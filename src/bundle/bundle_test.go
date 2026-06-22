package bundle

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeManifestDir creates a directory-style bundle with the given manifest
// written to manifest.json, and returns its path.
func writeManifestDir(t *testing.T, manifest BundleManifest) string {
	t.Helper()
	dir := t.TempDir()
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return dir
}

func sampleManifest() BundleManifest {
	m := CreateBundleManifest("test-bundle", "a test bundle", "1.0.0", TemplateBundleType, Author{})
	return m
}

// writeZip writes a zip file with the given name->content entries, using the
// names verbatim (so a "../escape" entry can be crafted for the zip-slip test).
func writeZip(t *testing.T, path string, entries map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip create entry %q: %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("zip write entry %q: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
}

func TestOpenBundle_RoundTrip(t *testing.T) {
	dir := writeManifestDir(t, sampleManifest())
	b, err := OpenBundle(dir)
	if err != nil {
		t.Fatalf("OpenBundle: %v", err)
	}
	if b.Manifest.Name != "test-bundle" {
		t.Fatalf("manifest name = %q, want test-bundle", b.Manifest.Name)
	}
	if b.IsVerified {
		t.Fatal("freshly opened bundle should not be marked verified")
	}
}

func TestOpenBundle_MissingPath(t *testing.T) {
	if _, err := OpenBundle(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("OpenBundle on missing path must error")
	}
}

func TestOpenBundle_MalformedManifest(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte("{not valid json"), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := OpenBundle(dir); err == nil {
		t.Fatal("OpenBundle with malformed manifest must error")
	}
}

func TestExtractBundle_RoundTrip(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "b.zip")
	writeZip(t, zipPath, map[string]string{
		"manifest.json":      `{"name":"x"}`,
		"templates/a.yaml":   "content-a",
		"templates/sub/b.txt": "content-b",
	})

	out := filepath.Join(t.TempDir(), "out")
	if err := ExtractBundle(zipPath, out); err != nil {
		t.Fatalf("ExtractBundle: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(out, "templates", "a.yaml"))
	if err != nil {
		t.Fatalf("extracted file missing: %v", err)
	}
	if string(got) != "content-a" {
		t.Fatalf("extracted content = %q", got)
	}
}

// TestExtractBundle_ZipSlip is the core security case from #230 / #174: an entry
// whose name escapes the output directory must be rejected, not written outside.
func TestExtractBundle_ZipSlip(t *testing.T) {
	zipPath := filepath.Join(t.TempDir(), "evil.zip")
	writeZip(t, zipPath, map[string]string{
		"../escaped.txt": "pwned",
	})

	outParent := t.TempDir()
	out := filepath.Join(outParent, "out")
	err := ExtractBundle(zipPath, out)
	if err == nil {
		t.Fatal("zip-slip entry must be rejected")
	}
	if !strings.Contains(err.Error(), "illegal file path") {
		t.Fatalf("expected illegal-file-path error, got: %v", err)
	}
	// The escaped file must NOT have been written to the parent directory.
	if _, statErr := os.Stat(filepath.Join(outParent, "escaped.txt")); statErr == nil {
		t.Fatal("zip-slip wrote a file outside the output directory")
	}
}

func TestIsWithinDir(t *testing.T) {
	base := t.TempDir()
	cases := []struct {
		path string
		want bool
	}{
		{filepath.Join(base, "a", "b.txt"), true},
		{filepath.Join(base, "x"), true},
		{filepath.Join(base, "..", "outside.txt"), false},
		{filepath.Join(base, "a", "..", "..", "etc"), false},
	}
	for _, c := range cases {
		if got := isWithinDir(base, c.path); got != c.want {
			t.Errorf("isWithinDir(%q, %q) = %v, want %v", base, c.path, got, c.want)
		}
	}
}

func TestContentAccessors(t *testing.T) {
	m := sampleManifest()
	m.AddContentItem("templates/a.yaml", TemplateContentType, "id-a", "1.0", "first")
	m.AddContentItem("templates/b.yaml", TemplateContentType, "id-b", "1.0", "second")
	m.AddContentItem("config/c.json", ConfigContentType, "id-c", "1.0", "third")

	b := &Bundle{Manifest: m, BundlePath: "/tmp/bundle"}

	if p := b.GetContentPath("id-b"); p != filepath.Join("/tmp/bundle", "templates/b.yaml") {
		t.Fatalf("GetContentPath = %q", p)
	}
	if b.GetContentPath("missing") != "" {
		t.Fatal("GetContentPath for missing id should be empty")
	}

	item := b.GetContentItem("id-c")
	if item == nil || item.Type != ConfigContentType {
		t.Fatalf("GetContentItem(id-c) = %+v", item)
	}
	if b.GetContentItem("missing") != nil {
		t.Fatal("GetContentItem for missing id should be nil")
	}

	templates := b.GetContentItemsByType(TemplateContentType)
	if len(templates) != 2 {
		t.Fatalf("expected 2 template items, got %d", len(templates))
	}
}

func TestCreateBundleManifest(t *testing.T) {
	m := CreateBundleManifest("n", "d", "2.0.0", ModuleBundleType, Author{})
	if m.Name != "n" || m.BundleType != ModuleBundleType || m.Version != "2.0.0" {
		t.Fatalf("unexpected manifest: %+v", m)
	}
	if m.SchemaVersion == "" || m.BundleID == "" {
		t.Fatal("manifest should populate schema version and bundle ID")
	}
	if m.Checksums.Content == nil {
		t.Fatal("manifest should initialize the content checksum map")
	}
}
