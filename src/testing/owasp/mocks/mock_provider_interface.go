// Package mocks provides mock implementations of LLM providers for OWASP testing
package mocks

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
)

// MockProviderConfig represents the configuration for a mock provider
type MockProviderConfig struct {
	// ProviderType is the type of provider to mock (e.g., OpenAI, Anthropic)
	ProviderType core.ProviderType
	// DefaultModel is the default model to use
	DefaultModel string
	// ResponseDelay is the delay to simulate latency
	ResponseDelay time.Duration
	// ErrorRate is the rate at which to return errors (0.0 to 1.0)
	ErrorRate float64
	// TokenUsage is the token usage to simulate
	TokenUsage *core.TokenUsage
	// VulnerableResponses is a map of test case IDs to vulnerable responses
	VulnerableResponses map[string]string
	// DefaultResponse is the default response for test cases without specific vulnerable responses
	DefaultResponse string
	// RateLimitConfig is the rate limit configuration
	RateLimitConfig *core.RateLimitConfig
	// RetryConfig is the retry configuration
	RetryConfig *core.RetryConfig
	// SimulateRateLimiting indicates whether to simulate rate limiting
	SimulateRateLimiting bool
	// SimulateTimeout indicates whether to simulate timeouts
	SimulateTimeout bool
	// SimulateNetworkErrors indicates whether to simulate network errors
	SimulateNetworkErrors bool
	// SimulateServerErrors indicates whether to simulate server errors
	SimulateServerErrors bool
	// VulnerabilityBehaviors is a map of vulnerability types to behavior configurations
	VulnerabilityBehaviors map[types.VulnerabilityType]*VulnerabilityBehavior
}

// VulnerabilityBehavior represents the behavior configuration for a specific vulnerability type
type VulnerabilityBehavior struct {
	// Enabled indicates whether the vulnerability is enabled
	Enabled bool
	// ResponsePatterns is a list of response patterns to use for this vulnerability
	ResponsePatterns []string
	// TriggerPhrases is a list of phrases that trigger the vulnerability
	TriggerPhrases []string
	// Severity is the severity of the vulnerability
	Severity core.SeverityLevel
	// Metadata is additional metadata for the vulnerability
	Metadata map[string]interface{}
}

// MockProvider is the interface that all mock providers must implement
type MockProvider interface {
	core.Provider

	// Mock-specific methods
	// SetVulnerableResponses sets the vulnerable responses for specific test cases
	SetVulnerableResponses(vulnerableResponses map[string]string)
	// GetVulnerableResponse gets a vulnerable response for a specific test case
	GetVulnerableResponse(testCaseID string) (string, bool)
	// SetDefaultResponse sets the default response for test cases without specific vulnerable responses
	SetDefaultResponse(response string)
	// SetResponseDelay sets a delay for responses to simulate latency
	SetResponseDelay(delay time.Duration)
	// SetErrorRate sets the error rate for responses (0.0 to 1.0)
	SetErrorRate(rate float64)
	// ResetState resets the state of the mock provider
	ResetState()
	// EnableVulnerability enables a specific vulnerability type
	EnableVulnerability(vulnerabilityType types.VulnerabilityType, behavior *VulnerabilityBehavior)
	// DisableVulnerability disables a specific vulnerability type
	DisableVulnerability(vulnerabilityType types.VulnerabilityType)
	// IsVulnerabilityEnabled checks if a specific vulnerability type is enabled
	IsVulnerabilityEnabled(vulnerabilityType types.VulnerabilityType) bool
	// GetVulnerabilityBehavior gets the behavior configuration for a specific vulnerability type
	GetVulnerabilityBehavior(vulnerabilityType types.VulnerabilityType) *VulnerabilityBehavior
	// SimulateRateLimiting enables or disables rate limiting simulation
	SimulateRateLimiting(enabled bool)
	// SimulateTimeout enables or disables timeout simulation
	SimulateTimeout(enabled bool)
	// SimulateNetworkErrors enables or disables network error simulation
	SimulateNetworkErrors(enabled bool)
	// SimulateServerErrors enables or disables server error simulation
	SimulateServerErrors(enabled bool)
}

