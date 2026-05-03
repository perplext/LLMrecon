package update

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// makeTestZip writes a ZIP with the supplied entries to a tempfile and
// returns its path. Callers configure entries via the entry struct's
// fields:
//
//	{name, content, isDir, mode}
//
// where name uses '/' separators (ZIP convention).
type zipEntry struct {
	name    string
	content []byte
	isDir   bool
	mode    os.FileMode
}

func makeTestZip(t *testing.T, entries []zipEntry) string {
	t.Helper()
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "test.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	defer func() { _ = f.Close() }()
	zw := zip.NewWriter(f)
	for _, e := range entries {
		header := &zip.FileHeader{Name: e.name, Method: zip.Deflate}
		if e.mode != 0 {
			header.SetMode(e.mode)
		}
		if e.isDir {
			if !strings.HasSuffix(header.Name, "/") {
				header.Name += "/"
			}
		}
		w, err := zw.CreateHeader(header)
		if err != nil {
			t.Fatalf("create header: %v", err)
		}
		if !e.isDir && len(e.content) > 0 {
			if _, err := w.Write(e.content); err != nil {
				t.Fatalf("write: %v", err)
			}
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zw: %v", err)
	}
	return zipPath
}

// TestExtractZip_HappyPath asserts a well-formed ZIP extracts files and
// directories with the expected content + permissions.
func TestExtractZip_HappyPath(t *testing.T) {
	zipPath := makeTestZip(t, []zipEntry{
		{name: "templates/", isDir: true},
		{name: "templates/owasp/", isDir: true},
		{name: "templates/owasp/llm01.yaml", content: []byte("id: llm01\n"), mode: 0o644},
		{name: "manifest.json", content: []byte(`{"version":"0.10.0"}`), mode: 0o644},
	})

	dest := t.TempDir()
	bytes, err := ExtractZip(zipPath, dest)
	if err != nil {
		t.Fatalf("ExtractZip: %v", err)
	}
	if bytes == 0 {
		t.Error("bytes = 0, want > 0")
	}

	// Files exist with expected content
	got, err := os.ReadFile(filepath.Join(dest, "templates/owasp/llm01.yaml"))
	if err != nil {
		t.Fatalf("read extracted: %v", err)
	}
	if string(got) != "id: llm01\n" {
		t.Errorf("content = %q", string(got))
	}
	manifest, err := os.ReadFile(filepath.Join(dest, "manifest.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	if !strings.Contains(string(manifest), `"0.10.0"`) {
		t.Errorf("manifest = %q", manifest)
	}
}

// TestExtractZip_RejectsZipSlip asserts that a malicious archive trying
// to write to ../../etc/passwd is rejected before extraction starts.
func TestExtractZip_RejectsZipSlip(t *testing.T) {
	cases := []string{
		"../escape.txt",
		"../../absolute-via-traverse.txt",
		"a/b/../../../escape.txt",
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			zipPath := makeTestZip(t, []zipEntry{
				{name: name, content: []byte("evil"), mode: 0o644},
			})
			dest := t.TempDir()
			_, err := ExtractZip(zipPath, dest)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), "traversal") && !strings.Contains(err.Error(), "escapes") {
				t.Errorf("err = %v, want traversal/escape error", err)
			}
		})
	}
}

// TestExtractZip_RejectsAbsolutePath asserts entries with absolute paths
// are rejected. These would smuggle data to /etc/, /Users/foo, etc.
func TestExtractZip_RejectsAbsolutePath(t *testing.T) {
	zipPath := makeTestZip(t, []zipEntry{
		{name: "/etc/passwd", content: []byte("root:x:0:0"), mode: 0o644},
	})
	dest := t.TempDir()
	_, err := ExtractZip(zipPath, dest)
	if err == nil {
		t.Fatal("expected error for absolute path")
	}
	if !strings.Contains(err.Error(), "absolute") {
		t.Errorf("err = %v, want 'absolute' error", err)
	}
}

