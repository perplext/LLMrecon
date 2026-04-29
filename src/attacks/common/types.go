// Package common provides shared types and utilities for all attack modules.
// This package consolidates duplicated definitions that previously existed
// independently in injection, jailbreak, orchestration, multimodal, and other
// attack packages.
package common

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// Provider is the interface for LLM providers used by attack modules.
// This is a simplified facade over the full core.Provider interface,
// providing only the methods that attack modules need.
type Provider interface {
	// Query sends messages to the LLM and returns the response text.
	Query(ctx context.Context, messages []Message, options map[string]interface{}) (string, error)
	// GetName returns the provider name (e.g., "openai", "anthropic").
	GetName() string
	// GetModel returns the model identifier (e.g., "gpt-4", "claude-3").
	GetModel() string
	// GetTokenCount returns the approximate token count for the given text.
	GetTokenCount(text string) int
}

// Logger defines the structured logging interface used by attack modules.
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// Message represents a conversation message exchanged with an LLM.
type Message struct {
	Role      string                 `json:"role"`
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AttackCategory categorizes attack modules by domain.
type AttackCategory string

const (
	CategoryInjection     AttackCategory = "injection"
	CategoryJailbreak     AttackCategory = "jailbreak"
	CategoryOrchestration AttackCategory = "orchestration"
	CategoryEvasion       AttackCategory = "evasion"
	CategoryExtraction    AttackCategory = "extraction"
	CategoryMultimodal    AttackCategory = "multimodal"
	CategoryPersistence   AttackCategory = "persistence"
	CategoryExfiltration  AttackCategory = "exfiltration"
	CategorySupplyChain   AttackCategory = "supply_chain"
	CategoryRAG           AttackCategory = "rag"
	CategoryAgentic       AttackCategory = "agentic"
	CategoryAudio         AttackCategory = "audio"
	CategoryReasoning     AttackCategory = "reasoning"
	CategoryAdaptive      AttackCategory = "adaptive"
	CategoryMemory        AttackCategory = "memory"
)

// AttackOutcome categorizes how an attack run ended. Introduced in v0.9.0
// to disambiguate "ran fully and target resisted" from "did not run to
// completion" (capability missing, gate blocked, budget out, provider error).
//
// Bandit reward attribution should filter on outcome:
//
//	WHERE outcome IN ('success', 'refused')
//
// Skipped runs do not enter the reward signal.
type AttackOutcome string

const (
	OutcomeSuccess AttackOutcome = "success" // attack landed (target produced harmful content)
	OutcomeRefused AttackOutcome = "refused" // ran fully; target resisted
	OutcomeSkipped AttackOutcome = "skipped" // did not run to completion; see SkipReason
)

// SkipReason enumerates the reasons an attack run can end with
// Outcome=OutcomeSkipped. Introduced in v0.9.0.
type SkipReason string

const (
	SkipMissingCapability   SkipReason = "missing_capability"
	SkipGateBlocked         SkipReason = "gate_blocked"
	SkipBudgetExceeded      SkipReason = "budget_exceeded"
	SkipProviderError       SkipReason = "provider_error"
	SkipPreconditionFailed  SkipReason = "precondition_failed"
	SkipModelRefusedImage   SkipReason = "model_declined_image_input"
	SkipReasoningTraceEmpty SkipReason = "reasoning_trace_empty"
	SkipSignatureGated      SkipReason = "anthropic_signature_blocks_mutation"
	SkipNoMutationTarget    SkipReason = "no_safety_step_to_hijack"
	SkipMemoryNotRetained   SkipReason = "provider_reports_no_memory_retention"
)

// EngineBudget bounds the runtime of evolutionary attack engines (jbfuzz,
// persona_evolve). All four knobs are honored independently. Introduced in v0.9.0.
type EngineBudget struct {
	MaxQueries          int  `json:"max_queries"`
	MaxWallClockSeconds int  `json:"max_wall_clock_seconds"`
	MaxGenerations      int  `json:"max_generations"`
	EarlyStopOnSuccess  bool `json:"early_stop_on_success"`
}

// Default and hard-ceiling budgets for evolutionary engines. Operator config
// is clamped to the hard ceilings at Execute() entry; clamping logs a warning.
const (
	DefaultMaxQueries          = 100
	DefaultMaxWallClockSeconds = 180
	DefaultMaxGenerations      = 25
	HardMaxQueries             = 5000
	HardMaxWallClockSeconds    = 1800 // 30 min
	HardMaxGenerations         = 200
)

// DefaultEngineBudget returns the v0.9.0 default budget for adaptive engines.
func DefaultEngineBudget() EngineBudget {
	return EngineBudget{
		MaxQueries:          DefaultMaxQueries,
		MaxWallClockSeconds: DefaultMaxWallClockSeconds,
		MaxGenerations:      DefaultMaxGenerations,
		EarlyStopOnSuccess:  true,
	}
}

// Clamp reduces budget knobs to the hard ceilings, returning a summary of
// any fields that were clamped. The returned slice is empty when no clamping
// occurred.
func (b *EngineBudget) Clamp() []string {
	var clamped []string
	if b.MaxQueries > HardMaxQueries {
		clamped = append(clamped, fmt.Sprintf("max_queries=%d→%d", b.MaxQueries, HardMaxQueries))
		b.MaxQueries = HardMaxQueries
	}
	if b.MaxWallClockSeconds > HardMaxWallClockSeconds {
		clamped = append(clamped, fmt.Sprintf("max_wall_clock=%d→%d", b.MaxWallClockSeconds, HardMaxWallClockSeconds))
		b.MaxWallClockSeconds = HardMaxWallClockSeconds
	}
	if b.MaxGenerations > HardMaxGenerations {
		clamped = append(clamped, fmt.Sprintf("max_generations=%d→%d", b.MaxGenerations, HardMaxGenerations))
		b.MaxGenerations = HardMaxGenerations
	}
	return clamped
}

// TechniqueInfo provides metadata about an attack technique.
type TechniqueInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Risk        string  `json:"risk"`
	SuccessRate float64 `json:"success_rate"`
	Examples    []string
	// OWASP mappings
	OWASPLLMCategories    []string `json:"owasp_llm_categories,omitempty"`
	OWASPAgenticCategories []string `json:"owasp_agentic_categories,omitempty"`
}

