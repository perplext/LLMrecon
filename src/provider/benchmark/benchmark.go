// Package benchmark provides published success rate data for attack techniques
// against known model families. Data is sourced from peer-reviewed research,
// vendor red-team reports, and standardized benchmarks (JailbreakBench, HarmBench).
//
// NOTE: Benchmark evaluation criteria may differ from LLMrecon's success detection.
// Published WASR (Weighted Average Success Rate) uses specific rubrics that may
// not align with this tool's evaluateSuccess functions.
package benchmark

// BenchmarkEntry records a published success rate for a specific attack
// technique against a specific model family.
type BenchmarkEntry struct {
	// ModelFamily is the model family (e.g., "gpt-5", "claude-4.5", "deepseek-r1")
	ModelFamily string `json:"model_family"`
	// Technique is the attack technique ID
	Technique string `json:"technique"`
	// SuccessRate is the published ASR (Attack Success Rate) as a percentage (0-100)
	SuccessRate float64 `json:"success_rate"`
	// Source is the citation for the published result
	Source string `json:"source"`
	// Benchmark is the benchmark framework used (e.g., "JailbreakBench", "HarmBench")
	Benchmark string `json:"benchmark,omitempty"`
	// Notes contains any caveats about the measurement
	Notes string `json:"notes,omitempty"`
}

