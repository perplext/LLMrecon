package bundle

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCompressionFactory_Handlers(t *testing.T) {
	f := NewCompressionFactory()
	for _, ct := range []CompressionType{CompressionGzip, CompressionZstd, CompressionNone} {
		if _, err := f.GetHandler(ct); err != nil {
			t.Errorf("GetHandler(%v): %v", ct, err)
		}
	}
	if _, err := f.GetHandler(CompressionType("brotli")); err == nil {
		t.Error("unknown compression type must error")
	}
}

func TestCompressionHandlers_RoundTrip(t *testing.T) {
	payload := bytes.Repeat([]byte("the quick brown fox jumps. "), 64)

	f := NewCompressionFactory()
	for _, ct := range []CompressionType{CompressionGzip, CompressionZstd, CompressionNone} {
		h, err := f.GetHandler(ct)
		if err != nil {
			t.Fatalf("GetHandler(%v): %v", ct, err)
		}

		var compressed bytes.Buffer
		if err := h.Compress(bytes.NewReader(payload), &compressed); err != nil {
			t.Fatalf("%v Compress: %v", ct, err)
		}

		var out bytes.Buffer
		if err := h.Decompress(bytes.NewReader(compressed.Bytes()), &out); err != nil {
			t.Fatalf("%v Decompress: %v", ct, err)
		}
		if !bytes.Equal(out.Bytes(), payload) {
			t.Fatalf("%v round-trip mismatch", ct)
		}
	}
}

func TestEncryptionFactory_Handlers(t *testing.T) {
	f := NewEncryptionFactory()
	for _, alg := range []string{"aes-256-gcm", "chacha20-poly1305"} {
		if _, err := f.GetHandler(alg); err != nil {
			t.Errorf("GetHandler(%q): %v", alg, err)
		}
	}
	if _, err := f.GetHandler("rot13"); err == nil {
		t.Error("unknown encryption algorithm must error")
	}
}

func TestEncryptionHandlers_RoundTripAndWrongPassword(t *testing.T) {
	plaintext := []byte("air-gapped bundle secret payload")
	const pw = "correct-horse"

	f := NewEncryptionFactory()
	for _, alg := range []string{"aes-256-gcm", "chacha20-poly1305"} {
		h, err := f.GetHandler(alg)
		if err != nil {
			t.Fatalf("GetHandler(%q): %v", alg, err)
		}

		ciphertext, err := h.Encrypt(plaintext, pw)
		if err != nil {
			t.Fatalf("%s Encrypt: %v", alg, err)
		}
		if bytes.Contains(ciphertext, plaintext) {
			t.Fatalf("%s ciphertext leaks plaintext", alg)
		}

		got, err := h.Decrypt(ciphertext, pw)
		if err != nil {
			t.Fatalf("%s Decrypt: %v", alg, err)
		}
		if !bytes.Equal(got, plaintext) {
			t.Fatalf("%s round-trip mismatch", alg)
		}

		// Wrong password must fail authentication, not return garbage.
		if _, err := h.Decrypt(ciphertext, "wrong-password"); err == nil {
			t.Fatalf("%s decrypt with wrong password must error", alg)
		}
	}
}

func TestDecompressBundle_MissingArchive(t *testing.T) {
	c := NewBundleCompressor()
	err := c.DecompressBundle(filepath.Join(t.TempDir(), "nope.zip"), t.TempDir(), DecompressOptions{})
	if err == nil {
		t.Fatal("DecompressBundle on a missing archive must error")
	}
}

// TestDecompressBundle_ExtractionStub documents the current honest contract: the
// zip/tar extraction paths are not yet implemented (the real, zip-slip-protected
// extraction lives in bundle.ExtractBundle). If extraction is implemented later,
// this test should be replaced with a real round-trip + traversal-rejection case.
func TestDecompressBundle_ExtractionStub(t *testing.T) {
	archive := filepath.Join(t.TempDir(), "b.zip")
	if err := os.WriteFile(archive, []byte("PK\x03\x04 not-a-real-zip"), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := c_decompress(t, archive); err == nil {
		t.Fatal("DecompressBundle zip extraction is a stub and must currently error")
	}
}

func c_decompress(t *testing.T, archive string) error {
	t.Helper()
	return NewBundleCompressor().DecompressBundle(archive, t.TempDir(), DecompressOptions{})
}
