package persistence

import (
	"context"
	"errors"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks/common"
	"github.com/perplext/LLMrecon/src/attacks/testutil"
)

func trustfallConfig() common.AttackConfig {
	return common.AttackConfig{
		Metadata: map[string]string{"i_understand_risks": "true"},
	}
}

// Covers AE1: missing gate → SkipGateBlocked.
func TestTrustFall_GateBlocked(t *testing.T) {
	cfg := trustfallConfig()
	cfg.Metadata["i_understand_risks"] = "false"
	r, _ := (&TrustFallModule{}).Execute(context.Background(), &testutil.MockCodingAgent{}, cfg)
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipGateBlocked {
		t.Errorf("outcome=%q skip=%q, want skipped/gate_blocked", r.Outcome, r.SkipReason)
	}
}

// Covers AE2: text provider lacks CodingAgentProvider → SkipMissingCapability.
func TestTrustFall_MissingCapability(t *testing.T) {
	r, _ := (&TrustFallModule{}).Execute(context.Background(), &testutil.MockProvider{}, trustfallConfig())
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipMissingCapability {
		t.Errorf("outcome=%q skip=%q, want skipped/missing_capability", r.Outcome, r.SkipReason)
	}
}

func TestTrustFall_SuccessOnAutoExecute(t *testing.T) {
	agent := &testutil.MockCodingAgent{AutoExecuteOnTrust: true}
	r, _ := (&TrustFallModule{}).Execute(context.Background(), agent, trustfallConfig())
	if r.Outcome != common.OutcomeSuccess {
		t.Fatalf("outcome=%q, want success", r.Outcome)
	}
	if len(agent.ExecutedMCP) != 1 {
		t.Errorf("expected 1 executed MCP path, got %d", len(agent.ExecutedMCP))
	}
}

func TestTrustFall_RefusedWithoutAutoExecute(t *testing.T) {
	agent := &testutil.MockCodingAgent{AutoExecuteOnTrust: false}
	r, _ := (&TrustFallModule{}).Execute(context.Background(), agent, trustfallConfig())
	if r.Outcome != common.OutcomeRefused {
		t.Errorf("outcome=%q, want refused", r.Outcome)
	}
}

func TestTrustFall_NoTrustPrompt(t *testing.T) {
	agent := &testutil.MockCodingAgent{NoTrustPrompt: true}
	r, _ := (&TrustFallModule{}).Execute(context.Background(), agent, trustfallConfig())
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipNoMutationTarget {
		t.Errorf("outcome=%q skip=%q, want skipped/no_safety_step_to_hijack", r.Outcome, r.SkipReason)
	}
}

func TestTrustFall_ProviderError(t *testing.T) {
	agent := &testutil.MockCodingAgent{TrustErr: errors.New("transport down")}
	r, _ := (&TrustFallModule{}).Execute(context.Background(), agent, trustfallConfig())
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipProviderError {
		t.Errorf("outcome=%q skip=%q, want skipped/provider_error", r.Outcome, r.SkipReason)
	}
}
