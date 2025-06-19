# OWASP LLM Templates Quick Start Guide

This guide provides a rapid introduction to using the OWASP LLM templates for testing LLM applications against security vulnerabilities.

## Prerequisites

- Go 1.18 or higher
- Access to an LLM application for testing

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/perplext/LLMrecon.git
   cd LLMrecon
   ```

2. Build the tools:
   ```bash
   go build -o bin/compliance-report ./cmd/compliance-report
   ```

## Running a Compliance Report

Generate a compliance report to see your current coverage:

```bash
./bin/compliance-report --templates examples/templates/owasp-llm --output report.json
```

This will analyze all templates in the directory and generate a report showing your coverage across the OWASP LLM categories.

## Testing a Specific Vulnerability

1. Choose a template to test, for example:

   ```bash
   cat examples/templates/owasp-llm/llm01-prompt-injection/direct-injection.yaml
   ```

2. Use the template prompt with your target LLM application and observe the response.

3. Compare the response against the detection criteria in the template.

## Template Structure

Each template follows this basic structure:

```yaml
id: unique-template-id
info:
  name: Template Name
  description: What this template tests
  severity: low|medium|high|critical
  tags:
    - owasp-llm
    - category-specific-tag
test:
  prompt: The prompt to send to the LLM
  expected_behavior: What the LLM should do
  detection:
    type: string_match|regex|semantic
    match: Pattern to look for in the response
compliance:
  owasp-llm:
    - LLM01:1.1  # Category and subcategory mapping
```

## Available Categories

The templates cover all 10 OWASP LLM categories:

1. **LLM01: Prompt Injection** - Test for vulnerabilities where malicious prompts can manipulate the LLM
2. **LLM02: Insecure Output Handling** - Test for output that could lead to XSS, SSRF, etc.
3. **LLM03: Training Data Poisoning** - Test for vulnerabilities from poisoned training data
4. **LLM04: Model Denial of Service** - Test for ways to exhaust resources or crash the LLM
5. **LLM05: Supply Chain Vulnerabilities** - Test for vulnerabilities in model dependencies
6. **LLM06: Sensitive Information Disclosure** - Test for leakage of sensitive data
7. **LLM07: Insecure Plugin Design** - Test for vulnerabilities in LLM plugins
8. **LLM08: Unauthorized Actions** - Test for unauthorized actions the LLM might perform
9. **LLM09: Overreliance** - Test for overreliance on potentially hallucinated information
10. **LLM10: Model Theft** - Test for vulnerabilities that could lead to model extraction

## Creating Your Own Templates

1. Copy an existing template as a starting point:
   ```bash
   cp examples/templates/owasp-llm/llm01-prompt-injection/direct-injection.yaml my-custom-template.yaml
   ```

2. Modify the template with your specific test case.

3. Validate your template format:
   ```bash
   go run scripts/verify-template-format.go my-custom-template.yaml
   ```

## Next Steps

- Read the [full documentation](README.md) for more details
- Check the [compliance mapping guide](COMPLIANCE.md) to understand how templates map to OWASP LLM categories
- Explore the [category guide](CATEGORY_GUIDE.md) for detailed information about each vulnerability category
- See the [usage guide](USAGE_GUIDE.md) for advanced usage patterns

## Troubleshooting

If you encounter issues:

1. Ensure your template follows the correct YAML format
2. Check that all required fields are present
3. Verify that compliance mappings are in the correct format (e.g., LLM01:1.1)
4. Run the debug script to understand how templates are being processed:
   ```bash
   go run scripts/debug-compliance-report.go --templates examples/templates/owasp-llm
   ```
