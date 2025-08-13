// Package security provides template security verification mechanisms
package security

import (
	"github.com/perplext/LLMrecon/src/reporting/common"
)

// SecurityIssueType represents the type of security issue found in a template
type SecurityIssueType string

// Security issue types
const (
	InjectionVulnerability SecurityIssueType = "injection_vulnerability"
	UnsanitizedInput       SecurityIssueType = "unsanitized_input"
	InsecurePattern        SecurityIssueType = "insecure_pattern"
	MissingValidation      SecurityIssueType = "missing_validation"
	OverpermissiveRegex    SecurityIssueType = "overpermissive_regex"
	DataLeakage            SecurityIssueType = "data_leakage"
	TemplateFormatError    SecurityIssueType = "template_format_error"
)

// SecurityIssue represents a security issue found in a template
type SecurityIssue struct {
	ID          string              `json:"id"`
	Type        SecurityIssueType   `json:"type"`
	Description string              `json:"description"`
	Location    string              `json:"location"`
	Severity    common.SeverityLevel `json:"severity"`
	Remediation string              `json:"remediation"`
	Context     string              `json:"context,omitempty"`
	LineNumber  int                 `json:"line_number,omitempty"`
	RawData     interface{}         `json:"raw_data,omitempty"`
}

// VerificationResult represents the result of a template security verification
type VerificationResult struct {
	TemplatePath string          `json:"template_path"`
	TemplateID   string          `json:"template_id"`
	TemplateName string          `json:"template_name"`
	Issues       []*SecurityIssue `json:"issues"`
	Passed       bool            `json:"passed"`
	Score        float64         `json:"score"`
	MaxScore     float64         `json:"max_score"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// VerificationOptions represents options for template security verification
type VerificationOptions struct {
	StrictMode         bool                   `json:"strict_mode"`
	IgnorePatterns     []string               `json:"ignore_patterns,omitempty"`
	CustomChecks       []string               `json:"custom_checks,omitempty"`
	SeverityThreshold  common.SeverityLevel    `json:"severity_threshold,omitempty"`
	IncludeInfo        bool                   `json:"include_info"`
	TemplateCategories []string               `json:"template_categories,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// DefaultVerificationOptions returns the default verification options
func DefaultVerificationOptions() *VerificationOptions {
	return &VerificationOptions{
		StrictMode:        false,
		IgnorePatterns:    []string{},
		CustomChecks:      []string{},
		SeverityThreshold: common.Low,
		IncludeInfo:       true,
		Metadata:          make(map[string]interface{}),
	}
}

// VerificationResultToTestResult converts a verification result to a test result
func VerificationResultToTestResult(result *VerificationResult) *common.TestResult {
	status := "passed"
	if !result.Passed {
		status = "failed"
	}

	severity := common.Medium
	if len(result.Issues) > 0 {
		// Find the highest severity issue
		for _, issue := range result.Issues {
			if isSeverityHigher(issue.Severity, severity) {
				severity = issue.Severity
			}
		}
	}

	details := ""
	if len(result.Issues) > 0 {
		details = "Security issues found in template: "
		for i, issue := range result.Issues {
			if i > 0 {
				details += "; "
			}
			details += issue.Description
		}
	} else {
		details = "No security issues found in template"
	}

	return &common.TestResult{
		ID:          result.TemplateID,
		Name:        "Template Security Verification: " + result.TemplateName,
		Description: "Security verification for template " + result.TemplatePath,
		Severity:    severity,
		Category:    "template_security",
		Status:      status,
		Details:     details,
		RawData:     result,
		Metadata:    result.Metadata,
	}
}

// isSeverityHigher returns true if severity1 is higher than severity2
func isSeverityHigher(severity1, severity2 common.SeverityLevel) bool {
	severityMap := map[common.SeverityLevel]int{
		common.Critical: 5,
		common.High:     4,
		common.Medium:   3,
		common.Low:      2,
		common.Info:     1,
	}

	return severityMap[severity1] > severityMap[severity2]
}
