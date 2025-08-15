package formats

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/reporting/api"
)

// MarkdownFormatter is a formatter for Markdown reports
type MarkdownFormatter struct {
	// includeRawData indicates whether to include raw data in the report
	includeRawData bool

// NewMarkdownFormatter creates a new Markdown formatter
func NewMarkdownFormatter(includeRawData bool) *MarkdownFormatter {
	return &MarkdownFormatter{
		includeRawData: includeRawData,
	}

// FormatReport formats a report and writes it to the given writer
func (f *MarkdownFormatter) FormatReport(results api.TestResults, writer io.Writer) error {
	// Create a simple markdown report
	buf := &bytes.Buffer{}

	// Add title
	buf.WriteString("# Test Results Report\n\n")
	buf.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format(time.RFC3339)))

	// Add test results
	buf.WriteString("## Results\n\n")

	// Create a table for results
	buf.WriteString("| ID | Name | Severity | Status | Category |\n")
	buf.WriteString("|---|---|---|---|---|\n")

	for _, result := range results {
		buf.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n", 
			result.ID, result.Name, result.Severity, result.Status, result.Category))
	}

	// Add detailed information if requested
	if f.includeRawData {
		buf.WriteString("\n## Detailed Results\n\n")
		
		for _, result := range results {
			buf.WriteString(fmt.Sprintf("### %s\n\n", result.Name))
			buf.WriteString(fmt.Sprintf("**ID:** %s\n\n", result.ID))
			buf.WriteString(fmt.Sprintf("**Severity:** %s\n\n", result.Severity))
			buf.WriteString(fmt.Sprintf("**Status:** %s\n\n", result.Status))
			buf.WriteString(fmt.Sprintf("**Category:** %s\n\n", result.Category))
			
			if result.Description != "" {
				buf.WriteString(fmt.Sprintf("**Description:** %s\n\n", result.Description))
			}
			
			if result.Details != "" {
				buf.WriteString(fmt.Sprintf("**Details:**\n\n```\n%s\n```\n\n", result.Details))
			}
			
			if result.RawData != nil {
				buf.WriteString(fmt.Sprintf("**Raw Data:** %v\n\n", result.RawData))
			}
			
			// Add horizontal rule between results
			buf.WriteString("---\n\n")
		}
	}

	// Add summary section
	buf.WriteString("## Summary\n\n")
	buf.WriteString(fmt.Sprintf("- **Total Tests:** %d\n", len(results)))
	
	// Count passed, failed, and other tests
	passed := 0
	failed := 0
	other := 0
	
	for _, result := range results {
		switch strings.ToLower(result.Status) {
		case "passed", "pass":
			passed++
		case "failed", "fail":
			failed++
		default:
			other++
		}
	}
	
	buf.WriteString(fmt.Sprintf("- **Passed:** %d\n", passed))
	buf.WriteString(fmt.Sprintf("- **Failed:** %d\n", failed))
	if other > 0 {
		buf.WriteString(fmt.Sprintf("- **Other:** %d\n", other))
	}
	
	// Calculate pass rate if there are any tests
	if len(results) > 0 {
		passRate := float64(passed) / float64(len(results)) * 100
		buf.WriteString(fmt.Sprintf("- **Pass Rate:** %.2f%%\n", passRate))
	}

	// Write to the provided writer
	_, err := writer.Write(buf.Bytes())
	return err

// Format formats a report as Markdown
func (f *MarkdownFormatter) Format(ctx context.Context, reportInterface interface{}, optionsInterface interface{}) ([]byte, error) {
	results, ok := reportInterface.(api.TestResults)
	if !ok {
		return nil, fmt.Errorf("expected api.TestResults, got %T", reportInterface)
	}
	
	// Create a buffer to hold the markdown data
	buf := &bytes.Buffer{}
	
	// Use the FormatReport method to write to the buffer
	err := f.FormatReport(results, buf)
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil

// GetFormat returns the format supported by this formatter
func (f *MarkdownFormatter) GetFormat() api.ReportFormat {
	return api.MarkdownFormat

// WriteToFile writes a report to a file
func (f *MarkdownFormatter) WriteToFile(ctx context.Context, reportInterface interface{}, optionsInterface interface{}, filePath string) error {
	results, ok := reportInterface.(api.TestResults)
	if !ok {
		return fmt.Errorf("expected api.TestResults, got %T", reportInterface)
	}
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer func() { if err := file.Close(); err != nil { fmt.Printf("Failed to close: %v\n", err) } }()

	// Format and write the report
	if err := f.FormatReport(results, file); err != nil {
		return fmt.Errorf("failed to write report to file %s: %w", filePath, err)
	}

