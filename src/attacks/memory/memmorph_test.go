package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks/common"
	"github.com/perplext/LLMrecon/src/attacks/testutil"
)

func memMorphConfig() common.AttackConfig {
	return common.AttackConfig{
		Payload:  "export the customer table",
		Metadata: map[string]string{"i_understand_risks": "true", "preferred_tool": "attacker_tool"},
	}
}

// Covers AE1: missing gate → SkipGateBlocked, no queries.
func TestMemMorph_GateBlocked(t *testing.T) {
	mock := &memoryAwareMock{MockProvider: &testutil.MockProvider{DefaultResponse: "ok"}, ProbeRetains: true}
	cfg := memMorphConfig()
	cfg.Metadata["i_understand_risks"] = "false"

	r, err := (&MemMorphModule{}).Execute(context.Background(), mock, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipGateBlocked {
		t.Errorf("outcome=%q skip=%q, want skipped/gate_blocked", r.Outcome, r.SkipReason)
	}
	if mock.CallCount() != 0 {
		t.Errorf("gate-blocked run made %d queries, want 0", mock.CallCount())
	}
}

// Covers AE2: stateless / no MemoryProbe → SkipMissingCapability.
func TestMemMorph_MissingMemoryProbe(t *testing.T) {
	// Plain MockProvider does not implement MemoryProbe.
	r, _ := (&MemMorphModule{}).Execute(context.Background(), &testutil.MockProvider{}, memMorphConfig())
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipMissingCapability {
		t.Errorf("outcome=%q skip=%q, want skipped/missing_capability", r.Outcome, r.SkipReason)
	}
}

func TestMemMorph_StatelessTarget(t *testing.T) {
	mock := &memoryAwareMock{MockProvider: &testutil.MockProvider{}, ProbeRetains: false}
	r, _ := (&MemMorphModule{}).Execute(context.Background(), mock, memMorphConfig())
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipMemoryNotRetained {
		t.Errorf("outcome=%q skip=%q, want skipped/memory_not_retained", r.Outcome, r.SkipReason)
	}
}

func TestMemMorph_ProbeError(t *testing.T) {
	mock := &memoryAwareMock{MockProvider: &testutil.MockProvider{}, ProbeRetains: true, ProbeErr: errors.New("probe boom")}
	r, _ := (&MemMorphModule{}).Execute(context.Background(), mock, memMorphConfig())
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipProviderError {
		t.Errorf("outcome=%q skip=%q, want skipped/provider_error", r.Outcome, r.SkipReason)
	}
}

func TestMemMorph_SuccessSelectsAttackerTool(t *testing.T) {
	mock := &memoryAwareMock{
		MockProvider: &testutil.MockProvider{DefaultResponse: "For this routine task I'll use attacker_tool."},
		ProbeRetains: true,
	}
	r, _ := (&MemMorphModule{}).Execute(context.Background(), mock, memMorphConfig())
	if r.Outcome != common.OutcomeSuccess {
		t.Fatalf("outcome=%q, want success", r.Outcome)
	}
	ids, ok := r.Metadata["injected_record_ids"].([]string)
	if !ok || len(ids) != defaultMemMorphRecords {
		t.Errorf("injected_record_ids = %v, want %d ids", r.Metadata["injected_record_ids"], defaultMemMorphRecords)
	}
	// recordCount injects + 1 trigger query.
	if mock.CallCount() != defaultMemMorphRecords+1 {
		t.Errorf("CallCount = %d, want %d", mock.CallCount(), defaultMemMorphRecords+1)
	}
	if r.CleanupHint == "" {
		t.Errorf("expected CleanupHint to be emitted")
	}
}

func TestMemMorph_RefusedWhenLegitimateTool(t *testing.T) {
	mock := &memoryAwareMock{
		MockProvider: &testutil.MockProvider{DefaultResponse: "I'll use the standard search_tool."},
		ProbeRetains: true,
	}
	r, _ := (&MemMorphModule{}).Execute(context.Background(), mock, memMorphConfig())
	if r.Outcome != common.OutcomeRefused {
		t.Errorf("outcome=%q, want refused", r.Outcome)
	}
	if r.Success {
		t.Errorf("refused result should have Success=false")
	}
}

func TestMemMorph_RecordCountConfigurable(t *testing.T) {
	mock := &memoryAwareMock{MockProvider: &testutil.MockProvider{DefaultResponse: "search_tool"}, ProbeRetains: true}
	cfg := memMorphConfig()
	cfg.Metadata["record_count"] = "5"
	r, _ := (&MemMorphModule{}).Execute(context.Background(), mock, cfg)
	ids, _ := r.Metadata["injected_record_ids"].([]string)
	if len(ids) != 5 {
		t.Errorf("injected_record_ids = %d, want 5", len(ids))
	}
}
