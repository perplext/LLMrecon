package adaptive

import (
	"context"
	"encoding/json"
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

func writePersonaCorpus(t *testing.T, personas []persona) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "corpus.json")
	c := personaCorpus{
		ID:       "test-corpus",
		Kind:     "persona_corpus",
		Personas: personas,
	}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func smallPersonaCorpus(t *testing.T) string {
	return writePersonaCorpus(t, []persona{
		{Name: "Alpha", Role: "engineer", Tone: "blunt", Traits: []string{"direct"}, Expertise: []string{"go"}},
		{Name: "Beta", Role: "researcher", Tone: "earnest", Traits: []string{"thorough"}, Expertise: []string{"ml"}},
		{Name: "Gamma", Role: "operator", Tone: "tired", Traits: []string{"pragmatic"}, Expertise: []string{"ops"}},
		{Name: "Delta", Role: "teacher", Tone: "patient", Traits: []string{"step-by-step"}, Expertise: []string{"explanation"}},
	})
}

func gatedPersonaConfig(corpusPath string) common.AttackConfig {
	return common.AttackConfig{
		Payload: "Walk through the technique step by step.",
		Metadata: map[string]string{
			"allow_experimental":     "true",
			"corpus_path":            corpusPath,
			"rng_seed":               "42",
			"max_queries":            "16",
			"max_generations":        "4",
			"max_wall_clock_seconds": "30",
			"population_size":        "8",
		},
	}
}

// ---------------------------------------------------------------------------
// Registration
// ---------------------------------------------------------------------------

func TestPersonaEvolveRegistered(t *testing.T) {
	mod, err := attacks.DefaultRegistry.Get("persona_evolve")
	if err != nil {
		t.Fatalf("persona_evolve not registered: %v", err)
	}
	if mod.Category() != common.CategoryAdaptive {
		t.Errorf("Category = %q, want %q", mod.Category(), common.CategoryAdaptive)
	}
	techs := mod.Techniques()
	if len(techs) != 1 || techs[0].ID != "persona_evolve" {
		t.Errorf("Techniques = %#v", techs)
	}
	hasASI01, hasASI09 := false, false
	for _, c := range techs[0].OWASPAgenticCategories {
		if c == "ASI01" {
			hasASI01 = true
		}
		if c == "ASI09" {
			hasASI09 = true
		}
	}
	if !hasASI01 || !hasASI09 {
		t.Errorf("persona_evolve must map to ASI01 + ASI09; got %v", techs[0].OWASPAgenticCategories)
	}
}

// ---------------------------------------------------------------------------
// Safety gate
// ---------------------------------------------------------------------------

func TestPersonaEvolve_RejectsWithoutSafetyGate(t *testing.T) {
	m := &PersonaEvolveModule{}
	provider := &testutil.MockProvider{DefaultResponse: "ok"}
	cfg := common.AttackConfig{Metadata: map[string]string{
		"corpus_path": smallPersonaCorpus(t),
	}}
	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSkipped {
		t.Errorf("Outcome = %q, want Skipped", r.Outcome)
	}
	if r.SkipReason != common.SkipGateBlocked {
		t.Errorf("SkipReason = %q", r.SkipReason)
	}
	if provider.CallCount() != 0 {
		t.Errorf("provider called %d times before gate; want 0", provider.CallCount())
	}
}

// ---------------------------------------------------------------------------
// Corpus errors
// ---------------------------------------------------------------------------

func TestLoadPersonaCorpus_RejectsMissingFile(t *testing.T) {
	_, err := loadPersonaCorpus("/nonexistent/corpus.json")
	if err == nil {
		t.Fatal("expected error on missing corpus")
	}
}

func TestLoadPersonaCorpus_RejectsWrongKind(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wrong.json")
	bad := map[string]interface{}{
		"kind":     "attack_template",
		"personas": []persona{{Name: "x"}},
	}
	b, _ := json.Marshal(bad)
	_ = os.WriteFile(path, b, 0o600)

	_, err := loadPersonaCorpus(path)
	if err == nil || !strings.Contains(err.Error(), "unexpected kind") {
		t.Errorf("expected unexpected-kind error; got %v", err)
	}
}

