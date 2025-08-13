package ui

import (
	"encoding/json"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ExportPreview provides preview functionality for different export formats
type ExportPreview struct {
	terminal *Terminal
	style    *DashboardStyle
}

// NewExportPreview creates a new export preview handler
func NewExportPreview(terminal *Terminal) *ExportPreview {
	return &ExportPreview{
		terminal: terminal,
		style:    newDashboardStyle(),
	}
}

// ShowFormatSelection displays available export formats with previews
func (ep *ExportPreview) ShowFormatSelection(data interface{}) (string, error) {
	ep.terminal.Clear()
	ep.terminal.HeaderBox("Export Format Selection")
	
	formats := []ExportFormat{
		{
			ID:          "json",
			Name:        "JSON",
			Description: "Machine-readable format for automation and integration",
			Extensions:  []string{".json"},
			Features:    []string{"Full data export", "API-friendly", "Parseable"},
		},
		{
			ID:          "yaml",
			Name:        "YAML",
			Description: "Human-readable format for configuration and documentation",
			Extensions:  []string{".yaml", ".yml"},
			Features:    []string{"Readable", "Comments supported", "Compact"},
		},
		{
			ID:          "markdown",
			Name:        "Markdown",
			Description: "Documentation format for reports and wikis",
			Extensions:  []string{".md"},
			Features:    []string{"Formatted text", "Tables", "GitHub/GitLab compatible"},
		},
		{
			ID:          "html",
			Name:        "HTML Report",
			Description: "Interactive web-based report with charts and navigation",
			Extensions:  []string{".html"},
			Features:    []string{"Interactive", "Charts", "Standalone file"},
		},
		{
			ID:          "pdf",
			Name:        "PDF Document",
			Description: "Professional report format for sharing and archiving",
			Extensions:  []string{".pdf"},
			Features:    []string{"Print-ready", "Professional", "Secure"},
		},
		{
			ID:          "csv",
			Name:        "CSV",
			Description: "Spreadsheet format for data analysis",
			Extensions:  []string{".csv"},
			Features:    []string{"Excel compatible", "Data analysis", "Simple"},
		},
		{
			ID:          "sarif",
			Name:        "SARIF",
			Description: "Static Analysis Results Interchange Format",
			Extensions:  []string{".sarif"},
			Features:    []string{"Standard format", "IDE integration", "CI/CD compatible"},
		},
		{
			ID:          "jira",
			Name:        "JIRA Issues",
			Description: "Create JIRA tickets for vulnerabilities",
			Extensions:  []string{},
			Features:    []string{"Issue tracking", "Workflow integration", "Team collaboration"},
		},
	}
	
	// Display format options
	for i, format := range formats {
		ep.terminal.Subsection(fmt.Sprintf("%d. %s", i+1, format.Name))
		ep.terminal.Info(format.Description)
		
		if len(format.Features) > 0 {
			ep.terminal.Muted("Features: " + strings.Join(format.Features, " • "))
		}
		
		fmt.Println()
	}
	
	// Format selection
	choice, err := ep.terminal.Select("Select export format:", ep.getFormatNames(formats))
	if err != nil {
		return "", err
	}
	
	selectedFormat := formats[choice]
	
	// Show preview
	ep.showPreview(selectedFormat, data)
	
	// Confirm selection
	confirm, err := ep.terminal.Confirm("Export in " + selectedFormat.Name + " format?")
	if err != nil {
		return "", err
	}
	
	if !confirm {
		return ep.ShowFormatSelection(data) // Recursive call to reselect
	}
	
	return selectedFormat.ID, nil
}

// ShowPreview displays a preview of the selected format
func (ep *ExportPreview) ShowPreview(formatID string, data interface{}) error {
	format := ep.getFormatByID(formatID)
	if format == nil {
		return fmt.Errorf("unknown format: %s", formatID)
	}
	
	ep.terminal.Clear()
	ep.terminal.HeaderBox("Preview: " + format.Name + " Export")
	
	// Generate preview based on format
	preview := ep.generatePreview(format, data)
	
	// Display preview in a scrollable box
	ep.terminal.Section("Format Preview")
	ep.displayPreview(preview, format)
	
	// Show export options
	ep.showExportOptions(format)
	
	return nil
}

// generatePreview creates a preview for the specified format
func (ep *ExportPreview) generatePreview(format *ExportFormat, data interface{}) string {
	// Sample data for preview
	sampleData := ep.getSampleData(data)
	
	switch format.ID {
	case "json":
		return ep.generateJSONPreview(sampleData)
	case "yaml":
		return ep.generateYAMLPreview(sampleData)
	case "markdown":
		return ep.generateMarkdownPreview(sampleData)
	case "html":
		return ep.generateHTMLPreview(sampleData)
	case "pdf":
		return ep.generatePDFPreview(sampleData)
	case "csv":
		return ep.generateCSVPreview(sampleData)
	case "sarif":
		return ep.generateSARIFPreview(sampleData)
	case "jira":
		return ep.generateJIRAPreview(sampleData)
	default:
		return "Preview not available for this format"
	}
}

// Format-specific preview generators

func (ep *ExportPreview) generateJSONPreview(data *SampleReportData) string {
	preview := map[string]interface{}{
		"report": map[string]interface{}{
			"id":        data.ID,
			"timestamp": data.Timestamp,
			"summary": map[string]interface{}{
				"total_tests":         data.TotalTests,
				"vulnerabilities":     data.VulnerabilityCount,
				"risk_score":          data.RiskScore,
				"compliance_status":   data.ComplianceStatus,
			},
			"vulnerabilities": []map[string]interface{}{
				{
					"id":          "VULN-001",
					"severity":    "Critical",
					"category":    "Prompt Injection",
					"description": "Direct prompt override vulnerability",
					"cvss_score":  9.8,
				},
				{
					"id":          "VULN-002",
					"severity":    "High",
					"category":    "Data Leakage",
					"description": "Sensitive data exposure through model inversion",
					"cvss_score":  7.5,
				},
			},
		},
	}
	
	jsonBytes, _ := json.MarshalIndent(preview, "", "  ")
	return string(jsonBytes)
}

func (ep *ExportPreview) generateYAMLPreview(data *SampleReportData) string {
	preview := map[string]interface{}{
		"report": map[string]interface{}{
			"id":        data.ID,
			"timestamp": data.Timestamp,
			"metadata": map[string]interface{}{
				"scanner_version": "1.5.0",
				"template_set":    "owasp-llm-v1",
			},
			"summary": map[string]interface{}{
				"duration":      "15m32s",
				"total_tests":   data.TotalTests,
				"passed":        120,
				"failed":        30,
				"error_rate":    "2.5%",
			},
			"findings": []map[string]interface{}{
				{
					"severity":     "critical",
					"type":         "prompt_injection",
					"confidence":   "high",
					"remediation":  "Implement input validation and context boundaries",
				},
			},
		},
	}
	
	yamlBytes, _ := yaml.Marshal(preview)
	return string(yamlBytes)
}

func (ep *ExportPreview) generateMarkdownPreview(data *SampleReportData) string {
	return fmt.Sprintf(`# Security Scan Report

**Report ID:** %s  
**Generated:** %s  
**Scanner Version:** 1.5.0

## Executive Summary

Total tests executed: **%d**  
Vulnerabilities found: **%d**  
Overall risk score: **%.1f/10**  
Compliance status: **%s**

## Vulnerability Summary

| Severity | Count | Percentage |
|----------|-------|------------|
| Critical | 2     | 13%%       |
| High     | 5     | 33%%       |
| Medium   | 8     | 53%%       |
| Low      | 0     | 0%%        |

## Top Findings

### 1. Direct Prompt Injection (Critical)

**Description:** System prompts can be overridden through crafted inputs.

**Impact:** Complete bypass of safety controls and system boundaries.

**Recommendation:** Implement robust input validation and prompt isolation.

### 2. Training Data Extraction (High)

**Description:** Model inversion attacks can extract sensitive training data.

**Impact:** Potential exposure of confidential information.

**Recommendation:** Apply differential privacy and output filtering.

## Compliance Mapping

- ✅ **OWASP LLM01:** Prompt Injection - 2 findings
- ✅ **OWASP LLM06:** Sensitive Information - 1 finding
- ✅ **ISO 42001:** AI Governance - Partial compliance

## Recommendations

1. Implement comprehensive input validation
2. Deploy prompt isolation mechanisms
3. Enable output content filtering
4. Regular security assessments

---
*Generated by LLMrecon v1.5.0*`,
		data.ID,
		data.Timestamp,
		data.TotalTests,
		data.VulnerabilityCount,
		data.RiskScore,
		data.ComplianceStatus,
	)
}

func (ep *ExportPreview) generateHTMLPreview(data *SampleReportData) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Security Scan Report - %s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 8px; }
        .metric { display: inline-block; margin: 10px; padding: 15px; 
                  background: white; border: 1px solid #ddd; border-radius: 4px; }
        .critical { color: #d32f2f; font-weight: bold; }
        .high { color: #f57c00; font-weight: bold; }
        .chart { width: 100%%; height: 300px; background: #fafafa; 
                 border: 1px solid #ddd; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Security Scan Report</h1>
        <p>Report ID: %s | Generated: %s</p>
    </div>
    
    <div class="metrics">
        <div class="metric">
            <h3>Total Tests</h3>
            <p style="font-size: 24px;">%d</p>
        </div>
        <div class="metric">
            <h3>Vulnerabilities</h3>
            <p style="font-size: 24px;" class="critical">%d</p>
        </div>
        <div class="metric">
            <h3>Risk Score</h3>
            <p style="font-size: 24px;">%.1f/10</p>
        </div>
    </div>
    
    <div class="chart">
        [Interactive Chart Placeholder]
    </div>
    
    <h2>Findings</h2>
    <table border="1" style="width: 100%%; border-collapse: collapse;">
        <tr>
            <th>Severity</th>
            <th>Category</th>
            <th>Description</th>
            <th>CVSS</th>
        </tr>
        <tr>
            <td class="critical">Critical</td>
            <td>Prompt Injection</td>
            <td>Direct prompt override vulnerability</td>
            <td>9.8</td>
        </tr>
    </table>
</body>
</html>`,
		data.ID,
		data.ID,
		data.Timestamp,
		data.TotalTests,
		data.VulnerabilityCount,
		data.RiskScore,
	)
}

func (ep *ExportPreview) generatePDFPreview(data *SampleReportData) string {
	return fmt.Sprintf(`PDF Document Preview
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

                    SECURITY SCAN REPORT
                    
    Report ID: %s
    Date: %s
    Classification: Confidential

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

EXECUTIVE SUMMARY

This report contains the results of security testing performed on
the target system. A total of %d tests were executed, revealing
%d vulnerabilities with an overall risk score of %.1f/10.

KEY FINDINGS

• Critical vulnerabilities: 2
• High severity issues: 5  
• Remediation required: Immediate
• Compliance gaps: 3 standards affected

[Chart: Vulnerability Distribution]
    Critical ████████░░░░░░░░ 13%%
    High     ████████████████ 33%%
    Medium   ████████████████ 53%%
    Low      ░░░░░░░░░░░░░░░░  0%%

DETAILED FINDINGS
─────────────────────────────────────────────────
1. PROMPT INJECTION VULNERABILITY (CRITICAL)
   
   Description: System prompts can be overridden...
   Impact: Complete security bypass
   CVSS Score: 9.8
   
   Remediation Steps:
   1. Implement input validation
   2. Use prompt isolation
   3. Deploy safety boundaries

[Page 1 of 15]`,
		data.ID,
		data.Timestamp,
		data.TotalTests,
		data.VulnerabilityCount,
		data.RiskScore,
	)
}

func (ep *ExportPreview) generateCSVPreview(data *SampleReportData) string {
	return `"ID","Severity","Category","Description","CVSS Score","Status","Remediation"
"VULN-001","Critical","Prompt Injection","Direct prompt override vulnerability","9.8","Open","Implement input validation"
"VULN-002","High","Data Leakage","Sensitive data exposure through model inversion","7.5","Open","Apply output filtering"
"VULN-003","High","Insecure Output","XSS vulnerability in generated content","7.2","Open","Sanitize HTML output"
"VULN-004","Medium","Access Control","Insufficient authorization checks","6.5","Open","Implement RBAC"
"VULN-005","Medium","Logging","Sensitive data in logs","5.3","Open","Redact sensitive information"

[Showing first 5 of 15 rows]`
}

func (ep *ExportPreview) generateSARIFPreview(data *SampleReportData) string {
	return `{
  "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "LLMrecon",
          "version": "1.5.0",
          "informationUri": "https://github.com/LLMrecon",
          "rules": [
            {
              "id": "LLM01",
              "name": "PromptInjection",
              "shortDescription": {
                "text": "Prompt Injection Vulnerability"
              },
              "fullDescription": {
                "text": "Direct prompt injection allows attackers to override system prompts"
              },
              "defaultConfiguration": {
                "level": "error"
              },
              "properties": {
                "tags": ["security", "prompt-injection", "owasp-llm"]
              }
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "LLM01",
          "level": "error",
          "message": {
            "text": "Prompt injection vulnerability detected in chat endpoint"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "api/chat/completions"
                }
              }
            }
          ]
        }
      ]
    }
  ]
}`
}

func (ep *ExportPreview) generateJIRAPreview(data *SampleReportData) string {
	return fmt.Sprintf(`JIRA Issue Preview
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Project: SECURITY
Issue Type: Bug
Priority: Highest

