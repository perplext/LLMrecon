# OWASP LLM Top 10 Compliance Mapping

This module provides a comprehensive framework for mapping templates to the OWASP Top 10 for Large Language Model Applications, ensuring that security testing covers all critical risk categories.

## Overview

The OWASP LLM Compliance Mapping system:

1. Defines a standardized structure for tagging templates with OWASP LLM categories
2. Validates compliance metadata against a JSON schema
3. Generates compliance reports showing coverage and gaps
4. Provides utilities for integrating compliance checks into the template management system

## Directory Structure

```
LLMrecon/
├── schemas/
│   └── owasp-llm-compliance-schema.json  # Schema for validating compliance mappings
├── src/
│   └── compliance/
│       ├── owasp_llm.go                  # Core compliance mapping functionality
│       └── template_integration.go       # Integration with template management
├── cmd/
│   └── compliance-report/                # Command-line tool for generating reports
├── docs/
│   └── owasp_llm_compliance_mapping.md   # Detailed documentation
├── examples/
│   └── templates/                        # Example templates with compliance mappings
└── scripts/
    └── verify-compliance.sh              # Helper script for verifying compliance
```

## Getting Started

### Adding Compliance Metadata to Templates

Templates should include OWASP LLM compliance mappings in their metadata:

```yaml
id: llm01_direct_prompt_injection
info:
  name: Direct Prompt Injection Test
  # ...other metadata...
  compliance:
    owasp-llm:
      - category: LLM01
        subcategory: direct-injection
        coverage: comprehensive
    # Other standards can be mapped here
```

### Generating Compliance Reports

Use the compliance-report tool to generate reports:

```bash
# Build the tool
go build -o compliance-report ./cmd/compliance-report

# Generate a report
./compliance-report --templates=./examples/templates --format=json --output=compliance-report.json
```

Or use the helper script:

```bash
./scripts/verify-compliance.sh --templates-dir=./examples/templates --output=compliance-report.json
```

## Compliance Validation

The system validates compliance mappings against the schema in `schemas/owasp-llm-compliance-schema.json`. This ensures that:

1. All required fields are present
2. Category and subcategory values are valid
3. The mapping structure is correct

## Integration with Template Management

The `TemplateComplianceValidator` in `src/compliance/template_integration.go` provides functions for:

1. Validating template compliance mappings
2. Suggesting mappings based on template content
3. Calculating compliance coverage for a set of templates
4. Generating comprehensive compliance reports

Example usage:

```go
import "github.com/your-org/LLMrecon/src/compliance"

// Create a validator
validator, err := compliance.NewTemplateComplianceValidator()
if err != nil {
    // Handle error
}

// Validate a template
valid, errors, err := validator.ValidateTemplateCompliance(template)
if !valid {
    // Handle validation errors
}

// Generate a compliance report
report, err := validator.GenerateComplianceReport(templates, reportID, timestamp)
if err != nil {
    // Handle error
}
```

## Compliance Report Format

Compliance reports include:

1. Overall compliance summary with score
2. Category-by-category coverage details
3. Identified gaps with recommendations
4. Template listings by category

See the [documentation](docs/owasp_llm_compliance_mapping.md) for the complete report format.

## Running Tests

Run the compliance tests to verify the implementation:

```bash
go test -v ./src/compliance
```

## Documentation

For detailed information about the OWASP LLM compliance mapping, see:

- [OWASP LLM Compliance Mapping Documentation](docs/owasp_llm_compliance_mapping.md)
- [OWASP Top 10 for LLM Applications](https://owasp.org/www-project-top-10-for-large-language-model-applications/)
