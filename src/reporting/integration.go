package reporting

import (
	"context"
	"fmt"
	"strings"

	"github.com/perplext/LLMrecon/src/reporting/common"
	"github.com/perplext/LLMrecon/src/template/management/types"
)

// TemplateResultConverter converts template results to test results
type TemplateResultConverter struct {
	// complianceProviders is a list of compliance mapping providers
	complianceProviders []ComplianceMappingProvider
}

// NewTemplateResultConverter creates a new template result converter
func NewTemplateResultConverter(complianceProviders []ComplianceMappingProvider) *TemplateResultConverter {
	return &TemplateResultConverter{
		complianceProviders: complianceProviders,
	}
}

// ConvertToTestSuite converts template results to a test suite
func (c *TemplateResultConverter) ConvertToTestSuite(ctx context.Context, results []*types.TemplateResult, suiteID string, suiteName string) (*TestSuite, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no template results to convert")
	}

	// Create test suite
	suite := &TestSuite{
		ID:          suiteID,
		Name:        suiteName,
		Description: fmt.Sprintf("Test suite for %d templates", len(results)),
		StartTime:   results[0].StartTime,
		EndTime:     results[len(results)-1].EndTime,
		Duration:    calculateTotalDuration(results),
		Results:     make([]*TestResult, 0, len(results)),
		Metadata:    map[string]interface{}{
			"template_count": len(results),
		},
	}

	// Convert each template result to a test result
	for _, result := range results {
		testResult, err := c.convertToTestResult(ctx, result)
		if err != nil {
			return nil, fmt.Errorf("failed to convert template result: %w", err)
		}
		suite.Results = append(suite.Results, testResult)
	}

	return suite, nil
}

// convertToTestResult converts a template result to a test result
func (c *TemplateResultConverter) convertToTestResult(ctx context.Context, result *types.TemplateResult) (*TestResult, error) {
	// Map status
	status := c.mapStatus(result.Status, result.Detected)

	// Map severity
	severity := c.mapSeverity(result.Score)

	// Create test result
	testResult := &TestResult{
		ID:          result.TemplateID,
		Name:        result.TemplateName,
		Description: result.Description,
		Status:      status,
		Severity:    severity,
		Score:       result.Score,
		StartTime:   result.StartTime,
		EndTime:     result.EndTime,
		Duration:    result.Duration,
		Tags:        result.Tags,
		Metadata:    result.Metadata,
	}

	// Set error if present
	if result.Error != nil {
		testResult.Error = result.Error.Error()
	}

	// Set input/output if present
	if result.Input != "" {
		testResult.Input = result.Input
	}
	if result.Output != "" {
		testResult.ActualOutput = result.Output
	}

	// Get compliance mappings
	for _, provider := range c.complianceProviders {
		mappings, err := provider.GetMappings(ctx, testResult)
		if err != nil {
			continue
		}
		testResult.ComplianceMappings = append(testResult.ComplianceMappings, mappings...)
	}

	return testResult, nil
}

// mapStatus maps template status to test status
func (c *TemplateResultConverter) mapStatus(status types.TemplateStatus, detected bool) TestStatus {
	switch status {
	case types.StatusCompleted:
		if detected {
			return FailStatus
		}
		return PassStatus
	case types.StatusFailed:
		return ErrorStatus
	case types.StatusLoaded:
		return SkippedStatus
	case types.StatusValidated:
		return PendingStatus
	default:
		return PendingStatus
	}
}

func convertSeverity(severity string) common.SeverityLevel {
	switch strings.ToLower(severity) {
	case "critical":
		return common.Critical
	case "high":
		return common.High
	case "medium":
		return common.Medium
	case "low":
		return common.Low
	default:
		return common.Info
	}
}

// mapSeverity maps score to severity level
func (c *TemplateResultConverter) mapSeverity(score int) common.SeverityLevel {
	return convertSeverity(c.mapSeverityString(score))
}

func (c *TemplateResultConverter) mapSeverityString(score int) string {
	switch {
	case score >= 90:
		return "critical"
	case score >= 70:
		return "high"
	case score >= 40:
		return "medium"
	case score >= 10:
		return "low"
	default:
		return "info"
	}
}

// calculateTotalDuration calculates the total duration of all template results
func calculateTotalDuration(results []*types.TemplateResult) time.Duration {
	var total time.Duration
	for _, result := range results {
		total += result.Duration
	}
	return total
}

// TemplateReportingService provides reporting services for template results
type TemplateReportingService struct {
	// converter is the template result converter
	converter *TemplateResultConverter
	// generator is the report generator
	generator common.ReportGenerator
}

// NewTemplateReportingService creates a new template reporting service
func NewTemplateReportingService(converter *TemplateResultConverter, generator common.ReportGenerator) *TemplateReportingService {
	return &TemplateReportingService{
		converter: converter,
		generator: generator,
	}
}

// GenerateReport generates a report from template results
func (s *TemplateReportingService) GenerateReport(ctx context.Context, results []*types.TemplateResult, options *ReportOptions) ([]byte, error) {
	// Convert template results to test suite
	suite, err := s.converter.ConvertToTestSuite(ctx, results, "template-suite", "Template Execution Results")
	if err != nil {
		return nil, fmt.Errorf("failed to convert template results: %w", err)
	}

	// Generate report
	report, err := s.generator.GenerateReport(ctx, []*TestSuite{suite}, options)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	// Get formatter for the specified format
	formatter, ok := s.generator.GetFormatter(options.Format)
	if !ok {
		return nil, fmt.Errorf("unsupported report format: %s", options.Format)
	}

	// Format report
	data, err := formatter.Format(ctx, report, options)
	if err != nil {
		return nil, fmt.Errorf("failed to format report: %w", err)
	}

	return data, nil
}

// BatchReportingService provides reporting services for multiple test suites
type BatchReportingService struct {
	// generator is the report generator
	generator common.ReportGenerator
}

// NewBatchReportingService creates a new batch reporting service
func NewBatchReportingService(generator common.ReportGenerator) *BatchReportingService {
	return &BatchReportingService{
		generator: generator,
	}
}

// GenerateReport generates a report from multiple test suites
func (s *BatchReportingService) GenerateReport(ctx context.Context, suites []*TestSuite, options *ReportOptions) ([]byte, error) {
	// Generate report
	report, err := s.generator.GenerateReport(ctx, suites, options)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	// Get formatter for the specified format
	formatter, ok := s.generator.GetFormatter(options.Format)
	if !ok {
		return nil, fmt.Errorf("unsupported report format: %s", options.Format)
	}

	// Format report
	data, err := formatter.Format(ctx, report, options)
	if err != nil {
		return nil, fmt.Errorf("failed to format report: %w", err)
	}

	return data, nil
}
