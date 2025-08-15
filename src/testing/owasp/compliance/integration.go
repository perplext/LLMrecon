// Package compliance provides compliance mapping and reporting functionality
package compliance

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/reporting/common"
	"github.com/perplext/LLMrecon/src/template/format"
	"github.com/perplext/LLMrecon/src/template/security"
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
)

// ComplianceIntegration handles the integration of compliance mapping with the testing framework
type ComplianceIntegration struct {
	complianceService ComplianceService
	templateVerifier  security.TemplateVerifier

// NewComplianceIntegration creates a new compliance integration
func NewComplianceIntegration(complianceService ComplianceService) *ComplianceIntegration {
	return &ComplianceIntegration{
		complianceService: complianceService,
		templateVerifier:  security.NewTemplateVerifier(),
	}

// RegisterWithTestFactory registers the compliance service with a test factory
func RegisterWithTestFactory(factory types.TestFactory, complianceService ComplianceService) error {
	if factory == nil {
		return fmt.Errorf("test factory cannot be nil")
	}

	if complianceService == nil {
		complianceService = NewComplianceService()
	}

	return factory.RegisterComplianceService(complianceService)

// GetComplianceService gets the compliance service from a test factory
func GetComplianceService(factory types.TestFactory) (ComplianceService, error) {
	if factory == nil {
		return nil, fmt.Errorf("test factory cannot be nil")
	}

	service, err := factory.GetComplianceService()
	if err != nil {
		return nil, err
	}

	complianceService, ok := service.(ComplianceService)
	if !ok {
		return nil, fmt.Errorf("service is not a ComplianceService")
	}

	return complianceService, nil

// VerifyTemplateAndGenerateReport verifies a template and generates a compliance report
func (ci *ComplianceIntegration) VerifyTemplateAndGenerateReport(
	ctx context.Context,
	templatePath string,
	testSuite *types.TestSuite,
	options *security.VerificationOptions,
	reportOptions *ComplianceReportOptions,
) (*security.VerificationResult, *ComplianceReport, error) {
	// Verify template security
	verificationResult, err := ci.templateVerifier.VerifyTemplateFile(ctx, templatePath, options)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to verify template security: %w", err)
	}

	// Generate compliance report
	if reportOptions == nil {
		reportOptions = &ComplianceReportOptions{
			Title:     "Compliance Report for " + verificationResult.TemplateName,
			Standards: []ComplianceStandard{OWASPLLMTop10, ISO42001},
		}
	}

	complianceReport, err := ci.complianceService.GenerateReport(ctx, testSuite, reportOptions)
	if err != nil {
		return verificationResult, nil, fmt.Errorf("failed to generate compliance report: %w", err)
	}

	return verificationResult, complianceReport, nil

// ConvertToTestResults converts verification and compliance results to test results
// VerifyTemplate verifies a template for security issues
func (ci *ComplianceIntegration) VerifyTemplate(ctx context.Context, template *format.Template, options *security.VerificationOptions) (*security.VerificationResult, error) {
	return ci.templateVerifier.VerifyTemplate(ctx, template, options)

// VerifyTemplateFile verifies a template file for security issues
func (ci *ComplianceIntegration) VerifyTemplateFile(ctx context.Context, templatePath string, options *security.VerificationOptions) (*security.VerificationResult, error) {
	return ci.templateVerifier.VerifyTemplateFile(ctx, templatePath, options)

// VerifyTemplateDirectory verifies all templates in a directory for security issues
func (ci *ComplianceIntegration) VerifyTemplateDirectory(ctx context.Context, directoryPath string, options *security.VerificationOptions) ([]*security.VerificationResult, error) {
	return ci.templateVerifier.VerifyTemplateDirectory(ctx, directoryPath, options)

// RegisterCheck registers a custom security check
func (ci *ComplianceIntegration) RegisterCheck(name string, check security.SecurityCheck) {
	ci.templateVerifier.RegisterCheck(name, check)

// GetChecks returns all registered security checks
func (ci *ComplianceIntegration) GetChecks() map[string]security.SecurityCheck {
	return ci.templateVerifier.GetChecks()

func (ci *ComplianceIntegration) ConvertToTestResults(
	verificationResult *security.VerificationResult,
	complianceReport *ComplianceReport,
) []*common.TestResult {
	var testResults []*common.TestResult

	// Add security verification test result
	if verificationResult != nil {
		securityTestResult := security.VerificationResultToTestResult(verificationResult)
		testResults = append(testResults, securityTestResult)
	}

	// Add compliance report test results
	if complianceReport != nil {
		// Add a summary test result
		summaryResult := &common.TestResult{
			ID:          "compliance_summary",
			Name:        "Compliance Report Summary",
			Description: "Summary of compliance verification results",
			Severity:    common.Medium,
			Category:    "compliance",
			Status:      "info",
			Details:     fmt.Sprintf("Test Suite: %s, Standards: %v", complianceReport.TestSuite.Name, getStandardNames(complianceReport.Standards)),
			RawData:     complianceReport,
			Metadata: map[string]interface{}{
				"test_suite_name": complianceReport.TestSuite.Name,
				"standards":       getStandardNames(complianceReport.Standards),
			},
		}

		testResults = append(testResults, summaryResult)

		// Add test results for each standard
		for _, standardResult := range complianceReport.StandardResults {
			status := "passed"
			if standardResult.CompliancePercentage < 80.0 {
				status = "failed"
			}

			testResult := &common.TestResult{
				ID:          fmt.Sprintf("compliance_%s", string(standardResult.Standard)),
				Name:        fmt.Sprintf("Compliance: %s", string(standardResult.Standard)),
				Description: fmt.Sprintf("Compliance verification for %s", string(standardResult.Standard)),
				Severity:    common.Medium,
				Category:    "compliance",
				Status:      status,
				Details:     fmt.Sprintf("Compliance: %.2f%%, Requirements Met: %d/%d",
					standardResult.CompliancePercentage,
					standardResult.RequirementsMet,
					standardResult.TotalRequirements),
				RawData: standardResult,
				Metadata: map[string]interface{}{
					"standard_id":           string(standardResult.Standard),
					"standard_name":         string(standardResult.Standard),
					"compliance_percentage": standardResult.CompliancePercentage,
					"requirements_met":      standardResult.RequirementsMet,
					"total_requirements":    standardResult.TotalRequirements,
				},
			}

			testResults = append(testResults, testResult)
		}
	}

	// Add combined test result
	if verificationResult != nil && complianceReport != nil {
		// Determine overall status
		overallStatus := "passed"
		if !verificationResult.Passed {
			overallStatus = "failed"
		} else {
			// Check if any standard failed
			for _, standardResult := range complianceReport.StandardResults {
				if standardResult.CompliancePercentage < 80.0 {
					overallStatus = "failed"
					break
				}
			}
		}

		overallResult := &common.TestResult{
			ID:          fmt.Sprintf("template_compliance_%s", verificationResult.TemplateID),
			Name:        fmt.Sprintf("Template Compliance: %s", verificationResult.TemplateName),
			Description: "Overall template security and compliance verification",
			Severity:    common.High,
			Category:    "template_compliance",
			Status:      overallStatus,
			Details:     fmt.Sprintf("Template: %s, Security: %t, Compliance: %s",
				verificationResult.TemplateName,
				verificationResult.Passed,
				overallStatus),
			Metadata: map[string]interface{}{
				"template_id":     verificationResult.TemplateID,
				"template_name":   verificationResult.TemplateName,
				"security_passed": verificationResult.Passed,
				"security_score":  verificationResult.Score,
			},
		}

		testResults = append(testResults, overallResult)
	}

	return testResults

// getStandardNames returns the names of the standards
func getStandardNames(standards []ComplianceStandard) []string {
	var names []string
	for _, standard := range standards {
		names = append(names, string(standard))
	}
