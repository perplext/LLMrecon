package cmd

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/perplext/LLMrecon/src/config"
)

// v0.10.0 #174 Tier 1: every "not implemented" stub in update_apply.go
// must return a non-nil error so the calling Run loop suppresses the
// "Successfully updated" message and the CLI exits non-zero.
//
// These tests are the contract: if a future change reverts a stub to
// `return nil`, this test catches it and the v0.10.0 honesty invariant
// is upheld.

func TestCreateBackup_ReturnsError(t *testing.T) {
	err := createBackup(&config.Config{})
	if err == nil {
		t.Fatal("createBackup returned nil; v0.10.0 #174 Tier 1 requires non-nil error from unimplemented stubs")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("error should mention 'not implemented'; got %q", err.Error())
	}
}

func TestApplyCoreBinaryUpdate_ReturnsError(t *testing.T) {
	err := applyCoreBinaryUpdate("/tmp/some-download.zip")
	if err == nil {
		t.Fatal("applyCoreBinaryUpdate returned nil; v0.10.0 #174 Tier 1 requires non-nil error")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("error should mention 'not implemented'; got %q", err.Error())
	}
	// The error message embeds the download path so operators can
	// recover by extracting it manually.
	if !strings.Contains(err.Error(), "/tmp/some-download.zip") {
		t.Errorf("error should include the download path so operators can recover; got %q", err.Error())
	}
}

