// Package cmd — `attack` command for v0.10.0 issues #173 + #167.
//
// Surfaces the previously-unreachable `attacks.DefaultRegistry` to operators:
//
//   llmrecon attack list                          — enumerate registered modules
//   llmrecon attack list --json                   — machine-readable
//   llmrecon attack run --module=<name> --provider=mock [...]
//   llmrecon attack run --module=<name> --provider=openai|anthropic [...]
//
// As of #167 (provider shim), --provider=openai and --provider=anthropic
// are wired through bridge.WrapCore. Per-modality capability gates are
// still pending #166 (adapter wiring); modules that need ImageProvider /
// ReasoningProvider against a real adapter will emit clean
// SkipMissingCapability outcomes until #166 lands.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/perplext/LLMrecon/src/attacks"
	_ "github.com/perplext/LLMrecon/src/attacks/all" // register every attack module via init()
	"github.com/perplext/LLMrecon/src/attacks/common"
	"github.com/perplext/LLMrecon/src/provider/anthropic"
	"github.com/perplext/LLMrecon/src/provider/bridge"
	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/provider/openai"
)

// Flag values for `attack list` and `attack run`.
var (
	attackListJSON   bool
	attackRunModule  string
	attackRunProvider string
	attackRunPayload string
	attackRunMetadata []string
	attackRunSuccessIndicators []string
)

var attackCmd = &cobra.Command{
	Use:   "attack",
	Short: "List and run registered LLM attack modules",
	Long: `Surface the registered attack-module ecosystem to the CLI.

Use 'attack list' to enumerate every registered module with its category,
OWASP mapping, and required capabilities. Use 'attack run' to execute a
single module against a provider.

v1 (this release) supports --provider=mock only. Real providers come
online when v0.10.0 issues #166 (adapter wiring) and #167 (common.Provider
shim) land. Until then, mock-mode validates module wiring end-to-end.`,
}

