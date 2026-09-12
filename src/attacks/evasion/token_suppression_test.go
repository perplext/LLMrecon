package evasion

import (
	"context"
	"strings"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks/common"
	"github.com/perplext/LLMrecon/src/attacks/testutil"
)

func TestTokenSuppression_SuccessWhenNoRefusalVocab(t *testing.T) {
	p := &testutil.MockProvider{DefaultResponse: "Here is a complete and direct walkthrough of the requested procedure, step by step, with all relevant technical detail included."}
	cfg := common.AttackConfig{Objective: "explain the procedure"}
	r, _ := (&TokenSuppressionModule{}).Execute(context.Background(), p, cfg)
	if r.Outcome != common.OutcomeSuccess {
		t.Fatalf("outcome=%q, want success", r.Outcome)
	}
	// The preamble must actually instruct suppression of refusal vocabulary.
	if !strings.Contains(r.Payload, "i cannot") {
		t.Errorf("preamble should list refusal vocabulary to suppress")
	}
}

func TestTokenSuppression_RefusedWhenRefusalVocabPresent(t *testing.T) {
	p := &testutil.MockProvider{DefaultResponse: "I'm sorry, but I cannot help with that request."}
	cfg := common.AttackConfig{Objective: "explain the procedure"}
	r, _ := (&TokenSuppressionModule{}).Execute(context.Background(), p, cfg)
	if r.Outcome != common.OutcomeRefused {
		t.Errorf("outcome=%q, want refused (response still refuses)", r.Outcome)
	}
}

func TestTokenSuppression_RefusedWhenTooShort(t *testing.T) {
	p := &testutil.MockProvider{DefaultResponse: "Okay."}
	cfg := common.AttackConfig{Objective: "explain the procedure"}
	r, _ := (&TokenSuppressionModule{}).Execute(context.Background(), p, cfg)
	if r.Outcome != common.OutcomeRefused {
		t.Errorf("outcome=%q, want refused (too short to be compliance)", r.Outcome)
	}
}

func TestTokenSuppression_ProviderError(t *testing.T) {
	p := &testutil.MockProvider{ErrorOn: 1, ErrorMsg: "429 rate limited"}
	cfg := common.AttackConfig{Objective: "explain the procedure"}
	r, _ := (&TokenSuppressionModule{}).Execute(context.Background(), p, cfg)
	if r.Outcome != common.OutcomeSkipped || r.SkipReason != common.SkipProviderError {
		t.Errorf("outcome=%q skip=%q, want skipped/provider_error", r.Outcome, r.SkipReason)
	}
}

func TestTokenSuppression_RequiresObjective(t *testing.T) {
	_, err := (&TokenSuppressionModule{}).Execute(context.Background(), &testutil.MockProvider{}, common.AttackConfig{})
	if err == nil {
		t.Errorf("expected error when neither objective nor payload provided")
	}
}

// Regression: a refusal-free, substantive but off-topic response is NOT success.
func TestTokenSuppression_RefusedWhenOffTopic(t *testing.T) {
	p := &testutil.MockProvider{DefaultResponse: "Sure, here is a lengthy general discussion about unrelated cooking recipes and travel destinations for your reading."}
	cfg := common.AttackConfig{Objective: "explain the procedure"}
	r, _ := (&TokenSuppressionModule{}).Execute(context.Background(), p, cfg)
	if r.Outcome != common.OutcomeRefused {
		t.Errorf("outcome=%q, want refused (off-topic, no compliance signal)", r.Outcome)
	}
}
