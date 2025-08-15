# Security Guidelines for LLMrecon

## Overview

This document outlines security best practices and guidelines for contributing to the LLMrecon project. All contributors must follow these guidelines to maintain the security posture of the codebase.

## Core Security Principles

### 1. Defensive Security Only
- This tool is for **defensive security research only**
- Never implement features that could be used maliciously
- Focus on detection, analysis, and defense capabilities
- Test only on systems you own or have explicit permission to test

### 2. No Hardcoded Secrets
- **NEVER** hardcode API keys, passwords, or tokens
- Use environment variables for all sensitive configuration
- Example:
  ```go
  // BAD
  apiKey := "sk-1234567890abcdef"
  
  // GOOD
  apiKey := os.Getenv("API_KEY")
  if apiKey == "" {
      return errors.New("API_KEY environment variable not set")
  }
  ```

### 3. Cryptographically Secure Random
- **ALWAYS** use `crypto/rand` for security-sensitive randomness
- Never use `math/rand` for tokens, nonces, or keys
- Example:
  ```go
  // BAD
  import "math/rand"
  token := rand.Intn(1000000)
  
  // GOOD
  import "crypto/rand"
  b := make([]byte, 32)
  if _, err := rand.Read(b); err != nil {
      return err
  }
  ```

### 4. HTTPS-Only Communication
- All external API calls must use HTTPS
- Never downgrade to HTTP for convenience
- Validate TLS certificates properly

### 5. Input Validation
- **ALWAYS** validate and sanitize user input
- Use `filepath.Clean()` for file paths
- Use parameterized queries for databases
- Example:
  ```go
  // Path traversal protection
  cleanPath := filepath.Clean(userPath)
  if !strings.HasPrefix(cleanPath, allowedDir) {
      return errors.New("invalid path")
  }
  ```

## Security Checklist for Contributors

Before submitting a PR, ensure:

- [ ] No hardcoded secrets or credentials
- [ ] All randomness uses crypto/rand
- [ ] All external connections use HTTPS
- [ ] All file operations use filepath.Clean()
- [ ] All database queries are parameterized
- [ ] Error messages don't leak sensitive information
- [ ] Proper authentication/authorization checks
- [ ] Input validation on all user-provided data
- [ ] No use of deprecated or insecure functions
- [ ] Security tests included for new features

## Common Security Patterns

### Environment Variable Management
```go
func getRequiredEnv(key string) (string, error) {
    value := os.Getenv(key)
    if value == "" {
        return "", fmt.Errorf("%s environment variable not set", key)
    }
    return value, nil
}
```

### Secure File Operations
```go
func readSecureFile(userPath string) ([]byte, error) {
    cleanPath := filepath.Clean(userPath)
    
    // Validate path is within allowed directory
    absPath, err := filepath.Abs(cleanPath)
    if err != nil {
        return nil, err
    }
    
    if !strings.HasPrefix(absPath, allowedDir) {
        return nil, errors.New("access denied")
    }
    
    return os.ReadFile(absPath)
}
```

### Secure Token Generation
```go
func generateSecureToken() (string, error) {
    b := make([]byte, 32)
    if _, err := rand.Read(b); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}
```

## Security Tools Integration

### 1. Gosec
Run before committing:
```bash
gosec -fmt json -out gosec-report.json ./...
```

### 2. CodeQL
Automatically runs on GitHub Actions for all PRs

### 3. Dependency Scanning
Regular updates using:
```bash
go get -u ./...
go mod tidy
```

## Incident Response

If you discover a security vulnerability:

1. **DO NOT** create a public issue
2. Contact the security team at security@llmrecon.dev
3. Provide detailed information about the vulnerability
4. Wait for confirmation before disclosure

## Compliance

This project aims to comply with:
- OWASP Top 10 for LLMs
- ISO/IEC 42001 AI Management Standards
- Common security best practices

## Regular Security Tasks

### Weekly
- Review dependency updates
- Check for new CVEs in dependencies

### Monthly
- Full security scan with gosec
- Review and update security documentation
- Audit access logs and permissions

### Quarterly
- Comprehensive security audit
- Penetration testing (authorized only)
- Update threat model

## Resources

- [OWASP LLM Top 10](https://owasp.org/www-project-top-10-for-large-language-model-applications/)
- [Go Security Best Practices](https://golang.org/doc/security)
- [CWE Top 25](https://cwe.mitre.org/top25/)

## Questions?

For security-related questions, contact the security team or open a discussion in the security channel.