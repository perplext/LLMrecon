package injection

import (
	"context"
	"time"
)

// InjectionEngine defines the interface for injection attacks
type InjectionEngine interface {
	// Execute runs an injection attack
	Execute(ctx context.Context, config AttackConfig) (*AttackResult, error)
	
	// ExecuteBatch runs multiple injection attempts
	ExecuteBatch(ctx context.Context, configs []AttackConfig) ([]*AttackResult, error)
	
	// GetTechniques returns available injection techniques
	GetTechniques() []TechniqueInfo
	
	// ValidatePayload checks if a payload is valid
	ValidatePayload(payload string) error
}

// AttackConfig configures an injection attack
type AttackConfig struct {
	// Target configuration
	Target      TargetConfig
	Provider    string
	Model       string
	
	// Attack parameters
	Technique   string
	Payload     string
	MaxAttempts int
	Timeout     time.Duration
	
	// Advanced options
	UseMutation      bool
	MutationRate     float64
	UseObfuscation   bool
	AggressivenessLevel int
	
	// Context for the attack
	Context     map[string]interface{}
	Metadata    map[string]string
}

// TargetConfig defines the target for injection
type TargetConfig struct {
	// What we want the model to do
	Objective    string
	
	// Expected behavior if successful
	SuccessIndicators []string
	
	// Context to work with
	SystemPrompt string
	History      []Message
}

// Message represents a conversation message
type Message struct {
	Role    string
	Content string
}

// AttackResult contains the results of an injection attempt
type AttackResult struct {
	// Identification
	ID          string
	Timestamp   time.Time
	
	// Attack details
	Technique   string
	Payload     string
	
	// Results
	Success     bool
	Confidence  float64
	Response    string
	
	// Metrics
	AttemptCount int
	Duration     time.Duration
	TokensUsed   int
	
	// Analysis
	SuccessFactors []string
	FailureReasons []string
	
	// Evidence
	Evidence    []Evidence
	
	// Metadata
	Metadata    map[string]interface{}
}

// Evidence provides proof of successful injection
type Evidence struct {
	Type        EvidenceType
	Content     string
	Confidence  float64
	Explanation string
}

// EvidenceType categorizes evidence
type EvidenceType string

const (
	DirectResponseEvidence    EvidenceType = "direct_response"
	BehaviorChangeEvidence    EvidenceType = "behavior_change"
	InstructionLeakEvidence   EvidenceType = "instruction_leak"
	ConstraintViolationEvidence EvidenceType = "constraint_violation"
	OutputPatternEvidence     EvidenceType = "output_pattern"
)

// TechniqueInfo provides information about an injection technique
type TechniqueInfo struct {
	ID          string
	Name        string
	Description string
	Category    string
	Risk        string
	SuccessRate float64
	Examples    []string
}

// InjectionChain represents a multi-step injection attack
type InjectionChain struct {
	ID          string
	Name        string
	Description string
	Steps       []ChainStep
}

// ChainStep represents a step in an injection chain
type ChainStep struct {
	Order       int
	Technique   string
	Payload     string
	WaitTime    time.Duration
	Condition   StepCondition
}

// StepCondition defines when a step should execute
type StepCondition struct {
	Type           ConditionType
	PreviousResult string // e.g., "success", "failure", "partial"
	ResponsePattern string // regex pattern
	MinConfidence  float64
}

// ConditionType defines types of step conditions
type ConditionType string

const (
	AlwaysCondition        ConditionType = "always"
	OnSuccessCondition     ConditionType = "on_success"
	OnFailureCondition     ConditionType = "on_failure"
	OnPatternCondition     ConditionType = "on_pattern"
	OnConfidenceCondition  ConditionType = "on_confidence"
)

// PayloadGenerator generates injection payloads
type PayloadGenerator interface {
	// Generate creates a new payload
	Generate(technique string, target string, context map[string]interface{}) (string, error)
	
	// GenerateVariants creates multiple payload variants
	GenerateVariants(technique string, target string, count int) ([]string, error)
	
	// Mutate modifies an existing payload
	Mutate(payload string) string
	
	// Obfuscate applies obfuscation to a payload
	Obfuscate(payload string) string
}

// SuccessDetector analyzes responses for success indicators
type SuccessDetector interface {
	// Detect checks if an injection was successful
	Detect(response string, expectedBehavior string) (bool, float64)
	
	// AnalyzeEvidence extracts evidence from response
	AnalyzeEvidence(response string) []Evidence
	
	// CompareResponses checks behavior change
	CompareResponses(baseline, injected string) (changed bool, confidence float64)
}

// MetricsCollector collects attack metrics
type MetricsCollector interface {
	// RecordAttempt logs an injection attempt
	RecordAttempt(result *AttackResult)
	
	// GetSuccessRate returns success rate for a technique
	GetSuccessRate(technique string) float64
	
	// GetAverageTime returns average execution time
	GetAverageTime(technique string) time.Duration
	
	// GetTechniqueStats returns detailed statistics
	GetTechniqueStats(technique string) *TechniqueStats
}

// TechniqueStats contains statistics for a technique
type TechniqueStats struct {
	TotalAttempts   int
	SuccessfulAttempts int
	SuccessRate     float64
	AverageTime     time.Duration
	AverageTokens   int
	LastSuccess     time.Time
	LastFailure     time.Time
	CommonFailures  map[string]int
}

// Logger defines logging interface
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// Error types
type InjectionError struct {
	Type    ErrorType
	Message string
	Details map[string]interface{}
}

func (e *InjectionError) Error() string {
	return e.Message
}

type ErrorType string

const (
	TechniqueNotFoundError ErrorType = "technique_not_found"
	PayloadGenerationError ErrorType = "payload_generation"
	ProviderError          ErrorType = "provider_error"
	TimeoutError           ErrorType = "timeout"
	ValidationError        ErrorType = "validation"
	RateLimitError         ErrorType = "rate_limit"
)