package trail

import (
	"strings"
	"testing"
)

func TestNewAuditLog_Defaults(t *testing.T) {
	log := NewAuditLog(OperationCreate, "templates", "created a template")

	if log.ID == "" {
		t.Error("expected a generated ID")
	}
	if log.Level != LogLevelInfo {
		t.Errorf("default level: got %q want %q", log.Level, LogLevelInfo)
	}
	if log.Status != "success" {
		t.Errorf("default status: got %q want success", log.Status)
	}
	if log.Operation != OperationCreate || log.Component != "templates" {
		t.Errorf("operation/component not set: %+v", log)
	}
	if log.Metadata == nil {
		t.Error("metadata map should be initialized, not nil")
	}
	if log.Timestamp.IsZero() {
		t.Error("timestamp should be set")
	}
	if log.Timestamp.Location().String() != "UTC" {
		t.Errorf("timestamp should be UTC, got %s", log.Timestamp.Location())
	}
}

func TestAuditLog_JSONRoundTrip(t *testing.T) {
	original := NewAuditLog(OperationUpdate, "config", "changed setting").
		WithUser("u-1", "alice").
		WithStatus("success", 200).
		WithTags("security", "config").
		WithMetadata("region", "us-east")

	jsonStr, err := original.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}

	parsed, err := FromJSON(jsonStr)
	if err != nil {
		t.Fatalf("FromJSON: %v", err)
	}

	if parsed.ID != original.ID ||
		parsed.Operation != original.Operation ||
		parsed.UserID != original.UserID ||
		parsed.User != original.User ||
		parsed.StatusCode != original.StatusCode {
		t.Errorf("round trip mismatch:\n got %+v\nwant %+v", parsed, original)
	}
	if len(parsed.Tags) != 2 {
		t.Errorf("tags lost in round trip: %v", parsed.Tags)
	}
}

func TestFromJSON_RejectsMalformed(t *testing.T) {
	if _, err := FromJSON("{not valid json"); err == nil {
		t.Fatal("expected an error parsing malformed JSON")
	}
}

func TestWithError_SetsErrorStatus(t *testing.T) {
	log := NewAuditLog(OperationExecute, "runner", "ran").
		WithError("E_TIMEOUT", "operation timed out")

	if log.Status != "error" {
		t.Errorf("WithError should flip status to error, got %q", log.Status)
	}
	if log.ErrorCode != "E_TIMEOUT" || log.ErrorMessage != "operation timed out" {
		t.Errorf("error fields not set: %+v", log)
	}
}

func TestBuilderChain_AccumulatesFields(t *testing.T) {
	log := NewAuditLog(OperationAuth, "login", "ok").
		WithLevel(LogLevelWarning).
		WithSession("sess-1").
		WithRequest("req-1", "trace-1").
		WithClient("10.0.0.1", "curl/8").
		WithResource("account", "acct-1").
		WithAction("login").
		WithDuration(42)

	if log.Level != LogLevelWarning ||
		log.SessionID != "sess-1" ||
		log.RequestID != "req-1" ||
		log.TraceID != "trace-1" ||
		log.IPAddress != "10.0.0.1" ||
		log.UserAgent != "curl/8" ||
		log.Resource != "account" ||
		log.ResourceID != "acct-1" ||
		log.Action != "login" ||
		log.Duration != 42 {
		t.Errorf("builder chain dropped a field: %+v", log)
	}
}

func TestWithMetadata_InitializesNilMap(t *testing.T) {
	// A zero-value AuditLog has a nil Metadata map; WithMetadata must not panic.
	log := &AuditLog{}
	log.WithMetadata("k", "v")
	if log.Metadata["k"] != "v" {
		t.Error("WithMetadata should lazily initialize the metadata map")
	}
}

func TestWithVersionAndVerification(t *testing.T) {
	log := NewAuditLog(OperationDeploy, "release", "deployed").
		WithVersion("1.0.0", "1.1.0", "minor").
		WithVerification(true, "checksum", map[string]interface{}{"sha": "abc"})

	if log.Version == nil || log.Version.Previous != "1.0.0" || log.Version.Current != "1.1.0" || log.Version.ChangeType != "minor" {
		t.Errorf("WithVersion not applied: %+v", log.Version)
	}
	if log.Verification == nil || !log.Verification.Success || log.Verification.Method != "checksum" {
		t.Errorf("WithVerification not applied: %+v", log.Verification)
	}
	if log.Verification.Details["sha"] != "abc" {
		t.Errorf("verification details lost: %+v", log.Verification.Details)
	}
}

func TestGenerateID_UniqueAndHex(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := generateID()
		if len(id) != 32 { // 16 random bytes hex-encoded
			t.Fatalf("unexpected ID length %d for %q", len(id), id)
		}
		if strings.TrimLeft(id, "0123456789abcdef") != "" {
			t.Fatalf("ID is not lowercase hex: %q", id)
		}
		if seen[id] {
			t.Fatalf("duplicate ID generated: %q", id)
		}
		seen[id] = true
	}
}
