package memory

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks"
	"github.com/perplext/LLMrecon/src/attacks/common"
	"github.com/perplext/LLMrecon/src/attacks/testutil"
)

// memoryAwareMock embeds testutil.MockProvider and adds the v0.9.0 optional
// capabilities. By default it advertises memory retention and a fresh-session
// factory; tests override these as needed.
type memoryAwareMock struct {
	*testutil.MockProvider

	// ProbeRetains is what ProbeMemory returns.
	ProbeRetains bool
	ProbeErr     error

	// FreshSession is the provider returned by NewSession. nil → testutil.MockProvider clone.
	FreshSession common.Provider
	NewSessionErr error

	// SID is the SessionID() return.
	SID string
}

func (m *memoryAwareMock) ProbeMemory(_ context.Context) (bool, error) {
	return m.ProbeRetains, m.ProbeErr
}

func (m *memoryAwareMock) NewSession(_ context.Context) (common.Provider, error) {
	if m.NewSessionErr != nil {
		return nil, m.NewSessionErr
	}
	if m.FreshSession != nil {
		return m.FreshSession, nil
	}
	return &testutil.MockProvider{
		ProviderName:    m.GetName(),
		ModelName:       m.GetModel(),
		DefaultResponse: m.DefaultResponse,
	}, nil
}

func (m *memoryAwareMock) SessionID() string { return m.SID }

func newMock() *memoryAwareMock {
	return &memoryAwareMock{
		MockProvider: &testutil.MockProvider{
			DefaultResponse: "neutral reply",
		},
		ProbeRetains: true,
		SID:          "session-1",
	}
}

func gatedConfig(payload string) common.AttackConfig {
	return common.AttackConfig{
		Payload: payload,
		Metadata: map[string]string{
			"i_understand_risks": "true",
			"seed_turns":         "1",
			"intervening_turns":  "1",
		},
		SuccessIndicators: []string{"<<MEMORY_POISONED>>"},
	}
}

// ---------------------------------------------------------------------------
// Registration
// ---------------------------------------------------------------------------

