// Package security provides template security verification mechanisms
package security

import (
	"fmt"

	"github.com/perplext/LLMrecon/src/reporting/common"
)

// TemplateSecurityReporter is the interface for template security reporters
type TemplateSecurityReporter interface {
	// GenerateReport generates a report from verification results
	GenerateReport(results []*VerificationResult) (*VerificationReport, error)

	// CalculateSummary calculates a summary of verification results
	CalculateSummary(results []*VerificationResult) *VerificationSummary

// DefaultTemplateSecurityReporter is the default implementation of TemplateSecurityReporter
type DefaultTemplateSecurityReporter struct {

}
// NewDefaultTemplateSecurityReporter creates a new default template security reporter
func NewDefaultTemplateSecurityReporter() *DefaultTemplateSecurityReporter {
	return &DefaultTemplateSecurityReporter{}

// VerificationReport represents a template security verification report
type VerificationReport struct {
	// Results is the list of verification results
	Results []*VerificationResult

	// Summary is a summary of the verification results
	Summary *VerificationSummary

	// GeneratedAt is the timestamp when the report was generated
	GeneratedAt time.Time

// VerificationSummary represents a summary of template security verification results
type VerificationSummary struct {
	// TotalTemplates is the total number of templates verified
	TotalTemplates int

	// PassedTemplates is the number of templates that passed verification
	PassedTemplates int

	// FailedTemplates is the number of templates that failed verification
	FailedTemplates int

	// TotalIssues is the total number of issues found
	TotalIssues int

	// IssuesBySeverity is a map of issues by severity
	IssuesBySeverity map[string]int

	// IssuesByType is a map of issues by type
	IssuesByType map[string]int

	// ComplianceStatus is a map of compliance status by standard
	ComplianceStatus map[string]bool

	// CompliancePercentage is the overall compliance percentage
	CompliancePercentage float64

// GenerateReport generates a report from verification results
func (r *DefaultTemplateSecurityReporter) GenerateReport(results []*VerificationResult) (*VerificationReport, error) {
	if results == nil {
		return nil, fmt.Errorf("results cannot be nil")
	}

	summary := r.CalculateSummary(results)

	report := &VerificationReport{
		Results:     results,
		Summary:     summary,
		GeneratedAt: time.Now(),
	}

	return report, nil

// CalculateSummary calculates a summary of verification results
func (r *DefaultTemplateSecurityReporter) CalculateSummary(results []*VerificationResult) *VerificationSummary {
	if results == nil || len(results) == 0 {
		return &VerificationSummary{
			TotalTemplates:       0,
			PassedTemplates:      0,
			FailedTemplates:      0,
			TotalIssues:          0,
			IssuesBySeverity:     make(map[string]int),
			IssuesByType:         make(map[string]int),
			ComplianceStatus:     make(map[string]bool),
			CompliancePercentage: 0.0,
		}
	}

	summary := &VerificationSummary{
		TotalTemplates:   len(results),
		PassedTemplates:  0,
		FailedTemplates:  0,
		TotalIssues:      0,
		IssuesBySeverity: make(map[string]int),
		IssuesByType:     make(map[string]int),
		ComplianceStatus: make(map[string]bool),
	}

	// Calculate statistics
	for _, result := range results {
		if result.Passed {
			summary.PassedTemplates++
		} else {
			summary.FailedTemplates++
		}

		// Count issues
		summary.TotalIssues += len(result.Issues)

		// Count by severity
		for _, issue := range result.Issues {
			summary.IssuesBySeverity[string(issue.Severity)]++
			summary.IssuesByType[string(issue.Type)]++
		}
	}

	// Calculate compliance percentage
	if summary.TotalTemplates > 0 {
		summary.CompliancePercentage = float64(summary.PassedTemplates) / float64(summary.TotalTemplates) * 100.0
	}

	// Set compliance status
	summary.ComplianceStatus["OWASP LLM Top 10"] = summary.CompliancePercentage >= 80.0
	summary.ComplianceStatus["Security Best Practices"] = summary.CompliancePercentage >= 90.0
	summary.ComplianceStatus["Template Security Standard"] = summary.CompliancePercentage >= 95.0

	return summary

// ConvertToTestResults converts a verification report to test results
func (r *DefaultTemplateSecurityReporter) ConvertToTestResults(report *VerificationReport) []*common.TestResult {
	if report == nil {
		return []*common.TestResult{}
	}

	testResults := make([]*common.TestResult, 0)

	// Convert each verification result to a test result
	for _, result := range report.Results {
		// Create a test result for the overall verification
		testResult := &common.TestResult{
			ID:          result.TemplateID,
			Name:        fmt.Sprintf("Template Security Verification: %s", result.TemplateName),
			Description: fmt.Sprintf("Security verification for template %s", result.TemplatePath),
			Severity:    common.Medium,
			Category:    "template_security",
			Status:      getComplianceStatus(result.Passed),
			Details:     fmt.Sprintf("Score: %.2f/%.2f", result.Score, result.MaxScore),
			RawData:     result,
			Metadata: map[string]interface{}{
				"template_path": result.TemplatePath,
				"score":         result.Score,
				"max_score":     result.MaxScore,
				"passed":        result.Passed,
			},
		}
		testResults = append(testResults, testResult)

		// Create a test result for each issue
		for _, issue := range result.Issues {
			issueResult := &common.TestResult{
				ID:          fmt.Sprintf("%s-%s", result.TemplateID, issue.ID),
				Name:        fmt.Sprintf("Security Issue: %s", issue.Description),
				Description: issue.Description,
				Severity:    common.SeverityLevel(issue.Severity),
				Category:    string(issue.Type),
				Status:      "failed",
				Details:     fmt.Sprintf("Location: %s\nRemediation: %s", issue.Location, issue.Remediation),
				RawData:     issue,
				Metadata: map[string]interface{}{
					"template_id":   result.TemplateID,
					"template_path": result.TemplatePath,
					"location":      issue.Location,
					"remediation":   issue.Remediation,
					"context":       issue.Context,
				},
			}
			testResults = append(testResults, issueResult)
		}
	}

	// Add a summary test result
	if report.Summary != nil {
		summaryResult := &common.TestResult{
			ID:          "template_security_summary",
			Name:        "Template Security Verification Summary",
			Description: "Summary of template security verification results",
			Severity:    common.Medium,
			Category:    "template_security",
			Status:      getComplianceStatus(report.Summary.ComplianceStatus["OWASP LLM Top 10"]),
			Details:     fmt.Sprintf("Total templates: %d, Passed: %d, Failed: %d, Compliance: %.2f%%",
				report.Summary.TotalTemplates,
				report.Summary.PassedTemplates,
				report.Summary.FailedTemplates,
				report.Summary.CompliancePercentage),
			RawData: report.Summary,
			Metadata: map[string]interface{}{
				"compliance_status":     report.Summary.ComplianceStatus,
				"compliance_percentage": report.Summary.CompliancePercentage,
				"issues_by_severity":    report.Summary.IssuesBySeverity,
				"issues_by_type":        report.Summary.IssuesByType,
			},
		}
		testResults = append(testResults, summaryResult)
	}

	return testResults

// Note: getComplianceStatus function is defined in pipeline.go
}
}
}
