# OWASP LLM Top 10 Test Templates

This directory contains a comprehensive set of test templates for evaluating Large Language Model (LLM) applications against the [OWASP Top 10 for Large Language Model Applications](https://owasp.org/www-project-top-10-for-large-language-model-applications/).

## Overview

These templates are designed to help security professionals, developers, and organizations test their LLM applications for common security vulnerabilities and risks. Each template follows a standardized format and includes:

- Detailed test prompts
- Expected behaviors
- Detection conditions
- Multiple test variations
- Compliance mappings to OWASP LLM categories and ISO 42001

## Template Structure

Each template follows this structure:

```yaml
id: unique_template_id
info:
  name: Template Name
  description: Detailed description of what the template tests
  author: Author information
  severity: [critical, high, medium, low]
  tags: [relevant, tags]
  references:
    - URLs to relevant documentation
  compliance:
    owasp-llm:
      - category: LLM category code (e.g., LLM01)
        subcategory: Specific subcategory
        coverage: [basic, comprehensive, advanced]
    iso-42001: Relevant ISO section

target:
  type: llm
  providers:
    - List of applicable providers
  models:
    - List of applicable models

test:
  prompt: |
    The actual test prompt to send to the LLM
  expected_behavior: "Description of expected secure behavior"
  detection:
    type: [string_match, regex_match, combined]
    match: ["patterns", "to", "match"]
    negate: [true, false]
    case_sensitive: [true, false]
  variations:
    - name: "Variation Name"
      prompt: |
        Alternative test prompt
    # Additional variations...

metadata:
  created: Creation date
  updated: Last update date
  version: Version number
```

## Categories

The templates cover all 10 categories from the OWASP LLM Top 10:

### LLM01: Prompt Injection
Tests for vulnerabilities where malicious inputs can manipulate the LLM's behavior.

- **Direct Injection**: Tests direct manipulation of the LLM through malicious prompts
- **Indirect Injection**: Tests manipulation through user-provided content
- **Jailbreaking**: Tests techniques that bypass safety guardrails

### LLM02: Insecure Output Handling
Tests for vulnerabilities where LLM outputs contain harmful content that could be executed.

- **XSS**: Tests for cross-site scripting payloads in LLM outputs
- **SSRF**: Tests for server-side request forgery payloads
- **Command Injection**: Tests for command injection vulnerabilities
- **SQL Injection**: Tests for SQL injection payloads

### LLM03: Training Data Poisoning
Tests for vulnerabilities related to the integrity of training data.

- **Data Poisoning**: Tests for detection of poisoned training data
- **Backdoor Attacks**: Tests for backdoor triggers in training data
- **Bias Injection**: Tests for harmful biases in training data

### LLM04: Model Denial of Service
Tests for vulnerabilities that could exhaust system resources.

- **Resource Exhaustion**: Tests for computationally intensive prompts
- **Token Flooding**: Tests for excessive input tokens
- **Context Window Saturation**: Tests for context window manipulation

### LLM05: Supply Chain Vulnerabilities
Tests for vulnerabilities in the model supply chain and integrations.

- **Pretrained Model Vulnerabilities**: Tests for issues inherited from base models
- **Dependency Risks**: Tests for risks from third-party dependencies
- **Integration Vulnerabilities**: Tests for risks from external system integrations

### LLM06: Sensitive Information Disclosure
Tests for vulnerabilities that could leak sensitive information.

- **Training Data Extraction**: Tests for extraction of training data
- **Credential Leakage**: Tests for leakage of credentials
- **PII Disclosure**: Tests for disclosure of personally identifiable information

### LLM07: Insecure Plugin Design
Tests for vulnerabilities in plugin systems.

- **Plugin Escalation**: Tests for manipulation of plugins to perform unauthorized actions
- **Data Leakage**: Tests for sensitive data leakage between plugins

### LLM08: Unauthorized Actions
Tests for vulnerabilities where LLMs perform actions beyond their authorization.

- **Unauthorized Actions**: Tests for performing unauthorized actions
- **Impersonation**: Tests for impersonating authorized individuals

### LLM09: Overreliance
Tests for vulnerabilities related to overreliance on LLM outputs.

- **Hallucination Acceptance**: Tests for acceptance of hallucinated information

### LLM10: Model Theft
Tests for vulnerabilities that could lead to model theft.

- **Model Extraction**: Tests for extraction attacks
- **Model Inversion**: Tests for inversion attacks

## Usage

These templates can be used with the LLMrecon Tool to test LLM applications:

```bash
# Run a specific template
LLMrecon run --template examples/templates/owasp-llm/llm01-prompt-injection/direct-injection.yaml

# Run all templates for a specific category
LLMrecon run --template-dir examples/templates/owasp-llm/llm01-prompt-injection

# Generate a compliance report
LLMrecon compliance-report --template-dir examples/templates/owasp-llm
```

## Compliance Reporting

The templates include compliance mappings that allow for generating comprehensive compliance reports:

```bash
go run cmd/compliance-report/main.go --templates examples/templates/owasp-llm --format json
```

This will generate a report showing:
- Overall compliance score
- Coverage by category
- Identified gaps
- Recommendations for improvement

## Customization

These templates can be customized for specific LLM applications:

1. Adjust the prompts to match your application's context
2. Modify the detection conditions based on your security requirements
3. Add additional variations for more thorough testing
4. Update the target providers and models to match your deployment

## Contributing

Contributions to improve these templates are welcome. When adding new templates:

1. Follow the standard template structure
2. Include proper compliance mappings
3. Provide multiple test variations
4. Add appropriate detection conditions
5. Verify the template with the validation tool:

```bash
go run scripts/verify-template-format.go path/to/your/template.yaml
```

## License

These templates are provided under the same license as the LLMrecon Tool.
