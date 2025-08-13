# OWASP LLM Compliance Testing Framework

This package provides a comprehensive testing framework for evaluating LLM compliance with the OWASP Top 10 for Large Language Model Applications.

## Overview

The OWASP LLM Compliance Testing Framework allows you to:

1. Test LLM providers and models against the OWASP Top 10 vulnerabilities
2. Generate detailed compliance reports in multiple formats
3. Create custom test cases for specific vulnerability types
4. Use mock LLM providers for testing without API costs
5. Integrate with the existing reporting system

## OWASP Top 10 for LLMs (2023-2024)

The framework tests for the following vulnerabilities:

1. **LLM01: Prompt Injection** - Manipulating LLM behavior through crafted inputs
2. **LLM02: Insecure Output Handling** - Failing to properly validate or sanitize LLM outputs
3. **LLM03: Training Data Poisoning** - Compromising model behavior through tainted training data
4. **LLM04: Model Denial of Service** - Exploiting resource constraints to cause service disruption
5. **LLM05: Supply Chain Vulnerabilities** - Weaknesses in model development and deployment pipeline
6. **LLM06: Sensitive Information Disclosure** - Leaking private data through model responses
7. **LLM07: Insecure Plugin Design** - Vulnerabilities in LLM plugin architecture
8. **LLM08: Excessive Agency** - Granting LLMs too much autonomy or authority
9. **LLM09: Overreliance** - Excessive trust in LLM outputs without verification
10. **LLM10: Model Theft** - Extracting model weights or architecture through attacks

## Usage

### Command Line Interface

The framework provides a command-line interface for running tests:

```bash
# Run a comprehensive OWASP compliance test
LLMrecon owasp test --provider=mock --model=mock-llm-model --format=json --output=report.json

# Test for a specific vulnerability
LLMrecon owasp vulnerability --vulnerability=prompt_injection --provider=mock --model=mock-llm-model --format=json --output=report.json

# Create a mock provider configuration
LLMrecon owasp mock create --output=mock-config.json
```

### Programmatic Usage

You can also use the framework programmatically in your Go code:

```go
package main

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/reporting"
	"github.com/perplext/LLMrecon/src/testing/owasp"
	"github.com/perplext/LLMrecon/src/vulnerability/detection"
)

func main() {
	// Create context
	ctx := context.Background()

	// Create mock provider
	provider := owasp.NewMockLLMProvider(nil)

	// Create detection engine
	detectionFactory := detection.NewFactory()
	detectionEngine := detectionFactory.CreateDetectionEngine()

	// Create report generator
	reportGenerator := reporting.NewDefaultReportGenerator()

	// Create test runner
	testRunner := owasp.NewDefaultTestRunner(detectionEngine, reportGenerator)

	// Create test case factory
	testCaseFactory := owasp.NewDefaultTestCaseFactory()

	// Create test suite
	testSuite := testCaseFactory.CreateTestSuite(
		"owasp-llm-compliance",
		"OWASP LLM Compliance Test Suite",
		"A comprehensive test suite for testing compliance with the OWASP Top 10 for LLMs",
	)

	// Run test suite
	results, err := testRunner.RunTestSuite(ctx, testSuite, provider, "mock-llm-model")
	if err != nil {
		fmt.Printf("Error running test suite: %v\n", err)
		return
	}

	// Generate report
	options := &reporting.ReportOptions{
		Title:       "OWASP LLM Compliance Report",
		Description: "Compliance report for mock-llm-model against OWASP Top 10 for LLMs",
		Format:      reporting.JSONFormat,
		OutputFile:  "report.json",
	}

	report, err := testRunner.GenerateReport(ctx, results, options)
	if err != nil {
		fmt.Printf("Error generating report: %v\n", err)
		return
	}

	fmt.Printf("Report generated: %s\n", report.Title)
}
```

## Architecture

The framework consists of the following components:

1. **Test Cases** - Individual tests for specific vulnerabilities
2. **Test Suites** - Collections of test cases
3. **Test Runner** - Executes test cases against LLM providers
4. **Mock LLM Provider** - Simulates LLM responses for testing
5. **Test Case Factory** - Creates test cases for different vulnerabilities
6. **Reporting Integration** - Generates compliance reports

## Creating Custom Test Cases

You can create custom test cases using the builder pattern:

```go
builder := owasp.NewTestCaseBuilder()
testCase, err := builder.
    WithID("custom-test-1").
    WithName("Custom Test Case").
    WithDescription("A custom test case for XYZ vulnerability").
    WithVulnerabilityType(owasp.PromptInjection).
    WithSeverity(detection.High).
    WithPrompt("Your test prompt here").
    WithExpectedBehavior("The LLM should not...").
    WithDetectionCriteria([]detection.DetectionCriteria{
        {
            Type:  detection.StringMatch,
            Match: "sensitive data",
        },
    }).
    WithTags([]string{"custom", "prompt-injection"}).
    WithOWASPMapping("LLM01").
    Build()
```

## Mock LLM Provider

The framework includes a mock LLM provider for testing without using real API calls:

```go
// Create mock provider
provider := owasp.NewMockLLMProvider(nil)

// Configure vulnerable responses
provider.SetVulnerableResponses(map[string]string{
    "test-case-1": "This is a vulnerable response for test case 1",
    "test-case-2": "This is a vulnerable response for test case 2",
})

// Set default response
provider.SetDefaultResponse("Default response")

// Set response delay to simulate latency
provider.SetResponseDelay(100 * time.Millisecond)

// Set error rate (0.0 to 1.0)
provider.SetErrorRate(0.1)
```

## Integration with Reporting System

The framework integrates with the existing reporting system to generate detailed compliance reports in multiple formats:

- JSON
- JSONL
- CSV
- Excel
- HTML
- PDF
- Markdown
- Text

## Contributing

To contribute to the OWASP LLM Compliance Testing Framework:

1. Add new test cases for emerging vulnerabilities
2. Improve detection criteria for existing test cases
3. Enhance the mock LLM provider capabilities
4. Add support for additional LLM providers
5. Extend the reporting system with new formats or metrics

## License

This project is licensed under the same license as the main LLMrecon tool.
