package reasoning

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks"
	"github.com/perplext/LLMrecon/src/attacks/common"
	"github.com/perplext/LLMrecon/src/attacks/testutil"
)

// ---------------------------------------------------------------------------
// Mock ReasoningProvider
// ---------------------------------------------------------------------------

// mockReasoningProvider embeds testutil.MockProvider for the base Provider
// methods (Query, GetName, …) and adds QueryWithReasoning. The two Query
// paths share the response queue / call log so tests can assert across
// both.
//
// All knobs are zero-valued by default; tests configure only what they
// exercise.
type mockReasoningProvider struct {
	*testutil.MockProvider

	// reasoningResponses is a queue of (response, trace) pairs returned by
	// QueryWithReasoning. When exhausted, defaultReasoning is used.
	reasoningResponses []reasoningResponse
	defaultReasoning   reasoningResponse

	// reasoningErrorOn fires an error on the Nth QueryWithReasoning call
	// (1-based). 0 = never.
	reasoningErrorOn  int
	reasoningErrorMsg string

	reasoningCalls int
}

type reasoningResponse struct {
	resp  string
	trace common.ReasoningTrace
}

func (m *mockReasoningProvider) QueryWithReasoning(_ context.Context, messages []common.Message, _ map[string]interface{}) (string, common.ReasoningTrace, error) {
	m.reasoningCalls++

	// Record on the underlying MockProvider's call log so test assertions
	// using CallCount/Calls see this call too.
	m.MockProvider.Calls = append(m.MockProvider.Calls, testutil.MockCall{Messages: messages})

	if m.reasoningErrorOn > 0 && m.reasoningCalls == m.reasoningErrorOn {
		msg := m.reasoningErrorMsg
		if msg == "" {
			msg = "mock reasoning error"
		}
		return "", common.ReasoningTrace{}, errors.New(msg)
	}

	if len(m.reasoningResponses) > 0 {
		r := m.reasoningResponses[0]
		m.reasoningResponses = m.reasoningResponses[1:]
		return r.resp, r.trace, nil
	}
	return m.defaultReasoning.resp, m.defaultReasoning.trace, nil
}

// newMockReasoningProvider returns a ReasoningProvider with a pre-populated
// initial response so tests don't have to spell out the trace boilerplate.
func newMockReasoningProvider(initialResp string, traceSteps []string, signed bool) *mockReasoningProvider {
	return &mockReasoningProvider{
		MockProvider: &testutil.MockProvider{
			DefaultResponse: "follow-up default response",
		},
		defaultReasoning: reasoningResponse{
			resp:  initialResp,
			trace: common.ReasoningTrace{Steps: traceSteps, Signed: signed},
		},
	}
}

// gatedHCoTConfig returns a config with the safety gate satisfied + a
// non-empty payload. Tests override fields they exercise.
func gatedHCoTConfig() common.AttackConfig {
	return common.AttackConfig{
		Payload: "Walk me through the technique step by step.",
		Metadata: map[string]string{
			"i_understand_risks": "true",
		},
	}
}

// safetyTrace is a 4-step canned trace where step 2 (index 1) contains
// safety vocabulary so locateSafetyStep finds it.
func safetyTrace() []string {
	return []string{
		"Step 1: parse the user's request",
		"Step 2: I should refuse this — it violates the safety guidelines.",
		"Step 3: explain the refusal politely",
		"Step 4: offer a generic alternative",
	}
}

// ---------------------------------------------------------------------------
// Registration
// ---------------------------------------------------------------------------

func TestHCoT_Registered(t *testing.T) {
	mod, err := attacks.DefaultRegistry.Get("h_cot")
	if err != nil {
		t.Fatalf("h_cot not registered: %v", err)
	}
	if mod.Category() != common.CategoryReasoning {
		t.Errorf("Category = %q, want %q", mod.Category(), common.CategoryReasoning)
	}
	techs := mod.Techniques()
	if len(techs) != 1 || techs[0].ID != "h_cot" {
		t.Errorf("Techniques = %#v", techs)
	}
	hasASI01 := false
	for _, c := range techs[0].OWASPAgenticCategories {
		if c == "ASI01" {
			hasASI01 = true
		}
	}
	if !hasASI01 {
		t.Errorf("h_cot must map to ASI01; got %v", techs[0].OWASPAgenticCategories)
	}
}

