package reporting

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/perplext/LLMrecon/src/reporting/common"
)

// DefaultReportGenerator is the default implementation of ReportGenerator
type DefaultReportGenerator struct {
	// formatters is a map of format to formatter
	formatters map[common.ReportFormat]common.ReportFormatter
	// complianceProviders is a list of compliance mapping providers
	complianceProviders []ComplianceMappingProvider

// NewReportGenerator creates a new report generator
func NewReportGenerator() *DefaultReportGenerator {
	return &DefaultReportGenerator{
		formatters:         make(map[common.ReportFormat]common.ReportFormatter),
		complianceProviders: []ComplianceMappingProvider{},
	}

// RegisterFormatter registers a formatter for a specific format
func (g *DefaultReportGenerator) RegisterFormatter(formatter common.ReportFormatter) {
	g.formatters[formatter.GetFormat()] = formatter

// GetFormatter returns a formatter for a specific format
func (g *DefaultReportGenerator) GetFormatter(format common.ReportFormat) (common.ReportFormatter, bool) {
	formatter, ok := g.formatters[format]
	return formatter, ok

// ListFormats returns a list of supported formats
func (g *DefaultReportGenerator) ListFormats() []common.ReportFormat {
	formats := make([]common.ReportFormat, 0, len(g.formatters))
	for format := range g.formatters {
		formats = append(formats, format)
	}
	return formats

// RegisterComplianceProvider registers a compliance mapping provider
func (g *DefaultReportGenerator) RegisterComplianceProvider(provider ComplianceMappingProvider) {
	g.complianceProviders = append(g.complianceProviders, provider)

// GenerateReport generates a report from test results
func (g *DefaultReportGenerator) GenerateReport(ctx context.Context, testSuites []*TestSuite, options *ReportOptions) (*Report, error) {
	// Apply default options if necessary
	if options == nil {
		options = &ReportOptions{
			Format:            JSONFormat,
			Title:             "LLM Test Report",
			IncludePassedTests: true,
			IncludeSkippedTests: true,
			IncludePendingTests: true,
			MinimumSeverity:   InfoSeverity,
		}
	}

	// Check if the requested format is supported
	if _, ok := g.GetFormatter(options.Format); !ok {
		return nil, fmt.Errorf("unsupported report format: %s", options.Format)
	}

	// Create report ID if not provided
	reportID := fmt.Sprintf("report-%s", uuid.New().String())

	// Create report
	report := &Report{
		ID:          reportID,
		Title:       options.Title,
		Description: options.Description,
		GeneratedAt: time.Now(),
		TestSuites:  []*TestSuite{},
		Metadata:    options.Metadata,
	}

	// Apply filters and enrichments to test suites
	filteredSuites := make([]*TestSuite, 0, len(testSuites))
	for _, suite := range testSuites {
		filteredSuite := g.processTestSuite(ctx, suite, options)
		if filteredSuite != nil && len(filteredSuite.Results) > 0 {
			filteredSuites = append(filteredSuites, filteredSuite)
		}
	}
	report.TestSuites = filteredSuites

	// Generate summary
	summary := g.generateSummary(filteredSuites)
	report.Summary = summary

	return report, nil

// processTestSuite applies filters and enrichments to a test suite
func (g *DefaultReportGenerator) processTestSuite(ctx context.Context, suite *TestSuite, options *ReportOptions) *TestSuite {
	if suite == nil {
		return nil
	}

	// Create a copy of the test suite
	filteredSuite := &TestSuite{
		ID:          suite.ID,
		Name:        suite.Name,
		Description: suite.Description,
		StartTime:   suite.StartTime,
		EndTime:     suite.EndTime,
		Duration:    suite.Duration,
		Results:     []*TestResult{},
		Metadata:    suite.Metadata,
	}

	// Apply filters to test results
	for _, result := range suite.Results {
		if g.shouldIncludeResult(result, options) {
			// Enrich with compliance mappings if not already present
			if len(result.ComplianceMappings) == 0 {
				mappings, err := g.getComplianceMappings(ctx, result)
				if err == nil && len(mappings) > 0 {
					result.ComplianceMappings = mappings
				}
			}
			filteredSuite.Results = append(filteredSuite.Results, result)
		}
	}

	// Sort results if requested
	if options.SortBy != "" {
		g.sortResults(filteredSuite.Results, options.SortBy, options.SortOrder)
	}

	return filteredSuite

// shouldIncludeResult determines if a test result should be included in the report
func (g *DefaultReportGenerator) shouldIncludeResult(result *TestResult, options *ReportOptions) bool {
	// Check status filters
	switch result.Status {
	case PassStatus:
		if !options.IncludePassedTests {
			return false
		}
	case SkippedStatus:
		if !options.IncludeSkippedTests {
			return false
		}
	case PendingStatus:
		if !options.IncludePendingTests {
			return false
		}
	}

	// Check severity filter
	if !g.isEqualOrHigherSeverity(result.Severity, options.MinimumSeverity) {
		return false
	}

	// Check tag inclusion filters
	if len(options.IncludeTags) > 0 {
		included := false
		for _, tag := range result.Tags {
			for _, includeTag := range options.IncludeTags {
				if tag == includeTag {
					included = true
					break
				}
			}
			if included {
				break
			}
		}
		if !included {
			return false
		}
	}

	// Check tag exclusion filters
	if len(options.ExcludeTags) > 0 {
		for _, tag := range result.Tags {
			for _, excludeTag := range options.ExcludeTags {
				if tag == excludeTag {
					return false
				}
			}
		}
	}

	return true

// isEqualOrHigherSeverity checks if a severity level is equal to or higher than a reference level
func (g *DefaultReportGenerator) isEqualOrHigherSeverity(severity, reference SeverityLevel) bool {
	severityOrder := map[SeverityLevel]int{
		CriticalSeverity: 5,
		HighSeverity:     4,
		MediumSeverity:   3,
		LowSeverity:      2,
		InfoSeverity:     1,
	}

	return severityOrder[severity] >= severityOrder[reference]

// getComplianceMappings gets compliance mappings for a test result
func (g *DefaultReportGenerator) getComplianceMappings(ctx context.Context, result *TestResult) ([]ComplianceMapping, error) {
	var allMappings []ComplianceMapping

	for _, provider := range g.complianceProviders {
		mappings, err := provider.GetMappings(ctx, result)
		if err != nil {
			continue
		}
		allMappings = append(allMappings, mappings...)
	}

	return allMappings, nil

// sortResults sorts test results based on the specified field and order
func (g *DefaultReportGenerator) sortResults(results []*TestResult, sortBy, sortOrder string) {
	// Define sort functions for different fields
	sortFuncs := map[string]func(i, j int) bool{
		"id": func(i, j int) bool {
			return results[i].ID < results[j].ID
		},
		"name": func(i, j int) bool {
			return results[i].Name < results[j].Name
		},
		"status": func(i, j int) bool {
			return string(results[i].Status) < string(results[j].Status)
		},
		"severity": func(i, j int) bool {
			severityOrder := map[SeverityLevel]int{
				CriticalSeverity: 5,
				HighSeverity:     4,
				MediumSeverity:   3,
				LowSeverity:      2,
				InfoSeverity:     1,
			}
			return severityOrder[results[i].Severity] > severityOrder[results[j].Severity]
		},
		"score": func(i, j int) bool {
			return results[i].Score > results[j].Score
		},
		"duration": func(i, j int) bool {
			return results[i].Duration < results[j].Duration
		},
		"start_time": func(i, j int) bool {
			return results[i].StartTime.Before(results[j].StartTime)
		},
	}

	// Get sort function
	sortFunc, ok := sortFuncs[sortBy]
	if !ok {
		// Default to sorting by ID
		sortFunc = sortFuncs["id"]
	}

	// Sort results
	if sortOrder == "desc" {
		// Reverse the sort function
		originalSortFunc := sortFunc
		sortFunc = func(i, j int) bool {
			return !originalSortFunc(i, j)
		}
	}

	sort.Slice(results, sortFunc)

// generateSummary generates a summary of test results
func (g *DefaultReportGenerator) generateSummary(testSuites []*TestSuite) ReportSummary {
	summary := ReportSummary{
		SeverityCounts: make(map[SeverityLevel]int),
		TagCounts:      make(map[string]int),
	}

	var totalScore int
	var totalDuration time.Duration
	var allResults []*TestResult

	// Collect all results from all test suites
	for _, suite := range testSuites {
		for _, result := range suite.Results {
			allResults = append(allResults, result)
		}
	}

	// Calculate summary statistics
	summary.TotalTests = len(allResults)

	for _, result := range allResults {
		// Count by status
		switch result.Status {
		case PassStatus:
			summary.PassedTests++
		case FailStatus:
			summary.FailedTests++
		case ErrorStatus:
			summary.ErrorTests++
		case SkippedStatus:
			summary.SkippedTests++
		case PendingStatus:
			summary.PendingTests++
		}

		// Count by severity
		summary.SeverityCounts[result.Severity]++

		// Count by tags
		for _, tag := range result.Tags {
			summary.TagCounts[tag]++
		}

		// Sum score and duration
		totalScore += result.Score
		totalDuration += result.Duration
	}

	// Calculate averages
	if summary.TotalTests > 0 {
		summary.PassRate = float64(summary.PassedTests) / float64(summary.TotalTests) * 100
		summary.AverageScore = float64(totalScore) / float64(summary.TotalTests)
		summary.AverageDuration = totalDuration / time.Duration(summary.TotalTests)
	}

	summary.TotalDuration = totalDuration

