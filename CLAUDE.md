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
- Novel attack techniques from 2024-2026 research
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
- `templates/` - Attack templates directory including 2024-2026 techniques

**Key Features:**
- Built-in attack templates (prompt injection, jailbreaking, data extraction)
- Novel 2024-2026 attack techniques (FlipAttack, DrAttack, Policy Puppetry, etc.)
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

## v0.8.0 New Attack Techniques (2025-2026 Research)

Version 0.8.0 adds 45+ attack modules implementing the latest LLM and agentic AI security research from 2025-2026.

### Attack Module Architecture

All attack modules implement the `AttackModule` interface (`src/attacks/attack.go`) and self-register via `init()` with `attacks.DefaultRegistry`. Shared types live in `src/attacks/common/types.go`.

```go
// Core interface all modules implement
type AttackModule interface {
    Name() string
    Category() common.AttackCategory
    Description() string
    Techniques() []common.TechniqueInfo
    Execute(ctx context.Context, provider common.Provider, config common.AttackConfig) (*common.AttackResult, error)
}
```

**Safety gates**: Dangerous modules require explicit metadata flags:
- `allow_experimental=true` — Required for deceptive alignment and agent collusion modules
- `i_understand_risks=true` — Required for RCE chain module
- `allow_autonomous=true` — Required for autonomous jailbreak module

### Attack Packages

#### Phase 1: Core Attack Techniques (9 types, 19 variants)
| Package | Techniques | Source |
|---------|-----------|--------|
| `src/attacks/orchestration/` | Crescendo, Skeleton Key, Bad Likert Judge, Many-Shot | arXiv 2404.01833, Microsoft Research |
| `src/attacks/evasion/` | MetaBreak, Poetry-Based, Content Concretization, Immersive World, Best-of-N | arXiv 2407.15211, arXiv 2410.02650 |

#### Phase 2: New Attack Surfaces (4 packages, 15 modules)
| Package | Modules | Description |
|---------|---------|-------------|
| `src/attacks/rag/` | 4 | RAG pipeline poisoning: document injection, vector embedding, KG poisoning, cross-encoder |
| `src/attacks/agentic/mcp/` | 4 | MCP protocol attacks: tool poisoning, schema manipulation, filesystem escape, supply chain |
| `src/attacks/agentic/browser/` | 3 | AI browser agent attacks: DOM injection, navigation hijack, screenshot exfiltration |
| `src/attacks/audio/` | 4 | Audio modality: jailbreak, speech model exploit, multilingual audio, BoN audio |

#### Phase 3: Reasoning & Adaptive Attacks (3 packages, 9 modules)
| Package | Modules | Description |
|---------|---------|-------------|
| `src/attacks/reasoning/` | 3 | Autonomous jailbreak (uses reasoning models as adversaries), CoT exploitation, reasoning loops |
| `src/attacks/agentic/tool_use/` | 3 | iMIST function transform, AIShellJack agent shell injection, tool-use exploitation |
| `src/attacks/adaptive/` | 3 | Gradient-based optimization, RL optimization, diffusion-based attacks |

#### Phase 4: Model Adapters & Profiles
- Provider adapters for 7 model families: OpenAI (GPT-5.x), Anthropic (Claude 4.x), Google (Gemini 2.5), Meta (Llama 4), DeepSeek (V3.2/R2), Alibaba (Qwen3), xAI (Grok 3)
- Model-specific YAML profiles: `templates/model_profiles/` (reasoning, long-context, MoE, agentic)

#### Phase 5: OWASP Agentic Top 10 2026 Compliance
- `src/compliance/owasp_agentic.go` — ASI01-ASI10 constants, compliance types, `TechniqueToAgenticCategories()` lookup
- `templates/owasp_agentic_2026.yaml` — Bidirectional mapping with 70 test cases, MITRE ATLAS tactics, MAESTRO layers

#### Phase 6: Multi-Agent Orchestration (4 packages, 11 modules)
| Package | Modules | Description |
|---------|---------|-------------|
| `src/attacks/agentic/multi_agent/` | 3 | Delegation escalation, toxic agent flow, recursive spawn abuse |
| `src/attacks/agentic/skill_injection/` | 2 | Marketplace injection, skill takeover chain |
| `src/attacks/agentic/persistence/` | 3 | Config rewrite, credential harvest, RCE chain |
| `src/attacks/agentic/deception/` | 2 | Deceptive alignment, agent collusion |

### Framework Attack Profiles

YAML-based attack sequences for specific multi-agent frameworks (`templates/framework_profiles/`):
- **OpenClaw** — 512 CVEs, malicious skill marketplace, queue lane bypass
- **CrewAI** — No per-agent RBAC, raw output passing, unrestricted delegation
- **LangGraph** — State machine manipulation, recursive sub-agent spawning ($38K incident)
- **AutoGen** — Auto-execute code blocks, Docker sandbox escape, GroupChat trust

### Running New Attack Modules

```bash
# Build with new modules (self-register via init())
go build -o llmrecon ./src/main.go

# Run specific attack categories
go test ./src/attacks/rag/...
go test ./src/attacks/agentic/...
go test ./src/attacks/reasoning/...
go test ./src/attacks/audio/...
go test ./src/compliance/...
```

## Security Considerations

- The tool is designed for security research and should only be used on systems you own or have permission to test
- Never use this tool against production systems without explicit authorization
- All testing should be conducted in isolated environments
- Results and data collected should be handled responsibly