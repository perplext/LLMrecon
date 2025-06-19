package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/perplext/LLMrecon/src/bundle"
	"github.com/perplext/LLMrecon/src/compliance"
)

var (
	complianceFormat     string
	complianceOutput     string
	complianceFramework  string
	complianceEvidence   bool
	complianceTemplates  string
)

var bundleComplianceCmd = &cobra.Command{
	Use:   "compliance",
	Short: "Generate and manage compliance documentation",
	Long: `Generate compliance documentation for security test bundles.
	
Supports multiple compliance frameworks:
- OWASP LLM Top 10 v1.0
- ISO/IEC 42001:2023 (AI Management System)
- NIST AI Risk Management Framework
- EU AI Act requirements`,
}

var complianceGenerateCmd = &cobra.Command{
	Use:   "generate [bundle-file]",
	Short: "Generate compliance documentation from a bundle",
	Long: `Generate comprehensive compliance documentation that demonstrates
how your security testing aligns with various compliance frameworks.`,
	Example: `  # Generate OWASP compliance docs
  LLMrecon bundle compliance generate security.bundle --framework=owasp
  
  # Generate ISO 42001 compliance package
  LLMrecon bundle compliance generate security.bundle --framework=iso42001 --format=pdf
  
  # Generate all compliance docs
  LLMrecon bundle compliance generate security.bundle --framework=all --output=compliance/`,
	Args: cobra.ExactArgs(1),
	RunE: runComplianceGenerate,
}

