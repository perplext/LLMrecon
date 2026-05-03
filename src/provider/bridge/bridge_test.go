package bridge

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/perplext/LLMrecon/src/attacks/common"
	"github.com/perplext/LLMrecon/src/provider/core"
)

// ---------------------------------------------------------------------------
// Mock core.Provider — minimal stub satisfying the full core.Provider
// interface. Only ChatCompletion / GetType / GetConfig / CountTokens have
// behavior tied to bridge tests; the rest return zero values.
// ---------------------------------------------------------------------------

type mockCoreProvider struct {
	cfg *core.ProviderConfig

	// LastRequest captures the most-recent ChatCompletion call so tests
	// can assert what the bridge built.
	LastRequest *core.ChatCompletionRequest

	// Response is what ChatCompletion returns. Tests configure it.
	Response *core.ChatCompletionResponse
	// Err is returned in place of Response when non-nil.
	Err error

	// CountTokensImpl, when set, drives GetTokenCount round-tripping;
	// when nil, returns len/4 fallback (mirrors the bridge's own
	// fallback to keep the test focused on bridge behavior, not the
	// counter's accuracy).
	CountTokensImpl func(text string) int
}

func (m *mockCoreProvider) GetType() core.ProviderType { return core.OpenAIProvider }
func (m *mockCoreProvider) GetConfig() *core.ProviderConfig {
	if m.cfg == nil {
		return &core.ProviderConfig{Type: core.OpenAIProvider, DefaultModel: "test-model"}
	}
	return m.cfg
}
func (m *mockCoreProvider) GetModels(_ context.Context) ([]core.ModelInfo, error) {
	return nil, nil
}
func (m *mockCoreProvider) GetModelInfo(_ context.Context, _ string) (*core.ModelInfo, error) {
	return nil, nil
}
func (m *mockCoreProvider) TextCompletion(_ context.Context, _ *core.TextCompletionRequest) (*core.TextCompletionResponse, error) {
	return nil, errors.New("not implemented in mock")
}
func (m *mockCoreProvider) ChatCompletion(_ context.Context, req *core.ChatCompletionRequest) (*core.ChatCompletionResponse, error) {
	m.LastRequest = req
	if m.Err != nil {
		return nil, m.Err
	}
	if m.Response != nil {
		return m.Response, nil
	}
	// Default: echo back the user's last message.
	last := ""
	for _, msg := range req.Messages {
		if msg.Role == "user" {
			last = msg.Content
		}
	}
	return &core.ChatCompletionResponse{
		Choices: []core.ChatCompletionChoice{{
			Index:        0,
			Message:      core.Message{Role: "assistant", Content: "echo: " + last},
			FinishReason: "stop",
		}},
	}, nil
}
func (m *mockCoreProvider) StreamingChatCompletion(_ context.Context, _ *core.ChatCompletionRequest, _ func(*core.ChatCompletionResponse) error) error {
	return errors.New("not implemented in mock")
}
func (m *mockCoreProvider) CreateEmbedding(_ context.Context, _ *core.EmbeddingRequest) (*core.EmbeddingResponse, error) {
	return nil, errors.New("not implemented in mock")
}
func (m *mockCoreProvider) CountTokens(_ context.Context, text string, _ string) (int, error) {
	if m.CountTokensImpl != nil {
		return m.CountTokensImpl(text), nil
	}
	return len(text) / 4, nil
}
func (m *mockCoreProvider) SupportsModel(_ context.Context, _ string) bool         { return true }
func (m *mockCoreProvider) SupportsCapability(_ context.Context, _ core.ModelCapability) bool {
	return true
}
func (m *mockCoreProvider) Close() error                          { return nil }
func (m *mockCoreProvider) GetRateLimitConfig() *core.RateLimitConfig { return nil }
func (m *mockCoreProvider) UpdateRateLimitConfig(_ *core.RateLimitConfig) error {
	return nil
}
func (m *mockCoreProvider) GetRetryConfig() *core.RetryConfig { return nil }
func (m *mockCoreProvider) UpdateRetryConfig(_ *core.RetryConfig) error {
	return nil
}
func (m *mockCoreProvider) GetUsageMetrics(_ string) (*core.UsageMetrics, error) {
	return nil, nil
}
func (m *mockCoreProvider) GetAllUsageMetrics() (map[string]*core.UsageMetrics, error) {
	return nil, nil
}
func (m *mockCoreProvider) ResetUsageMetrics() error { return nil }

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestWrapCore_ReturnsCommonProvider asserts the wrapped value satisfies
// the common.Provider interface. This is the core type-shape contract:
// if it compiles + the value isn't nil, the bridge's fundamental claim
// holds.
func TestWrapCore_ReturnsCommonProvider(t *testing.T) {
	mock := &mockCoreProvider{}
	var wrapped common.Provider = WrapCore(mock)
	if wrapped == nil {
		t.Fatal("WrapCore returned nil")
	}
}

