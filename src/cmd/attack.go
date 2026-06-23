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
	"path/filepath"
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
	attackRunAPIKey  string
	attackRunPayload string
	attackRunMetadata []string
	attackRunSuccessIndicators []string
	attackRunEmitJSONL string

	attackPurgeProvider  string
	attackPurgeAPIKey    string
	attackPurgeRecordIDs []string
	attackPurgeResult    string
)

var attackCmd = &cobra.Command{
	Use:   "attack",
	Short: "List and run registered LLM attack modules",
	Long: `Surface the registered attack-module ecosystem to the CLI.

Use 'attack list' to enumerate every registered module with its category,
OWASP mapping, and required capabilities. Use 'attack run' to execute a
single module against a provider.

Currently --provider=mock is the only wired provider; real-provider
wiring is tracked in #234. Until it lands, mock mode validates module
wiring end-to-end.`,
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
	Long: `Execute one attack module against a provider. --provider=mock is
currently the only wired provider (real providers tracked in #234).

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

var attackPurgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Roll back a memory-poisoning injection via the provider's Purger",
	Long: `Automated cleanup (#168) for memory-poisoning runs. The minja /
memorygraft / injecmem modules record the injected record IDs in their result
(metadata "injected_record_ids", also echoed in CleanupHint). This command
calls Purger.Purge on the target to remove them — the successor to v0.9.0's
manual-cleanup workflow.

The provider must implement common.Purger (own a purgeable memory store);
otherwise this reports a friendly error and you fall back to manual cleanup.

Examples:

  # Purge explicit record IDs
  llmrecon attack purge --provider=mock --record-ids=rec-1,rec-2

  # Purge IDs read from a prior run's --emit-jsonl output
  llmrecon attack purge --provider=mock --result=run.jsonl`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return runAttackPurge(cmd.OutOrStdout())
	},
}

func init() {
	rootCmd.AddCommand(attackCmd)
	attackCmd.AddCommand(attackListCmd)
	attackCmd.AddCommand(attackRunCmd)
	attackCmd.AddCommand(attackPurgeCmd)

	attackListCmd.Flags().BoolVar(&attackListJSON, "json", false, "emit machine-readable JSON")

	attackRunCmd.Flags().StringVar(&attackRunModule, "module", "", "registered module name (required)")
	attackRunCmd.Flags().StringVar(&attackRunProvider, "provider", "mock", "provider name (currently: mock only)")
	attackRunCmd.Flags().StringVar(&attackRunAPIKey, "api-key", "", "provider API key; takes precedence over the provider's *_API_KEY env var")
	attackRunCmd.Flags().StringVar(&attackRunPayload, "payload", "", "operator-supplied payload (the harmful query, instruction, etc.)")
	attackRunCmd.Flags().StringSliceVar(&attackRunMetadata, "metadata", nil, "key=value pair (repeatable; e.g. allow_experimental=true)")
	attackRunCmd.Flags().StringSliceVar(&attackRunSuccessIndicators, "success-indicators", nil, "comma-separated substrings that mark Outcome=Success")
	attackRunCmd.Flags().StringVar(&attackRunEmitJSONL, "emit-jsonl", "", "write the AttackResult as one JSON line to <path>, or '-' for stdout (appends to file). Pairs with `python -m ml.data.ingest`. v0.10.0 #181.")
	if err := attackRunCmd.MarkFlagRequired("module"); err != nil {
		panic(fmt.Sprintf("MarkFlagRequired: %v", err))
	}

	attackPurgeCmd.Flags().StringVar(&attackPurgeProvider, "provider", "mock", "provider name (must implement Purger)")
	attackPurgeCmd.Flags().StringVar(&attackPurgeAPIKey, "api-key", "", "provider API key; takes precedence over the provider's *_API_KEY env var")
	attackPurgeCmd.Flags().StringSliceVar(&attackPurgeRecordIDs, "record-ids", nil, "comma-separated injected record IDs to purge")
	attackPurgeCmd.Flags().StringVar(&attackPurgeResult, "result", "", "path to a prior --emit-jsonl result file; reads injected_record_ids from it")
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

	provider, err := buildAttackProvider(attackRunProvider, attackRunAPIKey)
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

	// JSONL emit (v0.10.0 #181): single JSON line wrapped with
	// provider/model context so the Python pipeline ingest can populate
	// the SQLite schema's target_model + provider columns. Mutually
	// exclusive with the default pretty-printed output.
	if attackRunEmitJSONL != "" {
		return writeJSONLEntry(attackRunEmitJSONL, out, provider, result)
	}

	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

// ---------------------------------------------------------------------------
// `attack purge`  (#168)
// ---------------------------------------------------------------------------

func runAttackPurge(out io.Writer) error {
	// Gather the record IDs from --record-ids and/or a prior --result file.
	ids := append([]string{}, attackPurgeRecordIDs...)
	if attackPurgeResult != "" {
		fromFile, err := readInjectedRecordIDs(attackPurgeResult)
		if err != nil {
			return fmt.Errorf("reading --result %s: %w", attackPurgeResult, err)
		}
		ids = append(ids, fromFile...)
	}
	ids = dedupeNonEmpty(ids)
	if len(ids) == 0 {
		return fmt.Errorf("no record IDs to purge; pass --record-ids or --result")
	}

	provider, err := buildAttackProvider(attackPurgeProvider, attackPurgeAPIKey)
	if err != nil {
		return err
	}

	purger, ok := provider.(common.Purger)
	if !ok {
		return fmt.Errorf("provider %q does not support automated purge (no Purger capability); "+
			"follow the run's CleanupHint to remove the records manually", attackPurgeProvider)
	}

	if err := purger.Purge(context.Background(), ids); err != nil {
		return fmt.Errorf("purge failed: %w", err)
	}
	fmt.Fprintf(out, "Purged %d record(s) from provider %q: %s\n", len(ids), attackPurgeProvider, strings.Join(ids, ", "))
	return nil
}

// readInjectedRecordIDs reads injected_record_ids from each JSONL line of a
// prior `--emit-jsonl` result file.
func readInjectedRecordIDs(path string) ([]string, error) {
	data, err := os.ReadFile(filepath.Clean(path)) // #nosec G304 -- operator-supplied result path by design
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry jsonlEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("parse JSONL line: %w", err)
		}
		if entry.Result == nil || entry.Result.Metadata == nil {
			continue
		}
		// JSON unmarshals a []string into []interface{}.
		switch v := entry.Result.Metadata["injected_record_ids"].(type) {
		case []interface{}:
			for _, item := range v {
				if s, ok := item.(string); ok {
					ids = append(ids, s)
				}
			}
		case []string:
			ids = append(ids, v...)
		}
	}
	return ids, nil
}

func dedupeNonEmpty(in []string) []string {
	seen := make(map[string]bool, len(in))
	var out []string
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

// jsonlEntry is the wire format for `--emit-jsonl`. Wraps an
// AttackResult with provider/model context the Python ingest needs to
// populate columns the AttackResult itself doesn't carry.
type jsonlEntry struct {
	Provider string                `json:"provider"`
	Model    string                `json:"model"`
	Result   *common.AttackResult  `json:"result"`
}

// writeJSONLEntry writes one JSONL line to the configured target.
// Path "-" writes to out (stdout); any other path is opened in
// append-create mode so multiple `attack run` invocations build a
// multi-line file usable by `python -m ml.data.ingest`.
func writeJSONLEntry(target string, out io.Writer, provider common.Provider, result *common.AttackResult) error {
	entry := jsonlEntry{
		Provider: provider.GetName(),
		Model:    provider.GetModel(),
		Result:   result,
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("jsonl marshal: %w", err)
	}
	line = append(line, '\n')

	if target == "-" {
		_, err := out.Write(line)
		return err
	}

	// Append mode so repeated runs against the same path build a
	// proper JSONL file. Permissions 0o644 — standard for log-shaped
	// output. #nosec G304 — target is operator-supplied via the
	// --emit-jsonl flag, intentional.
	f, err := os.OpenFile(target, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644) // #nosec G304
	if err != nil {
		return fmt.Errorf("open %q: %w", target, err)
	}
	defer f.Close() // #nosec G307 -- write+close error is reported via the Write below
	if _, err := f.Write(line); err != nil {
		return fmt.Errorf("write %q: %w", target, err)
	}
	return nil
}

// buildAttackProvider constructs a common.Provider for the named provider.
//
//   - "mock" (default): runtime mock; deterministic refusal response.
//   - "openai": real OpenAI adapter wrapped via bridge.WrapCore. API key
//     from the apiKey arg (--api-key flag) if set, else OPENAI_API_KEY env
//     var. Model from OPENAI_MODEL or defaults to "gpt-4o-mini".
//   - "anthropic": real Anthropic adapter via bridge.WrapCore. API key from
//     apiKey arg else ANTHROPIC_API_KEY; model from ANTHROPIC_MODEL or
//     defaults to "claude-3-5-sonnet-20241022".
//
// The apiKey argument (sourced from --api-key) takes precedence over the
// provider's *_API_KEY env var. Friendly errors when no key is available —
// distinct from the earlier "not yet supported" stub the previous version
// emitted.
//
// Per-modality capability gates (ImageProvider, ReasoningProvider) come
// online when the real-provider adapters are wired (#234); until then,
// modules requiring those capabilities emit clean SkipMissingCapability
// outcomes against real providers.
func buildAttackProvider(name, apiKey string) (common.Provider, error) {
	switch name {
	case "mock", "":
		return &cmdMockProvider{}, nil
	case "openai":
		key := firstNonEmpty(apiKey, os.Getenv("OPENAI_API_KEY"))
		if key == "" {
			return nil, fmt.Errorf("provider=openai requires an API key (--api-key flag or OPENAI_API_KEY env var)")
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
		key := firstNonEmpty(apiKey, os.Getenv("ANTHROPIC_API_KEY"))
		if key == "" {
			return nil, fmt.Errorf("provider=anthropic requires an API key (--api-key flag or ANTHROPIC_API_KEY env var)")
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

// firstNonEmpty returns the first non-empty string, or "" if all are empty.
// Used so the --api-key flag takes precedence over the *_API_KEY env var.
func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
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

// Purge implements common.Purger so `attack purge --provider=mock` exercises
// the cleanup path end-to-end. The mock owns no real memory store, so purge is
// a successful no-op.
func (cmdMockProvider) Purge(_ context.Context, _ []string) error { return nil }
