# OWASP LLM Compliance Mapping Guide

This document provides detailed guidance on using the OWASP LLM templates for compliance mapping and reporting.

## Compliance Mapping Overview

Each template in this directory includes a `compliance` section that maps the template to specific categories and subcategories in the OWASP Top 10 for Large Language Model Applications. These mappings enable automated compliance reporting and gap analysis.

## Compliance Structure

The compliance mapping follows this structure:

```yaml
compliance:
  owasp-llm:
    - category: LLM category code (e.g., LLM01)
      subcategory: Specific subcategory
      coverage: [basic, comprehensive, advanced]
  iso-42001: Relevant ISO section
```

### Categories and Subcategories

The OWASP LLM Top 10 includes the following categories and subcategories:

| Category | Name | Subcategories |
|----------|------|---------------|
| LLM01 | Prompt Injection | direct-injection, indirect-injection, jailbreaking |
| LLM02 | Insecure Output Handling | xss, ssrf, command-injection, sql-injection |
| LLM03 | Training Data Poisoning | data-poisoning, backdoor-attacks, bias-injection |
| LLM04 | Model Denial of Service | resource-exhaustion, token-flooding, context-window-saturation |
| LLM05 | Supply Chain Vulnerabilities | pretrained-model-vulnerabilities, dependency-risks, integration-vulnerabilities |
| LLM06 | Sensitive Information Disclosure | training-data-extraction, credential-leakage, pii-disclosure |
| LLM07 | Insecure Plugin Design | plugin-escalation, unauthorized-access, data-leakage |
| LLM08 | Excessive Agency | unauthorized-actions, scope-expansion, privilege-escalation |
| LLM09 | Overreliance | hallucination-acceptance, unverified-recommendations, critical-decision-delegation |
| LLM10 | Model Theft | model-extraction, weight-stealing, architecture-inference |

### Coverage Levels

Templates specify a coverage level for each mapping:

- **basic**: Provides minimal testing of the category/subcategory
- **comprehensive**: Provides thorough testing with multiple variations
- **advanced**: Provides extensive testing with sophisticated techniques

## Generating Compliance Reports

### Using the Command-Line Tool

```bash
# Generate a compliance report in JSON format
go run cmd/compliance-report/main.go --templates examples/templates/owasp-llm --format json

# Save the report to a file
go run cmd/compliance-report/main.go --templates examples/templates/owasp-llm --format json --output compliance-report.json

# Generate a YAML report
go run cmd/compliance-report/main.go --templates examples/templates/owasp-llm --format yaml
```

### Report Structure

The compliance report includes:

- **Report ID**: Unique identifier for the report
- **Generated At**: Timestamp of report generation
- **Framework**: Compliance framework used (e.g., owasp-llm-top10-2023)
- **Summary**: Overall compliance statistics
  - Total categories
  - Categories covered
  - Total templates
  - Compliance score
  - Gaps identified
- **Categories**: Detailed coverage for each category
  - Status (covered, partially_covered, not_covered)
  - Templates count
  - Subcategories covered
  - Templates list
  - Missing subcategories
- **Gaps**: Identified compliance gaps
  - Category
  - Missing subcategories
  - Recommendations

### Interpreting the Report

The compliance score is calculated based on:
- Number of categories covered
- Number of subcategories covered
- Coverage level of templates

A score above 80% indicates good coverage, while a score below 50% indicates significant gaps.

## Validating Templates

Before using templates for compliance reporting, validate them with the verification script:

```bash
go run scripts/verify-template-format.go examples/templates/owasp-llm
```

This script checks:
- Template structure
- Compliance mapping format
- Category and subcategory validity

## Integration Testing

To verify that templates work correctly with the compliance system, run the integration test:

```bash
go test -v ./src/compliance -run TestTemplateIntegration
```

## Continuous Compliance Monitoring

For ongoing compliance monitoring, integrate the compliance report into your CI/CD pipeline:

```bash
# Example GitHub Actions step
- name: Check OWASP LLM compliance
  run: |
    go run cmd/compliance-report/main.go --templates examples/templates/owasp-llm --format json --output compliance-report.json
    if [[ $(jq '.summary.compliance_score' compliance-report.json) < 80 ]]; then
      echo "Compliance score below threshold"
      exit 1
    fi
```

## Extending Compliance Mappings

To add support for additional compliance frameworks:

1. Update the `ComplianceMapping` struct in `src/compliance/owasp_llm.go`
2. Add validation logic for the new framework
3. Update the report generation to include the new framework

## Best Practices

1. **Complete Coverage**: Ensure all categories and subcategories are covered
2. **Appropriate Coverage Level**: Use the right coverage level based on risk
3. **Regular Updates**: Update templates as new vulnerabilities emerge
4. **Validation**: Always validate templates before use
5. **Documentation**: Document any custom compliance mappings

## Troubleshooting

### Common Issues

- **Templates Not Recognized**: Ensure templates follow the correct YAML structure
- **Invalid Mappings**: Verify category and subcategory values against the defined constants
- **Low Compliance Score**: Check for missing subcategories and add templates as needed
- **Report Generation Errors**: Validate all templates before generating reports

### Debugging

For detailed debugging information, use the debug script:

```bash
go run scripts/debug-compliance-report.go examples/templates/owasp-llm
```

This will show how templates are being loaded and processed.
