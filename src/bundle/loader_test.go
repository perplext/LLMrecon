package bundle

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadBundle_RoundTrip(t *testing.T) {
	dir := writeManifestDir(t, sampleManifest())
	b, err := LoadBundle(dir)
	if err != nil {
		t.Fatalf("LoadBundle: %v", err)
	}
	if b.Manifest.Name != "test-bundle" || b.BundlePath != dir {
		t.Fatalf("unexpected bundle: %+v", b)
	}
}

func TestLoadBundle_NotADirectory(t *testing.T) {
	f := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(f, []byte("x"), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := LoadBundle(f); err == nil {
		t.Fatal("LoadBundle on a file (not dir) must error")
	}
}

func TestLoadBundle_Missing(t *testing.T) {
	if _, err := LoadBundle(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("LoadBundle on missing path must error")
	}
}

func TestLoadBundle_MalformedManifest(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte("nope"), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := LoadBundle(dir); err == nil {
		t.Fatal("LoadBundle with malformed manifest must error")
	}
}

func TestSaveBundle_RoundTrip(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "saved")
	b := &Bundle{Manifest: sampleManifest(), BundlePath: dir}
	if err := SaveBundle(b); err != nil {
		t.Fatalf("SaveBundle: %v", err)
	}
	reloaded, err := LoadBundle(dir)
	if err != nil {
		t.Fatalf("LoadBundle after save: %v", err)
	}
	if reloaded.Manifest.Name != b.Manifest.Name {
		t.Fatalf("round-trip name mismatch: %q", reloaded.Manifest.Name)
	}
}

func TestCreateEmptyBundle(t *testing.T) {
	b, err := CreateEmptyBundle("/tmp/b", "1.0", "bundle-id", MixedBundleType, "name", "desc", "1.0.0")
	if err != nil {
		t.Fatalf("CreateEmptyBundle: %v", err)
	}
	if b.Manifest.BundleID != "bundle-id" || b.Manifest.Name != "name" {
		t.Fatalf("unexpected manifest: %+v", b.Manifest)
	}

	// Each required field, when missing, must error.
	bad := []struct {
		name                                                  string
		path, id, bundleName, version string
	}{
		{"no path", "", "id", "n", "1.0"},
		{"no id", "/tmp/b", "", "n", "1.0"},
		{"no name", "/tmp/b", "id", "", "1.0"},
		{"no version", "/tmp/b", "id", "n", ""},
	}
	for _, c := range bad {
		if _, err := CreateEmptyBundle(c.path, "1.0", c.id, MixedBundleType, c.bundleName, "d", c.version); err == nil {
			t.Errorf("%s: expected error", c.name)
		}
	}
}