func TestApplyTemplatesUpdate_ReturnsError(t *testing.T) {
	err := applyTemplatesUpdate("/tmp/templates.zip", "/usr/local/share/llmrecon/templates")
	if err == nil {
		t.Fatal("applyTemplatesUpdate returned nil; v0.10.0 #174 Tier 1 requires non-nil error")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("error should mention 'not implemented'; got %q", err.Error())
	}
	// Both paths should appear in the error so operators can recover.
	if !strings.Contains(err.Error(), "/tmp/templates.zip") {
		t.Errorf("error should include the download path; got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "/usr/local/share/llmrecon/templates") {
		t.Errorf("error should include the templates dir; got %q", err.Error())
	}
}

func TestApplyModuleUpdate_ReturnsError(t *testing.T) {
	err := applyModuleUpdate("/tmp/mod.zip", "best-of-n", "/usr/local/share/llmrecon/modules")
	if err == nil {
		t.Fatal("applyModuleUpdate returned nil; v0.10.0 #174 Tier 1 requires non-nil error")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("error should mention 'not implemented'; got %q", err.Error())
	}
	// Both the module ID and the modules dir should appear in the
	// error so operators can recover the right file in the right
	// location. Parity with TestApplyTemplatesUpdate_ReturnsError.
	if !strings.Contains(err.Error(), "best-of-n") {
		t.Errorf("error should include the module ID; got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "/usr/local/share/llmrecon/modules") {
		t.Errorf("error should include the modules dir; got %q", err.Error())
	}
}

// TestApplyStubsAreConsistent — meta-test asserting all four stubs
// follow the same pattern: error returned, "not implemented" mentioned.
// Catches regressions where one stub gets fixed but the others drift.
func TestApplyStubsAreConsistent(t *testing.T) {
	cases := []struct {
		name string
		fn   func() error
	}{
		{"createBackup", func() error { return createBackup(&config.Config{}) }},
		{"applyCoreBinaryUpdate", func() error { return applyCoreBinaryUpdate("x") }},
		{"applyTemplatesUpdate", func() error { return applyTemplatesUpdate("x", "y") }},
		{"applyModuleUpdate", func() error { return applyModuleUpdate("x", "y", "z") }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.fn(); err == nil {
				t.Errorf("%s returned nil; expected non-nil per v0.10.0 #174 Tier 1 honesty invariant", c.name)
			}
		})
	}
}

// makeTemplatesZip writes a minimal templates bundle to tempfile and
// returns the path. Used by Tier 2 cmd-level tests.
func makeTemplatesZip(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	zp := filepath.Join(dir, "templates.zip")
	f, err := os.Create(zp)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	zw := zip.NewWriter(f)
	w, err := zw.Create("templates/llm01.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("id: llm01\nversion: 0.10.0\n")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return zp
}

// makeModuleZip mirrors makeTemplatesZip for module bundles.
func makeModuleZip(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	zp := filepath.Join(dir, "module.zip")
	f, err := os.Create(zp)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	zw := zip.NewWriter(f)
	w, err := zw.Create("module.go")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("package m // stub\n")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return zp
}

// TestApplyTemplatesUpdateAtomic_FirstInstall is the v0.10.0 #174 Tier 2
// happy path: a fresh install lands the bundle into templatesDir.
func TestApplyTemplatesUpdateAtomic_FirstInstall(t *testing.T) {
	zp := makeTemplatesZip(t)
	parent := t.TempDir()
	dest := filepath.Join(parent, "templates")

	if err := applyTemplatesUpdateAtomic(zp, dest, false); err != nil {
		t.Fatalf("applyTemplatesUpdateAtomic: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dest, "templates/llm01.yaml"))
	if err != nil {
		t.Fatalf("read extracted: %v", err)
	}
	if !strings.Contains(string(got), "id: llm01") {
		t.Errorf("content = %q", got)
	}
}

// TestApplyTemplatesUpdateAtomic_RejectsEmptyBundle asserts the
// validateTemplatesBundle guard fires on empty bundles, leaving the
// pre-existing dest untouched.
func TestApplyTemplatesUpdateAtomic_RejectsEmptyBundle(t *testing.T) {
	dir := t.TempDir()
	zp := filepath.Join(dir, "empty.zip")
	f, err := os.Create(zp)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	parent := t.TempDir()
	dest := filepath.Join(parent, "templates")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	original := []byte("PRE-EXISTING\n")
	if err := os.WriteFile(filepath.Join(dest, "marker.txt"), original, 0o644); err != nil {
		t.Fatal(err)
	}

	err = applyTemplatesUpdateAtomic(zp, dest, false)
	if err == nil {
		t.Fatal("expected error for empty bundle")
	}

	// Pre-existing dest untouched.
	got, err := os.ReadFile(filepath.Join(dest, "marker.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Errorf("dest changed; got %q", got)
	}
}

// TestApplyTemplatesUpdateAtomic_KeepBackup asserts --backup keeps the
// .bak directory after a successful apply.
func TestApplyTemplatesUpdateAtomic_KeepBackup(t *testing.T) {
	zp := makeTemplatesZip(t)
	parent := t.TempDir()
	dest := filepath.Join(parent, "templates")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "old.yaml"), []byte("v1"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := applyTemplatesUpdateAtomic(zp, dest, true); err != nil {
		t.Fatal(err)
	}

	// .bak. sibling exists.
	entries, _ := os.ReadDir(parent)
	hasBak := false
	for _, e := range entries {
		if strings.Contains(e.Name(), ".bak.") {
			hasBak = true
		}
	}
	if !hasBak {
		t.Error("no .bak. sibling found; expected one with KeepBackup=true")
	}
}

// TestApplyModuleUpdateAtomic_RejectsEmptyModuleID asserts the empty-
// moduleID guard fires before any filesystem touch.
func TestApplyModuleUpdateAtomic_RejectsEmptyModuleID(t *testing.T) {
	zp := makeModuleZip(t)
	parent := t.TempDir()

	err := applyModuleUpdateAtomic(zp, "", parent, false)
	if err == nil {
		t.Fatal("expected error for empty moduleID")
	}
	if !strings.Contains(err.Error(), "empty moduleID") {
		t.Errorf("err = %v, want 'empty moduleID' substring", err)
	}
}

// TestApplyModuleUpdateAtomic_LandsAtModuleSubdir asserts module
// bundles land at modulesDir/<moduleID>/, not at modulesDir/.
func TestApplyModuleUpdateAtomic_LandsAtModuleSubdir(t *testing.T) {
	zp := makeModuleZip(t)
	parent := t.TempDir()

	if err := applyModuleUpdateAtomic(zp, "best-of-n", parent, false); err != nil {
		t.Fatalf("applyModuleUpdateAtomic: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(parent, "best-of-n", "module.go"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if !strings.Contains(string(got), "package m") {
		t.Errorf("content = %q", got)
	}
}
