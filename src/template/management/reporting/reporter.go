// Package reporting provides functionality for generating reports of template execution results.
package reporting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"sort"
	"time"

	"github.com/perplext/LLMrecon/src/template/management/interfaces"
	"github.com/xuri/excelize/v2"
	"gopkg.in/yaml.v3"
)

// ReportFormat represents the format of a report
type ReportFormat string

const (
	// JSONFormat represents a JSON report
	JSONFormat ReportFormat = "json"
	// YAMLFormat represents a YAML report
	YAMLFormat ReportFormat = "yaml"
	// HTMLFormat represents an HTML report
	HTMLFormat ReportFormat = "html"
	// PDFFormat represents a PDF report
	PDFFormat ReportFormat = "pdf"
	// ExcelFormat represents an Excel report
	ExcelFormat ReportFormat = "excel"
	// CSVFormat represents a CSV report
	CSVFormat ReportFormat = "csv"
)

// TemplateReporter is responsible for generating reports of template execution results
type TemplateReporter struct {
	// htmlTemplate is the HTML template for reports
	htmlTemplate *template.Template
	// customFormatters is a map of report format to custom formatter function
	customFormatters map[ReportFormat]ReportFormatter
}

// ReportFormatter is a function that formats a report
type ReportFormatter func(results []*interfaces.TemplateResult) ([]byte, error)

// NewTemplateReporter creates a new template reporter
func NewTemplateReporter() (*TemplateReporter, error) {
	// Parse HTML template
	htmlTemplate, err := template.New("report").Parse(defaultHTMLTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML template: %w", err)
	}

	return &TemplateReporter{
		htmlTemplate:     htmlTemplate,
		customFormatters: make(map[ReportFormat]ReportFormatter),
	}, nil
}

// RegisterCustomFormatter registers a custom formatter for a specific report format
func (r *TemplateReporter) RegisterCustomFormatter(format ReportFormat, formatter ReportFormatter) {
	r.customFormatters[format] = formatter
}

// SetHTMLTemplate sets the HTML template for reports
func (r *TemplateReporter) SetHTMLTemplate(htmlTemplate string) error {
	// Parse HTML template
	tmpl, err := template.New("report").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	r.htmlTemplate = tmpl
	return nil
}

// GenerateReport generates a report for template execution results
func (r *TemplateReporter) GenerateReport(results []*interfaces.TemplateResult, format string) ([]byte, error) {
	// Check if there's a custom formatter for this format
	if formatter, ok := r.customFormatters[ReportFormat(format)]; ok {
		return formatter(results)
	}

	// Use built-in formatters based on format
	switch ReportFormat(format) {
	case JSONFormat:
		return r.generateJSONReport(results)
	case YAMLFormat:
		return r.generateYAMLReport(results)
	case HTMLFormat:
		return r.generateHTMLReport(results)
	case ExcelFormat:
		return r.generateExcelReport(results)
	case CSVFormat:
		return r.generateCSVReport(results)
	case PDFFormat:
		return r.generatePDFReport(results)
	default:
		return nil, fmt.Errorf("unsupported report format: %s", format)
	}
}

// generateJSONReport generates a JSON report
func (r *TemplateReporter) generateJSONReport(results []*interfaces.TemplateResult) ([]byte, error) {
	// Create report data
	reportData := createReportData(results)

	// Marshal to JSON
	data, err := json.MarshalIndent(reportData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal report to JSON: %w", err)
	}

	return data, nil
}

// generateYAMLReport generates a YAML report
func (r *TemplateReporter) generateYAMLReport(results []*interfaces.TemplateResult) ([]byte, error) {
	// Create report data
	reportData := createReportData(results)

	// Marshal to YAML
	data, err := yaml.Marshal(reportData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal report to YAML: %w", err)
	}

	return data, nil
}

// generateHTMLReport generates an HTML report
func (r *TemplateReporter) generateHTMLReport(results []*interfaces.TemplateResult) ([]byte, error) {
	// Create report data
	reportData := createReportData(results)

	// Execute HTML template
	var buf bytes.Buffer
	if err := r.htmlTemplate.Execute(&buf, reportData); err != nil {
		return nil, fmt.Errorf("failed to execute HTML template: %w", err)
	}

	return buf.Bytes(), nil
}

