package reporting

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/reporting/common"
)

// Add type definitions that are missing
type SecurityReport struct {
	Metadata   ReportMetadata   `json:"metadata"`
	Summary    SecurityReportSummary    `json:"summary"`
	Findings   []Finding        `json:"findings"`
	Statistics ReportStatistics `json:"statistics"`
	Compliance map[string]*ComplianceStatus `json:"compliance,omitempty"`
}

type ReportMetadata struct {
	ID        string                 `json:"id"`
	Title     string                 `json:"title"`
	CreatedAt time.Time              `json:"created_at"`
	Version   string                 `json:"version"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type SecurityReportSummary struct {
	TotalFindings     int                `json:"total_findings"`
	RiskScore         float64            `json:"risk_score"`
	SeverityBreakdown map[string]int     `json:"severity_breakdown"`
	CategoryBreakdown map[string]int     `json:"category_breakdown"`
}

type Finding struct {
	ID           string                 `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	Category     string                 `json:"category"`
	Subcategory  string                 `json:"subcategory"`
	Severity     string                 `json:"severity"`
	Confidence   float64                `json:"confidence"`
	Title        string                 `json:"title"`
	Description  string                 `json:"description"`
	Remediation  Remediation            `json:"remediation"`
	OWASPMapping []string               `json:"owasp_mapping"`
	References   []Reference            `json:"references"`
	Evidence     Evidence               `json:"evidence"`
	Tags         []string               `json:"tags"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type Remediation struct {
	Summary string   `json:"summary"`
	Steps   []string `json:"steps"`
}

type Reference struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type Evidence struct {
	Request  string `json:"request,omitempty"`
	Response string `json:"response,omitempty"`
}

type ReportStatistics struct {
	StartTime         time.Time       `json:"start_time"`
	EndTime           time.Time       `json:"end_time"`
	FindingsPerMinute float64         `json:"findings_per_minute"`
	AverageConfidence float64         `json:"average_confidence"`
	TopCategories     []CategoryCount `json:"top_categories"`
}

type CategoryCount struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

type ComplianceStatus struct {
	Compliant bool     `json:"compliant"`
	Coverage  float64  `json:"coverage"`
	Gaps      []string `json:"gaps"`
}

// MultiFormatRenderer handles rendering reports in multiple formats
type MultiFormatRenderer struct {
	renderers map[common.ReportFormat]Renderer
	templates map[string]*template.Template
}

// Renderer interface for format-specific renderers
type Renderer interface {
	Render(report *SecurityReport, options RenderOptions) ([]byte, error)
	GetContentType() string
	GetFileExtension() string
}

// RenderOptions contains options for rendering
type RenderOptions struct {
	Template       string                 `json:"template"`
	Locale         string                 `json:"locale"`
	TimeZone       string                 `json:"timezone"`
	IncludeRawData bool                   `json:"includeRawData"`
	Filters        ReportFilter           `json:"filters"`
	CustomFields   map[string]interface{} `json:"customFields"`
}

// ReportFilter defines filtering options
type ReportFilter struct {
	Severity   []string               `json:"severity"`
	Categories []string               `json:"categories"`
	DateRange  *DateRange             `json:"dateRange"`
	Tags       []string               `json:"tags"`
	Custom     map[string]interface{} `json:"custom"`
}

// DateRange represents a time range
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// NewMultiFormatRenderer creates a new multi-format renderer
func NewMultiFormatRenderer() *MultiFormatRenderer {
	r := &MultiFormatRenderer{
		renderers: make(map[ReportFormat]Renderer),
		templates: make(map[string]*template.Template),
	}

	// Register default renderers
	r.RegisterRenderer(common.JSONFormat, &JSONRenderer{})
	r.RegisterRenderer(common.CSVFormat, &CSVRenderer{})
	r.RegisterRenderer(common.HTMLFormat, &HTMLRenderer{templates: r.templates})
	r.RegisterRenderer(common.MarkdownFormat, &MarkdownRenderer{})
	// PDF and Excel renderers would be registered if implemented

	return r
}

// RegisterRenderer registers a renderer for a format
func (r *MultiFormatRenderer) RegisterRenderer(format common.ReportFormat, renderer Renderer) {
	r.renderers[format] = renderer
}

// Render renders a report in the specified format
func (r *MultiFormatRenderer) Render(report *SecurityReport, format common.ReportFormat, options RenderOptions) ([]byte, error) {
	renderer, exists := r.renderers[format]
	if !exists {
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	// Apply filters if specified
	if options.Filters.Severity != nil || options.Filters.Categories != nil || options.Filters.Tags != nil || options.Filters.DateRange != nil {
		report = r.applyFilters(report, options.Filters)
	}

	return renderer.Render(report, options)
}

// applyFilters applies filters to a report
func (r *MultiFormatRenderer) applyFilters(report *SecurityReport, filter ReportFilter) *SecurityReport {
	filtered := &SecurityReport{
		Metadata:   report.Metadata,
		Summary:    report.Summary,
		Statistics: report.Statistics,
		Compliance: report.Compliance,
		Findings:   []Finding{},
	}

	for _, finding := range report.Findings {
		if r.matchesFilter(finding, filter) {
			filtered.Findings = append(filtered.Findings, finding)
		}
	}

	// Update summary based on filtered findings
	filtered.Summary = r.calculateSummary(filtered.Findings)

	return filtered
}

// matchesFilter checks if a finding matches the filter criteria
func (r *MultiFormatRenderer) matchesFilter(finding Finding, filter ReportFilter) bool {
	// Check severity
	if len(filter.Severity) > 0 {
		found := false
		for _, sev := range filter.Severity {
			if finding.Severity == sev {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check categories
	if len(filter.Categories) > 0 {
		found := false
		for _, cat := range filter.Categories {
			if finding.Category == cat {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check tags
	if len(filter.Tags) > 0 {
		found := false
		for _, tag := range filter.Tags {
			for _, findingTag := range finding.Tags {
				if tag == findingTag {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check date range
	if filter.DateRange != nil {
		if finding.Timestamp.Before(filter.DateRange.Start) || 
		   finding.Timestamp.After(filter.DateRange.End) {
			return false
		}
	}

	return true
}

// calculateSummary calculates summary statistics from findings
func (r *MultiFormatRenderer) calculateSummary(findings []Finding) SecurityReportSummary {
	summary := SecurityReportSummary{
		TotalFindings:     len(findings),
		SeverityBreakdown: make(map[string]int),
		CategoryBreakdown: make(map[string]int),
	}

	for _, finding := range findings {
		summary.SeverityBreakdown[finding.Severity]++
		summary.CategoryBreakdown[finding.Category]++
	}

	// Calculate risk score
	weights := map[string]float64{
		"critical": 10.0,
		"high":     5.0,
		"medium":   2.0,
		"low":      1.0,
		"info":     0.5,
	}

	for severity, count := range summary.SeverityBreakdown {
		if weight, ok := weights[strings.ToLower(severity)]; ok {
			summary.RiskScore += weight * float64(count)
		}
	}

	return summary
}

// JSONRenderer renders reports as JSON
type JSONRenderer struct{}

// Render renders a report as JSON
func (r *JSONRenderer) Render(report *SecurityReport, options RenderOptions) ([]byte, error) {
	if options.IncludeRawData {
		return json.MarshalIndent(report, "", "  ")
	}

	// Create a simplified version without raw data
	simplified := r.simplifyReport(report)
	return json.MarshalIndent(simplified, "", "  ")
}

// GetContentType returns the content type for JSON
func (r *JSONRenderer) GetContentType() string {
	return "application/json"
}

// GetFileExtension returns the file extension for JSON
func (r *JSONRenderer) GetFileExtension() string {
	return ".json"
}

// simplifyReport creates a simplified version of the report
func (r *JSONRenderer) simplifyReport(report *SecurityReport) interface{} {
	// Remove large raw data fields
	simplified := map[string]interface{}{
		"metadata":   report.Metadata,
		"summary":    report.Summary,
		"statistics": report.Statistics,
		"compliance": report.Compliance,
		"findings":   []interface{}{},
	}

	for _, finding := range report.Findings {
		simpleFinding := map[string]interface{}{
			"id":           finding.ID,
			"timestamp":    finding.Timestamp,
			"category":     finding.Category,
			"severity":     finding.Severity,
			"title":        finding.Title,
			"description":  finding.Description,
			"remediation":  finding.Remediation,
			"owaspMapping": finding.OWASPMapping,
		}
		simplified["findings"] = append(simplified["findings"].([]interface{}), simpleFinding)
	}

	return simplified
}

// CSVRenderer renders reports as CSV
type CSVRenderer struct{}

// Render renders a report as CSV
func (r *CSVRenderer) Render(report *SecurityReport, options RenderOptions) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write headers
	headers := []string{
		"ID", "Timestamp", "Category", "Subcategory", "Severity", 
		"Confidence", "Title", "Description", "OWASP_Mapping", 
		"Remediation", "References",
	}
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	// Write findings
	for _, finding := range report.Findings {
		row := []string{
			finding.ID,
			finding.Timestamp.Format(time.RFC3339),
			finding.Category,
			finding.Subcategory,
			finding.Severity,
			fmt.Sprintf("%.2f", finding.Confidence),
			finding.Title,
			finding.Description,
			strings.Join(finding.OWASPMapping, ";"),
			finding.Remediation.Summary,
			r.formatReferences(finding.References),
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}

// GetContentType returns the content type for CSV
func (r *CSVRenderer) GetContentType() string {
	return "text/csv"
}

// GetFileExtension returns the file extension for CSV
func (r *CSVRenderer) GetFileExtension() string {
	return ".csv"
}

// formatReferences formats references for CSV
func (r *CSVRenderer) formatReferences(refs []Reference) string {
	var formatted []string
	for _, ref := range refs {
		formatted = append(formatted, ref.URL)
	}
	return strings.Join(formatted, ";")
}

// HTMLRenderer renders reports as HTML
type HTMLRenderer struct {
	templates map[string]*template.Template
}

// Render renders a report as HTML
func (r *HTMLRenderer) Render(report *SecurityReport, options RenderOptions) ([]byte, error) {
	tmpl := r.getTemplate(options.Template)
	if tmpl == nil {
		// Use default template
		tmpl = r.createDefaultTemplate()
	}

	var buf bytes.Buffer
	data := r.prepareTemplateData(report, options)
	
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("template execution failed: %w", err)
	}

	return buf.Bytes(), nil
}

// GetContentType returns the content type for HTML
func (r *HTMLRenderer) GetContentType() string {
	return "text/html"
}

// GetFileExtension returns the file extension for HTML
func (r *HTMLRenderer) GetFileExtension() string {
	return ".html"
}

// getTemplate retrieves a template by name
func (r *HTMLRenderer) getTemplate(name string) *template.Template {
	if name == "" {
		name = "default"
	}
	return r.templates[name]
}

// createDefaultTemplate creates a default HTML template
func (r *HTMLRenderer) createDefaultTemplate() *template.Template {
	const defaultTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>{{.Metadata.Title}}</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { border-bottom: 2px solid #333; padding-bottom: 20px; }
        .summary { background: #f5f5f5; padding: 20px; margin: 20px 0; }
        .finding { border: 1px solid #ddd; padding: 15px; margin: 10px 0; }
        .critical { border-color: #d32f2f; }
        .high { border-color: #f57c00; }
        .medium { border-color: #fbc02d; }
        .low { border-color: #388e3c; }
        .info { border-color: #1976d2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>{{.Metadata.Title}}</h1>
        <p>Generated: {{.Metadata.CreatedAt.Format "2006-01-02 15:04:05"}}</p>
    </div>
    
    <div class="summary">
        <h2>Summary</h2>
        <p>Total Findings: {{.Summary.TotalFindings}}</p>
        <p>Risk Score: {{printf "%.2f" .Summary.RiskScore}}</p>
    </div>
    
    <div class="findings">
        <h2>Findings</h2>
        {{range .Findings}}
        <div class="finding {{.Severity | lower}}">
            <h3>{{.Title}}</h3>
            <p><strong>Severity:</strong> {{.Severity}}</p>
            <p><strong>Category:</strong> {{.Category}}</p>
            <p>{{.Description}}</p>
        </div>
        {{end}}
    </div>
</body>
</html>
`
	tmpl, _ := template.New("default").Funcs(template.FuncMap{
		"lower": strings.ToLower,
	}).Parse(defaultTemplate)
	return tmpl
}