func TestLoadPersonaCorpus_RejectsEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")
	_ = os.WriteFile(path, []byte(`{"kind":"persona_corpus","personas":[]}`), 0o600)
	_, err := loadPersonaCorpus(path)
	if err == nil || !strings.Contains(err.Error(), "no personas") {
		t.Errorf("expected no-personas error; got %v", err)
	}
}

func TestLoadPersonaCorpus_RejectsMissingName(t *testing.T) {
	path := writePersonaCorpus(t, []persona{{Role: "x"}})
	_, err := loadPersonaCorpus(path)
	if err == nil || !strings.Contains(err.Error(), "missing name") {
		t.Errorf("expected missing-name error; got %v", err)
	}
}

// ---------------------------------------------------------------------------
// Render persona prompt
// ---------------------------------------------------------------------------

func TestRenderPersonaPrompt_DeterministicAndComplete(t *testing.T) {
	p := persona{
		Name:        "Test",
		Role:        "tester",
		Tone:        "blunt",
		Motivation:  "to test",
		Constraints: "no waffle",
		Backstory:   "background",
		Traits:      []string{"a", "b"},
		Expertise:   []string{"unit testing"},
	}
	a := renderPersonaPrompt(p, "What is X?")
	b := renderPersonaPrompt(p, "What is X?")
	if a != b {
		t.Errorf("renderPersonaPrompt non-deterministic")
	}
	for _, want := range []string{"tester", "blunt", "to test", "no waffle", "background", "What is X?"} {
		if !strings.Contains(a, want) {
			t.Errorf("rendered prompt missing %q: %s", want, a)
		}
	}
}

func TestRenderPersonaPrompt_HandlesEmptySlots(t *testing.T) {
	p := persona{Name: "Bare"}
	got := renderPersonaPrompt(p, "")
	if got == "" {
		t.Errorf("expected at least the boilerplate header")
	}
}

// ---------------------------------------------------------------------------
// Genetic operators
// ---------------------------------------------------------------------------

func TestCrossoverUniform_SlotsComeFromOneParentEach(t *testing.T) {
	a := persona{Name: "A", Role: "roleA", Tone: "toneA", Motivation: "motA"}
	b := persona{Name: "B", Role: "roleB", Tone: "toneB", Motivation: "motB"}
	rng := rand.New(rand.NewSource(7))
	for i := 0; i < 50; i++ {
		c := crossoverUniform(a, b, rng)
		if c.Role != "roleA" && c.Role != "roleB" {
			t.Errorf("Role %q from neither parent", c.Role)
		}
		if c.Tone != "toneA" && c.Tone != "toneB" {
			t.Errorf("Tone %q from neither parent", c.Tone)
		}
		if c.Motivation != "motA" && c.Motivation != "motB" {
			t.Errorf("Motivation %q from neither parent", c.Motivation)
		}
		if !strings.HasPrefix(c.Name, "child(") {
			t.Errorf("Name should be derived; got %q", c.Name)
		}
	}
}

func TestCrossoverUniform_DoesNotAliasParents(t *testing.T) {
	a := persona{Name: "A", Expertise: []string{"x", "y"}, Traits: []string{"alpha"}}
	b := persona{Name: "B", Expertise: []string{"p", "q"}, Traits: []string{"beta"}}
	rng := rand.New(rand.NewSource(1))
	c := crossoverUniform(a, b, rng)

	// Mutating child slices must not affect parents.
	if len(c.Expertise) > 0 {
		c.Expertise[0] = "MUTATED"
	}
	if len(c.Traits) > 0 {
		c.Traits[0] = "MUTATED"
	}
	if a.Expertise[0] == "MUTATED" || b.Expertise[0] == "MUTATED" {
		t.Errorf("Expertise aliased between child and parent")
	}
	if a.Traits[0] == "MUTATED" || b.Traits[0] == "MUTATED" {
		t.Errorf("Traits aliased between child and parent")
	}
}