// generateExcelReport generates an Excel report
func (r *TemplateReporter) generateExcelReport(results []*interfaces.TemplateResult) ([]byte, error) {
	// Create new Excel file
	f := excelize.NewFile()

	// Create summary sheet
	summarySheet := "Summary"
	f.SetSheetName("Sheet1", summarySheet)

	// Set headers
	headers := []string{"Template ID", "Status", "Duration (ms)", "Detected", "Score"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(summarySheet, cell, header)
	}

	// Add data
	for i, result := range results {
		row := i + 2
		f.SetCellValue(summarySheet, fmt.Sprintf("A%d", row), result.TemplateID)
		f.SetCellValue(summarySheet, fmt.Sprintf("B%d", row), result.Status)
		f.SetCellValue(summarySheet, fmt.Sprintf("C%d", row), result.Duration.Milliseconds())
		f.SetCellValue(summarySheet, fmt.Sprintf("D%d", row), result.Detected)
		f.SetCellValue(summarySheet, fmt.Sprintf("E%d", row), result.Score)
	}

	// Create details sheet
	detailsSheet := "Details"
	f.NewSheet(detailsSheet)

	// Set headers for details
	detailsHeaders := []string{"Template ID", "Start Time", "End Time", "Duration (ms)", "Status", "Detected", "Score", "Error"}
	for i, header := range detailsHeaders {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(detailsSheet, cell, header)
	}

	// Add data to details sheet
	for i, result := range results {
		row := i + 2
		f.SetCellValue(detailsSheet, fmt.Sprintf("A%d", row), result.TemplateID)
		f.SetCellValue(detailsSheet, fmt.Sprintf("B%d", row), result.StartTime.Format(time.RFC3339))
		f.SetCellValue(detailsSheet, fmt.Sprintf("C%d", row), result.EndTime.Format(time.RFC3339))
		f.SetCellValue(detailsSheet, fmt.Sprintf("D%d", row), result.Duration.Milliseconds())
		f.SetCellValue(detailsSheet, fmt.Sprintf("E%d", row), result.Status)
		f.SetCellValue(detailsSheet, fmt.Sprintf("F%d", row), result.Detected)
		f.SetCellValue(detailsSheet, fmt.Sprintf("G%d", row), result.Score)
		if result.Error != nil {
			f.SetCellValue(detailsSheet, fmt.Sprintf("H%d", row), result.Error.Error())
		}
	}

	// Save to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write Excel file: %w", err)
	}

	return buf.Bytes(), nil
}

// generateCSVReport generates a CSV report
func (r *TemplateReporter) generateCSVReport(results []*interfaces.TemplateResult) ([]byte, error) {
	// Create CSV content
	var buf bytes.Buffer

	// Write headers
	headers := []string{"Template ID", "Status", "Start Time", "End Time", "Duration (ms)", "Detected", "Score", "Error"}
	fmt.Fprintln(&buf, joinCSV(headers))

	// Write data
	for _, result := range results {
		var errorStr string
		if result.Error != nil {
			errorStr = result.Error.Error()
		}

		row := []string{
			result.TemplateID,
			string(result.Status),
			result.StartTime.Format(time.RFC3339),
			result.EndTime.Format(time.RFC3339),
			fmt.Sprintf("%d", result.Duration.Milliseconds()),
			fmt.Sprintf("%t", result.Detected),
			fmt.Sprintf("%d", result.Score),
			errorStr,
		}

		fmt.Fprintln(&buf, joinCSV(row))
	}

	return buf.Bytes(), nil
}

// generatePDFReport generates a PDF report
// This is a placeholder implementation that returns an error
// In a real implementation, this would use a PDF generation library
func (r *TemplateReporter) generatePDFReport(results []*interfaces.TemplateResult) ([]byte, error) {
	// This is a placeholder implementation
	// In a real implementation, this would use a PDF generation library
	return nil, fmt.Errorf("PDF report generation not implemented")
}

// joinCSV joins strings with commas and handles escaping
func joinCSV(values []string) string {
	var buf bytes.Buffer
	for i, value := range values {
		if i > 0 {
			buf.WriteByte(',')
		}

		// Check if value needs to be quoted
		needsQuotes := false
		for _, c := range value {
			if c == '"' || c == ',' || c == '\n' || c == '\r' {
				needsQuotes = true
				break
			}
		}

		if needsQuotes {
			buf.WriteByte('"')
			for _, c := range value {
				if c == '"' {
					buf.WriteString("\"\"") // Escape quotes
				} else {
					buf.WriteRune(c)
				}
			}
			buf.WriteByte('"')
		} else {
			buf.WriteString(value)
		}
	}
	return buf.String()
}

// ReportData represents the data for a report
type ReportData struct {
	// GeneratedAt is the time the report was generated
	GeneratedAt time.Time `json:"generated_at"`
	// Summary is the summary of the report
	Summary ReportSummary `json:"summary"`
	// Results is the list of template execution results
	Results []*interfaces.TemplateResult `json:"results"`
}