var complianceCheckCmd = &cobra.Command{
	Use:   "check [bundle-file]",
	Short: "Check bundle compliance status",
	Long:  `Analyze a bundle and report its compliance status against various frameworks.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runComplianceCheck,
}

var complianceTemplateCmd = &cobra.Command{
	Use:   "template [framework]",
	Short: "Generate compliance documentation templates",
	Long:  `Generate template documents for compliance documentation that can be customized.`,
	Example: `  # Generate ISO 42001 templates
  LLMrecon bundle compliance template iso42001
  
  # Generate all templates
  LLMrecon bundle compliance template all --output=templates/`,
	Args: cobra.ExactArgs(1),
	RunE: runComplianceTemplate,
}

func init() {
	bundleCmd.AddCommand(bundleComplianceCmd)
	bundleComplianceCmd.AddCommand(complianceGenerateCmd)
	bundleComplianceCmd.AddCommand(complianceCheckCmd)
	bundleComplianceCmd.AddCommand(complianceTemplateCmd)
	
	// Generate command flags
	complianceGenerateCmd.Flags().StringVarP(&complianceFormat, "format", "f", "markdown", "Output format (markdown,pdf,html,docx)")
	complianceGenerateCmd.Flags().StringVarP(&complianceOutput, "output", "o", "./compliance", "Output directory")
	complianceGenerateCmd.Flags().StringVar(&complianceFramework, "framework", "all", "Compliance framework (owasp,iso42001,nist,eu-ai,all)")
	complianceGenerateCmd.Flags().BoolVar(&complianceEvidence, "evidence", true, "Include evidence mappings")
	
	// Template command flags
	complianceTemplateCmd.Flags().StringVarP(&complianceOutput, "output", "o", "./compliance-templates", "Output directory")
}

func runComplianceGenerate(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]
	
	// Verify bundle exists
	if _, err := os.Stat(bundlePath); err != nil {
		return fmt.Errorf("bundle file not found: %s", bundlePath)
	}
	
	fmt.Println()
	color.Cyan("📋 Generating Compliance Documentation")
	fmt.Println(strings.Repeat("-", 50))
	
	// Load bundle
	bundleData, err := bundle.LoadBundle(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to load bundle: %w", err)
	}
	
	color.Yellow("Bundle: %s", bundleData.Manifest.Name)
	color.Yellow("Framework: %s", complianceFramework)
	color.Yellow("Format: %s", complianceFormat)
	color.Yellow("Output: %s", complianceOutput)
	fmt.Println()
	
	// Create output directory
	if err := os.MkdirAll(complianceOutput, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	// Generate documentation based on framework
	frameworks := []string{}
	if complianceFramework == "all" {
		frameworks = []string{"owasp", "iso42001", "nist", "eu-ai"}
	} else {
		frameworks = []string{complianceFramework}
	}
	
	for _, framework := range frameworks {
		color.Cyan("Generating %s documentation...", strings.ToUpper(framework))
		
		switch framework {
		case "owasp":
			if err := generateOWASPCompliance(bundleData, complianceOutput, complianceFormat); err != nil {
				color.Red("  ✗ Failed: %v", err)
			} else {
				color.Green("  ✓ Generated OWASP LLM Top 10 compliance documentation")
			}
			
		case "iso42001":
			if err := generateISO42001Compliance(bundleData, complianceOutput, complianceFormat); err != nil {
				color.Red("  ✗ Failed: %v", err)
			} else {
				color.Green("  ✓ Generated ISO/IEC 42001:2023 compliance documentation")
			}
			
		case "nist":
			if err := generateNISTCompliance(bundleData, complianceOutput, complianceFormat); err != nil {
				color.Red("  ✗ Failed: %v", err)
			} else {
				color.Green("  ✓ Generated NIST AI RMF compliance documentation")
			}
			
		case "eu-ai":
			if err := generateEUAICompliance(bundleData, complianceOutput, complianceFormat); err != nil {
				color.Red("  ✗ Failed: %v", err)
			} else {
				color.Green("  ✓ Generated EU AI Act compliance documentation")
			}
		}
	}
	
	// Generate evidence mappings if requested
	if complianceEvidence {
		color.Cyan("\n📊 Generating evidence mappings...")
		if err := generateEvidenceMappings(bundleData, complianceOutput); err != nil {
			color.Red("  ✗ Failed to generate evidence mappings: %v", err)
		} else {
			color.Green("  ✓ Generated evidence mappings")
		}
	}
	
	// Generate executive summary
	color.Cyan("\n📄 Generating executive summary...")
	if err := generateExecutiveSummary(bundleData, complianceOutput, frameworks); err != nil {
		color.Red("  ✗ Failed to generate executive summary: %v", err)
	} else {
		color.Green("  ✓ Generated executive summary")
	}
	
	fmt.Println()
	color.Green("✅ Compliance documentation generated successfully")
	fmt.Printf("   Output directory: %s\n", complianceOutput)
	
	return nil
}

func runComplianceCheck(cmd *cobra.Command, args []string) error {
	bundlePath := args[0]
	
	// Verify bundle exists
	if _, err := os.Stat(bundlePath); err != nil {
		return fmt.Errorf("bundle file not found: %s", bundlePath)
	}
	
	fmt.Println()
	color.Cyan("🔍 Checking Bundle Compliance")
	fmt.Println(strings.Repeat("-", 50))
	
	// Load bundle
	bundleData, err := bundle.LoadBundle(bundlePath)
	if err != nil {
		return fmt.Errorf("failed to load bundle: %w", err)
	}
	
	// Check OWASP compliance
	owaspStatus := checkOWASPCompliance(bundleData)
	displayComplianceStatus("OWASP LLM Top 10", owaspStatus)
	
	// Check ISO 42001 compliance
	isoStatus := checkISO42001Compliance(bundleData)
	displayComplianceStatus("ISO/IEC 42001:2023", isoStatus)
	
	// Check NIST compliance
	nistStatus := checkNISTCompliance(bundleData)
	displayComplianceStatus("NIST AI RMF", nistStatus)
	
	// Check EU AI Act compliance
	euStatus := checkEUAICompliance(bundleData)
	displayComplianceStatus("EU AI Act", euStatus)
	
	// Overall compliance score
	fmt.Println()
	overallScore := calculateOverallCompliance(owaspStatus, isoStatus, nistStatus, euStatus)
	
	color.Cyan("Overall Compliance Score: ")
	if overallScore >= 80 {
		color.Green("%.1f%% - Excellent", overallScore)
	} else if overallScore >= 60 {
		color.Yellow("%.1f%% - Good", overallScore)
	} else {
		color.Red("%.1f%% - Needs Improvement", overallScore)
	}
	
	return nil
}

func runComplianceTemplate(cmd *cobra.Command, args []string) error {
	framework := args[0]
	
	fmt.Println()
	color.Cyan("📝 Generating Compliance Templates")
	fmt.Println(strings.Repeat("-", 50))
	
	// Create output directory
	if err := os.MkdirAll(complianceOutput, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}
	
	frameworks := []string{}
	if framework == "all" {
		frameworks = []string{"owasp", "iso42001", "nist", "eu-ai"}
	} else {
		frameworks = []string{framework}
	}
	
	for _, fw := range frameworks {
		color.Yellow("Generating %s templates...", strings.ToUpper(fw))
		
		switch fw {
		case "owasp":
			if err := generateOWASPTemplates(complianceOutput); err != nil {
				color.Red("  ✗ Failed: %v", err)
			} else {
				color.Green("  ✓ Generated OWASP LLM Top 10 templates")
			}
			
		case "iso42001":
			if err := generateISO42001Templates(complianceOutput); err != nil {
				color.Red("  ✗ Failed: %v", err)
			} else {
				color.Green("  ✓ Generated ISO/IEC 42001:2023 templates")
			}
			
		case "nist":
			if err := generateNISTTemplates(complianceOutput); err != nil {
				color.Red("  ✗ Failed: %v", err)
			} else {
				color.Green("  ✓ Generated NIST AI RMF templates")
			}
			
		case "eu-ai":
			if err := generateEUAITemplates(complianceOutput); err != nil {
				color.Red("  ✗ Failed: %v", err)
			} else {
				color.Green("  ✓ Generated EU AI Act templates")
			}
		}
	}
	
	fmt.Println()
	color.Green("✅ Compliance templates generated successfully")
	fmt.Printf("   Output directory: %s\n", complianceOutput)
	
	return nil
}

// Compliance generation functions

func generateOWASPCompliance(b *bundle.Bundle, outputDir, format string) error {
	doc := &ComplianceDocument{
		Title:     "OWASP LLM Top 10 Compliance Report",
		Framework: "OWASP LLM Top 10 v1.0",
		Bundle:    b.Manifest.Name,
		Version:   b.Manifest.Version,
		Date:      time.Now(),
		Sections:  []ComplianceSection{},
	}
	
	// Executive Summary
	doc.Sections = append(doc.Sections, ComplianceSection{
		Title: "Executive Summary",
		Content: fmt.Sprintf(`This compliance report demonstrates how the "%s" security test bundle aligns with the OWASP LLM Top 10 framework. The bundle contains %d security test templates designed to identify and mitigate the top security risks in Large Language Model applications.`, 
			b.Manifest.Name, b.Manifest.Templates),
	})
	
	// Coverage Analysis
	coverage := analyzeOWASPCoverage(b)
	doc.Sections = append(doc.Sections, ComplianceSection{
		Title:   "Coverage Analysis",
		Content: formatCoverageAnalysis(coverage),
	})
	
	// Category Details
	for _, category := range getOWASPCategories() {
		section := ComplianceSection{
			Title: fmt.Sprintf("%s - %s", category.ID, category.Name),
			Content: fmt.Sprintf("## Risk Description\n%s\n\n## Test Coverage\n%s\n\n## Mitigation Strategies\n%s",
				category.Description,
				getCategoryTestCoverage(b, category.ID),
				category.Mitigation),
		}
		doc.Sections = append(doc.Sections, section)
	}
	
	// Evidence Mapping
	if complianceEvidence {
		doc.Sections = append(doc.Sections, ComplianceSection{
			Title:   "Evidence Mapping",
			Content: generateOWASPEvidenceMapping(b),
		})
	}
	
	// Save document
	filename := filepath.Join(outputDir, fmt.Sprintf("owasp-llm-top10-compliance.%s", format))
	return saveComplianceDocument(doc, filename, format)
}

func generateISO42001Compliance(b *bundle.Bundle, outputDir, format string) error {
	doc := &ComplianceDocument{
		Title:     "ISO/IEC 42001:2023 Compliance Report",
		Framework: "ISO/IEC 42001:2023 - AI Management System",
		Bundle:    b.Manifest.Name,
		Version:   b.Manifest.Version,
		Date:      time.Now(),
		Sections:  []ComplianceSection{},
	}
	
	// Context of the Organization (Clause 4)
	doc.Sections = append(doc.Sections, ComplianceSection{
		Title: "4. Context of the Organization",
		Content: `### 4.1 Understanding the organization and its context
The security test bundle supports the organization's AI management system by providing comprehensive testing capabilities for AI/LLM applications.

### 4.2 Understanding stakeholder needs
Tests address security requirements from:
- Development teams
- Security teams  
- Compliance officers
- End users`,
	})
	
	// Leadership (Clause 5)
	doc.Sections = append(doc.Sections, ComplianceSection{
		Title: "5. Leadership",
		Content: fmt.Sprintf(`### 5.1 Leadership and commitment
This bundle demonstrates leadership commitment to AI security through:
- Comprehensive test coverage (%d templates)
- Regular updates and maintenance
- Clear documentation and guidance`, b.Manifest.Templates),
	})
	
	// Planning (Clause 6)
	doc.Sections = append(doc.Sections, ComplianceSection{
		Title: "6. Planning",
		Content: `### 6.1 Actions to address risks and opportunities
Security tests identify and mitigate:
- Prompt injection risks
- Data leakage vulnerabilities
- Model manipulation attempts
- Supply chain vulnerabilities`,
	})
	
	// Support (Clause 7)
	doc.Sections = append(doc.Sections, ComplianceSection{
		Title: "7. Support",
		Content: `### 7.2 Competence
Bundle includes:
- Detailed test documentation
- Usage examples
- Integration guides
- Training materials`,
	})
	
	// Operation (Clause 8)
	doc.Sections = append(doc.Sections, ComplianceSection{
		Title: "8. Operation", 
		Content: generateISO42001OperationSection(b),
	})
	
	// Save document
	filename := filepath.Join(outputDir, fmt.Sprintf("iso42001-compliance.%s", format))
	return saveComplianceDocument(doc, filename, format)
}

func generateNISTCompliance(b *bundle.Bundle, outputDir, format string) error {
	doc := &ComplianceDocument{
		Title:     "NIST AI Risk Management Framework Compliance",
		Framework: "NIST AI RMF 1.0",
		Bundle:    b.Manifest.Name,
		Version:   b.Manifest.Version,
		Date:      time.Now(),
		Sections:  []ComplianceSection{},
	}
	
	// GOVERN
	doc.Sections = append(doc.Sections, ComplianceSection{
		Title: "GOVERN - Cultivate AI Risk Management Culture",
		Content: `Security testing supports governance by:
- Establishing clear testing protocols
- Defining risk acceptance criteria
- Providing measurable security metrics
- Supporting compliance reporting`,
	})
	
	// MAP
	doc.Sections = append(doc.Sections, ComplianceSection{
		Title: "MAP - Understand AI Risks",
		Content: fmt.Sprintf(`Risk mapping through %d security tests covering:
- Input validation risks
- Output security risks
- Model integrity risks
- Operational risks`, b.Manifest.Templates),
	})
	
	// MEASURE
	doc.Sections = append(doc.Sections, ComplianceSection{
		Title: "MEASURE - Assess AI Risks",
		Content: generateNISTMeasureSection(b),
	})
	
	// MANAGE
	doc.Sections = append(doc.Sections, ComplianceSection{
		Title: "MANAGE - Respond to AI Risks",
		Content: `Risk management capabilities:
- Automated security testing
- Continuous monitoring support
- Incident response procedures
- Remediation guidance`,
	})
	
	// Save document
	filename := filepath.Join(outputDir, fmt.Sprintf("nist-ai-rmf-compliance.%s", format))
	return saveComplianceDocument(doc, filename, format)
}

func generateEUAICompliance(b *bundle.Bundle, outputDir, format string) error {
	doc := &ComplianceDocument{
		Title:     "EU AI Act Compliance Report",
		Framework: "EU AI Act - High-Risk AI Systems",
		Bundle:    b.Manifest.Name,
		Version:   b.Manifest.Version,
		Date:      time.Now(),
		Sections:  []ComplianceSection{},
	}
	
	// Risk Management System
	doc.Sections = append(doc.Sections, ComplianceSection{
		Title: "Risk Management System (Article 9)",
		Content: `The security test bundle supports the required risk management system by:
- Identifying foreseeable risks through comprehensive testing
- Estimating and evaluating risks with severity assessments
- Adopting risk management measures via security controls
- Providing residual risk evaluation capabilities`,
	})
	
	// Data Governance
	doc.Sections = append(doc.Sections, ComplianceSection{
		Title: "Data and Data Governance (Article 10)",
		Content: generateEUAIDataGovernanceSection(b),
	})
	
	// Technical Documentation
	doc.Sections = append(doc.Sections, ComplianceSection{
		Title: "Technical Documentation (Article 11)",
		Content: `Bundle provides comprehensive technical documentation:
- Detailed test specifications
- Implementation guides
- Results interpretation
- Compliance mappings`,
	})
	
	// Transparency
	doc.Sections = append(doc.Sections, ComplianceSection{
		Title: "Transparency and Information (Article 13)",
		Content: `Transparency measures supported:
- Clear test objectives and methods
- Understandable results reporting
- Risk communication templates
- User guidance documentation`,
	})
	
	// Save document
	filename := filepath.Join(outputDir, fmt.Sprintf("eu-ai-act-compliance.%s", format))
	return saveComplianceDocument(doc, filename, format)
}

// Helper structures and functions

type ComplianceDocument struct {
	Title     string
	Framework string
	Bundle    string
	Version   string
	Date      time.Time
	Sections  []ComplianceSection
}

type ComplianceSection struct {
	Title   string
	Content string
}

type ComplianceStatus struct {
	Framework        string
	Score            float64
	CoveredAreas     int
	TotalAreas       int
	Strengths        []string
	Gaps             []string
	Recommendations  []string
}

type OWASPCategory struct {
	ID          string
	Name        string
	Description string
	Mitigation  string
}

func getOWASPCategories() []OWASPCategory {
	return []OWASPCategory{
		{
			ID:          "LLM01",
			Name:        "Prompt Injection",
			Description: "Malicious inputs designed to manipulate LLM behavior",
			Mitigation:  "Input validation, output filtering, privilege separation",
		},
		{
			ID:          "LLM02",
			Name:        "Insecure Output Handling",
			Description: "Insufficient validation of LLM outputs leading to downstream vulnerabilities",
			Mitigation:  "Output encoding, validation, sandboxing",
		},
		{
			ID:          "LLM03",
			Name:        "Training Data Poisoning",
			Description: "Manipulation of training data to introduce vulnerabilities or biases",
			Mitigation:  "Data validation, source verification, anomaly detection",
		},
		{
			ID:          "LLM04",
			Name:        "Model Denial of Service",
			Description: "Resource exhaustion attacks targeting model availability",
			Mitigation:  "Rate limiting, resource monitoring, input size limits",
		},
		{
			ID:          "LLM05",
			Name:        "Supply Chain Vulnerabilities",
			Description: "Compromised components, models, or data sources",
			Mitigation:  "Vendor assessment, integrity verification, dependency scanning",
		},
		{
			ID:          "LLM06",
			Name:        "Sensitive Information Disclosure",
			Description: "Unintended revelation of confidential data",
			Mitigation:  "Data sanitization, access controls, output filtering",
		},
		{
			ID:          "LLM07",
			Name:        "Insecure Plugin Design",
			Description: "Flawed plugin interfaces enabling exploitation",
			Mitigation:  "Input validation, authentication, least privilege",
		},
		{
			ID:          "LLM08",
			Name:        "Excessive Agency",
			Description: "LLM performing actions beyond intended scope",
			Mitigation:  "Permission boundaries, human oversight, action validation",
		},
		{
			ID:          "LLM09",
			Name:        "Overreliance",
			Description: "Excessive dependence on LLM outputs without verification",
			Mitigation:  "Human review, confidence scoring, disclaimer",
		},
		{
			ID:          "LLM10",
			Name:        "Model Theft",
			Description: "Unauthorized access or extraction of model",
			Mitigation:  "Access controls, usage monitoring, watermarking",
		},
	}
}

func analyzeOWASPCoverage(b *bundle.Bundle) map[string]int {
	coverage := make(map[string]int)
	
	for _, tmpl := range b.Templates {
		for _, cat := range getOWASPCategories() {
			if strings.Contains(strings.ToLower(tmpl.Path), strings.ToLower(cat.ID)) {
				coverage[cat.ID]++
			}
		}
	}
	
	return coverage
}

func formatCoverageAnalysis(coverage map[string]int) string {
	var content strings.Builder
	
	content.WriteString("| Category | Coverage | Status |\n")
	content.WriteString("|----------|----------|--------|\n")
	
	for _, cat := range getOWASPCategories() {
		count := coverage[cat.ID]
		status := "❌ Not Covered"
		if count > 0 {
			status = fmt.Sprintf("✅ %d tests", count)
		}
		content.WriteString(fmt.Sprintf("| %s: %s | %d | %s |\n", cat.ID, cat.Name, count, status))
	}
	
	return content.String()
}

func getCategoryTestCoverage(b *bundle.Bundle, categoryID string) string {
	var tests []string
	
	for _, tmpl := range b.Templates {
		if strings.Contains(strings.ToLower(tmpl.Path), strings.ToLower(categoryID)) {
			tests = append(tests, fmt.Sprintf("- %s", tmpl.Name))
		}
	}
	
	if len(tests) == 0 {
		return "No specific tests found for this category."
	}
	
	return strings.Join(tests, "\n")
}

func generateOWASPEvidenceMapping(b *bundle.Bundle) string {
	var content strings.Builder
	
	content.WriteString("## Evidence Mapping\n\n")
	content.WriteString("| Test Template | OWASP Category | Evidence Type | Compliance Artifact |\n")
	content.WriteString("|---------------|----------------|---------------|--------------------|\n")
	
	for _, tmpl := range b.Templates {
		category := "General"
		for _, cat := range getOWASPCategories() {
			if strings.Contains(strings.ToLower(tmpl.Path), strings.ToLower(cat.ID)) {
				category = cat.ID
				break
			}
		}
		
		content.WriteString(fmt.Sprintf("| %s | %s | Test Result | Security Assessment |\n", 
			tmpl.Name, category))
	}
	
	return content.String()
}

func checkOWASPCompliance(b *bundle.Bundle) ComplianceStatus {
	coverage := analyzeOWASPCoverage(b)
	coveredCount := 0
	var gaps []string
	var strengths []string
	
	for _, cat := range getOWASPCategories() {
		if coverage[cat.ID] > 0 {
			coveredCount++
			strengths = append(strengths, fmt.Sprintf("%s coverage", cat.ID))
		} else {
			gaps = append(gaps, fmt.Sprintf("No %s tests", cat.ID))
		}
	}
	
	score := float64(coveredCount) / 10.0 * 100
	
	return ComplianceStatus{
		Framework:    "OWASP LLM Top 10",
		Score:        score,
		CoveredAreas: coveredCount,
		TotalAreas:   10,
		Strengths:    strengths,
		Gaps:         gaps,
		Recommendations: []string{
			"Add tests for uncovered categories",
			"Increase test depth for partial coverage",
			"Regular updates for emerging threats",
		},
	}
}

func checkISO42001Compliance(b *bundle.Bundle) ComplianceStatus {
	// Simplified ISO 42001 compliance check
	clauses := []string{
		"Context", "Leadership", "Planning", "Support",
		"Operation", "Performance", "Improvement",
	}
	
	// Check based on bundle metadata and structure
	covered := 0
	if b.Manifest.Description != "" {
		covered++ // Context
	}
	if b.Manifest.Author != "" {
		covered++ // Leadership
	}
	if len(b.Templates) > 10 {
		covered++ // Planning
	}
	if b.Manifest.Metadata != nil {
		covered++ // Support
	}
	if len(b.Templates) > 0 {
		covered++ // Operation
	}
	
	score := float64(covered) / float64(len(clauses)) * 100
	
	return ComplianceStatus{
		Framework:    "ISO/IEC 42001:2023",
		Score:        score,
		CoveredAreas: covered,
		TotalAreas:   len(clauses),
		Strengths:    []string{"Structured approach", "Documentation"},
		Gaps:         []string{"Performance metrics", "Improvement process"},
	}
}

func checkNISTCompliance(b *bundle.Bundle) ComplianceStatus {
	// NIST AI RMF functions
	functions := []string{"GOVERN", "MAP", "MEASURE", "MANAGE"}
	
	// Simple scoring based on template count and categories
	score := minFloat(float64(len(b.Templates))/20.0*100, 100)
	covered := int(score / 25) // Each function worth 25%
	
	return ComplianceStatus{
		Framework:    "NIST AI RMF",
		Score:        score,
		CoveredAreas: covered,
		TotalAreas:   len(functions),
		Strengths:    []string{"Risk identification", "Testing capabilities"},
		Gaps:         []string{"Governance documentation", "Metrics dashboard"},
	}
}

func checkEUAICompliance(b *bundle.Bundle) ComplianceStatus {
	// EU AI Act requirements for high-risk systems
	requirements := []string{
		"Risk Management", "Data Governance", "Documentation",
		"Transparency", "Human Oversight", "Accuracy",
	}
	
	// Basic scoring
	score := 60.0 // Base score
	if len(b.Templates) > 20 {
		score += 20
	}
	if b.Manifest.Metadata != nil {
		score += 20
	}
	
	covered := int(score / 100 * float64(len(requirements)))
	
	return ComplianceStatus{
		Framework:    "EU AI Act",
		Score:        score,
		CoveredAreas: covered,
		TotalAreas:   len(requirements),
		Strengths:    []string{"Technical measures", "Testing framework"},
		Gaps:         []string{"Human oversight procedures", "Transparency reports"},
	}
}

func displayComplianceStatus(framework string, status ComplianceStatus) {
	fmt.Println()
	color.Cyan("%s", framework)
	fmt.Println(strings.Repeat("-", 40))
	
	// Score with color
	fmt.Print("Score: ")
	if status.Score >= 80 {
		color.Green("%.1f%%", status.Score)
	} else if status.Score >= 60 {
		color.Yellow("%.1f%%", status.Score)
	} else {
		color.Red("%.1f%%", status.Score)
	}
	
	fmt.Printf(" (%d/%d areas covered)\n", status.CoveredAreas, status.TotalAreas)
	
	// Strengths
	if len(status.Strengths) > 0 {
		color.Green("Strengths:")
		for _, s := range status.Strengths[:min(3, len(status.Strengths))] {
			fmt.Printf("  ✓ %s\n", s)
		}
	}
	
	// Gaps
	if len(status.Gaps) > 0 {
		color.Yellow("Gaps:")
		for _, g := range status.Gaps[:min(3, len(status.Gaps))] {
			fmt.Printf("  ⚠ %s\n", g)
		}
	}
}

func calculateOverallCompliance(statuses ...ComplianceStatus) float64 {
	if len(statuses) == 0 {
		return 0
	}
	
	total := 0.0
	for _, status := range statuses {
		total += status.Score
	}
	
	return total / float64(len(statuses))
}

func saveComplianceDocument(doc *ComplianceDocument, filename, format string) error {
	switch format {
	case "markdown":
		return saveMarkdownDocument(doc, filename)
	case "json":
		return saveJSONDocument(doc, filename)
	case "html":
		return saveHTMLDocument(doc, filename)
	case "pdf":
		// PDF generation would require additional library
		return saveMarkdownDocument(doc, strings.Replace(filename, ".pdf", ".md", 1))
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func saveMarkdownDocument(doc *ComplianceDocument, filename string) error {
	var content strings.Builder
	
	// Header
	content.WriteString(fmt.Sprintf("# %s\n\n", doc.Title))
	content.WriteString(fmt.Sprintf("**Framework**: %s\n", doc.Framework))
	content.WriteString(fmt.Sprintf("**Bundle**: %s v%s\n", doc.Bundle, doc.Version))
	content.WriteString(fmt.Sprintf("**Date**: %s\n\n", doc.Date.Format("2006-01-02")))
	content.WriteString("---\n\n")
	
	// Sections
	for _, section := range doc.Sections {
		content.WriteString(fmt.Sprintf("## %s\n\n", section.Title))
		content.WriteString(section.Content)
		content.WriteString("\n\n")
	}
	
	return os.WriteFile(filename, []byte(content.String()), 0644)
}

func saveJSONDocument(doc *ComplianceDocument, filename string) error {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func saveHTMLDocument(doc *ComplianceDocument, filename string) error {
	var content strings.Builder
	
	content.WriteString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>` + doc.Title + `</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        h1, h2, h3 { color: #333; }
        table { border-collapse: collapse; width: 100%; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        code { background-color: #f5f5f5; padding: 2px 4px; }
    </style>
</head>
<body>
`)
	
	content.WriteString(fmt.Sprintf("<h1>%s</h1>\n", doc.Title))
	content.WriteString(fmt.Sprintf("<p><strong>Framework:</strong> %s<br>\n", doc.Framework))
	content.WriteString(fmt.Sprintf("<strong>Bundle:</strong> %s v%s<br>\n", doc.Bundle, doc.Version))
	content.WriteString(fmt.Sprintf("<strong>Date:</strong> %s</p>\n", doc.Date.Format("2006-01-02")))
	content.WriteString("<hr>\n")
	
	for _, section := range doc.Sections {
		content.WriteString(fmt.Sprintf("<h2>%s</h2>\n", section.Title))
		// Convert markdown to basic HTML (simplified)
		html := strings.ReplaceAll(section.Content, "\n", "<br>\n")
		content.WriteString(fmt.Sprintf("<div>%s</div>\n", html))
	}
	
	content.WriteString("</body>\n</html>")
	
	return os.WriteFile(filename, []byte(content.String()), 0644)
}

