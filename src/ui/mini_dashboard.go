package ui

import (
	"fmt"
	"strings"
	"time"
)

// MiniDashboard provides compact status views
type MiniDashboard struct {
	terminal *Terminal
	style    *DashboardStyle
}

// NewMiniDashboard creates a new mini dashboard
func NewMiniDashboard(terminal *Terminal) *MiniDashboard {
	return &MiniDashboard{
		terminal: terminal,
		style:    newDashboardStyle(),
	}
}

// ShowQuickStatus displays a quick status overview
func (md *MiniDashboard) ShowQuickStatus() {
	md.terminal.Clear()
	
	// Header
	md.terminal.HeaderBox("Quick Status Overview")
	
	// System Status
	md.showSystemStatus()
	
	// Recent Activity
	md.showRecentActivity()
	
	// Active Scans
	md.showActiveScans()
	
	// Alerts
	md.showAlerts()
}

// ShowScanProgress displays inline scan progress
func (md *MiniDashboard) ShowScanProgress(scan *ActiveScan) {
	// Clear previous line and move cursor up
	fmt.Print("\033[1A\033[2K")
	
	// Progress bar
	width := 40
	filled := (scan.Progress * width) / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	
	// Status line
	status := fmt.Sprintf(
		"[%s] %s %d%% | Tests: %d/%d | Findings: %d | ETA: %s",
		md.style.Chart.Render(bar),
		scan.Status,
		scan.Progress,
		scan.TestsCompleted,
		scan.TestsTotal,
		scan.FindingsCount,
		scan.ETA,
	)
	
	// Color based on status
	switch scan.Status {
	case "Running":
		fmt.Print(md.style.Warning.Render(status))
	case "Completed":
		fmt.Print(md.style.Success.Render(status))
	case "Failed":
		fmt.Print(md.style.Critical.Render(status))
	default:
		fmt.Print(status)
	}
}

// ShowCompactResults displays scan results in a compact format
func (md *MiniDashboard) ShowCompactResults(results *CompactResults) {
	md.terminal.Section("Scan Complete: " + results.ScanID)
	
	// Summary line
	summary := fmt.Sprintf(
		"Duration: %s | Tests: %d | ",
		results.Duration,
		results.TotalTests,
	)
	
	// Vulnerability counts with colors
	vulnCounts := []string{}
	if results.Critical > 0 {
		vulnCounts = append(vulnCounts, md.style.Critical.Render(fmt.Sprintf("%d Critical", results.Critical)))
	}
	if results.High > 0 {
		vulnCounts = append(vulnCounts, md.style.Warning.Render(fmt.Sprintf("%d High", results.High)))
	}
	if results.Medium > 0 {
		vulnCounts = append(vulnCounts, md.style.Info.Render(fmt.Sprintf("%d Medium", results.Medium)))
	}
	if results.Low > 0 {
		vulnCounts = append(vulnCounts, md.style.Success.Render(fmt.Sprintf("%d Low", results.Low)))
	}
	
	if len(vulnCounts) > 0 {
		summary += "Vulnerabilities: " + strings.Join(vulnCounts, ", ")
	} else {
		summary += md.style.Success.Render("No vulnerabilities found")
	}
	
	fmt.Println(summary)
	
	// Quick actions
	fmt.Println("\nQuick Actions:")
	fmt.Println("  • View full report: LLMrecon report view " + results.ScanID)
	fmt.Println("  • Export results: LLMrecon report export " + results.ScanID)
	fmt.Println("  • Re-run scan: LLMrecon scan --replay " + results.ScanID)
}

// ShowLiveMetrics displays real-time metrics
func (md *MiniDashboard) ShowLiveMetrics() {
	metrics := &LiveMetrics{
		ActiveScans:       3,
		QueuedScans:       7,
		RequestsPerSecond: 145.7,
		AverageLatency:    89,
		SystemLoad:        0.72,
		MemoryUsage:       62.3,
	}
	
	// Create sparkline for RPS
	sparkline := md.createSparkline([]float64{120, 135, 142, 138, 145, 149, 145})
	
	md.terminal.Box("Live System Metrics",
		fmt.Sprintf(`Active Scans: %d | Queued: %d
RPS: %.1f %s
Latency: %dms | Load: %.2f | Memory: %.1f%%`,
			metrics.ActiveScans,
			metrics.QueuedScans,
			metrics.RequestsPerSecond,
			sparkline,
			metrics.AverageLatency,
			metrics.SystemLoad,
			metrics.MemoryUsage,
		),
	)
}

