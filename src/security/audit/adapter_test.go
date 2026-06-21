package audit

import (
	"bytes"
	"errors"
	"testing"
)

func newAdapter(t *testing.T) (*AuditLoggerAdapter, *CredentialAuditLogger, *bytes.Buffer) {
	t.Helper()
	cl, _ := newLogger(t)
	buf := &bytes.Buffer{}
	return NewAuditLoggerAdapter(cl, buf, "tester"), cl, buf
}

func TestAdapter_LogCredentialAccess_WritesBothSinks(t *testing.T) {
	a, cl, buf := newAdapter(t)
	if err := a.LogCredentialAccess("cred-1", "openai", "read"); err != nil {
		t.Fatal(err)
	}
	// Credential logger sink:
	events, _ := cl.GetAuditEvents(0, nil)
	if len(events) != 1 {
		t.Errorf("credential logger should have 1 event, got %d", len(events))
	}
	// Standard audit logger sink:
	if buf.Len() == 0 {
		t.Errorf("standard audit logger should have written to the buffer")
	}
}

func TestAdapter_LogCredentialError(t *testing.T) {
	a, cl, buf := newAdapter(t)
	if err := a.LogCredentialError("cred-2", "svc", "delete", errors.New("nope")); err != nil {
		t.Fatal(err)
	}
	events, _ := cl.GetAuditEvents(0, nil)
	if len(events) != 1 || events[0].Success {
		t.Errorf("expected one failure event, got %+v", events)
	}
	if buf.Len() == 0 {
		t.Errorf("standard logger buffer empty")
	}
}

func TestAdapter_LogAlertAndKeyOperation(t *testing.T) {
	a, _, buf := newAdapter(t)
	if err := a.LogAlert("msg", "kind", map[string]string{"k": "v"}); err != nil {
		t.Fatal(err)
	}
	if err := a.LogKeyOperation("rotate", "key-1", "details"); err != nil {
		t.Fatal(err)
	}
	if buf.Len() == 0 {
		t.Errorf("expected alert + key-operation entries in the buffer")
	}
}

func TestAdapter_DelegatesGetAndRotateAndGetters(t *testing.T) {
	a, cl, _ := newAdapter(t)
	_ = a.LogCredentialAccess("c", "s", "read")

	if events, err := a.GetAuditEvents(0, nil); err != nil || len(events) != 1 {
		t.Errorf("GetAuditEvents delegation: events=%d err=%v", len(events), err)
	}
	if err := a.RotateLogFile(); err != nil {
		t.Errorf("RotateLogFile delegation: %v", err)
	}
	if a.GetCredentialAuditLogger() != cl {
		t.Errorf("GetCredentialAuditLogger should return the wired logger")
	}
	if a.GetStandardAuditLogger() == nil {
		t.Errorf("GetStandardAuditLogger should be non-nil for a wired adapter")
	}
}

func TestNullAdapter_AllMethodsAreSafeNoops(t *testing.T) {
	a := NewNullAuditLoggerAdapter()
	if err := a.LogCredentialAccess("c", "s", "read"); err != nil {
		t.Errorf("null LogCredentialAccess: %v", err)
	}
	if err := a.LogCredentialError("c", "s", "op", errors.New("x")); err != nil {
		t.Errorf("null LogCredentialError: %v", err)
	}
	if err := a.LogAlert("m", "t", nil); err != nil {
		t.Errorf("null LogAlert: %v", err)
	}
	if err := a.LogKeyOperation("op", "k", "d"); err != nil {
		t.Errorf("null LogKeyOperation: %v", err)
	}
	if err := a.RotateLogFile(); err != nil {
		t.Errorf("null RotateLogFile: %v", err)
	}
	events, err := a.GetAuditEvents(0, nil)
	if err != nil || len(events) != 0 {
		t.Errorf("null GetAuditEvents should be empty/no-error; got %d/%v", len(events), err)
	}
	if a.GetStandardAuditLogger() != nil || a.GetCredentialAuditLogger() != nil {
		t.Errorf("null adapter getters should return nil")
	}
}
