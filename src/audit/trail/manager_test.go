// Package trail provides a comprehensive audit trail and logging system.
//
// These are white-box tests (package trail) so they can exercise unexported
// helpers (logLevelValue) and inspect the manager's internal state directly.
package trail

import (
	"context"
	"errors"
	"testing"
	"time"
)

// captureLogger is a minimal AuditLogger test double that records every log it
// receives. A capturing logger is unavoidable here: InMemoryLogger keeps its
// backing slice unexported with no getter, so there is no production logger we
// can read entries back out of. This double tests the *manager's* dispatch
// behavior, not a mock's behavior.
type captureLogger struct {
	id       string
	received []*AuditLog
}

func (l *captureLogger) GetID() string { return l.id }

func (l *captureLogger) Log(_ context.Context, log *AuditLog) error {
	l.received = append(l.received, log)
	return nil
}

func (l *captureLogger) Close() error { return nil }

// failingLogger always returns an error from Log. There is no production logger
// that deterministically fails, so this double is required to test the
// manager's "continue on logger failure" contract.
type failingLogger struct {
	id  string
	err error
}

func (l *failingLogger) GetID() string { return l.id }

func (l *failingLogger) Log(_ context.Context, _ *AuditLog) error { return l.err }

func (l *failingLogger) Close() error { return nil }

// queryExportLogger implements the optional AuditQueryLogger and AuditExporter
// interfaces so the manager's Query/Export delegation can be exercised. No
// concrete production logger implements these, so a double is required.
type queryExportLogger struct {
	id           string
	queryResult  *LogQueryResult
	exportResult []byte
}

func (l *queryExportLogger) GetID() string                            { return l.id }
func (l *queryExportLogger) Log(_ context.Context, _ *AuditLog) error { return nil }
func (l *queryExportLogger) Close() error                             { return nil }

func (l *queryExportLogger) Query(_ context.Context, _ *LogQuery) (*LogQueryResult, error) {
	return l.queryResult, nil
}

func (l *queryExportLogger) Export(_ context.Context, _ []*AuditLog, _ ExportFormat) ([]byte, error) {
	return l.exportResult, nil
}

func tamperEvidentManager(t *testing.T) *AuditTrailManager {
	t.Helper()
	m, err := NewAuditTrailManager(&AuditConfig{
		MinLogLevel:   LogLevelDebug,
		TamperEvident: true,
		SigningKey:    "test-signing-key-32-bytes-or-more!!",
	})
	if err != nil {
		t.Fatalf("NewAuditTrailManager: %v", err)
	}
	return m
}

// --- Tamper-evidence: the audit trail's crown jewel -------------------------

func TestVerifyLogIntegrity_AcceptsUntamperedLog(t *testing.T) {
	m := tamperEvidentManager(t)
	log := NewAuditLog(OperationAuth, "login", "user authenticated")

	if err := m.Log(context.Background(), log); err != nil {
		t.Fatalf("Log: %v", err)
	}
	if log.Signature == "" {
		t.Fatal("expected Log to populate a signature when tamper-evident")
	}

	ok, err := m.VerifyLogIntegrity(log)
	if err != nil {
		t.Fatalf("VerifyLogIntegrity: %v", err)
	}
	if !ok {
		t.Fatal("expected untampered log to verify as intact")
	}
}

func TestVerifyLogIntegrity_DetectsTampering(t *testing.T) {
	m := tamperEvidentManager(t)
	log := NewAuditLog(OperationDelete, "templates", "deleted template X")
	if err := m.Log(context.Background(), log); err != nil {
		t.Fatalf("Log: %v", err)
	}

	// Mutate the signed entry the way an attacker covering their tracks would.
	log.Message = "deleted template Y"

	ok, err := m.VerifyLogIntegrity(log)
	if err != nil {
		t.Fatalf("VerifyLogIntegrity: %v", err)
	}
	if ok {
		t.Fatal("expected a mutated log to fail integrity verification")
	}
}

