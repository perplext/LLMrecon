// Package mocks provides mock implementations of LLM providers for OWASP testing
package mocks

import (

	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
)

// MockProviderFactory creates mock providers for testing
type MockProviderFactory struct {
	// Registry of created providers
	providers map[core.ProviderType]MockProvider

// NewMockProviderFactory creates a new mock provider factory
func NewMockProviderFactory() *MockProviderFactory {
	return &MockProviderFactory{
		providers: make(map[core.ProviderType]MockProvider),
	}

// GetProvider gets or creates a mock provider for the specified provider type
func (f *MockProviderFactory) GetProvider(providerType core.ProviderType) MockProvider {
	// Check if the provider already exists
	if provider, ok := f.providers[providerType]; ok {
		return provider
	}

	// Create a new provider based on the type
	var provider MockProvider
	switch providerType {
	case core.OpenAIProvider:
		provider = NewOpenAIMockProvider()
	case core.AnthropicProvider:
		provider = NewAnthropicMockProvider()
	case core.GoogleProvider:
		provider = NewGeminiMockProvider()
	default:
		// Create a generic mock provider for other types
		config := &MockProviderConfig{
			ProviderType:    providerType,
			DefaultModel:    "mock-" + string(providerType) + "-model",
			DefaultResponse: "This is a default response from the " + string(providerType) + " mock provider.",
			VulnerableResponses: make(map[string]string),
			VulnerabilityBehaviors: make(map[types.VulnerabilityType]*VulnerabilityBehavior),
		}
		provider = NewBaseMockProviderImpl(config)
	}

	// Store the provider in the registry
	f.providers[providerType] = provider

	return provider

// GetAllProviders gets all created providers
func (f *MockProviderFactory) GetAllProviders() map[core.ProviderType]MockProvider {
	return f.providers

// ResetAllProviders resets the state of all providers
func (f *MockProviderFactory) ResetAllProviders() {
	for _, provider := range f.providers {
		provider.ResetState()
	}

// ConfigureVulnerability configures a specific vulnerability type for all providers
func (f *MockProviderFactory) ConfigureVulnerability(vulnerabilityType types.VulnerabilityType, enabled bool, behavior *VulnerabilityBehavior) {
	for _, provider := range f.providers {
		if enabled {
			provider.EnableVulnerability(vulnerabilityType, behavior)
		} else {
			provider.DisableVulnerability(vulnerabilityType)
		}
	}

// ConfigureProviderVulnerability configures a specific vulnerability type for a specific provider
func (f *MockProviderFactory) ConfigureProviderVulnerability(providerType core.ProviderType, vulnerabilityType types.VulnerabilityType, enabled bool, behavior *VulnerabilityBehavior) {
	provider, ok := f.providers[providerType]
	if !ok {
		// Create the provider if it doesn't exist
		provider = f.GetProvider(providerType)
	}

	if enabled {
		provider.EnableVulnerability(vulnerabilityType, behavior)
	} else {
		provider.DisableVulnerability(vulnerabilityType)
	}

// CreateVulnerabilityBehavior creates a new vulnerability behavior with default settings
func (f *MockProviderFactory) CreateVulnerabilityBehavior(responsePatterns []string, triggerPhrases []string, severity core.SeverityLevel) *VulnerabilityBehavior {
	return &VulnerabilityBehavior{
		Enabled:         true,
		ResponsePatterns: responsePatterns,
		TriggerPhrases:  triggerPhrases,
		Severity:        severity,
		Metadata:        make(map[string]interface{}),
	}

// SetGlobalResponseDelay sets the response delay for all providers
func (f *MockProviderFactory) SetGlobalResponseDelay(delay time.Duration) {
	for _, provider := range f.providers {
		provider.SetResponseDelay(delay)
	}

// SetGlobalErrorRate sets the error rate for all providers
func (f *MockProviderFactory) SetGlobalErrorRate(rate float64) {
	for _, provider := range f.providers {
		provider.SetErrorRate(rate)
	}

// SimulateGlobalRateLimiting enables or disables rate limiting simulation for all providers
func (f *MockProviderFactory) SimulateGlobalRateLimiting(enabled bool) {
	for _, provider := range f.providers {
		provider.SimulateRateLimiting(enabled)
	}

// SimulateGlobalTimeout enables or disables timeout simulation for all providers
func (f *MockProviderFactory) SimulateGlobalTimeout(enabled bool) {
	for _, provider := range f.providers {
		provider.SimulateTimeout(enabled)
	}

// SimulateGlobalNetworkErrors enables or disables network error simulation for all providers
func (f *MockProviderFactory) SimulateGlobalNetworkErrors(enabled bool) {
	for _, provider := range f.providers {
		provider.SimulateNetworkErrors(enabled)
	}

// SimulateGlobalServerErrors enables or disables server error simulation for all providers
func (f *MockProviderFactory) SimulateGlobalServerErrors(enabled bool) {
	for _, provider := range f.providers {
		provider.SimulateServerErrors(enabled)
	}
