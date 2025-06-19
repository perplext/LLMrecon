// Package common provides common types and interfaces for the reporting system
package common

import (
	"context"
)

// ReportFormat defines the format of a report
type ReportFormat string

// Report formats
const (
	JSONFormat     ReportFormat = "json"
	JSONLFormat    ReportFormat = "jsonl"
	CSVFormat      ReportFormat = "csv"
	ExcelFormat    ReportFormat = "excel"
	TextFormat     ReportFormat = "text"
	MarkdownFormat ReportFormat = "markdown"
	PDFFormat      ReportFormat = "pdf"
	HTMLFormat     ReportFormat = "html"
)

// SeverityLevel defines the severity level of a test result
type SeverityLevel string

// Severity levels
const (
	Critical SeverityLevel = "critical"
	High     SeverityLevel = "high"
	Medium   SeverityLevel = "medium"
	Low      SeverityLevel = "low"
	Info     SeverityLevel = "info"
)

// TestResult represents the result of a security test
type TestResult struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Severity    SeverityLevel `json:"severity"`
	Category    string       `json:"category"`
	Status      string       `json:"status"`
	Details     string       `json:"details,omitempty"`
	RawData     interface{}  `json:"raw_data,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TestResults is a collection of test results
type TestResults []*TestResult

// ReportFormatter is the interface for report formatters
type ReportFormatter interface {
	// GetFormat returns the format of the formatter
	GetFormat() ReportFormat
	// Format formats a report
	Format(ctx context.Context, report interface{}, options interface{}) ([]byte, error)
	// WriteToFile writes a report to a file
	WriteToFile(ctx context.Context, report interface{}, options interface{}, filePath string) error
}

// ReportGenerator is the interface for report generators
type ReportGenerator interface {
	// GenerateReport generates a report from test results
	GenerateReport(ctx context.Context, testSuites interface{}, options interface{}) (interface{}, error)
	// RegisterFormatter registers a formatter for a specific format
	RegisterFormatter(formatter ReportFormatter)
	// GetFormatter returns a formatter for a specific format
	GetFormatter(format ReportFormat) (ReportFormatter, bool)
	// ListFormats returns a list of supported formats
	ListFormats() []ReportFormat
}
