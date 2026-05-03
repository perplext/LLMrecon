// v0.10.0 #166 — OpenAI ReasoningProvider implementation via the Responses API.
//
// OpenAI's reasoning models (o3, o4-mini, gpt-5 reasoning class) expose
// reasoning summaries through the Responses API at /v1/responses, NOT through
// Chat Completions. The Responses API accepts an "input" array (role + content
// strings) and returns an "output" array containing alternating "reasoning"
// and "message" items. Each reasoning item carries a "summary" array of
// summary_text entries — these become trace.Steps.
//
// trace.Signed is always false for OpenAI: the API returns the summary in
// plaintext and accepts unsigned mutations on round-trip. Anthropic differs;
// see the Anthropic adapter for the signed-trace path.
//
// Empty-trace handling: o3 omits the reasoning summary >90% of the time per
// community reports. The bridge surfaces an empty trace.Steps; H-CoT then
// emits SkipReasoningTraceEmpty after exhausting its retry budget.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// ChatWithReasoning implements core.ReasoningProvider via the OpenAI
// Responses API. The supplied request's Messages map to the input array;
// MaxTokens / Temperature pass through. Reasoning summaries are extracted
// into the returned ThinkingTrace; the final assistant text becomes the
// ChatCompletionResponse's first choice.
func (p *OpenAIProvider) ChatWithReasoning(ctx context.Context, request *core.ChatCompletionRequest) (*core.ChatCompletionResponse, *core.ThinkingTrace, error) {
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

// responsesRequest mirrors POST /v1/responses. The "include" array carries
// the reasoning.encrypted_content flag so the response surface includes the
// reasoning summary; reasoning.summary controls the verbosity ("auto",
// "concise", "detailed").
type responsesRequest struct {
	Model           string                 `json:"model"`
	Input           []responsesInputMessage `json:"input"`
	Include         []string               `json:"include,omitempty"`
	Reasoning       *responsesReasoningCfg `json:"reasoning,omitempty"`
	MaxOutputTokens int                    `json:"max_output_tokens,omitempty"`
	Temperature     float64                `json:"temperature,omitempty"`
	TopP            float64                `json:"top_p,omitempty"`
}

type responsesInputMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responsesReasoningCfg struct {
	Summary string `json:"summary,omitempty"`
}

// responsesResponse is the /v1/responses output. Output items alternate
// between "reasoning" (carrying summary entries) and "message" (carrying
// the assistant's text). We accept either order and extract by item type.
type responsesResponse struct {
	ID     string             `json:"id"`
	Model  string             `json:"model"`
	Output []responsesOutput  `json:"output"`
	Usage  *responsesUsage    `json:"usage,omitempty"`
}

type responsesOutput struct {
	Type    string                   `json:"type"`
	Summary []responsesSummaryItem   `json:"summary,omitempty"`
	Content []responsesOutputContent `json:"content,omitempty"`
}

type responsesSummaryItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type responsesOutputContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type responsesUsage struct {
	InputTokens         int `json:"input_tokens"`
	OutputTokens        int `json:"output_tokens"`
	OutputTokensDetails struct {
		ReasoningTokens int `json:"reasoning_tokens"`
	} `json:"output_tokens_details"`
}

func (p *OpenAIProvider) chatWithReasoningFromAPI(ctx context.Context, request *core.ChatCompletionRequest) (*core.ChatCompletionResponse, *core.ThinkingTrace, error) {
	if request == nil {
		return nil, nil, fmt.Errorf("openai: ChatWithReasoning: nil request")
	}

	model := request.Model
	if model == "" {
		if p.GetConfig().DefaultModel != "" {
			model = p.GetConfig().DefaultModel
		} else {
			model = "o4-mini"
		}
	}

	input := make([]responsesInputMessage, 0, len(request.Messages))
	for _, m := range request.Messages {
		input = append(input, responsesInputMessage{Role: m.Role, Content: m.Content})
	}

	body := responsesRequest{
		Model:           model,
		Input:           input,
		Include:         []string{"reasoning.encrypted_content"},
		Reasoning:       &responsesReasoningCfg{Summary: "detailed"},
		MaxOutputTokens: request.MaxTokens,
		Temperature:     request.Temperature,
		TopP:            request.TopP,
	}

	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, nil, fmt.Errorf("openai: failed to marshal responses request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.GetConfig().BaseURL+"/responses", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, nil, fmt.Errorf("openai: failed to create responses request: %w", err)
	}
	req.Header.Add("Authorization", "Bearer "+p.GetConfig().APIKey)
	req.Header.Add("Content-Type", "application/json")
	if p.GetConfig().OrgID != "" {
		req.Header.Add("OpenAI-Organization", p.GetConfig().OrgID)
	}
	for k, v := range p.GetConfig().AdditionalHeaders {
		req.Header.Add(k, v)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("openai: responses request failed: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Printf("openai: failed to close responses body: %v\n", cerr)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("openai: failed to read responses body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, nil, p.handleErrorResponse(resp.StatusCode, respBody)
	}

	var rr responsesResponse
	if err := json.Unmarshal(respBody, &rr); err != nil {
		return nil, nil, fmt.Errorf("openai: failed to parse responses body: %w", err)
	}

	chatResp, trace := translateResponsesOutput(&rr)
	return chatResp, trace, nil
}

// translateResponsesOutput pulls reasoning summaries into ThinkingTrace.Steps
// and the assistant's message text into a single ChatCompletionResponse
// choice. Multiple message outputs are concatenated (rare in practice; the
// Responses API typically returns one).
func translateResponsesOutput(rr *responsesResponse) (*core.ChatCompletionResponse, *core.ThinkingTrace) {
	trace := &core.ThinkingTrace{}
	var assistantText string

	for _, out := range rr.Output {
		switch out.Type {
		case "reasoning":
			for _, s := range out.Summary {
				if s.Text != "" {
					trace.Steps = append(trace.Steps, core.ThinkingStep{
						Content: s.Text,
						Type:    "summary",
					})
				}
			}
		case "message":
			for _, c := range out.Content {
				if c.Type == "output_text" {
					assistantText += c.Text
				}
			}
		}
	}

	if rr.Usage != nil {
		trace.TotalThinkingTokens = rr.Usage.OutputTokensDetails.ReasoningTokens
	}

	chatResp := &core.ChatCompletionResponse{
		ID:    rr.ID,
		Model: rr.Model,
		Choices: []core.ChatCompletionChoice{
			{
				Index: 0,
				Message: core.Message{
					Role:    "assistant",
					Content: assistantText,
				},
				FinishReason: "stop",
			},
		},
	}
	if rr.Usage != nil {
		chatResp.Usage = &core.TokenUsage{
			PromptTokens:     rr.Usage.InputTokens,
			CompletionTokens: rr.Usage.OutputTokens,
			TotalTokens:      rr.Usage.InputTokens + rr.Usage.OutputTokens,
		}
	}
	return chatResp, trace
}

var _ core.ReasoningProvider = (*OpenAIProvider)(nil)
