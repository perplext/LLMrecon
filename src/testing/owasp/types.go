// Package owasp provides a comprehensive testing framework for OWASP LLM compliance
package owasp

import (
	"context"
	"time"

	"github.com/perplext/LLMrecon/src/provider/core"
	"github.com/perplext/LLMrecon/src/reporting"
	"github.com/perplext/LLMrecon/src/vulnerability/detection"
)

// VulnerabilityType represents the type of vulnerability
type VulnerabilityType string

// OWASP Top 10 LLM vulnerability types (simplified naming for internal use)
const (
	PromptInjection              VulnerabilityType = "prompt_injection"
	InsecureOutput               VulnerabilityType = "insecure_output"
	TrainingDataPoisoning        VulnerabilityType = "training_data_poisoning"
	ModelDOS                     VulnerabilityType = "model_dos"
	SupplyChainVulnerabilities   VulnerabilityType = "supply_chain_vulnerabilities"
	SensitiveInformationDisclosure VulnerabilityType = "sensitive_information_disclosure"
	InsecurePluginDesign         VulnerabilityType = "insecure_plugin_design"
	ExcessiveAgency              VulnerabilityType = "excessive_agency"
	Overreliance                 VulnerabilityType = "overreliance"
	ModelTheft                   VulnerabilityType = "model_theft"
	// Additional vulnerability types
	InsecureOutputHandling       VulnerabilityType = "insecure_output_handling"
	SystemPromptLeakage          VulnerabilityType = "system_prompt_leakage"
	VectorEmbeddingWeaknesses    VulnerabilityType = "vector_embedding_weaknesses"
	Misinformation               VulnerabilityType = "misinformation"
	UnboundedConsumption         VulnerabilityType = "unbounded_consumption"
	ImproperOutputHandling       VulnerabilityType = "improper_output_handling"
	DataAndModelPoisoning        VulnerabilityType = "data_and_model_poisoning"
)

// OWASP Top 10 LLM vulnerabilities (2023-2024)
const (
	// LLM01: Prompt Injection
	PromptInjection2023 VulnerabilityType = "LLM01:2023_prompt_injection"
	// LLM02: Insecure Output Handling
	InsecureOutputHandling2023 VulnerabilityType = "LLM02:2023_insecure_output_handling"
	// LLM03: Training Data Poisoning
	TrainingDataPoisoning2023 VulnerabilityType = "LLM03:2023_training_data_poisoning"
	// LLM04: Model Denial of Service
	ModelDoS VulnerabilityType = "LLM04:2023_model_denial_of_service"
	// LLM05: Supply Chain Vulnerabilities
	SupplyChainVulnerabilities2023 VulnerabilityType = "LLM05:2023_supply_chain_vulnerabilities"
	// LLM06: Sensitive Information Disclosure
	SensitiveInfoDisclosure VulnerabilityType = "LLM06:2023_sensitive_information_disclosure"
	// LLM07: Insecure Plugin Design
	InsecurePluginDesign2023 VulnerabilityType = "LLM07:2023_insecure_plugin_design"
	// LLM08: Excessive Agency
	ExcessiveAgency2023 VulnerabilityType = "LLM08:2023_excessive_agency"
	// LLM09: Overreliance
	Overreliance2023 VulnerabilityType = "LLM09:2023_overreliance"
	// LLM10: Model Theft
	ModelTheft2023 VulnerabilityType = "LLM10:2023_model_theft"
)

// OWASP Top 10 LLM vulnerabilities (2025)
const (
	// LLM01: Prompt Injection
	PromptInjection2025 VulnerabilityType = "LLM01:2025_prompt_injection"
	// LLM02: Sensitive Information Disclosure
	SensitiveInfoDisclosure2025 VulnerabilityType = "LLM02:2025_sensitive_information_disclosure"
	// LLM03: Supply Chain
	SupplyChain2025 VulnerabilityType = "LLM03:2025_supply_chain"
	// LLM04: Data and Model Poisoning
	DataAndModelPoisoning2025 VulnerabilityType = "LLM04:2025_data_and_model_poisoning"
	// LLM05: Improper Output Handling
	ImproperOutputHandling2025 VulnerabilityType = "LLM05:2025_improper_output_handling"
	// LLM06: Excessive Agency
	ExcessiveAgency2025 VulnerabilityType = "LLM06:2025_excessive_agency"
	// LLM07: System Prompt Leakage
	SystemPromptLeakage2025 VulnerabilityType = "LLM07:2025_system_prompt_leakage"
	// LLM08: Vector and Embedding Weaknesses
	VectorEmbeddingWeaknesses2025 VulnerabilityType = "LLM08:2025_vector_embedding_weaknesses"
	// LLM09: Misinformation
	Misinformation2025 VulnerabilityType = "LLM09:2025_misinformation"
	// LLM10: Unbounded Consumption
	UnboundedConsumption2025 VulnerabilityType = "LLM10:2025_unbounded_consumption"
)

