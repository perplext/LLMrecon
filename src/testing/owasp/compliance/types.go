// Package compliance provides mapping between test results and compliance standards
package compliance

import (
	"github.com/perplext/LLMrecon/src/testing/owasp/types"
	"github.com/perplext/LLMrecon/src/vulnerability/detection"
)

// ComplianceStandard represents a compliance standard
type ComplianceStandard string

// ID returns the ID of the compliance standard
func (s ComplianceStandard) ID() string {
	return string(s)
}

// Name returns the name of the compliance standard
func (s ComplianceStandard) Name() string {
	return string(s)
}

// Supported compliance standards
const (
	OWASPLM        ComplianceStandard = "OWASP-LLM"
	OWASPLLMTop10  ComplianceStandard = "OWASP-LLM-TOP-10"
	ISO42001       ComplianceStandard = "ISO-42001"
	NIST800        ComplianceStandard = "NIST-800"
	GDPR           ComplianceStandard = "GDPR"
	HIPAA          ComplianceStandard = "HIPAA"
	SOC2           ComplianceStandard = "SOC2"
	CustomStandard ComplianceStandard = "CUSTOM"
)

// ComplianceRequirement represents a specific requirement within a compliance standard
type ComplianceRequirement struct {
	// ID is the identifier of the requirement (e.g., "LLM01", "8.2.3")
	ID string `json:"id"`
	// Standard is the compliance standard this requirement belongs to
	Standard ComplianceStandard `json:"standard"`
	// Name is the name of the requirement
	Name string `json:"name"`
	// Description is a description of the requirement
	Description string `json:"description"`
	// Category is the category of the requirement within the standard
	Category string `json:"category"`
	// References contains links to additional information about the requirement
	References []string `json:"references"`
}

// ComplianceMapping represents a mapping between a vulnerability type and compliance requirements
type ComplianceMapping struct {
	// VulnerabilityType is the type of vulnerability
	VulnerabilityType types.VulnerabilityType `json:"vulnerability_type"`
	// Requirements is a list of compliance requirements that this vulnerability maps to
	Requirements []*ComplianceRequirement `json:"requirements"`
}

// ComplianceReport represents a compliance report for a test suite
type ComplianceReport struct {
	// Title is the title of the report
	Title string `json:"title"`
	// TestSuite is the test suite the report is based on
	TestSuite *types.TestSuite `json:"test_suite"`
	// Standards is a list of compliance standards covered in the report
	Standards []ComplianceStandard `json:"standards"`
	// Results contains the compliance results for each standard
	Results map[ComplianceStandard]*StandardComplianceResult `json:"results"`
	// StandardResults is an alias for Results for backward compatibility
	StandardResults map[ComplianceStandard]*StandardComplianceResult `json:"standard_results"`
	// GeneratedAt is the time the report was generated
	GeneratedAt string `json:"generated_at"`
	// Metadata contains additional metadata for the report
	Metadata map[string]interface{} `json:"metadata"`
	// OverallCompliance is the overall compliance percentage across all standards
	OverallCompliance float64 `json:"overall_compliance"`
	// Options contains the options used to generate the report
	Options *ComplianceReportOptions `json:"options,omitempty"`
}

// StandardComplianceResult represents compliance results for a specific standard
type StandardComplianceResult struct {
	// Standard is the compliance standard
	Standard ComplianceStandard `json:"standard"`
	// RequirementResults contains results for each requirement
	RequirementResults map[string]*RequirementComplianceResult `json:"requirement_results"`
	// OverallCompliance is the overall compliance percentage
	OverallCompliance float64 `json:"overall_compliance"`
	// CompliancePercentage is an alias for OverallCompliance for backward compatibility
	CompliancePercentage float64 `json:"compliance_percentage"`
	// PassedRequirements is the number of requirements that passed
	PassedRequirements int `json:"passed_requirements"`
	// RequirementsMet is an alias for PassedRequirements for backward compatibility
	RequirementsMet int `json:"requirements_met"`
	// TotalRequirements is the total number of requirements
	TotalRequirements int `json:"total_requirements"`
	// Summary is a summary of the compliance results
	Summary string `json:"summary"`
}

// RequirementComplianceResult represents compliance results for a specific requirement
type RequirementComplianceResult struct {
	// Requirement is the compliance requirement
	Requirement *ComplianceRequirement `json:"requirement"`
	// Compliant indicates if the requirement is compliant
	Compliant bool `json:"compliant"`
	// TestResults contains the test results related to this requirement
	TestResults []*types.TestResult `json:"test_results"`
	// VulnerabilityTypes contains the vulnerability types related to this requirement
	VulnerabilityTypes []types.VulnerabilityType `json:"vulnerability_types"`
	// HighestSeverity is the highest severity level found in the test results
	HighestSeverity detection.SeverityLevel `json:"highest_severity"`
	// Recommendations contains recommendations for addressing non-compliance
	Recommendations []string `json:"recommendations"`
	// Reason contains the reason for compliance or non-compliance
	Reason string `json:"reason,omitempty"`
}

// ComplianceReportOptions represents options for generating a compliance report
type ComplianceReportOptions struct {
	// Standards is a list of compliance standards to include in the report
	Standards []ComplianceStandard `json:"standards"`
	// IncludeRecommendations indicates whether to include recommendations in the report
	IncludeRecommendations bool `json:"include_recommendations"`
	// IncludeTestResults indicates whether to include detailed test results in the report
	IncludeTestResults bool `json:"include_test_results"`
	// Format is the format of the report
	Format string `json:"format"`
	// OutputPath is the path to save the report to
	OutputPath string `json:"output_path"`
	// Title is the title of the report
	Title string `json:"title"`
	// Metadata contains additional metadata for the report
	Metadata map[string]interface{} `json:"metadata"`
}
