# Contributing to LLMrecon

Thank you for your interest in contributing to LLMrecon! This guide will help you get started with contributing to the project.

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Getting Started](#getting-started)
3. [How to Contribute](#how-to-contribute)
4. [Development Setup](#development-setup)
5. [Coding Standards](#coding-standards)
6. [Testing Guidelines](#testing-guidelines)
7. [Submitting Changes](#submitting-changes)
8. [Security Vulnerabilities](#security-vulnerabilities)
9. [Community](#community)

## Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

### Our Standards

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on constructive criticism
- Accept responsibility and apologize for mistakes
- Focus on what's best for the community

## Getting Started

### Prerequisites

Before contributing, ensure you have:

- Go 1.21 or higher installed
- Git configured with your GitHub account
- Basic understanding of security testing concepts
- Familiarity with the OWASP LLM Top 10

### Types of Contributions

We welcome various types of contributions:

- 🐛 **Bug fixes**: Fix issues in the codebase
- ✨ **Features**: Add new functionality
- 📝 **Documentation**: Improve docs or add examples
- 🧪 **Templates**: Create new security test templates
- 🌐 **Translations**: Translate documentation
- 🔍 **Security**: Report vulnerabilities (see [Security](#security-vulnerabilities))
- 💡 **Ideas**: Propose new features or improvements

## How to Contribute

### 1. Find an Issue

Look for issues labeled:
- `good first issue` - Great for newcomers
- `help wanted` - We need help with these
- `enhancement` - New features
- `bug` - Something needs fixing
- `documentation` - Doc improvements needed

Or create a new issue to discuss your idea.

### 2. Fork and Clone

```bash
# Fork the repository on GitHub, then:
git clone https://github.com/YOUR_USERNAME/LLMrecon.git
cd LLMrecon
git remote add upstream https://github.com/your-org/LLMrecon.git
```

### 3. Create a Branch

```bash
# Sync with upstream
git fetch upstream
git checkout -b feature/your-feature-name upstream/main

# For bugs:
git checkout -b fix/issue-description upstream/main

# For docs:
git checkout -b docs/update-description upstream/main
```

## Development Setup

### Environment Setup

```bash
# Install dependencies
go mod download

# Install development tools
make dev-tools

# This installs:
# - golangci-lint (linting)
# - gofumpt (formatting)
# - gomod (dependency management)
# - mockgen (test mocks)
```

### Build and Run

```bash
# Build the project
make build

# Run tests
make test

# Run with race detector
make test-race

# Run linters
make lint

# Format code
make fmt

# Run all checks (recommended before committing)
make check
```

### Project Structure

```
LLMrecon/
├── src/              # Source code
│   ├── cmd/         # CLI commands
│   ├── api/         # REST API
│   ├── template/    # Template engine
│   ├── provider/    # LLM providers
│   ├── security/    # Security features
│   └── reporting/   # Report generation
├── templates/        # Security test templates
├── docs/            # Documentation
├── examples/        # Example code
└── scripts/         # Build and utility scripts
```

## Coding Standards

### Go Code Style

We follow the standard Go style guide with some additions:

```go
// Package comments should be present
// Package security provides security testing functionality
package security

import (
    // Standard library imports first
    "context"
    "fmt"
    
    // External imports second
    "github.com/spf13/cobra"
    
    // Internal imports last
    "github.com/perplext/LLMrecon/src/types"
)

// SecurityScanner defines the interface for security scanning
// Always add interface comments
type SecurityScanner interface {
    // Scan performs a security scan on the target
    // Method comments explain what the method does
    Scan(ctx context.Context, target string) (*ScanResult, error)
}

// scanImpl implements SecurityScanner
// Unexported types need comments too
type scanImpl struct {
    provider Provider
    config   *Config
}

// Scan implements the SecurityScanner interface
func (s *scanImpl) Scan(ctx context.Context, target string) (*ScanResult, error) {
    // Validate inputs
    if target == "" {
        return nil, fmt.Errorf("target cannot be empty")
    }
    
    // Use meaningful variable names
    scanContext := &ScanContext{
        Target:    target,
        StartTime: time.Now(),
    }
    
    // Handle errors explicitly
    result, err := s.provider.Execute(ctx, scanContext)
    if err != nil {
        return nil, fmt.Errorf("scan failed: %w", err)
    }
    
    return result, nil
}
```

### Error Handling

```go
// Wrap errors with context
if err := scanner.Init(); err != nil {
    return fmt.Errorf("failed to initialize scanner: %w", err)
}

// Define error variables for common errors
var (
    ErrInvalidTemplate = errors.New("invalid template format")
    ErrProviderTimeout = errors.New("provider request timeout")
)

// Create custom error types for complex errors
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation error in %s: %s", e.Field, e.Message)
}
```

### Logging

```go
import "github.com/sirupsen/logrus"

// Use structured logging
log.WithFields(logrus.Fields{
    "scan_id": scanID,
    "target":  target,
    "duration": duration,
}).Info("Scan completed successfully")

// Log levels:
// - Debug: Detailed information for debugging
// - Info: General information
// - Warn: Warning conditions
// - Error: Error conditions
// - Fatal: Fatal errors (exits program)
```

## Testing Guidelines

### Unit Tests

Every new feature should include tests:

```go
func TestSecurityScanner_Scan(t *testing.T) {
    // Arrange
    mockProvider := mocks.NewMockProvider(t)
    scanner := NewSecurityScanner(mockProvider)
    
    mockProvider.EXPECT().
        Execute(gomock.Any(), gomock.Any()).
        Return(&ScanResult{Passed: true}, nil)
    
    // Act
    result, err := scanner.Scan(context.Background(), "test-target")
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.True(t, result.Passed)
}

// Table-driven tests for multiple cases
func TestValidateTemplate(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {
            name:    "valid template",
            input:   "id: test\ninfo:\n  name: Test",
            wantErr: false,
        },
        {
            name:    "missing id",
            input:   "info:\n  name: Test",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateTemplate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateTemplate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Integration Tests

```go
//go:build integration
// +build integration

func TestProviderIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Test with real provider
    provider := openai.NewProvider(os.Getenv("OPENAI_API_KEY"))
    scanner := NewSecurityScanner(provider)
    
    result, err := scanner.Scan(context.Background(), "test")
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Template Tests

For new templates:

```yaml
# templates/test-cases/prompt-injection-basic.yaml
test_cases:
  - name: "should detect basic injection"
    response: "INJECTION SUCCESS"
    expected: true
    
  - name: "should not trigger on normal response"
    response: "I cannot help with that request"
    expected: false
```

## Submitting Changes

### Commit Messages

Follow the conventional commits format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:
```
feat(templates): add advanced prompt injection detection

Add new template for detecting indirect prompt injection attacks
through external data sources. This addresses issue #123.

Closes #123
```

```
fix(api): handle rate limit errors gracefully

The API now properly returns 429 status codes when rate limited
and includes retry-after headers.
```

### Pull Request Process

1. **Update your branch**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all checks**:
   ```bash
   make check
   ```

3. **Push your changes**:
   ```bash
   git push origin feature/your-feature-name
   ```

4. **Create Pull Request**:
   - Use a clear, descriptive title
   - Reference any related issues
   - Describe what changes you made and why
   - Include screenshots for UI changes
   - Add tests for new functionality

5. **PR Template**:
   ```markdown
   ## Description
   Brief description of changes
   
   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Documentation update
   - [ ] Template addition
   
   ## Testing
   - [ ] Unit tests pass
   - [ ] Integration tests pass
   - [ ] Manual testing completed
   
   ## Checklist
   - [ ] Code follows project style
   - [ ] Self-review completed
   - [ ] Documentation updated
   - [ ] Tests added/updated
   
   Closes #(issue number)
   ```

### Code Review

- Be responsive to feedback
- Make requested changes promptly
- Ask questions if something is unclear
- Be patient - reviews ensure quality

## Security Vulnerabilities

### Reporting Security Issues

**DO NOT** open public issues for security vulnerabilities.

Instead:
1. Email security@LLMrecon.com
2. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We'll respond within 48 hours and work with you to resolve the issue.

### Security Review Process

All PRs undergo security review for:
- Input validation
- Authentication/authorization
- Sensitive data handling
- Dependency vulnerabilities
- Template security

## Community

### Getting Help

- 💬 [Discord](https://discord.gg/LLMrecon) - Real-time chat
- 🐦 [Twitter](https://twitter.com/llmredteam) - Updates and news
- 📧 [Mailing List](https://groups.google.com/g/LLMrecon) - Discussions
- 📺 [YouTube](https://youtube.com/@llmredteam) - Tutorials

### Meetings

- **Community Call**: First Tuesday of each month at 10 AM PST
- **Security SIG**: Every other Thursday at 2 PM PST
- **Template Authors**: Last Friday of each month at 1 PM PST

### Recognition

We recognize contributors in:
- Release notes
- Contributors page
- Annual contributors report
- Conference presentations

## Template Contribution Guide

### Template Standards

Templates must:
- Have unique IDs
- Include complete metadata
- Provide clear descriptions
- Include detection logic
- Pass validation tests
- Avoid false positives

### Template Review Process

1. Submit template PR
2. Automated validation runs
3. Security team reviews
4. Community testing (7 days)
5. Merge if approved

### Template Testing

```bash
# Validate template
LLMrecon template validate your-template.yaml

# Test against providers
LLMrecon template test your-template.yaml --all-providers

# Check for false positives
LLMrecon template fp-check your-template.yaml
```

## Release Process

We follow semantic versioning:
- **Major**: Breaking changes
- **Minor**: New features
- **Patch**: Bug fixes

Releases happen:
- Patch: As needed
- Minor: Monthly
- Major: Quarterly

## License

By contributing, you agree that your contributions will be licensed under the same license as the project (MIT License).

## Thank You!

Your contributions make LLMrecon better for everyone. We appreciate your time and effort in improving the project!

🙏 Special thanks to all our [contributors](https://github.com/your-org/LLMrecon/graphs/contributors)!