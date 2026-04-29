package core

import (
	"errors"
	"testing"
)

func TestTransientErrorIsAndAs(t *testing.T) {
	cause := errors.New("underlying network reset")
	te := &TransientError{Kind: TransientRateLimit, StatusCode: 429, Cause: cause, Message: "slow down"}

	// Error formatting includes Kind, status, and message.
	if got := te.Error(); got == "" {
		t.Errorf("Error() returned empty string")
	}

	// Unwrap returns the cause.
	if errors.Unwrap(te) != cause {
		t.Errorf("Unwrap() did not return cause")
	}

	// errors.As finds the typed error through wrapping.
	wrapped := errors.Join(errors.New("outer"), te)
	var found *TransientError
	if !errors.As(wrapped, &found) {
		t.Errorf("errors.As did not find *TransientError in wrapped chain")
	}

	// IsTransient is the convenience helper.
	if !IsTransient(te) {
		t.Errorf("IsTransient(te) = false")
	}
	if IsPermanent(te) {
		t.Errorf("IsPermanent(te) = true")
	}

	// errors.Is matches on Kind when both sides specify Kind.
	other := &TransientError{Kind: TransientRateLimit}
	if !errors.Is(te, other) {
		t.Errorf("errors.Is should match same Kind")
	}
	differentKind := &TransientError{Kind: TransientGateway}
	if errors.Is(te, differentKind) {
		t.Errorf("errors.Is should not match different Kind")
	}
	// Empty target Kind matches any Kind.
	anyTransient := &TransientError{}
	if !errors.Is(te, anyTransient) {
		t.Errorf("errors.Is should match empty-Kind target")
	}
}

func TestPermanentErrorIsAndAs(t *testing.T) {
	pe := &PermanentError{Kind: PermanentAuth, StatusCode: 401, Message: "invalid api key"}
	if !IsPermanent(pe) {
		t.Errorf("IsPermanent(pe) = false")
	}
	if IsTransient(pe) {
		t.Errorf("IsTransient(pe) = true")
	}

	other := &PermanentError{Kind: PermanentAuth}
	if !errors.Is(pe, other) {
		t.Errorf("errors.Is should match same Kind")
	}
	differentKind := &PermanentError{Kind: PermanentContextLength}
	if errors.Is(pe, differentKind) {
		t.Errorf("errors.Is should not match different Kind")
	}
}

func TestErrorTypesAreDisjoint(t *testing.T) {
	te := &TransientError{Kind: TransientTimeout}
	pe := &PermanentError{Kind: PermanentAuth}
	if errors.Is(te, pe) {
		t.Errorf("transient should not match permanent")
	}
	if errors.Is(pe, te) {
		t.Errorf("permanent should not match transient")
	}
}
