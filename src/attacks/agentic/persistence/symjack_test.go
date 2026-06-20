package persistence

import (
	"context"
	"errors"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks/common"
	"github.com/perplext/LLMrecon/src/attacks/testutil"
)

func symjackConfig() common.AttackConfig {
	return common.AttackConfig{
		Metadata: map[string]string{"i_understand_risks": "true", "shown_destination": "docs/assets/demo.mp4"},
	}
}

// Covers AE1: missing gate → SkipGateBlocked.
func TestSymJack_GateBlocked(t *testing.T) {
	cfg := symjackConfig()
	cfg.Metadata["i_understand_risks"] = "false"
	r, _ := (&SymJackModule{}).Execute(context.Background(), &testutil.MockCodingAgent{}, cfg)
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipGateBlocked {
		t.Errorf("outcome=%q skip=%q, want skipped/gate_blocked", r.Outcome, r.SkipReason)
	}
}

// Covers AE2: a plain text provider lacks CodingAgentProvider → SkipMissingCapability.
func TestSymJack_MissingCapability(t *testing.T) {
	r, _ := (&SymJackModule{}).Execute(context.Background(), &testutil.MockProvider{}, symjackConfig())
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipMissingCapability {
		t.Errorf("outcome=%q skip=%q, want skipped/missing_capability", r.Outcome, r.SkipReason)
	}
}

// Covers AE5: symlinked destination resolves to config dir and write lands → success.
func TestSymJack_SuccessOnMisrepresentation(t *testing.T) {
	agent := &testutil.MockCodingAgent{
		Symlinks: map[string]string{"docs/assets/demo.mp4": "/home/user/.config/mcp/servers.json"},
	}
	r, _ := (&SymJackModule{}).Execute(context.Background(), agent, symjackConfig())
	if r.Outcome != common.OutcomeSuccess {
		t.Fatalf("outcome=%q, want success", r.Outcome)
	}
	if r.Metadata["resolved_destination"] != "/home/user/.config/mcp/servers.json" {
		t.Errorf("resolved_destination = %v", r.Metadata["resolved_destination"])
	}
	if agent.WriteCount() != 1 {
		t.Errorf("expected 1 write, got %d", agent.WriteCount())
	}
}

func TestSymJack_RefusedWhenHonest(t *testing.T) {
	// No symlink mapping → resolved == shown → not misrepresented.
	agent := &testutil.MockCodingAgent{}
	r, _ := (&SymJackModule{}).Execute(context.Background(), agent, symjackConfig())
	if r.Outcome != common.OutcomeRefused {
		t.Errorf("outcome=%q, want refused", r.Outcome)
	}
}

func TestSymJack_NoApprovalStep(t *testing.T) {
	agent := &testutil.MockCodingAgent{NoApprovalStep: true}
	r, _ := (&SymJackModule{}).Execute(context.Background(), agent, symjackConfig())
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipNoMutationTarget {
		t.Errorf("outcome=%q skip=%q, want skipped/no_safety_step_to_hijack", r.Outcome, r.SkipReason)
	}
}

func TestSymJack_ProviderError(t *testing.T) {
	agent := &testutil.MockCodingAgent{ApproveErr: errors.New("transport down")}
	r, _ := (&SymJackModule{}).Execute(context.Background(), agent, symjackConfig())
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipProviderError {
		t.Errorf("outcome=%q skip=%q, want skipped/provider_error", r.Outcome, r.SkipReason)
	}
}