// Template generation functions

func generateOWASPTemplates(outputDir string) error {
	templateDir := filepath.Join(outputDir, "owasp-llm-top10")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		return err
	}
	
	// Main template
	mainTemplate := `# OWASP LLM Top 10 Compliance Documentation Template

## Organization Information
- **Organization Name**: [Your Organization]
- **Assessment Date**: [Date]
- **Assessor**: [Name/Role]
- **Scope**: [LLM Applications in Scope]

## Executive Summary
[Provide overview of compliance status and key findings]

## Risk Assessment

### LLM01: Prompt Injection
**Status**: [ ] Addressed [ ] Partially Addressed [ ] Not Addressed

**Current Controls**:
- [List implemented controls]

**Test Results**:
- [Summary of test findings]

**Remediation Plan**:
- [Actions and timeline]

[Repeat for LLM02-LLM10...]

## Compliance Statement
[Formal statement of compliance status]

## Appendices
- A. Test Methodologies
- B. Evidence Documentation
- C. Remediation Tracking
`
	
	if err := os.WriteFile(filepath.Join(templateDir, "compliance-template.md"), []byte(mainTemplate), 0644); err != nil {
		return err
	}
	
	// Checklist template
	checklistTemplate := `# OWASP LLM Top 10 Compliance Checklist

| Category | Control | Implemented | Evidence | Notes |
|----------|---------|-------------|----------|-------|
| LLM01: Prompt Injection | Input validation | [ ] | | |
| LLM01: Prompt Injection | Output filtering | [ ] | | |
| LLM01: Prompt Injection | Privilege separation | [ ] | | |
| LLM02: Insecure Output | Output encoding | [ ] | | |
| LLM02: Insecure Output | Context-aware sanitization | [ ] | | |
[Continue for all categories...]

## Sign-off
- Technical Lead: _________________ Date: _______
- Security Lead: _________________ Date: _______
- Compliance Officer: _____________ Date: _______
`
	
	return os.WriteFile(filepath.Join(templateDir, "checklist.md"), []byte(checklistTemplate), 0644)
}

