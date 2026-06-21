package keystore

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"
)

func storeWith(entries ...*KeyEntry) *Keystore {
	ks := &Keystore{keys: make(map[string]*KeyEntry)}
	for _, e := range entries {
		ks.keys[e.ID] = e
	}
	return ks
}

func testRSAEntry(t *testing.T, id string) (*KeyEntry, *rsa.PrivateKey) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA: %v", err)
	}
	return &KeyEntry{
		ID: id, Type: KeyTypeRSA, Algorithm: "RSA-2048", Key: priv,
		Metadata: map[string]string{"env": "test"},
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}, priv
}

func testAESEntry(id string) *KeyEntry {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}
	return &KeyEntry{
		ID: id, Type: KeyTypeAES, Algorithm: "AES-256", Key: key,
		Metadata: map[string]string{}, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
}

func TestExport_Guards(t *testing.T) {
	ke := NewKeyExporter(storeWith(testAESEntry("a1")))

	if _, err := ke.ExportKey("", KeyExportOptions{Format: FormatBinary}); err == nil {
		t.Error("empty key id must error")
	}
	if _, err := ke.ExportKey("missing", KeyExportOptions{Format: FormatBinary}); err == nil {
		t.Error("missing key must error")
	}
	if _, err := ke.ExportKey("a1", KeyExportOptions{Format: "tar"}); err == nil {
		t.Error("unsupported format must error")
	}
}

func TestExportPEM_TypeMismatch(t *testing.T) {
	// Declared RSA but holds a non-RSA value -> exportToPEM must reject it
	// rather than panic on the type assertion.
	bad := &KeyEntry{ID: "x", Type: KeyTypeRSA, Key: []byte("not-an-rsa-key")}
	ke := NewKeyExporter(storeWith(bad))
	if _, err := ke.ExportKey("x", KeyExportOptions{Format: FormatPEM}); err == nil {
		t.Fatal("PEM export of mistyped RSA key must error")
	}
}

func TestRoundTrip_RSA_PEM_Plain(t *testing.T) {
	entry, orig := testRSAEntry(t, "rsa-pem")
	ke := NewKeyExporter(storeWith(entry))

	exported, err := ke.ExportKey("rsa-pem", KeyExportOptions{Format: FormatPEM, Metadata: true})
	if err != nil {
		t.Fatalf("ExportKey: %v", err)
	}
	if exported.Fingerprint == "" {
		t.Error("export should produce a fingerprint")
	}

	// Import into a fresh keystore.
	ki := NewKeyImporter(storeWith())
	if err := ki.ImportKey(exported, KeyImportOptions{Format: FormatPEM, ValidateKey: true}); err != nil {
		t.Fatalf("ImportKey: %v", err)
	}
	got := ki.keystore.keys["rsa-pem"]
	imported, ok := got.Key.(*rsa.PrivateKey)
	if !ok {
		t.Fatalf("imported key is not *rsa.PrivateKey: %T", got.Key)
	}
	if imported.N.Cmp(orig.N) != 0 {
		t.Fatal("imported RSA modulus differs from original")
	}
}

func TestRoundTrip_RSA_PEM_Encrypted(t *testing.T) {
	entry, _ := testRSAEntry(t, "rsa-enc")
	ke := NewKeyExporter(storeWith(entry))

	exported, err := ke.ExportKey("rsa-enc", KeyExportOptions{
		Format: FormatPEM, Encrypted: true, Password: "s3cr3t-pass",
	})
	if err != nil {
		t.Fatalf("ExportKey encrypted: %v", err)
	}

	ki := NewKeyImporter(storeWith())

	// Wrong password must fail.
	if err := ki.ImportKey(exported, KeyImportOptions{Format: FormatPEM, Password: "wrong"}); err == nil {
		t.Fatal("import with wrong password must fail")
	}
	// Missing password for an encrypted block must fail.
	if err := ki.ImportKey(exported, KeyImportOptions{Format: FormatPEM}); err == nil {
		t.Fatal("import of encrypted PEM without password must fail")
	}
	// Correct password succeeds.
	if err := ki.ImportKey(exported, KeyImportOptions{Format: FormatPEM, Password: "s3cr3t-pass"}); err != nil {
		t.Fatalf("import with correct password: %v", err)
	}
}

