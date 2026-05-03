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

// ---------------------------------------------------------------------------
// Capability-promotion tests (v0.10.0 #166)
// ---------------------------------------------------------------------------
//
// These assert WrapCore returns the right wrapper type based on the
// underlying core.Provider's capabilities, and that the capability
// methods translate request/response shapes correctly.

// imageCapableMock embeds mockCoreProvider and adds ChatWithImages so
// it satisfies core.ImageProvider.
type imageCapableMock struct {
	*mockCoreProvider
	LastImages []core.ImageInput
}

func (m *imageCapableMock) ChatWithImages(_ context.Context, req *core.ChatCompletionRequest, images []core.ImageInput) (*core.ChatCompletionResponse, error) {
	m.LastRequest = req
	m.LastImages = images
	if m.Err != nil {
		return nil, m.Err
	}
	if m.Response != nil {
		return m.Response, nil
	}
	return &core.ChatCompletionResponse{
		Choices: []core.ChatCompletionChoice{{
			Index:        0,
			Message:      core.Message{Role: "assistant", Content: "saw an image"},
			FinishReason: "stop",
		}},
	}, nil
}

// reasoningCapableMock embeds mockCoreProvider and adds ChatWithReasoning
// so it satisfies core.ReasoningProvider.
type reasoningCapableMock struct {
	*mockCoreProvider
	Trace *core.ThinkingTrace
}

func (m *reasoningCapableMock) ChatWithReasoning(_ context.Context, req *core.ChatCompletionRequest) (*core.ChatCompletionResponse, *core.ThinkingTrace, error) {
	m.LastRequest = req
	if m.Err != nil {
		return nil, nil, m.Err
	}
	resp := m.Response
	if resp == nil {
		resp = &core.ChatCompletionResponse{
			Choices: []core.ChatCompletionChoice{{
				Index:        0,
				Message:      core.Message{Role: "assistant", Content: "thought about it"},
				FinishReason: "stop",
			}},
		}
	}
	return resp, m.Trace, nil
}

// signedReasoningMock additionally satisfies the signedReasoningProvider
// marker interface so the bridge surfaces Signed=true.
type signedReasoningMock struct {
	*reasoningCapableMock
}

func (signedReasoningMock) ReasoningTraceIsSigned() bool { return true }

// imageReasoningCapableMock satisfies BOTH core.ImageProvider and
// core.ReasoningProvider.
type imageReasoningCapableMock struct {
	*mockCoreProvider
	LastImages []core.ImageInput
	Trace      *core.ThinkingTrace
}

func (m *imageReasoningCapableMock) ChatWithImages(_ context.Context, req *core.ChatCompletionRequest, images []core.ImageInput) (*core.ChatCompletionResponse, error) {
	m.LastRequest = req
	m.LastImages = images
	return &core.ChatCompletionResponse{
		Choices: []core.ChatCompletionChoice{{Message: core.Message{Role: "assistant", Content: "img"}}},
	}, nil
}

func (m *imageReasoningCapableMock) ChatWithReasoning(_ context.Context, req *core.ChatCompletionRequest) (*core.ChatCompletionResponse, *core.ThinkingTrace, error) {
	m.LastRequest = req
	return &core.ChatCompletionResponse{
		Choices: []core.ChatCompletionChoice{{Message: core.Message{Role: "assistant", Content: "rsn"}}},
	}, m.Trace, nil
}

// TestWrapCore_PlainProviderDoesNotPromote asserts a plain core.Provider
// (no capabilities) wraps to a value that does NOT satisfy
// common.ImageProvider or common.ReasoningProvider — so attack-module
// type assertions correctly skip.
func TestWrapCore_PlainProviderDoesNotPromote(t *testing.T) {
	wrapped := WrapCore(&mockCoreProvider{})
	if _, ok := wrapped.(common.ImageProvider); ok {
		t.Error("plain provider promoted to common.ImageProvider; want skip")
	}
	if _, ok := wrapped.(common.ReasoningProvider); ok {
		t.Error("plain provider promoted to common.ReasoningProvider; want skip")
	}
}

