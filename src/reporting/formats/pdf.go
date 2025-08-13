package formats

import (
	"bytes"
	"context"
	"fmt"

	"github.com/jung-kurt/gofpdf"
	"github.com/perplext/LLMrecon/src/reporting/api"
)

// PDFFormatter is a formatter for PDF reports
type PDFFormatter struct {
	// customTemplate is the path to a custom template file
	customTemplate string
}

// NewPDFFormatter creates a new PDF formatter
func NewPDFFormatter(customTemplate string) *PDFFormatter {
	return &PDFFormatter{
		customTemplate: customTemplate,
	}
}

// FormatReport formats a report and writes it to the given writer
func (f *PDFFormatter) FormatReport(results api.TestResults, writer io.Writer) error {
	// Create a new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	
	// Set document properties
	pdf.SetTitle("Test Results Report", true)
	pdf.SetAuthor("LLMrecon Tool", true)
	pdf.SetCreator("LLMrecon Tool", true)
	
	// Add fonts
	pdf.SetFont("Arial", "", 10)
	
	// Add first page
	pdf.AddPage()
	
	// Generate the report content
	f.generateCoverPage(pdf, results)
	pdf.AddPage()
	f.generateResultsPage(pdf, results)
	
	// Generate the PDF
	return pdf.Output(writer)
}

// Format formats a report as PDF
func (f *PDFFormatter) Format(ctx context.Context, reportInterface interface{}, optionsInterface interface{}) ([]byte, error) {
	results, ok := reportInterface.(api.TestResults)
	if !ok {
		return nil, fmt.Errorf("expected api.TestResults, got %T", reportInterface)
	}
	
	// Create a buffer to hold the PDF data
	buf := &bytes.Buffer{}
	
	// Use the FormatReport method to write to the buffer
	err := f.FormatReport(results, buf)
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// GetFormat returns the format supported by this formatter
func (f *PDFFormatter) GetFormat() api.ReportFormat {
	return api.PDFFormat
}

// WriteToFile writes a report to a file
func (f *PDFFormatter) WriteToFile(ctx context.Context, reportInterface interface{}, optionsInterface interface{}, filePath string) error {
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

// generateCoverPage generates the cover page of the report
func (f *PDFFormatter) generateCoverPage(pdf *gofpdf.Fpdf, results api.TestResults) {
	// Set up the cover page
	pdf.SetFont("Arial", "B", 24)
	pdf.SetTextColor(0, 0, 0)
	
	// Add title
	pdf.Cell(0, 10, "Test Results Report")
	pdf.Ln(20)
	
	// Add generated date
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(0, 10, fmt.Sprintf("Generated: %s", time.Now().Format(time.RFC3339)))
	pdf.Ln(20)
	
	// Add summary statistics
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, fmt.Sprintf("Total Tests: %d", len(results)))
	pdf.Ln(10)
	
	// Add footer
	pdf.SetY(-30)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 10, "LLMrecon Tool")
	pdf.Ln(5)
	pdf.Cell(0, 10, fmt.Sprintf("© %d", time.Now().Year()))
}

// generateResultsPage generates the results page of the report
func (f *PDFFormatter) generateResultsPage(pdf *gofpdf.Fpdf, results api.TestResults) {
	// Set up the results page
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 10, "Test Results")
	pdf.Ln(15)
	
	// Create results table
	f.addResultsTable(pdf, results)
}

// addResultsTable adds a results table to the PDF
func (f *PDFFormatter) addResultsTable(pdf *gofpdf.Fpdf, results api.TestResults) {
	// Set up table
	pdf.SetFont("Arial", "B", 10)
	
	// Define column widths
	colWidths := []float64{10, 50, 30, 30, 30, 40}
	
	// Create table headers
	pdf.SetFillColor(200, 200, 200)
	pdf.Cell(colWidths[0], 8, "#")
	pdf.Cell(colWidths[1], 8, "Name")
	pdf.Cell(colWidths[2], 8, "ID")
	pdf.Cell(colWidths[3], 8, "Status")
	pdf.Cell(colWidths[4], 8, "Severity")
	pdf.Cell(colWidths[5], 8, "Category")
	pdf.Ln(-1)
	
	// Add table rows
	pdf.SetFont("Arial", "", 10)
	
	for i, result := range results {
		// Set background color based on status
		if result.Status == "failed" {
			pdf.SetFillColor(255, 200, 200)
		} else if result.Status == "passed" {
			pdf.SetFillColor(200, 255, 200)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		
		pdf.Cell(colWidths[0], 8, fmt.Sprintf("%d", i+1))
		pdf.Cell(colWidths[1], 8, f.truncateString(result.Name, 30))
		pdf.Cell(colWidths[2], 8, f.truncateString(result.ID, 15))
		pdf.Cell(colWidths[3], 8, result.Status)
		pdf.Cell(colWidths[4], 8, string(result.Severity))
		pdf.Cell(colWidths[5], 8, f.truncateString(result.Category, 20))
		pdf.Ln(-1)
		
		// Add details if available
		if result.Description != "" || result.Details != "" {
			pdf.SetFillColor(240, 240, 240)
			
			detailsText := ""
			if result.Description != "" {
				detailsText += "Description: " + result.Description + "\n"
			}
			if result.Details != "" {
				detailsText += "Details: " + result.Details
			}
			
			pdf.MultiCell(0, 8, detailsText, "", "", false)
			pdf.Ln(4)
		}
	}
}

// truncateString truncates a string to the specified length and adds "..." if truncated
func (f *PDFFormatter) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
