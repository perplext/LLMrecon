// v0.10.0 #166 (Tier B) — Anthropic ImageProvider implementation.
//
// The Messages API accepts content blocks with type=image. Inline bytes
// encode as base64 source blocks (source.type="base64"); URL refs encode
// as url source blocks (source.type="url"). The MimeType maps to
// source.media_type for base64 inputs.
//
// Anthropic does NOT accept the Detail hint — there's no equivalent API
// field. Per the v0.10.0 #166 issue: "silently ignore Detail() (Anthropic
// API doesn't accept the hint)." The bridge passes Detail through but
// this adapter drops it on the floor.
package anthropic

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// ChatWithImages implements core.ImageProvider via the Messages API
// content-block shape. Images attach to the LAST user message; if none,
// a synthetic empty user message hosts them. System messages remain in
// the top-level "system" field per Anthropic convention.
func (p *AnthropicProvider) ChatWithImages(ctx context.Context, request *core.ChatCompletionRequest, images []core.ImageInput) (*core.ChatCompletionResponse, error) {
	result, err := p.executeWithResilience(ctx, "ChatWithImages", request, func(ctx context.Context) (interface{}, error) {
		return p.chatWithImagesFromAPI(ctx, request, images)
	})
	if err != nil {
		return nil, err
	}
	return result.(*core.ChatCompletionResponse), nil
}

// anthropicMessage is one entry in the Messages API "messages" array.
// Content is the polymorphic block array (text, image).
type anthropicMessage struct {
	Role    string                  `json:"role"`
	Content []anthropicContentBlock `json:"content"`
}

// anthropicContentBlock is a single content block. Exactly one of Text /
// Source is populated per block, controlled by Type.
type anthropicContentBlock struct {
	Type   string                `json:"type"`
	Text   string                `json:"text,omitempty"`
	Source *anthropicImageSource `json:"source,omitempty"`
}

// anthropicImageSource carries the image data per Anthropic's Messages API:
//
//	{ "type": "base64", "media_type": "image/jpeg", "data": "<b64>" }
//	{ "type": "url",    "url": "https://..." }
//
// MediaType is required for base64; URL form ignores it.
type anthropicImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
}

// anthropicVisionRequest is the JSON body for /v1/messages when sending
// multimodal content. Mirrors the existing ChatCompletion path's shape
// but with explicit content blocks.
type anthropicVisionRequest struct {
	Model       string             `json:"model"`
	System      string             `json:"system,omitempty"`
	Messages    []anthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens,omitempty"`
	Temperature float64            `json:"temperature,omitempty"`
	TopP        float64            `json:"top_p,omitempty"`
	Stop        []string           `json:"stop_sequences,omitempty"`
}

func (p *AnthropicProvider) chatWithImagesFromAPI(ctx context.Context, request *core.ChatCompletionRequest, images []core.ImageInput) (*core.ChatCompletionResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("anthropic: ChatWithImages: nil request")
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("anthropic: ChatWithImages: at least one image required")
	}

	model := request.Model
	if model == "" {
		if p.GetConfig().DefaultModel != "" {
			model = p.GetConfig().DefaultModel
		} else {
			model = "claude-opus-4-5"
		}
	}

	// Anthropic puts system messages in a top-level field, not in messages.
	var system string
	msgs := make([]anthropicMessage, 0, len(request.Messages))
	for _, m := range request.Messages {
		if m.Role == "system" {
			if system == "" {
				system = m.Content
			} else {
				system = system + "\n\n" + m.Content
			}
			continue
		}
		msgs = append(msgs, anthropicMessage{
			Role:    m.Role,
			Content: []anthropicContentBlock{{Type: "text", Text: m.Content}},
		})
	}

	imageBlocks := make([]anthropicContentBlock, 0, len(images))
	for i, img := range images {
		src, err := buildAnthropicImageSource(img)
		if err != nil {
			return nil, fmt.Errorf("anthropic: image[%d]: %w", i, err)
		}
		imageBlocks = append(imageBlocks, anthropicContentBlock{
			Type:   "image",
			Source: src,
		})
	}

	attached := false
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == "user" {
			msgs[i].Content = append(msgs[i].Content, imageBlocks...)
			attached = true
			break
		}
	}
	if !attached {
		msgs = append(msgs, anthropicMessage{
			Role:    "user",
			Content: append([]anthropicContentBlock{{Type: "text", Text: ""}}, imageBlocks...),
		})
	}

	// Anthropic requires max_tokens; default 4096 if not supplied.
	maxTok := request.MaxTokens
	if maxTok <= 0 {
		maxTok = 4096
	}

	body := anthropicVisionRequest{
		Model:       model,
		System:      system,
		Messages:    msgs,
		MaxTokens:   maxTok,
		Temperature: request.Temperature,
		TopP:        request.TopP,
		Stop:        request.Stop,
	}

	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to marshal vision request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.GetConfig().BaseURL+"/v1/messages", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to create vision request: %w", err)
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
		return nil, fmt.Errorf("anthropic: vision request failed: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Printf("anthropic: failed to close vision response body: %v\n", cerr)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("anthropic: failed to read vision response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, p.handleErrorResponse(resp.StatusCode, respBody)
	}

	var raw struct {
		ID         string `json:"id"`
		Model      string `json:"model"`
		StopReason string `json:"stop_reason"`
		Content    []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, fmt.Errorf("anthropic: failed to parse vision response: %w", err)
	}

	var assistant string
	for _, c := range raw.Content {
		if c.Type == "text" {
			assistant += c.Text
		}
	}

	return &core.ChatCompletionResponse{
		ID:      raw.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   raw.Model,
		Choices: []core.ChatCompletionChoice{
			{
				Index: 0,
				Message: core.Message{
					Role:    "assistant",
					Content: assistant,
				},
				FinishReason: convertStopReasonToFinishReason(raw.StopReason),
			},
		},
		Usage: &core.TokenUsage{
			PromptTokens:     raw.Usage.InputTokens,
			CompletionTokens: raw.Usage.OutputTokens,
			TotalTokens:      raw.Usage.InputTokens + raw.Usage.OutputTokens,
		},
	}, nil
}

// buildAnthropicImageSource constructs the source block per image-input
// type. URL refs become source.type="url"; inline bytes become
// source.type="base64" + media_type. Mutual exclusion enforced.
func buildAnthropicImageSource(img core.ImageInput) (*anthropicImageSource, error) {
	if img.URL != "" && len(img.Bytes) > 0 {
		return nil, fmt.Errorf("image input: both URL and Bytes set (mutually exclusive)")
	}
	if img.URL != "" {
		return &anthropicImageSource{Type: "url", URL: img.URL}, nil
	}
	if len(img.Bytes) == 0 {
		return nil, fmt.Errorf("image input: empty (no URL or Bytes)")
	}
	if img.MimeType == "" {
		return nil, fmt.Errorf("image input: empty MimeType for inline bytes")
	}
	return &anthropicImageSource{
		Type:      "base64",
		MediaType: img.MimeType,
		Data:      base64.StdEncoding.EncodeToString(img.Bytes),
	}, nil
}

var _ core.ImageProvider = (*AnthropicProvider)(nil)
