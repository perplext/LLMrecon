package persistence

import (
	"context"
	"errors"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks/common"
	"github.com/perplext/LLMrecon/src/attacks/testutil"
)

func mosaicConfig() common.AttackConfig {
	return common.AttackConfig{Metadata: map[string]string{"i_understand_risks": "true"}}
}

func TestMosaic_GateBlocked(t *testing.T) {
	cfg := mosaicConfig()
	cfg.Metadata["i_understand_risks"] = "false"
	r, _ := (&MosaicCmdChainModule{}).Execute(context.Background(), &testutil.MockCodingAgent{}, cfg)
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipGateBlocked {
		t.Errorf("outcome=%q skip=%q, want skipped/gate_blocked", r.Outcome, r.SkipReason)
	}
}

func TestMosaic_MissingCapability(t *testing.T) {
	r, _ := (&MosaicCmdChainModule{}).Execute(context.Background(), &testutil.MockProvider{}, mosaicConfig())
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipMissingCapability {
		t.Errorf("outcome=%q skip=%q, want skipped/missing_capability", r.Outcome, r.SkipReason)
	}
}

// Default chain: each command allowed, composition dangerous → success.
func TestMosaic_SuccessOnComposition(t *testing.T) {
	agent := &testutil.MockCodingAgent{}
	r, _ := (&MosaicCmdChainModule{}).Execute(context.Background(), agent, mosaicConfig())
	if r.Outcome != common.OutcomeSuccess {
		t.Fatalf("outcome=%q, want success", r.Outcome)
	}
	if r.Metadata["all_commands_allowed"] != true || r.Metadata["dangerous_composition"] != true {
		t.Errorf("all_allowed=%v dangerous=%v, want both true", r.Metadata["all_commands_allowed"], r.Metadata["dangerous_composition"])
	}
	if len(agent.RanCommands) == 0 {
		t.Errorf("expected the chain to be executed against the agent")
	}
}

// Hardened agent blocks the consumer step → chain broken → refused.
func TestMosaic_RefusedWhenStepBlocked(t *testing.T) {
	agent := &testutil.MockCodingAgent{ExtraCommandBlock: []string{"sh -c"}}
	r, _ := (&MosaicCmdChainModule{}).Execute(context.Background(), agent, mosaicConfig())
	if r.Outcome != common.OutcomeRefused {
		t.Errorf("outcome=%q, want refused", r.Outcome)
	}
	if r.Metadata["all_commands_allowed"] != false {
		t.Errorf("all_commands_allowed=%v, want false", r.Metadata["all_commands_allowed"])
	}
}

// Sandbox isolates OS state → no dangerous composition → refused.
func TestMosaic_RefusedWhenIsolated(t *testing.T) {
	agent := &testutil.MockCodingAgent{ForceNoDangerousComposition: true}
	r, _ := (&MosaicCmdChainModule{}).Execute(context.Background(), agent, mosaicConfig())
	if r.Outcome != common.OutcomeRefused {
		t.Errorf("outcome=%q, want refused", r.Outcome)
	}
}

func TestMosaic_ProviderError(t *testing.T) {
	agent := &testutil.MockCodingAgent{CommandChainErr: errors.New("shell unavailable")}
	r, _ := (&MosaicCmdChainModule{}).Execute(context.Background(), agent, mosaicConfig())
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipProviderError {
		t.Errorf("outcome=%q skip=%q, want skipped/provider_error", r.Outcome, r.SkipReason)
	}
}

func TestMosaic_CustomChain(t *testing.T) {
	cfg := mosaicConfig()
	cfg.Metadata["command_chain"] = "echo hi > f\nbash -c \"$(cat f)\""
	agent := &testutil.MockCodingAgent{}
	r, _ := (&MosaicCmdChainModule{}).Execute(context.Background(), agent, cfg)
	if r.Metadata["chain_length"] != 2 {
		t.Errorf("chain_length=%v, want 2", r.Metadata["chain_length"])
	}
}