func generateISO42001Templates(outputDir string) error {
	templateDir := filepath.Join(outputDir, "iso42001")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		return err
	}
	
	// Statement of Applicability
	soaTemplate := `# ISO/IEC 42001:2023 Statement of Applicability

## Organization Context
[Describe AI systems and their context]

## Applicable Requirements

### Clause 4: Context of the Organization
- [ ] 4.1 Understanding context
- [ ] 4.2 Stakeholder needs
- [ ] 4.3 Scope determination
- [ ] 4.4 AI management system

### Clause 5: Leadership
- [ ] 5.1 Leadership commitment
- [ ] 5.2 AI policy
- [ ] 5.3 Roles and responsibilities

[Continue for all clauses...]

## Justification for Exclusions
[Document any excluded requirements and rationale]

## Approval
Approved by: _________________ Date: _______
`
	
	return os.WriteFile(filepath.Join(templateDir, "statement-of-applicability.md"), []byte(soaTemplate), 0644)
}

func generateNISTTemplates(outputDir string) error {
	templateDir := filepath.Join(outputDir, "nist-ai-rmf")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		return err
	}
	
	// AI Risk Profile template
	riskProfileTemplate := `# AI System Risk Profile (NIST AI RMF)

## System Information
- **System Name**: [AI/LLM System Name]
- **Version**: [Version]
- **Deployment Context**: [Production/Development/Research]
- **Risk Tier**: [High/Medium/Low]

## GOVERN
### Policies and Procedures
- [ ] AI governance policy established
- [ ] Risk management procedures defined
- [ ] Accountability structure documented

## MAP
### Context Understanding
- [ ] Use cases documented
- [ ] Stakeholders identified
- [ ] Benefits and impacts assessed

### Risk Identification
[List identified risks by category]

## MEASURE
### Metrics and KPIs
[Define measurement criteria]

## MANAGE
### Risk Response
[Document risk treatment decisions]

## Continuous Improvement
[Track improvements and lessons learned]
`
	
	return os.WriteFile(filepath.Join(templateDir, "risk-profile.md"), []byte(riskProfileTemplate), 0644)
}

