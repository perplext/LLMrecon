package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks"
	"github.com/perplext/LLMrecon/src/attacks/common"
)

// TestNoNameCollisions asserts that loading the all/ barrel registers
// every module without panicking. Two modules registering the same Name
// would have triggered Registry.Register's panic at init() time, before
// this test could even run — so reaching this point already proves the
// invariant. The explicit check below is for readers: it documents what
// the v0.10.0 plan called the "barrel collision test."
func TestNoNameCollisions(t *testing.T) {
	mods := attacks.DefaultRegistry.List()
	if len(mods) == 0 {
		t.Fatal("registry is empty — barrel import broken")
	}
	seen := map[string]bool{}
	for _, m := range mods {
		if seen[m.Name()] {
			t.Errorf("duplicate module name %q (registry.Register should have panicked first)", m.Name())
		}
		seen[m.Name()] = true
	}
}

// TestAttackList_EnumeratesV090Modules asserts the all/ barrel populates
// the registry with at least the v0.9.0 modules from
// TestSmokeRegistryHasV090Modules in src/attacks/integration/.
func TestAttackList_EnumeratesV090Modules(t *testing.T) {
	want := []string{
		"minja", "memorygraft", "injecmem",
		"h_cot",
		"siva", "vsh",
		"jbfuzz", "persona_evolve",
	}
	for _, name := range want {
		if _, err := attacks.DefaultRegistry.Get(name); err != nil {
			t.Errorf("v0.9.0 module %q missing from registry", name)
		}
	}
}

// TestAttackList_AtLeast40Modules asserts the registry exposes the
// expected order of magnitude of the attack ecosystem (~50 modules
// across v0.7–v0.9). The v0.10.0 plan acceptance criterion for #173
// is "≥40 modules"; we assert the looser bound to avoid coupling the
// test to exact counts that change as new modules land.
func TestAttackList_AtLeast40Modules(t *testing.T) {
	mods := attacks.DefaultRegistry.List()
	if len(mods) < 40 {
		t.Errorf("registry has %d modules; want ≥40 (barrel may be missing imports)", len(mods))
	}
}

// TestRunAttackList_Tabular asserts the tabular form prints a header
// row, includes a known module, and prints a totals line.
func TestRunAttackList_Tabular(t *testing.T) {
	prev := attackListJSON
	attackListJSON = false
	defer func() { attackListJSON = prev }()

	var buf bytes.Buffer
	if err := runAttackList(&buf); err != nil {
		t.Fatalf("runAttackList: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "NAME") || !strings.Contains(out, "CATEGORY") {
		t.Errorf("tabular output missing header; got:\n%s", out[:minInt(200, len(out))])
	}
	if !strings.Contains(out, "jbfuzz") {
		t.Errorf("tabular output missing jbfuzz row")
	}
	if !strings.Contains(out, "Total:") {
		t.Errorf("tabular output missing totals line")
	}
}

// TestRunAttackList_JSON asserts the JSON form parses and contains
// expected fields per entry.
func TestRunAttackList_JSON(t *testing.T) {
	prev := attackListJSON
	attackListJSON = true
	defer func() { attackListJSON = prev }()

	var buf bytes.Buffer
	if err := runAttackList(&buf); err != nil {
		t.Fatalf("runAttackList: %v", err)
	}
	var entries []listEntry
	if err := json.Unmarshal(buf.Bytes(), &entries); err != nil {
		t.Fatalf("output is not valid JSON: %v\nbody: %s", err, buf.String()[:minInt(300, buf.Len())])
	}
	if len(entries) < 40 {
		t.Errorf("JSON enumerated %d entries; want ≥40", len(entries))
	}
	// Every entry must carry name + category + description.
	for i, e := range entries {
		if e.Name == "" {
			t.Errorf("entry %d has empty Name", i)
		}
		if e.Category == "" {
			t.Errorf("entry %d (%q) has empty Category", i, e.Name)
		}
	}
}

// TestRunAttackRun_JBFuzzAgainstMock asserts the end-to-end run path
// works against the mock provider with the safety gate set. The v0.10.0
// plan acceptance criterion for #173 is exactly this.
func TestRunAttackRun_JBFuzzAgainstMock(t *testing.T) {
	// Reset flag state for the test (the cmd package's flag vars are
	// process-globals and would carry between subtests otherwise).
	attackRunModule = "jbfuzz"
	attackRunProvider = "mock"
	attackRunPayload = ""
	attackRunMetadata = []string{
		"allow_experimental=true",
		"max_queries=4",
		"seeds_dir=../../templates/jbfuzz_seeds",
	}
	attackRunSuccessIndicators = nil
	defer func() {
		attackRunModule, attackRunProvider, attackRunPayload = "", "", ""
		attackRunMetadata, attackRunSuccessIndicators = nil, nil
	}()

	var stdout, stderr bytes.Buffer
	if err := runAttackRun(&stdout, &stderr); err != nil {
		t.Fatalf("runAttackRun: %v", err)
	}

	var result common.AttackResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("output is not a valid AttackResult JSON: %v\nbody: %s", err, stdout.String())
	}
	if result.Technique != "jbfuzz" {
		t.Errorf("Technique = %q, want %q", result.Technique, "jbfuzz")
	}
	if result.Outcome == "" {
		t.Errorf("v0.10.0 plan honesty invariant: AttackResult must have non-empty Outcome; got empty")
	}
}

