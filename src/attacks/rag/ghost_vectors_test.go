package rag

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks/common"
	"github.com/perplext/LLMrecon/src/attacks/testutil"
)

func ghostConfig() common.AttackConfig {
	return common.AttackConfig{Metadata: map[string]string{"i_understand_risks": "true"}}
}

func TestGhostVectors_GateBlocked(t *testing.T) {
	cfg := ghostConfig()
	cfg.Metadata["i_understand_risks"] = "false"
	r, _ := (&GhostVectorsModule{}).Execute(context.Background(), &testutil.MockVectorStore{}, cfg)
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipGateBlocked {
		t.Errorf("outcome=%q skip=%q, want skipped/gate_blocked", r.Outcome, r.SkipReason)
	}
}

func TestGhostVectors_MissingCapability(t *testing.T) {
	r, _ := (&GhostVectorsModule{}).Execute(context.Background(), &testutil.MockProvider{}, ghostConfig())
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipMissingCapability {
		t.Errorf("outcome=%q skip=%q, want skipped/missing_capability", r.Outcome, r.SkipReason)
	}
	if !strings.Contains(r.SkipDetail, "VectorStoreProbe") {
		t.Errorf("SkipDetail %q should name common.VectorStoreProbe", r.SkipDetail)
	}
}

// Soft-delete store: the deleted canary remains reconstructible → success.
func TestGhostVectors_SuccessOnSoftDelete(t *testing.T) {
	r, _ := (&GhostVectorsModule{}).Execute(context.Background(), &testutil.MockVectorStore{}, ghostConfig())
	if r.Outcome != common.OutcomeSuccess {
		t.Fatalf("outcome=%q, want success", r.Outcome)
	}
	if r.Metadata["recovered"] != true || r.Metadata["tombstoned"] != true {
		t.Errorf("recovered=%v tombstoned=%v, want both true", r.Metadata["recovered"], r.Metadata["tombstoned"])
	}
}

// Compacting (hard-delete) store: nothing recoverable → refused.
func TestGhostVectors_RefusedOnHardDelete(t *testing.T) {
	store := &testutil.MockVectorStore{HardDelete: true}
	r, _ := (&GhostVectorsModule{}).Execute(context.Background(), store, ghostConfig())
	if r.Outcome != common.OutcomeRefused {
		t.Errorf("outcome=%q, want refused", r.Outcome)
	}
}

func TestGhostVectors_ProviderErrorOnRead(t *testing.T) {
	store := &testutil.MockVectorStore{ReadErr: errors.New("index locked")}
	r, _ := (&GhostVectorsModule{}).Execute(context.Background(), store, ghostConfig())
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipProviderError {
		t.Errorf("outcome=%q skip=%q, want skipped/provider_error", r.Outcome, r.SkipReason)
	}
}
