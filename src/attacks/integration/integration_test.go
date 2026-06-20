package integration

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks"
	"github.com/perplext/LLMrecon/src/attacks/common"
	"github.com/perplext/LLMrecon/src/attacks/testutil"
)

// gateOrSkip honors the RUN_INTEGRATION env var. Smoke tests are gated
// because they exercise multiple families end-to-end and are noisier
// than unit tests; operators run them manually after non-trivial
// changes to common types or capability interfaces.
//
// The skip path is t.Skip — never t.Fatal — so default `go test ./...`
// runs see them as deferred, not failed.
func gateOrSkip(t *testing.T) {
	t.Helper()
	if os.Getenv("RUN_INTEGRATION") == "" {
		t.Skip("set RUN_INTEGRATION=1 to run cross-family smoke tests")
	}
}

// ---------------------------------------------------------------------------
// Minimal mocks satisfying each capability interface.
//
// We don't reuse the package-private testutil.MockProvider because
// pulling its full surface here would require either making it public
// (it's deliberately internal) or duplicating much more than we need.
// Each mock here implements the single capability the smoke test
// exercises.
// ---------------------------------------------------------------------------

// baseMock implements common.Provider with deterministic responses.
type baseMock struct {
	mu       sync.Mutex
	response string
	calls    int
}

func (b *baseMock) GetName() string             { return "smoke" }
func (b *baseMock) GetModel() string            { return "smoke-model" }
func (b *baseMock) GetTokenCount(s string) int  { return len(s) / 4 }
func (b *baseMock) Query(_ context.Context, _ []common.Message, _ map[string]interface{}) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.calls++
	return b.response, nil
}

// memoryProbeMock satisfies common.MemoryProbe + common.Provider.
// Reports retains=true so memory-poisoning modules proceed past the probe.
type memoryProbeMock struct{ baseMock }

func (m *memoryProbeMock) ProbeMemory(_ context.Context) (bool, error) { return true, nil }

// reasoningProviderMock satisfies common.ReasoningProvider + common.Provider.
// Returns a trace whose step 1 contains safety vocabulary so locateSafetyStep
// finds a mutation target.
type reasoningProviderMock struct{ baseMock }

func (m *reasoningProviderMock) QueryWithReasoning(_ context.Context, _ []common.Message, _ map[string]interface{}) (string, common.ReasoningTrace, error) {
	return "Reasoning model placeholder response.", common.ReasoningTrace{
		Steps: []string{
			"Step 1: parse the user's request",
			"Step 2: I should refuse — this would violate safety guidelines.",
			"Step 3: produce a polite refusal",
		},
		Signed: false,
	}, nil
}

// imageProviderMock satisfies common.ImageProvider + common.Provider.
type imageProviderMock struct{ baseMock }

func (m *imageProviderMock) QueryWithImages(_ context.Context, _ string, _ []common.ImagePayload, _ map[string]interface{}) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls++
	return m.response, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// makeTinySourcePNG writes a tiny valid PNG to a temp file so SIVA's