Summary: [LLM-SEC] Critical: Prompt Injection Vulnerability

Description:
-----------
h3. Overview
Security scan %s identified a critical prompt injection vulnerability.

h3. Details
* *Severity:* Critical
* *CVSS Score:* 9.8
* *Category:* Prompt Injection
* *Scan Date:* %s

h3. Impact
Attackers can override system prompts and bypass security controls.

h3. Reproduction Steps
1. Send crafted prompt to /api/chat endpoint
2. Include injection payload in user message
3. Observe system prompt override

h3. Remediation
1. Implement input validation
2. Deploy prompt isolation mechanisms
3. Add context boundaries

Labels: security, llm, prompt-injection, critical
Components: API, Security
Affects Version: 2.0.0
Fix Version: 2.0.1

Attachments:
- Full scan report (PDF)
- Reproduction script
- Remediation guide`,
		data.ID,
		data.Timestamp,
	)
}

// Display helpers

func (ep *ExportPreview) displayPreview(preview string, format *ExportFormat) {
	lines := strings.Split(preview, "\n")
	maxLines := 30 // Maximum lines to show in preview
	
	if len(lines) > maxLines {
		// Show truncated preview
		for i := 0; i < maxLines-3; i++ {
			fmt.Println(ep.formatPreviewLine(lines[i], format))
		}
		fmt.Println(ep.style.Info.Render("..."))
		fmt.Println(ep.style.Info.Render(fmt.Sprintf("[%d more lines]", len(lines)-maxLines)))
		fmt.Println(ep.formatPreviewLine(lines[len(lines)-1], format))
	} else {
		// Show full preview
		for _, line := range lines {
			fmt.Println(ep.formatPreviewLine(line, format))
		}
	}
}

func (ep *ExportPreview) formatPreviewLine(line string, format *ExportFormat) string {
	// Apply syntax highlighting based on format
	switch format.ID {
	case "json", "sarif":
		return ep.highlightJSON(line)
	case "yaml":
		return ep.highlightYAML(line)
	case "markdown":
		return ep.highlightMarkdown(line)
	case "html":
		return ep.highlightHTML(line)
	case "csv":
		return ep.highlightCSV(line)
	default:
		return line
	}
}

// Syntax highlighting helpers

func (ep *ExportPreview) highlightJSON(line string) string {
	// Simple JSON syntax highlighting
	line = strings.ReplaceAll(line, `"`, ep.style.Success.Render(`"`))
	
	// Highlight numbers
	for _, num := range []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"} {
		if strings.Contains(line, num) && !strings.Contains(line, `"`+num) {
			line = strings.ReplaceAll(line, num, ep.style.Warning.Render(num))
		}
	}
	
	// Highlight keywords
	keywords := []string{"true", "false", "null"}
	for _, kw := range keywords {
		line = strings.ReplaceAll(line, kw, ep.style.Info.Render(kw))
	}
	
	return line
}

