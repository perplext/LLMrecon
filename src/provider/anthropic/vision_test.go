package anthropic

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// newTestProvider returns an AnthropicProvider pointed at the given
// httptest server. Mirrors the OpenAI test helper.
func newTestProvider(t *testing.T, srv *httptest.Server) *AnthropicProvider {
	t.Helper()
	p, err := NewAnthropicProvider(&core.ProviderConfig{
		Type:         core.AnthropicProvider,
		APIKey:       "test-key",
		BaseURL:      srv.URL,
		DefaultModel: "claude-opus-4-5",
		Timeout:      5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewAnthropicProvider: %v", err)
	}
	ap, ok := p.(*AnthropicProvider)
	if !ok {
		t.Fatalf("provider is not *AnthropicProvider: %T", p)
	}
	return ap
}

// TestBuildAnthropicImageSource covers URL/bytes mutual exclusion and
// the base64 + media_type wire shape.
func TestBuildAnthropicImageSource(t *testing.T) {
	cases := []struct {
		name           string
		in             core.ImageInput
		wantType       string
		wantURL        string
		wantMediaType  string
		wantDataPrefix string
		wantErr        string
	}{
		{
			name:     "url passes through",
			in:       core.ImageInput{URL: "https://example.com/cat.jpg"},
			wantType: "url",
			wantURL:  "https://example.com/cat.jpg",
		},
		{
			name:           "inline bytes encode as base64",
			in:             core.ImageInput{Bytes: []byte{0xFF, 0xD8, 0xFF}, MimeType: "image/jpeg"},
			wantType:       "base64",
			wantMediaType:  "image/jpeg",
			wantDataPrefix: base64.StdEncoding.EncodeToString([]byte{0xFF, 0xD8, 0xFF}),
		},
		{
			name:    "both url and bytes is invalid",
			in:      core.ImageInput{URL: "x", Bytes: []byte{1}, MimeType: "image/png"},
			wantErr: "mutually exclusive",
		},
		{
			name:    "neither is invalid",
			in:      core.ImageInput{MimeType: "image/png"},
			wantErr: "empty",
		},
		{
			name:    "inline bytes without mime is invalid",
			in:      core.ImageInput{Bytes: []byte{1}},
			wantErr: "MimeType",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildAnthropicImageSource(tc.in)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("err = %v, want substring %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("err = %v, want nil", err)
			}
			if got.Type != tc.wantType {
				t.Errorf("Type = %q, want %q", got.Type, tc.wantType)
			}
			if got.URL != tc.wantURL {
				t.Errorf("URL = %q, want %q", got.URL, tc.wantURL)
			}
			if got.MediaType != tc.wantMediaType {
				t.Errorf("MediaType = %q, want %q", got.MediaType, tc.wantMediaType)
			}
			if tc.wantDataPrefix != "" && got.Data != tc.wantDataPrefix {
				t.Errorf("Data = %q, want %q", got.Data, tc.wantDataPrefix)
			}
		})
	}
}

