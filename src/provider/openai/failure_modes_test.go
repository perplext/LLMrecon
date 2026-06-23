package openai

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// errorServer returns an httptest server that always replies with the given
// HTTP status and body — used to drive the provider's error-classification path.
func errorServer(status int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

func chatReq() *core.ChatCompletionRequest {
	return &core.ChatCompletionRequest{
		Messages: []core.Message{{Role: "user", Content: "hello"}},
	}
}

// TestChatCompletionFromAPI_ErrorClassification documents how the OpenAI
// provider maps non-200 HTTP responses: every failure surfaces as a
// *core.ProviderError carrying the HTTP status code, for both well-formed and
// malformed error bodies. We exercise chatCompletionFromAPI (the raw HTTP path)
// directly rather than ChatCompletion, because the latter's circuit breaker
// retries 5xx/429 and then returns an opaque "circuit breaker is open" error —
// a separate resilience layer from the error-classification under test here.
func TestChatCompletionFromAPI_ErrorClassification(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{"rate_limit", http.StatusTooManyRequests, `{"error":{"message":"rate limit reached","type":"rate_limit_error","code":"rate_limited"}}`},
		{"unauthorized", http.StatusUnauthorized, `{"error":{"message":"invalid api key","type":"invalid_request_error","code":"invalid_api_key"}}`},
		{"bad_request", http.StatusBadRequest, `{"error":{"message":"bad input","type":"invalid_request_error"}}`},
		{"server_error", http.StatusInternalServerError, `{"error":{"message":"internal"}}`},
		{"malformed_body", http.StatusServiceUnavailable, `not-json-at-all`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := errorServer(c.status, c.body)
			defer srv.Close()
			p := newTestProvider(t, srv)

			_, err := p.chatCompletionFromAPI(context.Background(), chatReq())
			if err == nil {
				t.Fatal("expected an error for non-200 response")
			}
			var pe *core.ProviderError
			if !errors.As(err, &pe) {
				t.Fatalf("expected *core.ProviderError in chain, got %T: %v", err, err)
			}
			if pe.StatusCode != c.status {
				t.Errorf("StatusCode = %d, want %d", pe.StatusCode, c.status)
			}
		})
	}
}

// TestChatCompletionFromAPI_RateLimitNotYetRetryable documents a current wiring
// gap (tracked separately): a 429 surfaces as a *core.ProviderError, which
// core.IsTransient does NOT recognize, so the RetryableQuery helper would treat
// it as non-retryable. When the provider→retry classification is implemented,
// this assertion will flip and should be updated alongside closing the gap.
func TestChatCompletionFromAPI_RateLimitNotYetRetryable(t *testing.T) {
	srv := errorServer(http.StatusTooManyRequests, `{"error":{"message":"slow down","type":"rate_limit_error"}}`)
	defer srv.Close()
	p := newTestProvider(t, srv)

	_, err := p.chatCompletionFromAPI(context.Background(), chatReq())
	if err == nil {
		t.Fatal("expected an error for 429")
	}
	if core.IsTransient(err) {
		t.Error("429 is now classified as transient — the retry wiring was implemented; update this test and close the gap issue")
	}
	if core.IsPermanent(err) {
		t.Error("429 is now classified as permanent — revisit: a rate limit should be transient")
	}
}
