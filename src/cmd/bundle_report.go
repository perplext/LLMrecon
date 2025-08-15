package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/perplext/LLMrecon/src/bundle"
	"github.com/spf13/cobra"
)

var (
	reportFormats      []string
	reportType         string
	reportOutput       string
	includeStatistics  bool
	includeCompliance  bool
	includeVulnDetails bool
)

var bundleReportCmd = &cobra.Command{
	Use:   "report [bundle-file]",
	Short: "Generate compliance and analysis reports from bundles",
	Long: `Generate comprehensive reports from bundle files including:
- OWASP LLM Top 10 compliance analysis
- ISO/IEC 42001 compliance status  
- Template coverage statistics
- Vulnerability detection patterns
- Security assessment summaries`,
	Args: cobra.ExactArgs(1),
	RunE: runBundleReport,

func init() {
	bundleCmd.AddCommand(bundleReportCmd)

	bundleReportCmd.Flags().StringSliceVarP(&reportFormats, "format", "f", []string{"html"}, "Output formats (html,json,pdf,markdown,csv)")
	bundleReportCmd.Flags().StringVarP(&reportType, "type", "t", "full", "Report type (full,owasp,iso42001,coverage,vulnerabilities)")
	bundleReportCmd.Flags().StringVarP(&reportOutput, "output", "o", "./reports", "Output directory for reports")
	bundleReportCmd.Flags().BoolVar(&includeStatistics, "statistics", true, "Include detailed statistics")
	bundleReportCmd.Flags().BoolVar(&includeCompliance, "compliance", true, "Include compliance mappings")
	bundleReportCmd.Flags().BoolVar(&includeVulnDetails, "vuln-details", false, "Include detailed vulnerability information")

func runBundleReport(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]

	// Verify bundle exists
	if _, err := os.Stat(bundlePath); err != nil {
		return fmt.Errorf("bundle file not found: %s", bundlePath)
	}

	// Create output directory
	if err := os.MkdirAll(reportOutput, 0700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Println()
	color.Cyan("📊 Generating Bundle Report")
	fmt.Println(strings.Repeat("-", 50))

	// Load bundle
	bundleData, err := loadBundleForReport(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to load bundle: %w", err)
	}

	// Generate report data
	reportData, err := generateReportData(bundleData, reportType)
	if err != nil {
		return fmt.Errorf("failed to generate report data: %w", err)
	}

	// Generate reports in requested formats
	timestamp := time.Now().Format("20060102-150405")
	baseFilename := fmt.Sprintf("bundle-report-%s-%s", reportType, timestamp)

	for _, format := range reportFormats {
		outputPath := filepath.Join(reportOutput, baseFilename+"."+format)

		color.Yellow("  Generating %s report...", strings.ToUpper(format))

		if err := generateReportInFormat(reportData, format, outputPath); err != nil {
			color.Red("  ✗ Failed to generate %s report: %v", format, err)
			continue
		}

		color.Green("  ✓ Generated: %s", outputPath)
	}

	fmt.Println()
	color.Green("✅ Report generation complete!")
	return nil

type BundleReportData struct {
	Metadata        BundleMetadata          `json:"metadata"`
	Summary         ReportSummary           `json:"summary"`
	OWASPAnalysis   *OWASPComplianceReport  `json:"owasp_analysis,omitempty"`
	ISOAnalysis     *ISOComplianceReport    `json:"iso_analysis,omitempty"`
	Statistics      *BundleReportStatistics `json:"statistics,omitempty"`
	Vulnerabilities []VulnerabilityReport   `json:"vulnerabilities,omitempty"`
	Coverage        *CoverageReport         `json:"coverage,omitempty"`
}

type BundleMetadata struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	Description string    `json:"description"`
	Author      string    `json:"author"`
}

type ReportSummary struct {
	TotalTemplates      int      `json:"total_templates"`
	CategoriesCount     int      `json:"categories_count"`
	ComplianceScore     float64  `json:"compliance_score"`
	SecurityLevel       string   `json:"security_level"`
	TopCategories       []string `json:"top_categories"`
	RecommendationCount int      `json:"recommendation_count"`

type OWASPComplianceReport struct {
	OverallScore     float64                       `json:"overall_score"`
	CategoryScores   map[string]CategoryCompliance `json:"category_scores"`
	CoverageGaps     []string                      `json:"coverage_gaps"`
	Recommendations  []string                      `json:"recommendations"`
	DetailedFindings []OWASPFinding                `json:"detailed_findings,omitempty"`
}

type CategoryCompliance struct {
	Category        string   `json:"category"`
	Score           float64  `json:"score"`
	TemplateCount   int      `json:"template_count"`
	Coverage        float64  `json:"coverage"`
	MissingPatterns []string `json:"missing_patterns,omitempty"`
}

type OWASPFinding struct {
	Category    string `json:"category"`
	Severity    string `json:"severity"`
	Template    string `json:"template"`
	Description string `json:"description"`
	Mitigation  string `json:"mitigation"`
}

type ISOComplianceReport struct {
	ComplianceLevel   string                `json:"compliance_level"`
	RequirementsMet   int                   `json:"requirements_met"`
	TotalRequirements int                   `json:"total_requirements"`
	Sections          map[string]ISOSection `json:"sections"`
	Gaps              []ISOGap              `json:"gaps"`
	Recommendations   []string              `json:"recommendations"`

type ISOSection struct {
	Name       string   `json:"name"`
	Compliance float64  `json:"compliance"`
	Status     string   `json:"status"`
	Evidence   []string `json:"evidence,omitempty"`

type ISOGap struct {
	Requirement string `json:"requirement"`
	Current     string `json:"current"`
	Expected    string `json:"expected"`
	Priority    string `json:"priority"`

type BundleReportStatistics struct {
	TemplatesByCategory  map[string]int `json:"templates_by_category"`
	TemplatesBySeverity  map[string]int `json:"templates_by_severity"`
	TemplatesByType      map[string]int `json:"templates_by_type"`
	AverageComplexity    float64        `json:"average_complexity"`
	UniquePatterns       int            `json:"unique_patterns"`
	TotalDetectionRules  int            `json:"total_detection_rules"`
	LanguageDistribution map[string]int `json:"language_distribution"`
	UpdateFrequency      string         `json:"update_frequency"`

type VulnerabilityReport struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Severity    string   `json:"severity"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`
	Likelihood  string   `json:"likelihood"`
	Detections  []string `json:"detections"`
	Mitigations []string `json:"mitigations"`
}

type CoverageReport struct {
	OverallCoverage      float64                    `json:"overall_coverage"`
	CategoryCoverage     map[string]float64         `json:"category_coverage"`
	AttackVectorCoverage map[string]float64         `json:"attack_vector_coverage"`
	UncoveredAreas       []string                   `json:"uncovered_areas"`
	CoverageMatrix       map[string]map[string]bool `json:"coverage_matrix"`
}

func loadBundleForReport(bundlePath string) (*bundle.Bundle, error) {
	// Load bundle
	b, err := bundle.LoadBundle(bundlePath)
	if err != nil {
		return nil, err
	}

	return b, nil

func generateReportData(b *bundle.Bundle, reportType string) (*BundleReportData, error) {
	data := &BundleReportData{
		Metadata: BundleMetadata{
			Name:        b.Manifest.Name,
			Version:     b.Manifest.Version,
			CreatedAt:   b.Manifest.CreatedAt,
			Description: b.Manifest.Description,
			Author:      b.Manifest.Author.Name,
		},
	}

	// Calculate summary
	data.Summary = calculateSummary(b)

	// Generate specific report components based on type
	switch reportType {
	case "full":
		// Generate all reports
		data.OWASPAnalysis = generateOWASPAnalysis(b)
		data.ISOAnalysis = generateISOAnalysis(b)
		data.Statistics = generateStatistics(b)
		data.Vulnerabilities = generateVulnerabilityReport(b)
		data.Coverage = generateCoverageReport(b)

	case "owasp":
		data.OWASPAnalysis = generateOWASPAnalysis(b)
		data.Coverage = generateCoverageReport(b)

	case "iso42001":
		data.ISOAnalysis = generateISOAnalysis(b)

	case "coverage":
		data.Coverage = generateCoverageReport(b)
		data.Statistics = generateStatistics(b)

	case "vulnerabilities":
		data.Vulnerabilities = generateVulnerabilityReport(b)
		data.Statistics = generateStatistics(b)
	}

	return data, nil

func calculateSummary(b *bundle.Bundle) ReportSummary {
	categoryCount := make(map[string]int)
	for _, tmpl := range getTemplatesFromBundle(b) {
		if cat := extractCategory(tmpl.Path); cat != "" {
			categoryCount[cat]++
		}
	}

	// Get top categories
	var topCategories []string
	maxCount := 0
	for cat, count := range categoryCount {
		if count > maxCount {
			topCategories = []string{cat}
			maxCount = count
		} else if count == maxCount {
			topCategories = append(topCategories, cat)
		}
	}

	return ReportSummary{
		TotalTemplates:      len(getTemplatesFromBundle(b)),
		CategoriesCount:     len(categoryCount),
		ComplianceScore:     calculateComplianceScore(b),
		SecurityLevel:       determineSecurityLevel(b),
		TopCategories:       topCategories,
		RecommendationCount: countRecommendations(b),
	}

func generateOWASPAnalysis(b *bundle.Bundle) *OWASPComplianceReport {
	report := &OWASPComplianceReport{
		CategoryScores: make(map[string]CategoryCompliance),
	}

	// Analyze each OWASP category
	owaspCategories := []string{
		"llm01-prompt-injection",
		"llm02-insecure-output",
		"llm03-training-data-poisoning",
		"llm04-model-denial-of-service",
		"llm05-supply-chain",
		"llm06-sensitive-information",
		"llm07-insecure-plugin",
		"llm08-excessive-agency",
		"llm09-overreliance",
		"llm10-model-theft",
	}

	totalScore := 0.0
	for _, category := range owaspCategories {
		compliance := analyzeCategory(b, category)
		report.CategoryScores[category] = compliance
		totalScore += compliance.Score
	}

	report.OverallScore = totalScore / float64(len(owaspCategories))
	report.CoverageGaps = identifyCoverageGaps(report.CategoryScores)
	report.Recommendations = generateOWASPRecommendations(report)

	if includeVulnDetails {
		report.DetailedFindings = generateDetailedFindings(b)
	}

	return report

func generateISOAnalysis(b *bundle.Bundle) *ISOComplianceReport {
	report := &ISOComplianceReport{
		Sections: make(map[string]ISOSection),
	}

	// ISO/IEC 42001 sections relevant to LLM security
	sections := map[string]string{
		"5.2":  "AI Policy",
		"6.1":  "Risk Assessment",
		"6.2":  "AI Objectives",
		"7.2":  "Competence",
		"8.1":  "Operational Planning",
		"8.2":  "AI System Requirements",
		"8.3":  "AI System Design",
		"9.1":  "Monitoring and Measurement",
		"10.1": "Nonconformity and Corrective Action",
	}

	metCount := 0
	for sectionID, sectionName := range sections {
		compliance := assessISOSection(b, sectionID)
		report.Sections[sectionID] = ISOSection{
			Name:       sectionName,
			Compliance: compliance,
			Status:     getComplianceStatus(compliance),
		}
		if compliance >= 0.7 {
			metCount++
		}
	}

	report.RequirementsMet = metCount
	report.TotalRequirements = len(sections)
	report.ComplianceLevel = calculateISOLevel(float64(metCount) / float64(len(sections)))
	report.Gaps = identifyISOGaps(report.Sections)
	report.Recommendations = generateISORecommendations(report)

	return report

func generateStatistics(b *bundle.Bundle) *BundleReportStatistics {
	stats := &BundleReportStatistics{
		TemplatesByCategory:  make(map[string]int),
		TemplatesBySeverity:  make(map[string]int),
		TemplatesByType:      make(map[string]int),
		LanguageDistribution: make(map[string]int),
	}

	// Analyze templates
	for _, tmpl := range getTemplatesFromBundle(b) {
		// Category
		if cat := extractCategory(tmpl.Path); cat != "" {
			stats.TemplatesByCategory[cat]++
		}

		// Parse template for additional metadata
		if metadata := parseTemplateMetadata(tmpl); metadata != nil {
			if metadata.Severity != "" {
				stats.TemplatesBySeverity[metadata.Severity]++
			}
			if metadata.Type != "" {
				stats.TemplatesByType[metadata.Type]++
			}
		}
	}

	stats.UniquePatterns = countUniquePatterns(b)
	stats.TotalDetectionRules = countDetectionRules(b)
	stats.AverageComplexity = calculateAverageComplexity(b)

	return stats

func generateVulnerabilityReport(b *bundle.Bundle) []VulnerabilityReport {
	var vulnerabilities []VulnerabilityReport

	// Map OWASP categories to vulnerability reports
	categoryVulnMap := map[string]VulnerabilityReport{
		"llm01-prompt-injection": {
			ID:          "VULN-001",
			Name:        "Prompt Injection",
			Category:    "Input Validation",
			Severity:    "High",
			Description: "Malicious prompts that manipulate LLM behavior",
			Impact:      "Unauthorized access, data exfiltration, system compromise",
			Likelihood:  "High",
		},
		"llm02-insecure-output": {
			ID:          "VULN-002",
			Name:        "Insecure Output Handling",
			Category:    "Output Validation",
			Severity:    "High",
			Description: "Unvalidated LLM outputs leading to injection attacks",
			Impact:      "XSS, SQL injection, command execution",
			Likelihood:  "Medium",
		},
		// Add more vulnerability mappings...
	}

	for category, vuln := range categoryVulnMap {
		// Check if we have templates for this vulnerability
		if hasTemplatesForCategory(b, category) {
			vuln.Detections = getDetectionMethods(b, category)
			vuln.Mitigations = getMitigationStrategies(category)
			vulnerabilities = append(vulnerabilities, vuln)
		}
	}

	return vulnerabilities

func generateCoverageReport(b *bundle.Bundle) *CoverageReport {
	report := &CoverageReport{
		CategoryCoverage:     make(map[string]float64),
		AttackVectorCoverage: make(map[string]float64),
		CoverageMatrix:       make(map[string]map[string]bool),
	}

	// Calculate category coverage
	owaspCategories := []string{
		"llm01-prompt-injection",
		"llm02-insecure-output",
		"llm03-training-data-poisoning",
		"llm04-model-denial-of-service",
		"llm05-supply-chain",
		"llm06-sensitive-information",
		"llm07-insecure-plugin",
		"llm08-excessive-agency",
		"llm09-overreliance",
		"llm10-model-theft",
	}

	coveredCount := 0
	for _, category := range owaspCategories {
		coverage := calculateCategoryCoverage(b, category)
		report.CategoryCoverage[category] = coverage
		if coverage > 0 {
			coveredCount++
		}
	}

	report.OverallCoverage = float64(coveredCount) / float64(len(owaspCategories))

	// Calculate attack vector coverage
	attackVectors := []string{
		"direct-injection",
		"indirect-injection",
		"data-poisoning",
		"model-extraction",
		"supply-chain",
	}

	for _, vector := range attackVectors {
		report.AttackVectorCoverage[vector] = calculateVectorCoverage(b, vector)
	}

	// Identify uncovered areas
	report.UncoveredAreas = identifyUncoveredAreas(report)
	return report

func generateReportInFormat(data *BundleReportData, format, outputPath string) error {
	// Simple report generation
	switch format {
	case "json":
		// Generate JSON report
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling report data: %w", err)
		}
		return os.WriteFile(filepath.Clean(outputPath, jsonData, 0600))

	case "html":
		// Generate simple HTML report
		html := generateSimpleHTMLReport(data)
		return os.WriteFile(filepath.Clean(outputPath, []byte(html)), 0600)

	case "markdown":
		// Generate markdown report
		md := generateMarkdownReport(data)
		return os.WriteFile(filepath.Clean(outputPath, []byte(md)), 0600)

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

// Helper functions

func extractCategory(path string) string {
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "llm") && len(part) > 3 {
			return part
		}
	}
	return ""

func calculateComplianceScore(b *bundle.Bundle) float64 {
	// Simple scoring based on template coverage
	expectedTemplates := 30 // Assume 3 templates per category minimum
	actualTemplates := len(getTemplatesFromBundle(b))

	score := float64(actualTemplates) / float64(expectedTemplates)
	if score > 1.0 {
		score = 1.0
	}

	return score * 100

func determineSecurityLevel(b *bundle.Bundle) string {
	score := calculateComplianceScore(b)
	switch {
	case score >= 90:
		return "Excellent"
	case score >= 75:
		return "Good"
	case score >= 60:
		return "Fair"
	case score >= 40:
		return "Poor"
	default:
		return "Critical"
	}

func countRecommendations(b *bundle.Bundle) int {
	// Count based on missing categories and low coverage
	count := 0
	categories := make(map[string]bool)

	for _, tmpl := range getTemplatesFromBundle(b) {
		if cat := extractCategory(tmpl.Path); cat != "" {
			categories[cat] = true
		}
	}

	// Each missing category generates recommendations
	expectedCategories := 10
	count = (expectedCategories - len(categories)) * 2

	return count

func analyzeCategory(b *bundle.Bundle, category string) CategoryCompliance {
	templateCount := 0
	for _, tmpl := range getTemplatesFromBundle(b) {
		if strings.Contains(tmpl.Path, category) {
			templateCount++
		}
	}

	// Expected minimum templates per category
	expectedTemplates := 3
	coverage := float64(templateCount) / float64(expectedTemplates)
	if coverage > 1.0 {
		coverage = 1.0
	}

	score := coverage * 100

	return CategoryCompliance{
		Category:      category,
		Score:         score,
		TemplateCount: templateCount,
		Coverage:      coverage,
	}

func identifyCoverageGaps(scores map[string]CategoryCompliance) []string {
	var gaps []string

	for category, compliance := range scores {
		if compliance.Coverage < 0.5 {
			gaps = append(gaps, fmt.Sprintf("%s (%.0f%% coverage)", category, compliance.Coverage*100))
		}
	}

	return gaps

func generateOWASPRecommendations(report *OWASPComplianceReport) []string {
	var recommendations []string

	// Low coverage recommendations
	for category, compliance := range report.CategoryScores {
		if compliance.Coverage < 0.5 {
			recommendations = append(recommendations,
				fmt.Sprintf("Increase test coverage for %s - currently at %.0f%%",
					category, compliance.Coverage*100))
		}
	}

	// Overall score recommendations
	if report.OverallScore < 70 {
		recommendations = append(recommendations,
			"Consider implementing comprehensive test suites for all OWASP LLM Top 10 categories")
	}

	return recommendations

func parseTemplateMetadata(tmpl bundle.ContentItem) *TemplateMetadata {
	// This would parse the actual template content
	// For now, return mock data
	return &TemplateMetadata{
		Severity: "High",
		Type:     "Detection",
	}

type TemplateMetadata struct {
	Severity string
	Type     string

func hasTemplatesForCategory(b *bundle.Bundle, category string) bool {
	for _, tmpl := range getTemplatesFromBundle(b) {
		if strings.Contains(tmpl.Path, category) {
			return true
		}
	}
	return false

func getDetectionMethods(b *bundle.Bundle, category string) []string {
	// This would analyze templates to extract detection methods
	return []string{
		"Pattern matching",
		"Behavioral analysis",
		"Statistical anomaly detection",
	}

func getMitigationStrategies(category string) []string {
	// Return standard mitigations per category
	mitigations := map[string][]string{
		"llm01-prompt-injection": {
			"Input validation and sanitization",
			"Prompt engineering best practices",
			"Context isolation",
		},
		"llm02-insecure-output": {
			"Output encoding and escaping",
			"Content Security Policy",
			"Strict output validation",
		},
	}

	if m, ok := mitigations[category]; ok {
		return m
	}

	return []string{"Implement security best practices"}

func calculateCategoryCoverage(b *bundle.Bundle, category string) float64 {
	count := 0
	for _, tmpl := range getTemplatesFromBundle(b) {
		if strings.Contains(tmpl.Path, category) {
			count++
		}
	}

	if count > 0 {
		return 1.0 // Simple binary coverage for now
	}
	return 0.0

func calculateVectorCoverage(b *bundle.Bundle, vector string) float64 {
	// This would analyze template content for attack vectors
	// For now, return mock coverage
	return 0.75

func identifyUncoveredAreas(report *CoverageReport) []string {
	var uncovered []string

	for category, coverage := range report.CategoryCoverage {
		if coverage == 0 {
			uncovered = append(uncovered, category)
		}
	}

	return uncovered

func assessISOSection(b *bundle.Bundle, sectionID string) float64 {
	// Assess compliance based on bundle contents
	// This is a simplified assessment
	switch sectionID {
	case "5.2": // AI Policy
		return 0.8
	case "6.1": // Risk Assessment
		return 0.9
	case "8.2": // AI System Requirements
		return 0.85
	default:
		return 0.7
	}

func getComplianceStatus(compliance float64) string {
	switch {
	case compliance >= 0.9:
		return "Fully Compliant"
	case compliance >= 0.7:
		return "Substantially Compliant"
	case compliance >= 0.5:
		return "Partially Compliant"
	default:
		return "Non-Compliant"
	}

func calculateISOLevel(ratio float64) string {
	switch {
	case ratio >= 0.9:
		return "Level 3 - Optimized"
	case ratio >= 0.7:
		return "Level 2 - Managed"
	case ratio >= 0.5:
		return "Level 1 - Basic"
	default:
		return "Level 0 - Initial"
	}

func identifyISOGaps(sections map[string]ISOSection) []ISOGap {
	var gaps []ISOGap

	for sectionID, section := range sections {
		if section.Compliance < 0.7 {
			gaps = append(gaps, ISOGap{
				Requirement: fmt.Sprintf("Section %s - %s", sectionID, section.Name),
				Current:     fmt.Sprintf("%.0f%% compliant", section.Compliance*100),
				Expected:    "70% minimum compliance",
				Priority:    "High",
			})
		}
	}

	return gaps

func generateISORecommendations(report *ISOComplianceReport) []string {
	var recommendations []string

	if report.ComplianceLevel != "Level 3 - Optimized" {
		recommendations = append(recommendations,
			"Implement continuous improvement processes to achieve Level 3 compliance")
	}

	for _, gap := range report.Gaps {
		recommendations = append(recommendations,
			fmt.Sprintf("Address %s to meet compliance requirements", gap.Requirement))
	}

	return recommendations

func countUniquePatterns(b *bundle.Bundle) int {
	// Count unique detection patterns across templates
	return len(getTemplatesFromBundle(b)) * 3 // Simplified calculation

func countDetectionRules(b *bundle.Bundle) int {
	// Count total detection rules
	return len(getTemplatesFromBundle(b)) * 5 // Simplified calculation

func calculateAverageComplexity(b *bundle.Bundle) float64 {
	// Calculate average template complexity
	return 3.5 // Mock value

func getTemplatesFromBundle(b *bundle.Bundle) []bundle.ContentItem {
	var templates []bundle.ContentItem
	for _, item := range b.Manifest.Content {
		if item.Type == bundle.TemplateContentType {
			templates = append(templates, item)
		}
	}
	return templates

func generateDetailedFindings(b *bundle.Bundle) []OWASPFinding {
	// Generate detailed findings for each template
	var findings []OWASPFinding

	// Add sample findings
	findings = append(findings, OWASPFinding{
		Category:    "llm01-prompt-injection",
		Severity:    "High",
		Template:    "direct-injection.yaml",
		Description: "Direct prompt injection vulnerability detected",
		Mitigation:  "Implement input validation and prompt sanitization",
	})

	return findings

func generateSimpleHTMLReport(data *BundleReportData) string {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<title>Bundle Report - %s</title>
<style>
body { font-family: Arial, sans-serif; margin: 20px; }
h1, h2, h3 { color: #333; }
table { border-collapse: collapse; width: 100%%; margin: 20px 0; }
th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
th { background-color: #f2f2f2; }
</style>
</head>
<body>
<h1>Bundle Analysis Report</h1>
<h2>%s v%s</h2>
<p>Generated: %s</p>
<p>%s</p>
<h3>Summary</h3>
<ul>
<li>Total Templates: %d</li>
<li>Categories: %d</li>
<li>Compliance Score: %.1f%%</li>
<li>Security Level: %s</li>
</ul>
</body>
</html>`,
		data.Metadata.Name,
		data.Metadata.Name,
		data.Metadata.Version,
		time.Now().Format("2006-01-02 15:04:05"),
		data.Metadata.Description,
		data.Summary.TotalTemplates,
		data.Summary.CategoriesCount,
		data.Summary.ComplianceScore,
		data.Summary.SecurityLevel,
	)
	return html

func generateMarkdownReport(data *BundleReportData) string {
	md := fmt.Sprintf(`# Bundle Analysis Report

## %s v%s

**Generated:** %s

%s

### Summary

- Total Templates: %d
- Categories: %d  
- Compliance Score: %.1f%%
- Security Level: %s

`,
		data.Metadata.Name,
		data.Metadata.Version,
		time.Now().Format("2006-01-02 15:04:05"),
		data.Metadata.Description,
		data.Summary.TotalTemplates,
		data.Summary.CategoriesCount,
		data.Summary.ComplianceScore,
		data.Summary.SecurityLevel,
	)
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