// ShowTestMatrix displays a test coverage matrix
func (md *MiniDashboard) ShowTestMatrix(matrix *TestMatrix) {
	md.terminal.Section("Test Coverage Matrix")
	
	// Header
	fmt.Print("         ")
	for _, target := range matrix.Targets {
		fmt.Printf("%-10s ", truncate(target, 10))
	}
	fmt.Println()
	
	// Rows
	for _, category := range matrix.Categories {
		fmt.Printf("%-8s ", truncate(category.Name, 8))
		
		for _, target := range matrix.Targets {
			coverage := matrix.GetCoverage(category.Name, target)
			cell := md.getCoverageCell(coverage)
			fmt.Print(cell + " ")
		}
		fmt.Println()
	}
	
	// Legend
	fmt.Println("\nLegend: " +
		md.style.Success.Render("█ >90%") + " " +
		md.style.Warning.Render("█ 50-90%") + " " +
		md.style.Critical.Render("█ <50%") + " " +
		md.style.Info.Render("░ Not tested"))
}

// Helper methods

func (md *MiniDashboard) showSystemStatus() {
	status := &SystemStatus{
		APIStatus:      "Operational",
		DatabaseStatus: "Operational",
		QueueStatus:    "Degraded",
		LastCheck:      time.Now(),
	}
	
	md.terminal.Subsection("System Status")
	
	statuses := []struct {
		Name   string
		Status string
	}{
		{"API", status.APIStatus},
		{"Database", status.DatabaseStatus},
		{"Queue", status.QueueStatus},
	}
	
	for _, s := range statuses {
		icon := "✅"
		style := md.style.Success
		
		switch s.Status {
		case "Degraded":
			icon = "⚠️"
			style = md.style.Warning
		case "Down":
			icon = "❌"
			style = md.style.Critical
		}
		
		fmt.Printf("%s %s: %s\n", icon, s.Name, style.Render(s.Status))
	}
	
	fmt.Printf("\nLast check: %s\n", status.LastCheck.Format("15:04:05"))
}

func (md *MiniDashboard) showRecentActivity() {
	md.terminal.Subsection("Recent Activity")
	
	activities := []Activity{
		{Time: time.Now().Add(-2 * time.Minute), Type: "scan_completed", User: "alice", Details: "OWASP scan on api.example.com"},
		{Time: time.Now().Add(-15 * time.Minute), Type: "template_created", User: "bob", Details: "New prompt injection template"},
		{Time: time.Now().Add(-1 * time.Hour), Type: "vulnerability_found", User: "system", Details: "Critical: Data exposure in LLM03"},
	}
	
	for _, activity := range activities[:3] {
		icon := md.getActivityIcon(activity.Type)
		timeStr := md.formatRelativeTime(activity.Time)
		
		fmt.Printf("%s %s %s - %s\n",
			icon,
			md.style.Info.Render(timeStr),
			md.style.Info.Render(activity.User),
			activity.Details,
		)
	}
}

func (md *MiniDashboard) showActiveScans() {
	md.terminal.Subsection("Active Scans")
	
	scans := []ActiveScan{
		{ID: "scan-001", Target: "api.prod.example.com", Progress: 72, Status: "Running", FindingsCount: 3},
		{ID: "scan-002", Target: "chat.example.com", Progress: 45, Status: "Running", FindingsCount: 1},
		{ID: "scan-003", Target: "llm.staging.example.com", Progress: 15, Status: "Initializing", FindingsCount: 0},
	}
	
	if len(scans) == 0 {
		fmt.Println("No active scans")
		return
	}
	
	for _, scan := range scans {
		// Mini progress bar
		width := 20
		filled := (scan.Progress * width) / 100
		bar := strings.Repeat("▰", filled) + strings.Repeat("▱", width-filled)
		
		fmt.Printf("%s [%s] %d%% - %s (%d findings)\n",
			scan.ID,
			bar,
			scan.Progress,
			truncate(scan.Target, 25),
			scan.FindingsCount,
		)
	}
}

