package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/perplext/LLMrecon/src/bundle"
	"github.com/spf13/cobra"
)

var (
	categorizeOutput   string
	categorizeFormat   string
	categorizeCopy     bool
	categorizeSymlink  bool
	categorizeMetadata bool
)

var bundleCategorizeCmd = &cobra.Command{
	Use:   "categorize [source]",
	Short: "Organize templates by OWASP LLM Top 10 categories",
	Long: `Analyze and organize security test templates according to OWASP LLM Top 10 categories.
	
This command helps you:
- Automatically categorize templates based on content analysis
- Reorganize template directories by OWASP categories
- Generate category mappings and reports
- Ensure comprehensive coverage of all vulnerability types`,
	Example: `  # Categorize templates in current directory
  LLMrecon bundle categorize ./templates
  
  # Categorize and copy to new structure
  LLMrecon bundle categorize ./templates --output=./categorized --copy
  
  # Generate category report
  LLMrecon bundle categorize ./templates --format=json --output=categories.json`,
	Args: cobra.ExactArgs(1),
	RunE: runBundleCategorize,

func init() {
	bundleCmd.AddCommand(bundleCategorizeCmd)

	bundleCategorizeCmd.Flags().StringVarP(&categorizeOutput, "output", "o", "", "Output directory or file")
	bundleCategorizeCmd.Flags().StringVarP(&categorizeFormat, "format", "f", "directory", "Output format (directory,json,report)")
	bundleCategorizeCmd.Flags().BoolVar(&categorizeCopy, "copy", false, "Copy files to new structure")
	bundleCategorizeCmd.Flags().BoolVar(&categorizeSymlink, "symlink", false, "Create symlinks instead of copying")
	bundleCategorizeCmd.Flags().BoolVar(&categorizeMetadata, "metadata", true, "Include category metadata")

// OWASPCategory represents a category with full details
type OWASPCategory struct {
	ID          string   `json:"id"`
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
	Patterns    []string `json:"patterns"`
	Examples    []string `json:"examples"`
}

// CategoryMapping represents template to category mapping
type CategoryMapping struct {
	Template    string   `json:"template"`
	Path        string   `json:"path"`
	Category    string   `json:"category"`
	Confidence  float64  `json:"confidence"`
	MatchedKeys []string `json:"matched_keywords"`
}

// CategoryReport represents categorization results
type CategoryReport struct {
	TotalTemplates     int                       `json:"total_templates"`
	CategorizedCount   int                       `json:"categorized_count"`
	UncategorizedCount int                       `json:"uncategorized_count"`
	Categories         map[string]*CategoryStats `json:"categories"`
	Mappings           []CategoryMapping         `json:"mappings"`
	Coverage           map[string]float64        `json:"coverage"`

// CategoryStats represents statistics for a category
type CategoryStats struct {
	Count      int            `json:"count"`
	Percentage float64        `json:"percentage"`
	Templates  []string       `json:"templates"`
	SubTypes   map[string]int `json:"subtypes,omitempty"`

func runBundleCategorize(cmd *cobra.Command, args []string) error {
	sourcePath := args[0]

	fmt.Println()
	color.Cyan("🗂️  OWASP LLM Top 10 Template Categorization")
	fmt.Println(strings.Repeat("-", 50))

	// Verify source exists
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("source not found: %s", sourcePath)
	}

	// Collect templates
	var templates []string
	if sourceInfo.IsDir() {
		color.Yellow("Scanning directory: %s", sourcePath)
		templates, err = collectTemplates(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to collect templates: %w", err)
		}
	} else if strings.HasSuffix(sourcePath, ".bundle") {
		color.Yellow("Loading bundle: %s", sourcePath)
		templates, err = collectTemplatesFromBundle(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to load bundle: %w", err)
		}
	} else {
		templates = []string{sourcePath}
	}

	fmt.Printf("Found %d templates to categorize\n\n", len(templates))

	// Perform categorization
	color.Cyan("🔍 Analyzing templates...")
	mappings := categorizeTemplates(templates)

	// Generate report
	report := generateCategoryReport(mappings)

	// Display results
	displayCategoryReport(report)

	// Process output based on format
	switch categorizeFormat {
	case "directory":
		if categorizeOutput == "" {
			categorizeOutput = "./owasp-categorized"
		}
		if err := organizeByCategoryDirectory(mappings, sourcePath, categorizeOutput); err != nil {
			return fmt.Errorf("failed to organize templates: %w", err)
		}

	case "json":
		if categorizeOutput == "" {
			categorizeOutput = "category-report.json"
		}
		if err := saveCategoryReport(report, categorizeOutput); err != nil {
			return fmt.Errorf("failed to save report: %w", err)
		}

	case "report":
		if categorizeOutput == "" {
			categorizeOutput = "category-report.md"
		}
		if err := saveCategoryReportMarkdown(report, categorizeOutput); err != nil {
			return fmt.Errorf("failed to save report: %w", err)
		}
	}

	// Show recommendations
	showCategoryRecommendations(report)

	return nil

// Get detailed OWASP categories
func getDetailedOWASPCategories() []OWASPCategory {
	return []OWASPCategory{
		{
			ID:          "LLM01",
			Code:        "llm01-prompt-injection",
			Name:        "Prompt Injection",
			Description: "Crafted inputs manipulating LLM behavior through prompt engineering",
			Keywords:    []string{"prompt", "injection", "jailbreak", "bypass", "override", "instruction"},
			Patterns:    []string{"ignore.*previous", "disregard.*instruction", "new.*directive"},
			Examples:    []string{"direct-injection", "indirect-injection", "jailbreaking"},
		},
		{
			ID:          "LLM02",
			Code:        "llm02-insecure-output",
			Name:        "Insecure Output Handling",
			Description: "Insufficient validation, sanitization, and handling of LLM outputs",
			Keywords:    []string{"output", "xss", "injection", "sql", "command", "ssrf", "sanitize"},
			Patterns:    []string{"<script", "'; DROP", "exec\\(", "../../"},
			Examples:    []string{"xss-injection", "sql-injection", "command-injection"},
		},
		{
			ID:          "LLM03",
			Code:        "llm03-training-data-poisoning",
			Name:        "Training Data Poisoning",
			Description: "Corruption of training data to introduce vulnerabilities or biases",
			Keywords:    []string{"training", "data", "poison", "backdoor", "bias", "corruption"},
			Patterns:    []string{"training.*data", "poison.*attack", "backdoor.*trigger"},
			Examples:    []string{"data-poisoning", "backdoor-attacks", "bias-injection"},
		},
		{
			ID:          "LLM04",
			Code:        "llm04-model-denial-of-service",
			Name:        "Model Denial of Service",
			Description: "Resource exhaustion through expensive operations or excessive requests",
			Keywords:    []string{"dos", "denial", "resource", "exhaustion", "flood", "overload"},
			Patterns:    []string{"resource.*exhaust", "token.*flood", "context.*overflow"},
			Examples:    []string{"resource-exhaustion", "token-flooding", "context-saturation"},
		},
		{
			ID:          "LLM05",
			Code:        "llm05-supply-chain",
			Name:        "Supply Chain Vulnerabilities",
			Description: "Compromised components, models, or data in the ML pipeline",
			Keywords:    []string{"supply", "chain", "dependency", "component", "third-party", "library"},
			Patterns:    []string{"vulnerable.*dependency", "compromised.*model", "malicious.*package"},
			Examples:    []string{"dependency-risks", "model-vulnerabilities", "integration-flaws"},
		},
		{
			ID:          "LLM06",
			Code:        "llm06-sensitive-information",
			Name:        "Sensitive Information Disclosure",
			Description: "Unauthorized access to confidential data through the model",
			Keywords:    []string{"sensitive", "pii", "confidential", "leak", "disclosure", "privacy"},
			Patterns:    []string{"personal.*information", "credential.*leak", "data.*exposure"},
			Examples:    []string{"pii-disclosure", "credential-leakage", "data-extraction"},
		},
		{
			ID:          "LLM07",
			Code:        "llm07-insecure-plugin",
			Name:        "Insecure Plugin Design",
			Description: "Flawed LLM plugin interfaces enabling exploitation",
			Keywords:    []string{"plugin", "extension", "interface", "api", "integration", "addon"},
			Patterns:    []string{"plugin.*vulnerability", "unsafe.*interface", "api.*flaw"},
			Examples:    []string{"plugin-exploitation", "interface-abuse", "api-vulnerabilities"},
		},
		{
			ID:          "LLM08",
			Code:        "llm08-excessive-agency",
			Name:        "Excessive Agency",
			Description: "LLM performing actions beyond intended authority",
			Keywords:    []string{"agency", "authorization", "permission", "privilege", "escalation", "unauthorized"},
			Patterns:    []string{"privilege.*escalation", "unauthorized.*action", "excessive.*permission"},
			Examples:    []string{"unauthorized-actions", "privilege-escalation", "scope-creep"},
		},
		{
			ID:          "LLM09",
			Code:        "llm09-overreliance",
			Name:        "Overreliance",
			Description: "Excessive dependence on LLM outputs without verification",
			Keywords:    []string{"overreliance", "hallucination", "misinformation", "accuracy", "verification"},
			Patterns:    []string{"hallucination.*accept", "unverified.*output", "blind.*trust"},
			Examples:    []string{"hallucination-acceptance", "misinformation-spread", "accuracy-issues"},
		},
		{
			ID:          "LLM10",
			Code:        "llm10-model-theft",
			Name:        "Model Theft",
			Description: "Unauthorized access, copying, or extraction of proprietary models",
			Keywords:    []string{"theft", "extraction", "stealing", "piracy", "intellectual", "proprietary"},
			Patterns:    []string{"model.*extraction", "weight.*theft", "architecture.*reverse"},
			Examples:    []string{"model-extraction", "weight-stealing", "ip-theft"},
		},
	}
// Collect templates from directory
func collectTemplates(dir string) ([]string, error) {
	var templates []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
			templates = append(templates, path)
		}

		return nil
	})

	return templates, err

// Collect templates from bundle
func collectTemplatesFromBundle(bundlePath string) ([]string, error) {
	b, err := bundle.LoadBundle(bundlePath)
	if err != nil {
		return nil, err
	}

	var templates []string
	for _, item := range b.Manifest.Content {
		if item.Type == bundle.TemplateContentType {
			templates = append(templates, item.Path)
		}
	}

	return templates, nil

// Categorize templates
func categorizeTemplates(templates []string) []CategoryMapping {
	categories := getDetailedOWASPCategories()
	var mappings []CategoryMapping

	for _, template := range templates {
		bestMatch := categorizeTemplate(template, categories)
		mappings = append(mappings, bestMatch)
	}

	return mappings

// Categorize single template
func categorizeTemplate(templatePath string, categories []OWASPCategory) CategoryMapping {
	mapping := CategoryMapping{
		Template:    filepath.Base(templatePath),
		Path:        templatePath,
		Category:    "uncategorized",
		Confidence:  0.0,
		MatchedKeys: []string{},
	}

	// Read template content if possible
	content := ""
	if data, err := os.ReadFile(filepath.Clean(templatePath)); err == nil {
		content = string(data)
	}

	// Combine path and content for analysis
	searchText := strings.ToLower(templatePath + " " + content)

	// Score each category
	for _, category := range categories {
		score := 0.0
		var matched []string

		// Check keywords
		for _, keyword := range category.Keywords {
			if strings.Contains(searchText, strings.ToLower(keyword)) {
				score += 10.0
				matched = append(matched, keyword)
			}
		}

		// Check if path contains category code
		if strings.Contains(strings.ToLower(templatePath), category.Code) {
			score += 50.0
			matched = append(matched, "path:"+category.Code)
		}

		// Check patterns in content
		for _, pattern := range category.Patterns {
			if strings.Contains(searchText, strings.ToLower(pattern)) {
				score += 15.0
				matched = append(matched, "pattern:"+pattern)
			}
		}

		// Check examples
		for _, example := range category.Examples {
			if strings.Contains(strings.ToLower(templatePath), example) {
				score += 25.0
				matched = append(matched, "example:"+example)
			}
		}

		// Update best match
		if score > mapping.Confidence {
			mapping.Category = category.Code
			mapping.Confidence = score
			mapping.MatchedKeys = matched
		}
	}

	// Normalize confidence to 0-100
	if mapping.Confidence > 100 {
		mapping.Confidence = 100
	}

	return mapping

// Generate category report
func generateCategoryReport(mappings []CategoryMapping) *CategoryReport {
	report := &CategoryReport{
		TotalTemplates: len(mappings),
		Categories:     make(map[string]*CategoryStats),
		Mappings:       mappings,
		Coverage:       make(map[string]float64),
	}

	// Initialize categories
	for _, cat := range getDetailedOWASPCategories() {
		report.Categories[cat.Code] = &CategoryStats{
			Count:     0,
			Templates: []string{},
			SubTypes:  make(map[string]int),
		}
	}

	// Count categorized templates
	for _, mapping := range mappings {
		if mapping.Category != "uncategorized" {
			report.CategorizedCount++
			if stats, ok := report.Categories[mapping.Category]; ok {
				stats.Count++
				stats.Templates = append(stats.Templates, mapping.Template)

				// Extract subtype from template name
				if subtype := extractSubtype(mapping.Template); subtype != "" {
					stats.SubTypes[subtype]++
				}
			}
		} else {
			report.UncategorizedCount++
		}
	}

	// Calculate percentages and coverage
	for code, stats := range report.Categories {
		if report.TotalTemplates > 0 {
			stats.Percentage = float64(stats.Count) / float64(report.TotalTemplates) * 100
		}

		// Coverage: 1.0 if has templates, 0.0 if not
		if stats.Count > 0 {
			report.Coverage[code] = 1.0
		} else {
			report.Coverage[code] = 0.0
		}
	}

	return report

// Extract subtype from template name
func extractSubtype(templateName string) string {
	// Remove extension
	name := strings.TrimSuffix(templateName, filepath.Ext(templateName))

	// Common subtypes
	subtypes := []string{
		"direct", "indirect", "basic", "advanced",
		"simple", "complex", "bypass", "evasion",
		"extraction", "manipulation", "detection",
	}

	nameLower := strings.ToLower(name)
	for _, subtype := range subtypes {
		if strings.Contains(nameLower, subtype) {
			return subtype
		}
	}

	// Check for version pattern (v1, v2, etc.)
	if strings.Contains(name, "_v") {
		parts := strings.Split(name, "_v")
		if len(parts) > 1 {
			return "v" + parts[1]
		}
	}

	return ""

// Display category report
func displayCategoryReport(report *CategoryReport) {
	fmt.Println()
	color.Cyan("📊 Categorization Results")
	fmt.Println(strings.Repeat("-", 50))

	fmt.Printf("Total Templates: %d\n", report.TotalTemplates)
	fmt.Printf("Categorized: %d (%.1f%%)\n",
		report.CategorizedCount,
		float64(report.CategorizedCount)/float64(report.TotalTemplates)*100)

	if report.UncategorizedCount > 0 {
		color.Yellow("Uncategorized: %d (%.1f%%)\n",
			report.UncategorizedCount,
			float64(report.UncategorizedCount)/float64(report.TotalTemplates)*100)
	}

	fmt.Println("\nCategory Distribution:")
	fmt.Println(strings.Repeat("-", 50))

	// Sort categories by count
	type catCount struct {
		Code  string
		Stats *CategoryStats
	}

	var sorted []catCount
	for code, stats := range report.Categories {
		sorted = append(sorted, catCount{code, stats})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Stats.Count > sorted[j].Stats.Count
	})

	// Display categories
	for _, cc := range sorted {
		cat := getCategoryByCode(cc.Code)
		if cat == nil {
			continue
		}

		if cc.Stats.Count > 0 {
			color.Green("%-30s %3d templates (%.1f%%)",
				fmt.Sprintf("%s: %s", cat.ID, cat.Name),
				cc.Stats.Count,
				cc.Stats.Percentage)

			// Show subtypes if any
			if len(cc.Stats.SubTypes) > 0 {
				fmt.Print("  Subtypes: ")
				var subtypes []string
				for subtype, count := range cc.Stats.SubTypes {
					subtypes = append(subtypes, fmt.Sprintf("%s(%d)", subtype, count))
				}
				fmt.Println(strings.Join(subtypes, ", "))
			}
		} else {
			color.Red("%-30s   0 templates ⚠️",
				fmt.Sprintf("%s: %s", cat.ID, cat.Name))
		}
	}

// Get category by code
func getCategoryByCode(code string) *OWASPCategory {
	for _, cat := range getDetailedOWASPCategories() {
		if cat.Code == code {
			return &cat
		}
	}
	return nil
// Organize templates by category directory
func organizeByCategoryDirectory(mappings []CategoryMapping, sourceDir, outputDir string) error {
	color.Cyan("\n📁 Organizing templates by category...")

	// Create output directory
	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create category directories and info files
	for _, cat := range getDetailedOWASPCategories() {
		catDir := filepath.Join(outputDir, cat.Code)
		if err := os.MkdirAll(catDir, 0700); err != nil {
			return fmt.Errorf("failed to create category directory: %w", err)
		}

		// Create category info file
		if categorizeMetadata {
			infoPath := filepath.Join(catDir, "README.md")
			if err := createCategoryInfo(cat, infoPath); err != nil {
				color.Red("  ⚠️  Failed to create info for %s: %v", cat.Code, err)
			}
		}
	}

	// Create uncategorized directory
	uncategorizedDir := filepath.Join(outputDir, "uncategorized")
	if err := os.MkdirAll(uncategorizedDir, 0700); err != nil {
		return fmt.Errorf("failed to create uncategorized directory: %w", err)
	}

	// Process each template
	for _, mapping := range mappings {
		targetDir := filepath.Join(outputDir, mapping.Category)
		if mapping.Category == "uncategorized" {
			targetDir = uncategorizedDir
		}
		targetPath := filepath.Join(targetDir, mapping.Template)

		// Skip if already exists
		if _, err := os.Stat(targetPath); err == nil {
			color.Yellow("  Skip: %s (already exists)", mapping.Template)
			continue
		}

		// Copy or symlink
		if categorizeSymlink {
			// Create absolute path for symlink
			absSource, err := filepath.Abs(mapping.Path)
			if err != nil {
				color.Red("  ✗ Failed to resolve %s: %v", mapping.Path, err)
				continue
			}

			if err := os.Symlink(absSource, targetPath); err != nil {
				color.Red("  ✗ Failed to symlink %s: %v", mapping.Template, err)
			} else {
				color.Green("  ✓ Linked: %s → %s", mapping.Template, mapping.Category)
			}
		} else if categorizeCopy {
			// Read and copy file
			data, err := os.ReadFile(filepath.Clean(mapping.Path))
			if err != nil {
				color.Red("  ✗ Failed to read %s: %v", mapping.Path, err)
				continue
			}

			if err := os.WriteFile(filepath.Clean(targetPath, data, 0600)); err != nil {
				color.Red("  ✗ Failed to copy %s: %v", mapping.Template, err)
			} else {
				color.Green("  ✓ Copied: %s → %s", mapping.Template, mapping.Category)
			}
		}
	}

	// Create mapping file
	if categorizeMetadata {
		mappingPath := filepath.Join(outputDir, "category-mappings.json")
		mappingData, _ := json.MarshalIndent(mappings, "", "  ")
		os.WriteFile(filepath.Clean(mappingPath, mappingData, 0600))
	}

	fmt.Println()
	color.Green("✅ Templates organized in: %s", outputDir)

	return nil

// Create category info file
func createCategoryInfo(cat OWASPCategory, infoPath string) error {
	var content strings.Builder

	content.WriteString(fmt.Sprintf("# %s: %s\n\n", cat.ID, cat.Name))
	content.WriteString(fmt.Sprintf("## Description\n%s\n\n", cat.Description))

	content.WriteString("## Keywords\n")
	for _, keyword := range cat.Keywords {
		content.WriteString(fmt.Sprintf("- %s\n", keyword))
	}
	content.WriteString("\n")

	content.WriteString("## Common Patterns\n")
	for _, pattern := range cat.Patterns {
		content.WriteString(fmt.Sprintf("- `%s`\n", pattern))
	}
	content.WriteString("\n")

	content.WriteString("## Example Templates\n")
	for _, example := range cat.Examples {
		content.WriteString(fmt.Sprintf("- %s\n", example))
	}
	content.WriteString("\n")

	content.WriteString("## Testing Guide\n")
	content.WriteString("1. Review template parameters\n")
	content.WriteString("2. Configure test environment\n")
	content.WriteString("3. Execute tests with appropriate safety measures\n")
	content.WriteString("4. Document findings and remediation steps\n")

	return os.WriteFile(filepath.Clean(infoPath, []byte(content.String())), 0600)

// Save category report as JSON
func saveCategoryReport(report *CategoryReport, outputPath string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Clean(outputPath, data, 0600)); err != nil {
		return err
	}

	color.Green("\n✅ Report saved to: %s", outputPath)
	return nil

// Save category report as Markdown
func saveCategoryReportMarkdown(report *CategoryReport, outputPath string) error {
	var content strings.Builder

	content.WriteString("# OWASP LLM Top 10 Template Categorization Report\n\n")
	content.WriteString(fmt.Sprintf("**Generated**: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	content.WriteString("## Summary\n\n")
	content.WriteString(fmt.Sprintf("- **Total Templates**: %d\n", report.TotalTemplates))
	content.WriteString(fmt.Sprintf("- **Categorized**: %d (%.1f%%)\n",
		report.CategorizedCount,
		float64(report.CategorizedCount)/float64(report.TotalTemplates)*100))
	content.WriteString(fmt.Sprintf("- **Uncategorized**: %d\n\n", report.UncategorizedCount))

	content.WriteString("## Category Coverage\n\n")
	content.WriteString("| Category | Templates | Coverage | Status |\n")
	content.WriteString("|----------|-----------|----------|--------|\n")

	for _, cat := range getDetailedOWASPCategories() {
		stats := report.Categories[cat.Code]
		status := "❌ Missing"
		if stats.Count > 0 {
			status = "✅ Covered"
		}

		content.WriteString(fmt.Sprintf("| %s: %s | %d | %.1f%% | %s |\n",
			cat.ID, cat.Name, stats.Count, stats.Percentage, status))
	}

	content.WriteString("\n## Template Mappings\n\n")
	content.WriteString("| Template | Category | Confidence | Matched Keywords |\n")
	content.WriteString("|----------|----------|------------|------------------|\n")

	// Sort mappings by confidence
	sort.Slice(report.Mappings, func(i, j int) bool {
		return report.Mappings[i].Confidence > report.Mappings[j].Confidence
	})

	// Show top mappings
	for i, mapping := range report.Mappings {
		if i >= 20 && mapping.Confidence < 50 {
			break // Limit output
		}

		content.WriteString(fmt.Sprintf("| %s | %s | %.0f%% | %s |\n",
			mapping.Template,
			mapping.Category,
			mapping.Confidence,
			strings.Join(mapping.MatchedKeys, ", ")))
	}

	if len(report.Mappings) > 20 {
		content.WriteString(fmt.Sprintf("\n*... and %d more templates*\n", len(report.Mappings)-20))
	}

	content.WriteString("\n## Recommendations\n\n")

	// Find missing categories
	var missing []string
	for code, stats := range report.Categories {
		if stats.Count == 0 {
			if cat := getCategoryByCode(code); cat != nil {
				missing = append(missing, fmt.Sprintf("%s: %s", cat.ID, cat.Name))
			}
		}
	}

	if len(missing) > 0 {
		content.WriteString("### Missing Coverage\n\n")
		content.WriteString("The following OWASP categories have no test templates:\n\n")
		for _, m := range missing {
			content.WriteString(fmt.Sprintf("- %s\n", m))
		}
		content.WriteString("\n")
	}

	if report.UncategorizedCount > 0 {
		content.WriteString("### Uncategorized Templates\n\n")
		content.WriteString(fmt.Sprintf("%d templates could not be automatically categorized. ", report.UncategorizedCount))
		content.WriteString("Consider:\n\n")
		content.WriteString("1. Reviewing template names and content\n")
		content.WriteString("2. Adding category indicators to file paths\n")
		content.WriteString("3. Including OWASP keywords in template metadata\n")
	}

	if err := os.WriteFile(filepath.Clean(outputPath, []byte(content.String())), 0600); err != nil {
		return err
	}

	color.Green("\n✅ Report saved to: %s", outputPath)
	return nil

// Show category recommendations
func showCategoryRecommendations(report *CategoryReport) {
	fmt.Println()
	color.Cyan("💡 Recommendations")
	fmt.Println(strings.Repeat("-", 50))

	// Check coverage
	missingCount := 0
	lowCoverage := []string{}

	for code, stats := range report.Categories {
		if stats.Count == 0 {
			missingCount++
		} else if stats.Count < 3 {
			if cat := getCategoryByCode(code); cat != nil {
				lowCoverage = append(lowCoverage, fmt.Sprintf("%s: %s (%d templates)",
					cat.ID, cat.Name, stats.Count))
			}
		}
	}

	if missingCount > 0 {
		color.Yellow("⚠️  Missing coverage for %d categories", missingCount)
		fmt.Println("   Consider adding templates for comprehensive security testing")
	}

	if len(lowCoverage) > 0 {
		color.Yellow("\n⚠️  Low coverage categories:")
		for _, cat := range lowCoverage {
			fmt.Printf("   - %s\n", cat)
		}
		fmt.Println("   Recommend at least 3 templates per category")
	}

	if report.UncategorizedCount > 0 {
		color.Yellow("\n⚠️  %d uncategorized templates", report.UncategorizedCount)
		fmt.Println("   Review and update template metadata for better organization")
	}

	// Overall assessment
	coverageScore := float64(report.CategorizedCount) / float64(report.TotalTemplates) * 100
	fmt.Println()

	if coverageScore >= 90 && missingCount == 0 {
		color.Green("✅ Excellent OWASP LLM Top 10 coverage!")
	} else if coverageScore >= 70 && missingCount <= 2 {
		color.Yellow("🔶 Good coverage with room for improvement")
	} else {
		color.Red("❌ Significant gaps in security test coverage")
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
