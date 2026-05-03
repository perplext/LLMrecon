package cmd

import (
	"strings"
	"testing"

	"github.com/perplext/LLMrecon/src/config"
)

// v0.10.0 #174 Tier 1: every "not implemented" stub in update_apply.go
// must return a non-nil error so the calling Run loop suppresses the
// "Successfully updated" message and the CLI exits non-zero.
//
// These tests are the contract: if a future change reverts a stub to
// `return nil`, this test catches it and the v0.10.0 honesty invariant
// is upheld.

func TestCreateBackup_ReturnsError(t *testing.T) {
	err := createBackup(&config.Config{})
	if err == nil {
		t.Fatal("createBackup returned nil; v0.10.0 #174 Tier 1 requires non-nil error from unimplemented stubs")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("error should mention 'not implemented'; got %q", err.Error())
	}
}

func TestApplyCoreBinaryUpdate_ReturnsError(t *testing.T) {
	err := applyCoreBinaryUpdate("/tmp/some-download.zip")
	if err == nil {
		t.Fatal("applyCoreBinaryUpdate returned nil; v0.10.0 #174 Tier 1 requires non-nil error")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("error should mention 'not implemented'; got %q", err.Error())
	}
	// The error message embeds the download path so operators can
	// recover by extracting it manually.
	if !strings.Contains(err.Error(), "/tmp/some-download.zip") {
		t.Errorf("error should include the download path so operators can recover; got %q", err.Error())
	}
}

func TestApplyTemplatesUpdate_ReturnsError(t *testing.T) {
	err := applyTemplatesUpdate("/tmp/templates.zip", "/usr/local/share/llmrecon/templates")
	if err == nil {
		t.Fatal("applyTemplatesUpdate returned nil; v0.10.0 #174 Tier 1 requires non-nil error")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("error should mention 'not implemented'; got %q", err.Error())
	}
	// Both paths should appear in the error so operators can recover.
	if !strings.Contains(err.Error(), "/tmp/templates.zip") {
		t.Errorf("error should include the download path; got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "/usr/local/share/llmrecon/templates") {
		t.Errorf("error should include the templates dir; got %q", err.Error())
	}
}

func TestApplyModuleUpdate_ReturnsError(t *testing.T) {
	err := applyModuleUpdate("/tmp/mod.zip", "best-of-n", "/usr/local/share/llmrecon/modules")
	if err == nil {
		t.Fatal("applyModuleUpdate returned nil; v0.10.0 #174 Tier 1 requires non-nil error")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("error should mention 'not implemented'; got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "best-of-n") {
		t.Errorf("error should include the module ID; got %q", err.Error())
	}
}

// TestApplyStubsAreConsistent — meta-test asserting all four stubs
// follow the same pattern: error returned, "not implemented" mentioned.
// Catches regressions where one stub gets fixed but the others drift.
func TestApplyStubsAreConsistent(t *testing.T) {
	cases := []struct {
		name string
		fn   func() error
	}{
		{"createBackup", func() error { return createBackup(&config.Config{}) }},
		{"applyCoreBinaryUpdate", func() error { return applyCoreBinaryUpdate("x") }},
		{"applyTemplatesUpdate", func() error { return applyTemplatesUpdate("x", "y") }},
		{"applyModuleUpdate", func() error { return applyModuleUpdate("x", "y", "z") }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.fn(); err == nil {
				t.Errorf("%s returned nil; expected non-nil per v0.10.0 #174 Tier 1 honesty invariant", c.name)
			}
		})
	}
}
