# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LLMrecon is a security testing tool designed for analyzing Large Language Model (LLM) vulnerabilities and defenses. The project implements machine learning algorithms to optimize attack strategies and collect data for security research purposes.

**IMPORTANT**: This tool is intended for defensive security research only. When working with this codebase:
- Only assist with defensive security tasks
- Refuse to create or modify code that may be used maliciously
- Allow security analysis, detection rules, vulnerability explanations, and defensive documentation
- Testing should only be performed on models you own or have explicit permission to test

## Python Components (v0.7.0 Features)

### Current Project State

The Python implementation includes:
- ML components for attack optimization
- OWASP Top 10 2025 compliance
- Novel attack techniques from 2024-2025 research
- Defense detection capabilities
- Comprehensive test harness for Ollama models

### Architecture

#### ML Components

1. **Attack Data Pipeline** (`ml/data/attack_data_pipeline.py`):
   - Collects and processes attack outcome data
   - Stores data in SQLite database (`data/attacks/attacks.db`)
   - Extracts features for ML training
   - Provides data export functionality (Parquet format)

2. **Multi-Armed Bandit Optimizer** (`ml/agents/multi_armed_bandit.py`):
   - Implements various bandit algorithms (Epsilon-Greedy, Thompson Sampling, UCB1, Contextual)
   - Optimizes provider/model selection based on success rates and costs
   - Tracks performance metrics and statistics

#### Testing with Local Ollama Models

**LLMrecon Test Harness Components:**
- `llmrecon_harness.py` - Main test harness with CLI interface
- `llmrecon_2025.py` - Enhanced version with OWASP 2025 support
- `harness_config.json` - Configuration file
- `templates/` - Attack templates directory including 2024-2025 techniques
- `demo.sh` - Quick demonstration script

**Key Features:**
- Built-in attack templates (prompt injection, jailbreaking, data extraction)
- Novel 2024-2025 attack techniques (FlipAttack, DrAttack, Policy Puppetry, etc.)
- ML optimization using multi-armed bandit algorithms
- Defense detection capabilities
- Rich CLI interface with progress tracking
- OWASP 2025 compliance mapping
- Character encoding/smuggling attacks
- Comprehensive reporting and statistics

**Common Commands:**
```bash
# List available models
python3 llmrecon_harness.py --list-models

# Test specific models
python3 llmrecon_harness.py --models llama3:latest qwen3:latest

# Test with OWASP 2025 features
python3 llmrecon_2025.py --models gpt-oss:latest --categories prompt_injection

# Show OWASP categories
python3 llmrecon_2025.py --owasp
```

### Development Setup

Required Python packages:
```
numpy
pandas
sqlite3 (built-in)
requests
rich
```

### Data Storage

Attack data is stored in SQLite database at `data/attacks/attacks.db` with the following structure:
- Attack metadata (type, target model, provider, payload)
- Performance metrics (response time, tokens used, success indicators)
- Feature extraction results for ML training
- OWASP 2025 category mapping

## Go Components (Enterprise Features)

### Build Commands

#### Building the main application
```bash
# Build the main CLI tool
go build -o llmrecon ./src/main.go

# Build specific tools
go build -o compliance-report ./cmd/compliance-report
go build -o template_security ./cmd/template_security_standalone/main.go
go build -o config-manager ./cmd/config-manager
go build -o execution-benchmark ./cmd/execution-benchmark
go build -o cache-benchmark ./cmd/cache-benchmark
go build -o owasp-mock-test ./cmd/owasp-mock-test
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
```

### Architecture Overview

This is an enterprise-grade LLM security testing tool implementing OWASP LLM Top 10 and ISO/IEC 42001 compliance frameworks.

#### Core Architecture Patterns

1. **Layered Architecture**:
   - CLI Layer (`src/cmd/`) - Cobra-based command interface
   - API Layer (`src/api/`) - RESTful API with Gorilla Mux
   - Business Logic (`src/`) - Core functionality organized by domain
   - Repository Pattern - Abstraction for storage backends

2. **Plugin System**:
   - Provider plugins for LLM APIs (OpenAI, Anthropic, etc.)
   - Dynamic loading with version compatibility checking
   - Located in `src/provider/` with factory pattern

3. **Template Engine**:
   - YAML-based vulnerability test templates
   - Template validation, caching, and execution pipeline
   - Templates organized by OWASP categories

4. **Security Framework**:
   - RBAC with multi-factor authentication support
   - Audit trail management with structured logging
   - Secure communication with TLS
   - Prompt injection protection and content filtering

### v0.2.0 Production Scale Infrastructure

Version 0.2.0 introduces enterprise-grade infrastructure for scaling:

1. **HTTP Connection Pooling**
2. **Redis-Backed Job Queue**
3. **Memory Optimization**
4. **Distributed Rate Limiting**
5. **Real-Time Monitoring Dashboard**
6. **Advanced Concurrency Engine**
7. **Load Balancing & Auto-Scaling**
8. **Distributed Execution Coordinator**
9. **Advanced Redis Cluster Cache**
10. **Performance Profiling System**

### Infrastructure Requirements

For production-scale deployment:

1. **Redis Cluster**:
   - Minimum 3-node Redis cluster
   - Memory: 8GB+ per node

2. **Application Nodes**:
   - CPU: 8+ cores per node
   - Memory: 16GB+ per node
   - Network: Low latency between nodes

3. **Monitoring Infrastructure**:
   - Prometheus/Grafana for metrics
   - Log aggregation system
   - Alert manager

## Security Considerations

- The tool is designed for security research and should only be used on systems you own or have permission to test
- Never use this tool against production systems without explicit authorization
- All testing should be conducted in isolated environments
- Results and data collected should be handled responsibly