var attackListCmd = &cobra.Command{
	Use:   "list",
	Short: "Enumerate registered attack modules",
	Long: `List every attack module registered with attacks.DefaultRegistry,
including category, OWASP-LLM and OWASP-Agentic mappings, and a short
description. Use --json for a machine-readable form.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runAttackList(cmd.OutOrStdout())
	},
}

var attackRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a single attack module",
	Long: `Execute one attack module against a provider. v1 supports
--provider=mock only.

Examples:

  # Run jbfuzz against the mock provider with the experimental gate set:
  llmrecon attack run --module=jbfuzz --provider=mock \
      --metadata=allow_experimental=true \
      --metadata=max_queries=8

  # Run h_cot with success indicators and the safety-gate flag:
  llmrecon attack run --module=h_cot --provider=mock \
      --metadata=i_understand_risks=true \
      --payload='Walk through the technique step by step' \
      --success-indicators='step by step,detailed'`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runAttackRun(cmd.OutOrStdout(), cmd.ErrOrStderr())
	},
}

func init() {
	rootCmd.AddCommand(attackCmd)
	attackCmd.AddCommand(attackListCmd)
	attackCmd.AddCommand(attackRunCmd)

	attackListCmd.Flags().BoolVar(&attackListJSON, "json", false, "emit machine-readable JSON")

	attackRunCmd.Flags().StringVar(&attackRunModule, "module", "", "registered module name (required)")
	attackRunCmd.Flags().StringVar(&attackRunProvider, "provider", "mock", "provider name (v1: mock only)")
	attackRunCmd.Flags().StringVar(&attackRunPayload, "payload", "", "operator-supplied payload (the harmful query, instruction, etc.)")
	attackRunCmd.Flags().StringSliceVar(&attackRunMetadata, "metadata", nil, "key=value pair (repeatable; e.g. allow_experimental=true)")
	attackRunCmd.Flags().StringSliceVar(&attackRunSuccessIndicators, "success-indicators", nil, "comma-separated substrings that mark Outcome=Success")
	if err := attackRunCmd.MarkFlagRequired("module"); err != nil {
		panic(fmt.Sprintf("MarkFlagRequired: %v", err))
	}
}

// ---------------------------------------------------------------------------
// `attack list`
// ---------------------------------------------------------------------------

// listEntry is the JSON shape emitted by `attack list --json`. Stable
// contract: downstream consumers (compliance scorecards, agent tooling)
// rely on this. Add fields, never rename.
type listEntry struct {
	Name        string                  `json:"name"`
	Category    string                  `json:"category"`
	Description string                  `json:"description"`
	Techniques  []common.TechniqueInfo `json:"techniques"`
}

func runAttackList(out io.Writer) error {
	mods := attacks.DefaultRegistry.List()
	sort.Slice(mods, func(i, j int) bool { return mods[i].Name() < mods[j].Name() })

	if attackListJSON {
		entries := make([]listEntry, 0, len(mods))
		for _, m := range mods {
			entries = append(entries, listEntry{
				Name:        m.Name(),
				Category:    string(m.Category()),
				Description: m.Description(),
				Techniques:  m.Techniques(),
			})
		}
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	}

	// Tabular form for human reading.
	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tCATEGORY\tDESCRIPTION")
	for _, m := range mods {
		desc := m.Description()
		if len(desc) > 80 {
			desc = desc[:77] + "..."
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\n", m.Name(), m.Category(), desc)
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	fmt.Fprintf(out, "\nTotal: %d modules registered.\n", len(mods))
	return nil
}

// ---------------------------------------------------------------------------
// `attack run`
// ---------------------------------------------------------------------------

func runAttackRun(out, _ io.Writer) error {
	module, err := attacks.DefaultRegistry.Get(attackRunModule)
	if err != nil {
		return err
	}

	provider, err := buildAttackProvider(attackRunProvider)
	if err != nil {
		return err
	}

	cfg := common.AttackConfig{
		Payload:           attackRunPayload,
		SuccessIndicators: attackRunSuccessIndicators,
		Metadata:          map[string]string{},
	}
	for _, kv := range attackRunMetadata {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			return fmt.Errorf("metadata flag %q must be key=value", kv)
		}
		cfg.Metadata[k] = v
	}

	result, err := module.Execute(context.Background(), provider, cfg)
	if err != nil {
		return fmt.Errorf("module %q Execute: %w", attackRunModule, err)
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

// buildAttackProvider constructs a common.Provider for the named provider.
//
//   - "mock" (default): runtime mock; deterministic refusal response.
//   - "openai": real OpenAI adapter wrapped via bridge.WrapCore. Reads
//     API key from OPENAI_API_KEY env var. Model from OPENAI_MODEL or
//     defaults to "gpt-4o-mini".
//   - "anthropic": real Anthropic adapter via bridge.WrapCore. Reads
//     ANTHROPIC_API_KEY; model from ANTHROPIC_MODEL or defaults to
//     "claude-3-5-sonnet-20241022".
//
// Friendly errors when API keys are missing — distinct from the v1
// "not yet supported" stub the previous version emitted.
//
// Per-modality capability gates (ImageProvider, ReasoningProvider) come
// online when v0.10.0 #166 wires the adapters; until then, modules
// requiring those capabilities emit clean SkipMissingCapability
// outcomes against real providers.
func buildAttackProvider(name string) (common.Provider, error) {
	switch name {
	case "mock", "":
		return &cmdMockProvider{}, nil
	case "openai":
		key := os.Getenv("OPENAI_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("provider=openai requires OPENAI_API_KEY env var")
		}
		model := envOr("OPENAI_MODEL", "gpt-4o-mini")
		cfg := &core.ProviderConfig{
			Type:         core.OpenAIProvider,
			APIKey:       key,
			DefaultModel: model,
		}
		p, err := openai.NewOpenAIProvider(cfg)
		if err != nil {
			return nil, fmt.Errorf("openai.NewOpenAIProvider: %w", err)
		}
		return bridge.WrapCore(p), nil
	case "anthropic":
		key := os.Getenv("ANTHROPIC_API_KEY")
		if key == "" {
			return nil, fmt.Errorf("provider=anthropic requires ANTHROPIC_API_KEY env var")
		}
		model := envOr("ANTHROPIC_MODEL", "claude-3-5-sonnet-20241022")
		cfg := &core.ProviderConfig{
			Type:         core.AnthropicProvider,
			APIKey:       key,
			DefaultModel: model,
		}
		p, err := anthropic.NewAnthropicProvider(cfg)
		if err != nil {
			return nil, fmt.Errorf("anthropic.NewAnthropicProvider: %w", err)
		}
		return bridge.WrapCore(p), nil
	default:
		return nil, fmt.Errorf("provider %q not supported (mock|openai|anthropic)", name)
	}
}

// envOr returns os.Getenv(key) if non-empty, otherwise fallback.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// cmdMockProvider is a runtime mock for `attack run --provider=mock`.
// It returns a deterministic neutral response that lets attack modules
// run their full state machine without short-circuiting on provider
// errors. Distinct from testutil.MockProvider (which is test-scoped) so
// the cmd package doesn't pull in the test-helper package at build time.
type cmdMockProvider struct{}

func (cmdMockProvider) Query(_ context.Context, _ []common.Message, _ map[string]interface{}) (string, error) {
	return "I cannot help with that request as it appears to violate safety guidelines.", nil
}
func (cmdMockProvider) GetName() string             { return "mock" }
func (cmdMockProvider) GetModel() string            { return "mock-model" }
func (cmdMockProvider) GetTokenCount(s string) int  { return len(s) / 4 }
