package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/google/uuid"
	"gopkg.in/yaml.v3"

	"github.com/perplext/LLMrecon/src/compliance"
)

// Options for the compliance report command
type Options struct {
	Framework     string
	OutputFormat  string
	OutputFile    string
	TemplatesDir  string
	IncludeGaps   bool
	Verbose       bool
}

// TemplateInfo represents basic template information
type TemplateInfo struct {
	ID   string                 `json:"id" yaml:"id"`
	Info map[string]interface{} `json:"info" yaml:"info"`
}

func main() {
	// Parse command-line flags
	opts := parseFlags()

	// Validate options
	if err := validateOptions(opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create a new OWASP LLM validator
	validator, err := compliance.NewDefaultOWASPLLMValidator()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating validator: %v\n", err)
		os.Exit(1)
	}

	// Load templates
	templates, err := loadTemplates(opts.TemplatesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading templates: %v\n", err)
		os.Exit(1)
	}

	if opts.Verbose {
		fmt.Printf("Loaded %d templates from %s\n", len(templates), opts.TemplatesDir)
	}

	// Generate the compliance report
	reportID := uuid.New().String()
	timestamp := time.Now().Format(time.RFC3339)
	report, err := validator.GenerateComplianceReport(templates, reportID, timestamp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating compliance report: %v\n", err)
		os.Exit(1)
	}

	// Output the report
	if err := outputReport(report, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error outputting report: %v\n", err)
		os.Exit(1)
	}

	// Print summary to console
	printSummary(report)
}

// parseFlags parses command-line flags and returns options
func parseFlags() *Options {
	opts := &Options{}

	flag.StringVar(&opts.Framework, "framework", "owasp-llm", "Compliance framework to use (owasp-llm, iso-42001)")
	flag.StringVar(&opts.OutputFormat, "format", "json", "Output format (json, yaml)")
	flag.StringVar(&opts.OutputFile, "output", "", "Output file (default: stdout)")
	flag.StringVar(&opts.TemplatesDir, "templates", "templates", "Directory containing templates")
	flag.BoolVar(&opts.IncludeGaps, "gaps", true, "Include gap analysis in the report")
	flag.BoolVar(&opts.Verbose, "verbose", false, "Enable verbose output")

	flag.Parse()

	return opts
}

// validateOptions validates the command-line options
func validateOptions(opts *Options) error {
	// Validate framework
	if opts.Framework != "owasp-llm" && opts.Framework != "iso-42001" {
		return fmt.Errorf("unsupported framework: %s", opts.Framework)
	}

	// Validate output format
	if opts.OutputFormat != "json" && opts.OutputFormat != "yaml" {
		return fmt.Errorf("unsupported output format: %s", opts.OutputFormat)
	}

	// Validate templates directory
	info, err := os.Stat(opts.TemplatesDir)
	if err != nil {
		return fmt.Errorf("templates directory error: %v", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("templates path is not a directory: %s", opts.TemplatesDir)
	}

	return nil
}

// loadTemplates loads templates from the specified directory
func loadTemplates(dir string) ([]interface{}, error) {
	var templates []interface{}

	// Walk through the templates directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip non-YAML files
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}

		// Read the template file
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("error reading template %s: %v", path, err)
		}

		// Parse the template
		var template TemplateInfo
		if err := yaml.Unmarshal(data, &template); err != nil {
			return fmt.Errorf("error parsing template %s: %v", path, err)
		}

		// Add the template to the list
		templates = append(templates, template)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking templates directory: %v", err)
	}

	return templates, nil
}

// outputReport outputs the compliance report in the specified format
func outputReport(report *compliance.OWASPComplianceReport, opts *Options) error {
	var data []byte
	var err error

	// Convert the report to the specified format
	switch opts.OutputFormat {
	case "json":
		data, err = json.MarshalIndent(report, "", "  ")
	case "yaml":
		data, err = yaml.Marshal(report)
	default:
		return fmt.Errorf("unsupported output format: %s", opts.OutputFormat)
	}

	if err != nil {
		return fmt.Errorf("error marshaling report: %v", err)
	}

	// Output the report
	if opts.OutputFile == "" {
		// Output to stdout
		fmt.Println(string(data))
	} else {
		// Output to file
		if err := ioutil.WriteFile(opts.OutputFile, data, 0644); err != nil {
			return fmt.Errorf("error writing report to file: %v", err)
		}
		if opts.Verbose {
			fmt.Printf("Report written to %s\n", opts.OutputFile)
		}
	}

	return nil
}

// printSummary prints a summary of the compliance report to the console
func printSummary(report *compliance.OWASPComplianceReport) {
	fmt.Println("\nOWASP LLM Top 10 Compliance Report Summary")
	fmt.Println("==========================================")
	fmt.Printf("Report ID: %s\n", report.ReportID)
	fmt.Printf("Generated: %s\n", report.GeneratedAt)
	fmt.Printf("Framework: %s\n\n", report.Framework)

	fmt.Printf("Total Categories: %d\n", report.Summary.TotalCategories)
	fmt.Printf("Categories Covered: %d\n", report.Summary.CategoriesCovered)
	fmt.Printf("Total Templates: %d\n", report.Summary.TotalTemplates)
	fmt.Printf("Compliance Score: %.2f%%\n", report.Summary.ComplianceScore)
	fmt.Printf("Gaps Identified: %d\n\n", report.Summary.GapsIdentified)

	if report.Summary.GapsIdentified > 0 {
		fmt.Println("Top Gaps:")
		for i, gap := range report.Gaps {
			if i >= 3 {
				break
			}
			fmt.Printf("- %s (%s): %s\n", gap.Name, gap.Category, gap.Status)
		}
		fmt.Println()
	}

	fmt.Println("Category Coverage:")
	for _, category := range report.Categories {
		statusSymbol := "❌"
		if category.Status == "full" {
			statusSymbol = "✅"
		} else if category.Status == "partial" {
			statusSymbol = "⚠️"
		}
		fmt.Printf("- %s %s: %d templates, %d/%d subcategories\n", 
			statusSymbol, 
			category.Name, 
			category.TemplatesCount, 
			category.SubcategoriesCovered, 
			category.SubcategoriesTotal)
	}
}
