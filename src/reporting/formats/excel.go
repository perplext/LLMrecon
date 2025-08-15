package formats

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/reporting/api"
	"github.com/xuri/excelize/v2"
)

// ExcelFormatter is a formatter for Excel reports
type ExcelFormatter struct {
	// includeRawData indicates whether to include raw data in the report
	includeRawData bool

// NewExcelFormatter creates a new Excel formatter
func NewExcelFormatter(includeRawData bool) *ExcelFormatter {
	return &ExcelFormatter{
		includeRawData: includeRawData,
	}

// FormatReport formats a report and writes it to the given writer
func (f *ExcelFormatter) FormatReport(results api.TestResults, writer io.Writer) error {
	// Create a new Excel file
	excel := excelize.NewFile()

	// Create summary sheet
	summarySheet := "Summary"
	excel.SetSheetName("Sheet1", summarySheet)

	// Add title
	excel.SetCellValue(summarySheet, "A1", "Test Results Report")
	excel.MergeCell(summarySheet, "A1", "D1")
	
	// Add generation timestamp
	excel.SetCellValue(summarySheet, "A2", fmt.Sprintf("Generated: %s", time.Now().Format(time.RFC3339)))
	excel.MergeCell(summarySheet, "A2", "D2")

	// Add headers
	headers := []string{"ID", "Name", "Severity", "Status", "Category"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c4", 'A'+i)
		excel.SetCellValue(summarySheet, cell, header)
	}

	// Add test results
	row := 5
	for _, result := range results {
		excel.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), result.ID)
		excel.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), result.Name)
		excel.SetCellValue(summarySheet, fmt.Sprintf("C%d", row), string(result.Severity))
		excel.SetCellValue(summarySheet, fmt.Sprintf("D%d", row), result.Status)
		excel.SetCellValue(summarySheet, fmt.Sprintf("E%d", row), result.Category)
		row++
	}

	// Create a details sheet if we have detailed information
	if f.includeRawData {
		detailsSheet := "Details"
		_, err := excel.NewSheet(detailsSheet)
		if err != nil {
			return fmt.Errorf("failed to create details sheet: %w", err)
		}

		// Add headers for details
		detailHeaders := []string{"ID", "Name", "Description", "Details"}
		for i, header := range detailHeaders {
			cell := fmt.Sprintf("%c1", 'A'+i)
			excel.SetCellValue(detailsSheet, cell, header)
		}

		// Add detailed information
		detailRow := 2
		for _, result := range results {
			excel.SetCellValue(detailsSheet, fmt.Sprintf("A%d", detailRow), result.ID)
			excel.SetCellValue(detailsSheet, fmt.Sprintf("B%d", detailRow), result.Name)
			excel.SetCellValue(detailsSheet, fmt.Sprintf("C%d", detailRow), result.Description)
			excel.SetCellValue(detailsSheet, fmt.Sprintf("D%d", detailRow), result.Details)
			detailRow++
		}
	}

	// Write to the provided writer
	return excel.Write(writer)

// Format formats a report as Excel
func (f *ExcelFormatter) Format(ctx context.Context, reportInterface interface{}, optionsInterface interface{}) ([]byte, error) {
	results, ok := reportInterface.(api.TestResults)
	if !ok {
		return nil, fmt.Errorf("expected api.TestResults, got %T", reportInterface)
	}

	// Create a buffer to hold the Excel data
	buf := &bytes.Buffer{}
	
	// Use the FormatReport method to write to the buffer
	err := f.FormatReport(results, buf)
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil

// GetFormat returns the format supported by this formatter
func (f *ExcelFormatter) GetFormat() api.ReportFormat {
	return api.ExcelFormat

// WriteToFile writes a report to a file
func (f *ExcelFormatter) WriteToFile(ctx context.Context, reportInterface interface{}, optionsInterface interface{}, filePath string) error {
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

	return nil

// sanitizeSheetName sanitizes a sheet name to be valid for Excel
func sanitizeSheetName(name string) string {
	// Excel sheet names have a maximum length of 31 characters
	if len(name) > 31 {
		name = name[:31]
	}
	
	// Excel sheet names cannot contain these characters: : \ / ? * [ ]
	invalidChars := []string{":", "\\", "/", "?", "*", "[", "]"}
	for _, char := range invalidChars {
		name = strings.ReplaceAll(name, char, "_")
	}
	
	// Excel sheet names cannot be empty
	if name == "" {
		name = "Sheet"
	}
	
