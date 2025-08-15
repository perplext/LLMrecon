package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Dashboard provides a comprehensive view of scan results
type Dashboard struct {
	terminal     *Terminal
	style        *DashboardStyle
	refreshRate  time.Duration
	widgets      []Widget
	layout       *Layout
}

// DashboardStyle defines the visual styling for the dashboard
type DashboardStyle struct {
	Border      lipgloss.Style
	Title       lipgloss.Style
	Widget      lipgloss.Style
	Metric      lipgloss.Style
	Chart       lipgloss.Style
	Critical    lipgloss.Style
	Warning     lipgloss.Style
	Success     lipgloss.Style
	Info        lipgloss.Style
}

// Widget represents a dashboard component
type Widget interface {
	Render(width, height int) string
	Update(data interface{})
	GetTitle() string

// Layout manages widget positioning
type Layout struct {
	Rows    []Row
	Columns int

// Row represents a row of widgets
type Row struct {
	Widgets []Widget
	Heights []int

// NewDashboard creates a new dashboard
func NewDashboard(terminal *Terminal) *Dashboard {
	return &Dashboard{
		terminal:    terminal,
		style:       newDashboardStyle(),
		refreshRate: 1 * time.Second,
		widgets:     []Widget{},
		layout:      &Layout{Columns: 12},
	}

func newDashboardStyle() *DashboardStyle {
	return &DashboardStyle{
		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229")),
		Widget: lipgloss.NewStyle().
			Padding(1).
			Margin(0, 1),
		Metric: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")),
		Chart: lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")),
		Critical: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),
		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")),
		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")),
		Info: lipgloss.NewStyle().
			Foreground(lipgloss.Color("69")),
	}

// ScanDashboard displays real-time scan progress and results
func (d *Dashboard) ScanDashboard(scanID string) error {
	// Clear screen
	d.terminal.Clear()
	
	// Create widgets
	d.widgets = []Widget{
		NewScanOverviewWidget(d.style),
		NewVulnerabilityChartWidget(d.style),
		NewTestProgressWidget(d.style),
		NewRecentFindingsWidget(d.style),
		NewPerformanceMetricsWidget(d.style),
		NewComplianceStatusWidget(d.style),
	}
	
	// Define layout
	d.layout.Rows = []Row{
		{Widgets: []Widget{d.widgets[0]}, Heights: []int{8}},           // Overview
		{Widgets: []Widget{d.widgets[1], d.widgets[2]}, Heights: []int{10, 10}}, // Charts
		{Widgets: []Widget{d.widgets[3]}, Heights: []int{12}},          // Findings
		{Widgets: []Widget{d.widgets[4], d.widgets[5]}, Heights: []int{8, 8}},   // Metrics
	}
	
	// Render loop
	ticker := time.NewTicker(d.refreshRate)
	defer ticker.Stop()
	
	for {
		// Update widget data
		d.updateWidgets(scanID)
		
		// Render dashboard
		d.render()
		
		// Check for user input
		select {
		case <-ticker.C:
			continue
		default:
			if d.terminal.HasInput() {
				key := d.terminal.ReadKey()
				if key == "q" || key == "Q" {
					return nil
				}
			}
		}
	}

// render displays the complete dashboard
func (d *Dashboard) render() {
	d.terminal.Clear()
	
	// Dashboard header
	header := d.style.Title.Render("🛡️  LLMrecon Security Dashboard")
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	headerLine := fmt.Sprintf("%s %s", header, d.style.Info.Render(timestamp))
	fmt.Println(headerLine)
	fmt.Println(strings.Repeat("─", 80))
	
	// Render rows
	for _, row := range d.layout.Rows {
		d.renderRow(row)
	}
	
	// Footer
	fmt.Println(strings.Repeat("─", 80))
	fmt.Println(d.style.Info.Render("Press 'q' to exit | Auto-refresh: " + d.refreshRate.String()))

// renderRow renders a single row of widgets
func (d *Dashboard) renderRow(row Row) {
	// Calculate widget widths
	widgetCount := len(row.Widgets)
	totalWidth := 80 // Terminal width
	widgetWidth := totalWidth / widgetCount
	
	// Render each widget
	renderedWidgets := []string{}
	maxHeight := 0
	
	for i, widget := range row.Widgets {
		height := row.Heights[i]
		rendered := widget.Render(widgetWidth-2, height)
		renderedWidgets = append(renderedWidgets, rendered)
		
		lines := strings.Split(rendered, "\n")
		if len(lines) > maxHeight {
			maxHeight = len(lines)
		}
	}
	
	// Print side by side
	for lineIdx := 0; lineIdx < maxHeight; lineIdx++ {
		line := ""
		for _, rendered := range renderedWidgets {
			lines := strings.Split(rendered, "\n")
			if lineIdx < len(lines) {
				line += lines[lineIdx]
			} else {
				line += strings.Repeat(" ", widgetWidth-2)
			}
			line += "  " // Spacing between widgets
		}
		fmt.Println(line)
	}

// updateWidgets updates all widget data
func (d *Dashboard) updateWidgets(scanID string) {
	// In a real implementation, this would fetch live data
	// For now, we'll use mock data
	
	scanData := &ScanData{
		ID:          scanID,
		Status:      "Running",
		Progress:    65,
		StartTime:   time.Now().Add(-10 * time.Minute),
		TestsTotal:  150,
		TestsPassed: 89,
		TestsFailed: 8,
		Findings: []Finding{
			{Severity: "Critical", Category: "Prompt Injection", Count: 2},
			{Severity: "High", Category: "Data Leakage", Count: 5},
			{Severity: "Medium", Category: "Output Handling", Count: 8},
			{Severity: "Low", Category: "Configuration", Count: 12},
		},
	}
	
	// Update each widget
	for _, widget := range d.widgets {
		widget.Update(scanData)
	}

// Widget implementations

// ScanOverviewWidget shows high-level scan status
type ScanOverviewWidget struct {
	style *DashboardStyle
	data  *ScanData

func NewScanOverviewWidget(style *DashboardStyle) *ScanOverviewWidget {
	return &ScanOverviewWidget{style: style}

func (w *ScanOverviewWidget) GetTitle() string {
	return "Scan Overview"

func (w *ScanOverviewWidget) Update(data interface{}) {
	if scanData, ok := data.(*ScanData); ok {
		w.data = scanData
	}

func (w *ScanOverviewWidget) Render(width, height int) string {
	if w.data == nil {
		return "No data"
	}
	
	box := w.style.Border.Width(width).Height(height)
	
	content := fmt.Sprintf(
		"%s\n\n%s %s\n%s %d%%\n%s %s\n%s %d/%d\n%s %d",
		w.style.Title.Render("📊 " + w.GetTitle()),
		w.style.Info.Render("Status:"),
		w.getStatusStyle(w.data.Status).Render(w.data.Status),
		w.style.Info.Render("Progress:"),
		w.data.Progress,
		w.style.Info.Render("Duration:"),
		time.Since(w.data.StartTime).Round(time.Second).String(),
		w.style.Info.Render("Tests:"),
		w.data.TestsPassed+w.data.TestsFailed,
		w.data.TestsTotal,
		w.style.Info.Render("Findings:"),
		w.getTotalFindings(),
	)
	
	return box.Render(content)

func (w *ScanOverviewWidget) getStatusStyle(status string) lipgloss.Style {
	switch status {
	case "Running":
		return w.style.Warning
	case "Completed":
		return w.style.Success
	case "Failed":
		return w.style.Critical
	default:
		return w.style.Info
	}

func (w *ScanOverviewWidget) getTotalFindings() int {
	total := 0
	for _, f := range w.data.Findings {
		total += f.Count
	}
	return total

// VulnerabilityChartWidget displays vulnerability distribution
type VulnerabilityChartWidget struct {
	style *DashboardStyle
	data  *ScanData

func NewVulnerabilityChartWidget(style *DashboardStyle) *VulnerabilityChartWidget {
	return &VulnerabilityChartWidget{style: style}

func (w *VulnerabilityChartWidget) GetTitle() string {
	return "Vulnerability Distribution"

func (w *VulnerabilityChartWidget) Update(data interface{}) {
	if scanData, ok := data.(*ScanData); ok {
		w.data = scanData
	}

func (w *VulnerabilityChartWidget) Render(width, height int) string {
	if w.data == nil {
		return "No data"
	}
	
	box := w.style.Border.Width(width).Height(height)
	
	content := w.style.Title.Render("🔍 " + w.GetTitle()) + "\n\n"
	
	// Create bar chart
	maxCount := 0
	for _, f := range w.data.Findings {
		if f.Count > maxCount {
			maxCount = f.Count
		}
	}
	
	barWidth := width - 20
	for _, f := range w.data.Findings {
		bar := w.renderBar(f, maxCount, barWidth)
		content += bar + "\n"
	}
	
	return box.Render(content)

func (w *VulnerabilityChartWidget) renderBar(finding Finding, maxCount, maxWidth int) string {
	label := fmt.Sprintf("%-8s", finding.Severity)
	
	barLength := 0
	if maxCount > 0 {
		barLength = (finding.Count * maxWidth) / maxCount
	}
	
	bar := strings.Repeat("█", barLength)
	count := fmt.Sprintf(" %d", finding.Count)
	
	style := w.style.Info
	switch finding.Severity {
	case "Critical":
		style = w.style.Critical
	case "High":
		style = w.style.Warning
	case "Medium":
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	case "Low":
		style = w.style.Success
	}
	
	return style.Render(label + bar + count)

// TestProgressWidget shows test execution progress
type TestProgressWidget struct {
	style *DashboardStyle
	data  *ScanData

func NewTestProgressWidget(style *DashboardStyle) *TestProgressWidget {
	return &TestProgressWidget{style: style}

func (w *TestProgressWidget) GetTitle() string {
	return "Test Progress"

func (w *TestProgressWidget) Update(data interface{}) {
	if scanData, ok := data.(*ScanData); ok {
		w.data = scanData
	}

func (w *TestProgressWidget) Render(width, height int) string {
	if w.data == nil {
		return "No data"
	}
	
	box := w.style.Border.Width(width).Height(height)
	
	content := w.style.Title.Render("⚡ " + w.GetTitle()) + "\n\n"
	
	// Progress bar
	progress := w.data.Progress
	barWidth := width - 10
	filled := (progress * barWidth) / 100
	empty := barWidth - filled
	
	progressBar := w.style.Success.Render(strings.Repeat("█", filled)) +
		w.style.Info.Render(strings.Repeat("░", empty))
	
	content += fmt.Sprintf("%s %d%%\n\n", progressBar, progress)
	
	// Test statistics
	content += fmt.Sprintf(
		"%s %d\n%s %d\n%s %d\n%s %d",
		w.style.Info.Render("Total Tests:"),
		w.data.TestsTotal,
		w.style.Success.Render("Passed:"),
		w.data.TestsPassed,
		w.style.Critical.Render("Failed:"),
		w.data.TestsFailed,
		w.style.Warning.Render("Pending:"),
		w.data.TestsTotal-w.data.TestsPassed-w.data.TestsFailed,
	)
	
	return box.Render(content)

// RecentFindingsWidget displays recent vulnerability findings
type RecentFindingsWidget struct {
	style    *DashboardStyle
	findings []DetailedFinding

func NewRecentFindingsWidget(style *DashboardStyle) *RecentFindingsWidget {
	return &RecentFindingsWidget{style: style}

func (w *RecentFindingsWidget) GetTitle() string {
	return "Recent Findings"

func (w *RecentFindingsWidget) Update(data interface{}) {
	// In real implementation, would fetch recent findings
	w.findings = []DetailedFinding{
		{
			Time:     time.Now().Add(-2 * time.Minute),
			Severity: "Critical",
			Type:     "Prompt Injection",
			Message:  "System prompt override detected in chat endpoint",
			TestCase: "test-prompt-override-001",
		},
		{
			Time:     time.Now().Add(-5 * time.Minute),
			Severity: "High",
			Type:     "Data Leakage",
			Message:  "Sensitive training data exposed through model inversion",
			TestCase: "test-data-extraction-042",
		},
		{
			Time:     time.Now().Add(-8 * time.Minute),
			Severity: "Medium",
			Type:     "Output Handling",
			Message:  "Insufficient output sanitization allows XSS",
			TestCase: "test-xss-output-017",
		},
	}

func (w *RecentFindingsWidget) Render(width, height int) string {
	box := w.style.Border.Width(width).Height(height)
	
	content := w.style.Title.Render("🚨 " + w.GetTitle()) + "\n\n"
	
	for _, finding := range w.findings {
		timeStr := finding.Time.Format("15:04:05")
		
		severityStyle := w.style.Info
		icon := "ℹ"
		switch finding.Severity {
		case "Critical":
			severityStyle = w.style.Critical
			icon = "🔴"
		case "High":
			severityStyle = w.style.Warning
			icon = "🟠"
		case "Medium":
			severityStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
			icon = "🟡"
		case "Low":
			severityStyle = w.style.Success
			icon = "🟢"
		}
		
		content += fmt.Sprintf(
			"%s %s %s %s\n  %s\n  %s\n\n",
			w.style.Info.Render(timeStr),
			icon,
			severityStyle.Render(finding.Severity),
			w.style.Info.Render(finding.Type),
			finding.Message,
			w.style.Info.Render("Test: "+finding.TestCase),
		)
	}
	
	return box.Render(content)

// PerformanceMetricsWidget displays performance statistics
type PerformanceMetricsWidget struct {
	style   *DashboardStyle
	metrics *PerformanceMetrics

func NewPerformanceMetricsWidget(style *DashboardStyle) *PerformanceMetricsWidget {
	return &PerformanceMetricsWidget{style: style}

func (w *PerformanceMetricsWidget) GetTitle() string {
	return "Performance Metrics"

func (w *PerformanceMetricsWidget) Update(data interface{}) {
	w.metrics = &PerformanceMetrics{
		RequestsPerSecond: 45.7,
		AverageLatency:    234,
		P95Latency:        567,
		P99Latency:        892,
		ErrorRate:         0.02,
		Throughput:        "1.2 MB/s",
	}

func (w *PerformanceMetricsWidget) Render(width, height int) string {
	if w.metrics == nil {
		return "No data"
	}
	
	box := w.style.Border.Width(width).Height(height)
	
	content := w.style.Title.Render("📈 " + w.GetTitle()) + "\n\n"
	
	content += fmt.Sprintf(
		"%s %.1f req/s\n%s %dms\n%s %dms\n%s %.1f%%\n%s %s",
		w.style.Info.Render("Requests/sec:"),
		w.metrics.RequestsPerSecond,
		w.style.Info.Render("Avg Latency:"),
		w.metrics.AverageLatency,
		w.style.Info.Render("P95 Latency:"),
		w.metrics.P95Latency,
		w.style.Info.Render("Error Rate:"),
		w.metrics.ErrorRate*100,
		w.style.Info.Render("Throughput:"),
		w.metrics.Throughput,
	)
	
	return box.Render(content)

// ComplianceStatusWidget shows compliance mapping
type ComplianceStatusWidget struct {
	style      *DashboardStyle
	compliance map[string]ComplianceStatus

func NewComplianceStatusWidget(style *DashboardStyle) *ComplianceStatusWidget {
	return &ComplianceStatusWidget{style: style}

func (w *ComplianceStatusWidget) GetTitle() string {
	return "Compliance Status"

func (w *ComplianceStatusWidget) Update(data interface{}) {
	w.compliance = map[string]ComplianceStatus{
		"OWASP LLM01": {Tested: 15, Passed: 12, Failed: 3},
		"OWASP LLM02": {Tested: 8, Passed: 8, Failed: 0},
		"OWASP LLM03": {Tested: 10, Passed: 7, Failed: 3},
		"ISO 42001":   {Tested: 25, Passed: 23, Failed: 2},
	}

func (w *ComplianceStatusWidget) Render(width, height int) string {
	box := w.style.Border.Width(width).Height(height)
	
	content := w.style.Title.Render("📋 " + w.GetTitle()) + "\n\n"
	
	for standard, status := range w.compliance {
		passRate := float64(status.Passed) / float64(status.Tested) * 100
		
		statusStyle := w.style.Success
		if passRate < 100 {
			statusStyle = w.style.Warning
		}
		if passRate < 80 {
			statusStyle = w.style.Critical
		}
		
		content += fmt.Sprintf(
			"%s: %s (%.0f%%)\n",
			w.style.Info.Render(standard),
			statusStyle.Render(fmt.Sprintf("%d/%d", status.Passed, status.Tested)),
			passRate,
		)
	}
	
	return box.Render(content)

// Data structures

type ScanData struct {
	ID          string
	Status      string
	Progress    int
	StartTime   time.Time
	TestsTotal  int
	TestsPassed int
	TestsFailed int
	Findings    []Finding

type Finding struct {
	Severity string
	Category string
	Count    int
}

type DetailedFinding struct {
	Time     time.Time
	Severity string
	Type     string
	Message  string
	TestCase string
}

type PerformanceMetrics struct {
	RequestsPerSecond float64
	AverageLatency    int
	P95Latency        int
	P99Latency        int
	ErrorRate         float64
	Throughput        string

type ComplianceStatus struct {
	Tested int
	Passed int
	Failed int
}

// Summary Dashboard for completed scans

// SummaryDashboard displays a summary of completed scan results
func (d *Dashboard) SummaryDashboard(scanResults *ScanResults) error {
	d.terminal.Clear()
	
	// Header
	d.terminal.HeaderBox("Scan Results Summary - " + scanResults.ID)
	
	// Executive Summary
	d.showExecutiveSummary(scanResults)
	
	// Vulnerability Breakdown
	d.showVulnerabilityBreakdown(scanResults)
	
	// Top Risks
	d.showTopRisks(scanResults)
	
	// Compliance Mapping
	d.showComplianceMapping(scanResults)
	
	// Recommendations
	d.showRecommendations(scanResults)
	
	// Export Options
	d.showExportOptions(scanResults)
	
	return nil

func (d *Dashboard) showExecutiveSummary(results *ScanResults) {
	d.terminal.Section("Executive Summary")
	
	summary := fmt.Sprintf(
		`Scan Duration: %s
Total Tests: %d
Vulnerabilities Found: %d
Risk Score: %.1f/10
Compliance: %s`,
		results.Duration,
		results.TotalTests,
		results.VulnerabilityCount,
		results.RiskScore,
		results.ComplianceStatus,
	)
	
	d.terminal.Box("Overview", summary)

func (d *Dashboard) showVulnerabilityBreakdown(results *ScanResults) {
	d.terminal.Section("Vulnerability Breakdown")
	
	// Create visual representation
	for _, vuln := range results.Vulnerabilities {
		bar := strings.Repeat("█", vuln.Count)
		d.terminal.Printf("%s%-10s %s %d%s\n",
			d.getSeverityIcon(vuln.Severity),
			vuln.Severity,
			d.getSeverityColor(vuln.Severity).Render(bar),
			vuln.Count,
			d.style.Info.Render(" vulnerabilities"),
		)
	}

func (d *Dashboard) showTopRisks(results *ScanResults) {
	d.terminal.Section("Top Security Risks")
	
	for i, risk := range results.TopRisks[:min(5, len(results.TopRisks))] {
		d.terminal.Printf("%d. %s%s%s\n   Impact: %s | Likelihood: %s\n   %s\n\n",
			i+1,
			d.getSeverityIcon(risk.Severity),
			d.getSeverityColor(risk.Severity).Render(risk.Title),
			d.style.Info.Render(" ("+risk.Category+")"),
			risk.Impact,
			risk.Likelihood,
			d.style.Info.Render(risk.Description),
		)
	}

func (d *Dashboard) showComplianceMapping(results *ScanResults) {
	d.terminal.Section("Compliance Mapping")
	
	table := [][]string{
		{"Standard", "Coverage", "Pass Rate", "Status"},
	}
	
	for _, comp := range results.ComplianceResults {
		status := "✅ Compliant"
		if comp.PassRate < 100 {
			status = "⚠️  Partial"
		}
		if comp.PassRate < 80 {
			status = "❌ Non-compliant"
		}
		
		table = append(table, []string{
			comp.Standard,
			fmt.Sprintf("%.0f%%", comp.Coverage),
			fmt.Sprintf("%.0f%%", comp.PassRate),
			status,
		})
	}
	
	d.terminal.Table(table)

func (d *Dashboard) showRecommendations(results *ScanResults) {
	d.terminal.Section("Key Recommendations")
	
	for i, rec := range results.Recommendations[:min(5, len(results.Recommendations))] {
		d.terminal.Printf("%d. %s%s%s\n   Priority: %s | Effort: %s\n\n",
			i+1,
			"💡 ",
			d.style.Info.Render(rec.Title),
			d.style.Info.Render(" - "+rec.Category),
			d.getPriorityColor(rec.Priority).Render(rec.Priority),
			rec.Effort,
		)
	}

func (d *Dashboard) showExportOptions(results *ScanResults) {
	d.terminal.Section("Export Options")
	
	options := []string{
		"1. Generate PDF Report",
		"2. Export as JSON",
		"3. Create HTML Dashboard",
		"4. Generate Compliance Report",
		"5. Export to JIRA/GitHub Issues",
		"6. Return to Main Menu",
	}
	
	for _, opt := range options {
		d.terminal.Info(opt)
	}

// Helper methods

func (d *Dashboard) getSeverityIcon(severity string) string {
	switch severity {
	case "Critical":
		return "🔴 "
	case "High":
		return "🟠 "
	case "Medium":
		return "🟡 "
	case "Low":
		return "🟢 "
	default:
		return "⚪ "
	}

func (d *Dashboard) getSeverityColor(severity string) lipgloss.Style {
	switch severity {
	case "Critical":
		return d.style.Critical
	case "High":
		return d.style.Warning
	case "Medium":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	case "Low":
		return d.style.Success
	default:
		return d.style.Info
	}

func (d *Dashboard) getPriorityColor(priority string) lipgloss.Style {
	switch priority {
	case "Immediate":
		return d.style.Critical
	case "High":
		return d.style.Warning
	case "Medium":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	case "Low":
		return d.style.Success
	default:
		return d.style.Info
	}

// ScanResults represents completed scan results
type ScanResults struct {
	ID                 string
	Duration           string
	TotalTests         int
	VulnerabilityCount int
	RiskScore          float64
	ComplianceStatus   string
	Vulnerabilities    []VulnerabilitySummary
	TopRisks           []RiskItem
	ComplianceResults  []ComplianceResult
	Recommendations    []Recommendation

type VulnerabilitySummary struct {
	Severity string
	Count    int

type RiskItem struct {
	Title       string
	Category    string
	Severity    string
	Impact      string
	Likelihood  string
	Description string

type ComplianceResult struct {
	Standard string
	Coverage float64
	PassRate float64
}

type Recommendation struct {
	Title    string
	Category string
	Priority string
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
}
