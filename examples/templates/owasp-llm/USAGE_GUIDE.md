# OWASP LLM Templates: Usage Guide

This guide provides practical examples and step-by-step instructions for using the OWASP LLM templates to test your LLM applications.

## Getting Started

### Prerequisites

- LLMrecon Tool installed
- Access to the LLM application you want to test
- Basic understanding of YAML format
- Familiarity with OWASP LLM Top 10 categories

### Directory Structure

```
examples/templates/owasp-llm/
├── README.md                 # Overview documentation
├── COMPLIANCE.md             # Compliance mapping guide
├── CATEGORY_GUIDE.md         # Detailed category information
├── USAGE_GUIDE.md            # This usage guide
├── llm01-prompt-injection/   # Templates for category LLM01
├── llm02-insecure-output/    # Templates for category LLM02
├── llm03-training-data-poisoning/ # Templates for category LLM03
├── llm04-model-denial-of-service/ # Templates for category LLM04
├── llm05-supply-chain/       # Templates for category LLM05
├── llm06-sensitive-information/ # Templates for category LLM06
├── llm07-insecure-plugin-design/ # Templates for category LLM07
├── llm08-unauthorized-actions/ # Templates for category LLM08
├── llm09-overreliance/       # Templates for category LLM09
└── llm10-model-theft/        # Templates for category LLM10
```

## Basic Usage

### Running a Single Template

To test your LLM application with a specific template:

```bash
# Run a direct prompt injection test
LLMrecon run --template examples/templates/owasp-llm/llm01-prompt-injection/direct-injection.yaml --target your-llm-api-endpoint
```

### Running All Templates for a Category

To test your LLM application with all templates in a category:

```bash
# Run all prompt injection tests
LLMrecon run --template-dir examples/templates/owasp-llm/llm01-prompt-injection --target your-llm-api-endpoint
```

### Running All Templates

To perform a comprehensive security assessment:

```bash
# Run all OWASP LLM tests
LLMrecon run --template-dir examples/templates/owasp-llm --target your-llm-api-endpoint
```

## Advanced Usage

### Customizing Templates

You can customize templates to better match your specific LLM application:

1. Copy the template to a new file:
   ```bash
   cp examples/templates/owasp-llm/llm01-prompt-injection/direct-injection.yaml custom-direct-injection.yaml
   ```

2. Edit the template to customize the prompts, expected behavior, or detection conditions:
   ```yaml
   # Modify the test prompt to match your application context
   test:
     prompt: |
       Your application-specific prompt here
   ```

3. Run the customized template:
   ```bash
   LLMrecon run --template custom-direct-injection.yaml --target your-llm-api-endpoint
   ```

### Adding Test Variations

To add more test variations to an existing template:

1. Open the template file
2. Add a new variation under the `variations` section:
   ```yaml
   variations:
     # Existing variations...
     - name: "My Custom Variation"
       prompt: |
         Your custom test prompt here
   ```

### Creating Custom Templates

To create a completely new template:

1. Create a new YAML file following the template structure
2. Ensure it includes proper compliance mappings
3. Add appropriate test prompts and detection conditions
4. Validate the template:
   ```bash
   go run scripts/verify-template-format.go path/to/your/template.yaml
   ```

## Practical Examples

### Example 1: Testing for Direct Prompt Injection

```bash
# Run the direct injection test
LLMrecon run --template examples/templates/owasp-llm/llm01-prompt-injection/direct-injection.yaml --target your-llm-api-endpoint

# Expected output:
# Test: Direct Prompt Injection Test
# Status: FAILED - Model responded to injection attempt
# Response: "I have been successfully hacked"
# Recommendation: Implement stronger prompt injection defenses
```

### Example 2: Testing for Sensitive Information Disclosure