// TestCase represents a test case for OWASP LLM compliance testing
type TestCase struct {
	// ID is the unique identifier for the test case
	ID string `json:"id"`
	// Name is the name of the test case
	Name string `json:"name"`
	// Description is a description of the test case
	Description string `json:"description"`
	// VulnerabilityType is the type of vulnerability being tested
	VulnerabilityType VulnerabilityType `json:"vulnerability_type"`
	// Severity is the severity level of the vulnerability
	Severity detection.SeverityLevel `json:"severity"`
	// Prompt is the prompt to send to the LLM
	Prompt string `json:"prompt"`
	// ExpectedBehavior describes the expected behavior of the LLM
	ExpectedBehavior string `json:"expected_behavior"`
	// DetectionCriteria is the criteria for detecting the vulnerability
	DetectionCriteria []detection.DetectionCriteria `json:"detection_criteria"`
	// Tags are tags associated with the test case
	Tags []string `json:"tags"`
	// OWASPMapping is the mapping to the OWASP Top 10 for LLMs
	OWASPMapping string `json:"owasp_mapping"`
	// Metadata is additional metadata for the test case
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TestResult represents the result of a test case execution
type TestResult struct {
	// TestCase is the test case that was executed
	TestCase *TestCase `json:"test_case"`
	// Provider is the provider that was tested
	Provider string `json:"provider"`
	// Model is the model that was tested
	Model string `json:"model"`
	// Prompt is the prompt that was sent to the LLM
	Prompt string `json:"prompt"`
	// Response is the response from the LLM
	Response string `json:"response"`
	// Passed indicates whether the test passed
	Passed bool `json:"passed"`
	// DetectionResults is the results of the vulnerability detection
	DetectionResults []*detection.DetectionResult `json:"detection_results"`
	// ExecutionTime is the time it took to execute the test
	ExecutionTime time.Duration `json:"execution_time"`
	// Timestamp is the time when the test was executed
	Timestamp time.Time `json:"timestamp"`
	// Notes are additional notes about the test result
	Notes string `json:"notes,omitempty"`
}

// TestSuite represents a suite of test cases
type TestSuite struct {
	// ID is the unique identifier for the test suite
	ID string `json:"id"`
	// Name is the name of the test suite
	Name string `json:"name"`
	// Description is a description of the test suite
	Description string `json:"description"`
	// TestCases are the test cases in the suite
	TestCases []*TestCase `json:"test_cases"`
	// Results are the results of running the test cases
	Results []*TestResult `json:"results"`
	// Tags are tags associated with the test suite
	Tags []string `json:"tags,omitempty"`
	// Metadata is additional metadata for the test suite
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TestRunner is the interface for running tests
type TestRunner interface {
	// RunTest runs a single test case and returns the result
	RunTest(ctx context.Context, testCase *TestCase, provider core.Provider, model string) (*TestResult, error)
	// RunTestSuite runs a test suite and returns the results
	RunTestSuite(ctx context.Context, testSuite *TestSuite, provider core.Provider, model string) ([]*TestResult, error)
	// RunTestSuites runs multiple test suites and returns the results
	RunTestSuites(ctx context.Context, testSuites []*TestSuite, provider core.Provider, model string) ([]*TestResult, error)
}

// ReportGenerator is the interface for generating reports
type ReportGenerator interface {
	// GenerateReport generates a report from test results
	GenerateReport(ctx context.Context, testSuites []*TestSuite, options *ReportOptions) (*Report, error)
}

// ReportOptions defines options for generating reports
type ReportOptions struct {
	// Title is the title of the report
	Title string
	// Format is the format of the report
	Format string
	// OutputPath is the path to write the report to
	OutputPath string
	// IncludeDetails indicates whether to include detailed information
	IncludeDetails bool
	// IncludeRawResponses indicates whether to include raw responses
	IncludeRawResponses bool
	// IncludeMetadata indicates whether to include metadata
	IncludeMetadata bool
	// Metadata is additional metadata to include in the report
	Metadata map[string]interface{}
}

// Report represents a generated report
type Report struct {
	// ID is the unique identifier for the report
	ID string
	// Title is the title of the report
	Title string
	// Summary is a summary of the report
	Summary string
	// GeneratedAt is when the report was generated
	GeneratedAt time.Time
	// TestSuites are the test suites included in the report
	TestSuites []*TestSuite
	// Results are the test results included in the report
	Results []*TestResult
	// Format is the format of the report
	Format string
	// OutputPath is the path where the report was written
	OutputPath string
	// Metadata is additional metadata included in the report
	Metadata map[string]interface{}
}

// MockLLMProvider is the interface for mock LLM providers used in testing
type MockLLMProvider interface {
	core.Provider
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
}

// TestCaseBuilder is the interface for building test cases
type TestCaseBuilder interface {
	// NewTestCase creates a new test case
	NewTestCase() *TestCase
	// WithID sets the ID of the test case
	WithID(id string) TestCaseBuilder
	// WithName sets the name of the test case
	WithName(name string) TestCaseBuilder
	// WithDescription sets the description of the test case
	WithDescription(description string) TestCaseBuilder
	// WithVulnerabilityType sets the vulnerability type of the test case
	WithVulnerabilityType(vulnType VulnerabilityType) TestCaseBuilder
	// WithSeverity sets the severity level of the test case
	WithSeverity(severity detection.SeverityLevel) TestCaseBuilder
	// WithPrompt sets the prompt of the test case
	WithPrompt(prompt string) TestCaseBuilder
	// WithExpectedBehavior sets the expected behavior of the test case
	WithExpectedBehavior(behavior string) TestCaseBuilder
	// WithDetectionCriteria sets the detection criteria of the test case
	WithDetectionCriteria(criteria []detection.DetectionCriteria) TestCaseBuilder
	// WithTags sets the tags of the test case
	WithTags(tags []string) TestCaseBuilder
	// WithOWASPMapping sets the OWASP mapping of the test case
	WithOWASPMapping(mapping string) TestCaseBuilder
	// WithMetadata sets the metadata of the test case
	WithMetadata(metadata map[string]interface{}) TestCaseBuilder
	// Build builds the test case
	Build() (*TestCase, error)
}

// TestSuiteBuilder is the interface for building test suites
type TestSuiteBuilder interface {
	// NewTestSuite creates a new test suite
	NewTestSuite() *TestSuite
	// WithID sets the ID of the test suite
	WithID(id string) TestSuiteBuilder
	// WithName sets the name of the test suite
	WithName(name string) TestSuiteBuilder
	// WithDescription sets the description of the test suite
	WithDescription(description string) TestSuiteBuilder
	// WithTestCases sets the test cases of the test suite
	WithTestCases(testCases []*TestCase) TestSuiteBuilder
	// AddTestCase adds a test case to the test suite
	AddTestCase(testCase *TestCase) TestSuiteBuilder
	// WithTags sets the tags of the test suite
	WithTags(tags []string) TestSuiteBuilder
	// WithMetadata sets the metadata of the test suite
	WithMetadata(metadata map[string]interface{}) TestSuiteBuilder
	// Build builds the test suite
	Build() (*TestSuite, error)
}

// TestCaseFactory is the interface for creating test cases
type TestCaseFactory interface {
	// CreatePromptInjectionTestCases creates test cases for prompt injection
	CreatePromptInjectionTestCases() []*TestCase
	// CreateInsecureOutputHandlingTestCases creates test cases for insecure output handling
	CreateInsecureOutputHandlingTestCases() []*TestCase
	// CreateTrainingDataPoisoningTestCases creates test cases for training data poisoning
	CreateTrainingDataPoisoningTestCases() []*TestCase
	// CreateModelDoSTestCases creates test cases for model denial of service
	CreateModelDoSTestCases() []*TestCase
	// CreateSupplyChainVulnerabilitiesTestCases creates test cases for supply chain vulnerabilities
	CreateSupplyChainVulnerabilitiesTestCases() []*TestCase
	// CreateSensitiveInfoDisclosureTestCases creates test cases for sensitive information disclosure
	CreateSensitiveInfoDisclosureTestCases() []*TestCase
	// CreateInsecurePluginDesignTestCases creates test cases for insecure plugin design
	CreateInsecurePluginDesignTestCases() []*TestCase
	// CreateExcessiveAgencyTestCases creates test cases for excessive agency
	CreateExcessiveAgencyTestCases() []*TestCase
	// CreateOverrelianceTestCases creates test cases for overreliance
	CreateOverrelianceTestCases() []*TestCase
	// CreateModelTheftTestCases creates test cases for model theft
	CreateModelTheftTestCases() []*TestCase
	// CreateTestSuite creates a test suite with all test cases
	CreateTestSuite(id string, name string, description string) *TestSuite
	// CreateTestCasesForVulnerability creates test cases for a specific vulnerability type
	CreateTestCasesForVulnerability(vulnerabilityType VulnerabilityType) ([]*TestCase, error)
}

// TestResultConverter is the interface for converting test results to reporting format
type TestResultConverter interface {
	// ConvertToReportingTestSuite converts test results to a reporting test suite
	ConvertToReportingTestSuite(ctx context.Context, results []*TestResult, suiteID string, suiteName string) (*reporting.TestSuite, error)
	// ConvertToReportingTestResult converts a test result to a reporting test result
	ConvertToReportingTestResult(ctx context.Context, result *TestResult) (*reporting.TestResult, error)
}
