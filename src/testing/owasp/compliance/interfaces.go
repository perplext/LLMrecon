// Package compliance provides mapping between test results and compliance standards
package compliance

import (
	"context"

	"github.com/perplext/LLMrecon/src/testing/owasp/types"
)

// ComplianceMapper is an interface for mapping test results to compliance standards
type ComplianceMapper interface {
	// MapTestResult maps a test result to compliance requirements
	MapTestResult(ctx context.Context, testResult *types.TestResult) ([]*ComplianceMapping, error)
	
	// MapTestSuite maps a test suite to compliance requirements
	MapTestSuite(ctx context.Context, testSuite *types.TestSuite) (map[types.VulnerabilityType][]*ComplianceMapping, error)
	
	// GetRequirementsForVulnerability returns compliance requirements for a specific vulnerability type
	GetRequirementsForVulnerability(ctx context.Context, vulnerabilityType types.VulnerabilityType) ([]*ComplianceRequirement, error)
	
	// GetRequirementsForStandard returns all requirements for a specific compliance standard
	GetRequirementsForStandard(ctx context.Context, standard ComplianceStandard) ([]*ComplianceRequirement, error)
	
	// GetSupportedStandards returns all supported compliance standards
	GetSupportedStandards(ctx context.Context) ([]ComplianceStandard, error)
}

// ComplianceReporter is an interface for generating compliance reports
type ComplianceReporter interface {
	// GenerateReport generates a compliance report for a test suite
	GenerateReport(ctx context.Context, testSuite *types.TestSuite, options *ComplianceReportOptions) (*ComplianceReport, error)
	
	// ExportReport exports a compliance report to a file
	ExportReport(ctx context.Context, report *ComplianceReport, format string, outputPath string) error
	
	// GetComplianceStatus returns the compliance status for a test suite
	GetComplianceStatus(ctx context.Context, testSuite *types.TestSuite, standard ComplianceStandard) (*StandardComplianceResult, error)
}

// ComplianceService is an interface that combines mapping and reporting capabilities
type ComplianceService interface {
	ComplianceMapper
	ComplianceReporter
}