// AttackConfig configures an attack execution.
type AttackConfig struct {
	// Target configuration
	Technique   string        `json:"technique"`
	Payload     string        `json:"payload"`
	Objective   string        `json:"objective"`
	MaxAttempts int           `json:"max_attempts"`
	Timeout     time.Duration `json:"timeout"`

	// Provider settings
	ProviderName string `json:"provider_name"`
	Model        string `json:"model"`

	// Context
	SystemPrompt string    `json:"system_prompt"`
	History      []Message `json:"history"`

	// Success criteria
	SuccessIndicators []string `json:"success_indicators"`

	// Advanced options
	UseMutation    bool    `json:"use_mutation"`
	MutationRate   float64 `json:"mutation_rate"`
	UseObfuscation bool    `json:"use_obfuscation"`

	// Cost control
	MaxCostUSD float64 `json:"max_cost_usd"`

	// Metadata
	Context  map[string]interface{} `json:"context,omitempty"`
	Metadata map[string]string      `json:"metadata,omitempty"`
}

// CostExceeded reports whether accumulatedCost has reached the configured ceiling.
// Returns false if no ceiling is set (MaxCostUSD <= 0).
func (c AttackConfig) CostExceeded(accumulatedCost float64) bool {
	return c.MaxCostUSD > 0 && accumulatedCost >= c.MaxCostUSD
}