// TestWrapCore_ImagePromotion asserts an underlying core.ImageProvider
// surfaces as common.ImageProvider AND that QueryWithImages translates
// the ImagePayload into a core.ImageInput.
func TestWrapCore_ImagePromotion(t *testing.T) {
	imgMock := &imageCapableMock{mockCoreProvider: &mockCoreProvider{}}
	wrapped := WrapCore(imgMock)

	ip, ok := wrapped.(common.ImageProvider)
	if !ok {
		t.Fatal("wrapper does not satisfy common.ImageProvider")
	}
	// Reasoning should NOT be promoted — only image is supported.
	if _, ok := wrapped.(common.ReasoningProvider); ok {
		t.Error("wrapper unexpectedly satisfies common.ReasoningProvider")
	}

	payload, err := common.NewImagePayloadBytes(
		[]byte{0x89, 0x50, 0x4E, 0x47}, // PNG magic, just any bytes for the test
		common.ImageMimePNG,
		common.ImageDetailHigh,
	)
	if err != nil {
		t.Fatalf("NewImagePayloadBytes: %v", err)
	}

	resp, err := ip.QueryWithImages(context.Background(), "what is this?", []common.ImagePayload{payload}, nil)
	if err != nil {
		t.Fatalf("QueryWithImages: %v", err)
	}
	if resp != "saw an image" {
		t.Errorf("response = %q, want %q", resp, "saw an image")
	}
	if len(imgMock.LastImages) != 1 {
		t.Fatalf("got %d images at core layer, want 1", len(imgMock.LastImages))
	}
	got := imgMock.LastImages[0]
	if got.MimeType != "image/png" {
		t.Errorf("MimeType = %q, want image/png", got.MimeType)
	}
	if got.Detail != "high" {
		t.Errorf("Detail = %q, want high", got.Detail)
	}
	if len(got.Bytes) != 4 {
		t.Errorf("Bytes len = %d, want 4", len(got.Bytes))
	}
}

// TestWrapCore_ImagePromotion_RejectsEmptyImages asserts the bridge
// fails fast on a zero-image call rather than dispatching to the
// provider with an empty slice.
func TestWrapCore_ImagePromotion_RejectsEmptyImages(t *testing.T) {
	wrapped := WrapCore(&imageCapableMock{mockCoreProvider: &mockCoreProvider{}})
	ip := wrapped.(common.ImageProvider)
	_, err := ip.QueryWithImages(context.Background(), "x", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "at least one") {
		t.Errorf("expected 'at least one image required' error; got %v", err)
	}
}

// TestWrapCore_ReasoningPromotion asserts an underlying
// core.ReasoningProvider surfaces as common.ReasoningProvider AND that
// QueryWithReasoning translates the ThinkingTrace.Steps into the
// minimal ReasoningTrace.Steps shape modules consume.
func TestWrapCore_ReasoningPromotion(t *testing.T) {
	rsnMock := &reasoningCapableMock{
		mockCoreProvider: &mockCoreProvider{},
		Trace: &core.ThinkingTrace{
			Steps: []core.ThinkingStep{
				{Content: "first I considered X", Type: "summary"},
				{Content: "then concluded Y", Type: "summary"},
				{Content: "", Type: "summary"}, // empty step should be elided
			},
			TotalThinkingTokens: 1234,
		},
	}
	wrapped := WrapCore(rsnMock)

	rp, ok := wrapped.(common.ReasoningProvider)
	if !ok {
		t.Fatal("wrapper does not satisfy common.ReasoningProvider")
	}
	if _, ok := wrapped.(common.ImageProvider); ok {
		t.Error("wrapper unexpectedly satisfies common.ImageProvider")
	}

	resp, trace, err := rp.QueryWithReasoning(context.Background(), []common.Message{
		{Role: "user", Content: "solve X"},
	}, nil)
	if err != nil {
		t.Fatalf("QueryWithReasoning: %v", err)
	}
	if resp != "thought about it" {
		t.Errorf("response = %q, want %q", resp, "thought about it")
	}
	if trace.Signed {
		t.Error("trace.Signed = true for non-signed provider; want false")
	}
	if got := len(trace.Steps); got != 2 {
		t.Errorf("Steps count = %d, want 2 (empty step should be elided)", got)
	}
	if trace.Steps[0] != "first I considered X" {
		t.Errorf("Steps[0] = %q, want %q", trace.Steps[0], "first I considered X")
	}
}

