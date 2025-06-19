// Package types provides common types for the reporting system
package types

import (
	"context"
	"time"
)

// ReportFormat represents the format of a report
type ReportFormat string

// Supported report formats
const (
	TextFormat    ReportFormat = "txt"
	MarkdownFormat ReportFormat = "md"
	JSONFormat    ReportFormat = "json"
	JSONLFormat   ReportFormat = "jsonl"
	CSVFormat     ReportFormat = "csv"
	ExcelFormat   ReportFormat = "xlsx"
	PDFFormat     ReportFormat = "pdf"
	HTMLFormat    ReportFormat = "html"
)

// SeverityLevel represents the severity level of a test result
type SeverityLevel string

// Supported severity levels
const (
	CriticalSeverity SeverityLevel = "critical"
	HighSeverity     SeverityLevel = "high"
	MediumSeverity   SeverityLevel = "medium"
	LowSeverity      SeverityLevel = "low"
	InfoSeverity     SeverityLevel = "info"
)

// Report represents a test report
type Report struct {
	// ID is the unique identifier for the report
	ID string `json:"id"`
	// Title is the title of the report
	Title string `json:"title"`
	// Description is the description of the report
	Description string `json:"description"`
	// CreatedAt is the time the report was created
	CreatedAt time.Time `json:"created_at"`
	// GeneratedAt is the time the report was generated
	GeneratedAt time.Time `json:"generated_at"`
	// Format is the format of the report
	Format ReportFormat `json:"format"`
	// Results is the list of test results
	Results []*TestResult `json:"results"`
	// TestSuites is the list of test suites
	TestSuites []*TestSuite `json:"test_suites"`
	// Summary is the summary of the report
	Summary *ReportSummary `json:"summary"`
	// Metadata is additional metadata for the report
	Metadata map[string]interface{} `json:"metadata"`
}

// ReportSummary represents a summary of test results
type ReportSummary struct {
	// TotalTests is the total number of tests
	TotalTests int `json:"total_tests"`
	// PassedTests is the number of passed tests
	PassedTests int `json:"passed_tests"`
	// FailedTests is the number of failed tests
	FailedTests int `json:"failed_tests"`
	// ErrorTests is the number of tests with errors
	ErrorTests int `json:"error_tests"`
	// SkippedTests is the number of skipped tests
	SkippedTests int `json:"skipped_tests"`
	// PendingTests is the number of pending tests
	PendingTests int `json:"pending_tests"`
	// PassRate is the pass rate as a percentage
	PassRate float64 `json:"pass_rate"`
	// AverageScore is the average score of all tests
	AverageScore float64 `json:"average_score"`
	// AverageDuration is the average duration of all tests
	AverageDuration time.Duration `json:"average_duration"`
	// TotalDuration is the total duration of all tests
	TotalDuration time.Duration `json:"total_duration"`
	// SeverityCounts is the count of tests by severity
	SeverityCounts map[SeverityLevel]int `json:"severity_counts"`
	// SeverityBreakdown is a breakdown of test results by severity
	SeverityBreakdown map[SeverityLevel]int `json:"severity_breakdown"`
	// CategoryBreakdown is a breakdown of test results by category
	CategoryBreakdown map[string]int `json:"category_breakdown"`
	// Score is the overall score (0-100)
	Score float64 `json:"score"`
}

// TestResult represents the result of a test
type TestResult struct {
	// ID is the unique identifier for the test result
	ID string `json:"id"`
	// TestID is the ID of the test
	TestID string `json:"test_id"`
	// Name is the name of the test
	Name string `json:"name"`
	// Description is the description of the test
	Description string `json:"description"`
	// Category is the category of the test
	Category string `json:"category"`
	// Severity is the severity of the test
	Severity SeverityLevel `json:"severity"`
	// Status is the status of the test
	Status TestStatus `json:"status"`
	// Passed indicates whether the test passed
	Passed bool `json:"passed"`
	// Score is the score of the test (0-100)
	Score int `json:"score"`
	// StartTime is the time the test started
	StartTime time.Time `json:"start_time"`
	// EndTime is the time the test ended
	EndTime time.Time `json:"end_time"`
	// Duration is the duration of the test
	Duration time.Duration `json:"duration"`
	// Error is any error that occurred during the test
	Error string `json:"error,omitempty"`
	// Input is the input to the test
	Input string `json:"input,omitempty"`
	// ExpectedOutput is the expected output of the test
	ExpectedOutput string `json:"expected_output,omitempty"`
	// ActualOutput is the actual output of the test
	ActualOutput string `json:"actual_output,omitempty"`
	// Message is a message associated with the test result
	Message string `json:"message"`
	// Tags is the list of tags for the test
	Tags []string `json:"tags,omitempty"`
	// ComplianceMappings is the list of compliance mappings for the test
	ComplianceMappings []ComplianceMapping `json:"compliance_mappings,omitempty"`
	// Details contains additional details about the test result
	Details map[string]interface{} `json:"details"`
	// Timestamp is the time the test was run
	Timestamp time.Time `json:"timestamp"`
}

