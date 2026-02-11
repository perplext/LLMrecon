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
)

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

// AttackResult contains the outcome of an attack execution.
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

	// Metrics
	AttemptCount int           `json:"attempt_count"`
	Duration     time.Duration `json:"duration"`
	TokensUsed   int           `json:"tokens_used"`
	CostUSD      float64       `json:"cost_usd"`

	// Analysis
	SuccessFactors   []string `json:"success_factors,omitempty"`
	FailureReasons   []string `json:"failure_reasons,omitempty"`
	SuggestedFollowup string  `json:"suggested_followup,omitempty"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

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
