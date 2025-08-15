package formats

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/perplext/LLMrecon/src/reporting/api"
)

// JSONFormatter is a formatter for JSON reports
type JSONFormatter struct {
	// pretty indicates whether to use pretty formatting
	pretty bool

// FormatReport formats a report and writes it to the given writer
func (f *JSONFormatter) FormatReport(results api.TestResults, writer io.Writer) error {
	data, err := f.Format(context.Background(), results, nil)
	if err != nil {
		return err
	}
	
	_, err = writer.Write(data)
	return err

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter(pretty bool) *JSONFormatter {
	return &JSONFormatter{
		pretty: pretty,
	}

// Format formats a report as JSON
func (f *JSONFormatter) Format(ctx context.Context, reportInterface interface{}, optionsInterface interface{}) ([]byte, error) {
	// We're using interface{} types now, so we don't need to check the specific types
	report := reportInterface
	
	// Just ensure we have a valid report object
	if report == nil {
		return nil, fmt.Errorf("report cannot be nil")
	}
	var data []byte
	var err error

	if f.pretty {
		data, err = json.MarshalIndent(report, "", "  ")
	} else {
		data, err = json.Marshal(report)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to marshal report to JSON: %w", err)
	}

	return data, nil

// GetFormat returns the format supported by this formatter
func (f *JSONFormatter) GetFormat() api.ReportFormat {
	return api.JSONFormat

// WriteToFile writes a report to a file
func (f *JSONFormatter) WriteToFile(ctx context.Context, reportInterface interface{}, optionsInterface interface{}, filePath string) error {
	// Format the report
	data, err := f.Format(ctx, reportInterface, optionsInterface)
	if err != nil {
		return fmt.Errorf("failed to format report: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write to file
	if err := os.WriteFile(filepath.Clean(filePath, data, 0600)); err != nil {
		return fmt.Errorf("failed to write report to file %s: %w", filePath, err)
	}

	return nil

// JSONLFormatter is a formatter for JSONL reports
type JSONLFormatter struct{}
// FormatReport formats a report and writes it to the given writer
func (f *JSONLFormatter) FormatReport(results api.TestResults, writer io.Writer) error {
	data, err := f.Format(context.Background(), results, nil)
	if err != nil {
		return err
	}
	
	_, err = writer.Write(data)
	return err

// NewJSONLFormatter creates a new JSONL formatter
func NewJSONLFormatter() *JSONLFormatter {
	return &JSONLFormatter{}

// Format formats a report as JSONL
func (f *JSONLFormatter) Format(ctx context.Context, reportInterface interface{}, optionsInterface interface{}) ([]byte, error) {
	// We're using interface{} types now, so we don't need to check the specific types
	report := reportInterface
	
	// Just ensure we have a valid report object
	if report == nil {
		return nil, fmt.Errorf("report cannot be nil")
	}
	
	// Simply marshal the report to JSON
	data, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal report to JSON: %w", err)
	}
	
	return data, nil

// GetFormat returns the format supported by this formatter
func (f *JSONLFormatter) GetFormat() api.ReportFormat {
	return api.JSONLFormat

// WriteToFile writes a report to a file
func (f *JSONLFormatter) WriteToFile(ctx context.Context, reportInterface interface{}, optionsInterface interface{}, filePath string) error {
	// Format the report
	data, err := f.Format(ctx, reportInterface, optionsInterface)
	if err != nil {
		return fmt.Errorf("failed to format report: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write to file
	if err := os.WriteFile(filepath.Clean(filePath, data, 0600)); err != nil {
		return fmt.Errorf("failed to write report to file %s: %w", filePath, err)
	}

