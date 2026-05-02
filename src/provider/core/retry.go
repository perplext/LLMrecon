// Provider retry helper introduced in v0.9.0.
//
// RetryableQuery wraps a single provider call with exponential-backoff retry
// for *TransientError, no retry for *PermanentError, and a hard cap on attempts
// and wall-clock time.
//
// Modules using a provider should call this helper rather than re-implementing
// backoff. On exhausted retries, the helper returns the last *TransientError;
// callers translate that to OutcomeSkipped + SkipProviderError on the result.

package core

import (
	"context"
	"errors"
	"math/rand"
	"time"
)

// RetryPolicy controls retry behavior. Zero value gives sane defaults via
// DefaultRetryPolicy.
type RetryPolicy struct {
	// MaxAttempts is the total number of attempts (initial + retries).
	// Must be >= 1; values < 1 are coerced to 1.
	MaxAttempts int
	// InitialBackoff is the delay before the first retry.
	InitialBackoff time.Duration
	// MaxBackoff caps any single backoff delay.
	MaxBackoff time.Duration
	// BackoffMultiplier is applied to the delay after each transient failure.
	BackoffMultiplier float64
	// JitterFraction adds [0, JitterFraction) * delay random jitter to each
	// backoff delay. Set to 0 for deterministic backoff.
	JitterFraction float64
	// MaxTotalWallClock caps the total wall-clock time spent retrying.
	// Zero means no wall-clock limit (only MaxAttempts applies).
	MaxTotalWallClock time.Duration
}

// DefaultRetryPolicy returns a reasonable policy: 3 attempts, exponential
// backoff starting at 500ms, capped at 30s, with 25% jitter, and a 60s
// total wall-clock budget.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:       3,
		InitialBackoff:    500 * time.Millisecond,
		MaxBackoff:        30 * time.Second,
		BackoffMultiplier: 2.0,
		JitterFraction:    0.25,
		MaxTotalWallClock: 60 * time.Second,
	}
}

// RetryableQuery invokes fn with exponential-backoff retry on *TransientError.
// Permanent errors are returned immediately. ctx cancellation aborts the
// retry loop with the context error.
//
// On exhausted retries, the most recent error is returned (typically a
// *TransientError). Callers should map that to OutcomeSkipped + SkipProviderError.
func RetryableQuery[T any](ctx context.Context, policy RetryPolicy, fn func(ctx context.Context) (T, error)) (T, error) {
	var zero T
	if policy.MaxAttempts < 1 {
		policy.MaxAttempts = 1
	}
	deadline := time.Now().Add(policy.MaxTotalWallClock)
	hasDeadline := policy.MaxTotalWallClock > 0

	delay := policy.InitialBackoff
	var lastErr error

	for attempt := 1; attempt <= policy.MaxAttempts; attempt++ {
		// Check ctx cancellation before each attempt — including the first.
		// fn may make a network call; if ctx is already cancelled we should
		// surface that without spending a syscall on it.
		//
		// When ctx is cancelled, return ctx.Err() unconditionally — even when
		// a prior iteration left lastErr set. The post-fn ctx check below
		// (after the call) returns ctx.Err() too, and these two paths must
		// agree: ctx cancellation is operator intent ("stop now") that
		// trumps a previous transient. Returning lastErr here lets the
		// transient mask the cancellation, which races the
		// ContextCancellationAbortsLoop test when the ctx-vs-timer select
		// goes to the timer first.
		if err := ctx.Err(); err != nil {
			return zero, err
		}
		// Check overall budget before each attempt (after the first).
		if attempt > 1 && hasDeadline && time.Now().After(deadline) {
			if lastErr != nil {
				return zero, lastErr
			}
			return zero, context.DeadlineExceeded
		}

		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}

		// Permanent: return immediately, no retry.
		if IsPermanent(err) {
			return zero, err
		}

		// Context cancelled: surface that, not the inner error.
		if ctxErr := ctx.Err(); ctxErr != nil {
			return zero, ctxErr
		}

		// Only *TransientError is retry-eligible. Unknown error types are
		// surfaced immediately so the caller can decide. Without this gate,
		// a buggy provider returning bare errors would stall the whole
		// retry budget on every call.
		if !IsTransient(err) {
			return zero, err
		}

		lastErr = err

		// Last attempt: don't sleep, just return.
		if attempt == policy.MaxAttempts {
			break
		}

		// Compute backoff: prefer Retry-After if the transient error supplies it.
		// Retry-After is the server's authoritative wait — it overrides our
		// local MaxBackoff (the server knows when it'll be available; we don't).
		// The wall-clock budget still applies as a hard ceiling below.
		sleep := delay
		fromRetryAfter := false
		var te *TransientError
		if errors.As(err, &te) && te.RetryAfter > 0 {
			sleep = te.RetryAfter
			fromRetryAfter = true
		}
		if policy.JitterFraction > 0 {
			sleep += time.Duration(rand.Float64()*policy.JitterFraction*float64(sleep)) // #nosec G404 -- math/rand used for non-security randomization (jitter/simulation/load distribution)
		}
		if !fromRetryAfter && sleep > policy.MaxBackoff {
			sleep = policy.MaxBackoff
		}
		if hasDeadline {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				return zero, lastErr
			}
			if sleep > remaining {
				sleep = remaining
			}
		}

		// Sleep with context-cancellation awareness.
		t := time.NewTimer(sleep)
		select {
		case <-ctx.Done():
			t.Stop()
			return zero, ctx.Err()
		case <-t.C:
		}

		// Geometric ramp for next iteration.
		delay = time.Duration(float64(delay) * policy.BackoffMultiplier)
		if delay > policy.MaxBackoff {
			delay = policy.MaxBackoff
		}
	}

	return zero, lastErr
}