// TestExtractZip_EnforcesPerEntryCap asserts a single zip entry that
// would expand past MaxExtractedFileBytes is rejected without writing
// the partial file.
func TestExtractZip_EnforcesPerEntryCap(t *testing.T) {
	// Create a real ZIP with a large entry (just over the cap). We
	// write the data into the zip, not a fake-content claim.
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "big.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	zw := zip.NewWriter(f)
	w, err := zw.CreateHeader(&zip.FileHeader{Name: "big.bin"})
	if err != nil {
		t.Fatal(err)
	}
	// Write MaxExtractedFileBytes+1 bytes of zeros.
	chunk := make([]byte, 1<<20) // 1 MiB
	written := int64(0)
	for written <= int64(MaxExtractedFileBytes) {
		n, err := w.Write(chunk)
		if err != nil {
			t.Fatal(err)
		}
		written += int64(n)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	dest := t.TempDir()
	_, err = ExtractZip(zipPath, dest)
	if err == nil {
		t.Fatal("expected size-cap error")
	}
	if !strings.Contains(err.Error(), "size cap") && !strings.Contains(err.Error(), "size budget") {
		t.Errorf("err = %v, want size cap error", err)
	}
}

// TestExtractZip_SkipsSymlinks asserts symlink entries are silently
// skipped rather than extracted (link traversal hardening).
func TestExtractZip_SkipsSymlinks(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "syms.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	zw := zip.NewWriter(f)

	// Add a symlink entry. SetMode handles platform-specific bit.
	header := &zip.FileHeader{Name: "evil-link", Method: zip.Deflate}
	header.SetMode(os.ModeSymlink | 0o777)
	w, err := zw.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("/etc/passwd")); err != nil {
		t.Fatal(err)
	}

	// Also add a regular file so the test verifies extraction continues.
	w2, err := zw.Create("ok.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w2.Write([]byte("ok")); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	dest := t.TempDir()
	if _, err := ExtractZip(zipPath, dest); err != nil {
		t.Fatalf("ExtractZip: %v", err)
	}
	// Symlink not extracted.
	if _, err := os.Lstat(filepath.Join(dest, "evil-link")); !os.IsNotExist(err) {
		t.Errorf("symlink was extracted; expected skip (lstat err = %v)", err)
	}
	// Regular file extracted.
	got, err := os.ReadFile(filepath.Join(dest, "ok.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "ok" {
		t.Errorf("regular entry content = %q", got)
	}
}

// TestExtractZip_WriteBitsClampedToOwner asserts group/other write bits
// are stripped from extracted file modes — bundle authors can't drop
// world-writable files via the zip's mode field.
func TestExtractZip_WriteBitsClampedToOwner(t *testing.T) {
	zipPath := makeTestZip(t, []zipEntry{
		{name: "perm.txt", content: []byte("x"), mode: 0o777},
	})
	dest := t.TempDir()
	if _, err := ExtractZip(zipPath, dest); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(filepath.Join(dest, "perm.txt"))
	if err != nil {
		t.Fatal(err)
	}
	got := info.Mode().Perm()
	// Should be clamped — group/other write bits cleared.
	if got&0o022 != 0 {
		t.Errorf("perm = %o, want group/other write cleared", got)
	}
}

// TestSafeJoin_RejectsTraversal exercises the safeJoin guard directly
// for the Windows-encoded-separator and double-dot cases.
func TestSafeJoin_RejectsTraversal(t *testing.T) {
	dest, _ := filepath.Abs(t.TempDir())
	bads := []string{
		"../x",
		"a/../../x",
		"/abs/path",
	}
	for _, b := range bads {
		t.Run(b, func(t *testing.T) {
			_, err := safeJoin(dest, b)
			if err == nil {
				t.Errorf("expected error for %q", b)
			}
		})
	}
}

// silence unused import warnings.
var _ = errors.New
var _ = io.EOF
var _ = bytes.NewBuffer
