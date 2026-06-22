package keystore

import (
	"path/filepath"
	"testing"
	"time"
)

// newTestStore creates a FileKeyStore in a temp dir with autosave on.
func newTestStore(t *testing.T) *FileKeyStore {
	t.Helper()
	dir := t.TempDir()
	ks, err := NewFileKeyStore(KeyStoreOptions{
		StoragePath: filepath.Join(dir, "keys.json"),
		Passphrase:  "correct horse battery staple",
		AutoSave:    true,
	})
	if err != nil {
		t.Fatalf("NewFileKeyStore: %v", err)
	}
	t.Cleanup(func() { _ = ks.Close() })
	return ks
}

func sampleKey(name string) *Key {
	return &Key{
		Metadata: KeyMetadata{
			Name:            name,
			Type:            SymmetricKey,
			Usage:           EncryptionKey,
			ProtectionLevel: SoftwareProtection,
			Algorithm:       "AES-256",
			Tags:            []string{"env:test"},
		},
		Material: KeyMaterial{
			Private: []byte("super-secret-material"),
			Public:  []byte("public-bytes"),
			Format:  "RAW",
		},
	}
}

func TestNewFileKeyStore_CreatesStore(t *testing.T) {
	ks := newTestStore(t)
	keys, err := ks.ListKeys()
	if err != nil {
		t.Fatalf("ListKeys: %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("new store should be empty, got %d keys", len(keys))
	}
}

func TestStoreAndGetKey_RoundTrip(t *testing.T) {
	ks := newTestStore(t)
	k := sampleKey("api-signing")

	if err := ks.StoreKey(k); err != nil {
		t.Fatalf("StoreKey: %v", err)
	}
	if k.Metadata.ID == "" {
		t.Fatal("StoreKey should populate a generated ID")
	}
	if k.Metadata.CreatedAt.IsZero() || k.Metadata.UpdatedAt.IsZero() {
		t.Fatal("StoreKey should set timestamps")
	}
	if k.Metadata.Fingerprint == "" {
		t.Fatal("StoreKey should compute a fingerprint when public material is present")
	}

	got, err := ks.GetKey(k.Metadata.ID)
	if err != nil {
		t.Fatalf("GetKey: %v", err)
	}
	if string(got.Material.Private) != "super-secret-material" {
		t.Fatalf("private material not preserved: %q", got.Material.Private)
	}
	if got.Metadata.LastUsedAt.IsZero() {
		t.Fatal("GetKey should stamp LastUsedAt")
	}
}

func TestStoreKey_NilKey(t *testing.T) {
	ks := newTestStore(t)
	if err := ks.StoreKey(nil); err == nil {
		t.Fatal("StoreKey(nil) must return an error")
	}
}

func TestStoreKey_HSMRequestedWithoutHSM(t *testing.T) {
	ks := newTestStore(t)
	k := sampleKey("hsm-key")
	k.Metadata.ProtectionLevel = HSMProtection
	if err := ks.StoreKey(k); err == nil {
		t.Fatal("HSM protection without configured HSM must error")
	}
}

func TestGetKey_NotFound(t *testing.T) {
	ks := newTestStore(t)
	if _, err := ks.GetKey("does-not-exist"); err == nil {
		t.Fatal("GetKey on missing id must return an error")
	}
}

func TestGetKeyMetadata(t *testing.T) {
	ks := newTestStore(t)
	k := sampleKey("meta")
	if err := ks.StoreKey(k); err != nil {
		t.Fatalf("StoreKey: %v", err)
	}

	md, err := ks.GetKeyMetadata(k.Metadata.ID)
	if err != nil {
		t.Fatalf("GetKeyMetadata: %v", err)
	}
	if md.Name != "meta" {
		t.Fatalf("unexpected metadata name: %q", md.Name)
	}

	// Mutating the returned copy must not affect the store.
	md.Name = "tampered"
	again, _ := ks.GetKeyMetadata(k.Metadata.ID)
	if again.Name != "meta" {
		t.Fatal("GetKeyMetadata must return a defensive copy")
	}

	if _, err := ks.GetKeyMetadata("nope"); err == nil {
		t.Fatal("GetKeyMetadata on missing id must error")
	}
}

func TestDeleteKey(t *testing.T) {
	ks := newTestStore(t)
	k := sampleKey("to-delete")
	if err := ks.StoreKey(k); err != nil {
		t.Fatalf("StoreKey: %v", err)
	}
	if err := ks.DeleteKey(k.Metadata.ID); err != nil {
		t.Fatalf("DeleteKey: %v", err)
	}
	if _, err := ks.GetKey(k.Metadata.ID); err == nil {
		t.Fatal("key should be gone after DeleteKey")
	}
	if err := ks.DeleteKey(k.Metadata.ID); err == nil {
		t.Fatal("DeleteKey on missing id must error")
	}
}

func TestListKeysByFilters(t *testing.T) {
	ks := newTestStore(t)

	sym := sampleKey("sym")
	sym.Metadata.Type = SymmetricKey
	sym.Metadata.Usage = EncryptionKey
	sym.Metadata.Tags = []string{"team:blue"}
	if err := ks.StoreKey(sym); err != nil {
		t.Fatalf("StoreKey sym: %v", err)
	}

	rsa := sampleKey("rsa")
	rsa.Metadata.Type = RSAKey
	rsa.Metadata.Usage = SigningKey
	rsa.Metadata.Tags = []string{"team:red"}
	if err := ks.StoreKey(rsa); err != nil {
		t.Fatalf("StoreKey rsa: %v", err)
	}

	all, err := ks.ListKeys()
	if err != nil {
		t.Fatalf("ListKeys: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(all))
	}

	byType, err := ks.ListKeysByType(RSAKey)
	if err != nil {
		t.Fatalf("ListKeysByType: %v", err)
	}
	if len(byType) != 1 || byType[0].Name != "rsa" {
		t.Fatalf("ListKeysByType returned %+v", byType)
	}

	byUsage, err := ks.ListKeysByUsage(SigningKey)
	if err != nil {
		t.Fatalf("ListKeysByUsage: %v", err)
	}
	if len(byUsage) != 1 || byUsage[0].Name != "rsa" {
		t.Fatalf("ListKeysByUsage returned %+v", byUsage)
	}

	byTag, err := ks.ListKeysByTag("team:blue")
	if err != nil {
		t.Fatalf("ListKeysByTag(team:blue): %v", err)
	}
	if len(byTag) != 1 || byTag[0].Name != "sym" {
		t.Fatalf("ListKeysByTag returned %+v", byTag)
	}

	none, err := ks.ListKeysByTag("team:green")
	if err != nil {
		t.Fatalf("ListKeysByTag(team:green): %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("ListKeysByTag for absent tag should be empty, got %d", len(none))
	}
}

// TestPersistence_EncryptionAtRestRoundTrip is the high-value case from #231:
// keys written to disk must survive a process restart (close + reopen) with the
// secret material intact, read back through the encrypted vault.
func TestPersistence_EncryptionAtRestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "keys.json")
	const pass = "round-trip-passphrase"

	ks1, err := NewFileKeyStore(KeyStoreOptions{
		StoragePath: path,
		Passphrase:  pass,
		AutoSave:    true,
	})
	if err != nil {
		t.Fatalf("open #1: %v", err)
	}
	k := sampleKey("persisted")
	if err := ks1.StoreKey(k); err != nil {
		t.Fatalf("StoreKey: %v", err)
	}
	id := k.Metadata.ID
	if err := ks1.Close(); err != nil {
		t.Fatalf("Close #1: %v", err)
	}

	// Reopen from the same path: simulates a process restart.
	ks2, err := NewFileKeyStore(KeyStoreOptions{
		StoragePath: path,
		Passphrase:  pass,
		AutoSave:    true,
	})
	if err != nil {
		t.Fatalf("open #2: %v", err)
	}
	t.Cleanup(func() { _ = ks2.Close() })

	got, err := ks2.GetKey(id)
	if err != nil {
		t.Fatalf("GetKey after reopen: %v", err)
	}
	if string(got.Material.Private) != "super-secret-material" {
		t.Fatalf("secret material did not survive reopen: %q", got.Material.Private)
	}
}

func TestGenerateKey_NilMetadata(t *testing.T) {
	ks := newTestStore(t)
	if _, err := ks.GenerateKey(SymmetricKey, "AES-256", nil); err == nil {
		t.Fatal("GenerateKey with nil metadata must error")
	}
}

func TestGenerateKey_UnsupportedType(t *testing.T) {
	ks := newTestStore(t)
	md := &KeyMetadata{Name: "x", Type: "bogus"}
	if _, err := ks.GenerateKey("bogus", "n/a", md); err == nil {
		t.Fatal("GenerateKey with unsupported type must error")
	}
}

func TestGenerateKey_StoresPlaceholder(t *testing.T) {
	ks := newTestStore(t)
	md := &KeyMetadata{Name: "gen", Type: SymmetricKey, ProtectionLevel: SoftwareProtection}
	k, err := ks.GenerateKey(SymmetricKey, "AES-256", md)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	if len(k.Material.Private) == 0 {
		t.Fatal("generated key should carry material")
	}
	if _, err := ks.GetKey(k.Metadata.ID); err != nil {
		t.Fatalf("generated key should be retrievable: %v", err)
	}
}

// TestUnimplementedAccessors locks in the honest "not implemented" contract of
// the placeholder crypto accessors so a future real implementation has to update
// the test deliberately rather than silently changing behavior.
func TestUnimplementedAccessors(t *testing.T) {
	ks := newTestStore(t)

	if _, err := ks.GetRSAPrivateKey("id"); err == nil {
		t.Error("GetRSAPrivateKey should report not implemented")
	}
	if _, err := ks.GetRSAPublicKey("id"); err == nil {
		t.Error("GetRSAPublicKey should report not implemented")
	}
	if _, err := ks.GetECDSAPrivateKey("id"); err == nil {
		t.Error("GetECDSAPrivateKey should report not implemented")
	}
	if _, err := ks.GetEd25519PrivateKey("id"); err == nil {
		t.Error("GetEd25519PrivateKey should report not implemented")
	}
	if _, err := ks.GetCertificate("id"); err == nil {
		t.Error("GetCertificate should report not implemented")
	}
	if _, err := ks.RotateKey("id"); err == nil {
		t.Error("RotateKey should report not implemented")
	}
	if _, err := ks.ExportKey("id", "PEM", false); err == nil {
		t.Error("ExportKey should report not implemented")
	}
	if _, err := ks.ImportKey([]byte("data"), "PEM", &KeyMetadata{}); err == nil {
		t.Error("ImportKey should report not implemented")
	}
}

func TestRotationAlertCallback(t *testing.T) {
	dir := t.TempDir()
	fired := make(chan int, 1)
	ks, err := NewFileKeyStore(KeyStoreOptions{
		StoragePath:           filepath.Join(dir, "keys.json"),
		Passphrase:            "pass",
		AutoSave:              true,
		RotationCheckInterval: 10 * time.Millisecond,
		AlertCallback: func(_ *KeyMetadata, days int) {
			select {
			case fired <- days:
			default:
			}
		},
	})
	if err != nil {
		t.Fatalf("NewFileKeyStore: %v", err)
	}
	t.Cleanup(func() { _ = ks.Close() })

	// A key created in the past with a short rotation period is overdue.
	k := sampleKey("overdue")
	k.Metadata.CreatedAt = time.Now().AddDate(0, 0, -30)
	k.Metadata.RotationPeriod = 1 // 1 day
	if err := ks.StoreKey(k); err != nil {
		t.Fatalf("StoreKey: %v", err)
	}

	select {
	case days := <-fired:
		if days > 0 {
			t.Fatalf("overdue key should report <= 0 days, got %d", days)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("rotation alert callback never fired for overdue key")
	}
}
