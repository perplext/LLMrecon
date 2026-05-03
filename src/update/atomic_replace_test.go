package update

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestAtomicReplace_FirstInstall asserts a brand-new install (no
// pre-existing dest dir) lands the bundle and reports FirstInstall=true.
func TestAtomicReplace_FirstInstall(t *testing.T) {
	zipPath := makeTestZip(t, []zipEntry{
		{name: "templates/", isDir: true},
		{name: "templates/x.yaml", content: []byte("a: b\n"), mode: 0o644},
	})
	parent := t.TempDir()
	dest := filepath.Join(parent, "install-dir")

	res, err := AtomicReplaceFromZip(StagedApplyOptions{
		ArchivePath: zipPath,
		DestDir:     dest,
	})
	if err != nil {
		t.Fatalf("AtomicReplaceFromZip: %v", err)
	}
	if !res.FirstInstall {
		t.Error("FirstInstall = false, want true")
	}
	if res.BackupPath != "" {
		t.Errorf("BackupPath = %q, want empty for first install", res.BackupPath)
	}
	got, err := os.ReadFile(filepath.Join(dest, "templates/x.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "a: b\n" {
		t.Errorf("content = %q", got)
	}
}

// TestAtomicReplace_ReplaceExisting asserts that an existing dest is
// renamed to .bak and the new content goes in atomically.
func TestAtomicReplace_ReplaceExisting(t *testing.T) {
	zipPath := makeTestZip(t, []zipEntry{
		{name: "v2.yaml", content: []byte("version: 2\n"), mode: 0o644},
	})
	parent := t.TempDir()
	dest := filepath.Join(parent, "install")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "v1.yaml"), []byte("version: 1\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := AtomicReplaceFromZip(StagedApplyOptions{
		ArchivePath: zipPath,
		DestDir:     dest,
		KeepBackup:  true, // explicitly keep so the test can assert
	})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if res.FirstInstall {
		t.Error("FirstInstall = true; expected false (dest existed)")
	}
	if res.BackupPath == "" {
		t.Fatal("BackupPath is empty; want non-empty for replace-with-keep")
	}

	// New content present.
	got, err := os.ReadFile(filepath.Join(dest, "v2.yaml"))
	if err != nil {
		t.Fatalf("read new: %v", err)
	}
	if string(got) != "version: 2\n" {
		t.Errorf("v2 content = %q", got)
	}
	// Old content NOT present in dest.
	if _, err := os.Stat(filepath.Join(dest, "v1.yaml")); !os.IsNotExist(err) {
		t.Errorf("v1.yaml still in dest; want gone (err = %v)", err)
	}
	// Old content IS present in backup.
	old, err := os.ReadFile(filepath.Join(res.BackupPath, "v1.yaml"))
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(old) != "version: 1\n" {
		t.Errorf("backup v1 = %q", old)
	}
}

