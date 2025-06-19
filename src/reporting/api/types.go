// Package api provides common types and interfaces for the reporting system
package api

import (
	"io"
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
	// GetFormat returns the format of the report
	GetFormat() ReportFormat
	// FormatReport formats a report and writes it to the given writer
	FormatReport(results TestResults, writer io.Writer) error
}

// FormatterCreator is a function that creates a formatter
type FormatterCreator func(options map[string]interface{}) (ReportFormatter, error)

// SeverityLevelMapping maps string representations to SeverityLevel constants
var SeverityLevelMapping = map[string]SeverityLevel{
	"critical": Critical,
	"high":     High,
	"medium":   Medium,
	"low":      Low,
	"info":     Info,
}
