package audit

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// mustGetEvents fails the test immediately if GetAuditEvents errors, so a
// retrieval failure surfaces as itself rather than as an ambiguous
// count/content mismatch downstream.
func mustGetEvents(t *testing.T, l *CredentialAuditLogger, limit int, filter map[string]string) []CredentialAuditEvent {
	t.Helper()
	events, err := l.GetAuditEvents(limit, filter)
	if err != nil {
		t.Fatalf("GetAuditEvents(%d, %v): %v", limit, filter, err)
	}
	return events
}

func newLogger(t *testing.T) (*CredentialAuditLogger, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "nested", "audit.log")
	l, err := NewCredentialAuditLogger(path, CredentialAuditLoggerOptions{
		UserIDProvider:   func() string { return "user-42" },
		SourceIPProvider: func() string { return "10.0.0.1" },
	})
	if err != nil {
		t.Fatalf("NewCredentialAuditLogger: %v", err)
	}
	return l, path
}

func TestLogCredentialAccess_RoundTrip(t *testing.T) {
	l, _ := newLogger(t)
	if err := l.LogCredentialAccess("cred-1", "openai", "read"); err != nil {
		t.Fatalf("LogCredentialAccess: %v", err)
	}
	events, err := l.GetAuditEvents(0, nil)
	if err != nil {
		t.Fatalf("GetAuditEvents: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	e := events[0]
	if e.EventType != "read_credential" || e.CredentialID != "cred-1" || e.Service != "openai" {
		t.Errorf("unexpected event fields: %+v", e)
	}
	if !e.Success {
		t.Errorf("access event should be Success=true")
	}
	if e.UserID != "user-42" || e.SourceIP != "10.0.0.1" {
		t.Errorf("provider-supplied UserID/SourceIP not recorded: %+v", e)
	}
	if e.Metadata["operation"] != "read" {
		t.Errorf("operation metadata missing: %v", e.Metadata)
	}
}

func TestLogCredentialError(t *testing.T) {
	l, _ := newLogger(t)
	if err := l.LogCredentialError("cred-2", "anthropic", "write", errors.New("denied")); err != nil {
		t.Fatal(err)
	}
	events := mustGetEvents(t, l, 0, nil)
	if len(events) != 1 || events[0].Success || events[0].ErrorMessage != "denied" {
		t.Errorf("error event not recorded correctly: %+v", events)
	}
	if events[0].EventType != "write_credential_error" {
		t.Errorf("event type = %q, want write_credential_error", events[0].EventType)
	}
}

func TestLogAlert(t *testing.T) {
	l, _ := newLogger(t)
	if err := l.LogAlert("suspicious access", "intrusion", map[string]string{"ip": "1.2.3.4"}); err != nil {
		t.Fatal(err)
	}
	events := mustGetEvents(t, l, 0, nil)
	if len(events) != 1 || events[0].EventType != "alert" {
		t.Fatalf("alert event not recorded: %+v", events)
	}
	if events[0].Metadata["alert_type"] != "intrusion" || events[0].Metadata["message"] != "suspicious access" {
		t.Errorf("alert metadata missing: %v", events[0].Metadata)
	}
}

func TestGetAuditEvents_Filter(t *testing.T) {
	l, _ := newLogger(t)
	_ = l.LogCredentialAccess("cred-A", "svc1", "read")
	_ = l.LogCredentialError("cred-B", "svc2", "delete", errors.New("x"))
	_ = l.LogCredentialAccess("cred-A", "svc1", "update")

	byID := mustGetEvents(t, l, 0, map[string]string{"credential_id": "cred-A"})
	if len(byID) != 2 {
		t.Errorf("credential_id filter: got %d, want 2", len(byID))
	}
	failed := mustGetEvents(t, l, 0, map[string]string{"success": "false"})
	if len(failed) != 1 || failed[0].CredentialID != "cred-B" {
		t.Errorf("success=false filter: got %+v", failed)
	}
	byType := mustGetEvents(t, l, 0, map[string]string{"event_type": "read_credential"})
	if len(byType) != 1 {
		t.Errorf("event_type filter: got %d, want 1", len(byType))
	}
}

func TestGetAuditEvents_Limit(t *testing.T) {
	l, _ := newLogger(t)
	for i := 0; i < 3; i++ {
		_ = l.LogCredentialAccess("c", "s", "read")
	}
	events := mustGetEvents(t, l, 2, nil)
	if len(events) != 2 {
		t.Errorf("limit=2: got %d events", len(events))
	}
}

func TestGetAuditEvents_NoFile(t *testing.T) {
	l, _ := newLogger(t) // nothing logged -> file doesn't exist yet
	events, err := l.GetAuditEvents(0, nil)
	if err != nil {
		t.Fatalf("unexpected error on missing file: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected empty slice for missing file, got %d", len(events))
	}
}

func TestRotateLogFile(t *testing.T) {
	l, path := newLogger(t)
	_ = l.LogCredentialAccess("c", "s", "read")

	if err := l.RotateLogFile(); err != nil {
		t.Fatalf("RotateLogFile: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("original log file should be gone after rotation")
	}
	// A backup (path.<unix>) should now exist.
	matches, _ := filepath.Glob(path + ".*")
	if len(matches) == 0 {
		t.Errorf("rotation should have produced a backup file")
	}
	// Fresh read returns empty (active file moved away).
	events := mustGetEvents(t, l, 0, nil)
	if len(events) != 0 {
		t.Errorf("post-rotation active log should be empty, got %d", len(events))
	}
}

func TestRotateLogFile_NoFileIsNoop(t *testing.T) {
	l, _ := newLogger(t)
	if err := l.RotateLogFile(); err != nil {
		t.Errorf("rotating a non-existent log should be a no-op, got %v", err)
	}
}