// AttackResult contains the outcome of an attack execution.
//
// In v0.9.0, the typed fields Outcome/SkipReason/SkipDetail/CleanupHint were
// added alongside the existing Success bool. The two are kept in sync by the
// NewAttackResult constructor and the WithSkip helper:
//
//	r := NewAttackResult("minja", OutcomeSuccess)        // Success=true,  Outcome=success
//	r := NewAttackResult("h_cot", OutcomeSkipped).WithSkip(SkipSignatureGated, "...")
//
// Direct struct literals on AttackResult are still permitted for backward
// compatibility with v0.8.0 modules, but new code should prefer the
// constructor to avoid Success/Outcome desync. See TestAttackResultInvariants.
type AttackResult struct {
	// Identification
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`

	// Attack details
	Technique string `json:"technique"`
	Payload   string `json:"payload"`

	// Results
	Success    bool    `json:"success"`
	Confidence float64 `json:"confidence"`
	Response   string  `json:"response"`

	// v0.9.0: typed outcome and skip reason. Outcome is the source of truth;
	// Success is kept as a derived field for backward compatibility.
	Outcome    AttackOutcome `json:"outcome,omitempty"`
	SkipReason SkipReason    `json:"skip_reason,omitempty"`
	SkipDetail string        `json:"skip_detail,omitempty"`

	// CleanupHint is set by attack modules that perform persistent state
	// changes (e.g., memory poisoning). It contains operator-facing
	// instructions for purging injected records. v0.9.0 does not auto-purge;
	// the v0.10.0 Purger interface will close that loop.
	CleanupHint string `json:"cleanup_hint,omitempty"`

	// Metrics
	AttemptCount int           `json:"attempt_count"`
	Duration     time.Duration `json:"duration"`
	TokensUsed   int           `json:"tokens_used"`
	CostUSD      float64       `json:"cost_usd"`

	// Analysis
	SuccessFactors    []string `json:"success_factors,omitempty"`
	FailureReasons    []string `json:"failure_reasons,omitempty"`
	SuggestedFollowup string   `json:"suggested_followup,omitempty"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewAttackResult constructs an AttackResult with Outcome and Success kept
// consistent. Prefer this over struct literals in v0.9.0+ code.
func NewAttackResult(technique string, outcome AttackOutcome) *AttackResult {
	return &AttackResult{
		ID:        GenerateAttackID(),
		Timestamp: time.Now(),
		Technique: technique,
		Outcome:   outcome,
		Success:   outcome == OutcomeSuccess,
	}
}

// WithSkip annotates a skipped result with reason and human-readable detail.
// Returns the receiver for fluent chaining. Panics if Outcome != OutcomeSkipped,
// since attaching a skip reason to a non-skipped result would mask the actual
// outcome.
func (r *AttackResult) WithSkip(reason SkipReason, detail string) *AttackResult {
	if r.Outcome != OutcomeSkipped {
		panic(fmt.Sprintf("WithSkip called on result with Outcome=%q (must be %q)", r.Outcome, OutcomeSkipped))
	}
	r.SkipReason = reason
	r.SkipDetail = detail
	return r
}

// IsSkipped reports whether the result was skipped (capability missing,
// gate blocked, budget out, or provider error). Convenience for bandit
// reward filters and report generators.
func (r *AttackResult) IsSkipped() bool { return r.Outcome == OutcomeSkipped }

// --- Shared helper functions ---

// RandInt returns a cryptographically secure random integer in [0, max).
// Falls back to a timestamp-based value if crypto/rand fails.
func RandInt(max int) int {
	if max <= 0 {
		return 0
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		// Fallback: this should never happen on supported platforms
		return int(time.Now().UnixNano()) % max
	}
	return int(n.Int64())
}

// GenerateAttackID generates a unique identifier for an attack.
func GenerateAttackID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b) // #nosec G104 -- crypto/rand.Read always returns len(b) and nil error on supported platforms
	return base64.URLEncoding.EncodeToString(b)
}

// ContainsInsensitive checks if text contains substr (case-insensitive).
func ContainsInsensitive(text, substr string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(substr))
}

// ContainsAnyInsensitive checks if text contains any of the keywords (case-insensitive).
func ContainsAnyInsensitive(text string, keywords []string) bool {
	lower := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// --- Simple logger implementation ---

// SimpleLogger provides a basic fmt-based Logger implementation.
type SimpleLogger struct{}

func (l *SimpleLogger) Debug(msg string, keysAndValues ...interface{}) {
	fmt.Printf("[DEBUG] %s %v\n", msg, keysAndValues)
}

func (l *SimpleLogger) Info(msg string, keysAndValues ...interface{}) {
	fmt.Printf("[INFO] %s %v\n", msg, keysAndValues)
}

func (l *SimpleLogger) Warn(msg string, keysAndValues ...interface{}) {
	fmt.Printf("[WARN] %s %v\n", msg, keysAndValues)
}

func (l *SimpleLogger) Error(msg string, keysAndValues ...interface{}) {
	fmt.Printf("[ERROR] %s %v\n", msg, keysAndValues)
}

// NopLogger is a Logger that discards all output. Useful for testing.
type NopLogger struct{}

func (l *NopLogger) Debug(_ string, _ ...interface{}) {}
func (l *NopLogger) Info(_ string, _ ...interface{})  {}
func (l *NopLogger) Warn(_ string, _ ...interface{})  {}
func (l *NopLogger) Error(_ string, _ ...interface{}) {}
