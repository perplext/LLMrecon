# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

### Building the main application
```bash
# Build the main CLI tool
go build -o LLMrecon ./src/main.go

# Build specific tools
go build -o compliance-report ./cmd/compliance-report
go build -o template_security ./cmd/template_security_standalone/main.go
go build -o config-manager ./cmd/config-manager
go build -o execution-benchmark ./cmd/execution-benchmark

# Build individual components
./scripts/build_component.sh template_security
./scripts/build_component.sh audit_logger
./scripts/build_component.sh memory_optimizer
./scripts/build_component.sh monitoring
```

### Running tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./src/template/...
go test ./src/security/...
go test ./src/bundle/...

# Run benchmarks
./scripts/run_memory_benchmark.sh
./scripts/run_execution_benchmark.sh
./scripts/benchmark_caching.sh
./scripts/benchmark_executors.sh
```

### Template development and validation
```bash
# Verify template compliance
./scripts/verify-compliance.sh

# Optimize templates
./scripts/optimize_templates.sh

# Run template security checks
go run ./cmd/template_security_standalone/main.go
```

## Architecture Overview

This is an enterprise-grade LLM security testing tool implementing OWASP LLM Top 10 and ISO/IEC 42001 compliance frameworks.

### Core Architecture Patterns

1. **Layered Architecture**:
   - CLI Layer (`src/cmd/`) - Cobra-based command interface
   - API Layer (`src/api/`) - RESTful API with Gorilla Mux
   - Business Logic (`src/`) - Core functionality organized by domain
   - Repository Pattern - Abstraction for storage backends (GitHub, GitLab, S3, local)

2. **Plugin System**:
   - Provider plugins for LLM APIs (OpenAI, Anthropic, etc.)
   - Dynamic loading with version compatibility checking
   - Located in `src/provider/` with factory pattern for instantiation

3. **Template Engine**:
   - YAML-based vulnerability test templates (similar to Nuclei)
   - Template validation, caching, and execution pipeline
   - Templates organized by OWASP categories in `examples/templates/owasp-llm/`

4. **Security Framework**:
   - RBAC with multi-factor authentication support
   - Audit trail management with structured logging
   - Secure communication with TLS and certificate management
   - Prompt injection protection and content filtering

### Key Components and Relationships

1. **Template Management System** (`src/template/`):
   - Loads and validates YAML templates
   - Manages template execution with rate limiting
   - Caches compiled templates for performance
   - Supports inheritance and modular composition

2. **Provider Framework** (`src/provider/`):
   - Interface-based design for LLM provider integration
   - Middleware stack: rate limiting, retries, circuit breaker, logging
   - Configuration management with encryption for sensitive data

3. **Update System** (`src/update/`):
   - Self-updating capability for binary and templates
   - Version management with semantic versioning
   - Signature verification for secure updates
   - Offline bundle support for air-gapped environments

4. **Reporting System** (`src/reporting/`):
   - Multiple output formats via factory pattern
   - Compliance-focused reporting for OWASP and ISO standards
   - Integration with vulnerability management systems

5. **Bundle System** (`src/bundle/`):
   - Offline distribution format for templates and modules
   - Conflict resolution for template updates
   - Import/export with validation and rollback support

### Important Design Decisions

1. **Interface-Heavy Design**: Most components define interfaces first, implementations second. This enables easy testing and extensibility.

2. **Factory Pattern Usage**: Providers, reports, and many other components use factories for instantiation, supporting runtime configuration.

3. **Middleware Architecture**: API and provider calls go through configurable middleware stacks for cross-cutting concerns.

4. **Template Inheritance**: Templates can inherit from base templates, promoting reuse and consistency.

5. **Multi-Repository Support**: Can sync templates from multiple sources (GitHub for production, GitLab for development).

### Current Development Focus

The codebase shows active development on:
- OWASP LLM Top 10 compliance implementation
- Memory optimization for large-scale operations
- Enhanced template security verification
- Offline bundle functionality
- Access control and authentication improvements

Note: Some dependencies need to be installed:
```bash
go get github.com/aws/aws-sdk-go-v2/aws
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/s3
go get github.com/go-sql-driver/mysql
go get github.com/lib/pq
go get golang.org/x/term
```