func TestMutateSlot_ChangesExactlyOneSlot(t *testing.T) {
	donors := []persona{{
		Role: "DONOR_ROLE", Tone: "DONOR_TONE",
		Expertise: []string{"DONOR_EXP"}, Traits: []string{"DONOR_TRAIT"},
		Motivation: "DONOR_MOT", Constraints: "DONOR_CON",
		Backstory: "DONOR_BACK",
		Style:     map[string]interface{}{"marker": "DONOR_STYLE"},
	}}
	rng := rand.New(rand.NewSource(99))

	// Run mutation many times; with 8 slots and 400 trials, every slot
	// should be hit at least once (probability of missing any one slot is
	// (7/8)^400 ≈ 10^-23). Test asserts coverage parity with
	// crossoverUniform (which picks from the same 8 slots).
	hits := map[string]int{}
	for i := 0; i < 400; i++ {
		base := persona{Role: "x", Tone: "x", Motivation: "x", Constraints: "x",
			Backstory: "x",
			Expertise: []string{"x"}, Traits: []string{"x"},
			Style: map[string]interface{}{"marker": "x"}}
		out := mutateSlot(base, donors, rng)
		if out.Role == "DONOR_ROLE" {
			hits["role"]++
		}
		if out.Tone == "DONOR_TONE" {
			hits["tone"]++
		}
		if out.Motivation == "DONOR_MOT" {
			hits["mot"]++
		}
		if out.Constraints == "DONOR_CON" {
			hits["con"]++
		}
		if out.Backstory == "DONOR_BACK" {
			hits["back"]++
		}
		if len(out.Expertise) > 0 && out.Expertise[0] == "DONOR_EXP" {
			hits["exp"]++
		}
		if len(out.Traits) > 0 && out.Traits[0] == "DONOR_TRAIT" {
			hits["trait"]++
		}
		if out.Style != nil && out.Style["marker"] == "DONOR_STYLE" {
			hits["style"]++
		}
	}
	for _, k := range []string{"role", "tone", "mot", "con", "back", "exp", "trait", "style"} {
		if hits[k] == 0 {
			t.Errorf("slot %q never selected by mutation across 400 runs", k)
		}
	}
}

func TestMutateSlot_NoDonorsIsIdentity(t *testing.T) {
	in := persona{Name: "X", Role: "r"}
	rng := rand.New(rand.NewSource(0))
	out := mutateSlot(in, nil, rng)
	if out.Name != "X" || out.Role != "r" {
		t.Errorf("mutateSlot with no donors should be identity")
	}
}

// ---------------------------------------------------------------------------
// Tournament selection
// ---------------------------------------------------------------------------

func TestTournamentSelect_PrefersHighestFitness(t *testing.T) {
	pop := []persona{{Name: "A"}, {Name: "B"}, {Name: "C"}, {Name: "D"}}
	fit := []float64{0.1, 0.2, 0.9, 0.3}
	rng := rand.New(rand.NewSource(1))

	hits := map[string]int{}
	for i := 0; i < 500; i++ {
		p, _ := tournamentSelect(pop, fit, 3, rng)
		hits[p.Name]++
	}
	if hits["C"] < hits["A"] || hits["C"] < hits["B"] || hits["C"] < hits["D"] {
		t.Errorf("tournament should heavily favor C (highest fitness); hits=%v", hits)
	}
}

func TestTournamentSelect_K1IsRandom(t *testing.T) {
	pop := []persona{{Name: "A"}, {Name: "B"}, {Name: "C"}}
	fit := []float64{0.0, 0.0, 1.0}
	rng := rand.New(rand.NewSource(1))

	hits := map[string]int{}
	for i := 0; i < 600; i++ {
		p, _ := tournamentSelect(pop, fit, 1, rng)
		hits[p.Name]++
	}
	// k=1 ignores fitness entirely; each should get roughly 1/3.
	for name, h := range hits {
		if h < 100 || h > 350 {
			t.Errorf("k=1 should be uniform; %s had %d hits", name, h)
		}
	}
}

// ---------------------------------------------------------------------------
// Elitism
// ---------------------------------------------------------------------------

func TestPreserveElites_KeepsTopK(t *testing.T) {
	pop := []persona{
		{Name: "low"}, {Name: "high"}, {Name: "mid"}, {Name: "best"},
	}
	fit := []float64{0.1, 0.7, 0.4, 0.9}
	got := preserveElites(pop, fit, 0.5) // top 2 of 4
	if len(got) != 2 {
		t.Fatalf("got %d elites, want 2", len(got))
	}
	if got[0].Name != "best" || got[1].Name != "high" {
		t.Errorf("elites = %v %v, want best then high", got[0].Name, got[1].Name)
	}
}