// TestSuite represents a collection of test results
type TestSuite struct {
	// ID is the unique identifier for the test suite
	ID string `json:"id"`
	// Name is the name of the test suite
	Name string `json:"name"`
	// Description is the description of the test suite
	Description string `json:"description,omitempty"`
	// StartTime is the time the test suite started
	StartTime time.Time `json:"start_time"`
	// EndTime is the time the test suite ended
	EndTime time.Time `json:"end_time"`
	// Duration is the duration of the test suite
	Duration time.Duration `json:"duration"`
	// Results is the list of test results
	Results []*TestResult `json:"results"`
	// Metadata is additional metadata for the test suite
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ReportOptions represents options for generating a report
type ReportOptions struct {
	// Title is the title of the report
	Title string `json:"title"`
	// Description is the description of the report
	Description string `json:"description"`
	// Format is the format of the report
	Format ReportFormat `json:"format"`
	// OutputPath is the path to write the report to
	OutputPath string `json:"output_path"`
	// IncludePassedTests indicates whether to include passed tests in the report
	IncludePassedTests bool `json:"include_passed_tests"`
	// IncludeDetails indicates whether to include test details in the report
	IncludeDetails bool `json:"include_details"`
	// Metadata is additional metadata for the report
	Metadata map[string]interface{} `json:"metadata"`
}

// TestResultGenerator is the interface for generating test results
type TestResultGenerator interface {
	// GenerateTestResults generates test results
	GenerateTestResults(results []*TestResult, options *ReportOptions) (*Report, error)
}

// ReportFormatter is the interface for formatting reports
type ReportFormatter interface {
	// Format formats a report
	Format(ctx context.Context, report *Report, options *ReportOptions) ([]byte, error)
	// GetFormat returns the format supported by this formatter
	GetFormat() ReportFormat
}

// TestStatus represents the status of a test
type TestStatus string

// Supported test statuses
const (
	PassStatus    TestStatus = "pass"
	FailStatus    TestStatus = "fail"
	ErrorStatus   TestStatus = "error"
	SkippedStatus TestStatus = "skipped"
	PendingStatus TestStatus = "pending"
)

// ComplianceFramework represents a compliance framework
type ComplianceFramework string

// Supported compliance frameworks
const (
	OWASPFramework ComplianceFramework = "owasp-top-10-llm"
	ISOFramework   ComplianceFramework = "iso-iec-42001"
	NISTFramework  ComplianceFramework = "nist-ai-risk"
	CustomFramework ComplianceFramework = "custom"
)

// ComplianceMapping represents a mapping to a compliance framework
type ComplianceMapping struct {
	// Framework is the compliance framework
	Framework ComplianceFramework `json:"framework"`
	// ID is the ID of the item within the framework
	ID string `json:"id"`
	// Name is the name of the item within the framework
	Name string `json:"name"`
	// Description is the description of the item within the framework
	Description string `json:"description,omitempty"`
	// URL is the URL to the item within the framework
	URL string `json:"url,omitempty"`
}

// ComplianceMappingProvider is the interface for compliance mapping providers
type ComplianceMappingProvider interface {
	// GetMappings returns compliance mappings for a test result
	GetMappings(ctx context.Context, testResult *TestResult) ([]ComplianceMapping, error)
	// GetFrameworks returns a list of supported compliance frameworks
	GetFrameworks() []ComplianceFramework
}

// FilterFunc is a function that filters test results
type FilterFunc func(result *TestResult) bool

// SortFunc is a function that sorts test results
type SortFunc func(results []*TestResult) []*TestResult

// ReportGenerator is the interface for report generators
type ReportGenerator interface {
	// GenerateReport generates a report from test results
	GenerateReport(ctx context.Context, testSuites []*TestSuite, options *ReportOptions) (*Report, error)
	// RegisterFormatter registers a formatter for a specific format
	RegisterFormatter(formatter ReportFormatter)
	// GetFormatter returns a formatter for a specific format
	GetFormatter(format ReportFormat) (ReportFormatter, bool)
	// ListFormats returns a list of supported formats
	ListFormats() []ReportFormat
}
