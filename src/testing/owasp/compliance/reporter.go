// Package compliance provides mapping between test results and compliance standards
package compliance

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/perplext/LLMrecon/src/testing/owasp/types"
	"github.com/perplext/LLMrecon/src/vulnerability/detection"
)

// ComplianceReporterImpl implements the ComplianceReporter interface
type ComplianceReporterImpl struct {
	mapper ComplianceMapper
}

// NewComplianceReporter creates a new compliance reporter
func NewComplianceReporter(mapper ComplianceMapper) *ComplianceReporterImpl {
	return &ComplianceReporterImpl{
		mapper: mapper,
	}
}

// GenerateReport generates a compliance report for a test suite
func (r *ComplianceReporterImpl) GenerateReport(ctx context.Context, testSuite *types.TestSuite, options *ComplianceReportOptions) (*ComplianceReport, error) {
	if testSuite == nil {
		return nil, fmt.Errorf("test suite cannot be nil")
	}

	if options == nil {
		options = &ComplianceReportOptions{
			Standards:              []ComplianceStandard{OWASPLM, ISO42001},
			IncludeRecommendations: true,
			IncludeTestResults:     true,
			Format:                 "json",
			Title:                  "Compliance Report",
		}
	}

	// Create a new compliance report
	report := &ComplianceReport{
		Title:       options.Title,
		TestSuite:   testSuite,
		Standards:   options.Standards,
		Results:     make(map[ComplianceStandard]*StandardComplianceResult),
		GeneratedAt: time.Now().Format(time.RFC3339),
		Metadata:    options.Metadata,
	}

	// Map test suite to compliance requirements
	mappings, err := r.mapper.MapTestSuite(ctx, testSuite)
	if err != nil {
		return nil, fmt.Errorf("error mapping test suite: %w", err)
	}

	// Process each standard
	for _, standard := range options.Standards {
		standardResult, err := r.generateStandardResult(ctx, testSuite, standard, mappings, options)
		if err != nil {
			return nil, fmt.Errorf("error generating standard result: %w", err)
		}
		report.Results[standard] = standardResult
	}

	return report, nil
}

// ExportReport exports a compliance report to a file
func (r *ComplianceReporterImpl) ExportReport(ctx context.Context, report *ComplianceReport, format string, outputPath string) error {
	if report == nil {
		return fmt.Errorf("report cannot be nil")
	}

	// Create the output directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	// Export the report based on the format
	switch format {
	case "json":
		return r.exportAsJSON(report, outputPath)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// GetComplianceStatus returns the compliance status for a test suite
func (r *ComplianceReporterImpl) GetComplianceStatus(ctx context.Context, testSuite *types.TestSuite, standard ComplianceStandard) (*StandardComplianceResult, error) {
	if testSuite == nil {
		return nil, fmt.Errorf("test suite cannot be nil")
	}

	// Map test suite to compliance requirements
	mappings, err := r.mapper.MapTestSuite(ctx, testSuite)
	if err != nil {
		return nil, fmt.Errorf("error mapping test suite: %w", err)
	}

	// Generate the standard result
	options := &ComplianceReportOptions{
		Standards:              []ComplianceStandard{standard},
		IncludeRecommendations: true,
		IncludeTestResults:     true,
	}
	return r.generateStandardResult(ctx, testSuite, standard, mappings, options)
}

// generateStandardResult generates compliance results for a specific standard
func (r *ComplianceReporterImpl) generateStandardResult(ctx context.Context, testSuite *types.TestSuite, standard ComplianceStandard, mappings map[types.VulnerabilityType][]*ComplianceMapping, options *ComplianceReportOptions) (*StandardComplianceResult, error) {
	// Get all requirements for the standard
	requirements, err := r.mapper.GetRequirementsForStandard(ctx, standard)
	if err != nil {
		return nil, fmt.Errorf("error getting requirements for standard: %w", err)
	}

	// Create a new standard result
	result := &StandardComplianceResult{
		Standard:          standard,
		RequirementResults: make(map[string]*RequirementComplianceResult),
		TotalRequirements: len(requirements),
	}

	// Process each requirement
	for _, req := range requirements {
		reqResult := &RequirementComplianceResult{
			Requirement:        req,
			Compliant:          true, // Assume compliant until a failing test is found
			TestResults:        []*types.TestResult{},
			VulnerabilityTypes: []types.VulnerabilityType{},
			HighestSeverity:    detection.Info,
			Recommendations:    []string{},
		}

		// Find all test results related to this requirement
		for vulnType, vulnMappings := range mappings {
			for _, mapping := range vulnMappings {
				for _, mappingReq := range mapping.Requirements {
					if mappingReq.ID == req.ID && mappingReq.Standard == req.Standard {
						// This vulnerability type is related to this requirement
						reqResult.VulnerabilityTypes = append(reqResult.VulnerabilityTypes, vulnType)

						// Find all test results for this vulnerability type
						for _, testResult := range testSuite.Results {
							if testResult.TestCase.VulnerabilityType == vulnType {
								reqResult.TestResults = append(reqResult.TestResults, testResult)

								// If any test fails, the requirement is not compliant
								if !testResult.Passed {
									reqResult.Compliant = false

									// Update the highest severity
									for _, detectionResult := range testResult.DetectionResults {
										if detectionResult.Severity > reqResult.HighestSeverity {
											reqResult.HighestSeverity = detectionResult.Severity
										}
									}

									// Add recommendations if enabled
									if options.IncludeRecommendations {
										recommendation := fmt.Sprintf("Fix issues in test case '%s' (%s) to comply with %s %s", 
											testResult.TestCase.Name, 
											testResult.TestCase.ID, 
											req.Standard, 
											req.ID)
										reqResult.Recommendations = append(reqResult.Recommendations, recommendation)
									}
								}
							}
						}
					}
				}
			}
		}

		// Add the requirement result to the standard result
		result.RequirementResults[req.ID] = reqResult

		// Update the passed requirements count
		if reqResult.Compliant {
			result.PassedRequirements++
		}
	}

	// Calculate the overall compliance percentage
	if result.TotalRequirements > 0 {
		result.OverallCompliance = float64(result.PassedRequirements) / float64(result.TotalRequirements) * 100
	}

	// Generate a summary
	result.Summary = fmt.Sprintf("Compliance with %s: %.2f%% (%d/%d requirements met)", 
		standard, 
		result.OverallCompliance, 
		result.PassedRequirements, 
		result.TotalRequirements)

	return result, nil
}

// exportAsJSON exports the report as JSON
func (r *ComplianceReporterImpl) exportAsJSON(report *ComplianceReport, outputPath string) error {
	// Marshal the report to JSON
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling report to JSON: %w", err)
	}

	// Write the JSON to the output file
	if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing report to file: %w", err)
	}

	return nil
}