func (ep *ExportPreview) highlightYAML(line string) string {
	// YAML key highlighting
	if strings.Contains(line, ":") && !strings.HasPrefix(strings.TrimSpace(line), "-") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			return ep.style.Info.Render(parts[0]) + ":" + parts[1]
		}
	}
	return line
}

func (ep *ExportPreview) highlightMarkdown(line string) string {
	// Headers
	if strings.HasPrefix(line, "#") {
		return ep.style.Title.Render(line)
	}
	
	// Bold text
	if strings.Contains(line, "**") {
		// Simple bold highlighting
		parts := strings.Split(line, "**")
		result := ""
		for i, part := range parts {
			if i%2 == 1 {
				result += ep.style.Metric.Render(part)
			} else {
				result += part
			}
		}
		return result
	}
	
	return line
}

func (ep *ExportPreview) highlightHTML(line string) string {
	// HTML tag highlighting
	if strings.Contains(line, "<") && strings.Contains(line, ">") {
		// Simple tag highlighting
		line = strings.ReplaceAll(line, "<", ep.style.Info.Render("<"))
		line = strings.ReplaceAll(line, ">", ep.style.Info.Render(">"))
	}
	return line
}

func (ep *ExportPreview) highlightCSV(line string) string {
	// CSV header row
	if strings.HasPrefix(line, `"ID"`) {
		return ep.style.Title.Render(line)
	}
	
	// Highlight severity column
	if strings.Contains(line, `"Critical"`) {
		line = strings.ReplaceAll(line, `"Critical"`, ep.style.Critical.Render(`"Critical"`))
	}
	if strings.Contains(line, `"High"`) {
		line = strings.ReplaceAll(line, `"High"`, ep.style.Warning.Render(`"High"`))
	}
	
	return line
}

