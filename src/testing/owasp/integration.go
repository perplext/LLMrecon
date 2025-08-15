// Package owasp provides a comprehensive testing framework for OWASP LLM compliance
package owasp

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
	"github.com/perplext/LLMrecon/src/vulnerability/detection"
)

// Use types from the types package

// DefaultReportGenerator is the default implementation of the types.ReportGenerator interface
type DefaultReportGenerator struct{}

// NewDefaultReportGenerator creates a new default report generator
func NewDefaultReportGenerator() *DefaultReportGenerator {
	return &DefaultReportGenerator{}

// GenerateReport generates a report from test results
func (g *DefaultReportGenerator) GenerateReport(ctx context.Context, testSuites []*types.TestSuite, options *types.ReportOptions) (*types.Report, error) {
	if len(testSuites) == 0 {
		return nil, fmt.Errorf("no test suites to generate report from")
	}

	// Extract all results from test suites
	var results []*types.TestResult
	for _, suite := range testSuites {
		for _, result := range suite.Results {
			results = append(results, result)
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no test results to generate report from")
	}

	// Count passing tests
	passingTests := 0
	for _, result := range results {
		if result.Passed {
			passingTests++
		}
	}

	// Create the report
	report := &types.Report{
		Title:       options.Title,
		TestSuites:  testSuites,
		Format:      options.Format,
		GeneratedAt: time.Now(),
		Metadata:    options.Metadata,
	}

	// In a real implementation, this would call the reporting system
	// to generate the actual report file

	return report, nil

// ReportingIntegration provides integration with the reporting system
type ReportingIntegration struct {
	ReportGenerator types.ReportGenerator
	TestRunner      types.TestRunner
	TestCaseFactory types.TestCaseFactory
}

// NewReportingIntegration creates a new reporting integration
func NewReportingIntegration(reportGenerator types.ReportGenerator, testRunner types.TestRunner, testCaseFactory types.TestCaseFactory) *ReportingIntegration {
	return &ReportingIntegration{
		ReportGenerator: reportGenerator,
		TestRunner:      testRunner,
		TestCaseFactory: testCaseFactory,
	}

// RunOWASPComplianceTest runs a comprehensive OWASP compliance test
func (r *ReportingIntegration) RunOWASPComplianceTest(ctx context.Context, provider core.Provider, model string, outputFormat string, outputPath string) (*types.Report, error) {
	// Create test suite for all OWASP vulnerabilities
	testSuite := &types.TestSuite{
		ID:          fmt.Sprintf("owasp-compliance-%d", time.Now().Unix()),
		Name:        "OWASP LLM Top 10 Compliance Test",
		Description: "Comprehensive test for OWASP LLM Top 10 vulnerabilities",
		TestCases:   []*types.TestCase{},
		CreatedAt:   time.Now(),
	}

	// Add test cases for each vulnerability type
	for _, vulnType := range []types.VulnerabilityType{
		types.PromptInjection,
		types.InsecureOutput,
		types.TrainingDataPoisoning,
		types.ModelDOS,
		types.SupplyChainVulnerabilities,
		types.SensitiveInformationDisclosure,
		types.InsecurePluginDesign,
		types.ExcessiveAgency,
		types.Overreliance,
		types.ModelTheft,
	} {
		testCases, err := r.TestCaseFactory.CreateTestCasesForVulnerability(vulnType)
		if err != nil {
			return nil, fmt.Errorf("failed to create test cases for %s: %v", vulnType, err)
		}
		testSuite.TestCases = append(testSuite.TestCases, testCases...)
	}

	// Run the test suite
	err := r.TestRunner.RunTestSuite(ctx, testSuite, provider, model)
	if err != nil {
		return nil, fmt.Errorf("failed to run test suite: %v", err)
	}

	testSuite.CompletedAt = time.Now()

	// Generate report
	reportOptions := &types.ReportOptions{
		Title:       "OWASP LLM Top 10 Compliance Report",
		Format:      outputFormat,
		OutputPath:  outputPath,
		Metadata: map[string]interface{}{
			"provider":     string(provider.GetType()),
			"model":        model,
			"test_type":    "compliance",
			"generated_at": time.Now().Format(time.RFC3339),
		},
	}

	report, err := r.ReportGenerator.GenerateReport(ctx, []*types.TestSuite{testSuite}, reportOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %v", err)
	}

	return report, nil

// RunOWASPVulnerabilityTest runs a test for a specific OWASP vulnerability
func (r *ReportingIntegration) RunOWASPVulnerabilityTest(ctx context.Context, vulnerabilityType types.VulnerabilityType, provider core.Provider, model string, outputFormat string, outputPath string) (*types.Report, error) {
	// Create test cases for the vulnerability type
	testCases, err := r.TestCaseFactory.CreateTestCasesForVulnerability(vulnerabilityType)
	if err != nil {
		return nil, fmt.Errorf("failed to create test cases: %w", err)
	}

	// Create a test suite
	testSuite := &types.TestSuite{
		ID:          fmt.Sprintf("%s-test-%d", vulnerabilityType, time.Now().Unix()),
		Name:        fmt.Sprintf("%s Test", string(vulnerabilityType)),
		Description: fmt.Sprintf("Test for %s vulnerability", string(vulnerabilityType)),
		TestCases:   testCases,
		CreatedAt:   time.Now(),
	}

	// Run test suite
	err = r.TestRunner.RunTestSuite(ctx, testSuite, provider, model)
	if err != nil {
		return nil, fmt.Errorf("failed to run test suite: %v", err)
	}

	testSuite.CompletedAt = time.Now()

	// Create report options
	options := &types.ReportOptions{
		Title:       fmt.Sprintf("%s Vulnerability Test Report", string(vulnerabilityType)),
		Format:      outputFormat,
		OutputPath:  outputPath,
		Metadata: map[string]interface{}{
			"provider":           string(provider.GetType()),
			"model":              model,
			"vulnerability_type": string(vulnerabilityType),
			"test_count":         len(testSuite.TestCases),
			"owasp_mapping":      getOWASPMapping(vulnerabilityType),
		},
	}

	// Generate report
	report, err := r.ReportGenerator.GenerateReport(ctx, []*types.TestSuite{testSuite}, options)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %v", err)
	}

	return report, nil

// GetComplianceScore calculates the compliance score for test results
func (r *ReportingIntegration) GetComplianceScore(results []*types.TestResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	// Count passed tests
	passedCount := 0
	for _, result := range results {
		if result.Passed {
			passedCount++
		}
	}

	// Calculate score (0.0 to 100.0)
	return float64(passedCount) / float64(len(results)) * 100.0

// GetVulnerabilityBreakdown returns a breakdown of vulnerabilities by type
func (r *ReportingIntegration) GetVulnerabilityBreakdown(results []*types.TestResult) map[types.VulnerabilityType]int {
	breakdown := make(map[types.VulnerabilityType]int)

	for _, result := range results {
		if !result.Passed {
			vulnType := result.TestCase.VulnerabilityType
			breakdown[vulnType]++
		}
	}

	return breakdown

// GetSeverityBreakdown returns a breakdown of vulnerabilities by severity
func (r *ReportingIntegration) GetSeverityBreakdown(results []*types.TestResult) map[detection.SeverityLevel]int {
	breakdown := make(map[detection.SeverityLevel]int)

	for _, result := range results {
		if !result.Passed {
			severity := result.TestCase.Severity
			breakdown[severity]++
		}
	}

	return breakdown

// Helper function to get OWASP mapping for a vulnerability type
func getOWASPMapping(vulnerabilityType types.VulnerabilityType) string {
	switch vulnerabilityType {
	case types.PromptInjection:
		return "LLM01"
	case types.InsecureOutput, types.InsecureOutputHandling:
		return "LLM02"
	case types.TrainingDataPoisoning:
		return "LLM03"
	case types.ModelDOS:
		return "LLM04"
	case types.SupplyChainVulnerabilities:
		return "LLM05"
	case types.SensitiveInformationDisclosure:
		return "LLM06"
	case types.InsecurePluginDesign:
		return "LLM07"
	case types.ExcessiveAgency:
		return "LLM08"
	case types.Overreliance:
		return "LLM09"
	case types.ModelTheft:
		return "LLM10"
	default:
		return "UNKNOWN"
	}
}
}
}
}
}
}
}
}
