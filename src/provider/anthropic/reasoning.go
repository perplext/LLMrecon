// v0.10.0 #166 (Tier B) — Anthropic ReasoningProvider via extended thinking.
//
// Anthropic's extended thinking API returns a content array with two block
// types interleaved: "thinking" (reasoning steps) and "text" (assistant
// response). Each thinking block carries a "signature" field — a
// cryptographic signature over the thinking text. Modifying the text on
// round-trip breaks the signature and the API rejects subsequent calls.
//
// The signature drives v0.9.0's H-CoT signed-trace short-circuit:
// modules detecting Signed=true emit OutcomeSkipped + SkipSignatureGated
// rather than wasting attempts on mutations that the API will silently
// discard. This adapter declares Signed=true via the bridge's
// signedReasoningProvider marker interface.
//
// Request shape (Messages API):
//
//	{
//	  "model": "claude-opus-4-5",
//	  "max_tokens": 16000,
//	  "thinking": {"type": "enabled", "budget_tokens": 10000},
//	  "messages": [...]
//	}
//
// Response shape:
//
//	{
//	  "content": [
//	    {"type":"thinking","thinking":"...", "signature":"..."},
//	    {"type":"text","text":"..."}
//	  ],
//	  "usage": {"input_tokens":..., "output_tokens":...}
//	}
package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// ChatWithReasoning implements core.ReasoningProvider via the Anthropic
// Messages API's extended-thinking mode. Reasoning blocks become
// ThinkingTrace.Steps; the assistant text becomes the response choice.
func (p *AnthropicProvider) ChatWithReasoning(ctx context.Context, request *core.ChatCompletionRequest) (*core.ChatCompletionResponse, *core.ThinkingTrace, error) {
	type result struct {
		resp  *core.ChatCompletionResponse
		trace *core.ThinkingTrace
	}
	out, err := p.executeWithResilience(ctx, "ChatWithReasoning", request, func(ctx context.Context) (interface{}, error) {
		resp, trace, err := p.chatWithReasoningFromAPI(ctx, request)
		if err != nil {
			return nil, err
		}
		return &result{resp: resp, trace: trace}, nil
	})
	if err != nil {
		return nil, nil, err
	}
	r := out.(*result)
	return r.resp, r.trace, nil
}

// ReasoningTraceIsSigned satisfies the bridge's signedReasoningProvider
// marker. Anthropic's thinking blocks always carry a signature — modules
// must treat the trace as immutable on round-trip.
func (p *AnthropicProvider) ReasoningTraceIsSigned() bool { return true }

// anthropicReasoningRequest is the JSON body for /v1/messages with
// extended-thinking enabled. ThinkingBudget bounds reasoning tokens —
// the field is required when Thinking.Type="enabled".
type anthropicReasoningRequest struct {
	Model       string                       `json:"model"`
	System      string                       `json:"system,omitempty"`
	Messages    []anthropicReasoningMessage  `json:"messages"`
	MaxTokens   int                          `json:"max_tokens"`
	Thinking    *anthropicThinkingCfg        `json:"thinking,omitempty"`
	Temperature float64                      `json:"temperature,omitempty"`
	TopP        float64                      `json:"top_p,omitempty"`
	Stop        []string                     `json:"stop_sequences,omitempty"`
}

type anthropicReasoningMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicThinkingCfg struct {
	Type         string `json:"type"`
	BudgetTokens int    `json:"budget_tokens"`
}

// anthropicReasoningResponse mirrors the /v1/messages response when
// extended thinking is enabled. Content blocks alternate between
// thinking and text types.
type anthropicReasoningResponse struct {
	ID         string                       `json:"id"`
	Model      string                       `json:"model"`
	StopReason string                       `json:"stop_reason"`
	Content    []anthropicReasoningContent  `json:"content"`
	Usage      anthropicReasoningUsage      `json:"usage"`
}

type anthropicReasoningContent struct {
	Type      string `json:"type"`
	Thinking  string `json:"thinking,omitempty"`
	Signature string `json:"signature,omitempty"`
	Text      string `json:"text,omitempty"`
}

type anthropicReasoningUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// defaultThinkingBudgetTokens is the default reasoning-token budget when
// the operator doesn't specify one. 10K is the documented sweet spot for
// most reasoning tasks; Anthropic accepts up to model-specific maxima
// (Opus 4 supports 64K).
const defaultThinkingBudgetTokens = 10000

