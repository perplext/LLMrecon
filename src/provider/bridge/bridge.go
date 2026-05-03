// Package bridge wraps a core.Provider to expose the common.Provider
// surface that attack modules consume.
//
// The two interfaces have different shapes:
//
//   core.Provider    — full LLM-API surface: ChatCompletion takes
//                      *ChatCompletionRequest with token / temperature /
//                      tool / metadata fields, returns *ChatCompletionResponse
//                      with id / created / model / choices / usage.
//   common.Provider  — minimal surface attack modules need: Query takes
//                      []Message + options map, returns plain string.
//
// The bridge translates between them: build a minimal request, call
// through, extract the first choice's content. Modules that need more
// (token counts for fitness, etc.) call the wrapper's GetTokenCount.
//
// This is the v0.10.0 issue #167 deliverable. Without this, the
// attack-module ecosystem (consuming common.Provider) was unreachable
// from the OpenAI/Anthropic adapters (returning core.Provider). After
// this lands, `attack run --provider=openai` flows from operator
// invocation through the factory → core.Provider → bridge.WrapCore →
// common.Provider → module.Execute.
package bridge

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/attacks/common"
	"github.com/perplext/LLMrecon/src/provider/core"
)

// WrapCore returns a common.Provider backed by the given core.Provider.
// The returned value also implements common.ImageProvider and/or
// common.ReasoningProvider when the wrapped provider implements the
// corresponding core capability interface (v0.10.0 #166).
//
// The 2x2 capability matrix is realized by four concrete wrapper types
// (coreAdapter, imageAdapter, reasoningAdapter, imageReasoningAdapter)
// rather than one struct with always-present methods. The reason: Go's
// type assertion is structural at compile time — a single struct with
// QueryWithImages would always satisfy common.ImageProvider, defeating
// the type-assertion gate that attack modules use to detect capability
// presence. Modules need OK to mean "this provider really does images,"
// not "the wrapper has the method."
//
// Returns the inner provider wrapped in a coreAdapter (or capability
// extension thereof). Caller retains ownership of the inner provider —
// the adapter does not call Close() on it; that's the caller's
// responsibility.
func WrapCore(p core.Provider) common.Provider {
	if p == nil {
		return nilProvider{}
	}
	base := &coreAdapter{inner: p}
	img, hasImg := p.(core.ImageProvider)
	rsn, hasRsn := p.(core.ReasoningProvider)
	switch {
	case hasImg && hasRsn:
		return &imageReasoningAdapter{coreAdapter: base, img: img, rsn: rsn}
	case hasImg:
		return &imageAdapter{coreAdapter: base, img: img}
	case hasRsn:
		return &reasoningAdapter{coreAdapter: base, rsn: rsn}
	default:
		return base
	}
}

// coreAdapter is the WrapCore-returned common.Provider.
type coreAdapter struct {
	inner core.Provider
}

// Query converts the operator-supplied []common.Message into a minimal
// core.ChatCompletionRequest, calls through, and extracts the first
// choice's content as the response string.
//
// Options pass-through is intentionally minimal in v1: well-known keys
// are mapped (model, max_tokens, temperature); unknown keys are ignored.
// Modules that need richer control should configure the underlying
// provider via core.ProviderConfig before WrapCore.
func (a *coreAdapter) Query(ctx context.Context, msgs []common.Message, opts map[string]interface{}) (string, error) {
	req := a.buildChatRequest(msgs, opts)
	resp, err := a.inner.ChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	if resp == nil || len(resp.Choices) == 0 {
		return "", fmt.Errorf("provider %s returned no choices", a.GetName())
	}
	return resp.Choices[0].Message.Content, nil
}

// GetName surfaces the wrapped provider's type as the common-side name.
// Stable across all OpenAI/Anthropic/etc. adapters for an operator.
func (a *coreAdapter) GetName() string {
	return string(a.inner.GetType())
}

// GetModel returns the wrapped provider's configured default model, or
// "unknown" when no config is available (defensive — providers in
// well-formed setups always have a config).
func (a *coreAdapter) GetModel() string {
	cfg := a.inner.GetConfig()
	if cfg == nil || cfg.DefaultModel == "" {
		return "unknown"
	}
	return cfg.DefaultModel
}

