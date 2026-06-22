package trail

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readAllEntries reads every *.log file in dir and returns the parsed entries.
// FileLogger names files with second-granularity timestamps, so size-based
// rotations within the same second land in one file; reading the whole
// directory makes the test robust to how many files rotation produced.
func readAllEntries(t *testing.T, dir string) []*AuditLog {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, "*.log"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	var entries []*AuditLog
	for _, path := range matches {
		data, err := os.ReadFile(path) // #nosec G304 -- test-controlled temp path
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			if line == "" {
				continue
			}
			entry, err := FromJSON(line)
			if err != nil {
				t.Fatalf("parse line %q: %v", line, err)
			}
			entries = append(entries, entry)
		}
	}
	return entries
}

func TestFileLogger_WritesRetrievableEntries(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewFileLogger(dir, 1<<20 /* 1 MiB, no rotation */, 5, false)
	if err != nil {
		t.Fatalf("NewFileLogger: %v", err)
	}
	defer logger.Close()

	for i := 0; i < 3; i++ {
		if err := logger.Log(context.Background(), NewAuditLog(OperationCreate, "c", "entry")); err != nil {
			t.Fatalf("Log: %v", err)
		}
	}

	entries := readAllEntries(t, dir)
	if len(entries) != 3 {
		t.Fatalf("expected 3 persisted entries, got %d", len(entries))
	}
	for _, e := range entries {
		if e.Operation != OperationCreate || e.Message != "entry" {
			t.Errorf("entry did not round-trip from disk: %+v", e)
		}
	}
}

func TestFileLogger_RotationPreservesAllEntries(t *testing.T) {
	dir := t.TempDir()
	// Tiny max size forces a rotation check before most writes.
	logger, err := NewFileLogger(dir, 16 /* bytes */, 10, false)
	if err != nil {
		t.Fatalf("NewFileLogger: %v", err)
	}
	defer logger.Close()

	const n = 8
	for i := 0; i < n; i++ {
		if err := logger.Log(context.Background(), NewAuditLog(OperationCreate, "c", "rot")); err != nil {
			t.Fatalf("Log %d: %v", i, err)
		}
	}

	// Whatever the file count, every entry must survive rotation.
	if got := len(readAllEntries(t, dir)); got != n {
		t.Fatalf("rotation lost entries: persisted %d, wrote %d", got, n)
	}
}

func TestFileLogger_RotationCreatesDistinctFiles(t *testing.T) {
	dir := t.TempDir()
	// 16-byte cap forces a rotation before nearly every write. All writes
	// happen within the same wall-clock second, which is exactly the case the
	// second-granularity filename scheme used to collide on.
	logger, err := NewFileLogger(dir, 16, 10, false)
	if err != nil {
		t.Fatalf("NewFileLogger: %v", err)
	}
	defer logger.Close()

	const n = 6
	for i := 0; i < n; i++ {
		if err := logger.Log(context.Background(), NewAuditLog(OperationCreate, "c", "rot")); err != nil {
			t.Fatalf("Log %d: %v", i, err)
		}
	}

	matches, err := filepath.Glob(filepath.Join(dir, "*.log"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	// Each rotation must produce a distinct file; same-second writes must not
	// silently share one filename.
	if len(matches) < 2 {
		t.Fatalf("expected rotation to create distinct files, got %d", len(matches))
	}
}

func TestFileLogger_CreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "audit")
	logger, err := NewFileLogger(dir, 1<<20, 5, false)
	if err != nil {
		t.Fatalf("NewFileLogger should create missing dirs: %v", err)
	}
	defer logger.Close()
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("expected directory to be created: %v", err)
	}
}

func TestInMemoryLogger_StoresImmutableCopy(t *testing.T) {
	logger := NewInMemoryLogger(10)
	log := NewAuditLog(OperationRead, "c", "original")
	if err := logger.Log(context.Background(), log); err != nil {
		t.Fatalf("Log: %v", err)
	}

	// Mutating the caller's struct must not change what the logger retained.
	log.Message = "mutated after logging"

	if logger.logs[0].Message != "original" {
		t.Errorf("logger stored a reference, not a copy: %q", logger.logs[0].Message)
	}
}

func TestInMemoryLogger_TrimsToMax(t *testing.T) {
	logger := NewInMemoryLogger(2)
	for _, msg := range []string{"first", "second", "third"} {
		if err := logger.Log(context.Background(), NewAuditLog(OperationRead, "c", msg)); err != nil {
			t.Fatalf("Log: %v", err)
		}
	}

	if len(logger.logs) != 2 {
		t.Fatalf("expected ring trimmed to 2, got %d", len(logger.logs))
	}
	// Oldest ("first") should have been dropped; newest two retained in order.
	if logger.logs[0].Message != "second" || logger.logs[1].Message != "third" {
		t.Errorf("trim dropped the wrong entries: %q, %q", logger.logs[0].Message, logger.logs[1].Message)
	}
}

func TestNewInMemoryLogger_DefaultsNonPositiveMax(t *testing.T) {
	if got := NewInMemoryLogger(0).maxLogs; got != 1000 {
		t.Errorf("expected default maxLogs 1000, got %d", got)
	}
	if got := NewInMemoryLogger(-5).maxLogs; got != 1000 {
		t.Errorf("expected default maxLogs 1000 for negative input, got %d", got)
	}
}
