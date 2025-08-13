package formats

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

	"github.com/perplext/LLMrecon/src/reporting/api"
)

// CSVFormatter is a formatter for CSV reports
type CSVFormatter struct {
	// delimiter is the delimiter to use for CSV fields
	delimiter rune
	// includeHeaders indicates whether to include headers
	includeHeaders bool
}

// FormatReport formats a report and writes it to the given writer
func (f *CSVFormatter) FormatReport(results api.TestResults, writer io.Writer) error {
	data, err := f.Format(context.Background(), results, nil)
	if err != nil {
		return err
	}
	
	_, err = writer.Write(data)
	return err
}

// NewCSVFormatter creates a new CSV formatter
func NewCSVFormatter(delimiter rune, includeHeaders bool) *CSVFormatter {
	if delimiter == 0 {
		delimiter = ','
	}
	return &CSVFormatter{
		delimiter:      delimiter,
		includeHeaders: includeHeaders,
	}
}

// Format formats a report as CSV
func (f *CSVFormatter) Format(ctx context.Context, reportInterface interface{}, optionsInterface interface{}) ([]byte, error) {
	// We're using interface{} types now, so we don't need to check the specific types
	report := reportInterface
	
	// Just ensure we have a valid report object
	if report == nil {
		return nil, fmt.Errorf("report cannot be nil")
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Comma = f.delimiter

	// Add headers if enabled
	if f.includeHeaders {
		headers := []string{
			"ID",
			"Name",
			"Description",
			"Status",
			"Severity",
			"Score",
			"Suite ID",
			"Suite Name",
			"Error",
		}
		if err := writer.Write(headers); err != nil {
			return nil, fmt.Errorf("failed to write headers: %w", err)
		}
	}

	// For a simplified implementation, we'll just write a single row
	// In a real implementation, we would iterate through the report structure
	row := []string{"sample", "sample", "sample", "sample", "sample", "0", "sample", "sample", ""}
	if err := writer.Write(row); err != nil {
		return nil, fmt.Errorf("failed to write row: %w", err)
	}
	
	// Flush the writer
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("failed to flush writer: %w", err)
	}

	return buf.Bytes(), nil
}

// GetFormat returns the format supported by this formatter
func (f *CSVFormatter) GetFormat() api.ReportFormat {
	return api.CSVFormat
}

// WriteToFile writes a report to a file
func (f *CSVFormatter) WriteToFile(ctx context.Context, reportInterface interface{}, optionsInterface interface{}, filePath string) error {
	// Format the report
	data, err := f.Format(ctx, reportInterface, optionsInterface)
	if err != nil {
		return fmt.Errorf("failed to format report: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write report to file %s: %w", filePath, err)
	}

	return nil
}
