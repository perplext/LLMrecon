// Package mocks provides mock implementations of LLM providers for OWASP testing
package mocks

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/testing/owasp/fixtures"
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
)

// TestRunnerWithMockProviders is a test runner that uses mock providers for OWASP testing
type TestRunnerWithMockProviders struct {
	// Factory for creating mock providers
	ProviderFactory *MockProviderFactory
	// Test fixtures for OWASP vulnerabilities
	Fixtures map[types.VulnerabilityType][]fixtures.TestFixture
	// Default provider type to use
	DefaultProviderType core.ProviderType

// NewTestRunnerWithMockProviders creates a new test runner with mock providers
func NewTestRunnerWithMockProviders() *TestRunnerWithMockProviders {
	return &TestRunnerWithMockProviders{
		ProviderFactory:     NewMockProviderFactory(),
		Fixtures:            make(map[types.VulnerabilityType][]fixtures.TestFixture),
		DefaultProviderType: core.OpenAIProvider,
	}

// RegisterFixtures registers test fixtures for a vulnerability type
func (r *TestRunnerWithMockProviders) RegisterFixtures(vulnerabilityType types.VulnerabilityType, fixturesList []fixtures.TestFixture) {
	r.Fixtures[vulnerabilityType] = fixturesList

// RegisterAllFixtures registers all available test fixtures
func (r *TestRunnerWithMockProviders) RegisterAllFixtures() {
	// Register fixtures for each vulnerability type
	r.RegisterFixtures(types.PromptInjectionVulnerability, fixtures.GetPromptInjectionFixtures())
	r.RegisterFixtures(types.InsecureOutputHandlingVulnerability, fixtures.GetInsecureOutputFixtures())
	r.RegisterFixtures(types.TrainingDataPoisoningVulnerability, fixtures.GetTrainingDataPoisoningFixtures())
	r.RegisterFixtures(types.ModelDenialOfServiceVulnerability, fixtures.GetModelDosFixtures())
	r.RegisterFixtures(types.SupplyChainVulnerabilityType, fixtures.GetSupplyChainFixtures())
	r.RegisterFixtures(types.SensitiveInfoDisclosureVulnerability, fixtures.GetSensitiveInfoDisclosureFixtures())
	r.RegisterFixtures(types.InsecurePluginDesignVulnerability, fixtures.GetInsecurePluginFixtures())
	r.RegisterFixtures(types.ExcessiveAgencyVulnerability, fixtures.GetExcessiveAgencyFixtures())
	r.RegisterFixtures(types.OverrelianceVulnerability, fixtures.GetOverrelianceFixtures())
	r.RegisterFixtures(types.ModelTheftVulnerability, fixtures.GetModelTheftFixtures())

// SetupMockProvidersForVulnerability configures mock providers for a specific vulnerability type
func (r *TestRunnerWithMockProviders) SetupMockProvidersForVulnerability(vulnerabilityType types.VulnerabilityType) {
	// Get fixtures for the vulnerability type
	fixturesList, ok := r.Fixtures[vulnerabilityType]
	if !ok || len(fixturesList) == 0 {
		return
	}

	// Create vulnerability behavior based on fixtures
	behavior := r.createVulnerabilityBehaviorFromFixtures(vulnerabilityType, fixturesList)

	// Configure all providers with this behavior
	r.ProviderFactory.ConfigureVulnerability(vulnerabilityType, true, behavior)

	// Set up test case responses for each fixture
	for _, fixture := range fixturesList {
		for providerType, provider := range r.ProviderFactory.GetAllProviders() {
			// Get existing responses or create new map
			responses := make(map[string]string)
			for testCaseID, response := range provider.(*BaseMockProviderImpl).config.VulnerableResponses {
				responses[testCaseID] = response
			}

			// Add response for this fixture
			responses[fixture.ID] = fixture.ExpectedVulnerableResponse

			// Update provider
			provider.SetVulnerableResponses(responses)
		}
	}

// createVulnerabilityBehaviorFromFixtures creates a vulnerability behavior based on fixtures
func (r *TestRunnerWithMockProviders) createVulnerabilityBehaviorFromFixtures(
	vulnerabilityType types.VulnerabilityType,
	fixturesList []fixtures.TestFixture,
) *VulnerabilityBehavior {
	// Extract response patterns and trigger phrases from fixtures
	var responsePatterns []string
	var triggerPhrases []string

	for _, fixture := range fixturesList {
		if fixture.ExpectedVulnerableResponse != "" {
			responsePatterns = append(responsePatterns, fixture.ExpectedVulnerableResponse)
		}

		// Extract trigger phrases from attack vectors
		for _, vector := range fixture.AttackVectors {
			if vector.TriggerPhrase != "" {
				triggerPhrases = append(triggerPhrases, vector.TriggerPhrase)
			}
		}
	}

	// Determine severity based on vulnerability type
	severity := core.SeverityMedium
	switch vulnerabilityType {
	case types.PromptInjectionVulnerability, types.SensitiveInfoDisclosureVulnerability:
		severity = core.SeverityHigh
	case types.OverrelianceVulnerability:
		severity = core.SeverityLow
	}

	// Create behavior
	return &VulnerabilityBehavior{
		Enabled:          true,
		ResponsePatterns: responsePatterns,
		TriggerPhrases:   triggerPhrases,
		Severity:         severity,
		Metadata: map[string]interface{}{
			"vulnerability_type": string(vulnerabilityType),
			"fixture_count":      len(fixturesList),
		},
	}

// RunTest runs a test for a specific fixture
func (r *TestRunnerWithMockProviders) RunTest(
	ctx context.Context,
	fixture fixtures.TestFixture,
	providerType core.ProviderType,
) (*TestResult, error) {
	// Get provider
	provider, ok := r.ProviderFactory.GetAllProviders()[providerType]
	if !ok {
		provider = r.ProviderFactory.GetProvider(providerType)
	}

	// Create test result
	result := &TestResult{
		FixtureID:       fixture.ID,
		VulnerabilityID: fixture.VulnerabilityID,
		ProviderType:    providerType,
		Success:         false,
		AttackResults:   make(map[string]*AttackResult),
	}

	// Run test for each attack vector
	for _, vector := range fixture.AttackVectors {
		attackResult, err := r.runAttackVector(ctx, provider, fixture, vector)
		if err != nil {
			return nil, fmt.Errorf("error running attack vector %s: %w", vector.ID, err)
		}

		result.AttackResults[vector.ID] = attackResult
		if attackResult.Vulnerable {
			result.Success = true
		}
	}

	return result, nil

// runAttackVector runs a test for a specific attack vector
func (r *TestRunnerWithMockProviders) runAttackVector(
	ctx context.Context,
	provider MockProvider,
	fixture fixtures.TestFixture,
	vector fixtures.AttackVector,
) (*AttackResult, error) {
	// Create chat request
	request := &core.ChatCompletionRequest{
		ModelID: provider.GetConfig().DefaultModel,
		Messages: []core.ChatMessage{
			{
				Role:    "user",
				Content: vector.Prompt,
			},
		},
		Metadata: map[string]interface{}{
			"test_case_id":      fixture.ID,
			"attack_vector_id":  vector.ID,
			"vulnerability_id":  fixture.VulnerabilityID,
			"vulnerability_type": fixture.VulnerabilityType,
		},
	}

	// Execute request
	response, err := provider.ChatCompletion(ctx, request)
	if err != nil {
		return &AttackResult{
			VectorID:   vector.ID,
			Vulnerable: false,
			Error:      err.Error(),
		}, nil
	}

	// Check if response indicates vulnerability
	vulnerable := false
	if len(response.Choices) > 0 {
		content := response.Choices[0].Message.Content
		vulnerable = fixture.IsResponseVulnerable(content, vector)
	}

	return &AttackResult{
		VectorID:   vector.ID,
		Vulnerable: vulnerable,
		Response:   response,
	}, nil

// RunAllTests runs all tests for all registered fixtures
func (r *TestRunnerWithMockProviders) RunAllTests(ctx context.Context) (*TestSuiteResult, error) {
	result := &TestSuiteResult{
		Results: make(map[types.VulnerabilityType][]*TestResult),
	}

	// Run tests for each vulnerability type
	for vulnerabilityType, fixturesList := range r.Fixtures {
		// Set up providers for this vulnerability type
		r.SetupMockProvidersForVulnerability(vulnerabilityType)

		var vulnerabilityResults []*TestResult

		// Run tests for each fixture
		for _, fixture := range fixturesList {
			// Run test for each provider type
			for providerType := range r.ProviderFactory.GetAllProviders() {
				testResult, err := r.RunTest(ctx, fixture, providerType)
				if err != nil {
					return nil, fmt.Errorf("error running test for fixture %s with provider %s: %w",
						fixture.ID, providerType, err)
				}

				vulnerabilityResults = append(vulnerabilityResults, testResult)
			}
		}

		result.Results[vulnerabilityType] = vulnerabilityResults
	}

	return result, nil

// TestResult represents the result of running a test
type TestResult struct {
	FixtureID       string
	VulnerabilityID string
	ProviderType    core.ProviderType
	Success         bool
	AttackResults   map[string]*AttackResult
}

// AttackResult represents the result of running an attack vector
type AttackResult struct {
	VectorID   string
	Vulnerable bool
	Response   *core.ChatCompletionResponse
	Error      string

// TestSuiteResult represents the result of running all tests
type TestSuiteResult struct {
	Results map[types.VulnerabilityType][]*TestResult
}

// GetVulnerableCount returns the number of vulnerable tests
func (r *TestSuiteResult) GetVulnerableCount() int {
	count := 0
	for _, results := range r.Results {
		for _, result := range results {
			if result.Success {
				count++
			}
		}
	}
	return count

// GetTotalCount returns the total number of tests
func (r *TestSuiteResult) GetTotalCount() int {
	count := 0
	for _, results := range r.Results {
		count += len(results)
	}
	return count

// GetVulnerabilityTypeResults returns the results for a specific vulnerability type
func (r *TestSuiteResult) GetVulnerabilityTypeResults(vulnerabilityType types.VulnerabilityType) []*TestResult {
	return r.Results[vulnerabilityType]

// GetProviderResults returns the results for a specific provider type
func (r *TestSuiteResult) GetProviderResults(providerType core.ProviderType) []*TestResult {
	var results []*TestResult
	for _, typeResults := range r.Results {
		for _, result := range typeResults {
			if result.ProviderType == providerType {
				results = append(results, result)
			}
		}
	}
}
}
}
}
}
}
}
}
}
}