// GetTokenCount delegates to the wrapped provider's CountTokens. Errors
// fall back to the rough-estimate len/4 used elsewhere in the codebase
// — token counting is a best-effort signal for fitness scoring, not a
// correctness boundary, so a tokenizer mismatch shouldn't kill the run.
func (a *coreAdapter) GetTokenCount(text string) int {
	cfg := a.inner.GetConfig()
	model := ""
	if cfg != nil {
		model = cfg.DefaultModel
	}
	n, err := a.inner.CountTokens(context.Background(), text, model)
	if err != nil {
		return len(text) / 4
	}
	return n
}

// ---------------------------------------------------------------------------
// Capability-extension wrappers (v0.10.0 #166)
// ---------------------------------------------------------------------------
//
// Each adapter embeds *coreAdapter so all common.Provider methods (Query,
// GetName, GetModel, GetTokenCount) inherit from the base. The capability
// methods (QueryWithImages, QueryWithReasoning) live on the extension
// type. The compile-time interface checks below catch missing methods.

// imageAdapter exposes common.ImageProvider on top of coreAdapter.
type imageAdapter struct {
	*coreAdapter
	img core.ImageProvider
}

// QueryWithImages translates the common-side prompt + ImagePayload list
// into a core.ChatCompletionRequest + []core.ImageInput, calls through
// the wrapped core.ImageProvider, and extracts the first choice's text.
//
// Constraint translation:
//   - common.ImagePayload's mutual-exclusion invariant (bytes XOR url)
//     was already validated at construction; the conversion here trusts
//     it but the underlying core.ImageInput documents the same.
//   - Empty images slice short-circuits to a typed error rather than
//     falling through to the core.ImageProvider — the bridge's contract
//     is that QueryWithImages with zero images is a programming error.
func (a *imageAdapter) QueryWithImages(ctx context.Context, prompt string, images []common.ImagePayload, opts map[string]interface{}) (string, error) {
	if len(images) == 0 {
		return "", fmt.Errorf("bridge: QueryWithImages: at least one image required")
	}
	req := a.buildChatRequest([]common.Message{{Role: "user", Content: prompt}}, opts)
	coreImages := make([]core.ImageInput, 0, len(images))
	for _, p := range images {
		coreImages = append(coreImages, core.ImageInput{
			Bytes:    p.Bytes(),
			URL:      p.URL(),
			MimeType: string(p.MimeType()),
			Detail:   string(p.Detail()),
		})
	}
	resp, err := a.img.ChatWithImages(ctx, req, coreImages)
	if err != nil {
		return "", err
	}
	if resp == nil || len(resp.Choices) == 0 {
		return "", fmt.Errorf("provider %s returned no choices", a.GetName())
	}
	return resp.Choices[0].Message.Content, nil
}

// reasoningAdapter exposes common.ReasoningProvider on top of coreAdapter.
type reasoningAdapter struct {
	*coreAdapter
	rsn core.ReasoningProvider
}

// QueryWithReasoning calls through to the wrapped core.ReasoningProvider
// and translates the heavier core.ThinkingTrace into the minimal
// common.ReasoningTrace shape attack modules consume.
//
// trace.Signed: the bridge defaults to false. Anthropic's adapter sets
// Signed=true on the core.ThinkingTrace via a sentinel marker (the
// Anthropic-specific implementation surfaces this via a field on
// core.ThinkingTrace once #166 lands the Anthropic side); for OpenAI
// the trace is plain summary text so Signed stays false. See the
// signedReasoningProvider sentinel below.
func (a *reasoningAdapter) QueryWithReasoning(ctx context.Context, msgs []common.Message, opts map[string]interface{}) (string, common.ReasoningTrace, error) {
	req := a.buildChatRequest(msgs, opts)
	resp, trace, err := a.rsn.ChatWithReasoning(ctx, req)
	if err != nil {
		return "", common.ReasoningTrace{}, err
	}
	if resp == nil || len(resp.Choices) == 0 {
		return "", common.ReasoningTrace{}, fmt.Errorf("provider %s returned no choices", a.GetName())
	}
	steps := make([]string, 0)
	if trace != nil {
		for _, s := range trace.Steps {
			if s.Content != "" {
				steps = append(steps, s.Content)
			}
		}
	}
	signed := false
	if sig, ok := a.rsn.(signedReasoningProvider); ok {
		signed = sig.ReasoningTraceIsSigned()
	}
	return resp.Choices[0].Message.Content, common.ReasoningTrace{Steps: steps, Signed: signed}, nil
}

