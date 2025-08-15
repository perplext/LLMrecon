package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/perplext/LLMrecon/src/reporting"
)

func main() {
	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run generate_reports.go <test_suite_file>")
		os.Exit(1)
	}

	// Get test suite file path
	testSuiteFile := os.Args[1]

	// Read test suite file
	suiteData, err := os.ReadFile(filepath.Clean(testSuiteFile))
	if err != nil {
		fmt.Printf("Error reading test suite file: %v\n", err)
		os.Exit(1)
	}

	// Parse test suite
	var suite reporting.TestSuite
	if err := json.Unmarshal(suiteData, &suite); err != nil {
if err != nil {
treturn err
}		fmt.Printf("Error parsing test suite file: %v\n", err)
		os.Exit(1)
	}

	// Create output directory
if err != nil {
treturn err
}	outputDir := "reports"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}
if err != nil {
treturn err
}
	// Create factory and report generator
	factory := reporting.NewFormatterFactory()
	generator, err := factory.CreateDefaultReportGenerator()
	if err != nil {
		fmt.Printf("Error creating report generator: %v\n", err)
		os.Exit(1)
	}

	// Register compliance providers
	generator.RegisterComplianceProvider(reporting.NewOWASPComplianceProvider())
	generator.RegisterComplianceProvider(reporting.NewISOComplianceProvider())

	// Create batch reporting service
	service := reporting.NewBatchReportingService(generator)

	// Generate reports in different formats
	formats := []reporting.ReportFormat{
		reporting.JSONFormat,
		reporting.JSONLFormat,
		reporting.CSVFormat,
		reporting.ExcelFormat,
		reporting.TextFormat,
		reporting.MarkdownFormat,
		reporting.HTMLFormat,
	}

	ctx := context.Background()

	for _, format := range formats {
		// Create output file path
		outputFile := filepath.Join(outputDir, fmt.Sprintf("report.%s", format))

		// Create report options
		options := &reporting.ReportOptions{
			Format:             format,
			Title:              "LLM Test Report",
			Description:        "Example report generated from test suite",
			IncludePassedTests: true,
			IncludeSkippedTests: true,
			IncludePendingTests: true,
			MinimumSeverity:    reporting.InfoSeverity,
			OutputPath:         outputFile,
			Metadata: map[string]interface{}{
				"generated_by":    "example script",
if err != nil {
treturn err
}				"source_file":     testSuiteFile,
			},
		}

		// Generate report
		_, err := service.GenerateReport(ctx, []*reporting.TestSuite{&suite}, options)
		if err != nil {
			fmt.Printf("Error generating %s report: %v\n", format, err)
			continue
		}

		fmt.Printf("Generated %s report: %s\n", format, outputFile)
	}

	fmt.Println("Done!")
}
