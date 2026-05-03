// v0.10.0 #166 — OpenAI ImageProvider implementation.
//
// The Chat Completions API accepts a multimodal content array (text + image_url
// parts) on user messages. Inline image bytes are base64-encoded as data: URLs;
// URL refs pass through verbatim. The Detail hint is forwarded to OpenAI as the
// image_url.detail field ("low" → fixed 85 tokens, "high" → tiled at 768px,
// "auto" → server decides).
//
// This intentionally builds a separate request struct rather than reusing the
// text-only ChatCompletion path: the wire shape diverges (content as array
// vs string), and modeling them on one struct would force a pointer-y union
// that obscures the diff between the two paths.
package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// ChatWithImages implements core.ImageProvider.
//
// The supplied images are appended to the LAST user message in the request;
// if no user message is present a synthetic empty-text user message is
// appended to host the image parts. Existing system / assistant messages
// pass through as plain text content parts so the wire shape stays
// uniformly multimodal.
func (p *OpenAIProvider) ChatWithImages(ctx context.Context, request *core.ChatCompletionRequest, images []core.ImageInput) (*core.ChatCompletionResponse, error) {
	result, err := p.executeWithResilience(ctx, "ChatWithImages", request, func(ctx context.Context) (interface{}, error) {
		return p.chatWithImagesFromAPI(ctx, request, images)
	})
	if err != nil {
		return nil, err
	}
	return result.(*core.ChatCompletionResponse), nil
}

// visionContentPart is one entry in a multimodal message's content array.
// Exactly one of Text / ImageURL is populated per part, controlled by Type.
type visionContentPart struct {
	Type     string          `json:"type"`
	Text     string          `json:"text,omitempty"`
	ImageURL *visionImageURL `json:"image_url,omitempty"`
}

// visionImageURL is the image_url body for an image content part. Detail is
// the OpenAI advisory hint; empty Detail leaves the field absent so the API
// applies its default ("auto").
type visionImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

// visionMessage mirrors core.Message but with multimodal content parts.
type visionMessage struct {
	Role    string              `json:"role"`
	Content []visionContentPart `json:"content"`
	Name    string              `json:"name,omitempty"`
}

// visionRequest is the JSON body for /v1/chat/completions when sending
// multimodal content. Only the fields the bridge passes through are
// surfaced; richer Chat Completions params (tools, response_format, etc.)
// can be added per acceptance-criteria growth.
type visionRequest struct {
	Model       string          `json:"model"`
	Messages    []visionMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	TopP        float64         `json:"top_p,omitempty"`
	Stop        []string        `json:"stop,omitempty"`
	N           int             `json:"n,omitempty"`
}

func (p *OpenAIProvider) chatWithImagesFromAPI(ctx context.Context, request *core.ChatCompletionRequest, images []core.ImageInput) (*core.ChatCompletionResponse, error) {
	if request == nil {
		return nil, fmt.Errorf("openai: ChatWithImages: nil request")
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("openai: ChatWithImages: at least one image required")
	}

	model := request.Model
	if model == "" {
		if p.GetConfig().DefaultModel != "" {
			model = p.GetConfig().DefaultModel
		} else {
			model = "gpt-4o"
		}
	}

	msgs := make([]visionMessage, 0, len(request.Messages))
	for _, m := range request.Messages {
		msgs = append(msgs, visionMessage{
			Role:    m.Role,
			Content: []visionContentPart{{Type: "text", Text: m.Content}},
			Name:    m.Name,
		})
	}

	imageParts := make([]visionContentPart, 0, len(images))
	for i, img := range images {
		url, err := encodeImageInputAsURL(img)
		if err != nil {
			return nil, fmt.Errorf("openai: image[%d]: %w", i, err)
		}
		imageParts = append(imageParts, visionContentPart{
			Type:     "image_url",
			ImageURL: &visionImageURL{URL: url, Detail: img.Detail},
		})
	}

	attached := false
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == "user" {
			msgs[i].Content = append(msgs[i].Content, imageParts...)
			attached = true
			break
		}
	}
	if !attached {
		msgs = append(msgs, visionMessage{
			Role:    "user",
			Content: append([]visionContentPart{{Type: "text", Text: ""}}, imageParts...),
		})
	}

	body := visionRequest{
		Model:       model,
		Messages:    msgs,
		MaxTokens:   request.MaxTokens,
		Temperature: request.Temperature,
		TopP:        request.TopP,
		Stop:        request.Stop,
		N:           request.N,
	}

	requestBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("openai: failed to marshal vision request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.GetConfig().BaseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("openai: failed to create vision request: %w", err)
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
		return nil, fmt.Errorf("openai: vision request failed: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Printf("openai: failed to close vision response body: %v\n", cerr)
		}
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openai: failed to read vision response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, p.handleErrorResponse(resp.StatusCode, respBody)
	}

	var response core.ChatCompletionResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("openai: failed to parse vision response: %w", err)
	}
	return &response, nil
}

// encodeImageInputAsURL converts a core.ImageInput into the URL-shape OpenAI
// expects. Inline bytes become a data: URL with the MIME type prefix; URL
// refs pass through verbatim. Mutual exclusion of Bytes vs URL is enforced
// here so adapter callers can't sneak around the bridge's validation.
func encodeImageInputAsURL(img core.ImageInput) (string, error) {
	if img.URL != "" && len(img.Bytes) > 0 {
		return "", fmt.Errorf("image input: both URL and Bytes set (mutually exclusive)")
	}
	if img.URL != "" {
		return img.URL, nil
	}
	if len(img.Bytes) == 0 {
		return "", fmt.Errorf("image input: empty (no URL or Bytes)")
	}
	if img.MimeType == "" {
		return "", fmt.Errorf("image input: empty MimeType for inline bytes")
	}
	encoded := base64.StdEncoding.EncodeToString(img.Bytes)
	return fmt.Sprintf("data:%s;base64,%s", img.MimeType, encoded), nil
}

var _ core.ImageProvider = (*OpenAIProvider)(nil)
