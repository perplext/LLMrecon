// Package compliance provides compliance mapping and reporting functionality
package compliance

import (
	"context"
	"fmt"

	"github.com/perplext/LLMrecon/src/reporting/common"
	"github.com/perplext/LLMrecon/src/template/security"
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
)

// ReportingIntegration handles the integration between compliance mapping and reporting
type ReportingIntegration struct {
	complianceService ComplianceService
	templateVerifier  security.TemplateVerifier
}

// NewReportingIntegration creates a new reporting integration
func NewReportingIntegration(complianceService ComplianceService, templateVerifier security.TemplateVerifier) *ReportingIntegration {
	if templateVerifier == nil {
		templateVerifier = security.NewTemplateVerifier()
	}

	return &ReportingIntegration{
		complianceService: complianceService,
		templateVerifier:  templateVerifier,
	}
}

// GenerateComplianceReport generates a compliance report for a test suite
func (ri *ReportingIntegration) GenerateComplianceReport(ctx context.Context, testSuite *types.TestSuite, options *ComplianceReportOptions) (*ComplianceReport, error) {
	if testSuite == nil {
		return nil, fmt.Errorf("test suite cannot be nil")
	}

	if options == nil {
		options = &ComplianceReportOptions{
			Title:     "Compliance Report",
			Standards: []ComplianceStandard{OWASPLLMTop10, ISO42001},
		}
	}

	// Generate compliance report using the compliance service
	report, err := ri.complianceService.GenerateReport(ctx, testSuite, options)
	if err != nil {
		return nil, fmt.Errorf("failed to generate compliance report: %w", err)
	}

	return report, nil
}

// ConvertToTestResults converts a compliance report to test results
func (ri *ReportingIntegration) ConvertToTestResults(report *ComplianceReport) []*common.TestResult {
	var testResults []*common.TestResult

	// Add a summary test result
	summaryResult := &common.TestResult{
		ID:          "compliance_summary",
		Name:        "Compliance Report Summary",
		Description: "Summary of compliance verification results",
		Severity:    common.Medium,
		Category:    "compliance",
		Status:      "info",
		Details:     fmt.Sprintf("Test Suite: %s, Standards: %v", report.TestSuite.Name, getComplianceStandardNames(report.Standards)),
		RawData:     report,
		Metadata: map[string]interface{}{
			"test_suite_name": report.TestSuite.Name,
			"standards":       getComplianceStandardNames(report.Standards),
		},
	}

	testResults = append(testResults, summaryResult)

	// Add test results for each standard
	for _, standardResult := range report.StandardResults {
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

		// Add test results for each requirement
		for _, reqResult := range standardResult.RequirementResults {
			reqStatus := "passed"
			if !reqResult.Compliant {
				reqStatus = "failed"
			}

			reqTestResult := &common.TestResult{
				ID:          fmt.Sprintf("compliance_%s_%s", string(standardResult.Standard), reqResult.Requirement.ID),
				Name:        fmt.Sprintf("Requirement: %s", reqResult.Requirement.Name),
				Description: reqResult.Requirement.Description,
				Severity:    common.Low,
				Category:    "compliance_requirement",
				Status:      reqStatus,
				Details:     fmt.Sprintf("Requirement: %s, Compliant: %t", reqResult.Requirement.Name, reqResult.Compliant),
				RawData:     reqResult,
				Metadata: map[string]interface{}{
					"standard_id":      string(standardResult.Standard),
					"requirement_id":   reqResult.Requirement.ID,
					"requirement_name": reqResult.Requirement.Name,
					"compliant":        reqResult.Compliant,
				},
			}

			testResults = append(testResults, reqTestResult)
		}
	}

	return testResults
}

