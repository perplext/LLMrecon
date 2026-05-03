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
   - Audit trail logging via `src/audit/trail` (`AuditTrailManager`).
   - Secure communication with TLS.
   - Prompt injection protection and content filtering.
   - **Note**: the v0.2.0 RBAC + MFA + auth subsystem under
     `src/security/access/` was removed in v0.10.0 (#180) — every
     constructor returned "not implemented", four CLI commands
     (`access_control` / `audit` / `auth` / `user`) sat as `.disabled`
     files, and no consumer ever wired the framework end-to-end. If
     auth/RBAC returns it'll be in v0.11.0+ as a fresh design rather
     than a partial revival.

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

## v0.9.0 New Attack Modules + Outcome Taxonomy + Provider Capabilities

Version 0.9.0 adds 6 Go attack modules and 5 templates covering Q4 2025 – Q2 2026 research, plus the cross-cutting infrastructure that makes outcomes machine-comparable across the ML/bandit/compliance stack.

### v0.9.0 Attack Module Inventory

| Module | Path | Source | OWASP Agentic | Safety gate |
|--------|------|--------|---------------|-------------|
| `minja` / `memorygraft` / `injecmem` | `src/attacks/memory/poisoning.go` | arXiv 2503.03704 / 2512.16962 / OpenReview QVX6hcJ2um | ASI06 (+ASI10 for memorygraft) | `i_understand_risks=true` |
| `h_cot` | `src/attacks/reasoning/h_cot.go` | arXiv 2502.12893, 2510.26418 | ASI01 | `i_understand_risks=true` + `common.ReasoningProvider` |
| `siva` | `src/attacks/multimodal/siva.go` | arXiv 2602.08136 | ASI01 | `common.ImageProvider` |
| `vsh` | `src/attacks/multimodal/vsh.go` | ScienceDirect S0031320325010520 | ASI01 | `common.ImageProvider` |
| `jbfuzz` | `src/attacks/adaptive/jbfuzz.go` | arXiv 2503.08990 | ASI01 | `allow_experimental=true` |
| `persona_evolve` | `src/attacks/adaptive/persona_evolve.go` | arXiv 2507.22171 | ASI01 + ASI09 | `allow_experimental=true` |

### `AttackOutcome` 3-state taxonomy (`src/attacks/common/types.go`)

Every v0.9.0 module returns one of three outcomes via `common.NewAttackResult`:

```go
type AttackOutcome string
const (
    OutcomeSuccess AttackOutcome = "success"   // attack landed
    OutcomeRefused AttackOutcome = "refused"   // ran fully, target resisted
    OutcomeSkipped AttackOutcome = "skipped"   // didn't run (capability/gate/budget/error)
)
```

Skipped runs **always** carry a typed `SkipReason`. The full enum lives in `common/types.go`; the attack-author rule of thumb:

- Use `SkipMissingCapability` when type assertion against an optional provider interface fails.
- Use `SkipGateBlocked` when a safety-gate metadata flag is missing.
- Use `SkipBudgetExceeded` when an evolutionary engine exhausted query/wall-clock/generation budget without success.
- Use `SkipProviderError` for transient/network failures — never silent `Success=false`.
- Use `SkipPreconditionFailed` for operator-config errors (missing payload, missing corpus path, etc).

The bandit-relevant invariant: **`OutcomeSkipped` rows are excluded from reward aggregation.** Skipped reflects engine/capability state, not target behavior; including them biases the bandit toward (or against) targets that simply lack capabilities. The Python pipeline exposes this filter via `AttackDataPipeline.get_bandit_rewards()`, the canonical single point of truth.

### Optional provider capabilities (`src/attacks/common/capabilities.go`)

Modules type-assert at `Execute()` entry and emit `OutcomeSkipped + SkipMissingCapability` when the assertion fails. Five interfaces:

```go
// Image input — multimodal SIVA/VSH.
type ImageProvider interface {
    Provider
    QueryWithImages(ctx context.Context, prompt string, images []ImagePayload, options map[string]interface{}) (string, error)
}

// Session lifecycle — memorygraft cross-session verification.
type SessionProvider interface {
    Provider
    SessionID() string
    NewSession(ctx context.Context) (Provider, error)
}

// Memory introspection — all memory-poisoning modes fail-fast on stateless targets.
type MemoryProbe interface {
    Provider
    ProbeMemory(ctx context.Context) (retains bool, err error)
}

// Reasoning trace — H-CoT mutation source.
type ReasoningProvider interface {
    Provider
    QueryWithReasoning(ctx context.Context, messages []Message, options map[string]interface{}) (response string, trace ReasoningTrace, err error)
}

// Module-side cleanup hook (NOT a provider interface).
type Cleaner interface {
    Cleanup(ctx context.Context, recordIDs []string) error
}
```

`ImagePayload` is constructor-only (`NewImagePayloadBytes` / `NewImagePayloadURL`) — direct struct literals fail at compile time because the fields are unexported. Constructors validate MIME type, size cap (`MaxImagePayloadBytes = 5 MiB`), and detail enum.

`ReasoningTrace.Signed=true` indicates an Anthropic-style cryptographically-signed thinking block whose text cannot be modified on round-trip; H-CoT short-circuits to `SkipSignatureGated` rather than wasting a re-injection.

### Safety-gate flag matrix

| Flag | Modules requiring it |
|------|----------------------|
| `i_understand_risks=true` | `minja`, `memorygraft`, `injecmem`, `h_cot`, `rce_chain` |
| `allow_experimental=true` | `jbfuzz`, `persona_evolve`, `autonomous_jailbreak`, `deceptive_alignment`, `agent_collusion` |
| `allow_autonomous=true` | `autonomous_jailbreak` (per v0.8.0) |

Modules emit `OutcomeSkipped + SkipGateBlocked` when the corresponding flag is missing.

### `EngineBudget` + hard ceilings (evolutionary engines)

JBFuzz and persona_evolve share `common.EngineBudget` knobs and hard ceilings:

```go
const (
    DefaultMaxQueries          = 100
    DefaultMaxWallClockSeconds = 180
    DefaultMaxGenerations      = 25
    HardMaxQueries             = 5000
    HardMaxWallClockSeconds    = 1800   // 30 min
    HardMaxGenerations         = 200
)
```

Hard ceilings clamp operator config; `(*EngineBudget).Clamp()` returns a slice of human-readable strings naming each clamped knob. Modules surface this in `result.Metadata["budget_clamped"]`.

### `RetryableQuery` retry helper (`src/provider/core/retry.go`)

Generic retry loop wrapping `func(ctx) (T, error)`. Two-class typed errors:

- `*TransientError` (rate limit, 5xx, network) — retry with exponential backoff, jitter, optional `Retry-After` honor, ctx-aware sleep via `select{case <-ctx.Done(): … case <-t.C: …}`.
- `*PermanentError` (auth, content-policy, schema mismatch) — surface immediately without retry.
- Any other type — surface immediately. Buggy provider returns must not absorb the retry budget.

ctx cancellation always wins: cancelling mid-loop returns `ctx.Err()`, never the previous transient. (PR #164 fixed an internal inconsistency where a top-of-loop check could let `lastErr` mask the cancellation.)

### OWASP Agentic 2026 codegen (`cmd/owasp-gen`)

Reads `templates/owasp_agentic_2026.yaml` (the canonical source) and emits `src/compliance/owasp_agentic_generated.go` containing `GeneratedTechniqueToAgenticCategories`. v0.9.0 ships the generator side-by-side with the existing hand-written `TechniqueToAgenticCategories`; v0.10.0 will switch the runtime lookup to the generated map and add a `go generate ./... && git diff --exit-code` drift check.

### ML pipeline migration (`ml/data/attack_data_pipeline.py`)

Two new columns: `outcome TEXT`, `parent_run_id TEXT`. Partial indexes (`idx_attacks_outcome`, `idx_attacks_parent_run_id`).

`_migrate_v090(conn, db_path)`:
- Idempotent: `PRAGMA table_info` introspection skips already-applied ALTERs; partial indexes use `IF NOT EXISTS`.
- Backs up via SQLite's **Online Backup API** (`Connection.backup`), not `shutil.copy2`. WAL-mode databases keep recently-committed transactions in `-wal`/`-shm` sidecar files; a filesystem copy can produce an inconsistent snapshot.
- Backup file (`{db}.bak.{UTC-timestamp}`) created only on the first run that detects a missing column.
- Backfills `outcome` from legacy `status` once: `'success' if status='success' else 'refused'`.

`_redact_sensitive_keys()` recursively scrubs dict keys matching `(?i)(key|token|secret|password|auth)` in `technique_params` and `features` before INSERT.

`get_bandit_rewards(target_model, attack_type, limit)` is the canonical filter: `WHERE outcome IN ('success', 'refused') ORDER BY timestamp DESC LIMIT ?` — applied INSIDE the LIMIT subquery so a burst of recent skipped rows doesn't displace rewardables. Returns `{success_rate, sample_count, skipped_count, outcomes}`.

### Integration smoke tests

`src/attacks/integration/integration_test.go` — one end-to-end smoke test per family (memory, reasoning, multimodal, adaptive). Each test calls `t.Skip()` (NOT `t.Fatal`) when `os.Getenv("RUN_INTEGRATION")` is unset, so CI is silent by default.

```bash
RUN_INTEGRATION=1 go test ./src/attacks/integration/...
```

Cost note: smoke tests against a local `MockLLMServer` are free. Running them against real providers (uncapped budgets, real keys) costs roughly $3–$8 per run on production-tier models — most of that is the GA engines.

## Security Considerations

- The tool is designed for security research and should only be used on systems you own or have permission to test
- Never use this tool against production systems without explicit authorization
- All testing should be conducted in isolated environments
- Results and data collected should be handled responsibly