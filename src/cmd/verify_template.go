package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/security"
	"github.com/perplext/LLMrecon/src/testing/owasp/compliance"
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
	"github.com/spf13/cobra"
)

var (
	templatePath       string
	templateDir        string
	outputDir          string
	outputFormat       string
	strictMode         bool
	includeInfo        bool
	severityThreshold  string
	complianceStandard string
)

// verifyTemplateCmd represents the verify-template command
var verifyTemplateCmd = &cobra.Command{
	Use:   "verify-template",
	Short: "Verify template security and compliance",
	Long: `Verify template security and compliance with OWASP LLM Top 10 and ISO/IEC 42001.
This command allows you to verify the security of templates and generate compliance reports.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Create verification options
		options := security.DefaultVerificationOptions()
		options.StrictMode = strictMode
		options.IncludeInfo = includeInfo

		// Set severity threshold
		switch strings.ToLower(severityThreshold) {
		case "critical":
			options.SeverityThreshold = "critical"
		case "high":
			options.SeverityThreshold = "high"
		case "medium":
			options.SeverityThreshold = "medium"
		case "low":
			options.SeverityThreshold = "low"
		case "info":
			options.SeverityThreshold = "info"
		}

		// Create compliance service
		complianceService := compliance.NewComplianceServiceImpl(
			compliance.NewBaseComplianceMapper(),
			compliance.NewComplianceReporterImpl(),
		)

		// Create compliance integration
		integration := compliance.NewComplianceIntegration(complianceService)

		// Create output directory if it doesn't exist
		if outputDir != "" {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				fmt.Printf("Error creating output directory: %v\n", err)
				os.Exit(1)
			}
		}

		// Verify template or directory
		if templatePath != "" {
			if err := verifyTemplate(ctx, integration, options); err != nil {
				fmt.Printf("Error verifying template: %v\n", err)
				os.Exit(1)
			}
		} else if templateDir != "" {
			if err := verifyTemplateDirectory(ctx, integration, options); err != nil {
				fmt.Printf("Error verifying template directory: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Please specify a template file or directory")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(verifyTemplateCmd)

	// Add flags
	verifyTemplateCmd.Flags().StringVar(&templatePath, "template", "", "Path to template file")
	verifyTemplateCmd.Flags().StringVar(&templateDir, "dir", "", "Path to template directory")
	verifyTemplateCmd.Flags().StringVar(&outputDir, "output", "", "Output directory for reports")
	verifyTemplateCmd.Flags().StringVar(&outputFormat, "format", "json", "Output format (json, text)")
	verifyTemplateCmd.Flags().BoolVar(&strictMode, "strict", false, "Enable strict mode for verification")
	verifyTemplateCmd.Flags().BoolVar(&includeInfo, "include-info", true, "Include info level issues in reports")
	verifyTemplateCmd.Flags().StringVar(&severityThreshold, "severity", "low", "Severity threshold (critical, high, medium, low, info)")
	verifyTemplateCmd.Flags().StringVar(&complianceStandard, "standard", "owasp", "Compliance standard (owasp, iso, all)")
}

// verifyTemplate verifies a single template
func verifyTemplate(ctx context.Context, integration *compliance.ComplianceIntegration, options *security.VerificationOptions) error {
	fmt.Printf("Verifying template: %s\n", templatePath)

	// Load template
	template, err := format.LoadFromFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to load template: %w", err)
	}

	// Create a test suite for the template
	testSuite := &types.TestSuite{
		ID:          "TS001",
		Name:        "Template Verification Suite",
		Description: "Test suite for template verification",
		TestCases: []*types.TestCase{
			{
				ID:          "TC001",
				Name:        template.Info.Name,
				Description: template.Info.Description,
				Category:    "template_verification",
			},
		},
	}

	// Create compliance report options
	var standards []compliance.ComplianceStandard
	switch strings.ToLower(complianceStandard) {
	case "owasp":
		standards = []compliance.ComplianceStandard{compliance.OWASPLLMTop10}
	case "iso":
		standards = []compliance.ComplianceStandard{compliance.ISO42001}
	default:
		standards = []compliance.ComplianceStandard{compliance.OWASPLLMTop10, compliance.ISO42001}
	}

	reportOptions := &compliance.ComplianceReportOptions{
		Title:     "Compliance Report for " + template.Info.Name,
		Standards: standards,
	}

	// Verify template and generate report
	verificationResult, complianceReport, err := integration.VerifyTemplateAndGenerateReport(
		ctx,
		templatePath,
		testSuite,
		options,
		reportOptions,
	)
	if err != nil {
		return fmt.Errorf("failed to verify template and generate report: %w", err)
	}

	// Print verification result
	printVerificationResult(verificationResult)

	// Print compliance report
	printComplianceReport(complianceReport)

	// Generate report file if output directory is specified
	if outputDir != "" {
		if err := generateReportFile(verificationResult, complianceReport); err != nil {
			return fmt.Errorf("failed to generate report file: %w", err)
		}
	}

	return nil
}

// verifyTemplateDirectory verifies all templates in a directory
func verifyTemplateDirectory(ctx context.Context, integration *compliance.ComplianceIntegration, options *security.VerificationOptions) error {
	fmt.Printf("Verifying templates in directory: %s\n", templateDir)

	// Find all template files in the directory
	templateFiles, err := filepath.Glob(filepath.Join(templateDir, "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to find template files: %w", err)
	}

	// Also check for .yml files
	ymlFiles, err := filepath.Glob(filepath.Join(templateDir, "*.yml"))
	if err != nil {
		return fmt.Errorf("failed to find template files: %w", err)
	}

	templateFiles = append(templateFiles, ymlFiles...)

	if len(templateFiles) == 0 {
		fmt.Println("No template files found in directory")
		return nil
	}

	// Verify each template file
	for _, templateFile := range templateFiles {
		templatePath = templateFile
		if err := verifyTemplate(ctx, integration, options); err != nil {
			fmt.Printf("Error verifying template %s: %v\n", templateFile, err)
			continue
		}
		fmt.Println()
	}

	return nil
}

// printVerificationResult prints a verification result
func printVerificationResult(result *security.VerificationResult) {
	fmt.Println("\n=== Template Security Verification ===")
	fmt.Printf("Template: %s (%s)\n", result.TemplateName, result.TemplateID)
	fmt.Printf("Path: %s\n", result.TemplatePath)
	fmt.Printf("Passed: %t\n", result.Passed)
	fmt.Printf("Score: %.2f/%.2f\n", result.Score, result.MaxScore)

	if len(result.Issues) > 0 {
		fmt.Printf("\nIssues (%d):\n", len(result.Issues))
		for i, issue := range result.Issues {
			fmt.Printf("  %d. [%s] %s\n", i+1, issue.Severity, issue.Description)
			fmt.Printf("     Location: %s\n", issue.Location)
			fmt.Printf("     Remediation: %s\n", issue.Remediation)
			if issue.Context != "" {
				fmt.Printf("     Context: %s\n", issue.Context)
			}
			fmt.Println()
		}
	} else {
		fmt.Println("\nNo security issues found")
	}
}

// printComplianceReport prints a compliance report
func printComplianceReport(report *compliance.ComplianceReport) {
	fmt.Println("\n=== Compliance Report ===")
	fmt.Printf("Title: %s\n", report.Title)
	fmt.Printf("Test Suite: %s\n", report.TestSuite.Name)
	fmt.Printf("Standards: %s\n", getStandardNames(report.Standards))

	fmt.Println("\nStandard Results:")
	for _, standardResult := range report.StandardResults {
		fmt.Printf("\n  Standard: %s (%s)\n", standardResult.Standard.Name, standardResult.Standard.ID)
		fmt.Printf("  Compliance: %.2f%%\n", standardResult.CompliancePercentage)
		fmt.Printf("  Requirements Met: %d/%d\n", standardResult.RequirementsMet, standardResult.TotalRequirements)

		fmt.Println("\n  Requirement Results:")
		for _, reqResult := range standardResult.RequirementResults {
			fmt.Printf("    - %s (%s): %t\n", reqResult.Requirement.Name, reqResult.Requirement.ID, reqResult.Compliant)
			if !reqResult.Compliant && reqResult.Reason != "" {
				fmt.Printf("      Reason: %s\n", reqResult.Reason)
			}
		}
	}
}

// generateReportFile generates a report file in the specified format
func generateReportFile(verificationResult *security.VerificationResult, complianceReport *compliance.ComplianceReport) error {
	// Create combined report
	combinedReport := struct {
		SecurityVerification *security.VerificationResult  `json:"security_verification"`
		ComplianceReport     *compliance.ComplianceReport `json:"compliance_report"`
	}{
		SecurityVerification: verificationResult,
		ComplianceReport:     complianceReport,
	}

	// Generate file name
	fileName := fmt.Sprintf("template_verification_%s.%s", verificationResult.TemplateID, outputFormat)
	filePath := filepath.Join(outputDir, fileName)

	// Generate report based on format
	switch strings.ToLower(outputFormat) {
	case "json":
		data, err := json.MarshalIndent(combinedReport, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal report to JSON: %w", err)
		}
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return fmt.Errorf("failed to write report to file: %w", err)
		}
	case "text":
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create report file: %w", err)
		}
		defer file.Close()

		// Write security verification
		fmt.Fprintf(file, "=== Template Security Verification ===\n")
		fmt.Fprintf(file, "Template: %s (%s)\n", verificationResult.TemplateName, verificationResult.TemplateID)
		fmt.Fprintf(file, "Path: %s\n", verificationResult.TemplatePath)
		fmt.Fprintf(file, "Passed: %t\n", verificationResult.Passed)
		fmt.Fprintf(file, "Score: %.2f/%.2f\n", verificationResult.Score, verificationResult.MaxScore)

		if len(verificationResult.Issues) > 0 {
			fmt.Fprintf(file, "\nIssues (%d):\n", len(verificationResult.Issues))
			for i, issue := range verificationResult.Issues {
				fmt.Fprintf(file, "  %d. [%s] %s\n", i+1, issue.Severity, issue.Description)
				fmt.Fprintf(file, "     Location: %s\n", issue.Location)
				fmt.Fprintf(file, "     Remediation: %s\n", issue.Remediation)
				if issue.Context != "" {
					fmt.Fprintf(file, "     Context: %s\n", issue.Context)
				}
				fmt.Fprintln(file)
			}
		} else {
			fmt.Fprintln(file, "\nNo security issues found")
		}

		// Write compliance report
		fmt.Fprintf(file, "\n=== Compliance Report ===\n")
		fmt.Fprintf(file, "Title: %s\n", complianceReport.Title)
		fmt.Fprintf(file, "Test Suite: %s\n", complianceReport.TestSuite.Name)
		fmt.Fprintf(file, "Standards: %s\n", getStandardNames(complianceReport.Standards))

		fmt.Fprintf(file, "\nStandard Results:\n")
		for _, standardResult := range complianceReport.StandardResults {
			fmt.Fprintf(file, "\n  Standard: %s (%s)\n", standardResult.Standard.Name, standardResult.Standard.ID)
			fmt.Fprintf(file, "  Compliance: %.2f%%\n", standardResult.CompliancePercentage)
			fmt.Fprintf(file, "  Requirements Met: %d/%d\n", standardResult.RequirementsMet, standardResult.TotalRequirements)

			fmt.Fprintf(file, "\n  Requirement Results:\n")
			for _, reqResult := range standardResult.RequirementResults {
				fmt.Fprintf(file, "    - %s (%s): %t\n", reqResult.Requirement.Name, reqResult.Requirement.ID, reqResult.Compliant)
				if !reqResult.Compliant && reqResult.Reason != "" {
					fmt.Fprintf(file, "      Reason: %s\n", reqResult.Reason)
				}
			}
		}
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}

	fmt.Printf("\nReport saved to %s\n", filePath)
	return nil
}

// getStandardNames returns a string representation of standard names
func getStandardNames(standards []compliance.ComplianceStandard) string {
	var names []string
	for _, standard := range standards {
		names = append(names, standard.Name())
	}
	return strings.Join(names, ", ")
}