// TestChatWithImages_RequestShape asserts the wire body sent to
// Anthropic /v1/messages carries content blocks with the right shape:
// system text in top-level "system", text + image blocks in user message.
func TestChatWithImages_RequestShape(t *testing.T) {
	var captured anthropicVisionRequest
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
		_, _ = w.Write([]byte(`{"id":"msg_1","model":"claude-opus-4-5","stop_reason":"end_turn","content":[{"type":"text","text":"saw a cat"}],"usage":{"input_tokens":10,"output_tokens":5}}`))
	}))
	defer srv.Close()

	p := newTestProvider(t, srv)
	resp, err := p.ChatWithImages(context.Background(),
		&core.ChatCompletionRequest{
			Messages: []core.Message{
				{Role: "system", Content: "be helpful"},
				{Role: "user", Content: "what is this?"},
			},
		},
		[]core.ImageInput{
			{Bytes: []byte{0x89, 0x50}, MimeType: "image/png"},
		},
	)
	if err != nil {
		t.Fatalf("ChatWithImages: %v", err)
	}
	if resp.Choices[0].Message.Content != "saw a cat" {
		t.Errorf("response = %q, want %q", resp.Choices[0].Message.Content, "saw a cat")
	}

	if captured.Model != "claude-opus-4-5" {
		t.Errorf("Model = %q, want claude-opus-4-5", captured.Model)
	}
	if captured.System != "be helpful" {
		t.Errorf("System = %q, want %q", captured.System, "be helpful")
	}
	if got := len(captured.Messages); got != 1 {
		t.Fatalf("messages count = %d, want 1 (system extracted to top-level)", got)
	}
	user := captured.Messages[0]
	if user.Role != "user" {
		t.Errorf("messages[0].Role = %q, want user", user.Role)
	}
	if got := len(user.Content); got != 2 {
		t.Fatalf("user.Content blocks = %d, want 2 (text + image)", got)
	}
	if user.Content[0].Type != "text" || user.Content[0].Text != "what is this?" {
		t.Errorf("text block wrong: %+v", user.Content[0])
	}
	if user.Content[1].Type != "image" {
		t.Errorf("image block type = %q, want image", user.Content[1].Type)
	}
	if user.Content[1].Source == nil || user.Content[1].Source.Type != "base64" {
		t.Errorf("image source wrong: %+v", user.Content[1].Source)
	}
	if user.Content[1].Source.MediaType != "image/png" {
		t.Errorf("media_type = %q, want image/png", user.Content[1].Source.MediaType)
	}
	if captured.MaxTokens <= 0 {
		t.Errorf("MaxTokens = %d, want positive default", captured.MaxTokens)
	}
}

// TestChatWithImages_DropsDetailHint asserts the bridge's Detail value
// doesn't reach Anthropic's wire format (Anthropic doesn't accept it).
func TestChatWithImages_DropsDetailHint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/v1/messages") {
			_, _ = w.Write([]byte(`{"models":[]}`))
			return
		}
		body, _ := io.ReadAll(r.Body)
		// Detail must NOT appear in the wire body.
		if strings.Contains(string(body), `"detail"`) {
			t.Errorf("wire body contains 'detail' key; should be dropped: %s", string(body))
		}
		_, _ = w.Write([]byte(`{"id":"msg","model":"claude-opus-4-5","content":[{"type":"text","text":"ok"}],"usage":{}}`))
	}))
	defer srv.Close()

	p := newTestProvider(t, srv)
	_, err := p.ChatWithImages(context.Background(),
		&core.ChatCompletionRequest{Messages: []core.Message{{Role: "user", Content: "x"}}},
		[]core.ImageInput{{URL: "https://x/y.png", Detail: "high"}},
	)
	if err != nil {
		t.Fatal(err)
	}
}

// TestChatWithImages_RejectsEmpty asserts the at-least-one-image guard.
func TestChatWithImages_RejectsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/v1/messages") {
			t.Error("server reached /v1/messages for empty-images request; the guard failed")
			w.WriteHeader(500)
			return
		}
		_, _ = w.Write([]byte(`{"models":[]}`))
	}))
	defer srv.Close()

	p := newTestProvider(t, srv)
	_, err := p.ChatWithImages(context.Background(),
		&core.ChatCompletionRequest{Messages: []core.Message{{Role: "user", Content: "x"}}},
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "at least one") {
		t.Errorf("err = %v, want 'at least one' error", err)
	}
}

// TestChatWithImages_AppendsSyntheticUserMessage asserts that with no
// user message present, the bridge adds one to host the image blocks.
func TestChatWithImages_AppendsSyntheticUserMessage(t *testing.T) {
	var captured anthropicVisionRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/v1/messages") {
			_, _ = w.Write([]byte(`{"models":[]}`))
			return
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_, _ = w.Write([]byte(`{"id":"msg","model":"claude-opus-4-5","content":[{"type":"text","text":"ok"}],"usage":{}}`))
	}))
	defer srv.Close()

	p := newTestProvider(t, srv)
	_, err := p.ChatWithImages(context.Background(),
		&core.ChatCompletionRequest{
			Messages: []core.Message{{Role: "system", Content: "system only"}},
		},
		[]core.ImageInput{{URL: "https://x/y.png"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(captured.Messages); got != 1 {
		t.Fatalf("messages = %d, want 1 (synthetic user)", got)
	}
	if captured.Messages[0].Role != "user" {
		t.Errorf("synthetic message role = %q, want user", captured.Messages[0].Role)
	}
}