// Export options

func (ep *ExportPreview) showExportOptions(format *ExportFormat) {
	ep.terminal.Section("Export Options")
	
	options := ep.getFormatOptions(format)
	
	for _, opt := range options {
		ep.terminal.Info(fmt.Sprintf("• %s: %s", opt.Name, opt.Description))
		if opt.Default != "" {
			ep.terminal.Muted("  Default: " + opt.Default)
		}
	}
	
	// File size estimate
	ep.terminal.Subsection("Estimated File Size")
	ep.terminal.Info(ep.estimateFileSize(format))
	
	// Compatibility notes
	if len(format.Compatible) > 0 {
		ep.terminal.Subsection("Compatible With")
		ep.terminal.Info(strings.Join(format.Compatible, ", "))
	}
}

func (ep *ExportPreview) getFormatOptions(format *ExportFormat) []ExportOption {
	switch format.ID {
	case "json":
		return []ExportOption{
			{Name: "Indent", Description: "Pretty-print with indentation", Default: "2 spaces"},
			{Name: "Include Metadata", Description: "Add scanner metadata", Default: "true"},
			{Name: "Minify", Description: "Compress output", Default: "false"},
		}
	case "html":
		return []ExportOption{
			{Name: "Theme", Description: "Visual theme", Default: "light"},
			{Name: "Include Charts", Description: "Add interactive charts", Default: "true"},
			{Name: "Embed Assets", Description: "Single file with embedded CSS/JS", Default: "true"},
		}
	case "pdf":
		return []ExportOption{
			{Name: "Page Size", Description: "Paper format", Default: "A4"},
			{Name: "Include TOC", Description: "Table of contents", Default: "true"},
			{Name: "Watermark", Description: "Add confidential watermark", Default: "false"},
		}
	default:
		return []ExportOption{}
	}
}