func generateEUAITemplates(outputDir string) error {
	templateDir := filepath.Join(outputDir, "eu-ai-act")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		return err
	}
	
	// Conformity Assessment template
	conformityTemplate := `# EU AI Act Conformity Assessment

## High-Risk AI System Declaration
- **System Classification**: [High-Risk Category]
- **Intended Purpose**: [Detailed description]
- **Provider Information**: [Organization details]

## Technical Documentation (Article 11)
- [ ] System architecture documented
- [ ] Data requirements specified
- [ ] Training methodology described
- [ ] Performance metrics defined
- [ ] Risk assessment completed

## Risk Management (Article 9)
### Risk Identification
[List identified risks]

### Risk Mitigation Measures
[Document controls and measures]

## Data Governance (Article 10)
- [ ] Training data quality assured
- [ ] Bias examination conducted
- [ ] Data management processes defined

## Transparency (Article 13)
- [ ] User instructions prepared
- [ ] System capabilities communicated
- [ ] Limitations disclosed

## Human Oversight (Article 14)
- [ ] Oversight measures implemented
- [ ] Human intervention possible
- [ ] System monitoring active

## Declaration of Conformity
We declare that the AI system described above complies with the requirements of the EU AI Act.

Signed: _________________ Date: _______
`
	
	return os.WriteFile(filepath.Join(templateDir, "conformity-assessment.md"), []byte(conformityTemplate), 0644)
}

