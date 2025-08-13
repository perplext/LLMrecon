# Troubleshooting Guide

This guide helps you resolve common issues with LLMrecon. If you can't find a solution here, please check our [GitHub Issues](https://github.com/your-org/LLMrecon/issues) or ask on [Discord](https://discord.gg/LLMrecon).

## Table of Contents

1. [Installation Issues](#installation-issues)
2. [Configuration Problems](#configuration-problems)
3. [API Connection Errors](#api-connection-errors)
4. [Scan Failures](#scan-failures)
5. [Template Issues](#template-issues)
6. [Performance Problems](#performance-problems)
7. [Report Generation](#report-generation)
8. [Update Problems](#update-problems)
9. [Debug Mode](#debug-mode)
10. [Getting Help](#getting-help)

## Installation Issues

### Binary Not Found

**Problem**: `command not found: LLMrecon`

**Solutions**:

1. Check if binary is in PATH:
   ```bash
   echo $PATH
   ls -la /usr/local/bin/LLMrecon
   ```

2. Add to PATH:
   ```bash
   # For bash
   echo 'export PATH=$PATH:/usr/local/bin' >> ~/.bashrc
   source ~/.bashrc
   
   # For zsh (macOS)
   echo 'export PATH=$PATH:/usr/local/bin' >> ~/.zshrc
   source ~/.zshrc
   ```

3. Use full path:
   ```bash
   /usr/local/bin/LLMrecon --help
   ```

### Permission Denied

**Problem**: `permission denied` when running LLMrecon

**Solutions**:

```bash
# Make executable
chmod +x /usr/local/bin/LLMrecon

# If installed with sudo, you might need:
sudo chmod 755 /usr/local/bin/LLMrecon
```

### Build Failures

**Problem**: Build fails with Go errors

**Solutions**:

1. Check Go version:
   ```bash
   go version  # Should be 1.21 or higher
   ```

2. Clean and rebuild:
   ```bash
   go clean -cache
   go mod download
   go build -o LLMrecon ./src/main.go
   ```

3. Update dependencies:
   ```bash
   go mod tidy
   go mod verify
   ```

### Missing Dependencies

**Problem**: `cannot find package` errors

**Solutions**:

```bash
# Update all dependencies
go get -u ./...

# If specific package is missing
go get github.com/missing/package

# Vendor dependencies
go mod vendor
go build -mod=vendor -o LLMrecon ./src/main.go
```

## Configuration Problems

### Config File Not Found

**Problem**: `configuration file not found`

**Solutions**:

1. Initialize configuration:
   ```bash
   LLMrecon init
   ```

2. Check config location:
   ```bash
   # Default location
   ls -la ~/.LLMrecon/config.yaml
   
   # Custom location
   export LLM_RED_TEAM_CONFIG_DIR=/custom/path
   ```

3. Create manually:
   ```bash
   mkdir -p ~/.LLMrecon
   cat > ~/.LLMrecon/config.yaml <<EOF
   providers:
     openai:
       api_key: ${OPENAI_API_KEY}
   EOF
   ```

### Invalid Configuration

**Problem**: `invalid configuration format`

**Solutions**:

1. Validate YAML syntax:
   ```bash
   # Online validator: https://www.yamllint.com/
   
   # Or use yamllint
   pip install yamllint
   yamllint ~/.LLMrecon/config.yaml
   ```

2. Check for common issues:
   - Incorrect indentation (use spaces, not tabs)
   - Missing colons after keys
   - Unclosed quotes
   - Invalid special characters

3. Use example config:
   ```bash
   LLMrecon config example > config.yaml
   LLMrecon config validate config.yaml
   ```

### Environment Variables Not Working

**Problem**: Environment variables not being substituted

**Solutions**:

1. Check variable format:
   ```yaml
   # Correct
   api_key: ${OPENAI_API_KEY}
   
   # Incorrect
   api_key: $OPENAI_API_KEY
   api_key: {OPENAI_API_KEY}
   ```

2. Verify environment:
   ```bash
   # Check if set
   echo $OPENAI_API_KEY
   
   # Export if needed
   export OPENAI_API_KEY="sk-..."
   ```

3. Debug substitution:
   ```bash
   LLMrecon config get openai.api_key --show-source
   ```

## API Connection Errors

### Authentication Failed

**Problem**: `401 Unauthorized` or `403 Forbidden`

**Solutions**:

1. Verify API key:
   ```bash
   # Check configuration
   LLMrecon config get openai.api_key
   
   # Test directly
   curl https://api.openai.com/v1/models \
     -H "Authorization: Bearer $OPENAI_API_KEY"
   ```

2. Check key format:
   - OpenAI: Should start with `sk-`
   - Anthropic: Should start with `sk-ant-`
   - Remove extra spaces or quotes

3. Verify key permissions:
   - Ensure key has necessary scopes
   - Check if key is active
   - Verify billing/credits available

### Connection Timeout

**Problem**: `connection timeout` or `context deadline exceeded`

**Solutions**:

1. Increase timeout:
   ```bash
   LLMrecon config set request.timeout 60s
   
   # Or per scan
   LLMrecon scan --timeout 120s
   ```

2. Check network:
   ```bash
   # Test connectivity
   ping api.openai.com
   curl -I https://api.openai.com
   
   # Check DNS
   nslookup api.openai.com
   ```

3. Use proxy if needed:
   ```bash
   export HTTP_PROXY=http://proxy.company.com:8080
   export HTTPS_PROXY=http://proxy.company.com:8080
   
   # Or in config
   LLMrecon config set network.proxy http://proxy.company.com:8080
   ```

### SSL/TLS Errors

**Problem**: `x509: certificate signed by unknown authority`

**Solutions**:

1. Update CA certificates:
   ```bash
   # Ubuntu/Debian
   sudo apt-get update && sudo apt-get install ca-certificates
   
   # macOS
   brew install ca-certificates
   ```

2. Disable verification (NOT for production):
   ```bash
   LLMrecon config set security.tls_verify false
   ```

3. Add custom CA:
   ```bash
   LLMrecon config set security.ca_bundle /path/to/ca-bundle.crt
   ```

### Rate Limiting

**Problem**: `429 Too Many Requests`

**Solutions**:

1. Configure rate limiting:
   ```bash
   # Set rate limit
   LLMrecon config set providers.openai.rate_limit "10/min"
   
   # Or in scan
   LLMrecon scan --rate-limit "5/min"
   ```

2. Use exponential backoff:
   ```bash
   LLMrecon config set retry.enabled true
   LLMrecon config set retry.max_attempts 5
   LLMrecon config set retry.backoff exponential
   ```

3. Reduce parallelism:
   ```bash
   LLMrecon scan --parallel 1
   ```

## Scan Failures

### Templates Not Found

**Problem**: `no templates found` or `template not found: xxx`

**Solutions**:

1. Update templates:
   ```bash
   LLMrecon template update
   ```

2. Check template location:
   ```bash
   LLMrecon template list
   LLMrecon template paths
   ```

3. Use absolute path:
   ```bash
   LLMrecon scan --template /full/path/to/template.yaml
   ```

4. Verify template ID:
   ```bash
   # List available
   LLMrecon template list --category prompt-injection
   
   # Search
   LLMrecon template search injection
   ```

### Provider Errors

**Problem**: `provider error: model not found`

**Solutions**:

1. Check available models:
   ```bash
   LLMrecon provider models openai
   ```

2. Update model name:
   ```bash
   # Fix common mistakes
   # Wrong: gpt4, gpt-4-turbo
   # Right: gpt-4, gpt-4-1106-preview
   
   LLMrecon config set openai.model gpt-4
   ```

3. Test provider:
   ```bash
   LLMrecon provider test openai
   ```

### Memory Issues

**Problem**: `out of memory` or scan crashes

**Solutions**:

1. Limit concurrency:
   ```bash
   LLMrecon scan --parallel 1 --batch-size 10
   ```

2. Enable streaming:
   ```bash
   LLMrecon config set processing.stream_mode true
   ```

3. Increase memory:
   ```bash
   # Set memory limit
   export GOGC=100
   export GOMEMLIMIT=4GiB
   ```

## Template Issues

### Template Validation Errors

**Problem**: `invalid template format`

**Solutions**:

1. Validate template:
   ```bash
   LLMrecon template validate my-template.yaml
   ```

2. Common fixes:
   ```yaml
   # Required fields
   id: unique-template-id  # Must be unique
   info:
     name: Template Name   # Required
     severity: high        # Required: critical|high|medium|low|info
   test:
     prompt: "Test"        # Required
     detection:            # Required
       type: string_match
       pattern: "response"
   ```

3. Check YAML syntax:
   ```bash
   # Install yamllint
   pip install yamllint
   yamllint my-template.yaml
   ```

### False Positives

**Problem**: Template triggers on benign responses

**Solutions**:

1. Refine detection:
   ```yaml
   detection:
     type: multi
     operator: AND
     conditions:
       - type: string_match
         pattern: "INJECTION"
       - type: not
         condition:
           type: string_match
           pattern: "I cannot"
   ```

2. Test with known responses:
   ```bash
   LLMrecon template test my-template.yaml \
     --response "Normal response text"
   ```

3. Add exclusions:
   ```yaml
   detection:
     exclude_patterns:
       - "I cannot help"
       - "I'm unable to"
       - "As an AI assistant"
   ```

### Template Performance

**Problem**: Templates run slowly

**Solutions**:

1. Simplify regex:
   ```yaml
   # Slow
   pattern: "(?i)(?:password|passwd|pwd).*(?:=|:).*(?:\w+)"
   
   # Faster
   pattern: "password[:=]"
   ```

2. Reduce variations:
   ```yaml
   # Limit variations
   test:
     variations_limit: 10
   ```

3. Cache results:
   ```bash
   LLMrecon config set cache.enabled true
   LLMrecon config set cache.ttl 3600
   ```

## Performance Problems

### Slow Scans

**Problem**: Scans take too long

**Solutions**:

1. Increase parallelism:
   ```bash
   LLMrecon scan --parallel 10
   ```

2. Use specific templates:
   ```bash
   # Instead of all templates
   LLMrecon scan --template prompt-injection/basic
   
   # Or by severity
   LLMrecon scan --severity critical,high
   ```

3. Enable caching:
   ```bash
   LLMrecon config set cache.enabled true
   LLMrecon config set cache.provider.ttl 3600
   ```

### High CPU Usage

**Problem**: Tool uses too much CPU

**Solutions**:

1. Limit workers:
   ```bash
   export GOMAXPROCS=2
   LLMrecon scan --parallel 2
   ```

2. Enable CPU profiling:
   ```bash
   LLMrecon scan --profile cpu.prof
   go tool pprof cpu.prof
   ```

3. Reduce processing:
   ```bash
   LLMrecon config set processing.light_mode true
   ```

### Memory Leaks

**Problem**: Memory usage grows over time

**Solutions**:

1. Enable memory profiling:
   ```bash
   LLMrecon scan --profile mem.prof
   go tool pprof -http=:8080 mem.prof
   ```

2. Limit cache size:
   ```bash
   LLMrecon config set cache.max_size 100MB
   LLMrecon config set cache.eviction_policy lru
   ```

3. Restart periodically:
   ```bash
   # For service mode
   LLMrecon service restart
   ```

## Report Generation

### Report Generation Fails

**Problem**: `failed to generate report`

**Solutions**:

1. Check scan results exist:
   ```bash
   LLMrecon scan list
   LLMrecon scan get SCAN_ID
   ```

2. Try different format:
   ```bash
   # If PDF fails, try HTML
   LLMrecon report --format html
   
   # Or simple JSON
   LLMrecon report --format json
   ```

3. Check disk space:
   ```bash
   df -h
   # Ensure tmp has space
   ```

### Missing Data in Reports

**Problem**: Reports have incomplete data

**Solutions**:

1. Include all data:
   ```bash
   LLMrecon report --include-evidence \
     --include-requests \
     --include-responses
   ```

2. Check scan completion:
   ```bash
   LLMrecon scan status SCAN_ID
   ```

3. Regenerate report:
   ```bash
   LLMrecon report regenerate --scan-id SCAN_ID
   ```

## Update Problems

### Update Check Fails

**Problem**: `failed to check for updates`

**Solutions**:

1. Check connectivity:
   ```bash
   curl https://api.github.com/repos/your-org/LLMrecon/releases/latest
   ```

2. Manual update:
   ```bash
   # Download latest
   wget https://github.com/your-org/LLMrecon/releases/latest/download/LLMrecon-linux-amd64.tar.gz
   
   # Install
   tar -xzf LLMrecon-linux-amd64.tar.gz
   sudo mv LLMrecon /usr/local/bin/
   ```

3. Disable auto-update:
   ```bash
   LLMrecon config set update.auto_check false
   ```

### Template Update Issues

**Problem**: `failed to update templates`

**Solutions**:

1. Clear template cache:
   ```bash
   rm -rf ~/.LLMrecon/templates/.cache
   LLMrecon template update --force
   ```

2. Check permissions:
   ```bash
   ls -la ~/.LLMrecon/templates
   chmod -R 755 ~/.LLMrecon/templates
   ```

3. Use different source:
   ```bash
   LLMrecon template add-source https://github.com/alt/templates
   ```

## Debug Mode

### Enable Debugging

For detailed troubleshooting:

```bash
# Maximum verbosity
LLMrecon --debug --verbose scan ...

# Save debug logs
LLMrecon --debug scan ... 2>debug.log

# Specific component
export LLM_RED_TEAM_DEBUG=provider,template
LLMrecon scan ...
```

### Debug Output

Understanding debug output:

```
[DEBUG] provider: Initializing OpenAI provider
[DEBUG] template: Loading template prompt-injection-basic
[DEBUG] http: POST https://api.openai.com/v1/chat/completions
[DEBUG] http: Response 200 OK (450ms)
[DEBUG] detection: Running string_match detector
[DEBUG] detection: Pattern "INJECTION" not found
```

### Common Debug Flags

```bash
# HTTP debugging
export LLM_RED_TEAM_DEBUG_HTTP=true

# Template debugging  
export LLM_RED_TEAM_DEBUG_TEMPLATE=true

# Provider debugging
export LLM_RED_TEAM_DEBUG_PROVIDER=true

# All debugging
export LLM_RED_TEAM_DEBUG=*
```

## Getting Help

### Diagnostic Information

Collect for bug reports:

```bash
# Generate diagnostic bundle
LLMrecon doctor --bundle diagnostic.zip

# Manual collection
LLMrecon version --full
LLMrecon config dump --sanitize
LLMrecon provider list
go version
uname -a
```

### Log Files

Default locations:

```bash
# Application logs
~/.LLMrecon/logs/LLMrecon.log

# Scan logs
~/.LLMrecon/scans/SCAN_ID/scan.log

# Error logs
~/.LLMrecon/logs/error.log
```

### Community Support

1. **GitHub Issues**: 
   - Search existing issues first
   - Use issue templates
   - Include diagnostic information

2. **Discord**:
   - `#help` channel for questions
   - `#bugs` for bug reports
   - `#templates` for template help

3. **Stack Overflow**:
   - Tag: `LLMrecon`
   - Include minimal reproducible example

### Commercial Support

For enterprise support:
- Email: support@LLMrecon.com
- SLA-based support available
- Training and consulting services

## Quick Fixes Checklist

Before seeking help, try:

- [ ] Update to latest version
- [ ] Check configuration syntax
- [ ] Verify API keys are valid
- [ ] Test network connectivity
- [ ] Clear cache directories
- [ ] Run with `--debug` flag
- [ ] Check disk space
- [ ] Review recent changes
- [ ] Test with minimal config
- [ ] Restart the tool/service

## Known Issues

Check our [Known Issues](https://github.com/your-org/LLMrecon/wiki/Known-Issues) page for:
- Current bugs and workarounds
- Platform-specific issues
- Provider-specific limitations
- Compatibility notes