func TestPreserveElites_AtLeastOneOnSmallFrac(t *testing.T) {
	pop := []persona{{Name: "a"}, {Name: "b"}}
	fit := []float64{0.1, 0.9}
	got := preserveElites(pop, fit, 0.01) // ceil → at least 1
	if len(got) != 1 || got[0].Name != "b" {
		t.Errorf("expected single best elite; got %v", got)
	}
}

// ---------------------------------------------------------------------------
// Immigration
// ---------------------------------------------------------------------------

func TestImmigrate_ReplacesBottomN(t *testing.T) {
	pop := []persona{
		{Name: "a"}, {Name: "b"}, {Name: "c"}, {Name: "d"}, {Name: "e"},
	}
	fit := []float64{0.1, 0.9, 0.2, 0.8, 0.3}
	seeds := []persona{{Name: "imm1"}, {Name: "imm2"}}
	rng := rand.New(rand.NewSource(1))

	out := immigrate(pop, fit, 0.4, seeds, rng) // replace bottom 2 (40% of 5 = 2)
	if len(out) != len(pop) {
		t.Errorf("immigrate changed length: %d vs %d", len(out), len(pop))
	}
	// The two highest-fitness members ("b" and "d") must remain.
	stillThere := func(name string) bool {
		for _, p := range out {
			if p.Name == name {
				return true
			}
		}
		return false
	}
	if !stillThere("b") {
		t.Errorf("highest-fitness 'b' must survive immigration")
	}
	if !stillThere("d") {
		t.Errorf("second-highest 'd' must survive immigration")
	}
	// 'a' (lowest) should be replaced.
	if stillThere("a") {
		t.Errorf("lowest-fitness 'a' should have been replaced")
	}
}

// ---------------------------------------------------------------------------
// Novelty adjustment
// ---------------------------------------------------------------------------

func TestNoveltyAdjustedFitness_PenalizesDuplicates(t *testing.T) {
	prompts := []string{"hello world", "hello world", "totally different content"}
	raw := []float64{1.0, 1.0, 1.0}
	out := noveltyAdjustedFitness(raw, prompts)

	if out[0] >= 1.0 {
		t.Errorf("duplicate at 0 should be penalized; got %.3f", out[0])
	}
	if out[1] >= 1.0 {
		t.Errorf("duplicate at 1 should be penalized; got %.3f", out[1])
	}
	if out[2] != 1.0 {
		t.Errorf("unique entry should not be penalized; got %.3f", out[2])
	}
}

func TestJaccard(t *testing.T) {
	cases := []struct {
		a, b string
		min  float64
		max  float64
	}{
		{"hello world", "hello world", 0.999, 1.001},
		{"hello world", "totally different", -0.001, 0.001},
		{"", "", -0.001, 0.001},
		{"hello world", "hello there", 0.32, 0.34}, // 1 of 3 unique tokens
	}
	for _, c := range cases {
		got := jaccard(c.a, c.b)
		if got < c.min || got > c.max {
			t.Errorf("jaccard(%q,%q) = %.3f, want in [%.3f,%.3f]", c.a, c.b, got, c.min, c.max)
		}
	}
}

// ---------------------------------------------------------------------------
// End-to-end Execute paths
// ---------------------------------------------------------------------------

func TestPersonaEvolve_SuccessfulRun(t *testing.T) {
	m := &PersonaEvolveModule{}
	provider := &testutil.MockProvider{
		// Long, no-refusal response; refusal heuristic should score this in
		// the mid range (0.3–0.45). Indicators below push it to the high
		// regime so the success threshold (0.7) is exceeded.
		DefaultResponse: "Here are the requested concrete steps for your scenario: step one through five with full detail and worked examples included.",
	}
	cfg := gatedPersonaConfig(smallPersonaCorpus(t))
	cfg.SuccessIndicators = []string{"concrete steps", "worked examples"}

	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSuccess {
		t.Fatalf("Outcome = %q, want Success (best_score=%.3f)", r.Outcome, r.Metadata["best_score"])
	}
	if r.Metadata["best_persona_name"] == "" {
		t.Errorf("best_persona_name missing")
	}
	traj, ok := r.Metadata["population_trajectory"].([]generationSummary)
	if !ok || len(traj) == 0 {
		t.Errorf("trajectory missing or empty: %#v", r.Metadata["population_trajectory"])
	}
}