// ReportSummary represents the summary of a report
type ReportSummary struct {
	// TotalTemplates is the total number of templates executed
	TotalTemplates int `json:"total_templates"`
	// SuccessfulTemplates is the number of templates that completed successfully
	SuccessfulTemplates int `json:"successful_templates"`
	// FailedTemplates is the number of templates that failed
	FailedTemplates int `json:"failed_templates"`
	// VulnerabilitiesDetected is the number of vulnerabilities detected
	VulnerabilitiesDetected int `json:"vulnerabilities_detected"`
	// AverageScore is the average score of all templates
	AverageScore float64 `json:"average_score"`
	// AverageDuration is the average duration of all templates
	AverageDuration time.Duration `json:"average_duration"`
	// TotalDuration is the total duration of all templates
	TotalDuration time.Duration `json:"total_duration"`
}

// createReportData creates report data from template execution results
func createReportData(results []*interfaces.TemplateResult) *ReportData {
	// Sort results by template ID
	sort.Slice(results, func(i, j int) bool {
		return results[i].TemplateID < results[j].TemplateID
	})

	// Calculate summary
	summary := ReportSummary{
		TotalTemplates: len(results),
	}

	var totalScore int
	var totalDuration time.Duration

	for _, result := range results {
		if result.Status == string(interfaces.StatusCompleted) {
			summary.SuccessfulTemplates++
		} else if result.Status == string(interfaces.StatusFailed) {
			summary.FailedTemplates++
		}

		if result.Detected {
			summary.VulnerabilitiesDetected++
		}

		totalScore += result.Score
		totalDuration += result.Duration
	}

	// Calculate averages
	if summary.TotalTemplates > 0 {
		summary.AverageScore = float64(totalScore) / float64(summary.TotalTemplates)
		summary.AverageDuration = totalDuration / time.Duration(summary.TotalTemplates)
	}

	summary.TotalDuration = totalDuration

	return &ReportData{
		GeneratedAt: time.Now(),
		Summary:     summary,
		Results:     results,
	}
}

// defaultHTMLTemplate is the default HTML template for reports
var defaultHTMLTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>LLMrecon Test Report</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 20px;
            color: #333;
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
        }
        .summary {
            background-color: #f5f5f5;
            padding: 20px;
            border-radius: 5px;
            margin-bottom: 30px;
        }
        .summary h2 {
            margin-top: 0;
        }
        .summary-grid {
            display: grid;
            grid-template-columns: repeat(3, 1fr);
            gap: 20px;
        }
        .summary-item {
            background-color: white;
            padding: 15px;
            border-radius: 5px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
        }
        .summary-item h3 {
            margin-top: 0;
            color: #555;
        }
        .summary-item p {
            font-size: 24px;
            font-weight: bold;
            margin: 10px 0 0;
        }
        .results {
            margin-bottom: 30px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th, td {
            padding: 12px 15px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #f8f8f8;
            font-weight: bold;
        }
        tr:hover {
            background-color: #f5f5f5;
        }
        .status-completed {
            color: green;
        }
        .status-failed {
            color: red;
        }
        .detected-true {
            color: red;
            font-weight: bold;
        }
        .detected-false {
            color: green;
        }
        .footer {
            text-align: center;
            margin-top: 30px;
            color: #777;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>LLMrecon Test Report</h1>
        <p>Generated at: {{.GeneratedAt.Format "Jan 02, 2006 15:04:05 MST"}}</p>
    </div>
    
    <div class="summary">
        <h2>Summary</h2>
        <div class="summary-grid">
            <div class="summary-item">
                <h3>Total Templates</h3>
                <p>{{.Summary.TotalTemplates}}</p>
            </div>
            <div class="summary-item">
                <h3>Successful</h3>
                <p>{{.Summary.SuccessfulTemplates}}</p>
            </div>
            <div class="summary-item">
                <h3>Failed</h3>
                <p>{{.Summary.FailedTemplates}}</p>
            </div>
            <div class="summary-item">
                <h3>Vulnerabilities Detected</h3>
                <p>{{.Summary.VulnerabilitiesDetected}}</p>
            </div>
            <div class="summary-item">
                <h3>Average Score</h3>
                <p>{{printf "%.1f" .Summary.AverageScore}}</p>
            </div>
            <div class="summary-item">
                <h3>Total Duration</h3>
                <p>{{.Summary.TotalDuration.Seconds}} seconds</p>
            </div>
        </div>
    </div>
    
    <div class="results">
        <h2>Test Results</h2>
        <table>
            <thead>
                <tr>
                    <th>Template ID</th>
                    <th>Status</th>
                    <th>Duration</th>
                    <th>Detected</th>
                    <th>Score</th>
                </tr>
            </thead>
            <tbody>
                {{range .Results}}
                <tr>
                    <td>{{.TemplateID}}</td>
                    <td class="status-{{.Status}}">{{.Status}}</td>
                    <td>{{.Duration.Milliseconds}} ms</td>
                    <td class="detected-{{.Detected}}">{{.Detected}}</td>
                    <td>{{.Score}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </div>
    
    <div class="footer">
        <p>LLMreconing Tool &copy; {{.GeneratedAt.Year}}</p>
    </div>
</body>
</html>`
