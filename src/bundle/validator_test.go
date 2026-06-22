package bundle

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/perplext/LLMrecon/src/version"
)

// buildValidatableBundle writes an on-disk bundle whose single content item's
// checksum matches the validator's hash format (raw hex, via calculateHash).
func buildValidatableBundle(t *testing.T) *Bundle {
	t.Helper()
	dir := t.TempDir()
	content := []byte("template-payload")
	if err := os.WriteFile(filepath.Join(dir, "a.yaml"), content, 0600); err != nil {
		t.Fatalf("write content: %v", err)
	}
	m := CreateBundleManifest("valid", "d", "1.0.0", MixedBundleType, Author{})
	m.AddContentItem("a.yaml", TemplateContentType, "id-a", "1.0", "")
	m.Content[0].Checksum = calculateHash(content)
	// manifest.json must exist on disk for ValidateChecksums.
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return &Bundle{Manifest: m, BundlePath: dir}
}

func newValidator() BundleValidator { return NewBundleValidator(io.Discard) }

func TestValidateManifest_Valid(t *testing.T) {
	m := sampleManifest() // no content items -> warning only, still valid
	res, err := newValidator().ValidateManifest(&m)
	if err != nil {
		t.Fatalf("ValidateManifest: %v", err)
	}
	if !res.Valid {
		t.Fatalf("manifest should be valid, errors: %v", res.Errors)
	}
}

func TestValidateManifest_Invalid(t *testing.T) {
	// Empty manifest: missing schema version, ID, name, and an invalid type.
	var empty BundleManifest
	res, _ := newValidator().ValidateManifest(&empty)
	if res.Valid {
		t.Fatal("empty manifest must be invalid")
	}
	if len(res.Errors) < 3 {
		t.Fatalf("expected several errors, got %v", res.Errors)
	}

	// Content item missing a checksum is rejected.
	m := sampleManifest()
	m.AddContentItem("a.yaml", TemplateContentType, "id", "1.0", "") // no checksum
	res, _ = newValidator().ValidateManifest(&m)
	if res.Valid {
		t.Fatal("content item without checksum must be invalid")
	}
}

func TestValidate_BasicRoutesToManifestOnly(t *testing.T) {
	// A manifest-only bundle (no content) passes basic without needing files.
	m := sampleManifest()
	b := &Bundle{Manifest: m, BundlePath: t.TempDir()}
	res, err := newValidator().Validate(b, BasicValidation)
	if err != nil {
		t.Fatalf("Validate(basic): %v", err)
	}
	if !res.Valid {
		t.Fatalf("basic validation should pass, errors: %v", res.Errors)
	}
}

func TestValidate_StandardChecksContent(t *testing.T) {
	b := buildValidatableBundle(t)
	res, err := newValidator().Validate(b, StandardValidation)
	if err != nil {
		t.Fatalf("Validate(standard): %v", err)
	}
	if !res.Valid {
		t.Fatalf("standard validation should pass, errors: %v", res.Errors)
	}

	// Corrupt the content -> standard validation must fail on checksums.
	if err := os.WriteFile(filepath.Join(b.BundlePath, "a.yaml"), []byte("tampered"), 0600); err != nil {
		t.Fatalf("tamper: %v", err)
	}
	if _, err := newValidator().Validate(b, StandardValidation); err == nil {
		t.Fatal("standard validation must fail on a checksum mismatch")
	}
}

func TestValidate_StrictChecksCompatibility(t *testing.T) {
	b := buildValidatableBundle(t) // MinVersion 1.0.0, current core is 1.0.0
	res, err := newValidator().Validate(b, StrictValidation)
	if err != nil {
		t.Fatalf("Validate(strict): %v", err)
	}
	if !res.Valid {
		t.Fatalf("strict validation should pass, errors: %v", res.Errors)
	}
}

func TestValidate_InvalidManifestShortCircuits(t *testing.T) {
	b := &Bundle{Manifest: BundleManifest{}, BundlePath: t.TempDir()}
	if _, err := newValidator().Validate(b, StandardValidation); err == nil {
		t.Fatal("invalid manifest should short-circuit with an error")
	}
}

func TestValidateCompatibility(t *testing.T) {
	v := newValidator().(*DefaultBundleValidator)

	core, err := version.ParseVersion("1.5.0")
	if err != nil {
		t.Fatalf("ParseVersion: %v", err)
	}
	current := map[string]*version.SemVersion{"core": &core}

	// Bundle requiring <= current core: compatible.
	ok := &Bundle{Manifest: BundleManifest{Compatibility: Compatibility{MinVersion: "1.0.0"}}}
	res, err := v.ValidateCompatibility(ok, current)
	if err != nil || !res.Valid {
		t.Fatalf("expected compatible, got valid=%v err=%v", res.Valid, err)
	}

	// Bundle requiring a newer core than installed: incompatible. Here an error
	// IS expected (compatibility failure), so assert it alongside !Valid.
	tooNew := &Bundle{Manifest: BundleManifest{Compatibility: Compatibility{MinVersion: "9.0.0"}}}
	res, err = v.ValidateCompatibility(tooNew, current)
	if err == nil {
		t.Fatal("incompatible bundle should return an error")
	}
	if res.Valid {
		t.Fatal("bundle requiring a newer core must be incompatible")
	}
}
