// Package compliance provides mapping between test results and compliance standards
package compliance

import (
	"context"

	"github.com/perplext/LLMrecon/src/testing/owasp/types"
)

// ComplianceServiceImpl implements the ComplianceService interface
type ComplianceServiceImpl struct {
	mapper   ComplianceMapper
	reporter ComplianceReporter

// NewComplianceService creates a new compliance service
func NewComplianceService() *ComplianceServiceImpl {
	mapper := NewBaseComplianceMapper()
	reporter := NewComplianceReporter(mapper)

	return &ComplianceServiceImpl{
		mapper:   mapper,
		reporter: reporter,
	}

// MapTestResult maps a test result to compliance requirements
func (s *ComplianceServiceImpl) MapTestResult(ctx context.Context, testResult *types.TestResult) ([]*ComplianceMapping, error) {
	return s.mapper.MapTestResult(ctx, testResult)

// MapTestSuite maps a test suite to compliance requirements
func (s *ComplianceServiceImpl) MapTestSuite(ctx context.Context, testSuite *types.TestSuite) (map[types.VulnerabilityType][]*ComplianceMapping, error) {
	return s.mapper.MapTestSuite(ctx, testSuite)

// GetRequirementsForVulnerability returns compliance requirements for a specific vulnerability type
func (s *ComplianceServiceImpl) GetRequirementsForVulnerability(ctx context.Context, vulnerabilityType types.VulnerabilityType) ([]*ComplianceRequirement, error) {
	return s.mapper.GetRequirementsForVulnerability(ctx, vulnerabilityType)

// GetRequirementsForStandard returns all requirements for a specific compliance standard
func (s *ComplianceServiceImpl) GetRequirementsForStandard(ctx context.Context, standard ComplianceStandard) ([]*ComplianceRequirement, error) {
	return s.mapper.GetRequirementsForStandard(ctx, standard)

// GetSupportedStandards returns all supported compliance standards
func (s *ComplianceServiceImpl) GetSupportedStandards(ctx context.Context) ([]ComplianceStandard, error) {
	return s.mapper.GetSupportedStandards(ctx)

// GenerateReport generates a compliance report for a test suite
func (s *ComplianceServiceImpl) GenerateReport(ctx context.Context, testSuite *types.TestSuite, options *ComplianceReportOptions) (*ComplianceReport, error) {
	return s.reporter.GenerateReport(ctx, testSuite, options)

// ExportReport exports a compliance report to a file
func (s *ComplianceServiceImpl) ExportReport(ctx context.Context, report *ComplianceReport, format string, outputPath string) error {
	return s.reporter.ExportReport(ctx, report, format, outputPath)

// GetComplianceStatus returns the compliance status for a test suite
func (s *ComplianceServiceImpl) GetComplianceStatus(ctx context.Context, testSuite *types.TestSuite, standard ComplianceStandard) (*StandardComplianceResult, error) {
	return s.reporter.GetComplianceStatus(ctx, testSuite, standard)
