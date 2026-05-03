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
// The returned value also implements common.* capability interfaces
// where the wrapped provider supports the underlying core capability:
// future v0.10.0 issue #166 work will extend this.
//
// Returns the inner provider wrapped in a coreAdapter. Caller retains
// ownership of the inner provider — the adapter does not call Close()
// on it; that's the caller's responsibility.
func WrapCore(p core.Provider) common.Provider {
	if p == nil {
		// Defensive: nil core.Provider would NPE inside Query. Operators
		// should never see this path; constructor errors should have
		// already returned, but the contract is clearer with the guard.
		return nilProvider{}
	}
	return &coreAdapter{inner: p}
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
	req := &core.ChatCompletionRequest{
		Messages: convertMessages(msgs),
	}

	// Default to the provider's configured model. Operator can override
	// via opts["model"] if they want to drive multiple models from a
	// single wrapped provider.
	if cfg := a.inner.GetConfig(); cfg != nil {
		req.Model = cfg.DefaultModel
	}
	applyOptions(req, opts)

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
