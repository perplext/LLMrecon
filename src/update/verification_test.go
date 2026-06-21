package update

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

// writeManifestStub creates the manifest.yaml file VerifyPackage stats for
// existence. Its contents are irrelevant — VerifyPackage reads pkg.Manifest
// (the in-memory struct), not this file.
func writeManifestStub(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "manifest.yaml"), []byte("schema_version: 1\n"), 0o600); err != nil {
		t.Fatalf("write manifest stub: %v", err)
	}
}

func newVerifier() *IntegrityVerifier { return NewIntegrityVerifier(io.Discard) }

// TestVerifyPackage_NoChecksumsRefuses is the core honesty check (#224): a
// manifest that declares no component checksums must NOT report success.
func TestVerifyPackage_NoChecksumsRefuses(t *testing.T) {
	dir := t.TempDir()
	writeManifestStub(t, dir)
	pkg := &UpdatePackage{PackagePath: dir, Manifest: PackageManifest{}}

	res, err := newVerifier().VerifyPackage(pkg)
	if err == nil {
		t.Fatalf("expected error when no checksums are declared")
	}
	if res.Success {
		t.Errorf("VerifyPackage must not report Success=true without verifying any checksum")
	}
}

func TestVerifyPackage_TemplatesChecksumMatch(t *testing.T) {
	dir := t.TempDir()
	writeManifestStub(t, dir)

	// Lay out a templates/ payload and compute its expected directory hash.
	tmplDir := filepath.Join(dir, "templates")
	if err := os.MkdirAll(tmplDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmplDir, "a.yaml"), []byte("payload-a"), 0o600); err != nil {
		t.Fatal(err)
	}
	want, err := calculateDirectoryHash(tmplDir)
	if err != nil {
		t.Fatal(err)
	}

	pkg := &UpdatePackage{PackagePath: dir, Manifest: PackageManifest{
		Components: Components{Templates: TemplatesComponentInfo{Checksum: want}},
	}}

	res, err := newVerifier().VerifyPackage(pkg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Success {
		t.Errorf("expected success on matching templates checksum; msg=%q", res.Message)
	}
	if res.Details["verified_components"] != 1 {
		t.Errorf("verified_components = %v, want 1", res.Details["verified_components"])
	}
}

func TestVerifyPackage_TemplatesChecksumMismatch(t *testing.T) {
	dir := t.TempDir()
	writeManifestStub(t, dir)
	tmplDir := filepath.Join(dir, "templates")
	if err := os.MkdirAll(tmplDir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmplDir, "a.yaml"), []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	pkg := &UpdatePackage{PackagePath: dir, Manifest: PackageManifest{
		Components: Components{Templates: TemplatesComponentInfo{Checksum: "deadbeef"}},
	}}

	res, err := newVerifier().VerifyPackage(pkg)
	if err == nil || res.Success {
		t.Errorf("expected failure on checksum mismatch; got success=%v err=%v", res.Success, err)
	}
}

func TestVerifyPackage_ModuleChecksumMatch(t *testing.T) {
	dir := t.TempDir()
	writeManifestStub(t, dir)
	modDir := filepath.Join(dir, "modules")
	if err := os.MkdirAll(modDir, 0o750); err != nil {
		t.Fatal(err)
	}
	content := []byte("module-payload")
	if err := os.WriteFile(filepath.Join(modDir, "mod1"), content, 0o600); err != nil {
		t.Fatal(err)
	}
	pkg := &UpdatePackage{PackagePath: dir, Manifest: PackageManifest{
		Components: Components{Modules: []ModuleComponentInfo{{ID: "mod1", Checksum: calculateFileHash(content)}}},
	}}

	res, err := newVerifier().VerifyPackage(pkg)
	if err != nil || !res.Success {
		t.Errorf("expected success on matching module checksum; success=%v err=%v", res.Success, err)
	}
}

func TestVerifyPackage_DeclaredPayloadMissing(t *testing.T) {
	dir := t.TempDir()
	writeManifestStub(t, dir)
	// Declares a module checksum but the modules/ payload is absent.
	pkg := &UpdatePackage{PackagePath: dir, Manifest: PackageManifest{
		Components: Components{Modules: []ModuleComponentInfo{{ID: "ghost", Checksum: "abc123"}}},
	}}

	res, err := newVerifier().VerifyPackage(pkg)
	if err == nil || res.Success {
		t.Errorf("expected failure when a declared payload is missing; success=%v err=%v", res.Success, err)
	}
}

func TestVerifyPackage_SignatureRefused(t *testing.T) {
	dir := t.TempDir()
	writeManifestStub(t, dir)
	pkg := &UpdatePackage{PackagePath: dir, Manifest: PackageManifest{Signature: "sig-data"}}

	res, err := newVerifier().VerifyPackage(pkg)
	if err == nil || res.Success {
		t.Errorf("expected refusal on a signed bundle (signature verification unimplemented); success=%v err=%v", res.Success, err)
	}
}

func TestVerifyPackage_MissingManifest(t *testing.T) {
	dir := t.TempDir() // no manifest.yaml written
	pkg := &UpdatePackage{PackagePath: dir, Manifest: PackageManifest{}}

	res, err := newVerifier().VerifyPackage(pkg)
	if err == nil || res.Success {
		t.Errorf("expected failure when manifest is absent; success=%v err=%v", res.Success, err)
	}
}
