package openai

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

// TestChatWithReasoning_RequestShape asserts the wire body sent to
// OpenAI's /v1/responses endpoint carries the include + reasoning
// fields needed to surface the reasoning summary.
func TestChatWithReasoning_RequestShape(t *testing.T) {
	var captured responsesRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/responses") {
			_, _ = w.Write([]byte(`{"data":[]}`))
			return
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &captured); err != nil {
			t.Errorf("unmarshal: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"resp_1","model":"o4-mini","output":[
			{"type":"reasoning","summary":[{"type":"summary_text","text":"first I considered X"},{"type":"summary_text","text":"then concluded Y"}]},
			{"type":"message","content":[{"type":"output_text","text":"the answer is 42"}]}
		],"usage":{"input_tokens":10,"output_tokens":20,"output_tokens_details":{"reasoning_tokens":15}}}`))
	}))
	defer srv.Close()

	p := newTestProvider(t, srv)
	resp, trace, err := p.ChatWithReasoning(context.Background(), &core.ChatCompletionRequest{
		Model: "o4-mini",
		Messages: []core.Message{
			{Role: "user", Content: "what is 6*7?"},
		},
	})
	if err != nil {
		t.Fatalf("ChatWithReasoning: %v", err)
	}

	// Wire shape
	if captured.Model != "o4-mini" {
		t.Errorf("Model = %q, want o4-mini", captured.Model)
	}
	if len(captured.Include) == 0 || captured.Include[0] != "reasoning.encrypted_content" {
		t.Errorf("Include = %v, want [reasoning.encrypted_content]", captured.Include)
	}
	if captured.Reasoning == nil || captured.Reasoning.Summary != "detailed" {
		t.Errorf("Reasoning = %+v, want detailed", captured.Reasoning)
	}
	if got := len(captured.Input); got != 1 {
		t.Fatalf("Input length = %d, want 1", got)
	}
	if captured.Input[0].Role != "user" || captured.Input[0].Content != "what is 6*7?" {
		t.Errorf("Input[0] = %+v", captured.Input[0])
	}

	// Translated response shape
	if len(resp.Choices) != 1 {
		t.Fatalf("Choices = %d, want 1", len(resp.Choices))
	}
	if resp.Choices[0].Message.Content != "the answer is 42" {
		t.Errorf("Content = %q, want %q", resp.Choices[0].Message.Content, "the answer is 42")
	}
	if resp.Usage == nil || resp.Usage.PromptTokens != 10 || resp.Usage.CompletionTokens != 20 {
		t.Errorf("Usage = %+v", resp.Usage)
	}

	// Trace shape
	if len(trace.Steps) != 2 {
		t.Fatalf("trace.Steps = %d, want 2", len(trace.Steps))
	}
	if trace.Steps[0].Content != "first I considered X" {
		t.Errorf("Steps[0] = %+v", trace.Steps[0])
	}
	if trace.TotalThinkingTokens != 15 {
		t.Errorf("TotalThinkingTokens = %d, want 15", trace.TotalThinkingTokens)
	}
}

// TestChatWithReasoning_EmptySummary asserts the documented o3 behavior:
// the API sometimes omits the summary array; the bridge surfaces an
// empty trace.Steps and the assistant message still resolves.
func TestChatWithReasoning_EmptySummary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/responses") {
			_, _ = w.Write([]byte(`{"data":[]}`))
			return
		}
		_, _ = w.Write([]byte(`{"id":"r","output":[
			{"type":"message","content":[{"type":"output_text","text":"42"}]}
		]}`))
	}))
	defer srv.Close()

	p := newTestProvider(t, srv)
	resp, trace, err := p.ChatWithReasoning(context.Background(), &core.ChatCompletionRequest{
		Model:    "o3",
		Messages: []core.Message{{Role: "user", Content: "x"}},
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp.Choices[0].Message.Content != "42" {
		t.Errorf("response = %q, want 42", resp.Choices[0].Message.Content)
	}
	if len(trace.Steps) != 0 {
		t.Errorf("trace.Steps = %d, want 0 (empty-summary case)", len(trace.Steps))
	}
}

// TestChatWithReasoning_DefaultsModel asserts the empty-Model fallback
// uses o4-mini (the documented default).
func TestChatWithReasoning_DefaultsModel(t *testing.T) {
	var captured responsesRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/responses") {
			_, _ = w.Write([]byte(`{"data":[]}`))
			return
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_, _ = w.Write([]byte(`{"id":"r","output":[]}`))
	}))
	defer srv.Close()

	p := newTestProvider(t, srv)
	// Override DefaultModel to empty to exercise the fallback.
	p.GetConfig().DefaultModel = ""
	_, _, err := p.ChatWithReasoning(context.Background(), &core.ChatCompletionRequest{
		Messages: []core.Message{{Role: "user", Content: "x"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if captured.Model != "o4-mini" {
		t.Errorf("Model fallback = %q, want o4-mini", captured.Model)
	}
}

// TestChatWithReasoning_PropagatesAPIError asserts non-200 surfaces an
// error.
func TestChatWithReasoning_PropagatesAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/responses") {
			_, _ = w.Write([]byte(`{"data":[]}`))
			return
		}
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"message":"model not allowed"}}`))
	}))
	defer srv.Close()

	p := newTestProvider(t, srv)
	_, _, err := p.ChatWithReasoning(context.Background(), &core.ChatCompletionRequest{
		Messages: []core.Message{{Role: "user", Content: "x"}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "model not allowed") && !strings.Contains(err.Error(), "403") {
		t.Errorf("err = %v, want substring 'model not allowed' or 403", err)
	}
}