func TestVerifyLogIntegrity_RestoresSignatureAfterCheck(t *testing.T) {
	// VerifyLogIntegrity temporarily blanks Signature to recompute it; it must
	// put the original back so the entry stays verifiable on a second call.
	m := tamperEvidentManager(t)
	log := NewAuditLog(OperationConfig, "settings", "changed retention")
	if err := m.Log(context.Background(), log); err != nil {
		t.Fatalf("Log: %v", err)
	}
	original := log.Signature

	if _, err := m.VerifyLogIntegrity(log); err != nil {
		t.Fatalf("VerifyLogIntegrity: %v", err)
	}
	if log.Signature != original {
		t.Fatalf("signature not restored: got %q want %q", log.Signature, original)
	}
}

func TestVerifyLogIntegrity_ErrorsWhenNotTamperEvident(t *testing.T) {
	m, err := NewAuditTrailManager(&AuditConfig{MinLogLevel: LogLevelDebug})
	if err != nil {
		t.Fatalf("NewAuditTrailManager: %v", err)
	}
	if _, err := m.VerifyLogIntegrity(NewAuditLog(OperationRead, "c", "m")); err == nil {
		t.Fatal("expected an error when tamper-evident logging is disabled")
	}
}

func TestVerifyLogIntegrity_ErrorsOnMissingSignature(t *testing.T) {
	m := tamperEvidentManager(t)
	// A log that was never passed through Log() has no signature.
	if _, err := m.VerifyLogIntegrity(NewAuditLog(OperationRead, "c", "m")); err == nil {
		t.Fatal("expected an error for a log with no signature")
	}
}

// --- Redaction: secrets must never reach a logger ---------------------------

func TestLog_RedactsConfiguredMetadataFields(t *testing.T) {
	m, err := NewAuditTrailManager(&AuditConfig{
		MinLogLevel:         LogLevelDebug,
		RedactSensitiveInfo: true,
		RedactFields:        []string{"password", "token"},
	})
	if err != nil {
		t.Fatalf("NewAuditTrailManager: %v", err)
	}

	log := NewAuditLog(OperationAuth, "login", "attempt").
		WithMetadata("password", "hunter2").
		WithMetadata("token", "abc123").
		WithMetadata("username", "alice")
	if err := m.Log(context.Background(), log); err != nil {
		t.Fatalf("Log: %v", err)
	}

	if got := log.Metadata["password"]; got != "[REDACTED]" {
		t.Errorf("password not redacted: got %v", got)
	}
	if got := log.Metadata["token"]; got != "[REDACTED]" {
		t.Errorf("token not redacted: got %v", got)
	}
	if got := log.Metadata["username"]; got != "alice" {
		t.Errorf("non-sensitive field altered: got %v", got)
	}
}

func TestLog_RedactsChangeBeforeAndAfterStates(t *testing.T) {
	m, err := NewAuditTrailManager(&AuditConfig{
		MinLogLevel:         LogLevelDebug,
		RedactSensitiveInfo: true,
		RedactFields:        []string{"secret"},
	})
	if err != nil {
		t.Fatalf("NewAuditTrailManager: %v", err)
	}

	log := NewAuditLog(OperationUpdate, "creds", "rotated").
		WithChanges(
			map[string]interface{}{"secret": "old"},
			map[string]interface{}{"secret": "new"},
			[]string{"secret"}, "rotation")
	if err := m.Log(context.Background(), log); err != nil {
		t.Fatalf("Log: %v", err)
	}

	if got := log.Changes.Before["secret"]; got != "[REDACTED]" {
		t.Errorf("before-state secret not redacted: got %v", got)
	}
	if got := log.Changes.After["secret"]; got != "[REDACTED]" {
		t.Errorf("after-state secret not redacted: got %v", got)
	}
}

// --- Level filtering --------------------------------------------------------