func (md *MiniDashboard) showAlerts() {
	md.terminal.Subsection("Alerts")
	
	alerts := []Alert{
		{Level: "critical", Message: "High error rate detected in scan engine", Count: 15},
		{Level: "warning", Message: "API rate limit approaching threshold", Count: 1},
		{Level: "info", Message: "New templates available for download", Count: 3},
	}
	
	hasAlerts := false
	for _, alert := range alerts {
		if alert.Level == "critical" || alert.Level == "warning" {
			hasAlerts = true
			icon := "⚠️"
			style := md.style.Warning
			
			if alert.Level == "critical" {
				icon = "🚨"
				style = md.style.Critical
			}
			
			message := alert.Message
			if alert.Count > 1 {
				message += fmt.Sprintf(" (×%d)", alert.Count)
			}
			
			fmt.Printf("%s %s\n", icon, style.Render(message))
		}
	}
	
	if !hasAlerts {
		fmt.Println(md.style.Success.Render("✅ No active alerts"))
	}
}

func (md *MiniDashboard) createSparkline(values []float64) string {
	if len(values) == 0 {
		return ""
	}
	
	// Find min and max
	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	
	// Normalize and create sparkline
	sparkChars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	sparkline := ""
	
	for _, v := range values {
		normalized := (v - min) / (max - min)
		charIndex := int(normalized * float64(len(sparkChars)-1))
		sparkline += sparkChars[charIndex]
	}
	
	return sparkline
}

func (md *MiniDashboard) getCoverageCell(coverage float64) string {
	if coverage < 0 {
		return md.style.Info.Render("    ░    ")
	}
	
	style := md.style.Critical
	if coverage >= 90 {
		style = md.style.Success
	} else if coverage >= 50 {
		style = md.style.Warning
	}
	
	return style.Render(fmt.Sprintf("  %3.0f%%  ", coverage))
}

func (md *MiniDashboard) getActivityIcon(activityType string) string {
	switch activityType {
	case "scan_completed":
		return "✅"
	case "scan_started":
		return "🚀"
	case "template_created":
		return "📝"
	case "vulnerability_found":
		return "🔍"
	case "user_login":
		return "👤"
	default:
		return "•"
	}
}

func (md *MiniDashboard) formatRelativeTime(t time.Time) string {
	diff := time.Since(t)
	
	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	} else {
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Data structures for mini dashboard

type ActiveScan struct {
	ID             string
	Target         string
	Progress       int
	Status         string
	TestsCompleted int
	TestsTotal     int
	FindingsCount  int
	ETA            string
}

type CompactResults struct {
	ScanID     string
	Duration   string
	TotalTests int
	Critical   int
	High       int
	Medium     int
	Low        int
}

type LiveMetrics struct {
	ActiveScans       int
	QueuedScans       int
	RequestsPerSecond float64
	AverageLatency    int
	SystemLoad        float64
	MemoryUsage       float64
}

type TestMatrix struct {
	Categories []TestCategory
	Targets    []string
	Coverage   map[string]map[string]float64
}

type TestCategory struct {
	Name  string
	Tests int
}

func (tm *TestMatrix) GetCoverage(category, target string) float64 {
	if tm.Coverage == nil {
		return -1
	}
	if catCov, ok := tm.Coverage[category]; ok {
		if cov, ok := catCov[target]; ok {
			return cov
		}
	}
	return -1
}

type SystemStatus struct {
	APIStatus      string
	DatabaseStatus string
	QueueStatus    string
	LastCheck      time.Time
}

type Activity struct {
	Time    time.Time
	Type    string
	User    string
	Details string
}

type Alert struct {
	Level   string
	Message string
	Count   int
}