package formats

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/reporting/api"
)

// TextFormatter is a formatter for plain text reports
type TextFormatter struct {
	// detailed indicates whether to include detailed information
	detailed bool
}

// NewTextFormatter creates a new text formatter
func NewTextFormatter(detailed bool) *TextFormatter {
	return &TextFormatter{
		detailed: detailed,
	}
}

// FormatReport formats a report and writes it to the given writer
func (f *TextFormatter) FormatReport(results api.TestResults, writer io.Writer) error {
	// Create a simple text report
	buf := &bytes.Buffer{}

	// Add title
	buf.WriteString("Test Results Report\n")
	buf.WriteString(strings.Repeat("=", 20) + "\n")
	buf.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))

	// Add test results
	buf.WriteString("RESULTS\n")
	buf.WriteString("-------\n")

	for i, result := range results {
		buf.WriteString(fmt.Sprintf("%d. %s (ID: %s)\n", i+1, result.Name, result.ID))
		buf.WriteString(fmt.Sprintf("   Status: %s\n", result.Status))
		buf.WriteString(fmt.Sprintf("   Severity: %s\n", result.Severity))
		buf.WriteString(fmt.Sprintf("   Category: %s\n", result.Category))
		
		if result.Description != "" {
			buf.WriteString(fmt.Sprintf("   Description: %s\n", result.Description))
		}
		
		if result.Details != "" {
			buf.WriteString(fmt.Sprintf("   Details: %s\n", result.Details))
		}
		
		if f.detailed && result.RawData != nil {
			buf.WriteString(fmt.Sprintf("   Raw Data: %v\n", result.RawData))
		}
		
		buf.WriteString("\n")
	}

	// Write to the provided writer
	_, err := writer.Write(buf.Bytes())
	return err
}

// Format formats a report as plain text
func (f *TextFormatter) Format(ctx context.Context, reportInterface interface{}, optionsInterface interface{}) ([]byte, error) {
	results, ok := reportInterface.(api.TestResults)
	if !ok {
		return nil, fmt.Errorf("expected api.TestResults, got %T", reportInterface)
	}
	
	// Create a buffer to hold the text data
	buf := &bytes.Buffer{}
	
	// Use the FormatReport method to write to the buffer
	err := f.FormatReport(results, buf)
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// GetFormat returns the format supported by this formatter
func (f *TextFormatter) GetFormat() api.ReportFormat {
	return api.TextFormat
}

// WriteToFile writes a report to a file
func (f *TextFormatter) WriteToFile(ctx context.Context, reportInterface interface{}, optionsInterface interface{}, filePath string) error {
	results, ok := reportInterface.(api.TestResults)
	if !ok {
		return fmt.Errorf("expected api.TestResults, got %T", reportInterface)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	// Format and write the report
	if err := f.FormatReport(results, file); err != nil {
		return fmt.Errorf("failed to write report to file %s: %w", filePath, err)
	}

	return nil
}