// TestAtomicReplace_BackupRemovedWhenNotKept asserts the default behavior:
// no --backup means the .bak is cleaned up after a successful apply.
func TestAtomicReplace_BackupRemovedWhenNotKept(t *testing.T) {
	zipPath := makeTestZip(t, []zipEntry{
		{name: "new.yaml", content: []byte("v: 2"), mode: 0o644},
	})
	parent := t.TempDir()
	dest := filepath.Join(parent, "install")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "old.yaml"), []byte("v: 1"), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := AtomicReplaceFromZip(StagedApplyOptions{
		ArchivePath: zipPath,
		DestDir:     dest,
		KeepBackup:  false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.BackupPath != "" {
		t.Errorf("BackupPath = %q, want empty when KeepBackup=false", res.BackupPath)
	}

	// Verify no .bak. siblings remain in parent.
	entries, _ := os.ReadDir(parent)
	for _, e := range entries {
		if strings.Contains(e.Name(), ".bak.") {
			t.Errorf("found leaked backup: %s", e.Name())
		}
	}
}

// TestAtomicReplace_ValidationFailure asserts a Validate error keeps
// the original dest untouched and cleans up the staged dir.
func TestAtomicReplace_ValidationFailure(t *testing.T) {
	zipPath := makeTestZip(t, []zipEntry{
		{name: "x.yaml", content: []byte("data"), mode: 0o644},
	})
	parent := t.TempDir()
	dest := filepath.Join(parent, "install")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	originalContent := []byte("ORIGINAL\n")
	if err := os.WriteFile(filepath.Join(dest, "marker.txt"), originalContent, 0o644); err != nil {
		t.Fatal(err)
	}

	wantErr := errors.New("manifest mismatch")
	_, err := AtomicReplaceFromZip(StagedApplyOptions{
		ArchivePath: zipPath,
		DestDir:     dest,
		Validate: func(stagedDir string) error {
			return wantErr
		},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "manifest mismatch") {
		t.Errorf("err = %v, want substring 'manifest mismatch'", err)
	}

	// Original dest untouched.
	got, err := os.ReadFile(filepath.Join(dest, "marker.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(originalContent) {
		t.Errorf("dest contents changed; got %q", got)
	}

	// No staged dir leaked.
	entries, _ := os.ReadDir(parent)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".llmrecon-staged-") {
			t.Errorf("leaked staged dir: %s", e.Name())
		}
	}
}

// TestAtomicReplace_RecoveryFromInterruptedApply simulates the kill-9
// window: dest renamed to .bak, but staged → dest rename never happened.
// RecoverFromInterruptedApply should restore.
func TestAtomicReplace_RecoveryFromInterruptedApply(t *testing.T) {
	parent := t.TempDir()
	dest := filepath.Join(parent, "install")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "marker.txt"), []byte("preserved"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Simulate the interrupted state directly: rename dest to a .bak
	// path, then run recovery.
	bakPath := dest + ".bak.20260101T000000Z"
	if err := os.Rename(dest, bakPath); err != nil {
		t.Fatalf("simulate rename: %v", err)
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatalf("dest should be absent before recovery; err = %v", err)
	}

	restored, err := RecoverFromInterruptedApply(dest)
	if err != nil {
		t.Fatalf("RecoverFromInterruptedApply: %v", err)
	}
	if restored != mustAbs(t, dest) {
		t.Errorf("restored = %q, want %q", restored, dest)
	}
	got, err := os.ReadFile(filepath.Join(dest, "marker.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "preserved" {
		t.Errorf("recovered content = %q, want preserved", got)
	}
}

// TestAtomicReplace_RecoverPicksMostRecentBackup asserts that with
// multiple .bak siblings, the highest timestamp wins.
func TestAtomicReplace_RecoverPicksMostRecentBackup(t *testing.T) {
	parent := t.TempDir()
	dest := filepath.Join(parent, "install")

	// Create two .bak directories with distinct contents.
	older := dest + ".bak.20250101T000000Z"
	newer := dest + ".bak.20260101T000000Z"
	for _, p := range []string{older, newer} {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(older, "x"), []byte("OLD"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(newer, "x"), []byte("NEW"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := RecoverFromInterruptedApply(dest); err != nil {
		t.Fatalf("recover: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dest, "x"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "NEW" {
		t.Errorf("recovered from older backup; got %q, want NEW", got)
	}
}

// TestAtomicReplace_RecoveryNoOpWhenDestExists asserts recovery only
// fires when dest is missing — running after a successful apply must
// not clobber the live install with a stale .bak.
func TestAtomicReplace_RecoveryNoOpWhenDestExists(t *testing.T) {
	parent := t.TempDir()
	dest := filepath.Join(parent, "install")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "live"), []byte("live data"), 0o644); err != nil {
		t.Fatal(err)
	}
	bak := dest + ".bak.20260101T000000Z"
	if err := os.MkdirAll(bak, 0o755); err != nil {
		t.Fatal(err)
	}

	restored, err := RecoverFromInterruptedApply(dest)
	if err != nil {
		t.Fatal(err)
	}
	if restored != "" {
		t.Errorf("restored = %q, want empty (no-op)", restored)
	}
	// Live data still in place.
	got, err := os.ReadFile(filepath.Join(dest, "live"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "live data" {
		t.Errorf("live = %q, want unchanged", got)
	}
}

// TestAtomicReplace_RecoveryNoOpWhenNoBackup asserts recovery returns
// empty + nil when there's neither a dest nor a .bak.
func TestAtomicReplace_RecoveryNoOpWhenNoBackup(t *testing.T) {
	parent := t.TempDir()
	dest := filepath.Join(parent, "install")
	restored, err := RecoverFromInterruptedApply(dest)
	if err != nil {
		t.Fatal(err)
	}
	if restored != "" {
		t.Errorf("restored = %q, want empty", restored)
	}
}

func mustAbs(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}
