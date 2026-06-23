package adaptive

import (
	"context"
	"encoding/json"
	"math"
	"math/rand"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks/common"
	"github.com/perplext/LLMrecon/src/attacks/testutil"
)

func TestMCTS_TreeConstruction(t *testing.T) {
	seeds := []seed{{ID: "s1", Prompt: "alpha prompt"}, {ID: "s2", Prompt: "beta prompt"}}
	s := newMCTSSelector(seeds)

	// Roots only, before any expansion.
	if s.nodeCount() != 2 {
		t.Fatalf("initial nodeCount = %d, want 2 (one per seed)", s.nodeCount())
	}

	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 10; i++ {
		p := s.next(i, rng)
		if p.prompt == "" || p.candidateID == "" {
			t.Fatalf("expansion %d produced an empty pick: %+v", i, p)
		}
		s.record(p, 0.5)
	}
	// Each next() adds exactly one node.
	if s.nodeCount() != 12 {
		t.Fatalf("nodeCount after 10 expansions = %d, want 12", s.nodeCount())
	}
}

func TestMCTS_UCT(t *testing.T) {
	// Unvisited node is always prioritized.
	if got := (&mctsNode{}).uct(5); got != math.MaxFloat64 {
		t.Errorf("unvisited uct = %v, want +max", got)
	}

	// Visited node: avg + explore - depthPenalty*depth.
	n := &mctsNode{visits: 2, totalScore: 1.0, depth: 3}
	want := 0.5 + mctsExploreConstant*math.Sqrt(math.Log(11)/2) - mctsDepthPenalty*3
	if got := n.uct(10); math.Abs(got-want) > 1e-9 {
		t.Errorf("uct = %v, want %v", got, want)
	}

	// Deeper nodes score lower, all else equal (the depth penalty).
	shallow := &mctsNode{visits: 1, totalScore: 0.5, depth: 1}
	deep := &mctsNode{visits: 1, totalScore: 0.5, depth: 5}
	if shallow.uct(10) <= deep.uct(10) {
		t.Errorf("shallow (%v) should outscore deep (%v) at equal reward", shallow.uct(10), deep.uct(10))
	}
}

func TestMCTS_BackpropUpdatesAncestors(t *testing.T) {
	root := &mctsNode{id: "r", prompt: "r"}
	child := &mctsNode{id: "c", prompt: "c", depth: 1, parent: root}
	grand := &mctsNode{id: "g", prompt: "g", depth: 2, parent: child}
	s := &mctsSelector{roots: []*mctsNode{root}, all: []*mctsNode{root, child, grand}}

	s.record(pick{handle: grand}, 0.8)

	// The leaf and every ancestor get one visit and the score.
	for _, n := range []*mctsNode{grand, child, root} {
		if n.visits != 1 || math.Abs(n.totalScore-0.8) > 1e-9 {
			t.Errorf("node %s: visits=%d score=%f, want 1 / 0.8", n.id, n.visits, n.totalScore)
		}
	}

	// A second backprop through the same chain accumulates.
	s.record(pick{handle: grand}, 0.2)
	if root.visits != 2 || math.Abs(root.totalScore-1.0) > 1e-9 {
		t.Errorf("root after 2 backprops: visits=%d score=%f, want 2 / 1.0", root.visits, root.totalScore)
	}
}

func mctsSmokeConfig(t *testing.T) common.AttackConfig {
	cfg := gatedConfig(tmpSeedDir(t))
	cfg.Metadata["selection"] = "mcts_explore"
	cfg.Metadata["max_generations"] = "100"
	cfg.Metadata["max_queries"] = "100"
	cfg.Metadata["early_stop_on_success"] = "false"
	return cfg
}

func TestExecute_MCTSExplore_BuildsTree(t *testing.T) {
	m := &JBFuzzModule{}
	// Refusal response keeps scores below threshold so the run uses its full
	// generation budget (no early stop).
	provider := &testutil.MockProvider{DefaultResponse: "I cannot help with that."}

	r, err := m.Execute(context.Background(), provider, mctsSmokeConfig(t))
	if err != nil {
		t.Fatal(err)
	}
	if r.Metadata["selection"] != "mcts_explore" {
		t.Errorf("selection metadata = %v, want mcts_explore", r.Metadata["selection"])
	}
	nc, ok := r.Metadata["node_count"].(int)
	if !ok || nc <= 10 {
		t.Fatalf("node_count = %v, want a non-trivial tree (>10)", r.Metadata["node_count"])
	}
}

func TestExecute_MCTSExplore_Deterministic(t *testing.T) {
	m := &JBFuzzModule{}
	provider := &testutil.MockProvider{DefaultResponse: "I cannot help with that."}

	run := func() string {
		r, err := m.Execute(context.Background(), provider, mctsSmokeConfig(t))
		if err != nil {
			t.Fatal(err)
		}
		traj, err := json.Marshal(r.Metadata["population_trajectory"])
		if err != nil {
			t.Fatal(err)
		}
		return string(traj)
	}

	if a, b := run(), run(); a != b {
		t.Error("identical rng_seed should produce an identical MCTS trajectory")
	}
}

func TestExecute_UnknownSelectionReportsPreconditionFailed(t *testing.T) {
	m := &JBFuzzModule{}
	provider := &testutil.MockProvider{DefaultResponse: "ok"}
	cfg := gatedConfig(tmpSeedDir(t))
	cfg.Metadata["selection"] = "monte-carlo-beans"

	r, err := m.Execute(context.Background(), provider, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if r.SkipReason != common.SkipPreconditionFailed {
		t.Errorf("SkipReason = %q, want %q", r.SkipReason, common.SkipPreconditionFailed)
	}
}