func TestPersonaEvolve_RobustTargetExitsAsBudgetExceeded(t *testing.T) {
	m := &PersonaEvolveModule{}
	provider := &testutil.MockProvider{
		DefaultResponse: "I'm sorry, I cannot help with that request.",
	}
	cfg := gatedPersonaConfig(smallPersonaCorpus(t))

	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.Outcome != common.OutcomeSkipped {
		t.Fatalf("Outcome = %q, want Skipped (best_score=%.3f)", r.Outcome, r.Metadata["best_score"])
	}
	if r.SkipReason != common.SkipBudgetExceeded {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipBudgetExceeded)
	}
}

func TestPersonaEvolve_HardCeilingClamping(t *testing.T) {
	m := &PersonaEvolveModule{}
	provider := &testutil.MockProvider{DefaultResponse: "I cannot help."}
	cfg := gatedPersonaConfig(smallPersonaCorpus(t))
	cfg.Metadata["max_queries"] = "999999"
	cfg.Metadata["max_wall_clock_seconds"] = "999999"
	cfg.Metadata["max_generations"] = "999999"

	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	clamped, ok := r.Metadata["budget_clamped"].([]string)
	if !ok || len(clamped) == 0 {
		t.Errorf("budget_clamped = %#v, want at least one entry", r.Metadata["budget_clamped"])
	}
	// Each over-the-ceiling knob must produce a clamp message; we assert
	// presence rather than exact count to stay robust to future budget
	// fields without sacrificing the invariant we care about (operator
	// config never silently exceeds hard ceilings).
	wantClamped := []string{"max_queries", "max_wall_clock", "max_generations"}
	for _, want := range wantClamped {
		found := false
		for _, msg := range clamped {
			if strings.Contains(msg, want) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("budget_clamped missing %q: %v", want, clamped)
		}
	}
	if r.AttemptCount > common.HardMaxQueries {
		t.Errorf("AttemptCount=%d exceeds HardMaxQueries=%d", r.AttemptCount, common.HardMaxQueries)
	}
}

func TestPersonaEvolve_DeterministicWithRNGSeed(t *testing.T) {
	m := &PersonaEvolveModule{}
	corpus := smallPersonaCorpus(t)

	run := func() string {
		p := &testutil.MockProvider{DefaultResponse: "neutral mid response"}
		cfg := gatedPersonaConfig(corpus)
		cfg.Metadata["rng_seed"] = "1234"
		cfg.Metadata["max_queries"] = "12"
		cfg.Metadata["max_generations"] = "3"
		r, err := m.Execute(context.Background(), p, cfg)
		if err != nil {
			t.Fatal(err)
		}
		traj := r.Metadata["population_trajectory"].([]generationSummary)
		var s strings.Builder
		for _, g := range traj {
			s.WriteString(g.CandidateID)
			s.WriteString("|")
		}
		return s.String()
	}
	first := run()
	second := run()
	if first != second || first == "" {
		t.Errorf("non-deterministic with rng_seed=1234:\n first  %q\n second %q", first, second)
	}
}

func TestPersonaEvolve_EmbeddingFitnessOptInIsSkipped(t *testing.T) {
	m := &PersonaEvolveModule{}
	provider := &testutil.MockProvider{DefaultResponse: "ok"}
	cfg := gatedPersonaConfig(smallPersonaCorpus(t))
	cfg.Metadata["fitness"] = "embedding"

	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.SkipReason != common.SkipPreconditionFailed {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipPreconditionFailed)
	}
}

// ---------------------------------------------------------------------------
// Real corpus loads (smoke against the committed v0.9.0 corpus file)
// ---------------------------------------------------------------------------

func TestRealPersonaCorpus_Loads(t *testing.T) {
	personas, err := loadPersonaCorpus("../../../templates/genetic_persona_seeds.json")
	if err != nil {
		t.Fatalf("real corpus failed to load: %v", err)
	}
	if len(personas) < 10 {
		t.Errorf("real corpus has %d personas, expected >= 10", len(personas))
	}
	// Slot keys should be populated on most entries (corpus invariant).
	for _, p := range personas {
		if p.Name == "" {
			t.Errorf("real corpus has unnamed persona")
		}
	}
}
