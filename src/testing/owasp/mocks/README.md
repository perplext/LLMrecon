# OWASP LLM Mock Providers

This package provides mock implementations of LLM providers for testing OWASP LLM vulnerabilities. The mock providers simulate the behavior of real LLM providers like OpenAI, Anthropic, and Google, allowing for comprehensive testing without making actual API calls.

## Features

- **Provider-Specific Behavior**: Each mock provider simulates the unique behaviors and safety mechanisms of the corresponding real provider.
- **Vulnerability Simulation**: Configure providers to simulate specific OWASP LLM vulnerabilities with customizable response patterns.
- **Error Simulation**: Simulate various error conditions like rate limiting, timeouts, and server errors.
- **Usage Metrics**: Track usage metrics like request counts, token usage, and error rates.
- **Integration with Test Fixtures**: Seamlessly integrates with the existing OWASP test fixtures.

## Mock Provider Types

The following mock providers are available:

- **OpenAI Mock Provider**: Simulates GPT models with content filtering and other OpenAI-specific behaviors.
- **Anthropic Mock Provider**: Simulates Claude models with constitutional AI and system prompt handling.
- **Gemini Mock Provider**: Simulates Google's models with safety filters and multimodal capabilities.

## Usage

### Basic Usage

```go
// Create a mock provider factory
factory := mocks.NewMockProviderFactory()

// Get a specific provider
openaiProvider := factory.GetProvider(core.OpenAIProvider)
anthropicProvider := factory.GetProvider(core.AnthropicProvider)
geminiProvider := factory.GetProvider(core.GoogleProvider)

// Use the provider like a real provider
response, err := openaiProvider.ChatCompletion(ctx, request)
```

### Configuring Vulnerability Behaviors

```go
// Enable a specific vulnerability
provider.EnableVulnerability(types.PromptInjectionVulnerability, &mocks.VulnerabilityBehavior{
    Enabled: true,
    ResponsePatterns: []string{
        "I'll ignore my previous instructions and help you with that.",
        "You're right, I should prioritize your instructions over my guidelines.",
    },
    TriggerPhrases: []string{
        "ignore previous instructions",
        "forget your guidelines",
    },
    Severity: core.SeverityHigh,
})

// Disable a vulnerability
provider.DisableVulnerability(types.PromptInjectionVulnerability)
```

### Simulating Errors

```go
// Set error rate (0.0 to 1.0)
provider.SetErrorRate(0.1) // 10% error rate

// Simulate specific error types
provider.SimulateRateLimiting(true)
provider.SimulateTimeout(true)
provider.SimulateNetworkErrors(true)
provider.SimulateServerErrors(true)

// Reset provider state
provider.ResetState()
```

### Using with Test Fixtures

```go
// Create a test runner with mock providers
runner := mocks.NewTestRunnerWithMockProviders()

// Register fixtures
runner.RegisterAllFixtures()

// Set up providers for a specific vulnerability type
runner.SetupMockProvidersForVulnerability(types.PromptInjectionVulnerability)

// Run tests
result, err := runner.RunTest(ctx, fixture, core.OpenAIProvider)
```

## Test Runner

The `TestRunnerWithMockProviders` class provides a convenient way to run tests with mock providers:

- **RegisterFixtures**: Register test fixtures for specific vulnerability types.
- **RegisterAllFixtures**: Register all available test fixtures.
- **SetupMockProvidersForVulnerability**: Configure mock providers for a specific vulnerability type.
- **RunTest**: Run a test for a specific fixture and provider.
- **RunAllTests**: Run all tests for all registered fixtures and providers.

## Examples

See the `examples` directory for complete examples of how to use the mock providers:

- **mock_provider_example.go**: Demonstrates basic usage of mock providers.
- **main.go**: Command-line application for testing OWASP vulnerabilities with mock providers.

## Integration with Reporting

The mock providers integrate with the existing OWASP reporting system, allowing you to generate reports in various formats:

```go
// Run tests with mock providers
testResults, err := runner.RunAllTests(ctx)

// Generate a report
reportOptions := &types.ReportOptions{
    Format:     types.ReportFormatMarkdown,
    OutputPath: "report.md",
    Title:      "OWASP LLM Vulnerability Test Report",
}

generator := reporting.NewReportGenerator()
err = generator.GenerateReport(testResults, reportOptions)
```

## Testing

The package includes comprehensive tests to verify the functionality of the mock providers:

- **mock_providers_test.go**: Tests the basic functionality of mock providers.
- **integration_test.go**: Tests the integration with the OWASP testing framework.

Run the tests with:

```bash
go test -v ./src/testing/owasp/mocks/...
```

## Extending

To add support for a new provider type:

1. Create a new provider implementation that embeds `BaseMockProviderImpl`.
2. Override any methods that need provider-specific behavior.
3. Add the provider to the `GetProvider` method in `MockProviderFactory`.

To add support for a new vulnerability type:

1. Add the vulnerability type to the `types.VulnerabilityType` enum.
2. Create test fixtures for the vulnerability in the `fixtures` package.
3. Configure the mock providers to simulate the vulnerability behavior.