func TestRoundTrip_AES_JSON_Encrypted(t *testing.T) {
	ke := NewKeyExporter(storeWith(testAESEntry("aes-json")))
	exported, err := ke.ExportKey("aes-json", KeyExportOptions{
		Format: FormatJSON, Encrypted: true, Password: "pw-correct",
	})
	if err != nil {
		t.Fatalf("ExportKey: %v", err)
	}

	ki := NewKeyImporter(storeWith())

	// Wrong password -> GCM auth failure.
	if err := ki.ImportKey(exported, KeyImportOptions{Format: FormatJSON, Password: "pw-wrong"}); err == nil {
		t.Fatal("AES-GCM import with wrong password must fail")
	}
	if err := ki.ImportKey(exported, KeyImportOptions{Format: FormatJSON, Password: "pw-correct"}); err != nil {
		t.Fatalf("import correct password: %v", err)
	}
	got, ok := ki.keystore.keys["aes-json"].Key.([]byte)
	if !ok || len(got) != 32 {
		t.Fatalf("decrypted AES key wrong: ok=%v len=%d", ok, len(got))
	}
}

func TestRoundTrip_AES_Binary_Encrypted(t *testing.T) {
	ke := NewKeyExporter(storeWith(testAESEntry("aes-bin")))
	exported, err := ke.ExportKey("aes-bin", KeyExportOptions{
		Format: FormatBinary, Encrypted: true, Password: "binpass",
	})
	if err != nil {
		t.Fatalf("ExportKey: %v", err)
	}

	ki := NewKeyImporter(storeWith())
	if err := ki.ImportKey(exported, KeyImportOptions{Format: FormatBinary, Password: "binpass"}); err != nil {
		t.Fatalf("ImportKey: %v", err)
	}
}

func TestImport_Guards(t *testing.T) {
	ki := NewKeyImporter(storeWith(testAESEntry("exists")))

	if err := ki.ImportKey(nil, KeyImportOptions{}); err == nil {
		t.Error("nil exported key must error")
	}
	if err := ki.ImportKey(&ExportedKey{}, KeyImportOptions{Format: FormatBinary}); err == nil {
		t.Error("empty ID must error")
	}

	dup := &ExportedKey{ID: "exists", Type: KeyTypeAES, Data: "AAAA"}
	if err := ki.ImportKey(dup, KeyImportOptions{Format: FormatBinary}); err == nil {
		t.Error("importing existing id without Overwrite must error")
	}

	if err := ki.ImportKey(&ExportedKey{ID: "new", Data: "AAAA"}, KeyImportOptions{Format: "xml"}); err == nil {
		t.Error("unsupported import format must error")
	}
}

func TestValidateKey(t *testing.T) {
	ki := NewKeyImporter(storeWith())

	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	if err := ki.validateKey(priv, KeyTypeRSA); err != nil {
		t.Errorf("valid RSA key should pass: %v", err)
	}
	if err := ki.validateKey([]byte("short"), KeyTypeRSA); err == nil {
		t.Error("non-RSA value for RSA type must fail")
	}

	for _, size := range []int{16, 24, 32} {
		if err := ki.validateKey(make([]byte, size), KeyTypeAES); err != nil {
			t.Errorf("AES key of %d bytes should be valid: %v", size, err)
		}
	}
	if err := ki.validateKey(make([]byte, 20), KeyTypeAES); err == nil {
		t.Error("AES key of invalid size must fail")
	}
	if err := ki.validateKey([]byte{}, "exotic"); err == nil {
		t.Error("unsupported key type must fail validation")
	}
}
