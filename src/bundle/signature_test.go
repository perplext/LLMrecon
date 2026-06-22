package bundle

import (
	"crypto/ed25519"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildSignableBundle creates an on-disk bundle directory with one content file
// and a manifest referencing it, ready for checksum/signature operations.
func buildSignableBundle(t *testing.T) *Bundle {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("hello world"), 0600); err != nil {
		t.Fatalf("write content: %v", err)
	}
	m := CreateBundleManifest("signable", "d", "1.0.0", MixedBundleType, Author{})
	m.AddContentItem("a.txt", ResourceContentType, "id-a", "1.0", "")
	return &Bundle{Manifest: m, BundlePath: dir}
}

func TestGenerateKeyPair(t *testing.T) {
	priv, pub, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	if len(priv) != ed25519.PrivateKeySize {
		t.Fatalf("private key size = %d, want %d", len(priv), ed25519.PrivateKeySize)
	}
	if len(pub) != ed25519.PublicKeySize {
		t.Fatalf("public key size = %d, want %d", len(pub), ed25519.PublicKeySize)
	}
}

func TestNewSignatureManager_BadPublicKey(t *testing.T) {
	if _, err := NewSignatureManager(Ed25519Algorithm, nil, []byte("too-short")); err == nil {
		t.Fatal("NewSignatureManager with wrong-size public key must error")
	}
}

func TestCalculateFileChecksum(t *testing.T) {
	f := filepath.Join(t.TempDir(), "f.txt")
	if err := os.WriteFile(f, []byte("data"), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}
	sum, err := CalculateFileChecksum(f)
	if err != nil {
		t.Fatalf("CalculateFileChecksum: %v", err)
	}
	if !strings.HasPrefix(sum, "sha256:") {
		t.Fatalf("checksum missing sha256 prefix: %q", sum)
	}
	// Deterministic.
	again, _ := CalculateFileChecksum(f)
	if again != sum {
		t.Fatal("checksum should be deterministic")
	}

	if _, err := CalculateFileChecksum(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("checksum of missing file must error")
	}
}

func TestBundleChecksums_RoundTripAndTamper(t *testing.T) {
	b := buildSignableBundle(t)

	if err := UpdateBundleChecksums(b); err != nil {
		t.Fatalf("UpdateBundleChecksums: %v", err)
	}
	res, err := VerifyBundleChecksums(b)
	if err != nil {
		t.Fatalf("VerifyBundleChecksums: %v", err)
	}
	if !res.Valid {
		t.Fatalf("checksums should verify, errors: %v", res.Errors)
	}

	// Tamper the content after checksums were computed.
	if err := os.WriteFile(filepath.Join(b.BundlePath, "a.txt"), []byte("TAMPERED"), 0600); err != nil {
		t.Fatalf("tamper: %v", err)
	}
	res, err = VerifyBundleChecksums(b)
	if err != nil {
		t.Fatalf("VerifyBundleChecksums (tampered): %v", err)
	}
	if res.Valid {
		t.Fatal("tampered content must fail checksum verification")
	}
}

// TestSignVerifyBundle_RoundTrip is the core #230 case: a signed bundle verifies,
// and tampering with either the content or the manifest is detected.
func TestSignVerifyBundle_RoundTrip(t *testing.T) {
	priv, pub, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}

	b := buildSignableBundle(t)
	if err := SignBundle(b, priv); err != nil {
		t.Fatalf("SignBundle: %v", err)
	}
	if b.Manifest.Signature == "" {
		t.Fatal("SignBundle should populate the manifest signature")
	}

	res, err := VerifyBundle(b, pub)
	if err != nil {
		t.Fatalf("VerifyBundle: %v", err)
	}
	if !res.Valid {
		t.Fatalf("signed bundle should verify, errors: %v", res.Errors)
	}
}

func TestVerifyBundle_TamperedContent(t *testing.T) {
	priv, pub, _ := GenerateKeyPair()
	b := buildSignableBundle(t)
	if err := SignBundle(b, priv); err != nil {
		t.Fatalf("SignBundle: %v", err)
	}

	// Corrupt the content file after signing -> checksum mismatch.
	if err := os.WriteFile(filepath.Join(b.BundlePath, "a.txt"), []byte("corrupted"), 0600); err != nil {
		t.Fatalf("tamper: %v", err)
	}
	res, _ := VerifyBundle(b, pub)
	if res.Valid {
		t.Fatal("tampered content must fail bundle verification")
	}
}

func TestVerifyBundle_TamperedManifest(t *testing.T) {
	priv, pub, _ := GenerateKeyPair()
	b := buildSignableBundle(t)
	if err := SignBundle(b, priv); err != nil {
		t.Fatalf("SignBundle: %v", err)
	}

	// Mutate a signed manifest field -> signature no longer matches.
	b.Manifest.Name = "evil-rename"
	res, _ := VerifyBundle(b, pub)
	if res.Valid {
		t.Fatal("tampered manifest must fail signature verification")
	}
}

func TestVerifyBundle_WrongKey(t *testing.T) {
	priv, _, _ := GenerateKeyPair()
	_, otherPub, _ := GenerateKeyPair()

	b := buildSignableBundle(t)
	if err := SignBundle(b, priv); err != nil {
		t.Fatalf("SignBundle: %v", err)
	}
	res, _ := VerifyBundle(b, otherPub)
	if res.Valid {
		t.Fatal("verification with the wrong public key must fail")
	}
}