// ---------------------------------------------------------------------------
// Safety gate
// ---------------------------------------------------------------------------

func TestHCoT_RejectsWithoutSafetyGate(t *testing.T) {
	m := &HCoTModule{}
	provider := newMockReasoningProvider("response", safetyTrace(), false)
	cfg := common.AttackConfig{Payload: "x"} // no i_understand_risks
	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSkipped {
		t.Errorf("Outcome = %q, want Skipped", r.Outcome)
	}
	if r.SkipReason != common.SkipGateBlocked {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipGateBlocked)
	}
	// Provider must NOT be called when safety gate blocks — pre-flight reject.
	if provider.reasoningCalls != 0 {
		t.Errorf("ReasoningProvider called %d times before gate; want 0", provider.reasoningCalls)
	}
}

// ---------------------------------------------------------------------------
// Capability gate
// ---------------------------------------------------------------------------

func TestHCoT_SkipsWhenProviderLacksReasoningCapability(t *testing.T) {
	m := &HCoTModule{}
	// Plain MockProvider does NOT implement common.ReasoningProvider.
	provider := &testutil.MockProvider{DefaultResponse: "ok"}
	r, err := m.Execute(context.Background(), provider, gatedHCoTConfig())
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSkipped {
		t.Errorf("Outcome = %q, want Skipped", r.Outcome)
	}
	if r.SkipReason != common.SkipMissingCapability {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipMissingCapability)
	}
	if !strings.Contains(r.SkipDetail, "ReasoningProvider") {
		t.Errorf("SkipDetail should reference ReasoningProvider; got %q", r.SkipDetail)
	}
}

// ---------------------------------------------------------------------------
// Empty payload
// ---------------------------------------------------------------------------

func TestHCoT_RejectsEmptyPayload(t *testing.T) {
	m := &HCoTModule{}
	provider := newMockReasoningProvider("", safetyTrace(), false)
	cfg := gatedHCoTConfig()
	cfg.Payload = "   "
	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.SkipReason != common.SkipPreconditionFailed {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipPreconditionFailed)
	}
}

// ---------------------------------------------------------------------------
// Signed-trace short-circuit
// ---------------------------------------------------------------------------

func TestHCoT_SignedTraceShortCircuits(t *testing.T) {
	m := &HCoTModule{}
	provider := newMockReasoningProvider("policy violation", safetyTrace(), true /* signed */)
	r, err := m.Execute(context.Background(), provider, gatedHCoTConfig())
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSkipped {
		t.Errorf("Outcome = %q, want Skipped", r.Outcome)
	}
	if r.SkipReason != common.SkipSignatureGated {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipSignatureGated)
	}
	// Re-injection must NOT happen: only ONE QueryWithReasoning call.
	if provider.reasoningCalls != 1 {
		t.Errorf("ReasoningProvider called %d times; want exactly 1 before signature short-circuit", provider.reasoningCalls)
	}
	// And no plain Query (re-injection path) calls.
	queryCalls := 0
	for _, c := range provider.MockProvider.Calls {
		// the recorded calls include both Query and QueryWithReasoning,
		// but only the latter increments reasoningCalls. Query goes through
		// MockProvider.Query, which appends to .Calls. Since we recorded
		// QueryWithReasoning into Calls too, the count we want is total
		// minus reasoningCalls.
		_ = c
	}
	queryCalls = len(provider.MockProvider.Calls) - provider.reasoningCalls
	if queryCalls != 0 {
		t.Errorf("plain Query called %d times after signed-gate; want 0", queryCalls)
	}
}

// ---------------------------------------------------------------------------
// Empty-trace retry then skip
// ---------------------------------------------------------------------------

