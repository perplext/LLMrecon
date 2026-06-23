// Provider error taxonomy introduced in v0.9.0.
//
// Two struct types implementing errors.Is partition provider failures into
// retry-eligible (TransientError) and non-retry-eligible (PermanentError).
// Modules consuming providers should branch on errors.As, not on string
// matching:
//
//	var te *TransientError
//	if errors.As(err, &te) {
//	    time.Sleep(te.RetryAfter)
//	    // retry…
//	}
//
// Specific subcategories (rate limit, gateway, timeout, auth, content
// length) are carried as the Kind field on the struct so call sites can
// branch on a single typed value rather than juggling sentinel errors.

package core

import (
	"errors"
	"fmt"
	"time"
)

// TransientErrorKind classifies retry-eligible provider failures.
type TransientErrorKind string

const (
	TransientRateLimit TransientErrorKind = "rate_limit"
	TransientGateway   TransientErrorKind = "gateway"   // 502/503/504
	TransientTimeout   TransientErrorKind = "timeout"   // network or context deadline
	TransientServer    TransientErrorKind = "server"    // 5xx that doesn't fit the others
)

// PermanentErrorKind classifies non-retry-eligible provider failures.
type PermanentErrorKind string

const (
	PermanentAuth          PermanentErrorKind = "auth"           // 401/403
	PermanentBadRequest    PermanentErrorKind = "bad_request"    // 400, malformed prompt
	PermanentNotFound      PermanentErrorKind = "not_found"      // 404, model unavailable
	PermanentContextLength PermanentErrorKind = "context_length" // input too long
	PermanentModelMismatch PermanentErrorKind = "model_mismatch" // capability not supported by model
)

// TransientError represents a retry-eligible provider failure. Carries an
// optional RetryAfter duration for rate-limit responses (Retry-After header).
type TransientError struct {
	Kind       TransientErrorKind
	StatusCode int
	RequestID  string
	RetryAfter time.Duration
	Message    string
	Cause      error
}

// Error implements the error interface.
func (e *TransientError) Error() string {
	suffix := e.Message
	if suffix == "" && e.Cause != nil {
		suffix = e.Cause.Error()
	}
	if e.StatusCode != 0 {
		return fmt.Sprintf("transient error (%s, http %d): %s", e.Kind, e.StatusCode, suffix)
	}
	return fmt.Sprintf("transient error (%s): %s", e.Kind, suffix)
}

// Unwrap returns the underlying cause if any, supporting errors.Is/As chains.
func (e *TransientError) Unwrap() error { return e.Cause }

// Is reports whether the target matches this transient error. Matches if the
// target is also a *TransientError with the same Kind, or if Kind is empty.
func (e *TransientError) Is(target error) bool {
	t, ok := target.(*TransientError)
	if !ok {
		return false
	}
	return t.Kind == "" || e.Kind == t.Kind
}

// PermanentError represents a non-retry-eligible provider failure.
type PermanentError struct {
	Kind       PermanentErrorKind
	StatusCode int
	RequestID  string
	Message    string
	Cause      error
}

// Error implements the error interface.
func (e *PermanentError) Error() string {
	suffix := e.Message
	if suffix == "" && e.Cause != nil {
		suffix = e.Cause.Error()
	}
	if e.StatusCode != 0 {
		return fmt.Sprintf("permanent error (%s, http %d): %s", e.Kind, e.StatusCode, suffix)
	}
	return fmt.Sprintf("permanent error (%s): %s", e.Kind, suffix)
}

// Unwrap returns the underlying cause.
func (e *PermanentError) Unwrap() error { return e.Cause }

// Is reports whether the target matches this permanent error. Matches if the
// target is also a *PermanentError with the same Kind, or if Kind is empty.
func (e *PermanentError) Is(target error) bool {
	t, ok := target.(*PermanentError)
	if !ok {
		return false
	}
	return t.Kind == "" || e.Kind == t.Kind
}

// IsTransient reports whether err (or any error in its chain) is a
// *TransientError. Convenience for retry-loop branching.
func IsTransient(err error) bool {
	var te *TransientError
	return errors.As(err, &te)
}

// IsPermanent reports whether err (or any error in its chain) is a
// *PermanentError. Convenience for retry-loop branching.
func IsPermanent(err error) bool {
	var pe *PermanentError
	return errors.As(err, &pe)
}
