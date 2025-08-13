# Quick Start Guide

Get up and running with LLMrecon in 5 minutes.

## 1. Install

### Option A: Download Binary (Fastest)

```bash
# macOS
curl -L https://github.com/your-org/LLMrecon/releases/latest/download/LLMrecon-darwin-amd64.tar.gz | tar xz
sudo mv LLMrecon /usr/local/bin/

# Linux
curl -L https://github.com/your-org/LLMrecon/releases/latest/download/LLMrecon-linux-amd64.tar.gz | tar xz
sudo mv LLMrecon /usr/local/bin/

# Verify installation
LLMrecon version
```

### Option B: Install with Go

```bash
go install github.com/your-org/LLMrecon/src@latest
```

## 2. Configure

```bash
# Initialize configuration
LLMrecon init

# Set your OpenAI API key
export OPENAI_API_KEY="sk-..."

# Or configure directly
LLMrecon config set openai.api_key "sk-..."
```

## 3. Run Your First Scan

```bash
# Test for prompt injection vulnerabilities
LLMrecon scan --target openai --template prompt-injection/basic

# Run all OWASP LLM Top 10 tests
LLMrecon scan --target openai --compliance owasp-llm
```

## 4. View Results

```bash
# Get scan results
LLMrecon scan list
LLMrecon scan results <scan-id>

# Generate HTML report
LLMrecon report --format html --output report.html
open report.html  # macOS
# or
xdg-open report.html  # Linux
```

## Common Commands

```bash
# Update templates to latest
LLMrecon template update

# List available templates
LLMrecon template list

# Test specific vulnerability category
LLMrecon scan --target openai --category data-leakage

# Scan with custom template
LLMrecon scan --target openai --template ./my-template.yaml

# Get help
LLMrecon help
LLMrecon scan --help
```

## Example: Full Security Audit

```bash
# 1. Update everything
LLMrecon update
LLMrecon template update

# 2. Run comprehensive scan
LLMrecon scan \
  --target openai \
  --compliance owasp-llm \
  --output-dir ./audit-results \
  --parallel 5

# 3. Generate executive report
LLMrecon report \
  --scan-id latest \
  --format pdf \
  --template executive-summary \
  --output security-audit.pdf

# 4. View critical findings
LLMrecon scan results latest --severity critical,high
```

## Next Steps

- 📖 Read the [User Guide](user-guide.md) for detailed usage
- 🧪 Learn to [write custom templates](template-guide.md)
- 🔧 Set up [CI/CD integration](user-guide.md#integration-with-cicd)
- 🤝 [Contribute](../CONTRIBUTING.md) to the project

## Need Help?

- 💬 [Discord Community](https://discord.gg/LLMrecon)
- 📚 [Full Documentation](https://docs.LLMrecon.com)
- 🐛 [Report Issues](https://github.com/your-org/LLMrecon/issues)
- 📧 [Email Support](mailto:support@LLMrecon.com)