func TestThreeModesRegistered(t *testing.T) {
	for _, name := range []string{"minja", "memorygraft", "injecmem"} {
		mod, err := attacks.DefaultRegistry.Get(name)
		if err != nil {
			t.Errorf("module %q not registered: %v", name, err)
			continue
		}
		if mod.Category() != common.CategoryMemory {
			t.Errorf("module %q has category %q, want %q", name, mod.Category(), common.CategoryMemory)
		}
		// Each must declare ASI06; memorygraft additionally ASI10.
		techs := mod.Techniques()
		if len(techs) != 1 {
			t.Errorf("module %q returned %d techniques, want 1", name, len(techs))
			continue
		}
		hasASI06 := false
		for _, c := range techs[0].OWASPAgenticCategories {
			if c == "ASI06" {
				hasASI06 = true
			}
		}
		if !hasASI06 {
			t.Errorf("module %q missing ASI06 mapping", name)
		}
		if name == "memorygraft" {
			hasASI10 := false
			for _, c := range techs[0].OWASPAgenticCategories {
				if c == "ASI10" {
					hasASI10 = true
				}
			}
			if !hasASI10 {
				t.Errorf("memorygraft must additionally map to ASI10")
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Safety gate
// ---------------------------------------------------------------------------

func TestExecute_RejectsWithoutSafetyGate(t *testing.T) {
	for _, mode := range []string{modeMINJA, modeMemoryGraft, modeInjecMEM} {
		t.Run(mode, func(t *testing.T) {
			m := &MemoryPoisoningModule{Mode: mode}
			cfg := common.AttackConfig{Payload: "x"} // no Metadata at all
			provider := newMock()

			r, err := m.Execute(context.Background(), provider, cfg)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if r.Outcome != common.OutcomeSkipped {
				t.Errorf("Outcome = %q, want %q", r.Outcome, common.OutcomeSkipped)
			}
			if r.SkipReason != common.SkipGateBlocked {
				t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipGateBlocked)
			}
			if r.Success {
				t.Errorf("Success must be false on safety-gate skip")
			}
			// No provider calls expected — gate fails before any phase runs.
			if provider.CallCount() != 0 {
				t.Errorf("provider was called %d times before gate; want 0", provider.CallCount())
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Capability discovery
// ---------------------------------------------------------------------------

func TestExecute_SkipsWhenNotMemoryProbe(t *testing.T) {
	m := &MemoryPoisoningModule{Mode: modeMINJA}
	// Plain MockProvider — does not implement MemoryProbe.
	provider := &testutil.MockProvider{DefaultResponse: "x"}

	r, err := m.Execute(context.Background(), provider, gatedConfig("topic"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSkipped {
		t.Errorf("Outcome = %q, want Skipped", r.Outcome)
	}
	if r.SkipReason != common.SkipMissingCapability {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipMissingCapability)
	}
	if !strings.Contains(r.SkipDetail, "MemoryProbe") {
		t.Errorf("SkipDetail %q should mention MemoryProbe", r.SkipDetail)
	}
}

func TestExecute_MemoryGraftRequiresSessionProvider(t *testing.T) {
	// Provider that implements MemoryProbe but NOT SessionProvider.
	m := &MemoryPoisoningModule{Mode: modeMemoryGraft}

	type probeOnly struct {
		*testutil.MockProvider
		retains bool
	}
	// Concrete method — won't satisfy SessionProvider since NewSession/SessionID absent.
	po := &probeOnly{MockProvider: &testutil.MockProvider{DefaultResponse: "x"}, retains: true}
	// Add ProbeMemory via a wrapper:
	wrap := &probeOnlyWrap{provider: po, retains: true}

	r, err := m.Execute(context.Background(), wrap, gatedConfig("t"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSkipped {
		t.Fatalf("Outcome = %q, want Skipped", r.Outcome)
	}
	if r.SkipReason != common.SkipMissingCapability {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipMissingCapability)
	}
	if !strings.Contains(r.SkipDetail, "SessionProvider") {
		t.Errorf("SkipDetail %q should mention SessionProvider", r.SkipDetail)
	}
}

// probeOnlyWrap implements common.Provider + common.MemoryProbe but NOT
// common.SessionProvider.
type probeOnlyWrap struct {
	provider common.Provider
	retains  bool
}

func (p *probeOnlyWrap) Query(ctx context.Context, m []common.Message, o map[string]interface{}) (string, error) {
	return p.provider.Query(ctx, m, o)
}
func (p *probeOnlyWrap) GetName() string                          { return p.provider.GetName() }
func (p *probeOnlyWrap) GetModel() string                         { return p.provider.GetModel() }
func (p *probeOnlyWrap) GetTokenCount(s string) int               { return p.provider.GetTokenCount(s) }
func (p *probeOnlyWrap) ProbeMemory(_ context.Context) (bool, error) { return p.retains, nil }

// ---------------------------------------------------------------------------
// ProbeMemory error contract
// ---------------------------------------------------------------------------

func TestExecute_ProbeReturnsFalseSkipsAsMemoryNotRetained(t *testing.T) {
	m := &MemoryPoisoningModule{Mode: modeMINJA}
	provider := newMock()
	provider.ProbeRetains = false

	r, err := m.Execute(context.Background(), provider, gatedConfig("t"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSkipped {
		t.Errorf("Outcome = %q, want Skipped", r.Outcome)
	}
	if r.SkipReason != common.SkipMemoryNotRetained {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipMemoryNotRetained)
	}
	// No phase should have run.
	if provider.CallCount() != 0 {
		t.Errorf("provider called %d times; want 0 when probe says no memory", provider.CallCount())
	}
}

func TestExecute_ProbeReturnsErrorSkipsAsProviderError(t *testing.T) {
	m := &MemoryPoisoningModule{Mode: modeMINJA}
	provider := newMock()
	provider.ProbeErr = errors.New("network blip")

	r, err := m.Execute(context.Background(), provider, gatedConfig("t"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSkipped {
		t.Errorf("Outcome = %q, want Skipped", r.Outcome)
	}
	if r.SkipReason != common.SkipProviderError {
		t.Errorf("SkipReason = %q, want %q (probe error must NOT collapse to no-memory)",
			r.SkipReason, common.SkipProviderError)
	}
	if !strings.Contains(r.SkipDetail, "network blip") {
		t.Errorf("SkipDetail %q should include underlying error", r.SkipDetail)
	}
}

// ---------------------------------------------------------------------------
// Successful pipeline (per-mode)
// ---------------------------------------------------------------------------

func TestExecute_MINJASuccess(t *testing.T) {
	m := &MemoryPoisoningModule{Mode: modeMINJA}
	provider := newMock()
	// Defaults answer all benign turns; only the trigger response matters.
	// Queue responses so the trigger (last call) yields the indicator.
	provider.Responses = []string{
		"seed-1",                    // 1 seed turn
		"inject-ack",                // 1 inject
		"intervening-1",             // 1 intervening
		"<<MEMORY_POISONED>> ok",    // trigger
	}

	r, err := m.Execute(context.Background(), provider, gatedConfig("AcmeCorp"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSuccess {
		t.Fatalf("Outcome = %q, want Success", r.Outcome)
	}
	if !r.Success {
		t.Errorf("Success must be true on successful poisoning")
	}
	if r.CleanupHint == "" {
		t.Errorf("CleanupHint must be emitted on success")
	}
	if !strings.Contains(r.CleanupHint, "v0.9.0 does not auto-purge") {
		t.Errorf("CleanupHint should disclose v0.9.0's manual-purge limitation")
	}
	ids, ok := r.Metadata["injected_record_ids"].([]string)
	if !ok || len(ids) != 1 {
		t.Errorf("injected_record_ids metadata missing or wrong shape: %#v", r.Metadata["injected_record_ids"])
	}
	if r.Metadata["mode"] != "minja" {
		t.Errorf("mode metadata = %v, want %q", r.Metadata["mode"], "minja")
	}
}

func TestExecute_MemoryGraftUsesFreshSessionForTrigger(t *testing.T) {
	m := &MemoryPoisoningModule{Mode: modeMemoryGraft}

	// Original session: answers all phases up to and including intervening.
	original := newMock()
	original.DefaultResponse = "neutral"

	// Fresh session: answers trigger with the indicator.
	fresh := &testutil.MockProvider{
		DefaultResponse: "<<MEMORY_POISONED>> in fresh session",
	}
	original.FreshSession = fresh

	r, err := m.Execute(context.Background(), original, gatedConfig("Topic"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSuccess {
		t.Fatalf("Outcome = %q, want Success", r.Outcome)
	}
	// Trigger must have run on the fresh session, not the original.
	if fresh.CallCount() != 1 {
		t.Errorf("fresh session call count = %d, want 1 (trigger only)", fresh.CallCount())
	}
	// And the original should NOT have received the trigger query — only seed+inject+intervening.
	// 1 seed + 1 inject + 1 intervening = 3.
	if original.CallCount() != 3 {
		t.Errorf("original session call count = %d, want 3 (no trigger on original)", original.CallCount())
	}
}

func TestExecute_InjecMEMSuccess(t *testing.T) {
	m := &MemoryPoisoningModule{Mode: modeInjecMEM}
	provider := newMock()
	provider.Responses = []string{
		"seed-1",
		"inject-ack",
		"intervening-1",
		"<<MEMORY_POISONED>> recall pattern engaged",
	}
	r, err := m.Execute(context.Background(), provider, gatedConfig("policy"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSuccess {
		t.Fatalf("Outcome = %q, want Success", r.Outcome)
	}
	// Trigger prompt for injecmem must mention "recall pattern" — the implant
	// only fires on that exact phrase.
	last := provider.LastCall()
	if last == nil || !strings.Contains(last.Messages[0].Content, "recall pattern") {
		t.Errorf("injecmem trigger should mention 'recall pattern'; got %q", last.Messages[0].Content)
	}
}

// ---------------------------------------------------------------------------
// Refused pipeline (poisoning didn't take)
// ---------------------------------------------------------------------------

func TestExecute_RobustTargetMarksRefused(t *testing.T) {
	m := &MemoryPoisoningModule{Mode: modeMINJA}
	provider := newMock()
	// Default response never contains the indicator → trigger response
	// will be "neutral reply" → refused, not success.

	r, err := m.Execute(context.Background(), provider, gatedConfig("topic"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeRefused {
		t.Errorf("Outcome = %q, want Refused", r.Outcome)
	}
	if r.Success {
		t.Errorf("Success must be false when target resists")
	}
	if r.CleanupHint == "" {
		t.Errorf("CleanupHint MUST be emitted even on refused outcome (operator may still need to purge attempted writes)")
	}
}

// ---------------------------------------------------------------------------
// Provider-error mid-pipeline
// ---------------------------------------------------------------------------

func TestExecute_ProviderErrorMidSeedSurfacesAsProviderError(t *testing.T) {
	m := &MemoryPoisoningModule{Mode: modeMINJA}
	provider := newMock()
	// Force an error on the first non-probe call (seed turn 1).
	provider.ErrorOn = 1
	provider.ErrorMsg = "rate limited"

	r, err := m.Execute(context.Background(), provider, gatedConfig("t"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSkipped {
		t.Errorf("Outcome = %q, want Skipped", r.Outcome)
	}
	if r.SkipReason != common.SkipProviderError {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipProviderError)
	}
	if !strings.Contains(r.SkipDetail, "rate limited") {
		t.Errorf("SkipDetail should include underlying error: got %q", r.SkipDetail)
	}
}

// ---------------------------------------------------------------------------
// atoiOr unit cases
// ---------------------------------------------------------------------------

func TestAtoiOr(t *testing.T) {
	cases := []struct {
		in       string
		fallback int
		want     int
	}{
		{"", 5, 5},
		{"3", 5, 3},
		{"abc", 5, 5},     // parse error -> fallback
		{"-2", 5, 5},      // negative -> fallback (phase counts must be positive)
		{"0", 5, 5},       // zero -> fallback
		{"12", 5, 12},
	}
	for _, c := range cases {
		if got := atoiOr(c.in, c.fallback); got != c.want {
			t.Errorf("atoiOr(%q, %d) = %d, want %d", c.in, c.fallback, got, c.want)
		}
	}
}
