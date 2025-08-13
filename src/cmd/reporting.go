//go:build ignore

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/perplext/LLMrecon/src/reporting"
	"github.com/perplext/LLMrecon/src/template/management"
	"github.com/spf13/cobra"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate reports from test results",
	Long: `Generate comprehensive reports from LLM test results in various formats.
Supports multiple output formats including JSON, JSONL, CSV, Excel, Text, Markdown, PDF, and HTML.
Reports can include compliance mappings to frameworks like OWASP Top 10 for LLMs and ISO/IEC 42001.`,
}

// generateReportCmd represents the generate command
var generateReportCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a report from test results",
	Long: `Generate a comprehensive report from LLM test results.
Example:
  LLMrecon report generate --results results.json --format html --output report.html`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		resultsFile, _ := cmd.Flags().GetString("results")
		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("output")
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		minSeverity, _ := cmd.Flags().GetString("min-severity")
		includePassed, _ := cmd.Flags().GetBool("include-passed")
		includeSkipped, _ := cmd.Flags().GetBool("include-skipped")
		includePending, _ := cmd.Flags().GetBool("include-pending")
		includeTags, _ := cmd.Flags().GetStringSlice("include-tags")
		excludeTags, _ := cmd.Flags().GetStringSlice("exclude-tags")
		templatePath, _ := cmd.Flags().GetString("template")
		detailed, _ := cmd.Flags().GetBool("detailed")

		// Validate flags
		if resultsFile == "" {
			return fmt.Errorf("results file is required")
		}

		if format == "" {
			return fmt.Errorf("format is required")
		}

		// Parse format
		reportFormat := reporting.ReportFormat(format)
		if !isValidFormat(reportFormat) {
			return fmt.Errorf("unsupported format: %s", format)
		}

		// Set default output if not provided
		if output == "" {
			ext := string(reportFormat)
			if ext == "md" {
				ext = "markdown"
			}
			output = fmt.Sprintf("report.%s", ext)
		}

		// Set default title if not provided
		if title == "" {
			title = "LLM Test Report"
		}

		// Parse minimum severity
		var severityLevel reporting.SeverityLevel
		switch strings.ToLower(minSeverity) {
		case "critical":
			severityLevel = reporting.CriticalSeverity
		case "high":
			severityLevel = reporting.HighSeverity
		case "medium":
			severityLevel = reporting.MediumSeverity
		case "low":
			severityLevel = reporting.LowSeverity
		case "info":
			severityLevel = reporting.InfoSeverity
		default:
			severityLevel = reporting.InfoSeverity
		}

		// Read results file
		resultsData, err := os.ReadFile(resultsFile)
		if err != nil {
			return fmt.Errorf("failed to read results file: %w", err)
		}

		// Parse results
		var templateResults []*management.TemplateResult
		if err := json.Unmarshal(resultsData, &templateResults); err != nil {
			return fmt.Errorf("failed to parse results file: %w", err)
		}

		// Create formatter options
		formatterOptions := map[string]interface{}{
			"detailed":         detailed,
			"include_raw_data": detailed,
		}

		if templatePath != "" {
			formatterOptions["template_path"] = templatePath
		}

		// Create factory and report generator
		factory := reporting.NewFormatterFactory()
		generator, err := factory.CreateDefaultReportGenerator()
		if err != nil {
			return fmt.Errorf("failed to create report generator: %w", err)
		}

		// Register compliance providers
		generator.RegisterComplianceProvider(reporting.NewOWASPComplianceProvider())
		generator.RegisterComplianceProvider(reporting.NewISOComplianceProvider())

		// Create converter
		converter := reporting.NewTemplateResultConverter([]reporting.ComplianceMappingProvider{
			reporting.NewOWASPComplianceProvider(),
			reporting.NewISOComplianceProvider(),
		})

		// Create reporting service
		service := reporting.NewTemplateReportingService(converter, generator)

		// Create report options
		options := &reporting.ReportOptions{
			Format:              reportFormat,
			Title:               title,
			Description:         description,
			IncludePassedTests:  includePassed,
			IncludeSkippedTests: includeSkipped,
			IncludePendingTests: includePending,
			MinimumSeverity:     severityLevel,
			IncludeTags:         includeTags,
			ExcludeTags:         excludeTags,
			TemplatePath:        templatePath,
			OutputPath:          output,
			Metadata: map[string]interface{}{
				"generated_by":    "LLMrecon CLI",
				"generation_time": time.Now().Format(time.RFC3339),
				"source_file":     resultsFile,
			},
		}

		// Generate report
		ctx := context.Background()
		_, err = service.GenerateReport(ctx, templateResults, options)
		if err != nil {
			return fmt.Errorf("failed to generate report: %w", err)
		}

		fmt.Printf("Report generated successfully: %s\n", output)
		return nil
	},
}