func TestLog_DropsEntriesBelowMinLevel(t *testing.T) {
	m, err := NewAuditTrailManager(&AuditConfig{MinLogLevel: LogLevelWarning})
	if err != nil {
		t.Fatalf("NewAuditTrailManager: %v", err)
	}
	cap := &captureLogger{id: "cap"}
	m.AddLogger(cap)

	// Below threshold: must be dropped silently.
	if err := m.Log(context.Background(), NewAuditLog(OperationRead, "c", "debug").WithLevel(LogLevelInfo)); err != nil {
		t.Fatalf("Log(info): %v", err)
	}
	if len(cap.received) != 0 {
		t.Fatalf("info entry should have been dropped below warning threshold, got %d", len(cap.received))
	}

	// At/above threshold: must be delivered.
	if err := m.Log(context.Background(), NewAuditLog(OperationRead, "c", "warn").WithLevel(LogLevelError)); err != nil {
		t.Fatalf("Log(error): %v", err)
	}
	if len(cap.received) != 1 {
		t.Fatalf("error entry should have been delivered, got %d", len(cap.received))
	}
}

func TestLogLevelValue_Ordering(t *testing.T) {
	levels := []LogLevel{LogLevelDebug, LogLevelInfo, LogLevelWarning, LogLevelError, LogLevelCritical}
	for i := 1; i < len(levels); i++ {
		if logLevelValue(levels[i-1]) >= logLevelValue(levels[i]) {
			t.Errorf("level %q should rank below %q", levels[i-1], levels[i])
		}
	}
	// Unknown levels default to info's rank.
	if logLevelValue(LogLevel("bogus")) != logLevelValue(LogLevelInfo) {
		t.Error("unknown level should default to info rank")
	}
}

// --- Dispatch resilience ----------------------------------------------------

func TestLog_ContinuesAfterLoggerFailure(t *testing.T) {
	m, err := NewAuditTrailManager(&AuditConfig{MinLogLevel: LogLevelDebug})
	if err != nil {
		t.Fatalf("NewAuditTrailManager: %v", err)
	}
	boom := errors.New("disk full")
	m.AddLogger(&failingLogger{id: "bad", err: boom})
	good := &captureLogger{id: "good"}
	m.AddLogger(good)

	err = m.Log(context.Background(), NewAuditLog(OperationCreate, "c", "m"))
	if !errors.Is(err, boom) {
		t.Fatalf("expected the failing logger's error to surface, got %v", err)
	}
	if len(good.received) != 1 {
		t.Fatalf("healthy logger should still receive the entry, got %d", len(good.received))
	}
}

// --- Logger management & query validation -----------------------------------

func TestRemoveLogger(t *testing.T) {
	m, _ := NewAuditTrailManager(&AuditConfig{MinLogLevel: LogLevelDebug})
	m.AddLogger(&captureLogger{id: "keep"})

	if err := m.RemoveLogger("does-not-exist"); !errors.Is(err, ErrLoggerNotFound) {
		t.Fatalf("expected ErrLoggerNotFound, got %v", err)
	}
	if err := m.RemoveLogger("keep"); err != nil {
		t.Fatalf("removing existing logger: %v", err)
	}
}

func TestQuery_RejectsInvertedTimeRange(t *testing.T) {
	m, _ := NewAuditTrailManager(&AuditConfig{MinLogLevel: LogLevelDebug})
	now := time.Now()
	_, err := m.Query(context.Background(), &LogQuery{
		StartTime: now,
		EndTime:   now.Add(-time.Hour),
	})
	if !errors.Is(err, ErrInvalidTimeRange) {
		t.Fatalf("expected ErrInvalidTimeRange, got %v", err)
	}
}

func TestQuery_EmptyWhenNoLoggers(t *testing.T) {
	m, _ := NewAuditTrailManager(&AuditConfig{MinLogLevel: LogLevelDebug})
	res, err := m.Query(context.Background(), &LogQuery{})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if res.TotalCount != 0 || len(res.Logs) != 0 || res.HasMore {
		t.Fatalf("expected empty result, got %+v", res)
	}
}

