package core

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func zeroJitterPolicy(maxAttempts int) RetryPolicy {
	return RetryPolicy{
		MaxAttempts:       maxAttempts,
		InitialBackoff:    1 * time.Millisecond,
		MaxBackoff:        10 * time.Millisecond,
		BackoffMultiplier: 2.0,
		JitterFraction:    0,
		MaxTotalWallClock: 5 * time.Second,
	}
}

func TestRetryableQuery_SuccessFirstAttempt(t *testing.T) {
	var calls int32
	got, err := RetryableQuery(context.Background(), zeroJitterPolicy(3), func(_ context.Context) (string, error) {
		atomic.AddInt32(&calls, 1)
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "ok" {
		t.Errorf("got %q, want %q", got, "ok")
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}

func TestRetryableQuery_RetriesTransient(t *testing.T) {
	var calls int32
	got, err := RetryableQuery(context.Background(), zeroJitterPolicy(3), func(_ context.Context) (string, error) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			return "", &TransientError{Kind: TransientRateLimit, Message: "throttled"}
		}
		return "eventually", nil
	})
	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if got != "eventually" {
		t.Errorf("got %q, want %q", got, "eventually")
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestRetryableQuery_DoesNotRetryPermanent(t *testing.T) {
	var calls int32
	_, err := RetryableQuery(context.Background(), zeroJitterPolicy(5), func(_ context.Context) (string, error) {
		atomic.AddInt32(&calls, 1)
		return "", &PermanentError{Kind: PermanentAuth, Message: "401"}
	})
	if err == nil {
		t.Fatal("expected permanent error")
	}
	var pe *PermanentError
	if !errors.As(err, &pe) {
		t.Errorf("expected *PermanentError, got %T", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (no retries on permanent)", calls)
	}
}

func TestRetryableQuery_ExhaustsAttemptsAndReturnsLastTransient(t *testing.T) {
	var calls int32
	_, err := RetryableQuery(context.Background(), zeroJitterPolicy(3), func(_ context.Context) (string, error) {
		atomic.AddInt32(&calls, 1)
		return "", &TransientError{Kind: TransientGateway, StatusCode: 503}
	})
	if err == nil {
		t.Fatal("expected error after exhausted retries")
	}
	if !IsTransient(err) {
		t.Errorf("expected transient, got %v", err)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestRetryableQuery_ContextCancellationAbortsLoop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var calls int32
	go func() {
		// Cancel mid-sleep on the first iteration.
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	policy := zeroJitterPolicy(10)
	// InitialBackoff alone isn't enough — zeroJitterPolicy sets MaxBackoff
	// to 10ms which would silently clamp the sleep below the cancel time
	// (20ms). Override BOTH so the cancel reliably lands during the first
	// iteration's backoff sleep, and the select-on-ctx-Done path is
	// deterministically exercised.
	policy.InitialBackoff = 200 * time.Millisecond
	policy.MaxBackoff = 200 * time.Millisecond
	_, err := RetryableQuery(ctx, policy, func(_ context.Context) (string, error) {
		atomic.AddInt32(&calls, 1)
		return "", &TransientError{Kind: TransientRateLimit}
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if calls < 1 {
		t.Errorf("expected at least 1 call, got %d", calls)
	}
}

// TestRetryableQuery_CtxCancelBetweenIterationsReturnsCtxErr targets the
// race window where the timer wins the select (sleep finishes first) but
// ctx is cancelled before the next iteration's check. Pre-fix, the
// top-of-loop check returned lastErr (the previous transient) instead of
// ctx.Err(), masking the cancellation.
//
// Synthesized deterministically: pre-cancel an already-set-up ctx, then
// run with a transient-returning fn that sets lastErr on the first call,
// and assert the final error wraps context.Canceled rather than the
// transient.
func TestRetryableQuery_CtxCancelBetweenIterationsReturnsCtxErr(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var calls int32
	policy := zeroJitterPolicy(3)
	// Tiny backoff so the timer wins the select reliably, leaving the
	// top-of-loop ctx check as the only place ctx.Err() can be observed.
	policy.InitialBackoff = 1 * time.Millisecond
	policy.MaxBackoff = 1 * time.Millisecond

	_, err := RetryableQuery(ctx, policy, func(_ context.Context) (string, error) {
		c := atomic.AddInt32(&calls, 1)
		// Cancel after the first transient is observed but before the
		// second call would run. With 1ms backoff the timer fires fast,
		// so the cancellation must land between iterations.
		if c == 1 {
			cancel()
		}
		return "", &TransientError{Kind: TransientRateLimit}
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v (lastErr should NOT mask ctx cancellation)", err)
	}
}

func TestRetryableQuery_ContextCancelDuringFn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel
	var calls int32
	_, err := RetryableQuery(ctx, zeroJitterPolicy(3), func(c context.Context) (string, error) {
		atomic.AddInt32(&calls, 1)
		return "", c.Err()
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// TestRetryableQuery_UnknownErrorReturnsImmediately verifies that errors
// that are neither *TransientError nor *PermanentError surface immediately.
// The retry budget must not absorb buggy provider returns.
func TestRetryableQuery_UnknownErrorReturnsImmediately(t *testing.T) {
	plain := errors.New("plain unknown error")
	var calls int32
	_, err := RetryableQuery(context.Background(), zeroJitterPolicy(5), func(_ context.Context) (string, error) {
		atomic.AddInt32(&calls, 1)
		return "", plain
	})
	if !errors.Is(err, plain) {
		t.Errorf("expected unknown error to surface, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call (no retry on unknown error), got %d", calls)
	}
}

// TestRetryableQuery_PreCancelledCtxSkipsFn verifies that when ctx is already
// cancelled before the first attempt, fn is NOT invoked at all. Avoids
// firing a provider network call when the caller has already given up.
func TestRetryableQuery_PreCancelledCtxSkipsFn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var calls int32
	_, err := RetryableQuery(ctx, zeroJitterPolicy(3), func(_ context.Context) (string, error) {
		atomic.AddInt32(&calls, 1)
		return "ok", nil
	})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if calls != 0 {
		t.Errorf("expected fn to be skipped on pre-cancelled ctx, got calls=%d", calls)
	}
}

func TestRetryableQuery_HonorsRetryAfter(t *testing.T) {
	var calls int32
	policy := zeroJitterPolicy(2)
	start := time.Now()
	_, _ = RetryableQuery(context.Background(), policy, func(_ context.Context) (string, error) {
		atomic.AddInt32(&calls, 1)
		return "", &TransientError{Kind: TransientRateLimit, RetryAfter: 30 * time.Millisecond}
	})
	elapsed := time.Since(start)
	if calls != 2 {
		t.Errorf("calls = %d, want 2", calls)
	}
	// Should have waited at least the RetryAfter (with some jitter slack).
	if elapsed < 25*time.Millisecond {
		t.Errorf("elapsed=%v < RetryAfter; helper did not honor it", elapsed)
	}
}

func TestRetryableQuery_MaxAttemptsCoercion(t *testing.T) {
	var calls int32
	policy := zeroJitterPolicy(0) // zero attempts -> coerced to 1
	_, _ = RetryableQuery(context.Background(), policy, func(_ context.Context) (string, error) {
		atomic.AddInt32(&calls, 1)
		return "", &TransientError{Kind: TransientGateway}
	})
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (zero MaxAttempts coerced to 1)", calls)
	}
}

func TestRetryableQuery_WallClockBudget(t *testing.T) {
	var calls int32
	policy := RetryPolicy{
		MaxAttempts:       100, // many attempts
		InitialBackoff:    20 * time.Millisecond,
		MaxBackoff:        20 * time.Millisecond,
		BackoffMultiplier: 1.0,
		JitterFraction:    0,
		MaxTotalWallClock: 50 * time.Millisecond, // tight budget
	}
	start := time.Now()
	_, _ = RetryableQuery(context.Background(), policy, func(_ context.Context) (string, error) {
		atomic.AddInt32(&calls, 1)
		return "", &TransientError{Kind: TransientGateway}
	})
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		t.Errorf("wall-clock budget not honored; elapsed=%v", elapsed)
	}
	if calls > 5 {
		t.Errorf("too many calls under wall-clock cap: %d", calls)
	}
}