// signedReasoningProvider is an opt-in marker any core.ReasoningProvider
// can implement to declare its reasoning trace is cryptographically
// signed (Anthropic). Modules consuming common.ReasoningTrace.Signed
// gate mutation paths on the field; without this marker the bridge
// defaults to Signed=false.
type signedReasoningProvider interface {
	ReasoningTraceIsSigned() bool
}

// imageReasoningAdapter exposes both common.ImageProvider and
// common.ReasoningProvider. The capability methods come from the two
// embedded extension types; the embedded *coreAdapter remains the
// common.Provider implementer.
type imageReasoningAdapter struct {
	*coreAdapter
	img core.ImageProvider
	rsn core.ReasoningProvider
}

// QueryWithImages — see imageAdapter.QueryWithImages.
func (a *imageReasoningAdapter) QueryWithImages(ctx context.Context, prompt string, images []common.ImagePayload, opts map[string]interface{}) (string, error) {
	tmp := &imageAdapter{coreAdapter: a.coreAdapter, img: a.img}
	return tmp.QueryWithImages(ctx, prompt, images, opts)
}

// QueryWithReasoning — see reasoningAdapter.QueryWithReasoning.
func (a *imageReasoningAdapter) QueryWithReasoning(ctx context.Context, msgs []common.Message, opts map[string]interface{}) (string, common.ReasoningTrace, error) {
	tmp := &reasoningAdapter{coreAdapter: a.coreAdapter, rsn: a.rsn}
	return tmp.QueryWithReasoning(ctx, msgs, opts)
}

// Compile-time interface checks. These fail at build time if the
// wrappers stop satisfying the common.* contracts.
var (
	_ common.Provider          = (*coreAdapter)(nil)
	_ common.ImageProvider     = (*imageAdapter)(nil)
	_ common.ReasoningProvider = (*reasoningAdapter)(nil)
	_ common.ImageProvider     = (*imageReasoningAdapter)(nil)
	_ common.ReasoningProvider = (*imageReasoningAdapter)(nil)
)

// buildChatRequest centralizes the common.Message → core.ChatCompletionRequest
// conversion so the capability adapters share a single source of truth
// with the plain-text Query path.
func (a *coreAdapter) buildChatRequest(msgs []common.Message, opts map[string]interface{}) *core.ChatCompletionRequest {
	req := &core.ChatCompletionRequest{
		Messages: convertMessages(msgs),
	}
	if cfg := a.inner.GetConfig(); cfg != nil {
		req.Model = cfg.DefaultModel
	}
	applyOptions(req, opts)
	return req
}

// ---------------------------------------------------------------------------
// nilProvider — sentinel for the WrapCore(nil) defensive guard.
// ---------------------------------------------------------------------------

// nilProvider returns a "wrap returned nil" error from every Query so
// downstream code paths see a typed failure rather than a panic.
type nilProvider struct{}

func (nilProvider) Query(_ context.Context, _ []common.Message, _ map[string]interface{}) (string, error) {
	return "", fmt.Errorf("bridge.WrapCore: wrapped core.Provider is nil")
}
func (nilProvider) GetName() string             { return "nil" }
func (nilProvider) GetModel() string            { return "nil" }
func (nilProvider) GetTokenCount(_ string) int  { return 0 }

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// convertMessages translates common.Message into core.Message. Both
// types carry Role + Content + Timestamp; common.Message has Metadata
// (silently dropped — core has no equivalent), core.Message has
// Name/FunctionCall/ToolCalls (defaulted to zero values — attack
// modules don't populate them).
func convertMessages(in []common.Message) []core.Message {
	out := make([]core.Message, len(in))
	for i, m := range in {
		out[i] = core.Message{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp,
		}
	}
	return out
}

// applyOptions maps the small subset of well-known opts to request
// fields. Keys not on the recognized list are silently dropped; this
// is the bridge's narrow contract — modules wanting richer control
// should configure the core.Provider directly before WrapCore.
//
// Recognized keys (all optional):
//   - "model"        — string, overrides ProviderConfig.DefaultModel
//   - "max_tokens"   — int
//   - "temperature"  — float64
func applyOptions(req *core.ChatCompletionRequest, opts map[string]interface{}) {
	if opts == nil {
		return
	}
	if v, ok := opts["model"]; ok {
		if s, ok := v.(string); ok && s != "" {
			req.Model = s
		}
	}
	if v, ok := opts["max_tokens"]; ok {
		if n, ok := v.(int); ok {
			req.MaxTokens = n
		}
	}
	if v, ok := opts["temperature"]; ok {
		if f, ok := v.(float64); ok {
			req.Temperature = f
		}
	}
}