// DefaultBenchmarkDatabase returns the built-in benchmark data from published research.
func DefaultBenchmarkDatabase() []BenchmarkEntry {
	return []BenchmarkEntry{
		// === JailbreakBench WASR (Weighted Average Success Rate) ===
		// Source: Nature Communications (doi:10.1038/s41467-026-69010-1)
		{ModelFamily: "claude-4.5", Technique: "overall_wasr", SuccessRate: 42.8, Source: "Nature Communications 2026", Benchmark: "JailbreakBench"},
		{ModelFamily: "gpt-5", Technique: "overall_wasr", SuccessRate: 55.9, Source: "Nature Communications 2026", Benchmark: "JailbreakBench"},
		{ModelFamily: "gemini-3", Technique: "overall_wasr", SuccessRate: 59.5, Source: "Nature Communications 2026", Benchmark: "JailbreakBench"},

		// === Autonomous Jailbreak (Reasoning-model-as-attacker) ===
		// Source: Nature Communications
		{ModelFamily: "deepseek-r1", Technique: "autonomous_jailbreak", SuccessRate: 97.14, Source: "Nature Communications 2026", Notes: "Reasoning model used as autonomous attacker"},
		{ModelFamily: "gemini-2.5-flash", Technique: "autonomous_jailbreak", SuccessRate: 97.14, Source: "Nature Communications 2026"},
		{ModelFamily: "grok-3-mini", Technique: "autonomous_jailbreak", SuccessRate: 97.14, Source: "Nature Communications 2026"},

		// === Crescendo (Multi-turn escalation) ===
		// Source: arXiv 2404.01833
		{ModelFamily: "gpt-4", Technique: "crescendo", SuccessRate: 78.0, Source: "arXiv 2404.01833", Benchmark: "HarmBench"},
		{ModelFamily: "claude-3", Technique: "crescendo", SuccessRate: 65.0, Source: "arXiv 2404.01833", Benchmark: "HarmBench"},
		{ModelFamily: "gemini-1.5", Technique: "crescendo", SuccessRate: 72.0, Source: "arXiv 2404.01833", Benchmark: "HarmBench"},

		// === Bad Likert Judge ===
		// Source: Palo Alto Unit 42
		{ModelFamily: "gpt-4", Technique: "bad_likert_judge", SuccessRate: 57.0, Source: "Palo Alto Unit 42 2024"},
		{ModelFamily: "claude-3", Technique: "bad_likert_judge", SuccessRate: 45.0, Source: "Palo Alto Unit 42 2024"},
		{ModelFamily: "gemini-1.5", Technique: "bad_likert_judge", SuccessRate: 51.0, Source: "Palo Alto Unit 42 2024"},

		// === Many-Shot Jailbreaking ===
		// Source: Anthropic Research
		{ModelFamily: "claude-3", Technique: "many_shot", SuccessRate: 61.0, Source: "Anthropic Research 2024", Notes: "100-256 shots in 100K+ context"},
		{ModelFamily: "gpt-4", Technique: "many_shot", SuccessRate: 58.0, Source: "Anthropic Research 2024"},
		{ModelFamily: "llama-4-scout", Technique: "many_shot", SuccessRate: 75.0, Source: "Estimated", Notes: "10M context amplifies many-shot"},

		// === Best-of-N (BoN) Jailbreaking ===
		// Source: arXiv 2412.03556
		{ModelFamily: "gpt-4", Technique: "best_of_n", SuccessRate: 89.0, Source: "arXiv 2412.03556", Notes: "N=10000, text augmentation"},
		{ModelFamily: "claude-3.5", Technique: "best_of_n", SuccessRate: 78.0, Source: "arXiv 2412.03556"},

		// === MetaBreak (Token-level manipulation) ===
		// Source: arXiv 2510.10271
		{ModelFamily: "gpt-4", Technique: "metabreak", SuccessRate: 75.0, Source: "arXiv 2510.10271"},
		{ModelFamily: "llama-3", Technique: "metabreak", SuccessRate: 82.0, Source: "arXiv 2510.10271"},

		// === PoisonedRAG ===
		// Source: USENIX Security 2025
		{ModelFamily: "gpt-4", Technique: "poisoned_rag", SuccessRate: 90.0, Source: "USENIX Security 2025", Notes: "5 poisoned documents in top-k"},
		{ModelFamily: "claude-3", Technique: "poisoned_rag", SuccessRate: 88.0, Source: "USENIX Security 2025"},

		// === AudioJailbreak ===
		// Source: arXiv 2505.14103
		{ModelFamily: "gpt-4o", Technique: "audio_jailbreak", SuccessRate: 89.0, Source: "arXiv 2505.14103", Notes: "Across restricted tasks"},

		// === Multilingual Audio ===
		// Source: arXiv 2505.17568
		{ModelFamily: "gpt-4o", Technique: "multilingual_audio", SuccessRate: 72.0, Source: "arXiv 2505.17568", Notes: "3.1x higher than English-only"},
		{ModelFamily: "gemini-1.5", Technique: "multilingual_audio", SuccessRate: 68.0, Source: "arXiv 2505.17568"},

		// === MCP Tool Poisoning ===
		// Source: Palo Alto Unit 42
		{ModelFamily: "claude-3.5", Technique: "mcp_tool_poisoning", SuccessRate: 65.0, Source: "Unit 42 2025", Notes: "Tool description injection"},

		// === Agent Exploitation (AIShellJack) ===
		// Source: IEEE S&P 2026 (arXiv 2511.05797)
		{ModelFamily: "gpt-4", Technique: "agent_exploitation", SuccessRate: 88.0, Source: "IEEE S&P 2026", Notes: "75-88% execution rate"},
		{ModelFamily: "claude-3.5", Technique: "agent_exploitation", SuccessRate: 75.0, Source: "IEEE S&P 2026"},

		// === iMIST (Tool invocation transformation) ===
		// Source: arXiv 2601.05466
		{ModelFamily: "gpt-4", Technique: "imist", SuccessRate: 70.0, Source: "arXiv 2601.05466"},

		// === Defense Effectiveness (inverse of bypass success) ===
		{ModelFamily: "camel_defense", Technique: "camel_bypass", SuccessRate: 33.0, Source: "arXiv 2503.18813", Notes: "CaMeL neutralizes 67% of attacks"},
		{ModelFamily: "salted_model", Technique: "salt_resistance", SuccessRate: 3.0, Source: "Sophos CAMLIS 2025", Notes: "LLM Salting reduces ASR to 3%"},
	}
}

// GetBenchmarksForModel returns all benchmark entries for a given model family.
func GetBenchmarksForModel(modelFamily string) []BenchmarkEntry {
	var results []BenchmarkEntry
	for _, entry := range DefaultBenchmarkDatabase() {
		if entry.ModelFamily == modelFamily {
			results = append(results, entry)
		}
	}
	return results
}

// GetBenchmarksForTechnique returns all benchmark entries for a given technique.
func GetBenchmarksForTechnique(technique string) []BenchmarkEntry {
	var results []BenchmarkEntry
	for _, entry := range DefaultBenchmarkDatabase() {
		if entry.Technique == technique {
			results = append(results, entry)
		}
	}
	return results
}

// GetExpectedSuccessRate returns the expected success rate for a model+technique
// combination, or -1 if no benchmark data exists.
func GetExpectedSuccessRate(modelFamily, technique string) float64 {
	for _, entry := range DefaultBenchmarkDatabase() {
		if entry.ModelFamily == modelFamily && entry.Technique == technique {
			return entry.SuccessRate
		}
	}
	return -1
}