func (ep *ExportPreview) estimateFileSize(format *ExportFormat) string {
	// Rough estimates based on format
	sizes := map[string]string{
		"json":     "~250 KB",
		"yaml":     "~300 KB",
		"markdown": "~150 KB",
		"html":     "~500 KB",
		"pdf":      "~2 MB",
		"csv":      "~100 KB",
		"sarif":    "~200 KB",
		"jira":     "N/A (API call)",
	}
	
	if size, ok := sizes[format.ID]; ok {
		return size
	}
	return "Unknown"
}

// Helper methods

func (ep *ExportPreview) getFormatNames(formats []ExportFormat) []string {
	names := make([]string, len(formats))
	for i, f := range formats {
		names[i] = f.Name
	}
	return names
}

func (ep *ExportPreview) getFormatByID(id string) *ExportFormat {
	formats := []ExportFormat{
		{ID: "json", Name: "JSON"},
		{ID: "yaml", Name: "YAML"},
		{ID: "markdown", Name: "Markdown"},
		{ID: "html", Name: "HTML Report"},
		{ID: "pdf", Name: "PDF Document"},
		{ID: "csv", Name: "CSV"},
		{ID: "sarif", Name: "SARIF"},
		{ID: "jira", Name: "JIRA Issues"},
	}
	
	for _, f := range formats {
		if f.ID == id {
			return &f
		}
	}
	return nil
}