// TestWrapCore_ReasoningPromotion_SignedFlag asserts a core provider
// marked via signedReasoningProvider surfaces Signed=true on the
// returned trace. This is the v0.10.0 #166 hook for Anthropic's
// extended-thinking signed-trace handling.
func TestWrapCore_ReasoningPromotion_SignedFlag(t *testing.T) {
	signed := signedReasoningMock{
		reasoningCapableMock: &reasoningCapableMock{
			mockCoreProvider: &mockCoreProvider{},
			Trace:            &core.ThinkingTrace{Steps: []core.ThinkingStep{{Content: "step"}}},
		},
	}
	wrapped := WrapCore(signed)
	rp, ok := wrapped.(common.ReasoningProvider)
	if !ok {
		t.Fatal("wrapper does not satisfy common.ReasoningProvider")
	}
	_, trace, err := rp.QueryWithReasoning(context.Background(), []common.Message{{Role: "user", Content: "x"}}, nil)
	if err != nil {
		t.Fatalf("QueryWithReasoning: %v", err)
	}
	if !trace.Signed {
		t.Error("trace.Signed = false for signed provider; want true")
	}
}

// TestWrapCore_ImageReasoningPromotion asserts a provider supporting
// BOTH capabilities returns a wrapper satisfying both interfaces, and
// each capability method routes to the correct underlying provider.
func TestWrapCore_ImageReasoningPromotion(t *testing.T) {
	dual := &imageReasoningCapableMock{
		mockCoreProvider: &mockCoreProvider{},
		Trace:            &core.ThinkingTrace{Steps: []core.ThinkingStep{{Content: "thinking"}}},
	}
	wrapped := WrapCore(dual)

	ip, okI := wrapped.(common.ImageProvider)
	rp, okR := wrapped.(common.ReasoningProvider)
	if !okI {
		t.Fatal("dual-capable wrapper does not satisfy common.ImageProvider")
	}
	if !okR {
		t.Fatal("dual-capable wrapper does not satisfy common.ReasoningProvider")
	}

	payload, _ := common.NewImagePayloadBytes(
		[]byte{0x00, 0x01},
		common.ImageMimeJPEG,
		common.ImageDetailAuto,
	)

	imgResp, err := ip.QueryWithImages(context.Background(), "look", []common.ImagePayload{payload}, nil)
	if err != nil {
		t.Fatalf("QueryWithImages: %v", err)
	}
	if imgResp != "img" {
		t.Errorf("image response = %q, want %q", imgResp, "img")
	}

	rsnResp, _, err := rp.QueryWithReasoning(context.Background(), []common.Message{{Role: "user", Content: "think"}}, nil)
	if err != nil {
		t.Fatalf("QueryWithReasoning: %v", err)
	}
	if rsnResp != "rsn" {
		t.Errorf("reasoning response = %q, want %q", rsnResp, "rsn")
	}
}

// TestWrapCore_ImagePromotion_URLPayload asserts URL-referenced
// ImagePayloads round-trip with empty Bytes and the URL field set.
func TestWrapCore_ImagePromotion_URLPayload(t *testing.T) {
	imgMock := &imageCapableMock{mockCoreProvider: &mockCoreProvider{}}
	wrapped := WrapCore(imgMock)
	ip := wrapped.(common.ImageProvider)

	payload, err := common.NewImagePayloadURL(
		"https://example.com/x.png",
		common.ImageMimePNG,
		common.ImageDetailLow,
	)
	if err != nil {
		t.Fatalf("NewImagePayloadURL: %v", err)
	}
	_, err = ip.QueryWithImages(context.Background(), "p", []common.ImagePayload{payload}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(imgMock.LastImages) != 1 {
		t.Fatalf("got %d images, want 1", len(imgMock.LastImages))
	}
	got := imgMock.LastImages[0]
	if got.URL != "https://example.com/x.png" {
		t.Errorf("URL = %q, want %q", got.URL, "https://example.com/x.png")
	}
	if len(got.Bytes) != 0 {
		t.Errorf("Bytes len = %d, want 0 for URL payload", len(got.Bytes))
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