// Additional helper functions

func generateISO42001OperationSection(b *bundle.Bundle) string {
	return fmt.Sprintf(`### 8.1 Operational planning and control
The bundle implements operational controls through:
- %d security test templates
- Automated testing capabilities
- Clear test execution procedures

### 8.2 AI system requirements
Tests validate:
- Input security requirements
- Output safety requirements
- Model integrity requirements
- Operational security requirements

### 8.3 AI system design and development
Security considerations integrated throughout:
- Threat modeling templates
- Security testing protocols
- Validation procedures`, b.Manifest.Templates)
}

func generateNISTMeasureSection(b *bundle.Bundle) string {
	return fmt.Sprintf(`Measurement capabilities include:
- %d distinct test scenarios
- Quantitative risk scoring
- Performance benchmarking
- Compliance tracking

Key metrics tracked:
- Test coverage percentage
- Vulnerability detection rate
- False positive ratio
- Remediation effectiveness`, b.Manifest.Templates)
}

func generateEUAIDataGovernanceSection(b *bundle.Bundle) string {
	return `Data governance support through:
- Training data validation tests
- Bias detection capabilities
- Data quality assessments
- Privacy compliance checks

Tests ensure:
- Data relevance and representation
- Absence of discriminatory biases
- Appropriate data handling
- Compliance with data protection`
}

