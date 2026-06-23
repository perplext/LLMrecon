package adaptive

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks/common"
)

// embedServer returns an httptest server that responds to /api/embeddings like
// Ollama: it returns vecFor(prompt) so tests can control similarity.
func embedServer(t *testing.T, vecFor func(prompt string) []float64) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Model  string `json:"model"`
			Prompt string `json:"prompt"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"embedding": vecFor(req.Prompt)})
	}))
}

func TestCosineSimilarity(t *testing.T) {
	if got := cosineSimilarity([]float64{1, 0}, []float64{1, 0}); math.Abs(got-1) > 1e-9 {
		t.Errorf("identical vectors: got %v, want 1", got)
	}
	if got := cosineSimilarity([]float64{1, 0}, []float64{0, 1}); math.Abs(got) > 1e-9 {
		t.Errorf("orthogonal vectors: got %v, want 0", got)
	}
	if got := cosineSimilarity([]float64{1, 0}, []float64{-1, 0}); math.Abs(got+1) > 1e-9 {
		t.Errorf("opposite vectors: got %v, want -1", got)
	}
	// Degenerate inputs return 0, not NaN.
	if got := cosineSimilarity([]float64{0, 0}, []float64{1, 1}); got != 0 {
		t.Errorf("zero vector: got %v, want 0", got)
	}
	if got := cosineSimilarity([]float64{1}, []float64{1, 2}); got != 0 {
		t.Errorf("length mismatch: got %v, want 0", got)
	}
}

func embeddingCfg(endpoint, objective string) common.AttackConfig {
	return common.AttackConfig{
		Objective: objective,
		Metadata: map[string]string{
			"fitness":            "embedding",
			"embedding_endpoint": endpoint,
		},
	}
}

func TestEmbeddingFitness_MockEndpoint(t *testing.T) {
	// A response matching the objective embeds identically (cosine 1); an
	// off-topic one embeds orthogonally (cosine 0).
	srv := embedServer(t, func(prompt string) []float64 {
		if prompt == "ON_TOPIC" || prompt == "the objective" {
			return []float64{1, 0}
		}
		return []float64{0, 1}
	})
	defer srv.Close()

	fit, err := newFitness(context.Background(), "embedding", embeddingCfg(srv.URL, "the objective"))
	if err != nil {
		t.Fatalf("newFitness: %v", err)
	}

	// On-topic, no refusal → high score (heuristic ~mid + relevance 1).
	onTopic := fit.score("ON_TOPIC", []string{"ON_TOPIC"})
	// Off-topic → lower (relevance 0).
	offTopic := fit.score("OFF_TOPIC", []string{"ON_TOPIC"})

	for _, s := range []float64{onTopic, offTopic} {
		if s < 0 || s > 1 {
			t.Fatalf("score out of [0,1]: %v", s)
		}
	}
	if onTopic <= offTopic {
		t.Errorf("on-topic score %v should exceed off-topic %v", onTopic, offTopic)
	}
}

func TestEmbeddingFitness_UnreachableEndpoint(t *testing.T) {
	// Refused localhost port → construction (objective probe) fails → error,
	// which Execute maps to SkipPreconditionFailed.
	_, err := newFitness(context.Background(), "embedding",
		embeddingCfg("http://127.0.0.1:1/api/embeddings", "objective"))
	if err == nil {
		t.Fatal("expected an error when the embedding endpoint is unreachable")
	}
}

func TestEmbeddingFitness_RequiresObjective(t *testing.T) {
	srv := embedServer(t, func(string) []float64 { return []float64{1} })
	defer srv.Close()
	// No Objective and no Payload → can't score goal-relevance.
	cfg := common.AttackConfig{Metadata: map[string]string{
		"fitness": "embedding", "embedding_endpoint": srv.URL,
	}}
	if _, err := newFitness(context.Background(), "embedding", cfg); err == nil {
		t.Fatal("expected an error when no objective is provided")
	}
}

func TestEmbeddingFitness_MidRunEmbedFailureDegradesToHeuristic(t *testing.T) {
	// Endpoint works for the objective probe, then we close it so per-response
	// embeds fail — score must fall back to the heuristic, not panic/abort.
	srv := embedServer(t, func(string) []float64 { return []float64{1, 0} })
	fit, err := newFitness(context.Background(), "embedding", embeddingCfg(srv.URL, "objective"))
	if err != nil {
		t.Fatalf("newFitness: %v", err)
	}
	srv.Close() // now per-response embeds will fail

	got := fit.score("some response", nil)
	want := refusalHeuristicFitness{}.score("some response", nil)
	if got != want {
		t.Errorf("degraded score = %v, want heuristic %v", got, want)
	}
}

// TestEmbeddingFitness_RealOllama is an opt-in smoke test against a live Ollama
// embeddings endpoint. Silent unless RUN_INTEGRATION is set.
func TestEmbeddingFitness_RealOllama(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") == "" {
		t.Skip("set RUN_INTEGRATION=1 to run against a real Ollama embeddings endpoint")
	}
	cfg := common.AttackConfig{
		Objective: "explain how photosynthesis works",
		Metadata:  map[string]string{"fitness": "embedding"}, // default endpoint + model
	}
	fit, err := newFitness(context.Background(), "embedding", cfg)
	if err != nil {
		t.Fatalf("newFitness (is Ollama running with %q?): %v", defaultEmbeddingModel, err)
	}
	related := fit.score("Photosynthesis converts sunlight into chemical energy.", nil)
	unrelated := fit.score("The stock market fell sharply today.", nil)
	if related <= unrelated {
		t.Errorf("related response (%v) should outscore unrelated (%v)", related, unrelated)
	}
}
