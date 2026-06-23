package middleware

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// fastRetry builds a RetryMiddleware with negligible backoff so Execute tests
// don't sleep for real seconds.
func fastRetry(maxRetries int) *RetryMiddleware {
	return NewRetryMiddleware(&core.RetryConfig{
		MaxRetries:        maxRetries,
		InitialBackoff:    time.Millisecond,
		MaxBackoff:        time.Millisecond,
		BackoffMultiplier: 1.0,
		RetryableStatusCodes: []int{
			http.StatusTooManyRequests,
			http.StatusInternalServerError,
			http.StatusBadGateway,
			http.StatusServiceUnavailable,
			http.StatusGatewayTimeout,
		},
	})
}

// TestIsRetryableError is the live retry-classification decision (#304): the
// provider retry path keys on *core.ProviderError status codes — 429 and 5xx
// retry, 4xx auth/validation do not. (This replaced the unused
// core.RetryableQuery/Transient/Permanent taxonomy.)
func TestIsRetryableError(t *testing.T) {
	m := fastRetry(3)
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"429 rate limit", &core.ProviderError{StatusCode: http.StatusTooManyRequests}, true},
		{"500 server error", &core.ProviderError{StatusCode: http.StatusInternalServerError}, true},
		{"503 unavailable", &core.ProviderError{StatusCode: http.StatusServiceUnavailable}, true},
		{"400 bad request", &core.ProviderError{StatusCode: http.StatusBadRequest}, false},
		{"401 unauthorized", &core.ProviderError{StatusCode: http.StatusUnauthorized}, false},
		{"404 not found", &core.ProviderError{StatusCode: http.StatusNotFound}, false},
		{"non-provider error", errors.New("plain error"), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := m.isRetryableError(c.err); got != c.want {
				t.Errorf("isRetryableError(%v) = %v, want %v", c.err, got, c.want)
			}
		})
	}
}

func TestExecute_RetriesTransientThenSucceeds(t *testing.T) {
	m := fastRetry(3)
	calls := 0
	result, err := m.Execute(context.Background(), func(context.Context) (interface{}, error) {
		calls++
		if calls < 3 {
			return nil, &core.ProviderError{StatusCode: http.StatusTooManyRequests}
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if result != "ok" {
		t.Fatalf("result = %v, want ok", result)
	}
	if calls != 3 {
		t.Fatalf("calls = %d, want 3 (two 429 retries then success)", calls)
	}
}

func TestExecute_DoesNotRetryPermanent(t *testing.T) {
	m := fastRetry(3)
	calls := 0
	_, err := m.Execute(context.Background(), func(context.Context) (interface{}, error) {
		calls++
		return nil, &core.ProviderError{StatusCode: http.StatusUnauthorized}
	})
	if err == nil {
		t.Fatal("expected error for 401")
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1 (401 is not retryable)", calls)
	}
}

func TestExecute_ExhaustsRetries(t *testing.T) {
	m := fastRetry(2)
	calls := 0
	_, err := m.Execute(context.Background(), func(context.Context) (interface{}, error) {
		calls++
		return nil, &core.ProviderError{StatusCode: http.StatusServiceUnavailable}
	})
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	// MaxRetries=2 → initial attempt + 2 retries = 3 calls.
	if calls != 3 {
		t.Fatalf("calls = %d, want 3 (1 + MaxRetries)", calls)
	}
}

func TestExecute_ContextCancellationStopsRetry(t *testing.T) {
	m := fastRetry(5)
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	_, err := m.Execute(ctx, func(context.Context) (interface{}, error) {
		calls++
		cancel() // cancel after the first failing attempt
		return nil, &core.ProviderError{StatusCode: http.StatusTooManyRequests}
	})
	if err == nil {
		t.Fatal("expected error after context cancellation")
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1 (cancelled before retry)", calls)
	}
}
