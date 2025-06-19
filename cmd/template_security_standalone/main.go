package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/template/security"
	"github.com/perplext/LLMrecon/src/testing/owasp/compliance"
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
)

var (
	templatePath      string
	templateDir       string
	outputDir         string
	severityThreshold string
)

func main() {
	// Parse command line flags
	flag.StringVar(&templatePath, "template", "", "Path to template file")
	flag.StringVar(&templateDir, "dir", "", "Path to template directory")
	flag.StringVar(&outputDir, "output", "", "Output directory for reports")
	flag.StringVar(&severityThreshold, "severity", "medium", "Severity threshold (critical, high, medium, low, info)")
	flag.Parse()

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
	complianceService := compliance.NewComplianceService()

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
}

// verifyTemplate verifies a single template
func verifyTemplate(ctx context.Context, verifier security.TemplateVerifier, options *security.VerificationOptions) error {
	fmt.Printf("Verifying template: %s\n", templatePath)

	// Verify template
	result, err := verifier.VerifyTemplateFile(ctx, templatePath, options)
	if err != nil {
		return fmt.Errorf("failed to verify template: %w", err)
	}

	// Print verification result
	printVerificationResult(result)

	// Skip reporting integration for now - would need to pass complianceService
	// reportingIntegration := compliance.NewReportingIntegration(complianceService, verifier)

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
	templateComplianceResult, err := reportingIntegration.VerifyTemplateSecurityAndCompliance(ctx, templatePath, testSuite, options)
	if err != nil {
		return fmt.Errorf("failed to verify template compliance: %w", err)
	}

	// Print compliance results
	printComplianceResult(templateComplianceResult)

	// Save results to JSON files if output directory is specified
	if outputDir != "" {
		// Save verification result
		verificationResultPath := filepath.Join(outputDir, "verification_result.json")
		verificationResultJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal verification result: %w", err)
		}
		if err := os.WriteFile(verificationResultPath, verificationResultJSON, 0644); err != nil {
			return fmt.Errorf("failed to save verification result: %w", err)
		}
		fmt.Printf("Verification result saved to %s\n", verificationResultPath)

		// Save compliance result
		complianceResultPath := filepath.Join(outputDir, fmt.Sprintf("compliance_result_%s.json", 
			strings.ReplaceAll(templateComplianceResult.TemplateID, "/", "_")))
		complianceResultJSON, err := json.MarshalIndent(templateComplianceResult, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal compliance result: %w", err)
		}
		if err := os.WriteFile(complianceResultPath, complianceResultJSON, 0644); err != nil {
			return fmt.Errorf("failed to save compliance result: %w", err)
		}
		fmt.Printf("Compliance result saved to %s\n", complianceResultPath)
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

	// Combine the files
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

// calculateSummary calculates a summary of template verification results
func calculateSummary(results []*security.VerificationResult) *security.VerificationSummary {
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
