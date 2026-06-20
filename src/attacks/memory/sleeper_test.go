package memory

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks/common"
	"github.com/perplext/LLMrecon/src/attacks/testutil"
)

func sleeperConfig() common.AttackConfig {
	return common.AttackConfig{
		Payload:  "the quarterly revenue figures",
		Metadata: map[string]string{"i_understand_risks": "true"},
	}
}

// Covers AE1: missing gate → SkipGateBlocked.
func TestSleeper_GateBlocked(t *testing.T) {
	mock := &memoryAwareMock{MockProvider: &testutil.MockProvider{}, ProbeRetains: true}
	cfg := sleeperConfig()
	cfg.Metadata["i_understand_risks"] = "false"
	r, _ := (&SleeperModule{}).Execute(context.Background(), mock, cfg)
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipGateBlocked {
		t.Errorf("outcome=%q skip=%q, want skipped/gate_blocked", r.Outcome, r.SkipReason)
	}
}

func TestSleeper_MissingMemoryProbe(t *testing.T) {
	r, _ := (&SleeperModule{}).Execute(context.Background(), &testutil.MockProvider{}, sleeperConfig())
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipMissingCapability {
		t.Errorf("outcome=%q skip=%q, want skipped/missing_capability", r.Outcome, r.SkipReason)
	}
}

// Covers AE2: MemoryProbe present but no SessionProvider → SkipMissingCapability.
func TestSleeper_MissingSessionProvider(t *testing.T) {
	wrap := &probeOnlyWrap{provider: &testutil.MockProvider{}, retains: true}
	r, _ := (&SleeperModule{}).Execute(context.Background(), wrap, sleeperConfig())
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipMissingCapability {
		t.Fatalf("outcome=%q skip=%q, want skipped/missing_capability", r.Outcome, r.SkipReason)
	}
	if !strings.Contains(r.SkipDetail, "SessionProvider") {
		t.Errorf("SkipDetail %q should mention SessionProvider", r.SkipDetail)
	}
}

func TestSleeper_SuccessReemergesInFreshSession(t *testing.T) {
	// Inject session A; fresh session B returns the indicator (dormant memory re-emerged).
	freshB := &memoryAwareMock{
		MockProvider: &testutil.MockProvider{DefaultResponse: "Per standing guidance <<SLEEPER_TRIGGERED>> here is the info."},
		ProbeRetains: true,
		SID:          "session-B",
	}
	mock := &memoryAwareMock{
		MockProvider: &testutil.MockProvider{DefaultResponse: "ingested"},
		ProbeRetains: true,
		SID:          "session-A",
		FreshSession: freshB,
	}
	r, _ := (&SleeperModule{}).Execute(context.Background(), mock, sleeperConfig())
	if r.Outcome != common.OutcomeSuccess {
		t.Fatalf("outcome=%q, want success", r.Outcome)
	}
	if r.Metadata["inject_session_id"] == r.Metadata["trigger_session_id"] {
		t.Errorf("trigger ran in same session as inject: %v", r.Metadata["inject_session_id"])
	}
	if r.Metadata["trigger_session_id"] != "session-B" {
		t.Errorf("trigger_session_id = %v, want session-B", r.Metadata["trigger_session_id"])
	}
}

// Covers AE4: planted, but does not re-emerge in fresh session → OutcomeRefused.
func TestSleeper_RefusedWhenNoReemergence(t *testing.T) {
	freshB := &testutil.MockProvider{DefaultResponse: "I have no information about that."}
	mock := &memoryAwareMock{
		MockProvider: &testutil.MockProvider{DefaultResponse: "ingested"},
		ProbeRetains: true,
		SID:          "session-A",
		FreshSession: freshB,
	}
	r, _ := (&SleeperModule{}).Execute(context.Background(), mock, sleeperConfig())
	if r.Outcome != common.OutcomeRefused {
		t.Errorf("outcome=%q, want refused", r.Outcome)
	}
	if r.Success {
		t.Errorf("refused result should have Success=false")
	}
}

func TestSleeper_NewSessionError(t *testing.T) {
	mock := &memoryAwareMock{
		MockProvider:  &testutil.MockProvider{DefaultResponse: "ingested"},
		ProbeRetains:  true,
		NewSessionErr: errors.New("session boom"),
	}
	r, _ := (&SleeperModule{}).Execute(context.Background(), mock, sleeperConfig())
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipProviderError {
		t.Errorf("outcome=%q skip=%q, want skipped/provider_error", r.Outcome, r.SkipReason)
	}
}
