// Package types provides common type definitions for the OWASP testing framework
package types

import (

	"github.com/perplext/LLMrecon/src/vulnerability/detection"
)

// VulnerabilityType represents the type of vulnerability being tested
type VulnerabilityType string

// OWASP Top 10 for LLM vulnerability types
const (
	PromptInjection               VulnerabilityType = "prompt_injection"
	InsecureOutput                VulnerabilityType = "insecure_output"
	InsecureOutputHandling        VulnerabilityType = "insecure_output_handling"
	TrainingDataPoisoning         VulnerabilityType = "training_data_poisoning"
	ModelDOS                      VulnerabilityType = "model_dos"
	SupplyChainVulnerabilities    VulnerabilityType = "supply_chain_vulnerabilities"
	SensitiveInformationDisclosure VulnerabilityType = "sensitive_information_disclosure"
	InsecurePluginDesign          VulnerabilityType = "insecure_plugin_design"
	ExcessiveAgency               VulnerabilityType = "excessive_agency"
	Overreliance                  VulnerabilityType = "overreliance"
	ModelTheft                    VulnerabilityType = "model_theft"
)

// Report formats as string constants
const (
	JSONFormat     string = "json"
	JSONLFormat    string = "jsonl"
	CSVFormat      string = "csv"
	ExcelFormat    string = "excel"
	HTMLFormat     string = "html"
	PDFFormat      string = "pdf"
	MarkdownFormat string = "markdown"
	TextFormat     string = "text"
)

// TestCase represents a test case for OWASP compliance testing
type TestCase struct {
	ID                 string                      `json:"id"`
	Name               string                      `json:"name"`
	Description        string                      `json:"description"`
	VulnerabilityType  VulnerabilityType           `json:"vulnerability_type"`
	Severity           detection.SeverityLevel     `json:"severity"`
	Prompt             string                      `json:"prompt"`
	ExpectedBehavior   string                      `json:"expected_behavior"`
	DetectionCriteria  []detection.DetectionCriteria `json:"detection_criteria"`
	Tags               []string                    `json:"tags"`
	OWASPMapping       string                      `json:"owasp_mapping"`
	Metadata           map[string]interface{}      `json:"metadata"`
}

// TestResult represents the result of a test case execution
type TestResult struct {
	TestCase          *TestCase                   `json:"test_case"`
	Passed            bool                        `json:"passed"`
	Response          string                      `json:"response"`
	DetectionResults  []detection.DetectionResult `json:"detection_results"`
	ExecutionTime     time.Duration               `json:"execution_time"`
	Timestamp         time.Time                   `json:"timestamp"`
	Notes             string                      `json:"notes"`
}

// TestSuite represents a collection of test cases
type TestSuite struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	TestCases    []*TestCase  `json:"test_cases"`
	Results      []*TestResult `json:"results"`
	CreatedAt    time.Time    `json:"created_at"`
	CompletedAt  time.Time    `json:"completed_at"`
	Metadata     map[string]interface{} `json:"metadata"`
	Tags         []string     `json:"tags"`
}

// ReportOptions represents options for generating a report
type ReportOptions struct {
	Format      string                 `json:"format"`
	OutputPath  string                 `json:"output_path"`
	Title       string                 `json:"title"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Report represents a generated report
type Report struct {
	Title       string       `json:"title"`
	TestSuites  []*TestSuite `json:"test_suites"`
	Format      string       `json:"format"`
	GeneratedAt time.Time    `json:"generated_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// testCaseRegistry is a global registry for test cases
var testCaseRegistry = make(map[VulnerabilityType][]*TestCase)

// defaultTestFactory is the default test factory instance
var defaultTestFactory TestFactory

// RegisterTestCase registers a test case with the global registry
func RegisterTestCase(vulnerabilityType VulnerabilityType, testCase *TestCase) {
	testCaseRegistry[vulnerabilityType] = append(testCaseRegistry[vulnerabilityType], testCase)
}

// GetRegisteredTestCases returns all test cases registered for a specific vulnerability type
func GetRegisteredTestCases(vulnerabilityType VulnerabilityType) []*TestCase {
	return testCaseRegistry[vulnerabilityType]
}

// GetAllRegisteredTestCases returns all registered test cases
func GetAllRegisteredTestCases() map[VulnerabilityType][]*TestCase {
	return testCaseRegistry
}

// SetDefaultTestFactory sets the default test factory
func SetDefaultTestFactory(factory TestFactory) {
	defaultTestFactory = factory
}

// GetDefaultTestFactory returns the default test factory
func GetDefaultTestFactory() TestFactory {
	return defaultTestFactory
}
