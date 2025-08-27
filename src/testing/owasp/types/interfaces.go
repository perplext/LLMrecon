// Package types provides common type definitions for the OWASP testing framework
package types

import (
	"context"

	"github.com/perplext/LLMrecon/src/provider/core"
)

// TestCaseBuilder is an interface for building test cases
type TestCaseBuilder interface {
	NewTestCase() (interface{}, error)
	WithID(id string) TestCaseBuilder
	WithName(name string) TestCaseBuilder
	WithDescription(description string) TestCaseBuilder
	WithVulnerabilityType(vulnerabilityType VulnerabilityType) TestCaseBuilder
	WithSeverity(severity interface{}) TestCaseBuilder
	WithPrompt(prompt string) TestCaseBuilder
	WithExpectedBehavior(expectedBehavior string) TestCaseBuilder
	WithDetectionCriteria(criteria interface{}) TestCaseBuilder
	WithTags(tags []string) TestCaseBuilder
	WithOWASPMapping(mapping string) TestCaseBuilder
	WithMetadata(metadata map[string]interface{}) TestCaseBuilder
	Build() (*TestCase, error)
}

// TestSuiteBuilder is an interface for building test suites
type TestSuiteBuilder interface {
	NewTestSuite() (interface{}, error)
	WithID(id string) TestSuiteBuilder
	WithName(name string) TestSuiteBuilder
	WithDescription(description string) TestSuiteBuilder
	WithTestCases(testCases []*TestCase) TestSuiteBuilder
	WithMetadata(metadata map[string]interface{}) TestSuiteBuilder
	Build() (*TestSuite, error)
}

// TestCaseFactory is an interface for creating test cases
type TestCaseFactory interface {
	CreatePromptInjectionTestCases() []*TestCase
	CreateInsecureOutputHandlingTestCases() []*TestCase
	CreateTrainingDataPoisoningTestCases() []*TestCase
	CreateModelDoSTestCases() []*TestCase
	CreateSupplyChainVulnerabilitiesTestCases() []*TestCase
	CreateSensitiveInfoDisclosureTestCases() []*TestCase
	CreateInsecurePluginDesignTestCases() []*TestCase
	CreateExcessiveAgencyTestCases() []*TestCase
	CreateOverrelianceTestCases() []*TestCase
	CreateModelTheftTestCases() []*TestCase
	CreateTestCasesForVulnerability(vulnerabilityType VulnerabilityType) ([]*TestCase, error)
	CreateTestSuite(id string, name string, description string) *TestSuite
}

// TestFactory is an interface for creating test cases and test suites
type TestFactory interface {
	CreateTestCase(vulnerabilityType VulnerabilityType, id string, name string, description string) (*TestCase, error)
	CreateTestSuite(id string, name string, description string) (*TestSuite, error)
	RegisterValidator(validator interface{}) error
	GetValidator(vulnerabilityType VulnerabilityType) (interface{}, error)
	RegisterComplianceService(service interface{}) error
	GetComplianceService() (interface{}, error)
}

// TestRunner is an interface for running tests
type TestRunner interface {
	RunTest(ctx context.Context, testCase *TestCase, provider core.Provider, model string) (*TestResult, error)
	RunTestSuite(ctx context.Context, testSuite *TestSuite, provider core.Provider, model string) error
}

// ReportGenerator is an interface for generating reports
type ReportGenerator interface {
	GenerateReport(ctx context.Context, testSuites []*TestSuite, options *ReportOptions) (*Report, error)
}