func (ep *ExportPreview) getSampleData(data interface{}) *SampleReportData {
	// If real data provided, extract sample
	// Otherwise, use default sample data
	return &SampleReportData{
		ID:                 "SCAN-2024-001",
		Timestamp:          time.Now().Format("2006-01-02 15:04:05"),
		TotalTests:         150,
		VulnerabilityCount: 15,
		RiskScore:          7.8,
		ComplianceStatus:   "Partial",
	}
}

// ShowComparisonPreview shows a side-by-side format comparison
func (ep *ExportPreview) ShowComparisonPreview(formats []string, data interface{}) error {
	ep.terminal.Clear()
	ep.terminal.HeaderBox("Export Format Comparison")
	
	// Generate previews for each format
	previews := make(map[string]string)
	for _, formatID := range formats {
		format := ep.getFormatByID(formatID)
		if format != nil {
			previews[formatID] = ep.generatePreview(format, data)
		}
	}
	
	// Display side by side (simplified for 2 formats)
	if len(formats) == 2 {
		ep.showSideBySide(formats[0], formats[1], previews)
	} else {
		// Sequential display for more formats
		for _, formatID := range formats {
			format := ep.getFormatByID(formatID)
			ep.terminal.Section(format.Name + " Preview")
			lines := strings.Split(previews[formatID], "\n")
			for i, line := range lines[:min(10, len(lines))] {
				fmt.Println(ep.formatPreviewLine(line, format))
			}
			if len(lines) > 10 {
				ep.terminal.Info("... (truncated)")
			}
			fmt.Println()
		}
	}
	
	return nil
}

func (ep *ExportPreview) showSideBySide(format1, format2 string, previews map[string]string) {
	f1 := ep.getFormatByID(format1)
	f2 := ep.getFormatByID(format2)
	
	lines1 := strings.Split(previews[format1], "\n")
	lines2 := strings.Split(previews[format2], "\n")
	
	maxLines := max(len(lines1), len(lines2))
	if maxLines > 20 {
		maxLines = 20
	}
	
	// Headers
	fmt.Printf("%-40s | %-40s\n", 
		ep.style.Title.Render(f1.Name),
		ep.style.Title.Render(f2.Name))
	fmt.Println(strings.Repeat("─", 81))
	
	// Content
	for i := 0; i < maxLines; i++ {
		line1 := ""
		line2 := ""
		
		if i < len(lines1) {
			line1 = truncate(lines1[i], 40)
		}
		if i < len(lines2) {
			line2 = truncate(lines2[i], 40)
		}
		
		fmt.Printf("%-40s | %-40s\n", line1, line2)
	}
	
	if maxLines < max(len(lines1), len(lines2)) {
		fmt.Printf("%-40s | %-40s\n",
			ep.style.Info.Render("... (truncated)"),
			ep.style.Info.Render("... (truncated)"))
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Data structures

type ExportFormat struct {
	ID          string
	Name        string
	Description string
	Extensions  []string
	Features    []string
	Compatible  []string
}

type ExportOption struct {
	Name        string
	Description string
	Default     string
}

type SampleReportData struct {
	ID                 string
	Timestamp          string
	TotalTests         int
	VulnerabilityCount int
	RiskScore          float64
	ComplianceStatus   string
}