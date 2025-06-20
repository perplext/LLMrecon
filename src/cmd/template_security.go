//go:build ignore

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/reporting/common"
	"github.com/perplext/LLMrecon/src/template/security"
	"github.com/perplext/LLMrecon/src/testing/owasp/compliance"
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
	"github.com/spf13/cobra"
)

var (
	templateSecurityPath      string
	templateSecurityDir       string
	templateSecurityOutputDir string
	templateReportFormats     []string
	templatePipelineConfig    string
	templateSeverityThreshold string
)

// templateSecurityCmd represents the template-security command
var templateSecurityCmd = &cobra.Command{
	Use:   "template-security",
	Short: "Verify template security",
	Long:  `Verify template security using the template security verifier.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		// Create verification options
		options := security.DefaultVerificationOptions()
		options.StrictMode = true

		// Set severity threshold
		switch severityThreshold {
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
		if templateSecurityOutputDir != "" {
			if err := os.MkdirAll(templateSecurityOutputDir, 0755); err != nil {
				fmt.Printf("Error creating output directory: %v\n", err)
				os.Exit(1)
			}
		}

		// Run pipeline if pipeline config is specified
		if templatePipelineConfig != "" {
			if err := runSecurityPipeline(ctx, integration, options); err != nil {
				fmt.Printf("Error running security pipeline: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// Verify template or directory
		if templateSecurityPath != "" {
			if err := verifyTemplate(ctx, integration, options); err != nil {
				fmt.Printf("Error verifying template: %v\n", err)
				os.Exit(1)
			}
		} else if templateSecurityDir != "" {
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
	rootCmd.AddCommand(templateSecurityCmd)

	// Add flags
	templateSecurityCmd.Flags().StringVar(&templateSecurityPath, "template", "", "Path to template file")
	templateSecurityCmd.Flags().StringVar(&templateSecurityDir, "dir", "", "Path to template directory")
	templateSecurityCmd.Flags().StringVar(&templateSecurityOutputDir, "output", "", "Output directory for reports")
	templateSecurityCmd.Flags().StringSliceVar(&templateReportFormats, "format", []string{"JSON", "HTML"}, "Report formats (JSON, HTML, CSV, XML)")
	templateSecurityCmd.Flags().StringVar(&templatePipelineConfig, "pipeline-config", "", "Path to pipeline configuration file")
	templateSecurityCmd.Flags().StringVar(&templateSeverityThreshold, "severity", "medium", "Severity threshold (critical, high, medium, low, info)")
}

// verifyTemplate verifies a single template
func verifyTemplate(ctx context.Context, verifier security.TemplateVerifier, options *security.VerificationOptions) error {
	fmt.Printf("Verifying template: %s\n", templateSecurityPath)

	// Verify template
	result, err := verifier.VerifyTemplateFile(ctx, templateSecurityPath, options)
	if err != nil {
		return fmt.Errorf("failed to verify template: %w", err)
	}

	// Print verification result
	printVerificationResult(result)

	// Generate reports if output directory is specified
	if templateSecurityOutputDir != "" {
		if err := generateReports([]*security.VerificationResult{result}); err != nil {
			return fmt.Errorf("failed to generate reports: %w", err)
		}
	}

	// Create compliance service directly
	complianceService := compliance.NewComplianceServiceImpl(
		compliance.NewBaseComplianceMapper(),
		compliance.NewComplianceReporterImpl(),
	)

	// Create reporting integration
	integration := compliance.NewReportingIntegration(complianceService, verifier)

	// Create a test suite directly
	testSuite := &types.TestSuite{
		ID:          "template-security-test-suite",
		Name:        "Template Security Test Suite",
		Description: "Test suite for template security verification",
		CreatedAt:   time.Now(),
		Tags:        []string{"security", "template"},
		Metadata:    make(map[string]interface{}),
	}

	// Verify template compliance
	fmt.Println("\nVerifying template compliance...")
	templateComplianceResult, err := integration.VerifyTemplateSecurityAndCompliance(ctx, templateSecurityPath, testSuite, options)
	if err != nil {
		return fmt.Errorf("failed to verify template compliance: %w", err)
	}

	// Print compliance results
	printComplianceResult(templateComplianceResult)

	// Generate compliance reports if output directory is specified
	if templateSecurityOutputDir != "" {
		if err := generateComplianceReports(integration, templateComplianceResult); err != nil {
			return fmt.Errorf("failed to generate compliance reports: %w", err)
		}
	}

	return nil
}

// verifyTemplateDirectory verifies all templates in a directory
func verifyTemplateDirectory(ctx context.Context, integration *compliance.ComplianceIntegration, options *security.VerificationOptions) error {
	fmt.Printf("Verifying templates in directory: %s\n", templateSecurityDir)

	// Find all template files in the directory
	templateFiles, err := filepath.Glob(filepath.Join(templateSecurityDir, "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to find template files: %w", err)
	}

	// Also check for .yml files
	ymlFiles, err := filepath.Glob(filepath.Join(templateSecurityDir, "*.yml"))
	if err != nil {
		return fmt.Errorf("failed to find template files: %w", err)
	}

	// Combine the files
	templateFiles = append(templateFiles, ymlFiles...)

	if len(templateFiles) == 0 {
		fmt.Println("No template files found in directory")
		return nil
	}

	// Verify each template file
	for _, templateFile := range templateFiles {
		templateSecurityPath = templateFile
		if err := verifyTemplate(ctx, integration, options); err != nil {
			fmt.Printf("Error verifying template %s: %v\n", templateFile, err)
			continue
		}
		fmt.Println()
	}

	return nil
}

// runSecurityPipeline runs the template security pipeline
func runSecurityPipeline(ctx context.Context, verifier security.TemplateVerifier, options *security.VerificationOptions) error {
	// Load pipeline configuration
	var config security.PipelineConfig
	if templatePipelineConfig != "" {
		data, err := os.ReadFile(templatePipelineConfig)
		if err != nil {
			return fmt.Errorf("failed to read pipeline configuration: %w", err)
		}

		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse pipeline configuration: %w", err)
		}
	} else {
		// Use default configuration
		config = security.PipelineConfig{
			TemplateDirectories: []string{templateSecurityDir},
			OutputDirectory:     templateSecurityOutputDir,
			VerificationOptions: options,
			ReportFormats:       convertReportFormats(templateReportFormats),
		}
	}

	// Create pipeline
	pipeline := security.NewPipeline(verifier, options)

	// Run pipeline
	fmt.Println("Running template security pipeline...")
	if err := pipeline.RunVerification(ctx, &config); err != nil {
		return fmt.Errorf("failed to run pipeline: %w", err)
	}

	// Print results
	results := pipeline.GetResults()
	for _, result := range results {
		printVerificationResult(result)
		fmt.Println()
	}

	// Print summary
	summary := pipeline.GetSummary()
	printSummary(summary)

	fmt.Printf("\nPipeline completed successfully. Reports saved to %s\n", config.OutputDirectory)
	return nil
}

// printVerificationResult prints a verification result
func printVerificationResult(result *security.VerificationResult) {
	fmt.Printf("Template: %s (%s)\n", result.TemplateName, result.TemplateID)
	fmt.Printf("Path: %s\n", result.TemplatePath)
	fmt.Printf("Passed: %t\n", result.Passed)
	fmt.Printf("Score: %.2f/%.2f\n", result.Score, result.MaxScore)

	if len(result.Issues) > 0 {
		fmt.Printf("Issues (%d):\n", len(result.Issues))
		for i, issue := range result.Issues {
			fmt.Printf("  %d. [%s] %s\n", i+1, issue.Severity, issue.Description)
			fmt.Printf("     Location: %s\n", issue.Location)
			fmt.Printf("     Remediation: %s\n", issue.Remediation)
			if issue.Context != "" {
				fmt.Printf("     Context: %s\n", issue.Context)
			}
		}
	} else {
		fmt.Println("No issues found")
	}
}

// printComplianceResult prints a compliance result
func printComplianceResult(result *compliance.TemplateComplianceResult) {
	fmt.Printf("Template: %s (%s)\n", result.TemplateName, result.TemplateID)
	fmt.Printf("Path: %s\n", result.TemplatePath)
	fmt.Printf("Overall Compliance: %t\n", result.OverallCompliance)

	fmt.Println("Compliance by Standard:")
	for standard, compliant := range result.ComplianceByStandard {
		fmt.Printf("  %s: %t\n", standard, compliant)
	}

	fmt.Println("Security Result:")
	fmt.Printf("  Passed: %t\n", result.SecurityResult.Passed)
	fmt.Printf("  Score: %.2f/%.2f\n", result.SecurityResult.Score, result.SecurityResult.MaxScore)
	fmt.Printf("  Issues: %d\n", len(result.SecurityResult.Issues))
}

// printSummary prints a verification summary
func printSummary(summary *security.VerificationSummary) {
	fmt.Println("\nVerification Summary:")
	fmt.Printf("Total Templates: %d\n", summary.TotalTemplates)
	fmt.Printf("Passed Templates: %d\n", summary.PassedTemplates)
	fmt.Printf("Failed Templates: %d\n", summary.FailedTemplates)
	fmt.Printf("Total Issues: %d\n", summary.TotalIssues)
	fmt.Printf("Average Score: %.2f\n", summary.AverageScore)
	fmt.Printf("Compliance Percentage: %.2f%%\n", summary.CompliancePercentage)

	fmt.Println("\nIssues by Severity:")
	for severity, count := range summary.IssuesBySeverity {
		fmt.Printf("  %s: %d\n", severity, count)
	}

	fmt.Println("\nIssues by Type:")
	for issueType, count := range summary.IssuesByType {
		fmt.Printf("  %s: %d\n", issueType, count)
	}

	fmt.Println("\nCompliance Status:")
	for standard, compliant := range summary.ComplianceStatus {
		fmt.Printf("  %s: %t\n", standard, compliant)
	}
}

// calculateTemplateSummary calculates a summary of template verification results
func calculateTemplateSummary(results []*security.VerificationResult) *security.VerificationSummary {
	summary := &security.VerificationSummary{
		TotalTemplates:   len(results),
		PassedTemplates:  0,
		FailedTemplates:  0,
		TotalIssues:      0,
		IssuesBySeverity: make(map[string]int),
		IssuesByType:     make(map[string]int),
	}

	// Calculate statistics
	for _, result := range results {
		if result.Passed {
			summary.PassedTemplates++
		} else {
			summary.FailedTemplates++
		}

		// Count issues
		for _, issue := range result.Issues {
			summary.TotalIssues++

			// Count by severity
			summary.IssuesBySeverity[string(issue.Severity)]++

			// Count by type
			summary.IssuesByType[string(issue.Type)]++
		}
	}

	return summary
}

// generateReports generates reports for verification results
func generateReports(results []*security.VerificationResult) error {
	// Create a new report with the results
	report := &security.TemplateSecurityReport{
		Title:           "Template Security Verification Report",
		GeneratedAt:     time.Now(),
		TemplateResults: results,
		Options:         security.DefaultVerificationOptions(),
		// We'll calculate the summary ourselves instead of using a non-existent method
		Summary: calculateSummary(results),
	}

	// Convert to test results
	testResults := security.ConvertToTestResults(report)

	// Generate reports in the specified formats
	for _, formatStr := range templateReportFormats {
		format := common.ReportFormat(formatStr)
		outputPath := filepath.Join(templateSecurityOutputDir, fmt.Sprintf("template_security_report.%s", strings.ToLower(string(format))))

		// Get formatter for the specified format
		formatterCreator, ok := common.GetFormatterCreatorFromDefault(format)
		if !ok {
			return fmt.Errorf("formatter not found for format: %s", format)
		}

		formatter, err := formatterCreator(nil)
		if err != nil {
			return fmt.Errorf("failed to create formatter for format %s: %w", format, err)
		}

		// Format the report
		formattedReport, err := formatter.FormatTestResults(testResults)
		if err != nil {
			return fmt.Errorf("failed to format report: %w", err)
		}

		// Save the report to file
		if err := os.WriteFile(outputPath, []byte(formattedReport), 0644); err != nil {
			return fmt.Errorf("failed to save report to file: %w", err)
		}

		fmt.Printf("Report saved to %s\n", outputPath)
	}

	return nil
}

// generateComplianceReports generates compliance reports
func generateComplianceReports(integration *compliance.ReportingIntegration, result *compliance.TemplateComplianceResult) error {
	// Convert to test results
	testResults := integration.ConvertTemplateComplianceToTestResults(result)

	// Generate reports in the specified formats
	for _, formatStr := range templateReportFormats {
		format := common.ReportFormat(formatStr)
		outputPath := filepath.Join(templateSecurityOutputDir, fmt.Sprintf("template_compliance_report_%s.%s",
			strings.ReplaceAll(result.TemplateID, "/", "_"),
			strings.ToLower(string(format))))

		// Get formatter for the specified format
		formatterCreator, ok := common.GetFormatterCreatorFromDefault(format)
		if !ok {
			return fmt.Errorf("formatter not found for format: %s", format)
		}

		formatter, err := formatterCreator(nil)
		if err != nil {
			return fmt.Errorf("failed to create formatter for format %s: %w", format, err)
		}

		// Format the report
		formattedReport, err := formatter.FormatTestResults(testResults)
		if err != nil {
			return fmt.Errorf("failed to format report: %w", err)
		}

		// Save the report to file
		if err := os.WriteFile(outputPath, []byte(formattedReport), 0644); err != nil {
			return fmt.Errorf("failed to save report to file: %w", err)
		}

		fmt.Printf("Compliance report saved to %s\n", outputPath)
	}

	return nil
}

// convertReportFormats converts string report formats to common.ReportFormat
func convertReportFormats(formats []string) []common.ReportFormat {
	var result []common.ReportFormat
	for _, format := range formats {
		result = append(result, common.ReportFormat(format))
	}
	return result
}
