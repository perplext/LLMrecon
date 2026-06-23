package anthropic

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// dummyProvider returns an AnthropicProvider instance for exercising the
// error-classification method. The base URL points at a closed server because
// handleErrorResponse performs no I/O.
func dummyProvider(t *testing.T) *AnthropicProvider {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	srv.Close()
	return newTestProvider(t, srv)
}

// TestHandleErrorResponse_Classification documents how the Anthropic provider
// maps non-200 statuses: each surfaces as a *core.ProviderError carrying the
// HTTP status code, for both well-formed and malformed error bodies. We call
// handleErrorResponse directly because ChatCompletion's HTTP path is wrapped by
// the circuit breaker (a separate resilience layer) with no raw entry point.
func TestHandleErrorResponse_Classification(t *testing.T) {
	p := dummyProvider(t)
	cases := []struct {
		name   string
		status int
		body   string
	}{
		{"rate_limit", http.StatusTooManyRequests, `{"type":"error","error":{"type":"rate_limit_error","message":"slow down"}}`},
		{"unauthorized", http.StatusUnauthorized, `{"type":"error","error":{"type":"authentication_error","message":"bad key"}}`},
		{"bad_request", http.StatusBadRequest, `{"type":"error","error":{"type":"invalid_request_error","message":"bad"}}`},
		{"server_error", http.StatusInternalServerError, `{"type":"error","error":{"type":"api_error","message":"oops"}}`},
		{"malformed_body", http.StatusServiceUnavailable, `not-json`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := p.handleErrorResponse(c.status, []byte(c.body))
			if err == nil {
				t.Fatal("expected an error")
			}
			var pe *core.ProviderError
			if !errors.As(err, &pe) {
				t.Fatalf("expected *core.ProviderError, got %T: %v", err, err)
			}
			if pe.StatusCode != c.status {
				t.Errorf("StatusCode = %d, want %d", pe.StatusCode, c.status)
			}
		})
	}
}

// TestHandleErrorResponse_RateLimitIsProviderError documents that a 429
// surfaces as a *core.ProviderError carrying the status code. The retry
// decision lives in middleware.RetryMiddleware (see
// src/provider/middleware/retry_test.go), which retries 429/5xx. (#304 removed
// the redundant, unused core.RetryableQuery/Transient/Permanent system.)
func TestHandleErrorResponse_RateLimitIsProviderError(t *testing.T) {
	p := dummyProvider(t)
	err := p.handleErrorResponse(http.StatusTooManyRequests, []byte(`{"error":{"type":"rate_limit_error"}}`))
	if err == nil {
		t.Fatal("expected an error for 429")
	}
	var pe *core.ProviderError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *core.ProviderError, got %T: %v", err, err)
	}
	if pe.StatusCode != http.StatusTooManyRequests {
		t.Errorf("StatusCode = %d, want 429", pe.StatusCode)
	}
}
