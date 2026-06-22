package keystore

import "testing"

// The HSMManager is a simulated stub (no real PKCS#11 backend). These tests
// lock in its documented contract: connection lifecycle works, the simulated
// store/delete succeed, and every key-material accessor refuses honestly rather
// than returning fake material.

func enabledHSM(t *testing.T) *HSMManager {
	t.Helper()
	m, err := NewHSMManager(HSMConfig{Enabled: true, Provider: "pkcs11"})
	if err != nil {
		t.Fatalf("NewHSMManager: %v", err)
	}
	t.Cleanup(func() { _ = m.Close() })
	return m
}

func TestNewHSMManager_Disabled(t *testing.T) {
	if _, err := NewHSMManager(HSMConfig{Enabled: false}); err == nil {
		t.Fatal("NewHSMManager must error when HSM is disabled")
	}
}

func TestHSM_SimulatedStoreDelete(t *testing.T) {
	m := enabledHSM(t)
	if err := m.StoreKey(&Key{Metadata: KeyMetadata{ID: "k"}}); err != nil {
		t.Errorf("simulated StoreKey should succeed: %v", err)
	}
	if err := m.DeleteKey("k"); err != nil {
		t.Errorf("simulated DeleteKey should succeed: %v", err)
	}
}

func TestHSM_GetKeyNotSupported(t *testing.T) {
	m := enabledHSM(t)
	if _, err := m.GetKey("k"); err == nil {
		t.Error("direct GetKey from HSM must be unsupported")
	}
}

func TestHSM_ExportRefused(t *testing.T) {
	m := enabledHSM(t)
	if _, err := m.ExportKey("k", "PEM", true); err == nil {
		t.Error("private key export from HSM must be refused")
	}
	if _, err := m.ExportKey("k", "PEM", false); err == nil {
		t.Error("HSM export is not implemented and must error")
	}
	if _, err := m.ImportKey([]byte("d"), "PEM", &KeyMetadata{}); err == nil {
		t.Error("HSM import is not implemented and must error")
	}
}

func TestHSM_TypedAccessorsNotImplemented(t *testing.T) {
	m := enabledHSM(t)
	if _, err := m.GetRSAPrivateKey("k"); err == nil {
		t.Error("GetRSAPrivateKey must report not implemented")
	}
	if _, err := m.GetRSAPublicKey("k"); err == nil {
		t.Error("GetRSAPublicKey must report not implemented")
	}
	if _, err := m.GetECDSAPrivateKey("k"); err == nil {
		t.Error("GetECDSAPrivateKey must report not implemented")
	}
	if _, err := m.GetEd25519PublicKey("k"); err == nil {
		t.Error("GetEd25519PublicKey must report not implemented")
	}
}

func TestHSM_Close(t *testing.T) {
	m, err := NewHSMManager(HSMConfig{Enabled: true})
	if err != nil {
		t.Fatalf("NewHSMManager: %v", err)
	}
	if err := m.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
	// Close should be idempotent for the simulated manager.
	if err := m.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
}