// prepareTemplateData prepares data for template rendering
func (r *HTMLRenderer) prepareTemplateData(report *SecurityReport, options RenderOptions) interface{} {
	// Add any additional data processing here
	return report
}

// MarkdownRenderer renders reports as Markdown
type MarkdownRenderer struct{}

// Render renders a report as Markdown
func (r *MarkdownRenderer) Render(report *SecurityReport, options RenderOptions) ([]byte, error) {
	var buf bytes.Buffer
	
	// Write header
	fmt.Fprintf(&buf, "# %s\n\n", report.Metadata.Title)
	fmt.Fprintf(&buf, "**Report ID:** %s  \n", report.Metadata.ID)
	fmt.Fprintf(&buf, "**Created:** %s  \n", report.Metadata.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(&buf, "**Version:** %s  \n\n", report.Metadata.Version)

	// Write summary
	fmt.Fprintf(&buf, "## Executive Summary\n\n")
	fmt.Fprintf(&buf, "Total findings: **%d**  \n", report.Summary.TotalFindings)
	fmt.Fprintf(&buf, "Risk score: **%.2f**  \n\n", report.Summary.RiskScore)

	// Write severity breakdown
	fmt.Fprintf(&buf, "### Severity Breakdown\n\n")
	for severity, count := range report.Summary.SeverityBreakdown {
		fmt.Fprintf(&buf, "- %s: %d\n", severity, count)
	}
	fmt.Fprintf(&buf, "\n")

	// Write findings
	fmt.Fprintf(&buf, "## Findings\n\n")
	
	// Sort findings by severity
	findings := make([]Finding, len(report.Findings))
	copy(findings, report.Findings)
	sort.Slice(findings, func(i, j int) bool {
		return r.severityWeight(findings[i].Severity) > r.severityWeight(findings[j].Severity)
	})

	for i, finding := range findings {
		fmt.Fprintf(&buf, "### %d. %s\n\n", i+1, finding.Title)
		fmt.Fprintf(&buf, "**Severity:** %s  \n", finding.Severity)
		fmt.Fprintf(&buf, "**Category:** %s  \n", finding.Category)
		if len(finding.OWASPMapping) > 0 {
			fmt.Fprintf(&buf, "**OWASP:** %s  \n", strings.Join(finding.OWASPMapping, ", "))
		}
		fmt.Fprintf(&buf, "**Confidence:** %.0f%%  \n\n", finding.Confidence*100)
		
		fmt.Fprintf(&buf, "**Description:**  \n%s\n\n", finding.Description)
		
		if finding.Evidence.Request != "" || finding.Evidence.Response != "" {
			fmt.Fprintf(&buf, "**Evidence:**  \n")
			if finding.Evidence.Request != "" {
				fmt.Fprintf(&buf, "```\nRequest: %s\n```\n", finding.Evidence.Request)
			}
			if finding.Evidence.Response != "" {
				fmt.Fprintf(&buf, "```\nResponse: %s\n```\n", finding.Evidence.Response)
			}
			fmt.Fprintf(&buf, "\n")
		}
		
		if finding.Remediation.Summary != "" {
			fmt.Fprintf(&buf, "**Remediation:**  \n%s\n\n", finding.Remediation.Summary)
			if len(finding.Remediation.Steps) > 0 {
				for j, step := range finding.Remediation.Steps {
					fmt.Fprintf(&buf, "%d. %s\n", j+1, step)
				}
				fmt.Fprintf(&buf, "\n")
			}
		}
		
		if len(finding.References) > 0 {
			fmt.Fprintf(&buf, "**References:**  \n")
			for _, ref := range finding.References {
				fmt.Fprintf(&buf, "- [%s](%s)\n", ref.Title, ref.URL)
			}
			fmt.Fprintf(&buf, "\n")
		}
		
		fmt.Fprintf(&buf, "---\n\n")
	}

	// Write compliance section if present
	if report.Compliance != nil {
		fmt.Fprintf(&buf, "## Compliance Status\n\n")
		for framework, status := range report.Compliance {
			fmt.Fprintf(&buf, "### %s\n\n", framework)
			fmt.Fprintf(&buf, "- Compliant: %v\n", status.Compliant)
			fmt.Fprintf(&buf, "- Coverage: %.1f%%\n", status.Coverage*100)
			if len(status.Gaps) > 0 {
				fmt.Fprintf(&buf, "- Gaps: %s\n", strings.Join(status.Gaps, ", "))
			}
			fmt.Fprintf(&buf, "\n")
		}
	}

	return buf.Bytes(), nil
}

// GetContentType returns the content type for Markdown
func (r *MarkdownRenderer) GetContentType() string {
	return "text/markdown"
}

// GetFileExtension returns the file extension for Markdown
func (r *MarkdownRenderer) GetFileExtension() string {
	return ".md"
}

// severityWeight returns a weight for severity sorting
func (r *MarkdownRenderer) severityWeight(severity string) int {
	weights := map[string]int{
		"critical": 5,
		"high":     4,
		"medium":   3,
		"low":      2,
		"info":     1,
	}
	if weight, ok := weights[strings.ToLower(severity)]; ok {
		return weight
	}
	return 0
}

// StreamingRenderer supports streaming large reports
type StreamingRenderer interface {
	StartRender(w io.Writer, metadata ReportMetadata, options RenderOptions) error
	RenderFinding(w io.Writer, finding Finding) error
	FinishRender(w io.Writer, summary SecurityReportSummary) error
}

// ReportBuilder builds reports with various options
type ReportBuilder struct {
	metadata  ReportMetadata
	findings  []Finding
	filters   []func(Finding) bool
	sorters   []func([]Finding)
	enrichers []func(*Finding)
}

// NewReportBuilder creates a new report builder
func NewReportBuilder() *ReportBuilder {
	return &ReportBuilder{
		findings:  []Finding{},
		filters:   []func(Finding) bool{},
		sorters:   []func([]Finding){},
		enrichers: []func(*Finding){},
	}
}

// WithMetadata sets report metadata
func (rb *ReportBuilder) WithMetadata(metadata ReportMetadata) *ReportBuilder {
	rb.metadata = metadata
	return rb
}

// AddFinding adds a finding to the report
func (rb *ReportBuilder) AddFinding(finding Finding) *ReportBuilder {
	rb.findings = append(rb.findings, finding)
	return rb
}

// AddFilter adds a filter function
func (rb *ReportBuilder) AddFilter(filter func(Finding) bool) *ReportBuilder {
	rb.filters = append(rb.filters, filter)
	return rb
}

// Build builds the final report
func (rb *ReportBuilder) Build() *SecurityReport {
	// Apply filters
	filtered := []Finding{}
	for _, finding := range rb.findings {
		include := true
		for _, filter := range rb.filters {
			if !filter(finding) {
				include = false
				break
			}
		}
		if include {
			filtered = append(filtered, finding)
		}
	}

	// Apply enrichers
	for i := range filtered {
		for _, enricher := range rb.enrichers {
			enricher(&filtered[i])
		}
	}

	// Apply sorters
	for _, sorter := range rb.sorters {
		sorter(filtered)
	}

	// Calculate summary and statistics
	summary := rb.calculateSummary(filtered)
	stats := rb.calculateStatistics(filtered)

	return &SecurityReport{
		Metadata:   rb.metadata,
		Summary:    summary,
		Findings:   filtered,
		Statistics: stats,
	}
}

// calculateSummary calculates report summary
func (rb *ReportBuilder) calculateSummary(findings []Finding) SecurityReportSummary {
	summary := SecurityReportSummary{
		TotalFindings:     len(findings),
		SeverityBreakdown: make(map[string]int),
		CategoryBreakdown: make(map[string]int),
	}

	for _, finding := range findings {
		summary.SeverityBreakdown[finding.Severity]++
		summary.CategoryBreakdown[finding.Category]++
	}

	// Calculate risk score
	weights := map[string]float64{
		"critical": 10.0,
		"high":     5.0,
		"medium":   2.0,
		"low":      1.0,
		"info":     0.5,
	}

	for severity, count := range summary.SeverityBreakdown {
		if weight, ok := weights[strings.ToLower(severity)]; ok {
			summary.RiskScore += weight * float64(count)
		}
	}

	return summary
}

// calculateStatistics calculates report statistics
func (rb *ReportBuilder) calculateStatistics(findings []Finding) ReportStatistics {
	stats := ReportStatistics{
		StartTime:         rb.metadata.CreatedAt,
		EndTime:           time.Now(),
		FindingsPerMinute: 0,
		AverageConfidence: 0,
		TopCategories:     []CategoryCount{},
	}

	if len(findings) > 0 {
		duration := stats.EndTime.Sub(stats.StartTime).Minutes()
		if duration > 0 {
			stats.FindingsPerMinute = float64(len(findings)) / duration
		}

		totalConfidence := 0.0
		for _, finding := range findings {
			totalConfidence += finding.Confidence
		}
		stats.AverageConfidence = totalConfidence / float64(len(findings))
	}

	// Calculate top categories
	categoryMap := make(map[string]int)
	for _, finding := range findings {
		categoryMap[finding.Category]++
	}
	
	for cat, count := range categoryMap {
		stats.TopCategories = append(stats.TopCategories, CategoryCount{
			Category: cat,
			Count:    count,
		})
	}
	
	sort.Slice(stats.TopCategories, func(i, j int) bool {
		return stats.TopCategories[i].Count > stats.TopCategories[j].Count
	})

	return stats
}