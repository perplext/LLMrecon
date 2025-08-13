package ui

import (
	"fmt"
	"io"
	"strings"
	"time"
)

// StyledOutput provides styled terminal output
type StyledOutput struct {
	writer    io.Writer
	formatter *Formatter
	icons     *Icons
	boxChars  *BoxChars
	width     int
}

// NewStyledOutput creates a new styled output handler
func NewStyledOutput(writer io.Writer, colorEnabled bool, width int) *StyledOutput {
	return &StyledOutput{
		writer:    writer,
		formatter: NewFormatter(DefaultColorScheme(), colorEnabled),
		icons:     DefaultIcons(),
		boxChars:  DefaultBoxChars(),
		width:     width,
	}
}

// SetColorScheme sets the color scheme
func (so *StyledOutput) SetColorScheme(scheme *ColorScheme) {
	so.formatter.scheme = scheme
}

// SetASCIIMode enables ASCII-only mode
func (so *StyledOutput) SetASCIIMode() {
	so.icons = ASCIIIcons()
	so.boxChars = ASCIIBoxChars()
}

// Banner prints a large banner
func (so *StyledOutput) Banner(text string) {
	width := so.width
	if width > 80 {
		width = 80
	}

	border := strings.Repeat("═", width)
	padding := (width - len(text) - 2) / 2
	
	fmt.Fprintln(so.writer, so.formatter.Header(border))
	fmt.Fprintf(so.writer, "%s%s%s%s%s\n",
		so.formatter.Header("║"),
		strings.Repeat(" ", padding),
		so.formatter.Header(text),
		strings.Repeat(" ", width-padding-len(text)-2),
		so.formatter.Header("║"),
	)
	fmt.Fprintln(so.writer, so.formatter.Header(border))
}

// Section prints a section header
func (so *StyledOutput) Section(title string) {
	fmt.Fprintf(so.writer, "\n%s %s\n%s\n",
		so.formatter.Subheader("▶"),
		so.formatter.Subheader(title),
		so.formatter.Muted(strings.Repeat("─", len(title)+2)),
	)
}

// StatusLine prints a status line with icon
func (so *StyledOutput) StatusLine(status, message string, args ...interface{}) {
	var icon string
	var format func(string, ...interface{}) string

	switch strings.ToLower(status) {
	case "success", "ok", "done":
		icon = so.icons.Success
		format = so.formatter.Success
	case "error", "fail", "failed":
		icon = so.icons.Error
		format = so.formatter.Error
	case "warning", "warn":
		icon = so.icons.Warning
		format = so.formatter.Warning
	case "info":
		icon = so.icons.Info
		format = so.formatter.Info
	case "debug":
		icon = so.icons.Debug
		format = so.formatter.Debug
	default:
		icon = so.icons.Bullet
		format = func(s string, a ...interface{}) string {
			return fmt.Sprintf(s, a...)
		}
	}

	fmt.Fprintf(so.writer, "%s %s\n", format(icon, args...), fmt.Sprintf(message, args...))
}

// KeyValue prints a key-value pair
func (so *StyledOutput) KeyValue(key string, value interface{}) {
	fmt.Fprintf(so.writer, "%s: %s\n", 
		so.formatter.Label(key),
		so.formatter.Value("%v", value),
	)
}

// KeyValueList prints a list of key-value pairs
func (so *StyledOutput) KeyValueList(pairs map[string]interface{}) {
	maxKeyLen := 0
	for key := range pairs {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}

	for key, value := range pairs {
		fmt.Fprintf(so.writer, "%s: %s\n",
			so.formatter.Label(fmt.Sprintf("%-*s", maxKeyLen, key)),
			so.formatter.Value("%v", value),
		)
	}
}

// VulnerabilityFinding prints a formatted vulnerability finding
func (so *StyledOutput) VulnerabilityFinding(finding VulnerabilityFinding) {
	// Severity badge
	severityBadge := fmt.Sprintf(" %s ", strings.ToUpper(finding.Severity))
	fmt.Fprintf(so.writer, "%s %s\n",
		so.formatter.Severity(severityBadge),
		so.formatter.Label(finding.Title),
	)

	// Details
	if finding.Description != "" {
		fmt.Fprintf(so.writer, "  %s %s\n", so.icons.Arrow, finding.Description)
	}

	// Template info
	if finding.TemplateID != "" {
		fmt.Fprintf(so.writer, "  %s %s: %s\n",
			so.icons.Bullet,
			so.formatter.Muted("Template"),
			so.formatter.Code(finding.TemplateID),
		)
	}

	// Evidence
	if finding.Evidence != "" {
		fmt.Fprintf(so.writer, "  %s %s\n",
			so.icons.Bullet,
			so.formatter.Muted("Evidence:"),
		)
		so.CodeBlock(finding.Evidence, "  ")
	}

	// Remediation
	if finding.Remediation != "" {
		fmt.Fprintf(so.writer, "  %s %s %s\n",
			so.icons.Shield,
			so.formatter.Muted("Remediation:"),
			finding.Remediation,
		)
	}

	fmt.Fprintln(so.writer)
}