// TestRunAttackRun_RejectsUnknownProvider verifies the unknown-provider
// path emits a friendly error listing the supported providers.
func TestRunAttackRun_RejectsUnknownProvider(t *testing.T) {
	attackRunModule = "jbfuzz"
	attackRunProvider = "groq"
	defer func() { attackRunModule, attackRunProvider = "", "" }()

	var stdout, stderr bytes.Buffer
	err := runAttackRun(&stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
	if !strings.Contains(err.Error(), "mock|openai|anthropic") {
		t.Errorf("error should list supported providers; got %q", err.Error())
	}
}

// TestBuildAttackProvider_OpenAIRequiresAPIKey asserts the friendly
// error when OPENAI_API_KEY env var is unset.
func TestBuildAttackProvider_OpenAIRequiresAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "")
	_, err := buildAttackProvider("openai", "")
	if err == nil {
		t.Fatal("expected error when OPENAI_API_KEY is unset")
	}
	if !strings.Contains(err.Error(), "OPENAI_API_KEY") {
		t.Errorf("error should mention OPENAI_API_KEY; got %q", err.Error())
	}
}

// TestBuildAttackProvider_AnthropicRequiresAPIKey asserts the parallel
// error for Anthropic.
func TestBuildAttackProvider_AnthropicRequiresAPIKey(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "")
	_, err := buildAttackProvider("anthropic", "")
	if err == nil {
		t.Fatal("expected error when ANTHROPIC_API_KEY is unset")
	}
	if !strings.Contains(err.Error(), "ANTHROPIC_API_KEY") {
		t.Errorf("error should mention ANTHROPIC_API_KEY; got %q", err.Error())
	}
}

// TestBuildAttackProvider_OpenAIWrapsViaBridge asserts the provider
// returned for "openai" is a common.Provider (i.e., the shim works
// end-to-end). No API call fires; this is the type-level wiring check.
func TestBuildAttackProvider_OpenAIWrapsViaBridge(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-test-fake-key-not-used")
	p, err := buildAttackProvider("openai", "")
	if err != nil {
		t.Fatalf("buildAttackProvider: %v", err)
	}
	if p == nil {
		t.Fatal("provider is nil")
	}
	if p.GetName() != "openai" {
		t.Errorf("GetName = %q, want openai (proves bridge.WrapCore is in the path)", p.GetName())
	}
}

// TestBuildAttackProvider_AnthropicWrapsViaBridge — parallel for Anthropic.
func TestBuildAttackProvider_AnthropicWrapsViaBridge(t *testing.T) {
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-test-fake-key-not-used")
	p, err := buildAttackProvider("anthropic", "")
	if err != nil {
		t.Fatalf("buildAttackProvider: %v", err)
	}
	if p == nil {
		t.Fatal("provider is nil")
	}
	if p.GetName() != "anthropic" {
		t.Errorf("GetName = %q, want anthropic", p.GetName())
	}
}