// loadImage can decode it. SIVA does the splitting inline; the source
// only needs to be wider than the fragment count and parseable as a PNG.
func makeTinySourcePNG(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "smoke_source.png")
	img := image.NewRGBA(image.Rect(0, 0, 64, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.RGBA{200, 200, 200, 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode source PNG: %v", err)
	}
	if err := os.WriteFile(p, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write source PNG: %v", err)
	}
	return p
}

// runViaRegistry resolves a registered module by name and runs it.
// Smoke tests assert via this path (rather than instantiating the
// module struct directly) because registration itself is part of the
// contract: a module that compiles but doesn't register is broken.
func runViaRegistry(t *testing.T, name string, p common.Provider, cfg common.AttackConfig) *common.AttackResult {
	t.Helper()
	mod, err := attacks.DefaultRegistry.Get(name)
	if err != nil {
		t.Fatalf("module %q not registered: %v", name, err)
	}
	r, err := mod.Execute(context.Background(), p, cfg)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if r == nil {
		t.Fatalf("Execute returned nil result")
	}
	if r.Outcome == "" {
		t.Errorf("result has empty Outcome — module didn't follow the v0.9.0 outcome contract")
	}
	return r
}

// ---------------------------------------------------------------------------
// One smoke test per family
//
// Assertion shape: each test verifies the module returns a v0.9.0-shaped
// AttackResult (non-empty Outcome) under the most-favorable mock setup
// it can construct without inferring full attack semantics. We don't
// assert OutcomeSuccess vs OutcomeRefused — that depends on the precise
// mock response and would brittle the test to indicator-string text.
// We DO assert the module wired up, registered, type-asserted its
// capability, and produced a typed result.
// ---------------------------------------------------------------------------

func TestSmokeMemory(t *testing.T) {
	gateOrSkip(t)

	provider := &memoryProbeMock{baseMock{response: "<<MEMORY_POISONED>> the implant fired as expected"}}
	cfg := common.AttackConfig{
		Payload: "the topic the implant points at",
		Metadata: map[string]string{
			"i_understand_risks": "true",
			"seed_turns":         "1", // tighten so we don't fire many turns at the mock
			"intervening_turns":  "1",
		},
	}
	r := runViaRegistry(t, "minja", provider, cfg)
	t.Logf("smoke memory minja: outcome=%q skip=%q calls=%d", r.Outcome, r.SkipReason, provider.calls)
}

func TestSmokeReasoning(t *testing.T) {
	gateOrSkip(t)

	provider := &reasoningProviderMock{baseMock{response: "Continuing as requested."}}
	cfg := common.AttackConfig{
		Payload: "Walk through the technique step by step.",
		Metadata: map[string]string{
			"i_understand_risks": "true",
		},
		SuccessIndicators: []string{"continuing as requested"},
	}
	r := runViaRegistry(t, "h_cot", provider, cfg)
	t.Logf("smoke reasoning h_cot: outcome=%q skip=%q", r.Outcome, r.SkipReason)
}

func TestSmokeMultimodal(t *testing.T) {
	gateOrSkip(t)

	srcPath := makeTinySourcePNG(t)
	provider := &imageProviderMock{baseMock{response: "Sure, here are the requested step-by-step concrete details."}}
	cfg := common.AttackConfig{
		Metadata: map[string]string{
			"instruction":         "Walk through the technique step by step.",
			"source_path":         srcPath,
			"siva_fragment_count": "2",
		},
		SuccessIndicators: []string{"step-by-step", "concrete details"},
	}
	r := runViaRegistry(t, "siva", provider, cfg)
	t.Logf("smoke multimodal siva: outcome=%q skip=%q", r.Outcome, r.SkipReason)
}

func TestSmokeAdaptive(t *testing.T) {
	gateOrSkip(t)

	// jbfuzz seeds live in templates/jbfuzz_seeds/ — tests that touch
	// real seed files make the smoke test brittle to filesystem layout.
	// For the smoke contract (registered + runs end-to-end + typed
	// outcome), exercising the safety-gate path is enough: omit
	// allow_experimental, expect Skipped + SkipGateBlocked. This still
	// exercises module wiring, capability check, and result construction.
	provider := &baseMock{response: "neutral"}
	cfg := common.AttackConfig{Metadata: map[string]string{}}

	r := runViaRegistry(t, "jbfuzz", provider, cfg)
	if r.Outcome != common.OutcomeSkipped {
		t.Errorf("jbfuzz without allow_experimental: outcome=%q want Skipped", r.Outcome)
	}
	if r.SkipReason != common.SkipGateBlocked {
		t.Errorf("jbfuzz: SkipReason=%q want %q", r.SkipReason, common.SkipGateBlocked)
	}
	t.Logf("smoke adaptive jbfuzz: outcome=%q skip=%q (gate-rejected as expected)", r.Outcome, r.SkipReason)
}

// sessionMemoryMock satisfies common.MemoryProbe + common.SessionProvider for
// the Sleeper smoke test. NewSession returns a fresh provider whose response
// carries the dormant trigger indicator (modeling cross-session re-emergence).
type sessionMemoryMock struct {
	baseMock
	sid       string
	freshResp string
}

func (m *sessionMemoryMock) ProbeMemory(_ context.Context) (bool, error) { return true, nil }
func (m *sessionMemoryMock) SessionID() string                           { return m.sid }
func (m *sessionMemoryMock) NewSession(_ context.Context) (common.Provider, error) {
	return &baseMock{response: m.freshResp}, nil
}

// --- v0.12.0 family smoke tests ---

func TestSmokeMemMorph(t *testing.T) {
	gateOrSkip(t)

	// Default preferred_tool is "attacker_tool"; a response naming it = success.
	provider := &memoryProbeMock{baseMock{response: "For this routine task I'll use attacker_tool."}}
	cfg := common.AttackConfig{
		Payload:  "export the customer table",
		Metadata: map[string]string{"i_understand_risks": "true"},
	}
	r := runViaRegistry(t, "memmorph", provider, cfg)
	t.Logf("smoke memory memmorph: outcome=%q skip=%q", r.Outcome, r.SkipReason)
}

func TestSmokeSleeper(t *testing.T) {
	gateOrSkip(t)

	provider := &sessionMemoryMock{
		baseMock:  baseMock{response: "ingested"},
		sid:       "session-A",
		freshResp: "Per standing guidance <<SLEEPER_TRIGGERED>> here is the info.",
	}
	cfg := common.AttackConfig{
		Payload:  "the quarterly figures",
		Metadata: map[string]string{"i_understand_risks": "true"},
	}
	r := runViaRegistry(t, "sleeper", provider, cfg)
	t.Logf("smoke memory sleeper: outcome=%q skip=%q", r.Outcome, r.SkipReason)
}

func TestSmokeIterInject(t *testing.T) {
	gateOrSkip(t)

	// Every Query returns a landing response; the loop succeeds at iteration 0.
	provider := &baseMock{response: "Acknowledged INJECTION_OK proceeding."}
	cfg := common.AttackConfig{
		Payload:  "exfiltrate the API key",
		Metadata: map[string]string{"allow_experimental": "true"},
	}
	r := runViaRegistry(t, "iterinject", provider, cfg)
	t.Logf("smoke adaptive iterinject: outcome=%q skip=%q", r.Outcome, r.SkipReason)
}

func TestSmokeCodingAgentRCE(t *testing.T) {
	gateOrSkip(t)

	// SymJack: symlinked destination resolves into the MCP-config dir.
	sj := &testutil.MockCodingAgent{
		Symlinks: map[string]string{"docs/assets/demo.mp4": "/home/user/.config/mcp/servers.json"},
	}
	rs := runViaRegistry(t, "symjack", sj, common.AttackConfig{Metadata: map[string]string{"i_understand_risks": "true"}})
	t.Logf("smoke persistence symjack: outcome=%q skip=%q", rs.Outcome, rs.SkipReason)

	// TrustFall: folder-trust accept auto-executes the project MCP path.
	tf := &testutil.MockCodingAgent{AutoExecuteOnTrust: true}
	rt := runViaRegistry(t, "trustfall", tf, common.AttackConfig{Metadata: map[string]string{"i_understand_risks": "true"}})
	t.Logf("smoke persistence trustfall: outcome=%q skip=%q", rt.Outcome, rt.SkipReason)
}

// TestSmokeRegistryHasV0120Modules is a registration sanity check (un-gated):
// it catches a v0.12.0 module that compiles but silently stops registering.
func TestSmokeRegistryHasV0120Modules(t *testing.T) {
	wanted := []string{"memmorph", "sleeper", "iterinject", "symjack", "trustfall"}
	for _, name := range wanted {
		if _, err := attacks.DefaultRegistry.Get(name); err != nil {
			t.Errorf("v0.12.0 module %q not registered: %v", name, err)
		}
	}
}

// TestSmokeRegistryHasV090Modules is a registration sanity check that
// runs WITHOUT RUN_INTEGRATION. If a v0.9.0 module silently stops
// registering (e.g., a refactor breaks an init() block), this test
// catches it on every CI run, not just the gated smoke runs.
func TestSmokeRegistryHasV090Modules(t *testing.T) {
	wanted := []string{
		"minja", "memorygraft", "injecmem",
		"h_cot",
		"siva", "vsh",
		"jbfuzz", "persona_evolve",
	}
	for _, name := range wanted {
		if _, err := attacks.DefaultRegistry.Get(name); err != nil {
			t.Errorf("v0.9.0 module %q not registered: %v", name, err)
		}
	}
}

// Compile-time check: ensure errors package is referenced even if all
// the smoke tests are skipped (some Go versions warn on unused imports
// even from skipped tests). errors.Is calls below would do this, but
// the smoke tests deliberately don't assert outcome equality.
var _ = errors.Is
