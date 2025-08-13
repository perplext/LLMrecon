// Package reporting provides a comprehensive reporting system for LLM test results.
package reporting

import (
	"context"
	"time"

	"github.com/perplext/LLMrecon/src/reporting/common"
)

// ReportFormat represents the format of a report
type ReportFormat = common.ReportFormat

// Supported report formats
const (
	TextFormat    = common.TextFormat
	MarkdownFormat = common.MarkdownFormat
	JSONFormat    = common.JSONFormat
	JSONLFormat   = common.JSONLFormat
	CSVFormat     = common.CSVFormat
	ExcelFormat   = common.ExcelFormat
	PDFFormat     = common.PDFFormat
	HTMLFormat    = common.HTMLFormat
)

// SeverityLevelMapping maps string representations to SeverityLevel constants
var SeverityLevelMapping = map[string]common.SeverityLevel{
	"critical": common.Critical,
	"high":     common.High,
	"medium":   common.Medium,
	"low":      common.Low,
	"info":     common.Info,
}

// SeverityLevel represents the severity level of a test
type SeverityLevel = common.SeverityLevel

// Supported severity levels
const (
	CriticalSeverity = common.Critical
	HighSeverity     = common.High
	MediumSeverity   = common.Medium
	LowSeverity      = common.Low
	InfoSeverity     = common.Info
)

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

// TestResult represents a single test result
type TestResult struct {
	// ID is the unique identifier for the test result
	ID string `json:"id"`
	// Name is the name of the test
	Name string `json:"name"`
	// Description is the description of the test
	Description string `json:"description,omitempty"`
	// Status is the status of the test
	Status TestStatus `json:"status"`
	// Severity is the severity level of the test
	Severity common.SeverityLevel `json:"severity"`
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
	// ComplianceMappings is the list of compliance mappings for the test
	ComplianceMappings []ComplianceMapping `json:"compliance_mappings,omitempty"`
	// Tags is the list of tags for the test
	Tags []string `json:"tags,omitempty"`
	// Metadata is additional metadata for the test
	Metadata map[string]interface{} `json:"metadata,omitempty"`
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

// ReportSummary represents a summary of a report
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
	SeverityCounts map[common.SeverityLevel]int `json:"severity_counts"`
	// TagCounts is the count of tests by tag
	TagCounts map[string]int `json:"tag_counts"`
}

// Report represents a complete test report
type Report struct {
	// ID is the unique identifier for the report
	ID string `json:"id"`
	// Title is the title of the report
	Title string `json:"title"`
	// Description is the description of the report
	Description string `json:"description,omitempty"`
	// GeneratedAt is the time the report was generated
	GeneratedAt time.Time `json:"generated_at"`
	// Summary is the summary of the report
	Summary ReportSummary `json:"summary"`
	// TestSuites is the list of test suites
	TestSuites []*TestSuite `json:"test_suites"`
	// Metadata is additional metadata for the report
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ReportOptions represents options for generating a report
type ReportOptions struct {
	// Format is the format of the report
	Format common.ReportFormat `json:"format"`
	// Title is the title of the report
	Title string `json:"title"`
	// Description is the description of the report
	Description string `json:"description,omitempty"`
	// IncludePassedTests indicates whether to include passed tests
	IncludePassedTests bool `json:"include_passed_tests"`
	// IncludeSkippedTests indicates whether to include skipped tests
	IncludeSkippedTests bool `json:"include_skipped_tests"`
	// IncludePendingTests indicates whether to include pending tests
	IncludePendingTests bool `json:"include_pending_tests"`
	// MinimumSeverity is the minimum severity level to include
	MinimumSeverity common.SeverityLevel `json:"minimum_severity"`
	// IncludeTags is the list of tags to include
	IncludeTags []string `json:"include_tags,omitempty"`
	// ExcludeTags is the list of tags to exclude
	ExcludeTags []string `json:"exclude_tags,omitempty"`
	// SortBy is the field to sort by
	SortBy string `json:"sort_by,omitempty"`
	// SortOrder is the sort order (asc or desc)
	SortOrder string `json:"sort_order,omitempty"`
	// TemplatePath is the path to a custom template
	TemplatePath string `json:"template_path,omitempty"`
	// OutputPath is the path to write the report to
	OutputPath string `json:"output_path,omitempty"`
	// Metadata is additional metadata for the report
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ReportGenerator interface defines methods for report generation
type ReportGenerator interface {
	// GenerateReport generates a report from test results
	GenerateReport(ctx context.Context, testSuites []*TestSuite, options *ReportOptions) (*Report, error)
	// RegisterFormatter registers a formatter for a specific format
	RegisterFormatter(formatter common.ReportFormatter)
	// GetFormatter returns a formatter for a specific format
	GetFormatter(format common.ReportFormat) (common.ReportFormatter, bool)
	// ListFormats returns a list of supported formats
	ListFormats() []common.ReportFormat
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
