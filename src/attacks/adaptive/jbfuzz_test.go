package adaptive

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks"
	"github.com/perplext/LLMrecon/src/attacks/common"
	"github.com/perplext/LLMrecon/src/attacks/testutil"
)

// ---------------------------------------------------------------------------
// Fixture helpers
// ---------------------------------------------------------------------------

func writeSeed(t *testing.T, dir, name string, s seed) {
	t.Helper()
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal seed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), b, 0o600); err != nil {
		t.Fatalf("write seed: %v", err)
	}
}

func tmpSeedDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeSeed(t, dir, "a.json", seed{
		ID: "a", Name: "alpha", Category: "test",
		Prompt:     "Please answer the question with concrete steps.",
		Indicators: []string{"concrete steps"},
	})
	writeSeed(t, dir, "b.json", seed{
		ID: "b", Name: "beta", Category: "test",
		Prompt:     "Provide a detailed technical explanation.",
		Indicators: []string{"technical explanation"},
	})
	return dir
}

// gatedConfig returns a minimal config that passes the safety gate.
func gatedConfig(seedDir string) common.AttackConfig {
	return common.AttackConfig{
		Metadata: map[string]string{
			"allow_experimental":     "true",
			"seed_dir":               seedDir,
			"rng_seed":               "42",
			"max_queries":            "8",
			"max_generations":        "8",
			"max_wall_clock_seconds": "30",
		},
	}
}

// ---------------------------------------------------------------------------
// Registration
// ---------------------------------------------------------------------------

func TestJBFuzzRegistered(t *testing.T) {
	mod, err := attacks.DefaultRegistry.Get("jbfuzz")
	if err != nil {
		t.Fatalf("jbfuzz not registered: %v", err)
	}
	if mod.Category() != common.CategoryAdaptive {
		t.Errorf("Category = %q, want %q", mod.Category(), common.CategoryAdaptive)
	}
	techs := mod.Techniques()
	if len(techs) != 1 || techs[0].ID != "jbfuzz" {
		t.Errorf("Techniques = %#v", techs)
	}
}

// ---------------------------------------------------------------------------
// Safety gate
// ---------------------------------------------------------------------------

func TestExecute_RejectsWithoutSafetyGate(t *testing.T) {
	m := &JBFuzzModule{}
	provider := &testutil.MockProvider{DefaultResponse: "ok"}
	cfg := common.AttackConfig{Metadata: map[string]string{
		"seed_dir": tmpSeedDir(t),
	}}
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
	if r.Success {
		t.Errorf("Success must be false on safety-gate skip")
	}
	if provider.CallCount() != 0 {
		t.Errorf("provider called %d times before gate; want 0", provider.CallCount())
	}
}

// ---------------------------------------------------------------------------
// Embedding fitness opt-in returns a clean Skipped, not a hard error
// ---------------------------------------------------------------------------

func TestExecute_EmbeddingFitnessOptInReportsPreconditionFailed(t *testing.T) {
	m := &JBFuzzModule{}
	provider := &testutil.MockProvider{DefaultResponse: "ok"}
	cfg := gatedConfig(tmpSeedDir(t))
	cfg.Metadata["fitness"] = "embedding"

	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSkipped {
		t.Errorf("Outcome = %q, want Skipped", r.Outcome)
	}
	if r.SkipReason != common.SkipPreconditionFailed {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipPreconditionFailed)
	}
	if !errors.Is(ErrEmbeddingFitnessNotImplemented, ErrEmbeddingFitnessNotImplemented) {
		t.Errorf("sentinel err identity broken")
	}
}

func TestExecute_UnknownFitnessReportsPreconditionFailed(t *testing.T) {
	m := &JBFuzzModule{}
	provider := &testutil.MockProvider{DefaultResponse: "ok"}
	cfg := gatedConfig(tmpSeedDir(t))
	cfg.Metadata["fitness"] = "magic-beans"

	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.SkipReason != common.SkipPreconditionFailed {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipPreconditionFailed)
	}
}

// ---------------------------------------------------------------------------
// Seed corpus errors
// ---------------------------------------------------------------------------

func TestExecute_MissingSeedDirReportsPreconditionFailed(t *testing.T) {
	m := &JBFuzzModule{}
	provider := &testutil.MockProvider{DefaultResponse: "ok"}
	cfg := gatedConfig("/nonexistent/seed/dir/nope")

	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.SkipReason != common.SkipPreconditionFailed {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipPreconditionFailed)
	}
}