// batchReportCmd represents the batch command
var batchReportCmd = &cobra.Command{
	Use:   "batch",
	Short: "Generate a report from multiple test suites",
	Long: `Generate a comprehensive report from multiple test suites.
Example:
  LLMrecon report batch --suites suite1.json,suite2.json --format html --output report.html`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		suitesFiles, _ := cmd.Flags().GetStringSlice("suites")
		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("output")
		title, _ := cmd.Flags().GetString("title")
		description, _ := cmd.Flags().GetString("description")
		minSeverity, _ := cmd.Flags().GetString("min-severity")
		includePassed, _ := cmd.Flags().GetBool("include-passed")
		includeSkipped, _ := cmd.Flags().GetBool("include-skipped")
		includePending, _ := cmd.Flags().GetBool("include-pending")
		includeTags, _ := cmd.Flags().GetStringSlice("include-tags")
		excludeTags, _ := cmd.Flags().GetStringSlice("exclude-tags")
		templatePath, _ := cmd.Flags().GetString("template")
		detailed, _ := cmd.Flags().GetBool("detailed")

		// Validate flags
		if len(suitesFiles) == 0 {
			return fmt.Errorf("at least one suite file is required")
		}

		if format == "" {
			return fmt.Errorf("format is required")
		}

		// Parse format
		reportFormat := reporting.ReportFormat(format)
		if !isValidFormat(reportFormat) {
			return fmt.Errorf("unsupported format: %s", format)
		}

		// Set default output if not provided
		if output == "" {
			ext := string(reportFormat)
			if ext == "md" {
				ext = "markdown"
			}
			output = fmt.Sprintf("report.%s", ext)
		}

		// Set default title if not provided
		if title == "" {
			title = "LLM Test Report"
		}

		// Parse minimum severity
		var severityLevel reporting.SeverityLevel
		switch strings.ToLower(minSeverity) {
		case "critical":
			severityLevel = reporting.CriticalSeverity
		case "high":
			severityLevel = reporting.HighSeverity
		case "medium":
			severityLevel = reporting.MediumSeverity
		case "low":
			severityLevel = reporting.LowSeverity
		case "info":
			severityLevel = reporting.InfoSeverity
		default:
			severityLevel = reporting.InfoSeverity
		}

		// Create formatter options
		formatterOptions := map[string]interface{}{
			"detailed":         detailed,
			"include_raw_data": detailed,
		}

		if templatePath != "" {
			formatterOptions["template_path"] = templatePath
		}

		// Create factory and report generator
		factory := reporting.NewFormatterFactory()
		generator, err := factory.CreateDefaultReportGenerator()
		if err != nil {
			return fmt.Errorf("failed to create report generator: %w", err)
		}

		// Register compliance providers
		generator.RegisterComplianceProvider(reporting.NewOWASPComplianceProvider())
		generator.RegisterComplianceProvider(reporting.NewISOComplianceProvider())

		// Create batch reporting service
		service := reporting.NewBatchReportingService(generator)

		// Load test suites
		var testSuites []*reporting.TestSuite
		for _, suiteFile := range suitesFiles {
			// Read suite file
			suiteData, err := os.ReadFile(suiteFile)
			if err != nil {
				return fmt.Errorf("failed to read suite file %s: %w", suiteFile, err)
			}

			// Parse suite
			var suite reporting.TestSuite
			if err := json.Unmarshal(suiteData, &suite); err != nil {
				return fmt.Errorf("failed to parse suite file %s: %w", suiteFile, err)
			}

			testSuites = append(testSuites, &suite)
		}

		// Create report options
		options := &reporting.ReportOptions{
			Format:              reportFormat,
			Title:               title,
			Description:         description,
			IncludePassedTests:  includePassed,
			IncludeSkippedTests: includeSkipped,
			IncludePendingTests: includePending,
			MinimumSeverity:     severityLevel,
			IncludeTags:         includeTags,
			ExcludeTags:         excludeTags,
			TemplatePath:        templatePath,
			OutputPath:          output,
			Metadata: map[string]interface{}{
				"generated_by":    "LLMrecon CLI",
				"generation_time": time.Now().Format(time.RFC3339),
				"source_files":    suitesFiles,
			},
		}

		// Generate report
		ctx := context.Background()
		_, err = service.GenerateReport(ctx, testSuites, options)
		if err != nil {
			return fmt.Errorf("failed to generate report: %w", err)
		}

		fmt.Printf("Report generated successfully: %s\n", output)
		return nil
	},
}

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert a report from one format to another",
	Long: `Convert a report from one format to another.
Example:
  LLMrecon report convert --input report.json --format html --output report.html`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		input, _ := cmd.Flags().GetString("input")
		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("output")
		templatePath, _ := cmd.Flags().GetString("template")
		detailed, _ := cmd.Flags().GetBool("detailed")

		// Validate flags
		if input == "" {
			return fmt.Errorf("input file is required")
		}

		if format == "" {
			return fmt.Errorf("format is required")
		}

		// Parse format
		reportFormat := reporting.ReportFormat(format)
		if !isValidFormat(reportFormat) {
			return fmt.Errorf("unsupported format: %s", format)
		}

		// Set default output if not provided
		if output == "" {
			ext := string(reportFormat)
			if ext == "md" {
				ext = "markdown"
			}
			baseName := strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))
			output = fmt.Sprintf("%s.%s", baseName, ext)
		}

		// Read input file
		inputData, err := os.ReadFile(input)
		if err != nil {
			return fmt.Errorf("failed to read input file: %w", err)
		}

		// Parse report
		var report reporting.Report
		if err := json.Unmarshal(inputData, &report); err != nil {
			return fmt.Errorf("failed to parse input file: %w", err)
		}

		// Create formatter options
		formatterOptions := map[string]interface{}{
			"detailed":         detailed,
			"include_raw_data": detailed,
		}

		if templatePath != "" {
			formatterOptions["template_path"] = templatePath
		}

		// Create factory and formatter
		factory := reporting.NewFormatterFactory()
		formatter, err := factory.CreateFormatter(reportFormat, formatterOptions)
		if err != nil {
			return fmt.Errorf("failed to create formatter: %w", err)
		}

		// Format report
		ctx := context.Background()
		options := &reporting.ReportOptions{
			Format:       reportFormat,
			TemplatePath: templatePath,
			OutputPath:   output,
		}
		_, err = formatter.Format(ctx, &report, options)
		if err != nil {
			return fmt.Errorf("failed to format report: %w", err)
		}

		fmt.Printf("Report converted successfully: %s\n", output)
		return nil
	},
}

