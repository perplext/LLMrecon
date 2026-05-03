package cmd

import (
	"bytes"
	"strings"
	"testing"
)

// runImport mirrors runVerify in bundle_verify_test.go: invokes the
// command's RunE directly, bypassing cobra's argument parsing so the
// test exercises only the import logic.
func runImport(t *testing.T, path, target, level string, force bool) (stdout, stderr *bytes.Buffer, err error) {
	t.Helper()
	stdout = &bytes.Buffer{}
	stderr = &bytes.Buffer{}
	bundleImportCmd.SetOut(stdout)
	bundleImportCmd.SetErr(stderr)
	if err := bundleImportCmd.Flags().Set("target", target); err != nil {
		t.Fatalf("set --target: %v", err)
	}
	if err := bundleImportCmd.Flags().Set("level", level); err != nil {
		t.Fatalf("set --level: %v", err)
	}
	if err := bundleImportCmd.Flags().Set("backup", ""); err != nil {
		t.Fatalf("set --backup: %v", err)
	}
	forceVal := "false"
	if force {
		forceVal = "true"
	}
	if err := bundleImportCmd.Flags().Set("force", forceVal); err != nil {
		t.Fatalf("set --force: %v", err)
	}
	err = bundleImportCmd.RunE(bundleImportCmd, []string{path})
	return stdout, stderr, err
}

// TestBundleImport_RejectsEmptyTarget asserts the import fails fast
// without a --target flag, before touching the bundle.
func TestBundleImport_RejectsEmptyTarget(t *testing.T) {
	dir := writeMinimalBundle(t)
	_, _, err := runImport(t, dir, "", "standard", false)
	if err == nil {
		t.Fatal("expected error for empty --target")
	}
	if !strings.Contains(err.Error(), "target") {
		t.Errorf("err = %v, want 'target' substring", err)
	}
}

// TestBundleImport_RejectsInvalidLevel asserts a bad --level fails fast.
func TestBundleImport_RejectsInvalidLevel(t *testing.T) {
	dir := writeMinimalBundle(t)
	target := t.TempDir()
	_, _, err := runImport(t, dir, target, "garbage", false)
	if err == nil {
		t.Fatal("expected error for invalid --level")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("err = %v, want 'unknown' substring", err)
	}
}

// TestBundleImport_RejectsMissingBundle asserts the import fails fast
// when the bundle path doesn't exist.
func TestBundleImport_RejectsMissingBundle(t *testing.T) {
	target := t.TempDir()
	_, _, err := runImport(t, "/nonexistent/path/to/bundle", target, "standard", false)
	if err == nil {
		t.Fatal("expected error for missing bundle path")
	}
}