// VulnerabilityFinding represents a security finding
type VulnerabilityFinding struct {
	Severity    string
	Title       string
	Description string
	TemplateID  string
	Evidence    string
	Remediation string
}

// ScanSummary prints a scan summary
func (so *StyledOutput) ScanSummary(summary ScanSummary) {
	// Header box
	content := fmt.Sprintf(
		"Total Tests: %d\nPassed: %d\nFailed: %d\nDuration: %s",
		summary.TotalTests,
		summary.Passed,
		summary.Failed,
		summary.Duration,
	)
	
	box := RenderBox("Scan Summary", content, 40, so.boxChars, so.formatter)
	fmt.Fprintln(so.writer, box)

	// Severity distribution
	if summary.Critical > 0 || summary.High > 0 || summary.Medium > 0 || summary.Low > 0 {
		fmt.Fprintln(so.writer, "\nSeverity Distribution:")
		
		// Bar chart
		bar := RenderSeverityBar(summary.Critical, summary.High, summary.Medium, summary.Low, 40, so.formatter)
		fmt.Fprintln(so.writer, bar)
		
		// Legend
		fmt.Fprintf(so.writer, "%s Critical: %d  %s High: %d  %s Medium: %d  %s Low: %d\n",
			so.formatter.format(so.formatter.scheme.Critical, "█"),
			summary.Critical,
			so.formatter.format(so.formatter.scheme.High, "█"),
			summary.High,
			so.formatter.format(so.formatter.scheme.Medium, "█"),
			summary.Medium,
			so.formatter.format(so.formatter.scheme.Low, "█"),
			summary.Low,
		)
	}

	// Success rate
	if summary.TotalTests > 0 {
		successRate := float64(summary.Passed) / float64(summary.TotalTests) * 100
		fmt.Fprintf(so.writer, "\nSuccess Rate: %s\n",
			so.formatter.Value("%.1f%%", successRate),
		)
		
		// Progress bar
		bar := RenderProgressBar(summary.Passed, summary.TotalTests, 40, so.formatter)
		fmt.Fprintln(so.writer, bar)
	}
}

// ScanSummary represents scan results summary
type ScanSummary struct {
	TotalTests int
	Passed     int
	Failed     int
	Critical   int
	High       int
	Medium     int
	Low        int
	Duration   time.Duration
}

// TemplateInfo prints template information
func (so *StyledOutput) TemplateInfo(template TemplateInfo) {
	// Header
	fmt.Fprintf(so.writer, "%s %s\n",
		so.formatter.Header(so.icons.Database),
		so.formatter.Header(template.Name),
	)

	// Metadata
	so.KeyValue("ID", so.formatter.Code(template.ID))
	so.KeyValue("Category", template.Category)
	so.KeyValue("Severity", so.formatter.Severity(template.Severity))
	so.KeyValue("Author", template.Author)
	
	// Description
	if template.Description != "" {
		fmt.Fprintf(so.writer, "\n%s\n", template.Description)
	}

	// Tags
	if len(template.Tags) > 0 {
		fmt.Fprintf(so.writer, "\n%s: ", so.formatter.Label("Tags"))
		for i, tag := range template.Tags {
			if i > 0 {
				fmt.Fprint(so.writer, ", ")
			}
			fmt.Fprint(so.writer, so.formatter.Code(tag))
		}
		fmt.Fprintln(so.writer)
	}
}

// TemplateInfo represents template information
type TemplateInfo struct {
	ID          string
	Name        string
	Description string
	Category    string
	Severity    string
	Author      string
	Tags        []string
}

// CodeBlock prints a formatted code block
func (so *StyledOutput) CodeBlock(code string, indent string) {
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		fmt.Fprintf(so.writer, "%s%s\n", indent, so.formatter.Code(line))
	}
}

// Quote prints a formatted quote
func (so *StyledOutput) Quote(text string, author string) {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		fmt.Fprintf(so.writer, "  %s %s\n", 
			so.formatter.Muted("│"),
			so.formatter.Quote(line),
		)
	}
	if author != "" {
		fmt.Fprintf(so.writer, "  %s %s\n",
			so.formatter.Muted("└─"),
			so.formatter.Muted(author),
		)
	}
}

