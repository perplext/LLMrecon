package vault

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newVault(t *testing.T, autoSave bool) (*SecureVault, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "vault.enc")
	v, err := NewSecureVault(path, VaultOptions{
		Passphrase: "vault-test-passphrase",
		AutoSave:   autoSave,
		// Long interval so the background rotation checker never fires mid-test.
		RotationCheckInterval: time.Hour,
	})
	if err != nil {
		t.Fatalf("NewSecureVault: %v", err)
	}
	t.Cleanup(func() { _ = v.Close() })
	return v, path
}

func sampleCred(id string) *Credential {
	return &Credential{
		ID:      id,
		Name:    "openai-key",
		Type:    APIKeyCredential,
		Service: "openai",
		Value:   "sk-secret-value",
		Tags:    []string{"prod"},
	}
}

func TestStoreAndGetCredential(t *testing.T) {
	v, _ := newVault(t, false)

	if err := v.StoreCredential(sampleCred("c1")); err != nil {
		t.Fatalf("StoreCredential: %v", err)
	}
	got, err := v.GetCredential("c1")
	if err != nil {
		t.Fatalf("GetCredential: %v", err)
	}
	if got.Value != "sk-secret-value" {
		t.Fatalf("value mismatch: %q", got.Value)
	}
	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Fatal("timestamps should be set on store")
	}
}

func TestStoreCredential_EmptyID(t *testing.T) {
	v, _ := newVault(t, false)
	if err := v.StoreCredential(&Credential{Value: "x"}); err == nil {
		t.Fatal("StoreCredential with empty ID must error")
	}
}

func TestGetCredential_NotFound(t *testing.T) {
	v, _ := newVault(t, false)
	if _, err := v.GetCredential("ghost"); err == nil {
		t.Fatal("GetCredential on missing id must error")
	}
}

func TestDeleteCredential(t *testing.T) {
	v, _ := newVault(t, false)
	_ = v.StoreCredential(sampleCred("c1"))

	if err := v.DeleteCredential("c1"); err != nil {
		t.Fatalf("DeleteCredential: %v", err)
	}
	if _, err := v.GetCredential("c1"); err == nil {
		t.Fatal("credential should be gone after delete")
	}
	if err := v.DeleteCredential("c1"); err == nil {
		t.Fatal("deleting missing credential must error")
	}
}

func TestListCredentialsByFilters(t *testing.T) {
	v, _ := newVault(t, false)

	a := sampleCred("a")
	a.Service = "openai"
	a.Type = APIKeyCredential
	a.Tags = []string{"prod"}
	b := sampleCred("b")
	b.Service = "anthropic"
	b.Type = TokenCredential
	b.Tags = []string{"dev"}
	_ = v.StoreCredential(a)
	_ = v.StoreCredential(b)

	all, _ := v.ListCredentials()
	if len(all) != 2 {
		t.Fatalf("expected 2 credentials, got %d", len(all))
	}

	bySvc, _ := v.ListCredentialsByService("anthropic")
	if len(bySvc) != 1 || bySvc[0].ID != "b" {
		t.Fatalf("ListCredentialsByService = %+v", bySvc)
	}

	byType, _ := v.ListCredentialsByType(APIKeyCredential)
	if len(byType) != 1 || byType[0].ID != "a" {
		t.Fatalf("ListCredentialsByType = %+v", byType)
	}

	byTag, _ := v.ListCredentialsByTag("dev")
	if len(byTag) != 1 || byTag[0].ID != "b" {
		t.Fatalf("ListCredentialsByTag = %+v", byTag)
	}
}

func TestRotateCredential(t *testing.T) {
	v, _ := newVault(t, false)
	cred := sampleCred("c1")
	cred.RotationPolicy = &RotationPolicy{Enabled: true, IntervalDays: 30}
	_ = v.StoreCredential(cred)

	if err := v.RotateCredential("c1", "sk-new-value"); err != nil {
		t.Fatalf("RotateCredential: %v", err)
	}
	got, _ := v.GetCredential("c1")
	if got.Value != "sk-new-value" {
		t.Fatalf("rotated value not applied: %q", got.Value)
	}
	if got.RotationPolicy.LastRotation.IsZero() {
		t.Fatal("rotation should stamp LastRotation")
	}

	if err := v.RotateCredential("missing", "x"); err == nil {
		t.Fatal("rotating missing credential must error")
	}
}