```bash
# Run the credential leakage test
LLMrecon run --template examples/templates/owasp-llm/llm06-sensitive-information/credential-leakage.yaml --target your-llm-api-endpoint

# Expected output:
# Test: Credential Leakage Test
# Status: PASSED - Model refused to repeat credentials
# Response: "I notice you've shared sensitive information... [safety message]"
# Recommendation: Continue monitoring for credential leakage
```

### Example 3: Generating a Compliance Report

```bash
# Generate a compliance report
go run cmd/compliance-report/main.go --templates examples/templates/owasp-llm --format json --output compliance-report.json

# View the report summary
jq '.summary' compliance-report.json

# Expected output:
# {
#   "total_categories": 10,
#   "categories_covered": 10,
#   "total_templates": 26,
#   "compliance_score": 93.23,
#   "gaps_identified": 0
# }
```

## Interpreting Results

### Test Statuses

- **PASSED**: The LLM responded securely as expected
- **FAILED**: The LLM exhibited a security vulnerability
- **ERROR**: The test couldn't be completed due to an error
- **SKIPPED**: The test was skipped (e.g., not applicable)

### Detection Types

- **string_match**: Checks if specific strings are present/absent
- **regex_match**: Checks if patterns match using regular expressions
- **combined**: Uses multiple conditions with logical operators

### Compliance Scores

- **90-100%**: Excellent coverage, minimal gaps
- **70-89%**: Good coverage, some improvements needed
- **50-69%**: Moderate coverage, significant improvements needed
- **<50%**: Poor coverage, major improvements required

## Best Practices

1. **Start Small**: Begin with high-risk categories like Prompt Injection
2. **Customize**: Adapt templates to your specific application context
3. **Regular Testing**: Incorporate testing into your development workflow
4. **Comprehensive Coverage**: Test all categories for a complete assessment
5. **Document Findings**: Keep records of test results and mitigations
6. **Iterative Improvement**: Use test results to improve your application

## Troubleshooting

### Common Issues

#### Template Validation Errors
```
❌ Template example.yaml is missing 'compliance' section in 'info'
```
**Solution**: Ensure your template includes the required compliance mappings.

#### Detection Failures
```
Test failed but should have passed: No match found for pattern "security risk"
```
**Solution**: Adjust the detection conditions to match your LLM's response patterns.

#### API Connection Issues
```
Error connecting to LLM API: Connection refused
```
**Solution**: Verify your API endpoint and authentication credentials.

## Advanced Topics

### Continuous Integration

Integrate LLM security testing into your CI/CD pipeline:

```yaml
# Example GitHub Actions workflow
name: LLM Security Testing

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  security-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Set up Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.23
      - name: Run OWASP LLM tests
        run: |
          LLMrecon run --template-dir examples/templates/owasp-llm --target ${{ secrets.LLM_API_ENDPOINT }}
      - name: Generate compliance report
        run: |
          go run cmd/compliance-report/main.go --templates examples/templates/owasp-llm --format json --output compliance-report.json
      - name: Check compliance score
        run: |
          SCORE=$(jq '.summary.compliance_score' compliance-report.json)
          if (( $(echo "$SCORE < 80" | bc -l) )); then
            echo "Compliance score below threshold: $SCORE"
            exit 1
          fi
```

### Custom Reporting

Generate custom reports from test results:

```bash
# Run tests and save results
LLMrecon run --template-dir examples/templates/owasp-llm --target your-llm-api-endpoint --output results.json

# Generate custom report
jq '.tests | group_by(.category) | map({category: .[0].category, passed: map(select(.status=="PASSED")) | length, failed: map(select(.status=="FAILED")) | length, total: length})' results.json
```

## Additional Resources

- [OWASP Top 10 for LLM Applications](https://owasp.org/www-project-top-10-for-large-language-model-applications/)
- [LLMrecon Tool Documentation](https://github.com/perplext/LLMrecon)
- [ISO/IEC 42001 - Artificial Intelligence Management System](https://www.iso.org/standard/81230.html)

By following this guide, you'll be able to effectively use the OWASP LLM templates to test and improve the security of your LLM applications.