func TestGetConfig_ReturnsDefensiveCopy(t *testing.T) {
	m := tamperEvidentManager(t)
	got := m.GetConfig()
	got.MinLogLevel = LogLevelCritical // mutate the copy

	if m.GetLogLevel() == LogLevelCritical {
		t.Fatal("GetConfig must return a copy; mutating it changed the manager")
	}
}

func TestSetLogLevel(t *testing.T) {
	m := tamperEvidentManager(t)
	m.SetLogLevel(LogLevelError)
	if m.GetLogLevel() != LogLevelError {
		t.Fatalf("SetLogLevel not reflected: got %q", m.GetLogLevel())
	}
}

func TestQuery_DelegatesToQueryableLogger(t *testing.T) {
	m, _ := NewAuditTrailManager(&AuditConfig{MinLogLevel: LogLevelDebug})
	want := &LogQueryResult{
		Logs:       []*AuditLog{NewAuditLog(OperationRead, "c", "hit")},
		TotalCount: 1,
	}
	// A plain capture logger does not support querying; the queryable one does.
	m.AddLogger(&captureLogger{id: "plain"})
	m.AddLogger(&queryExportLogger{id: "queryable", queryResult: want})

	got, err := m.Query(context.Background(), &LogQuery{})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if got.TotalCount != 1 || len(got.Logs) != 1 {
		t.Fatalf("expected delegation to the queryable logger, got %+v", got)
	}
}

func TestQuery_ErrorsWhenNoLoggerSupportsQuerying(t *testing.T) {
	m, _ := NewAuditTrailManager(&AuditConfig{MinLogLevel: LogLevelDebug})
	m.AddLogger(&captureLogger{id: "plain"}) // no Query method
	if _, err := m.Query(context.Background(), &LogQuery{}); err == nil {
		t.Fatal("expected an error when no logger supports querying")
	}
}

func TestExport_DelegatesToExporter(t *testing.T) {
	m, _ := NewAuditTrailManager(&AuditConfig{MinLogLevel: LogLevelDebug})
	m.AddLogger(&queryExportLogger{
		id:           "exporter",
		queryResult:  &LogQueryResult{Logs: []*AuditLog{}},
		exportResult: []byte("exported-bytes"),
	})

	out, err := m.Export(context.Background(), &LogQuery{}, FormatJSON)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if string(out) != "exported-bytes" {
		t.Fatalf("expected delegation to exporter, got %q", out)
	}
}

func TestDefaultAuditConfig_SecureDefaults(t *testing.T) {
	cfg := DefaultAuditConfig()
	// Security posture: tamper-evidence and redaction must be ON by default.
	if !cfg.TamperEvident {
		t.Error("tamper-evident logging should default to enabled")
	}
	if !cfg.RedactSensitiveInfo {
		t.Error("sensitive-info redaction should default to enabled")
	}
	for _, want := range []string{"password", "token", "key", "secret", "credential"} {
		found := false
		for _, f := range cfg.RedactFields {
			if f == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("default redact fields should include %q", want)
		}
	}
	if cfg.MinLogLevel != LogLevelInfo {
		t.Errorf("default min level: got %q want info", cfg.MinLogLevel)
	}
}

func TestLoggerGetID(t *testing.T) {
	if got := NewInMemoryLogger(10).GetID(); got != "memory-logger" {
		t.Errorf("InMemoryLogger ID: got %q", got)
	}
}

func TestClose_ClearsLoggers(t *testing.T) {
	m, _ := NewAuditTrailManager(&AuditConfig{MinLogLevel: LogLevelDebug})
	cap := &captureLogger{id: "c"}
	m.AddLogger(cap)
	if err := m.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	// After Close, a subsequent log reaches no logger.
	if err := m.Log(context.Background(), NewAuditLog(OperationRead, "c", "m")); err != nil {
		t.Fatalf("Log after Close: %v", err)
	}
	if len(cap.received) != 0 {
		t.Fatal("loggers should be cleared after Close")
	}
}