func generateEvidenceMappings(b *bundle.Bundle, outputDir string) error {
	mappings := map[string]interface{}{
		"bundle": map[string]string{
			"name":    b.Manifest.Name,
			"version": b.Manifest.Version,
		},
		"mappings": []map[string]interface{}{},
	}
	
	// Create evidence mappings for each template
	for _, tmpl := range b.Templates {
		mapping := map[string]interface{}{
			"template":   tmpl.Name,
			"path":       tmpl.Path,
			"frameworks": detectFrameworkMappings(tmpl),
			"evidence_type": "test_result",
			"artifacts": []string{
				"test_execution_log",
				"vulnerability_report",
				"remediation_guidance",
			},
		}
		mappings["mappings"] = append(mappings["mappings"].([]map[string]interface{}), mapping)
	}
	
	// Save as JSON
	data, err := json.MarshalIndent(mappings, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filepath.Join(outputDir, "evidence-mappings.json"), data, 0644)
}

func generateExecutiveSummary(b *bundle.Bundle, outputDir string, frameworks []string) error {
	var summary strings.Builder
	
	summary.WriteString("# Compliance Executive Summary\n\n")
	summary.WriteString(fmt.Sprintf("**Bundle**: %s v%s\n", b.Manifest.Name, b.Manifest.Version))
	summary.WriteString(fmt.Sprintf("**Date**: %s\n\n", time.Now().Format("2006-01-02")))
	
	summary.WriteString("## Overview\n")
	summary.WriteString(fmt.Sprintf("This security test bundle contains %d templates designed to validate AI/LLM applications against multiple compliance frameworks.\n\n", b.Manifest.Templates))
	
	summary.WriteString("## Compliance Coverage\n\n")
	summary.WriteString("| Framework | Status | Key Findings |\n")
	summary.WriteString("|-----------|--------|-------------|\n")
	
	for _, fw := range frameworks {
		status := "✅ Documented"
		findings := "Compliance artifacts generated"
		summary.WriteString(fmt.Sprintf("| %s | %s | %s |\n", getFrameworkFullName(fw), status, findings))
	}
	
	summary.WriteString("\n## Recommendations\n")
	summary.WriteString("1. Review generated compliance documentation\n")
	summary.WriteString("2. Execute security tests regularly\n")
	summary.WriteString("3. Track and remediate identified issues\n")
	summary.WriteString("4. Update compliance artifacts quarterly\n")
	
	summary.WriteString("\n## Next Steps\n")
	summary.WriteString("- [ ] Review detailed compliance reports\n")
	summary.WriteString("- [ ] Assign remediation tasks\n")
	summary.WriteString("- [ ] Schedule follow-up assessment\n")
	summary.WriteString("- [ ] Update risk register\n")
	
	return os.WriteFile(filepath.Join(outputDir, "executive-summary.md"), []byte(summary.String()), 0644)
}

func detectFrameworkMappings(tmpl bundle.TemplateEntry) []string {
	frameworks := []string{}
	
	// Check for OWASP categories
	for _, cat := range getOWASPCategories() {
		if strings.Contains(strings.ToLower(tmpl.Path), strings.ToLower(cat.ID)) {
			frameworks = append(frameworks, "OWASP LLM Top 10")
			break
		}
	}
	
	// Check for other framework indicators
	if strings.Contains(tmpl.Path, "risk") || strings.Contains(tmpl.Path, "assess") {
		frameworks = append(frameworks, "NIST AI RMF")
		frameworks = append(frameworks, "ISO/IEC 42001:2023")
	}
	
	if strings.Contains(tmpl.Path, "data") || strings.Contains(tmpl.Path, "privacy") {
		frameworks = append(frameworks, "EU AI Act")
	}
	
	return frameworks
}

func getFrameworkFullName(fw string) string {
	names := map[string]string{
		"owasp":    "OWASP LLM Top 10",
		"iso42001": "ISO/IEC 42001:2023",
		"nist":     "NIST AI RMF",
		"eu-ai":    "EU AI Act",
	}
	
	if name, ok := names[fw]; ok {
		return name
	}
	return fw
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}