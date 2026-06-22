package keystore

import (
	"crypto/rsa"
	"testing"
	"time"
)

func newRotatorWith(entries ...*KeyEntry) (*Keystore, *KeyRotator) {
	ks := &Keystore{keys: make(map[string]*KeyEntry)}
	for _, e := range entries {
		ks.keys[e.ID] = e
	}
	return ks, NewKeyRotator(ks)
}

func rsaEntry(id string) *KeyEntry {
	return &KeyEntry{
		ID: id, Type: KeyTypeRSA, Algorithm: "RSA-2048",
		Metadata: map[string]string{"env": "prod"},
		CreatedAt: time.Now().AddDate(0, 0, -100),
		UpdatedAt: time.Now(),
	}
}

func TestRotator_SetAndGetPolicy(t *testing.T) {
	_, kr := newRotatorWith(rsaEntry("k1"))

	if err := kr.SetRotationPolicy("", RotationPolicy{}); err == nil {
		t.Error("empty key id must error")
	}
	if err := kr.SetRotationPolicy("ghost", RotationPolicy{}); err == nil {
		t.Error("policy for missing key must error")
	}

	policy := RotationPolicy{RotationPeriod: 24 * time.Hour, AutoRotate: true}
	if err := kr.SetRotationPolicy("k1", policy); err != nil {
		t.Fatalf("SetRotationPolicy: %v", err)
	}

	got, err := kr.GetRotationPolicy("k1")
	if err != nil {
		t.Fatalf("GetRotationPolicy: %v", err)
	}
	if got.RotationPeriod != 24*time.Hour {
		t.Fatalf("policy not stored correctly: %+v", got)
	}

	if _, err := kr.GetRotationPolicy("k2"); err == nil {
		t.Error("GetRotationPolicy for key without policy must error")
	}
}

func TestRotator_RotateRSAKey_RemovesOldByDefault(t *testing.T) {
	ks, kr := newRotatorWith(rsaEntry("k1"))

	if err := kr.RotateKey("", "r", "tester"); err == nil {
		t.Error("empty id must error")
	}

	if err := kr.RotateKey("k1", "scheduled refresh", "tester"); err != nil {
		t.Fatalf("RotateKey: %v", err)
	}

	// Old key removed (no archive policy set).
	if _, ok := ks.keys["k1"]; ok {
		t.Fatal("old key should be removed when ArchiveOldKeys is false")
	}

	// Exactly one new key, and it must hold a real *rsa.PrivateKey.
	if len(ks.keys) != 1 {
		t.Fatalf("expected 1 key after rotation, got %d", len(ks.keys))
	}
	for id, entry := range ks.keys {
		if _, ok := entry.Key.(*rsa.PrivateKey); !ok {
			t.Fatalf("rotated key %s is not a real RSA private key: %T", id, entry.Key)
		}
		if entry.Metadata["rotated_from"] != "k1" {
			t.Fatalf("rotated key missing provenance metadata: %+v", entry.Metadata)
		}
	}

	hist := kr.GetRotationHistory()
	if len(hist) != 1 || hist[0].OldKeyID != "k1" {
		t.Fatalf("rotation history not recorded: %+v", hist)
	}
}

