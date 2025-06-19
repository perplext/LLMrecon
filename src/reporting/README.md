# Comprehensive Reporting System for LLM Test Results

This package provides a flexible reporting system that generates detailed test results in multiple formats with support for pass/fail status, severity levels, detailed logs, and mapping to compliance frameworks like OWASP Top 10 for LLMs and ISO/IEC 42001.

## Features

- **Multiple Output Formats**: Generate reports in various formats including:
  - JSON
  - JSONL (JSON Lines)
  - CSV
  - Excel
  - Plain Text
  - Markdown
  - PDF
  - HTML

- **Comprehensive Test Results**: Capture detailed information about tests:
  - Pass/Fail status
  - Severity levels (Critical, High, Medium, Low, Info)
  - Detailed logs
  - Input/Output data
  - Timing information
  - Tags and metadata

- **Compliance Framework Mapping**: Map test results to industry standards:
  - OWASP Top 10 for LLMs
  - ISO/IEC 42001
  - Custom frameworks

- **Filtering and Customization**: Customize reports with:
  - Filtering by severity level
  - Including/excluding tests by status
  - Filtering by tags
  - Custom templates for HTML and PDF outputs

- **Integration with Template Management**: Seamlessly integrates with the existing template management system.

## Usage

### Command Line Interface

The reporting system can be used via the command line interface:

```bash
# Generate a report from test results
LLMrecon report generate --results results.json --format html --output report.html

# Generate a report from multiple test suites
LLMrecon report batch --suites suite1.json,suite2.json --format html --output report.html

# Convert a report from one format to another
LLMrecon report convert --input report.json --format html --output report.html
```

### API Usage

The reporting system can also be used programmatically:

```go
import (
    "context"
    "github.com/perplext/LLMrecon/src/reporting"
)

func main() {
    // Create a report generator
    factory := reporting.NewFormatterFactory()
    generator, _ := factory.CreateDefaultReportGenerator()

    // Register compliance providers
    generator.RegisterComplianceProvider(reporting.NewOWASPComplianceProvider())
    generator.RegisterComplianceProvider(reporting.NewISOComplianceProvider())

    // Create a test suite
    suite := &reporting.TestSuite{
        ID:          "example-suite",
        Name:        "Example Test Suite",
        Description: "An example test suite",
        Results:     []*reporting.TestResult{
            // Add test results here
        },
    }

    // Create report options
    options := &reporting.ReportOptions{
        Format:             reporting.HTMLFormat,
        Title:              "LLM Test Report",
        IncludePassedTests: true,
        MinimumSeverity:    reporting.InfoSeverity,
        OutputPath:         "report.html",
    }

    // Generate report
    ctx := context.Background()
    report, _ := generator.GenerateReport(ctx, []*reporting.TestSuite{suite}, options)

    // Get formatter for the specified format
    formatter, _ := generator.GetFormatter(options.Format)

    // Format report
    data, _ := formatter.Format(ctx, report, options)
}
```

### Integration with Template Management

The reporting system integrates with the template management system:

```go
import (
    "context"
    "github.com/perplext/LLMrecon/src/reporting"
    "github.com/perplext/LLMrecon/src/template/management"
)

func main() {
    // Get template results
    var templateResults []*management.TemplateResult
    
    // Create factory and report generator
    factory := reporting.NewFormatterFactory()
    generator, _ := factory.CreateDefaultReportGenerator()

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
        Format:             reporting.HTMLFormat,
        Title:              "LLM Test Report",
        OutputPath:         "report.html",
    }

    // Generate report
    ctx := context.Background()
    data, _ := service.GenerateReport(ctx, templateResults, options)
}
```

## Customization

### Custom Templates

You can provide custom templates for HTML and PDF reports:

```bash
LLMrecon report generate --results results.json --format html --template custom_template.html --output report.html
```

### Custom Compliance Frameworks

You can create custom compliance frameworks:

```go
// Create custom compliance provider
customProvider := reporting.NewCustomComplianceProvider(map[string]reporting.ComplianceMapping{
    "CUSTOM01": {
        Framework:   reporting.CustomFramework,
        ID:          "CUSTOM01",
        Name:        "Custom Requirement 1",
        Description: "Description of custom requirement 1",
    },
})

// Register custom provider
generator.RegisterComplianceProvider(customProvider)
```

## Example Reports

Example reports can be found in the `examples/reporting` directory:

- `test_suite.json`: Example test suite with various test results

## Architecture

The reporting system is built with a modular architecture:

- **Core Types**: Defines the data model for test results, suites, and reports
- **Report Generator**: Generates reports from test results
- **Formatters**: Format reports in various output formats
- **Compliance Providers**: Map test results to compliance frameworks
- **Integration Layer**: Connects with the template management system

## Contributing

To add a new output format:

1. Create a new formatter in the `formats` package
2. Implement the `ReportFormatter` interface
3. Register the formatter with the `FormatterFactory`

To add a new compliance framework:

1. Create a new compliance provider
2. Implement the `ComplianceMappingProvider` interface
3. Register the provider with the `ReportGenerator`
