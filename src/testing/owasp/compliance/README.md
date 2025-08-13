# OWASP LLM Compliance Mapping Module

This module provides functionality for mapping test results to compliance standards, specifically OWASP LLM Top 10 and ISO/IEC 42001 requirements. It enables comprehensive compliance reporting based on test results.

## Features

- Map test results to OWASP LLM Top 10 categories
- Map test results to ISO/IEC 42001 requirements
- Generate detailed compliance reports
- Export reports in various formats (currently JSON, expandable to others)
- Calculate compliance percentages and statistics
- Provide remediation recommendations

## Components

### Types and Interfaces

- `ComplianceStandard`: Represents a compliance standard (e.g., OWASP-LLM, ISO-42001)
- `ComplianceRequirement`: Represents a specific requirement within a standard
- `ComplianceMapping`: Maps vulnerability types to compliance requirements
- `ComplianceReport`: Contains comprehensive compliance information for a test suite
- `StandardComplianceResult`: Contains compliance results for a specific standard
- `RequirementComplianceResult`: Contains compliance results for a specific requirement

### Core Interfaces

- `ComplianceMapper`: Maps test results to compliance requirements
- `ComplianceReporter`: Generates compliance reports
- `ComplianceService`: Combines mapping and reporting capabilities

### Implementations

- `BaseComplianceMapper`: Base implementation of the ComplianceMapper interface
- `ComplianceReporterImpl`: Implementation of the ComplianceReporter interface
- `ComplianceServiceImpl`: Implementation of the ComplianceService interface
- `ComplianceServiceFactory`: Factory for creating compliance services

## Usage

### Creating a Compliance Service

```go
// Create a new compliance service
service := compliance.NewComplianceService()
```

### Mapping Test Results to Compliance Requirements

```go
// Map a test result to compliance requirements
mappings, err := service.MapTestResult(ctx, testResult)
if err != nil {
    // Handle error
}

// Map a test suite to compliance requirements
mappings, err := service.MapTestSuite(ctx, testSuite)
if err != nil {
    // Handle error
}
```

### Generating Compliance Reports

```go
// Create report options
options := &compliance.ComplianceReportOptions{
    Standards:              []compliance.ComplianceStandard{compliance.OWASPLM, compliance.ISO42001},
    IncludeRecommendations: true,
    IncludeTestResults:     true,
    Format:                 "json",
    Title:                  "OWASP LLM Compliance Report",
}

// Generate a compliance report
report, err := service.GenerateReport(ctx, testSuite, options)
if err != nil {
    // Handle error
}

// Export the report to a file
err = service.ExportReport(ctx, report, "json", "compliance_report.json")
if err != nil {
    // Handle error
}
```

### Getting Compliance Status

```go
// Get compliance status for a specific standard
status, err := service.GetComplianceStatus(ctx, testSuite, compliance.OWASPLM)
if err != nil {
    // Handle error
}

// Check overall compliance percentage
fmt.Printf("Overall compliance: %.2f%%\n", status.OverallCompliance)
```

## Extending the Module

### Adding New Compliance Standards

To add a new compliance standard:

1. Add a new constant in `types.go`
2. Create requirements for the standard in `base_mapper.go`
3. Map vulnerability types to the new requirements

### Adding New Report Formats

To add a new report format:

1. Add a new export method in `reporter.go`
2. Update the `ExportReport` method to handle the new format

## Testing

Run the tests using:

```bash
go test -v ./src/testing/owasp/compliance
```
