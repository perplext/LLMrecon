package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/perplext/LLMrecon/src/bundle"
)

// writeMinimalBundle writes a directory with a manifest.json that
// passes ValidateManifest. Used by both verify and import tests.
func writeMinimalBundle(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	manifest := bundle.BundleManifest{
		SchemaVersion: "1.0.0",
		BundleID:      "test-bundle",
		BundleType:    bundle.TemplateBundleType,
		Name:          "test bundle",
		Description:   "v0.10.0 #177 verify test fixture",
		Version:       "0.10.0",
		CreatedAt:     time.Now(),
		Author:        bundle.Author{Name: "tester", Email: "test@example.com"},
		Content:       []bundle.ContentItem{},
		Checksums:     bundle.Checksums{Manifest: "", Content: map[string]string{}},
		Compatibility: bundle.Compatibility{MinVersion: "0.10.0"},
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestParseValidationLevel covers the operator-friendly --level
// string → bundle.ValidationLevel mapping. Exercised by both
// bundle_import.go and (via offline CLI's switch) other entry points.
func TestParseValidationLevel(t *testing.T) {
	cases := []struct {
		in      string
		want    bundle.ValidationLevel
		wantErr bool
	}{
		{"", bundle.StandardValidation, false},
		{"basic", bundle.BasicValidation, false},
		{"BASIC", bundle.BasicValidation, false},
		{"standard", bundle.StandardValidation, false},
		{"strict", bundle.StrictValidation, false},
		{"manifest", bundle.ManifestValidationLevel, false},
		{"checksum", bundle.ChecksumValidationLevel, false},
		{"signature", bundle.SignatureValidationLevel, false},
		{"compatibility", bundle.CompatibilityValidationLevel, false},
		{"  strict  ", bundle.StrictValidation, false},
		{"unknown", "", true},
		{"sig", "", true},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			got, err := parseValidationLevel(c.in)
			if c.wantErr {
				if err == nil {
					t.Errorf("parseValidationLevel(%q) = %v, want error", c.in, got)
				}
				return
			}
			if err != nil {
				t.Errorf("parseValidationLevel(%q) error: %v", c.in, err)
			}
			if got != c.want {
				t.Errorf("parseValidationLevel(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// runVerify invokes bundleVerifyCmd.RunE directly with the given args
// and flag values. Bypasses cobra's argument parsing (which would walk
// up to the root command and emit the help banner) so tests can
// exercise the RunE logic in isolation.
func runVerify(t *testing.T, path, level, publicKey string) (stdout, stderr *bytes.Buffer, err error) {
	t.Helper()
	stdout = &bytes.Buffer{}
	stderr = &bytes.Buffer{}
	bundleVerifyCmd.SetOut(stdout)
	bundleVerifyCmd.SetErr(stderr)
	// Reset flags between runs so a previous test's --level doesn't
	// bleed into this one.
	if err := bundleVerifyCmd.Flags().Set("level", level); err != nil {
		t.Fatalf("set --level: %v", err)
	}
	if err := bundleVerifyCmd.Flags().Set("public-key", publicKey); err != nil {
		t.Fatalf("set --public-key: %v", err)
	}
	err = bundleVerifyCmd.RunE(bundleVerifyCmd, []string{path})
	return stdout, stderr, err
}

// TestBundleVerify_DefaultManifestLevel runs verify against a
// well-formed bundle with the default level and expects success.
func TestBundleVerify_DefaultManifestLevel(t *testing.T) {
	dir := writeMinimalBundle(t)
	stdout, stderr, err := runVerify(t, dir, "manifest", "")
	if err != nil {
		t.Fatalf("verify failed: %v\nstderr=%s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "✓") {
		t.Errorf("stdout missing success marker: %q", stdout.String())
	}
}

// TestBundleVerify_RejectsInvalidLevel asserts unknown --level values
// fail fast before touching the bundle.
func TestBundleVerify_RejectsInvalidLevel(t *testing.T) {
	dir := writeMinimalBundle(t)
	_, _, err := runVerify(t, dir, "garbage", "")
	if err == nil {
		t.Fatal("expected error for invalid --level")
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("err = %v, want 'unknown' substring", err)
	}
}

// TestBundleVerify_SignatureLevelRequiresPublicKey asserts that
// --level=signature without --public-key fails fast.
func TestBundleVerify_SignatureLevelRequiresPublicKey(t *testing.T) {
	dir := writeMinimalBundle(t)
	_, _, err := runVerify(t, dir, "signature", "")
	if err == nil {
		t.Fatal("expected error for missing --public-key")
	}
	if !strings.Contains(err.Error(), "public-key") {
		t.Errorf("err = %v, want substring 'public-key'", err)
	}
}

// TestBundleVerify_RejectsMissingBundle asserts a path that doesn't
// exist exits with a non-zero error.
func TestBundleVerify_RejectsMissingBundle(t *testing.T) {
	_, _, err := runVerify(t, "/nonexistent/path/to/bundle", "manifest", "")
	if err == nil {
		t.Fatal("expected error for missing path")
	}
}

// silence unused-import lint when the build tag flips.
var _ = errors.New