// VerifyTemplateSecurityAndCompliance verifies template security and compliance
func (ri *ReportingIntegration) VerifyTemplateSecurityAndCompliance(ctx context.Context, templatePath string, testSuite *types.TestSuite, options *security.VerificationOptions) (*TemplateComplianceResult, error) {
	if options == nil {
		options = security.DefaultVerificationOptions()
	}

	// Verify template security
	verificationResult, err := ri.templateVerifier.VerifyTemplateFile(ctx, templatePath, options)
	if err != nil {
		return nil, fmt.Errorf("failed to verify template security: %w", err)
	}

	// Get compliance status for the test suite
	complianceOptions := &ComplianceReportOptions{
		Title:     "Template Compliance Report",
		Standards: []ComplianceStandard{OWASPLLMTop10, ISO42001},
	}

	complianceReport, err := ri.GenerateComplianceReport(ctx, testSuite, complianceOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to generate compliance report: %w", err)
	}

	// Create template compliance result
	result := &TemplateComplianceResult{
		TemplatePath:        templatePath,
		TemplateID:          verificationResult.TemplateID,
		TemplateName:        verificationResult.TemplateName,
		SecurityResult:      verificationResult,
		ComplianceReport:    complianceReport,
		OverallCompliance:   true,
		ComplianceByStandard: make(map[string]bool),
	}

	// Calculate overall compliance
	for _, standardResult := range complianceReport.StandardResults {
		compliant := standardResult.CompliancePercentage >= 80.0
		result.ComplianceByStandard[string(standardResult.Standard)] = compliant
		
		if !compliant {
			result.OverallCompliance = false
		}
	}

	return result, nil
}

// TemplateComplianceResult represents the combined result of template security verification and compliance mapping
type TemplateComplianceResult struct {
	TemplatePath         string                 `json:"template_path"`
	TemplateID           string                 `json:"template_id"`
	TemplateName         string                 `json:"template_name"`
	SecurityResult       *security.VerificationResult `json:"security_result"`
	ComplianceReport     *ComplianceReport      `json:"compliance_report"`
	OverallCompliance    bool                   `json:"overall_compliance"`
	ComplianceByStandard map[string]bool        `json:"compliance_by_standard"`
}

// ConvertTemplateComplianceToTestResults converts a template compliance result to test results
func (ri *ReportingIntegration) ConvertTemplateComplianceToTestResults(result *TemplateComplianceResult) []*common.TestResult {
	var testResults []*common.TestResult

	// Add security verification test results
	securityTestResult := security.VerificationResultToTestResult(result.SecurityResult)
	testResults = append(testResults, securityTestResult)

	// Add compliance test results
	complianceTestResults := ri.ConvertToTestResults(result.ComplianceReport)
	testResults = append(testResults, complianceTestResults...)

	// Add overall template compliance result
	status := "passed"
	if !result.OverallCompliance || !result.SecurityResult.Passed {
		status = "failed"
	}

	overallResult := &common.TestResult{
		ID:          fmt.Sprintf("template_compliance_%s", result.TemplateID),
		Name:        fmt.Sprintf("Template Compliance: %s", result.TemplateName),
		Description: "Overall template security and compliance verification",
		Severity:    common.High,
		Category:    "template_compliance",
		Status:      status,
		Details:     fmt.Sprintf("Template: %s, Security: %t, Compliance: %t",
			result.TemplateName,
			result.SecurityResult.Passed,
			result.OverallCompliance),
		RawData: result,
		Metadata: map[string]interface{}{
			"template_id":          result.TemplateID,
			"template_name":        result.TemplateName,
			"security_passed":      result.SecurityResult.Passed,
			"security_score":       result.SecurityResult.Score,
			"overall_compliance":   result.OverallCompliance,
			"compliance_by_standard": result.ComplianceByStandard,
		},
	}

	testResults = append(testResults, overallResult)

	return testResults
}

// getComplianceStandardNames returns the names of the standards
func getComplianceStandardNames(standards []ComplianceStandard) []string {
	var names []string
	for _, standard := range standards {
		names = append(names, string(standard))
	}
	return names
}
