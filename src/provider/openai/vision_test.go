package openai

import (
	"bytes"
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

// newTestProvider returns an OpenAIProvider pointed at the given test
// server, bypassing the heavyweight middleware setup used by
// NewOpenAIProvider. The provider is concrete and exposes the methods
// under test.
func newTestProvider(t *testing.T, srv *httptest.Server) *OpenAIProvider {
	t.Helper()
	p, err := NewOpenAIProvider(&core.ProviderConfig{
		Type:         core.OpenAIProvider,
		APIKey:       "test-key",
		BaseURL:      srv.URL,
		DefaultModel: "gpt-4o",
		Timeout:      5 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewOpenAIProvider: %v", err)
	}
	op, ok := p.(*OpenAIProvider)
	if !ok {
		t.Fatalf("provider is not *OpenAIProvider: %T", p)
	}
	return op
}

// TestEncodeImageInputAsURL covers the URL/bytes mutual exclusion and
// the data: URL formatting for inline-bytes inputs.
func TestEncodeImageInputAsURL(t *testing.T) {
	cases := []struct {
		name    string
		in      core.ImageInput
		want    string
		wantErr string
	}{
		{
			name: "url passes through",
			in:   core.ImageInput{URL: "https://example.com/cat.jpg", MimeType: "image/jpeg"},
			want: "https://example.com/cat.jpg",
		},
		{
			name: "inline bytes encode as data URL",
			in:   core.ImageInput{Bytes: []byte{0xFF, 0xD8, 0xFF}, MimeType: "image/jpeg"},
			want: "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString([]byte{0xFF, 0xD8, 0xFF}),
		},
		{
			name:    "both url and bytes is invalid",
			in:      core.ImageInput{URL: "x", Bytes: []byte{1}, MimeType: "image/png"},
			wantErr: "mutually exclusive",
		},
		{
			name:    "neither url nor bytes is invalid",
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
			got, err := encodeImageInputAsURL(tc.in)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("err = %v, want substring %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("err = %v, want nil", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// TestChatWithImages_RequestShape asserts the wire body sent to OpenAI
// has the expected multimodal shape: messages[].content is an array
// with text + image_url parts, and the image part carries the data URL
// + Detail hint.
func TestChatWithImages_RequestShape(t *testing.T) {
	var captured visionRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The provider's NewOpenAIProvider kicks off a GetModels call
		// in a background goroutine; ignore it so the test only captures
		// the chat completions body.
		if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":[]}`))
			return
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &captured); err != nil {
			t.Errorf("failed to unmarshal request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"r","model":"gpt-4o","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}]}`))
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
			{Bytes: []byte{0x89, 0x50}, MimeType: "image/png", Detail: "high"},
		},
	)
	if err != nil {
		t.Fatalf("ChatWithImages: %v", err)
	}
	if resp.Choices[0].Message.Content != "ok" {
		t.Errorf("response content = %q, want %q", resp.Choices[0].Message.Content, "ok")
	}

	if captured.Model != "gpt-4o" {
		t.Errorf("Model = %q, want gpt-4o", captured.Model)
	}
	if got := len(captured.Messages); got != 2 {
		t.Fatalf("messages count = %d, want 2", got)
	}
	user := captured.Messages[1]
	if user.Role != "user" {
		t.Errorf("messages[1].Role = %q, want user", user.Role)
	}
	if got := len(user.Content); got != 2 {
		t.Fatalf("user.Content parts = %d, want 2 (text + image)", got)
	}
	if user.Content[0].Type != "text" || user.Content[0].Text != "what is this?" {
		t.Errorf("text part wrong: %+v", user.Content[0])
	}
	if user.Content[1].Type != "image_url" {
		t.Errorf("image part type = %q, want image_url", user.Content[1].Type)
	}
	if user.Content[1].ImageURL == nil {
		t.Fatal("image_url missing")
	}
	if !strings.HasPrefix(user.Content[1].ImageURL.URL, "data:image/png;base64,") {
		t.Errorf("image data URL = %q, want data:image/png;base64,...", user.Content[1].ImageURL.URL)
	}
	if user.Content[1].ImageURL.Detail != "high" {
		t.Errorf("detail = %q, want high", user.Content[1].ImageURL.Detail)
	}
}

// TestChatWithImages_AppendsSyntheticUserMessage asserts that when the
// request has no user message (e.g., system-only), the bridge adds a
// synthetic user message to host the image parts.
func TestChatWithImages_AppendsSyntheticUserMessage(t *testing.T) {
	var captured visionRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			_, _ = w.Write([]byte(`{"data":[]}`))
			return
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &captured)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer srv.Close()

	p := newTestProvider(t, srv)
	_, err := p.ChatWithImages(context.Background(),
		&core.ChatCompletionRequest{
			Messages: []core.Message{{Role: "system", Content: "system only"}},
		},
		[]core.ImageInput{{URL: "https://x/y.png", MimeType: "image/png"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(captured.Messages); got != 2 {
		t.Fatalf("messages = %d, want 2 (system + synthetic user)", got)
	}
	if captured.Messages[1].Role != "user" {
		t.Errorf("synthetic message.Role = %q, want user", captured.Messages[1].Role)
	}
}

// TestChatWithImages_RejectsEmpty asserts at-least-one-image guard.
func TestChatWithImages_RejectsEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Error("server should not be called for empty-images request")
		w.WriteHeader(500)
	}))
	defer srv.Close()
	p := newTestProvider(t, srv)
	_, err := p.ChatWithImages(context.Background(),
		&core.ChatCompletionRequest{Messages: []core.Message{{Role: "user", Content: "hi"}}},
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "at least one") {
		t.Errorf("err = %v, want 'at least one' error", err)
	}
}

// TestChatWithImages_PropagatesAPIError asserts non-200 responses
// surface as a typed *core.ProviderError.
func TestChatWithImages_PropagatesAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			_, _ = w.Write([]byte(`{"data":[]}`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"message":"bad image","type":"invalid_request_error"}}`))
	}))
	defer srv.Close()
	p := newTestProvider(t, srv)

	_, err := p.ChatWithImages(context.Background(),
		&core.ChatCompletionRequest{Messages: []core.Message{{Role: "user", Content: "x"}}},
		[]core.ImageInput{{URL: "https://x/y.png", MimeType: "image/png"}},
	)
	if err == nil {
		t.Fatal("expected error")
	}
	pe, ok := err.(*core.ProviderError)
	if !ok {
		// retry middleware may wrap, accept substring match on Error()
		if !strings.Contains(err.Error(), "bad image") {
			t.Fatalf("err = %v, want ProviderError or substring 'bad image'", err)
		}
		return
	}
	if pe.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want 400", pe.StatusCode)
	}
}

// TestChatWithImages_AuthorizationHeader asserts the API key reaches
// the request as a Bearer header — minimal smoke test for the
// authentication wiring.
func TestChatWithImages_AuthorizationHeader(t *testing.T) {
	gotAuth := ""
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			_, _ = w.Write([]byte(`{"data":[]}`))
			return
		}
		gotAuth = r.Header.Get("Authorization")
		_, _ = io.Copy(io.Discard, r.Body)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer srv.Close()
	p := newTestProvider(t, srv)
	_, _ = p.ChatWithImages(context.Background(),
		&core.ChatCompletionRequest{Messages: []core.Message{{Role: "user", Content: "x"}}},
		[]core.ImageInput{{URL: "https://x/y.png", MimeType: "image/png"}},
	)
	if !strings.HasPrefix(gotAuth, "Bearer ") {
		t.Errorf("Authorization header = %q, want Bearer prefix", gotAuth)
	}
}

// silence unused-import lint when building without these.
var _ = bytes.NewBuffer