func TestRotator_RotateAESKey(t *testing.T) {
	aes := &KeyEntry{
		ID: "a1", Type: KeyTypeAES, Algorithm: "AES-256",
		Metadata: map[string]string{}, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	ks, kr := newRotatorWith(aes)
	if err := kr.RotateKey("a1", "manual", "tester"); err != nil {
		t.Fatalf("RotateKey: %v", err)
	}
	for _, entry := range ks.keys {
		material, ok := entry.Key.([]byte)
		if !ok {
			t.Fatalf("rotated AES key is not a byte slice: %T", entry.Key)
		}
		if len(material) != 32 {
			t.Fatalf("rotated AES key should be 256-bit, got %d bytes", len(material))
		}
	}
}

func TestRotator_RotateUnsupportedType(t *testing.T) {
	bad := &KeyEntry{ID: "x", Type: "exotic", CreatedAt: time.Now()}
	_, kr := newRotatorWith(bad)
	if err := kr.RotateKey("x", "r", "tester"); err == nil {
		t.Fatal("rotating unsupported key type must error")
	}
}

func TestRotator_ArchivePolicyKeepsOldKey(t *testing.T) {
	ks, kr := newRotatorWith(rsaEntry("k1"))
	if err := kr.SetRotationPolicy("k1", RotationPolicy{ArchiveOldKeys: true}); err != nil {
		t.Fatalf("SetRotationPolicy: %v", err)
	}
	if err := kr.RotateKey("k1", "r", "tester"); err != nil {
		t.Fatalf("RotateKey: %v", err)
	}
	old, ok := ks.keys["k1"]
	if !ok {
		t.Fatal("archived old key should be retained")
	}
	if old.Metadata["status"] != "archived" {
		t.Fatalf("old key should be marked archived, got %q", old.Metadata["status"])
	}
}

func TestRotator_StatusCalculations(t *testing.T) {
	_, kr := newRotatorWith(rsaEntry("expired"), rsaEntry("expiring"))

	// MaxAge already exceeded (key created 100 days ago) -> expired.
	_ = kr.SetRotationPolicy("expired", RotationPolicy{MaxAge: 24 * time.Hour})
	// Long MaxAge but NotifyBefore window covers now -> expiring.
	_ = kr.SetRotationPolicy("expiring", RotationPolicy{
		MaxAge:       200 * 24 * time.Hour,
		NotifyBefore: 150 * 24 * time.Hour,
	})

	statuses := kr.CheckRotationStatus()
	if statuses["expired"] != StatusExpired {
		t.Fatalf("expected expired, got %s", statuses["expired"])
	}
	if statuses["expiring"] != StatusExpiring {
		t.Fatalf("expected expiring, got %s", statuses["expiring"])
	}

	if got := kr.GetExpiredKeys(); len(got) != 1 {
		t.Fatalf("GetExpiredKeys = %d, want 1", len(got))
	}
	if got := kr.GetExpiringKeys(); len(got) != 1 {
		t.Fatalf("GetExpiringKeys = %d, want 1", len(got))
	}
}

func TestRotator_GetRotationInfo(t *testing.T) {
	_, kr := newRotatorWith(rsaEntry("k1"))
	if _, err := kr.GetRotationInfo(""); err == nil {
		t.Error("empty id must error")
	}
	if _, err := kr.GetRotationInfo("k1"); err == nil {
		t.Error("info before any policy/rotation must error")
	}
	_ = kr.SetRotationPolicy("k1", RotationPolicy{RotationPeriod: time.Hour})
	if _, err := kr.GetRotationInfo("k1"); err != nil {
		t.Fatalf("GetRotationInfo after policy: %v", err)
	}
}

func TestRotator_AutoRotationLifecycle(t *testing.T) {
	_, kr := newRotatorWith(rsaEntry("k1"))

	// First start must succeed on a fresh rotator (regression: stopChan was
	// previously initialized open, making this always fail with "already
	// running" — auto-rotation could never start).
	if err := kr.StartAutoRotation(); err != nil {
		t.Fatalf("StartAutoRotation on fresh rotator: %v", err)
	}
	if err := kr.StartAutoRotation(); err == nil {
		t.Error("double StartAutoRotation must error")
	}
	if err := kr.StopAutoRotation(); err != nil {
		t.Fatalf("StopAutoRotation: %v", err)
	}
	if err := kr.StopAutoRotation(); err == nil {
		t.Error("double StopAutoRotation must error")
	}

	// After a clean stop, the rotator must be restartable.
	if err := kr.StartAutoRotation(); err != nil {
		t.Fatalf("restart after stop: %v", err)
	}
	if err := kr.StopAutoRotation(); err != nil {
		t.Fatalf("final stop: %v", err)
	}
}