// TestWrapCore_NilInputReturnsErroringProvider asserts the defensive
// guard for WrapCore(nil) returns a typed-failure provider rather than
// crashing later inside Query.
func TestWrapCore_NilInputReturnsErroringProvider(t *testing.T) {
	wrapped := WrapCore(nil)
	if wrapped == nil {
		t.Fatal("WrapCore(nil) returned untyped nil")
	}
	_, err := wrapped.Query(context.Background(), nil, nil)
	if err == nil || !strings.Contains(err.Error(), "nil") {
		t.Errorf("expected 'nil' in error from nilProvider.Query; got %v", err)
	}
	if wrapped.GetName() != "nil" {
		t.Errorf("nilProvider.GetName = %q, want %q", wrapped.GetName(), "nil")
	}
}

// TestQuery_RoundTrip asserts the happy path: bridge converts messages,
// calls ChatCompletion, extracts the first choice's content.
func TestQuery_RoundTrip(t *testing.T) {
	mock := &mockCoreProvider{
		Response: &core.ChatCompletionResponse{
			Choices: []core.ChatCompletionChoice{{
				Index:   0,
				Message: core.Message{Role: "assistant", Content: "the answer is 42"},
			}},
		},
	}
	wrapped := WrapCore(mock)

	resp, err := wrapped.Query(context.Background(), []common.Message{
		{Role: "system", Content: "you are helpful"},
		{Role: "user", Content: "what is the answer?"},
	}, nil)
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if resp != "the answer is 42" {
		t.Errorf("Query response = %q, want %q", resp, "the answer is 42")
	}

	// Inspect the request the mock saw.
	if mock.LastRequest == nil {
		t.Fatal("mock didn't see a ChatCompletion call")
	}
	if got := len(mock.LastRequest.Messages); got != 2 {
		t.Errorf("request had %d messages; want 2", got)
	}
	if mock.LastRequest.Messages[1].Role != "user" {
		t.Errorf("messages[1].Role = %q, want user", mock.LastRequest.Messages[1].Role)
	}
}