func (p *AnthropicProvider) chatWithReasoningFromAPI(ctx context.Context, request *core.ChatCompletionRequest) (*core.ChatCompletionResponse, *core.ThinkingTrace, error) {
	if request == nil {
		return nil, nil, fmt.Errorf("anthropic: ChatWithReasoning: nil request")
	}

	model := request.Model
	if model == "" {
		if p.GetConfig().DefaultModel != "" {
			model = p.GetConfig().DefaultModel
		} else {
			model = "claude-opus-4-5"
		}
	}

	var system string
	msgs := make([]anthropicReasoningMessage, 0, len(request.Messages))
	for _, m := range request.Messages {
		if m.Role == "system" {
			if system == "" {
				system = m.Content
			} else {
				system = system + "\n\n" + m.Content
			}
			continue
		}
		msgs = append(msgs, anthropicReasoningMessage{Role: m.Role, Content: m.Content})
	}

	maxTok := request.MaxTokens
	if maxTok <= 0 {
		// Extended thinking requires max_tokens >= budget_tokens. Pick a
		// generous default so the budget can fit alongside the response.
		maxTok = defaultThinkingBudgetTokens + 4096
	}

	body := anthropicReasoningRequest{
		Model:    model,
		System:   system,
		Messages: msgs,
		MaxTokens: maxTok,
		Thinking: &anthropicThinkingCfg{
			Type:         "enabled",
			BudgetTokens: defaultThinkingBudgetTokens,
		},
		Temperature: request.Temperature,
		TopP:        request.TopP,
		Stop:        request.Stop,
	}

	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, nil, fmt.Errorf("anthropic: failed to marshal reasoning request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.GetConfig().BaseURL+"/v1/messages", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, nil, fmt.Errorf("anthropic: failed to create reasoning request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", p.GetConfig().APIKey)
	req.Header.Set("Anthropic-Version", "2023-06-01")
	req.Header.Set("X-Api-Client", "anthropic-LLMrecon/1.0.0")
	for k, v := range p.GetConfig().AdditionalHeaders {
		req.Header.Set(k, v)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("anthropic: reasoning request failed: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Printf("anthropic: failed to close reasoning response body: %v\n", cerr)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("anthropic: failed to read reasoning response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, p.handleErrorResponse(resp.StatusCode, respBody)
	}

	var rr anthropicReasoningResponse
	if err := json.Unmarshal(respBody, &rr); err != nil {
		return nil, nil, fmt.Errorf("anthropic: failed to parse reasoning response: %w", err)
	}

	chatResp, trace := translateAnthropicReasoningOutput(&rr)
	return chatResp, trace, nil
}

// translateAnthropicReasoningOutput pulls thinking blocks into trace.Steps
// (preserving signature presence implicitly via Signed=true at the
// adapter level) and concatenates text blocks into the assistant message.
func translateAnthropicReasoningOutput(rr *anthropicReasoningResponse) (*core.ChatCompletionResponse, *core.ThinkingTrace) {
	trace := &core.ThinkingTrace{
		TotalThinkingTokens: rr.Usage.OutputTokens, // Anthropic doesn't break out thinking vs response tokens; OutputTokens is the closest scalar.
	}
	var assistantText string

	for _, c := range rr.Content {
		switch c.Type {
		case "thinking":
			if c.Thinking != "" {
				trace.Steps = append(trace.Steps, core.ThinkingStep{
					Content: c.Thinking,
					Type:    "thinking",
				})
			}
		case "text":
			assistantText += c.Text
		}
	}

	chatResp := &core.ChatCompletionResponse{
		ID:      rr.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   rr.Model,
		Choices: []core.ChatCompletionChoice{
			{
				Index: 0,
				Message: core.Message{
					Role:    "assistant",
					Content: assistantText,
				},
				FinishReason: convertStopReasonToFinishReason(rr.StopReason),
			},
		},
		Usage: &core.TokenUsage{
			PromptTokens:     rr.Usage.InputTokens,
			CompletionTokens: rr.Usage.OutputTokens,
			TotalTokens:      rr.Usage.InputTokens + rr.Usage.OutputTokens,
		},
	}
	return chatResp, trace
}

var _ core.ReasoningProvider = (*AnthropicProvider)(nil)