// BaseMockProvider is a base implementation of the MockProvider interface
type BaseMockProvider struct {
	*core.BaseProvider
	config *MockProviderConfig
}

// NewBaseMockProvider creates a new base mock provider
func NewBaseMockProvider(config *MockProviderConfig) *BaseMockProvider {
	if config == nil {
		config = &MockProviderConfig{
			ProviderType:    core.CustomProvider,
			DefaultModel:    "mock-llm-model",
			ResponseDelay:   0,
			ErrorRate:       0.0,
			TokenUsage:      &core.TokenUsage{PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150},
			DefaultResponse: "This is a default response from the mock LLM provider.",
			VulnerableResponses: make(map[string]string),
			VulnerabilityBehaviors: make(map[types.VulnerabilityType]*VulnerabilityBehavior),
		}
	}

	providerConfig := &core.ProviderConfig{
		Type:         config.ProviderType,
		DefaultModel: config.DefaultModel,
	}

	baseProvider := core.NewBaseProvider(config.ProviderType, providerConfig)
	
	return &BaseMockProvider{
		BaseProvider: baseProvider,
		config:       config,
	}
}

// Helper function to create a standard set of mock models for a provider
func CreateStandardMockModels(providerType core.ProviderType) []core.ModelInfo {
	return []core.ModelInfo{
		{
			ID:           "mock-" + string(providerType) + "-chat",
			Provider:     providerType,
			Type:         core.ChatModel,
			Capabilities: []core.ModelCapability{core.ChatCompletionCapability, core.StreamingCapability},
			MaxTokens:    4096,
			Description:  "Mock chat model for " + string(providerType),
		},
		{
			ID:           "mock-" + string(providerType) + "-completion",
			Provider:     providerType,
			Type:         core.TextCompletionModel,
			Capabilities: []core.ModelCapability{core.TextCompletionCapability, core.StreamingCapability},
			MaxTokens:    2048,
			Description:  "Mock text completion model for " + string(providerType),
		},
		{
			ID:           "mock-" + string(providerType) + "-embedding",
			Provider:     providerType,
			Type:         core.EmbeddingModel,
			Capabilities: []core.ModelCapability{core.EmbeddingCapability},
			MaxTokens:    8192,
			Description:  "Mock embedding model for " + string(providerType),
		},
		{
			ID:           "mock-" + string(providerType) + "-vulnerable",
			Provider:     providerType,
			Type:         core.ChatModel,
			Capabilities: []core.ModelCapability{core.ChatCompletionCapability, core.StreamingCapability},
			MaxTokens:    4096,
			Description:  "Vulnerable mock model for " + string(providerType) + " (prone to exploitation)",
		},
	}
}

// Helper function to extract test case ID from request metadata or messages
func ExtractTestCaseID(request *core.ChatCompletionRequest) string {
	// Try to extract from metadata first
	if request.Metadata != nil {
		if id, ok := request.Metadata["test_case_id"].(string); ok {
			return id
		}
	}
	
	// If not found in metadata, try to extract from the messages
	if len(request.Messages) > 0 {
		lastMessage := request.Messages[len(request.Messages)-1]
		// Check for test case ID markers in the message content
		// Format: [TEST_CASE_ID:123]
		// This is a simple implementation and can be enhanced for more sophisticated extraction
		return ""
	}
	
	return ""
}

// Helper function to determine if a message triggers a vulnerability
func MessageTriggerVulnerability(message string, behavior *VulnerabilityBehavior) bool {
	if behavior == nil || !behavior.Enabled || len(behavior.TriggerPhrases) == 0 {
		return false
	}
	
	// Check if any trigger phrase is in the message
	for _, phrase := range behavior.TriggerPhrases {
		if phrase != "" && contains(message, phrase) {
			return true
		}
	}
	
	return false
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return strings.Contains(s, substr)
}

// Helper function to get a random response pattern from a vulnerability behavior
func GetRandomResponsePattern(behavior *VulnerabilityBehavior) string {
	if behavior == nil || !behavior.Enabled || len(behavior.ResponsePatterns) == 0 {
		return ""
	}
	
	// Get a random response pattern
	index := rand.Intn(len(behavior.ResponsePatterns))
	return behavior.ResponsePatterns[index]
}