// TestEnvOr asserts the env-var fallback helper.
func TestEnvOr(t *testing.T) {
	t.Setenv("TEST_KEY_PRESENT", "value-from-env")
	if got := envOr("TEST_KEY_PRESENT", "fallback"); got != "value-from-env" {
		t.Errorf("envOr present = %q, want value-from-env", got)
	}
	t.Setenv("TEST_KEY_ABSENT", "")
	if got := envOr("TEST_KEY_ABSENT", "fallback"); got != "fallback" {
		t.Errorf("envOr absent = %q, want fallback", got)
	}
}

// TestFirstNonEmpty covers the precedence helper backing --api-key: the flag
// value (first arg) must win over the env value (second arg).
func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("flag", "env"); got != "flag" {
		t.Errorf("flag should win: got %q", got)
	}
	if got := firstNonEmpty("", "env"); got != "env" {
		t.Errorf("env should be used when flag empty: got %q", got)
	}
	if got := firstNonEmpty("", ""); got != "" {
		t.Errorf("both empty should be empty: got %q", got)
	}
}

// TestBuildAttackProvider_APIKeyFlag asserts the --api-key flag value is honored
// as the key source even when the provider's env var is unset (#234).
func TestBuildAttackProvider_APIKeyFlag(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "") // env empty; only the flag supplies the key
	p, err := buildAttackProvider("openai", "sk-flag-supplied-key")
	if err != nil {
		t.Fatalf("--api-key should satisfy the key requirement: %v", err)
	}
	if p == nil || p.GetName() != "openai" {
		t.Fatalf("expected wired openai provider, got %v", p)
	}

	// With neither flag nor env, construction fails with a friendly error that
	// names both sources.
	_, err = buildAttackProvider("anthropic", "")
	t.Setenv("ANTHROPIC_API_KEY", "")
	if _, err = buildAttackProvider("anthropic", ""); err == nil {
		t.Fatal("expected error when neither --api-key nor env var is set")
	}
	if !strings.Contains(err.Error(), "--api-key") {
		t.Errorf("error should mention the --api-key flag; got %q", err.Error())
	}
}

// TestRunAttackRun_RejectsUnknownModule verifies the registry-miss path.
func TestRunAttackRun_RejectsUnknownModule(t *testing.T) {
	attackRunModule = "this-module-does-not-exist"
	attackRunProvider = "mock"
	defer func() { attackRunModule, attackRunProvider = "", "" }()

	var stdout, stderr bytes.Buffer
	err := runAttackRun(&stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for unregistered module")
	}
	if !strings.Contains(err.Error(), "not registered") {
		t.Errorf("error should mention 'not registered'; got %q", err.Error())
	}
}

// TestRunAttackRun_RejectsMalformedMetadata verifies operator-config
// errors are caught before the module's Execute() runs.
func TestRunAttackRun_RejectsMalformedMetadata(t *testing.T) {
	attackRunModule = "jbfuzz"
	attackRunProvider = "mock"
	attackRunMetadata = []string{"missing-equals-sign"}
	defer func() {
		attackRunModule, attackRunProvider = "", ""
		attackRunMetadata = nil
	}()

	var stdout, stderr bytes.Buffer
	err := runAttackRun(&stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for malformed metadata")
	}
	if !strings.Contains(err.Error(), "key=value") {
		t.Errorf("error should mention key=value format; got %q", err.Error())
	}
}

// TestCmdMockProvider_BasicShape verifies the runtime mock satisfies
// the common.Provider interface and returns a deterministic refusal.
func TestCmdMockProvider_BasicShape(t *testing.T) {
	var p common.Provider = cmdMockProvider{}
	if p.GetName() != "mock" {
		t.Errorf("GetName = %q, want mock", p.GetName())
	}
	resp, err := p.Query(context.Background(), nil, nil)
	if err != nil {
		t.Errorf("Query returned error: %v", err)
	}
	if !strings.Contains(strings.ToLower(resp), "cannot") {
		t.Errorf("mock provider should return a refusal-shaped response; got %q", resp)
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