// TestQuery_PreservesTimestamp asserts common.Message.Timestamp is
// preserved on conversion to core.Message (both have the field).
func TestQuery_PreservesTimestamp(t *testing.T) {
	mock := &mockCoreProvider{}
	wrapped := WrapCore(mock)

	now := time.Now()
	_, err := wrapped.Query(context.Background(), []common.Message{
		{Role: "user", Content: "x", Timestamp: now},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := mock.LastRequest.Messages[0].Timestamp; !got.Equal(now) {
		t.Errorf("Timestamp = %v, want %v", got, now)
	}
}

// TestQuery_DefaultsToProviderModel asserts the request's Model field
// is auto-populated from ProviderConfig.DefaultModel when opts don't
// override.
func TestQuery_DefaultsToProviderModel(t *testing.T) {
	mock := &mockCoreProvider{
		cfg: &core.ProviderConfig{Type: core.OpenAIProvider, DefaultModel: "gpt-5"},
	}
	wrapped := WrapCore(mock)

	_, err := wrapped.Query(context.Background(), []common.Message{{Role: "user", Content: "hi"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if mock.LastRequest.Model != "gpt-5" {
		t.Errorf("request.Model = %q, want gpt-5", mock.LastRequest.Model)
	}
}

// TestQuery_OptsModelOverride asserts opts["model"] takes precedence
// over the provider's default.
func TestQuery_OptsModelOverride(t *testing.T) {
	mock := &mockCoreProvider{
		cfg: &core.ProviderConfig{Type: core.OpenAIProvider, DefaultModel: "gpt-5"},
	}
	wrapped := WrapCore(mock)

	_, err := wrapped.Query(
		context.Background(),
		[]common.Message{{Role: "user", Content: "hi"}},
		map[string]interface{}{"model": "gpt-5-mini"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if mock.LastRequest.Model != "gpt-5-mini" {
		t.Errorf("request.Model = %q, want override gpt-5-mini", mock.LastRequest.Model)
	}
}

// TestQuery_OptsMaxTokensAndTemperature asserts the recognized opts
// reach the request.
func TestQuery_OptsMaxTokensAndTemperature(t *testing.T) {
	mock := &mockCoreProvider{}
	wrapped := WrapCore(mock)

	_, err := wrapped.Query(
		context.Background(),
		[]common.Message{{Role: "user", Content: "hi"}},
		map[string]interface{}{"max_tokens": 100, "temperature": 0.7},
	)
	if err != nil {
		t.Fatal(err)
	}
	if mock.LastRequest.MaxTokens != 100 {
		t.Errorf("MaxTokens = %d, want 100", mock.LastRequest.MaxTokens)
	}
	if mock.LastRequest.Temperature != 0.7 {
		t.Errorf("Temperature = %v, want 0.7", mock.LastRequest.Temperature)
	}
}

// TestQuery_UnknownOptsAreIgnored asserts unrecognized opts don't
// cause errors. The bridge's contract is "minimal pass-through" —
// modules wanting full control configure core.Provider directly.
func TestQuery_UnknownOptsAreIgnored(t *testing.T) {
	mock := &mockCoreProvider{}
	wrapped := WrapCore(mock)

	_, err := wrapped.Query(
		context.Background(),
		[]common.Message{{Role: "user", Content: "hi"}},
		map[string]interface{}{"unknown_field": "value", "another": 42},
	)
	if err != nil {
		t.Errorf("unknown opts shouldn't error; got %v", err)
	}
}

// TestQuery_PropagatesProviderError asserts a ChatCompletion error
// reaches the caller verbatim.
func TestQuery_PropagatesProviderError(t *testing.T) {
	mock := &mockCoreProvider{Err: errors.New("rate limit")}
	wrapped := WrapCore(mock)

	_, err := wrapped.Query(context.Background(), []common.Message{{Role: "user", Content: "x"}}, nil)
	if err == nil || !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("expected 'rate limit' error; got %v", err)
	}
}

// TestQuery_NoChoicesReturnsError asserts a malformed response (no
// choices) doesn't silently return empty string.
func TestQuery_NoChoicesReturnsError(t *testing.T) {
	mock := &mockCoreProvider{
		Response: &core.ChatCompletionResponse{Choices: nil},
	}
	wrapped := WrapCore(mock)

	_, err := wrapped.Query(context.Background(), []common.Message{{Role: "user", Content: "x"}}, nil)
	if err == nil || !strings.Contains(err.Error(), "no choices") {
		t.Errorf("expected 'no choices' error; got %v", err)
	}
}

// TestGetName_SurfacesType asserts GetName returns the provider type
// string.
func TestGetName_SurfacesType(t *testing.T) {
	mock := &mockCoreProvider{}
	wrapped := WrapCore(mock)
	if got := wrapped.GetName(); got != string(core.OpenAIProvider) {
		t.Errorf("GetName = %q, want %q", got, string(core.OpenAIProvider))
	}
}

// TestGetModel_FromConfigDefault asserts GetModel reads
// ProviderConfig.DefaultModel.
func TestGetModel_FromConfigDefault(t *testing.T) {
	mock := &mockCoreProvider{
		cfg: &core.ProviderConfig{Type: core.OpenAIProvider, DefaultModel: "gpt-5-mini"},
	}
	wrapped := WrapCore(mock)
	if got := wrapped.GetModel(); got != "gpt-5-mini" {
		t.Errorf("GetModel = %q, want gpt-5-mini", got)
	}
}

// TestGetModel_FallsBackOnEmptyConfig asserts the defensive default
// when no config is present.
func TestGetModel_FallsBackOnEmptyConfig(t *testing.T) {
	mock := &mockCoreProvider{cfg: &core.ProviderConfig{}}
	wrapped := WrapCore(mock)
	if got := wrapped.GetModel(); got != "unknown" {
		t.Errorf("GetModel = %q, want unknown fallback", got)
	}
}

// TestGetTokenCount_DelegatesToProvider asserts the wrapper calls
// CountTokens on the inner provider.
func TestGetTokenCount_DelegatesToProvider(t *testing.T) {
	called := false
	mock := &mockCoreProvider{
		CountTokensImpl: func(text string) int {
			called = true
			return len(text) // distinct from the bridge's len/4 fallback
		},
	}
	wrapped := WrapCore(mock)
	got := wrapped.GetTokenCount("hello world")
	if !called {
		t.Error("CountTokens was not called on the inner provider")
	}
	if got != 11 {
		t.Errorf("GetTokenCount = %d, want 11", got)
	}
}
