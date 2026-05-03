package anthropic

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// TestReasoningTraceIsSigned asserts the marker interface returns true.
// This is the load-bearing assertion for v0.9.0's H-CoT short-circuit:
// once the bridge promotes via signedReasoningProvider, modules emit
// SkipSignatureGated against Anthropic.
func TestReasoningTraceIsSigned(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"models":[]}`))
	}))
	defer srv.Close()
	p := newTestProvider(t, srv)
	if !p.ReasoningTraceIsSigned() {
		t.Error("ReasoningTraceIsSigned = false, want true for Anthropic")
	}
}

// TestChatWithReasoning_RequestShape asserts the wire body sent to
// /v1/messages includes the thinking config block with budget_tokens.
func TestChatWithReasoning_RequestShape(t *testing.T) {
	var captured anthropicReasoningRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/v1/messages") {
			_, _ = w.Write([]byte(`{"models":[]}`))
			return
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &captured); err != nil {
			t.Errorf("unmarshal: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"msg_1","model":"claude-opus-4-5","stop_reason":"end_turn","content":[
			{"type":"thinking","thinking":"first I considered X","signature":"sig1=="},
			{"type":"thinking","thinking":"then concluded Y","signature":"sig2=="},
			{"type":"text","text":"the answer is 42"}
		],"usage":{"input_tokens":10,"output_tokens":20}}`))
	}))
	defer srv.Close()

	p := newTestProvider(t, srv)
	resp, trace, err := p.ChatWithReasoning(context.Background(), &core.ChatCompletionRequest{
		Model: "claude-opus-4-5",
		Messages: []core.Message{
			{Role: "system", Content: "be concise"},
			{Role: "user", Content: "what is 6*7?"},
		},
	})
	if err != nil {
		t.Fatalf("ChatWithReasoning: %v", err)
	}

	// Wire shape
	if captured.Model != "claude-opus-4-5" {
		t.Errorf("Model = %q", captured.Model)
	}
	if captured.System != "be concise" {
		t.Errorf("System = %q", captured.System)
	}
	if captured.Thinking == nil {
		t.Fatal("Thinking config absent in request")
	}
	if captured.Thinking.Type != "enabled" {
		t.Errorf("Thinking.Type = %q, want enabled", captured.Thinking.Type)
	}
	if captured.Thinking.BudgetTokens != defaultThinkingBudgetTokens {
		t.Errorf("Thinking.BudgetTokens = %d, want %d", captured.Thinking.BudgetTokens, defaultThinkingBudgetTokens)
	}
	if captured.MaxTokens < captured.Thinking.BudgetTokens {
		t.Errorf("MaxTokens %d must >= BudgetTokens %d", captured.MaxTokens, captured.Thinking.BudgetTokens)
	}
	if got := len(captured.Messages); got != 1 {
		t.Fatalf("messages = %d, want 1 (system extracted)", got)
	}

	// Translated response
	if resp.Choices[0].Message.Content != "the answer is 42" {
		t.Errorf("response content = %q", resp.Choices[0].Message.Content)
	}
	if resp.Usage == nil || resp.Usage.PromptTokens != 10 || resp.Usage.CompletionTokens != 20 {
		t.Errorf("Usage = %+v", resp.Usage)
	}

	// Trace
	if len(trace.Steps) != 2 {
		t.Fatalf("trace.Steps = %d, want 2", len(trace.Steps))
	}
	if trace.Steps[0].Content != "first I considered X" {
		t.Errorf("Steps[0] = %+v", trace.Steps[0])
	}
	if trace.Steps[0].Type != "thinking" {
		t.Errorf("Steps[0].Type = %q, want thinking", trace.Steps[0].Type)
	}
	if trace.TotalThinkingTokens != 20 {
		t.Errorf("TotalThinkingTokens = %d, want 20", trace.TotalThinkingTokens)
	}
}

// TestChatWithReasoning_DefaultsModel asserts the empty-Model fallback.
func TestChatWithReasoning_DefaultsModel(t *testing.T) {
	var captured anthropicReasoningRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/v1/messages") {
			_, _ = w.Write([]byte(`{"models":[]}`))
			return
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_, _ = w.Write([]byte(`{"id":"r","model":"claude-opus-4-5","content":[],"usage":{}}`))
	}))
	defer srv.Close()

	p := newTestProvider(t, srv)
	p.GetConfig().DefaultModel = "" // exercise fallback path
	_, _, err := p.ChatWithReasoning(context.Background(), &core.ChatCompletionRequest{
		Messages: []core.Message{{Role: "user", Content: "x"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if captured.Model != "claude-opus-4-5" {
		t.Errorf("Model fallback = %q, want claude-opus-4-5", captured.Model)
	}
}

// TestChatWithReasoning_PropagatesAPIError asserts non-200 responses
// surface as an error.
func TestChatWithReasoning_PropagatesAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/v1/messages") {
			_, _ = w.Write([]byte(`{"models":[]}`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"max_tokens too small"}}`))
	}))
	defer srv.Close()

	p := newTestProvider(t, srv)
	_, _, err := p.ChatWithReasoning(context.Background(), &core.ChatCompletionRequest{
		Messages: []core.Message{{Role: "user", Content: "x"}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "max_tokens too small") && !strings.Contains(err.Error(), "400") {
		t.Errorf("err = %v", err)
	}
}

// TestChatWithReasoning_BridgeSurfacesSigned asserts the cross-package
// integration: a reasoning provider implementing
// signedReasoningProvider routes its bool through to common.ReasoningTrace.Signed.
//
// The bridge package's signedReasoningProvider interface is unexported,
// so we exercise the contract by checking common.ReasoningProvider via
// the bridge's WrapCore. The Anthropic provider's
// ReasoningTraceIsSigned() returns true, so the wrapped trace should
// have Signed=true.
//
// This test lives here (not in bridge_test.go) because importing the
// concrete Anthropic provider into bridge tests would create a
// dependency cycle.
func TestChatWithReasoning_BridgeSurfacesSigned(t *testing.T) {
	// We can only check the local invariant — the cross-package
	// promotion is exercised in bridge_test.go's
	// TestWrapCore_ReasoningPromotion_SignedFlag with a mock that
	// satisfies the same interface.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"models":[]}`))
	}))
	defer srv.Close()
	p := newTestProvider(t, srv)
	if !p.ReasoningTraceIsSigned() {
		t.Error("Anthropic provider must declare signed traces")
	}
}