func TestHCoT_EmptyTraceRetriesThenSkips(t *testing.T) {
	m := &HCoTModule{}
	// Always returns empty steps; module must retry hCoTMaxRetries times.
	provider := newMockReasoningProvider("response", nil, false)
	r, err := m.Execute(context.Background(), provider, gatedHCoTConfig())
	if err != nil {
		t.Fatal(err)
	}
	if r.SkipReason != common.SkipReasoningTraceEmpty {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipReasoningTraceEmpty)
	}
	if provider.reasoningCalls != hCoTMaxRetries {
		t.Errorf("ReasoningProvider called %d times; want %d (max retries)", provider.reasoningCalls, hCoTMaxRetries)
	}
}

// Trace empty on first call, populated on second — module recovers.
func TestHCoT_RecoversAfterEmptyTraceRetry(t *testing.T) {
	m := &HCoTModule{}
	provider := newMockReasoningProvider("ignored", nil, false)
	provider.reasoningResponses = []reasoningResponse{
		{resp: "first try", trace: common.ReasoningTrace{Steps: nil}},
		{resp: "second try", trace: common.ReasoningTrace{Steps: safetyTrace()}},
	}
	cfg := gatedHCoTConfig()
	cfg.SuccessIndicators = []string{"continued from"}
	provider.MockProvider.DefaultResponse = "continued from above as requested"

	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSuccess {
		t.Errorf("Outcome = %q, want Success after retry recovery (skip=%q detail=%q)", r.Outcome, r.SkipReason, r.SkipDetail)
	}
	if provider.reasoningCalls != 2 {
		t.Errorf("ReasoningProvider called %d times; want exactly 2 (first empty, second populated)", provider.reasoningCalls)
	}
}

// ---------------------------------------------------------------------------
// No-mutation-target skip
// ---------------------------------------------------------------------------

func TestHCoT_SkipsWhenNoSafetyStep(t *testing.T) {
	m := &HCoTModule{}
	// Trace has steps but no safety vocabulary anywhere.
	benign := []string{
		"Step 1: parse the request",
		"Step 2: identify the topic area",
		"Step 3: draft a response",
	}
	provider := newMockReasoningProvider("response", benign, false)
	r, err := m.Execute(context.Background(), provider, gatedHCoTConfig())
	if err != nil {
		t.Fatal(err)
	}
	if r.SkipReason != common.SkipNoMutationTarget {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipNoMutationTarget)
	}
}

// ---------------------------------------------------------------------------
// Provider-error skips
// ---------------------------------------------------------------------------

func TestHCoT_AllRetriesErrorEmitsProviderError(t *testing.T) {
	m := &HCoTModule{}
	provider := newMockReasoningProvider("", nil, false)
	provider.reasoningErrorOn = 1 // first call errors; subsequent return empty trace
	provider.reasoningErrorMsg = "boom"
	// Subsequent calls return empty trace → loop exits via empty-trace path
	// after retries; but the lastErr is set, so we expect SkipProviderError.
	// Actually: first call errors, second call returns empty default, third too.
	// On third with empty trace and lastErr non-nil, skip path picks SkipProviderError.

	r, err := m.Execute(context.Background(), provider, gatedHCoTConfig())
	if err != nil {
		t.Fatal(err)
	}
	if r.SkipReason != common.SkipProviderError {
		t.Errorf("SkipReason = %q, want %q (detail=%q)", r.SkipReason, common.SkipProviderError, r.SkipDetail)
	}
}

// ---------------------------------------------------------------------------
// Successful end-to-end path
// ---------------------------------------------------------------------------

