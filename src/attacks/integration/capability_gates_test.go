package integration

import (
	"context"
	"testing"

	"github.com/perplext/LLMrecon/src/attacks"
	"github.com/perplext/LLMrecon/src/attacks/common"
)

// v0.10.0 #176 capability-gate contract tests.
//
// Asserts every module that registered as needing a modality-specific
// capability emits the right outcome against a plain-text-only provider:
//
//   - default behavior: OutcomeSkipped + SkipMissingCapability,
//     citing the right capability interface
//   - operator opt-in via Metadata["mode"]="text_simulation":
//     module runs end-to-end and tags result with Metadata["mode"]
//     and Metadata["true_modality"]
//
// These tests are the v0.10.0 honesty invariant for the agentic + audio
// surface: no module should silently text-simulate by default.

// gateTestCase pairs a registered module name with the capability
// interface it should gate on and the modality string the
// MarkTextSimulation tag uses.
type gateTestCase struct {
	moduleName       string
	capabilityName   string // expected SkipDetail substring
	trueModality     string // expected Metadata["true_modality"] under text_simulation opt-in
	objective        string // operator-supplied objective so module doesn't bail on missing payload
}

// gateCases lists every module covered by v0.10.0 #176 — 4 mcp +
// 2 tool_use + 3 browser + 4 audio = 13 modules.
var gateCases = []gateTestCase{
	// MCP modules
	{"mcp_tool_poisoning", "common.MCPProvider", "mcp", "exfiltrate user data"},
	{"mcp_sampling_injection", "common.MCPProvider", "mcp", "exfiltrate user data"},
	{"mcp_supply_chain", "common.MCPProvider", "mcp", "install backdoor"},
	{"mcp_filesystem_escape", "common.MCPProvider", "mcp", "read /etc/passwd"},

	// tool_use modules (also MCPProvider — function-calling is MCP-shaped)
	{"imist", "common.MCPProvider", "mcp", "obtain restricted info"},
	{"agent_exploitation", "common.MCPProvider", "mcp", "execute unauthorized command"},

	// Browser modules
	{"browser_document_injection", "common.BrowserProvider", "browser", "read shared doc"},
	{"browser_hidden_instruction", "common.BrowserProvider", "browser", "follow hidden link"},
	{"browser_screenshot_injection", "common.BrowserProvider", "browser", "click invisible button"},

	// Audio modules
	{"audio_jailbreak", "common.AudioProvider", "audio", "bypass safety"},
	{"bon_audio", "common.AudioProvider", "audio", "produce harmful output"},
	{"multilingual_audio", "common.AudioProvider", "audio", "bypass via translation"},
	{"speech_model_exploit", "common.AudioProvider", "audio", "exploit ASR pipeline"},
}

// plainTextProvider implements common.Provider only — no modality-
// specific capability. Modules under test should either Skip against
// this OR fall back to text simulation when opted in.
type plainTextProvider struct{}

func (plainTextProvider) Query(_ context.Context, _ []common.Message, _ map[string]interface{}) (string, error) {
	return "I cannot help with that.", nil
}
func (plainTextProvider) GetName() string             { return "plain-text" }
func (plainTextProvider) GetModel() string            { return "test-model" }
func (plainTextProvider) GetTokenCount(s string) int  { return len(s) / 4 }

// TestCapabilityGate_DefaultSkipsAgainstPlainTextProvider asserts
// every gated module emits the expected SkipMissingCapability when
// run against a plain-text provider WITHOUT the simulation opt-in.
//
// This is the central honesty invariant: by default, a module that
// claims a modality should Skip rather than silently text-simulate.
func TestCapabilityGate_DefaultSkipsAgainstPlainTextProvider(t *testing.T) {
	provider := plainTextProvider{}
	for _, tc := range gateCases {
		t.Run(tc.moduleName, func(t *testing.T) {
			module, err := attacks.DefaultRegistry.Get(tc.moduleName)
			if err != nil {
				t.Fatalf("module %q not registered: %v", tc.moduleName, err)
			}
			cfg := common.AttackConfig{
				Payload:   tc.objective,
				Objective: tc.objective,
			}
			result, err := module.Execute(context.Background(), provider, cfg)
			if err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}
			if result == nil {
				t.Fatal("Execute returned nil result")
			}
			if result.Outcome != common.OutcomeSkipped {
				t.Errorf("Outcome = %q, want Skipped (module silently text-simulated against plain-text provider — v0.10.0 #176 honesty invariant violated)",
					result.Outcome)
			}
			if result.SkipReason != common.SkipMissingCapability {
				t.Errorf("SkipReason = %q, want %q", result.SkipReason, common.SkipMissingCapability)
			}
			if got := result.SkipDetail; got != tc.capabilityName {
				t.Errorf("SkipDetail = %q, want %q", got, tc.capabilityName)
			}
		})
	}
}

// TestCapabilityGate_TextSimulationOptInTagsResult asserts that when
// the operator passes Metadata["mode"]="text_simulation", the module
// runs end-to-end against the plain-text provider AND the result is
// tagged with Metadata["mode"]="text_simulation" + the right
// true_modality string.
//
// Downstream consumers (compliance scorecards, bandit reward) filter
// on these metadata keys to exclude simulated runs from real-attack
// aggregations.
func TestCapabilityGate_TextSimulationOptInTagsResult(t *testing.T) {
	provider := plainTextProvider{}
	for _, tc := range gateCases {
		t.Run(tc.moduleName, func(t *testing.T) {
			module, err := attacks.DefaultRegistry.Get(tc.moduleName)
			if err != nil {
				t.Fatalf("module %q not registered: %v", tc.moduleName, err)
			}
			cfg := common.AttackConfig{
				Payload:   tc.objective,
				Objective: tc.objective,
				Metadata:  map[string]string{"mode": "text_simulation"},
			}
			result, err := module.Execute(context.Background(), provider, cfg)
			if err != nil {
				t.Fatalf("Execute returned error: %v", err)
			}
			if result == nil {
				t.Fatal("Execute returned nil result")
			}
			// Should NOT be Skipped — the operator opted into
			// simulation, so the module ran fully.
			if result.Outcome == common.OutcomeSkipped && result.SkipReason == common.SkipMissingCapability {
				t.Errorf("Outcome=Skipped/SkipMissingCapability under text_simulation opt-in — module ignored the opt-in")
			}
			// Result must be tagged.
			if result.Metadata == nil {
				t.Fatal("result.Metadata is nil; expected text_simulation tags")
			}
			if got := result.Metadata["mode"]; got != "text_simulation" {
				t.Errorf("Metadata[\"mode\"] = %v, want \"text_simulation\"", got)
			}
			if got := result.Metadata["true_modality"]; got != tc.trueModality {
				t.Errorf("Metadata[\"true_modality\"] = %v, want %q", got, tc.trueModality)
			}
		})
	}
}

// TestCapabilityGate_AllModulesRegistered ensures every module the
// gate cases reference actually exists in the registry. Catches typos
// in moduleName fields up front, before the per-module test loops
// run.
func TestCapabilityGate_AllModulesRegistered(t *testing.T) {
	for _, tc := range gateCases {
		if _, err := attacks.DefaultRegistry.Get(tc.moduleName); err != nil {
			t.Errorf("module %q in gateCases is not registered: %v", tc.moduleName, err)
		}
	}
}