// Tree prints a tree structure
func (so *StyledOutput) Tree(root TreeNode, indent string) {
	so.printTreeNode(root, indent, true, true)
}

// TreeNode represents a node in a tree structure
type TreeNode struct {
	Name     string
	Type     string
	Children []TreeNode
}

// printTreeNode recursively prints tree nodes
func (so *StyledOutput) printTreeNode(node TreeNode, indent string, isLast bool, isRoot bool) {
	// Node prefix
	prefix := "├─"
	if isLast {
		prefix = "└─"
	}
	if isRoot {
		prefix = ""
	}

	// Node icon based on type
	icon := so.icons.Bullet
	switch node.Type {
	case "folder", "directory":
		icon = "📁"
	case "file":
		icon = "📄"
	case "template":
		icon = "📋"
	case "config":
		icon = "⚙"
	}

	// Print node
	if !isRoot {
		fmt.Fprintf(so.writer, "%s%s %s %s\n", indent, prefix, icon, node.Name)
	} else {
		fmt.Fprintf(so.writer, "%s %s\n", icon, so.formatter.Header(node.Name))
	}

	// Print children
	for i, child := range node.Children {
		childIndent := indent
		if !isRoot {
			if isLast {
				childIndent += "   "
			} else {
				childIndent += "│  "
			}
		}
		so.printTreeNode(child, childIndent, i == len(node.Children)-1, false)
	}
}

// ComparisonTable prints a comparison table
func (so *StyledOutput) ComparisonTable(title string, headers []string, rows [][]string) {
	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}
	
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Print title
	if title != "" {
		fmt.Fprintln(so.writer, so.formatter.Subheader(title))
	}

	// Print header
	var headerLine strings.Builder
	var separatorLine strings.Builder
	
	for i, header := range headers {
		if i > 0 {
			headerLine.WriteString(" │ ")
			separatorLine.WriteString("─┼─")
		}
		headerLine.WriteString(so.formatter.Label(fmt.Sprintf("%-*s", colWidths[i], header)))
		separatorLine.WriteString(strings.Repeat("─", colWidths[i]))
	}
	
	fmt.Fprintln(so.writer, headerLine.String())
	fmt.Fprintln(so.writer, so.formatter.Muted(separatorLine.String()))

	// Print rows with alternating colors
	for i, row := range rows {
		var rowLine strings.Builder
		
		for j, cell := range row {
			if j > 0 {
				rowLine.WriteString(" │ ")
			}
			
			// Apply special formatting
			formatted := cell
			if j == 0 {
				// First column often contains names/IDs
				formatted = so.formatter.Value(cell)
			} else if strings.Contains(strings.ToLower(cell), "pass") || strings.Contains(strings.ToLower(cell), "success") {
				formatted = so.formatter.Success(cell)
			} else if strings.Contains(strings.ToLower(cell), "fail") || strings.Contains(strings.ToLower(cell), "error") {
				formatted = so.formatter.Error(cell)
			} else if strings.Contains(strings.ToLower(cell), "warn") {
				formatted = so.formatter.Warning(cell)
			}
			
			if j < len(colWidths) {
				rowLine.WriteString(fmt.Sprintf("%-*s", colWidths[j], formatted))
			} else {
				rowLine.WriteString(formatted)
			}
		}
		
		fmt.Fprintln(so.writer, rowLine.String())
	}
}

// Alert prints an alert box
func (so *StyledOutput) Alert(alertType, title, message string) {
	var icon string
	var colorFunc func(string, ...interface{}) string

	switch strings.ToLower(alertType) {
	case "success":
		icon = so.icons.Success
		colorFunc = so.formatter.Success
	case "error":
		icon = so.icons.Error
		colorFunc = so.formatter.Error
	case "warning":
		icon = so.icons.Warning
		colorFunc = so.formatter.Warning
	case "info":
		icon = so.icons.Info
		colorFunc = so.formatter.Info
	default:
		icon = so.icons.Info
		colorFunc = so.formatter.Info
	}

	// Box content
	content := fmt.Sprintf("%s %s\n\n%s", icon, title, message)
	
	// Create colored box
	width := 60
	if so.width < 60 {
		width = so.width - 4
	}
	
	box := RenderBox("", content, width, so.boxChars, so.formatter)
	
	// Apply color to entire box
	lines := strings.Split(box, "\n")
	for _, line := range lines {
		fmt.Fprintln(so.writer, colorFunc(line))
	}
}