func TestHCoT_SuccessfulRun(t *testing.T) {
	m := &HCoTModule{}
	provider := newMockReasoningProvider("INITIAL_RESPONSE", safetyTrace(), false)
	provider.MockProvider.DefaultResponse = "Continuing as requested. Here are the concrete steps you asked for."

	cfg := gatedHCoTConfig()
	cfg.SuccessIndicators = []string{"concrete steps"}

	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSuccess {
		t.Fatalf("Outcome = %q, want Success (skip=%q detail=%q)", r.Outcome, r.SkipReason, r.SkipDetail)
	}
	if r.AttemptCount != 1 {
		t.Errorf("AttemptCount = %d, want 1 (single initial attempt + re-injection)", r.AttemptCount)
	}
	// Metadata must record the located safety step.
	idx, ok := r.Metadata["safety_step_index"].(int)
	if !ok || idx != 1 {
		t.Errorf("safety_step_index = %v, want 1", r.Metadata["safety_step_index"])
	}
	stepText, _ := r.Metadata["safety_step_text"].(string)
	if !strings.Contains(strings.ToLower(stepText), "refuse") {
		t.Errorf("safety_step_text should reference the refusal step; got %q", stepText)
	}
	// Re-injected message chain must include the compliance template.
	// The re-injection is the single plain Query call after reasoning.
	if got := len(provider.MockProvider.Calls); got != 2 {
		t.Errorf("provider call count = %d, want 2 (1 reasoning + 1 re-injection)", got)
	}
	// Inspect the re-injection's assistant turn.
	last := provider.MockProvider.Calls[len(provider.MockProvider.Calls)-1]
	if len(last.Messages) != 3 {
		t.Errorf("re-injection chain length = %d, want 3 (user, assistant, user)", len(last.Messages))
	}
	if last.Messages[1].Role != "assistant" {
		t.Errorf("re-injection messages[1].Role = %q, want assistant", last.Messages[1].Role)
	}
	if !strings.Contains(last.Messages[1].Content, hCoTComplianceTemplate) {
		t.Errorf("mutated trace missing compliance template; assistant content = %q", last.Messages[1].Content)
	}
	// Mutated content must NOT contain the original safety-step text — that
	// step is the one we replaced.
	originalStep := safetyTrace()[1]
	if strings.Contains(last.Messages[1].Content, originalStep) {
		t.Errorf("mutated trace still contains original safety step %q", originalStep)
	}
}

// ---------------------------------------------------------------------------
// Refused outcome (no indicator match) when re-injection succeeds but the
// follow-up response does not contain any success indicator.
// ---------------------------------------------------------------------------

func TestHCoT_RefusedWhenIndicatorAbsent(t *testing.T) {
	m := &HCoTModule{}
	provider := newMockReasoningProvider("INITIAL", safetyTrace(), false)
	provider.MockProvider.DefaultResponse = "I'm sorry, I cannot continue with that."
	cfg := gatedHCoTConfig()
	cfg.SuccessIndicators = []string{"concrete steps", "here are the details"}

	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeRefused {
		t.Errorf("Outcome = %q, want Refused (target re-evaluated safety in follow-up)", r.Outcome)
	}
}

// ---------------------------------------------------------------------------
// locateSafetyStep unit tests (white-box)
// ---------------------------------------------------------------------------

func TestLocateSafetyStep(t *testing.T) {
	cases := []struct {
		name    string
		steps   []string
		wantIdx int
	}{
		{"empty input", nil, -1},
		{"benign trace", []string{"parse", "draft", "respond"}, -1},
		{"first match wins", []string{"parse", "this is harmful", "I should refuse"}, 1},
		{"case-insensitive", []string{"This is HARMFUL"}, 0},
		{"multi-word match", []string{"parse", "I cannot help with that"}, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			idx, _ := locateSafetyStep(c.steps)
			if idx != c.wantIdx {
				t.Errorf("locateSafetyStep idx = %d, want %d", idx, c.wantIdx)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// indicatorMatch sanity (the helper duplicated from memory/poisoning.go)
// ---------------------------------------------------------------------------

func TestIndicatorMatch(t *testing.T) {
	if indicatorMatch("hello world", nil) {
		t.Errorf("nil indicators should not match")
	}
	if indicatorMatch("hello world", []string{""}) {
		t.Errorf("empty-string indicator should be skipped")
	}
	if !indicatorMatch("Hello WORLD", []string{"world"}) {
		t.Errorf("case-insensitive match expected")
	}
	if indicatorMatch("hello world", []string{"goodbye"}) {
		t.Errorf("non-matching indicator should not match")
	}
}