func TestGetCredentialsNeedingRotation(t *testing.T) {
	v, _ := newVault(t, false)

	due := sampleCred("due")
	due.CreatedAt = time.Now().AddDate(0, 0, -60)
	due.RotationPolicy = &RotationPolicy{Enabled: true, IntervalDays: 30}
	_ = v.StoreCredential(due)
	// StoreCredential overwrites CreatedAt for new creds; reset it to the past
	// so the rotation math sees an overdue credential.
	due.CreatedAt = time.Now().AddDate(0, 0, -60)

	fresh := sampleCred("fresh")
	fresh.RotationPolicy = &RotationPolicy{Enabled: true, IntervalDays: 30}
	_ = v.StoreCredential(fresh)

	needing, err := v.GetCredentialsNeedingRotation()
	if err != nil {
		t.Fatalf("GetCredentialsNeedingRotation: %v", err)
	}
	if len(needing) != 1 || needing[0].ID != "due" {
		t.Fatalf("expected only 'due' needing rotation, got %+v", needing)
	}
}

// TestEncryptionAtRest_RoundTrip is the core #231 case for vault: a credential
// written to disk must survive a close + reopen with the secret intact, read
// back through the encrypted file.
func TestEncryptionAtRest_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.enc")
	const pass = "round-trip-pass"

	v1, err := NewSecureVault(path, VaultOptions{Passphrase: pass})
	if err != nil {
		t.Fatalf("open #1: %v", err)
	}
	_ = v1.StoreCredential(sampleCred("persist"))
	if err := v1.Close(); err != nil { // Close saves
		t.Fatalf("Close #1: %v", err)
	}

	// The on-disk file must not contain the plaintext secret.
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read vault file: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("vault file is empty after save")
	}
	if containsPlaintext(raw, "sk-secret-value") {
		t.Fatal("plaintext secret found in vault file — encryption at rest broken")
	}

	v2, err := NewSecureVault(path, VaultOptions{Passphrase: pass})
	if err != nil {
		t.Fatalf("open #2: %v", err)
	}
	t.Cleanup(func() { _ = v2.Close() })

	got, err := v2.GetCredential("persist")
	if err != nil {
		t.Fatalf("GetCredential after reopen: %v", err)
	}
	if got.Value != "sk-secret-value" {
		t.Fatalf("secret did not survive reopen: %q", got.Value)
	}
}

// TestWrongPassphrase_CannotOpen verifies the central security property: a vault
// encrypted under one passphrase cannot be opened (decrypted) with another.
func TestWrongPassphrase_CannotOpen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.enc")

	v1, err := NewSecureVault(path, VaultOptions{Passphrase: "right-pass"})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	_ = v1.StoreCredential(sampleCred("c1"))
	_ = v1.Close()

	if _, err := NewSecureVault(path, VaultOptions{Passphrase: "wrong-pass"}); err == nil {
		t.Fatal("opening vault with wrong passphrase must fail (GCM auth)")
	}
}

// TestTamperedCiphertext_FailsAuth verifies AES-GCM integrity: a flipped byte in
// the stored ciphertext must be detected on load rather than silently accepted.
func TestTamperedCiphertext_FailsAuth(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.enc")
	const pass = "tamper-pass"

	v1, _ := NewSecureVault(path, VaultOptions{Passphrase: pass})
	_ = v1.StoreCredential(sampleCred("c1"))
	_ = v1.Close()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	// Flip the last base64 char to corrupt the ciphertext/tag.
	if raw[len(raw)-1] == 'A' {
		raw[len(raw)-1] = 'B'
	} else {
		raw[len(raw)-1] = 'A'
	}
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatalf("write tampered: %v", err)
	}

	if _, err := NewSecureVault(path, VaultOptions{Passphrase: pass}); err == nil {
		t.Fatal("tampered ciphertext must fail to load")
	}
}

func TestCorruptedFile_FailsToLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.enc")
	if err := os.WriteFile(path, []byte("not-valid-base64-or-ciphertext!!!"), 0600); err != nil {
		t.Fatalf("write: %v", err)
	}
	if _, err := NewSecureVault(path, VaultOptions{Passphrase: "x"}); err == nil {
		t.Fatal("corrupted vault file must fail to load")
	}
}

func TestGenerateCredentialID(t *testing.T) {
	id1 := GenerateCredentialID("openai", "key")
	id2 := GenerateCredentialID("openai", "key")
	if id1 == "" {
		t.Fatal("ID must not be empty")
	}
	if id1 == id2 {
		t.Fatal("IDs should be unique even for same service/name (nanosecond entropy)")
	}
	if len(id1) != 16 { // 8 bytes hex-encoded
		t.Fatalf("unexpected ID length %d: %s", len(id1), id1)
	}
}

func containsPlaintext(haystack []byte, needle string) bool {
	n := []byte(needle)
	for i := 0; i+len(n) <= len(haystack); i++ {
		if string(haystack[i:i+len(n)]) == needle {
			return true
		}
	}
	return false
}