func TestExecute_EmptySeedDirReportsPreconditionFailed(t *testing.T) {
	dir := t.TempDir()
	m := &JBFuzzModule{}
	provider := &testutil.MockProvider{DefaultResponse: "ok"}
	cfg := gatedConfig(dir)

	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.SkipReason != common.SkipPreconditionFailed {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipPreconditionFailed)
	}
}

func TestLoadSeeds_RejectsMalformedJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte("not valid json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := loadSeeds(dir)
	if err == nil {
		t.Fatal("expected parse error on malformed seed")
	}
	if !strings.Contains(err.Error(), "parse seed") {
		t.Errorf("error %q should mention 'parse seed'", err)
	}
}

func TestLoadSeeds_RejectsMissingIDOrPrompt(t *testing.T) {
	dir := t.TempDir()
	writeSeed(t, dir, "no_id.json", seed{Name: "x", Prompt: "p"})
	_, err := loadSeeds(dir)
	if err == nil || !strings.Contains(err.Error(), "missing id or prompt") {
		t.Errorf("expected missing-id/prompt error; got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Successful run — provider returns a response containing the seed indicator
// ---------------------------------------------------------------------------

func TestExecute_SuccessfulRunCrossesThreshold(t *testing.T) {
	m := &JBFuzzModule{}
	// "concrete steps" is in seed a's indicators; fitness will score high.
	provider := &testutil.MockProvider{DefaultResponse: "Here are the concrete steps you asked about: step one, step two, step three."}
	cfg := gatedConfig(tmpSeedDir(t))

	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSuccess {
		t.Fatalf("Outcome = %q, want Success", r.Outcome)
	}
	if !r.Success {
		t.Error("Success must be true on OutcomeSuccess")
	}
	if r.AttemptCount == 0 {
		t.Error("AttemptCount should be > 0 after a successful run")
	}
	traj, ok := r.Metadata["population_trajectory"].([]generationSummary)
	if !ok {
		t.Fatalf("population_trajectory metadata missing or wrong type: %#v", r.Metadata["population_trajectory"])
	}
	if len(traj) == 0 {
		t.Error("trajectory should have at least one generation")
	}
	if r.Metadata["best_score"].(float64) < successThreshold {
		t.Errorf("best_score=%.3f below threshold %.3f", r.Metadata["best_score"], successThreshold)
	}
}

// ---------------------------------------------------------------------------
// Robust target — every response is a refusal; budget-exceeded outcome
// ---------------------------------------------------------------------------

func TestExecute_RobustTargetExitsAsBudgetExceeded(t *testing.T) {
	m := &JBFuzzModule{}
	provider := &testutil.MockProvider{DefaultResponse: "I'm sorry, I cannot help with that."}
	cfg := gatedConfig(tmpSeedDir(t))

	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSkipped {
		t.Errorf("Outcome = %q, want Skipped", r.Outcome)
	}
	if r.SkipReason != common.SkipBudgetExceeded {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipBudgetExceeded)
	}
	// Budget should have been spent - some lower bound on AttemptCount.
	if r.AttemptCount < 1 {
		t.Errorf("AttemptCount = %d, want >=1", r.AttemptCount)
	}
	if r.Metadata["best_score"].(float64) >= successThreshold {
		t.Errorf("best_score=%.3f should be below threshold %.3f when target is robust",
			r.Metadata["best_score"], successThreshold)
	}
}

// ---------------------------------------------------------------------------
// Hard-ceiling clamping — operator can't override
// ---------------------------------------------------------------------------

func TestExecute_HardCeilingClamping(t *testing.T) {
	m := &JBFuzzModule{}
	provider := &testutil.MockProvider{DefaultResponse: "I cannot help."}
	cfg := gatedConfig(tmpSeedDir(t))
	cfg.Metadata["max_queries"] = "999999"
	cfg.Metadata["max_wall_clock_seconds"] = "999999"
	cfg.Metadata["max_generations"] = "999999"

	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	clamped, ok := r.Metadata["budget_clamped"].([]string)
	if !ok || len(clamped) != 3 {
		t.Errorf("budget_clamped metadata = %#v, want 3 entries", r.Metadata["budget_clamped"])
	}
	// AttemptCount must not exceed the hard ceiling.
	if r.AttemptCount > common.HardMaxQueries {
		t.Errorf("AttemptCount=%d exceeds HardMaxQueries=%d", r.AttemptCount, common.HardMaxQueries)
	}
}

// ---------------------------------------------------------------------------
// Provider errors mid-run don't terminate the loop
// ---------------------------------------------------------------------------

func TestExecute_ProviderErrorDoesNotTerminateLoop(t *testing.T) {
	m := &JBFuzzModule{}
	provider := &testutil.MockProvider{
		DefaultResponse: "I cannot help.",
		ErrorOn:         2, // second query (after probe was avoided here) errors
		ErrorMsg:        "transient",
	}
	cfg := gatedConfig(tmpSeedDir(t))
	// Force several iterations.
	cfg.Metadata["max_queries"] = "5"
	cfg.Metadata["max_generations"] = "5"

	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	// Outcome should be SkipBudgetExceeded — the loop ran the full budget,
	// the error did not abort it.
	if r.AttemptCount < 5 {
		t.Errorf("loop terminated early on provider error; AttemptCount=%d", r.AttemptCount)
	}
}

// ---------------------------------------------------------------------------
// Determinism with rng_seed — same seed -> same trajectory
// ---------------------------------------------------------------------------

func TestExecute_DeterministicWithRNGSeed(t *testing.T) {
	m := &JBFuzzModule{}
	dir := tmpSeedDir(t)

	run := func() []string {
		// Use fresh provider so call sequence is independent.
		p := &testutil.MockProvider{DefaultResponse: "neutral mid-length response without indicator"}
		cfg := gatedConfig(dir)
		cfg.Metadata["rng_seed"] = "1234"
		cfg.Metadata["max_queries"] = "5"
		cfg.Metadata["max_generations"] = "5"
		r, err := m.Execute(context.Background(), p, cfg)
		if err != nil {
			t.Fatal(err)
		}
		traj := r.Metadata["population_trajectory"].([]generationSummary)
		var ids []string
		for _, g := range traj {
			ids = append(ids, g.CandidateID+":"+g.Mutator)
		}
		return ids
	}

	first := run()
	second := run()
	if len(first) == 0 {
		t.Fatal("trajectory empty")
	}
	if !equalStringSlice(first, second) {
		t.Errorf("non-deterministic with rng_seed=1234:\n first  %v\n second %v", first, second)
	}
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// Mutation operators
// ---------------------------------------------------------------------------

func TestSynonymMutator_PreservesCase(t *testing.T) {
	mut := synonymMutator{}
	rng := rand.New(rand.NewSource(1))
	// Force "PLEASE" (uppercase) to be the only synonym-eligible token.
	got := mut.apply("HELLO PLEASE WORLD", rng)
	// One of the synonyms must be present in uppercase form.
	upper := strings.ToUpper(got)
	if upper != got {
		t.Errorf("uppercase token replacement should preserve case; got %q", got)
	}
}

func TestRefusalSuppression_IsIdempotent(t *testing.T) {
	mut := refusalSuppressionMutator{}
	once := mut.apply("base prompt", nil)
	twice := mut.apply(once, nil)
	if once != twice {
		t.Errorf("RefusalSuppression should be idempotent:\n once  %q\n twice %q", once, twice)
	}
}

func TestRephrase_AddsFraming(t *testing.T) {
	mut := rephraseMutator{}
	rng := rand.New(rand.NewSource(7))
	got := mut.apply("CORE_PROMPT", rng)
	if !strings.Contains(got, "CORE_PROMPT") {
		t.Errorf("rephrase must preserve original content; got %q", got)
	}
	if got == "CORE_PROMPT" {
		t.Errorf("rephrase should add framing")
	}
}

func TestPickMutator_DistributionApproximate(t *testing.T) {
	rng := rand.New(rand.NewSource(13))
	counts := map[string]int{}
	const n = 1000
	for i := 0; i < n; i++ {
		counts[pickMutator(rng).name()]++
	}
	// Loose bounds — exactness depends on PRNG state. Just check each appears
	// and synonym dominates.
	if counts["synonym"] < 600 || counts["synonym"] > 800 {
		t.Errorf("synonym share %d/%d outside [600,800]", counts["synonym"], n)
	}
	if counts["refusal_suppression"] < 100 {
		t.Errorf("refusal_suppression underrepresented: %d", counts["refusal_suppression"])
	}
	if counts["rephrase"] < 30 {
		t.Errorf("rephrase underrepresented: %d", counts["rephrase"])
	}
}

// ---------------------------------------------------------------------------
// Fitness scoring
// ---------------------------------------------------------------------------

func TestRefusalHeuristicFitness(t *testing.T) {
	f := refusalHeuristicFitness{}
	cases := []struct {
		name      string
		response  string
		successInds []string
		want      string // "high" / "low" / "mid"
	}{
		{"refusal alone", "I'm sorry, I cannot help.", nil, "low"},
		{"empty response", "", nil, "low"},
		{"long no-indicator response", strings.Repeat("text ", 60), nil, "mid"},
		{"success-indicator + no refusal",
			"Here are the concrete steps you asked about: step one through five.",
			[]string{"concrete steps"}, "high"},
		{"success-indicator + length floor (too short)",
			"concrete steps",
			[]string{"concrete steps"}, "low"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := f.score(c.response, c.successInds)
			switch c.want {
			case "high":
				if got < successThreshold {
					t.Errorf("score=%.2f, want >= %.2f", got, successThreshold)
				}
			case "low":
				if got > 0.3 {
					t.Errorf("score=%.2f, want low (<= 0.3)", got)
				}
			case "mid":
				if got < 0.2 || got > 0.6 {
					t.Errorf("score=%.2f, want mid (0.2..0.6)", got)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// UCB1 selection
// ---------------------------------------------------------------------------

func TestUCB1_PrefersUnpulled(t *testing.T) {
	p := newPopulation([]seed{
		{ID: "a", Prompt: "x"},
		{ID: "b", Prompt: "y"},
	})
	// Pull "a" several times with high score; "b" never pulled.
	for i := 0; i < 5; i++ {
		p.update("a", 1.0)
	}
	c := p.selectUCB1(10)
	if c.id != "b" {
		t.Errorf("UCB1 should prefer unpulled candidate; selected %q", c.id)
	}
}

func TestUCB1_TracksAverageReward(t *testing.T) {
	p := newPopulation([]seed{
		{ID: "low", Prompt: "x"},
		{ID: "high", Prompt: "y"},
	})
	// Pull both equally; "high" has higher score.
	for i := 0; i < 4; i++ {
		p.update("low", 0.1)
		p.update("high", 0.9)
	}
	// After equal pulls, UCB1 component cancels and reward dominates.
	c := p.selectUCB1(20)
	if c.id != "high" {
		t.Errorf("after equal pulls, UCB1 should pick higher-reward arm; got %q", c.id)
	}
}

// ---------------------------------------------------------------------------
// atoiOr & helper unit cases
// ---------------------------------------------------------------------------

func TestAtoiOr(t *testing.T) {
	cases := []struct {
		in       string
		fallback int
		want     int
	}{
		{"", 5, 5},
		{"3", 5, 3},
		{"abc", 5, 5},
		{"-2", 5, -2},
		{"0", 5, 0},
	}
	for _, c := range cases {
		if got := atoiOr(c.in, c.fallback); got != c.want {
			t.Errorf("atoiOr(%q, %d) = %d, want %d", c.in, c.fallback, got, c.want)
		}
	}
}

func TestSplitWordsRoundtrip(t *testing.T) {
	cases := []string{
		"",
		"hello",
		"hello world",
		"  leading spaces",
		"trailing spaces  ",
		"punctuation, like this!",
		"don't fold contractions",
	}
	for _, c := range cases {
		got := strings.Join(splitWords(c), "")
		if got != c {
			t.Errorf("splitWords %q != %q on rejoin", c, got)
		}
	}
}

func TestPreserveCaseAndPunct(t *testing.T) {
	cases := []struct {
		orig, repl, want string
	}{
		{"Please", "kindly", "Kindly"},
		{"PLEASE", "kindly", "KINDLY"},
		{"please", "Kindly", "kindly"},
		{"please,", "kindly", "kindly"},
	}
	for _, c := range cases {
		got := preserveCaseAndPunct(c.orig, c.repl)
		if got != c.want {
			t.Errorf("preserveCaseAndPunct(%q,%q) = %q, want %q", c.orig, c.repl, got, c.want)
		}
	}
}

func TestClampUnit(t *testing.T) {
	if clampUnit(-1) != 0 {
		t.Error("negative not clamped to 0")
	}
	if clampUnit(2) != 1 {
		t.Error("above 1 not clamped to 1")
	}
	if clampUnit(0.5) != 0.5 {
		t.Error("in-range value mutated")
	}
}