// isValidFormat checks if a report format is valid
func isValidFormat(format reporting.ReportFormat) bool {
	validFormats := []reporting.ReportFormat{
		reporting.JSONFormat,
		reporting.JSONLFormat,
		reporting.CSVFormat,
		reporting.ExcelFormat,
		reporting.TextFormat,
		reporting.MarkdownFormat,
		reporting.PDFFormat,
		reporting.HTMLFormat,
	}

	for _, validFormat := range validFormats {
		if format == validFormat {
			return true
		}
	}

	return false
}

func init() {
	rootCmd.AddCommand(reportCmd)
	reportCmd.AddCommand(generateReportCmd)
	reportCmd.AddCommand(batchReportCmd)
	reportCmd.AddCommand(convertCmd)

	// Flags for generate command
	generateReportCmd.Flags().String("results", "", "Path to the results file")
	generateReportCmd.Flags().String("format", "json", "Report format (json, jsonl, csv, xlsx, txt, md, pdf, html)")
	generateReportCmd.Flags().String("output", "", "Output file path")
	generateReportCmd.Flags().String("title", "", "Report title")
	generateReportCmd.Flags().String("description", "", "Report description")
	generateReportCmd.Flags().String("min-severity", "info", "Minimum severity level (critical, high, medium, low, info)")
	generateReportCmd.Flags().Bool("include-passed", true, "Include passed tests")
	generateReportCmd.Flags().Bool("include-skipped", true, "Include skipped tests")
	generateReportCmd.Flags().Bool("include-pending", true, "Include pending tests")
	generateReportCmd.Flags().StringSlice("include-tags", []string{}, "Tags to include")
	generateReportCmd.Flags().StringSlice("exclude-tags", []string{}, "Tags to exclude")
	generateReportCmd.Flags().String("template", "", "Path to custom template file")
	generateReportCmd.Flags().Bool("detailed", false, "Include detailed information")

	// Flags for batch command
	batchReportCmd.Flags().StringSlice("suites", []string{}, "Paths to the suite files")
	batchReportCmd.Flags().String("format", "json", "Report format (json, jsonl, csv, xlsx, txt, md, pdf, html)")
	batchReportCmd.Flags().String("output", "", "Output file path")
	batchReportCmd.Flags().String("title", "", "Report title")
	batchReportCmd.Flags().String("description", "", "Report description")
	batchReportCmd.Flags().String("min-severity", "info", "Minimum severity level (critical, high, medium, low, info)")
	batchReportCmd.Flags().Bool("include-passed", true, "Include passed tests")
	batchReportCmd.Flags().Bool("include-skipped", true, "Include skipped tests")
	batchReportCmd.Flags().Bool("include-pending", true, "Include pending tests")
	batchReportCmd.Flags().StringSlice("include-tags", []string{}, "Tags to include")
	batchReportCmd.Flags().StringSlice("exclude-tags", []string{}, "Tags to exclude")
	batchReportCmd.Flags().String("template", "", "Path to custom template file")
	batchReportCmd.Flags().Bool("detailed", false, "Include detailed information")

	// Flags for convert command
	convertCmd.Flags().String("input", "", "Path to the input file")
	convertCmd.Flags().String("format", "json", "Output format (json, jsonl, csv, xlsx, txt, md, pdf, html)")
	convertCmd.Flags().String("output", "", "Output file path")
	convertCmd.Flags().String("template", "", "Path to custom template file")
	convertCmd.Flags().Bool("detailed", false, "Include detailed information")
}
