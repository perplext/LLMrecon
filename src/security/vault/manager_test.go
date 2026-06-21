package vault

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/perplext/LLMrecon/src/provider/core"
)

func newManager(t *testing.T) *CredentialManager {
	t.Helper()
	m, err := NewCredentialManager(ManagerOptions{
		ConfigDir:  t.TempDir(),
		Passphrase: "manager-test-pass",
		AutoSave:   false,
	})
	if err != nil {
		t.Fatalf("NewCredentialManager: %v", err)
	}
	t.Cleanup(func() { _ = m.Close() })
	return m
}

func TestManager_SetAndGetAPIKey(t *testing.T) {
	m := newManager(t)

	if err := m.SetAPIKey(core.OpenAIProvider, "sk-openai", "test key"); err != nil {
		t.Fatalf("SetAPIKey: %v", err)
	}
	got, err := m.GetAPIKey(core.OpenAIProvider)
	if err != nil {
		t.Fatalf("GetAPIKey: %v", err)
	}
	if got != "sk-openai" {
		t.Fatalf("GetAPIKey = %q, want sk-openai", got)
	}
}

func TestManager_UnknownProvider(t *testing.T) {
	m := newManager(t)
	bogus := core.ProviderType("nonexistent-provider")

	if _, err := m.GetAPIKey(bogus); err == nil {
		t.Error("GetAPIKey for unknown provider must error")
	}
	if err := m.SetAPIKey(bogus, "x", ""); err == nil {
		t.Error("SetAPIKey for unknown provider must error")
	}
}

func TestManager_GetAPIKey_NoneStored(t *testing.T) {
	m := newManager(t)
	if _, err := m.GetAPIKey(core.AnthropicProvider); err == nil {
		t.Error("GetAPIKey with no stored key and no env var must error")
	}
}

func TestManager_GetAPIKey_EnvFallback(t *testing.T) {
	// No credential stored, but a matching env var exists -> fallback returns it.
	t.Setenv("LLMRT_ANTHROPIC_API_KEY", "sk-env-anthropic")
	m := newManager(t)

	got, err := m.GetAPIKey(core.AnthropicProvider)
	if err != nil {
		t.Fatalf("GetAPIKey env fallback: %v", err)
	}
	if got != "sk-env-anthropic" {
		t.Fatalf("env fallback = %q", got)
	}
}

func TestManager_LoadFromEnv(t *testing.T) {
	// Set before constructing: NewCredentialManager calls LoadFromEnv.
	t.Setenv("LLMRT_OPENAI_API_KEY", "sk-loaded-from-env")
	m := newManager(t)

	creds, err := m.ListCredentialsByService("openai")
	if err != nil {
		t.Fatalf("ListCredentialsByService: %v", err)
	}
	found := false
	for _, c := range creds {
		if c.Value == "sk-loaded-from-env" && c.Type == APIKeyCredential {
			found = true
		}
	}
	if !found {
		t.Fatalf("LoadFromEnv did not import the openai key; got %+v", creds)
	}
}

func TestManager_DelegatesCRUD(t *testing.T) {
	m := newManager(t)
	cred := sampleCred("delegated")

	if err := m.StoreCredential(cred); err != nil {
		t.Fatalf("StoreCredential: %v", err)
	}
	got, err := m.GetCredential("delegated")
	if err != nil {
		t.Fatalf("GetCredential: %v", err)
	}
	if got.Value != "sk-secret-value" {
		t.Fatalf("delegated get value = %q", got.Value)
	}
	if err := m.RotateCredential("delegated", "rotated"); err != nil {
		t.Fatalf("RotateCredential: %v", err)
	}
	if err := m.DeleteCredential("delegated"); err != nil {
		t.Fatalf("DeleteCredential: %v", err)
	}
	if _, err := m.GetCredential("delegated"); err == nil {
		t.Fatal("credential should be gone after delete")
	}
}

func TestManager_InstallGitHookInDir(t *testing.T) {
	m := newManager(t)

	// Missing .git directory -> error.
	if err := m.InstallGitHookInDir(t.TempDir()); err == nil {
		t.Fatal("InstallGitHookInDir without a .git dir must error")
	}

	// Realistic repo layout: <dir>/.git/hooks must exist for the hook write.
	dir := t.TempDir()
	hooks := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hooks, 0700); err != nil {
		t.Fatalf("mkdir hooks: %v", err)
	}

	if err := m.InstallGitHookInDir(dir); err != nil {
		t.Fatalf("InstallGitHookInDir: %v", err)
	}
	hook, err := os.ReadFile(filepath.Join(hooks, "pre-commit"))
	if err != nil {
		t.Fatalf("hook not written: %v", err)
	}
	if !strings.Contains(string(hook), "LLMrecon credential check") {
		t.Fatal("installed hook missing the credential-check marker")
	}

	// Idempotent: installing again is a no-op (marker already present).
	if err := m.InstallGitHookInDir(dir); err != nil {
		t.Fatalf("second InstallGitHookInDir should be a no-op: %v", err)
	}
}
