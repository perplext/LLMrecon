---
title: "feat: Add 2025-2026 LLM Attack Techniques and Model Adapters"
type: feat
date: 2026-02-11
deepened: 2026-02-11
---

# Add 2025-2026 LLM Attack Techniques, Model Adapters, and OWASP Agentic Mapping

## Enhancement Summary

**Deepened on:** 2026-02-11
**Research agents used:** Security Sentinel, Architecture Strategist, Performance Oracle, Pattern Recognition Specialist, Agent-Native Reviewer, Go Security Testing Patterns Researcher, OWASP Agentic Implementation Researcher (7 agents)

### Key Improvements
1. **Phase 0 added**: Architecture prerequisites (shared `AttackModule` interface, `src/attacks/common/` package, security fixes) must land before Phase 1 begins
2. **Security hardening**: 4 critical, 9 high findings identified -- cost bombs, credential leakage, self-DoS vectors, filesystem escape risks all require mitigations
3. **Performance guardrails**: Streaming payload construction for many-shot (10M tokens), BoN cost ceilings, agent tree depth/breadth limits, per-provider rate limiting
4. **6 missing attack categories**: Structured output exploitation, system prompt extraction, fine-tuning tax attacks, embedding space adversarial, multimodal fusion confusion, model merging exploitation
5. **Agent-native API surface**: 9/15 capabilities currently inaccessible to orchestrating agents -- attack execution, chaining, and multi-turn orchestration endpoints needed for Phases 3 and 6
6. **OWASP test case matrix**: 70+ concrete test cases across ASI01-ASI10 with MITRE ATLAS and MAESTRO framework cross-references
7. **Existing codebase debt**: Duplicated types across 6+ packages (Provider, Logger, Message, randInt, generateAttackID, contains) must be consolidated before adding 30+ new modules

### New Considerations Discovered
- Autonomous jailbreak module creates uncontrolled feedback loops -- needs human-in-the-loop gating
- Framework-specific profiles (`profile_openclaw.go` etc.) should be YAML-driven, not Go code, for maintainability
- China-based provider endpoints (DeepSeek, Qwen) have data residency implications requiring documentation
- Go/Python parity is a false goal as framed -- need explicit parity tiers (template, API, Python-only)
- Existing tools (DeepTeam, Promptfoo, ASB) already implement parts of OWASP Agentic testing -- consider integration vs. reimplementation

---

## Overview

LLMrecon currently implements 130+ attack techniques covering the OWASP LLM Top 10 2025. Since the last major update, 19 new attack categories, 7+ new frontier model families, a brand-new OWASP framework (Agentic Top 10 2026), and critical new attack surfaces (MCP, RAG pipelines, browser agents, audio modality) have emerged. This plan covers adding these capabilities across both the Python (ML-optimized) and Go (enterprise) components.

## Problem Statement

The LLM security landscape has evolved dramatically in 2025-2026:

1. **New attack surfaces**: MCP protocol (437K+ downloads of vulnerable packages), RAG pipelines (90% manipulation with 5 documents), AI browser agents (OpenAI admits "may never be fully solved"), and audio-language models (89% ASR)
2. **New model architectures**: Reasoning models (CoT exploitation, 97.14% autonomous jailbreak success), MoE models (expert routing manipulation), 10M-token context windows (many-shot amplification)
3. **New regulatory framework**: OWASP Top 10 for Agentic Applications 2026 creates compliance requirements the tool must map to
4. **Defense evolution**: CaMel (capability-based defense, 67% attack neutralization) and LLM Salting (reduces ASR to 3%) require adaptive attack generation

Without these updates, the tool cannot adequately test against current threats or provide compliance coverage for the 2026 agentic security landscape.

## Proposed Solution

A phased implementation adding 19 new attack technique categories plus 11 multi-agent/orchestrator attack modules, 7 model family adapters, OWASP Agentic 2026 compliance mapping, and 2 defense-bypass modules across both Python and Go components.

## Technical Approach

### Architecture

All new attack techniques follow the existing plugin architecture:
- **Go**: New attack modules under `src/attacks/` with YAML template definitions
- **Python**: New templates under `templates/` with ML pipeline integration via `ml/data/attack_data_pipeline.py`
- **Shared**: OWASP mapping extensions in both harness configurations

New attack categories that don't fit existing module directories will get their own packages:
- `src/attacks/rag/` - RAG pipeline poisoning
- `src/attacks/agentic/` - Browser agent, MCP, inter-agent attacks
- `src/attacks/audio/` - Audio modality attacks
- `src/attacks/reasoning/` - CoT exploitation, reasoning manipulation
- `src/attacks/adaptive/` - Defense-bypass optimization

### Implementation Phases

#### Phase 0: Architecture Prerequisites (Must Complete Before Phase 1)

Critical architectural debt and security fixes that must land before adding 30+ new modules.

**0.1 Shared Attack Module Interface** (`src/attacks/attack.go`)
- Define shared `AttackModule` interface: `Name()`, `Category()`, `Execute(ctx, config) (Result, error)`, `Techniques() []TechniqueInfo`
- Create `Registry` with `Register()`, `Get()`, `List()`, `ListByCategory()` methods
- Use `init()` self-registration pattern (nuclei-style) so each module registers itself on import

**0.2 Common Types Package** (`src/attacks/common/`)
- Consolidate 6 duplicated types into single canonical definitions:
  - `Provider` interface (currently 3 incompatible definitions across packages)
  - `Logger` interface (3 byte-for-byte identical copies)
  - `Message` struct (3 definitions)
  - `randInt()` helper (duplicated in 6 packages)
  - `generateAttackID()` (4 different implementations)
  - `contains()` (3 different implementations)
- Remove `min`/`max` redefinitions (Go 1.21+ builtins; `go.mod` specifies 1.23.0)

**0.3 Provider Interface Extensions** (`src/provider/core/provider.go`)
- Add optional capability interfaces (not expanding base Provider):
  - `ReasoningProvider`: `ChatWithReasoning(ctx, messages, config) (Response, ThinkingTrace, error)`
  - `AudioProvider`: `ChatWithAudio(ctx, messages, audioData, config) (Response, error)`
  - `LongContextProvider`: `MaxContextTokens() int`
  - `MCPProvider`: `ListTools()`, `InvokeTool()` for MCP-capable providers
- Update `ModelCapability` enum to include: `Reasoning`, `AudioInput`, `LongContext`, `MCPToolUse`

**0.4 Security Hardening**
- Fix weak KDF at `src/provider/config/encryption.go:89-92`: Replace bare SHA-256 with argon2id
- Remove hardcoded salt at `src/security/vault/vault.go:124`: Generate random salt per key
- Enforce TLS for all provider `BaseURL` fields: Reject `http://` URLs in provider config validation
- Add cost ceiling to usage tracker: Cap API spend per-session (prevent $38K+ BoN/recursive-spawn incidents)
- Enable PII redaction by default in `src/provider/middleware/logging.go`
- Mask `Value` field in `vault.ListCredentials()` output

**0.5 Test Infrastructure** (`src/attacks/testutil/`)
- Create shared test helpers: mock provider, mock logger, test fixture loader
- Table-driven test template for new attack modules
- Mock LLM server (HTTP handler returning configurable responses)
- testdata fixtures with sample payloads and expected results

**Phase 0 Success Criteria:**
- [x] `AttackModule` interface and `Registry` defined at `src/attacks/attack.go`
- [x] `src/attacks/common/` package with all consolidated types
- [x] Optional provider interfaces added without breaking existing code
- [x] All 4 security fixes applied and verified
- [x] `src/attacks/testutil/` with mock provider, logger, and test helpers
- [x] All existing tests still pass
- [x] 11 existing modules migrated to use common types (randInt, generateAttackID, contains)
- Estimated scope: ~8 new/modified Go files, 0 new templates → Actual: 15 files modified/created

### Research Insights: Phase 0

**Best Practices (from Go Security Testing Patterns):**
- Use nuclei-style dual `Request`/`Executer` interfaces for maximum flexibility
- Global registry with `init()` self-registration keeps module code decoupled from registry management
- Factory + middleware chain pattern (already used in `src/provider/`) should be extended to attack modules

**Security (from Security Sentinel):**
- Weak KDF finding is P0 -- bare SHA-256 is trivially brute-forceable for passphrases
- Hardcoded salt means identical passphrases produce identical keys across all installations
- `ListCredentials()` returning full `Value` field allows any authenticated user to exfiltrate all secrets
- No TLS enforcement means provider API keys can be intercepted in transit

**Architecture (from Architecture Strategist):**
- Without shared `AttackModule` interface, 30+ new modules will each reinvent registration, configuration, and result reporting
- Optional interfaces (Go's implicit interface satisfaction) are the right pattern for provider capabilities -- avoids breaking existing Provider implementations
- Type fragmentation is the single biggest codebase debt item; every new module that copies types makes consolidation harder

**References:**
- [ProjectDiscovery nuclei architecture](https://github.com/projectdiscovery/nuclei) -- registry + self-registration pattern
- [Go optional interfaces pattern](https://blog.merovius.de/posts/2017-09-12-diminishing-returns-of-static-typing/) -- implicit satisfaction
- [OWASP Cryptographic Failures](https://owasp.org/Top10/A02_2021-Cryptographic_Failures/) -- KDF requirements

---

#### Phase 1: Core Attack Techniques (High-Impact, Lower Complexity)

New techniques that build on existing infrastructure with minimal new dependencies.

**1.1 Multi-Turn Escalation Patterns**
- **Crescendo** (`src/attacks/orchestration/crescendo.go`)
  - Benign-to-malicious escalation in <5 turns
  - Exploits pattern-following and recency bias in self-generated text
  - Source: [arXiv 2404.01833](https://arxiv.org/abs/2404.01833), USENIX Security 2025
- **Skeleton Key** (`src/attacks/orchestration/skeleton_key.go`)
  - Multi-turn mode-switching to enable direct harmful requests
  - Source: [Microsoft Security Blog](https://www.microsoft.com/en-us/security/blog/2024/06/26/mitigating-skeleton-key-a-new-type-of-generative-ai-jailbreak-technique/)
- **Bad Likert Judge** (`src/attacks/orchestration/bad_likert_judge.go`)
  - Evaluation-capability misuse via Likert scale scoring
  - 60%+ ASR increase over plain prompts
  - Source: [Palo Alto Unit 42](https://unit42.paloaltonetworks.com/multi-turn-technique-jailbreaks-llms/)
- Templates: `templates/crescendo_*.yaml`, `templates/skeleton_key_*.yaml`, `templates/bad_likert_*.yaml`

**1.2 Token & Format Manipulation**
- **MetaBreak** (`src/attacks/evasion/metabreak.go`)
  - 4 primitives: response injection, turn masking, role confusion, context manipulation
  - Targets BOS, EOS, role tokens
  - Outperforms PAP by 11.6% and GPTFuzzer by 34.8%
  - Source: [arXiv 2510.10271](https://arxiv.org/abs/2510.10271)
- **Poetry-Based Attacks** (`src/attacks/evasion/poetry_attacks.go`)
  - Poetic structure framing to bypass safety filters
  - 62% ASR across major LLMs
- **Content Concretization** (`src/attacks/evasion/content_concretization.go`)
  - Iterative abstract-to-concrete transformation
  - Source: GameSec 2025
- **Immersive World Technique** (`src/attacks/evasion/immersive_world.go`)
  - Sophisticated narrative engineering / world-building
  - Source: Cato Networks 2025

**1.3 Long-Context Exploitation**
- **Many-Shot Jailbreaking** (`src/attacks/orchestration/many_shot.go`)
  - Configurable example count (100-10,000+ for Llama 4 Scout's 10M context)
  - Example generation pipeline for in-context learning priming
  - Source: [Anthropic Research](https://www.anthropic.com/research/many-shot-jailbreaking)

**1.4 Statistical Sampling**
- **Best-of-N (BoN)** (`src/attacks/evasion/best_of_n.go`)
  - Augmentation strategies: character scrambling, random capitalization, character noising
  - Configurable N (default 100, up to 10,000)
  - Cross-modal: text, vision, audio
  - 78% ASR on Claude 3.5 Sonnet at N=10,000
  - Source: [arXiv 2412.03556](https://arxiv.org/html/2412.03556v1)

**Phase 1 Success Criteria:**
- [ ] All 9 attack types implemented with Go modules and YAML templates
- [ ] Python template equivalents added for ML pipeline integration
- [ ] Unit tests for each attack module
- [ ] OWASP LLM 2025 mapping updated for new techniques
- Estimated scope: ~15 new Go files, ~9 new YAML templates, ~9 new Python templates

### Research Insights: Phase 1

**Best Practices (from Go Security Testing Patterns):**
- Each attack module should implement the Phase 0 `AttackModule` interface and self-register via `init()`
- Template validation is critical -- CVE-2024-43405 (nuclei signature bypass) exploited YAML parsing inconsistencies
- Use `errgroup.SetLimit()` + semaphore work pool for parallel BoN execution
- Table-driven tests with mock provider for each technique; testdata fixtures for payload/response pairs

**Performance (from Performance Oracle):**
- **Many-Shot (1.3):** 10M-token payloads at 4 bytes/token = 40MB+ raw text per payload. Current `MemoryPoolManager` default is 100MB -- a single many-shot payload risks OOM. **Mitigation:** Stream payload construction (don't buffer full payload in memory); increase `MaxMemoryUsage` to 2GB for many-shot mode.
- **BoN (1.4):** At N=10,000 and $0.01/request, a single BoN run costs $100+. Current usage tracker has no ceiling. **Mitigation:** Add per-run cost ceiling (default $10, configurable) and per-session ceiling (default $50). Add `--max-cost` CLI flag.
- Per-provider rate limiting needed: OpenAI (10K RPM tier 5), Anthropic (4K RPM), Google (1.5K RPM), DeepSeek (variable). Current `distributed_rate_limiter.go` has single `DefaultLimit: 100` for all providers.

**Security (from Security Sentinel):**
- BoN augmentation (random capitalization, character noising) should not be applied to attack payloads containing code -- could break syntax and produce false negatives
- Many-shot example generation pipeline must not use live model outputs as examples (circular dependency, potential for harmful content caching)
- MetaBreak's BOS/EOS token manipulation is provider-specific -- incorrect token IDs produce API errors, not attacks

**Edge Cases:**
- Crescendo multi-turn sessions need cleanup on failure (dangling provider sessions consume quota)
- Poetry-based attacks with non-Latin scripts may trigger different safety filters than English
- Skeleton Key's mode-switching may not work on models without explicit "mode" concept (e.g., Llama 4)

**References:**
- [CVE-2024-43405 nuclei signature bypass](https://github.com/projectdiscovery/nuclei/security/advisories) -- template validation importance
- [Go errgroup patterns](https://pkg.go.dev/golang.org/x/sync/errgroup) -- bounded concurrency
- [Anthropic rate limits](https://docs.anthropic.com/en/api/rate-limits) -- per-provider configuration

---

#### Phase 2: New Attack Surfaces (RAG, MCP, Agentic, Audio)

New attack domains requiring new package structure.

**2.1 RAG Pipeline Poisoning** (`src/attacks/rag/`)
- `poisoned_rag.go` - Document injection attacks (PoisonedRAG pattern)
  - Craft semantically meaningful poisoned texts for RAG databases
  - 5 documents → 90% manipulation
  - Source: [USENIX Security 2025](https://www.usenix.org/system/files/usenixsecurity25-zou-poisonedrag.pdf)
- `vector_embedding_attack.go` - Adversarial embedding generation
  - Craft documents whose vector representations cluster near target queries
  - Exploits high-dimensionality (768/1536-dim) of embedding space
  - Source: [Prompt Security](https://prompt.security/blog/the-embedded-threat-in-your-llm-poisoning-rag-pipelines-via-vector-embeddings)
- `kg_rag_poisoning.go` - Knowledge Graph RAG attacks
  - Insert perturbation triples to create misleading inference chains
  - Source: [ScienceDirect](https://www.sciencedirect.com/science/article/abs/pii/S1566253525009625)
- `indirect_rag_injection.go` - Third-party content injection via web scraping
  - Embed instructions in publicly accessible documents ingested by RAG scrapers
  - Source: [Hidden-in-Plain-Text benchmark (arXiv 2601.10923)](https://arxiv.org/html/2601.10923v1)

**2.2 MCP Protocol Attacks** (`src/attacks/agentic/mcp/`)
- `tool_poisoning.go` - Manipulate MCP tool metadata/descriptions
  - Cause LLM agents to invoke compromised tools via description manipulation
  - Source: [Palo Alto Unit 42](https://unit42.paloaltonetworks.com/model-context-protocol-attack-vectors/)
- `sampling_injection.go` - Prompt injection via MCP sampling API
  - New attack vectors through the sampling protocol extension
  - Source: [Unit 42 MCP sampling research](https://unit42.paloaltonetworks.com/model-context-protocol-attack-vectors/)
- `mcp_supply_chain.go` - Malicious MCP server/package attacks
  - Auth endpoint manipulation (CVE-2025-6514 pattern)
  - Source: [Practical DevSecOps](https://www.practical-devsecops.com/mcp-security-vulnerabilities/)
- `filesystem_escape.go` - MCP filesystem/git sandbox escapes
  - Symlink bypass, path validation bypass, argument injection
  - Source: [Adversa AI](https://adversa.ai/blog/top-mcp-security-resources-february-2026/)

**2.3 AI Browser Agent Attacks** (`src/attacks/agentic/browser/`)
- `hidden_instruction.go` - Hidden prompts in web content
  - Faint text on colored backgrounds, CSS-hidden content, invisible Unicode
  - Source: [Brave/Comet research](https://brave.com/blog/comet-prompt-injection/), [Auth0](https://auth0.com/blog/prompt-injection-ai-browser/)
- `document_injection.go` - Injection via Google Docs, shared documents
  - Exploit AI browsers that read document content
  - Source: [OpenAI Atlas research](https://openai.com/index/hardening-atlas-against-prompt-injection/)
- `screenshot_injection.go` - Prompts embedded in screenshots/images
  - Instructions hidden within images that AI can read but humans can't easily see
  - Source: [Brave red team research](https://brave.com/blog/comet-prompt-injection/)

**2.4 Audio Modality Attacks** (`src/attacks/audio/`)
- `audio_jailbreak.go` - Adversarial audio perturbation attacks
  - Asynchronous, universal, stealthy audio perturbations
  - 89% ASR across restricted tasks
  - Source: [AudioJailbreak (arXiv 2505.14103)](https://arxiv.org/abs/2505.14103)
- `speech_model_exploit.go` - White-box audio-language model attacks
  - Target SpeechGPT and similar end-to-end models
  - Source: [arXiv 2505.18864](https://arxiv.org/abs/2505.18864)
- `multilingual_audio.go` - Non-English audio bypass
  - 3.1x higher success rate than text-only attacks
  - Source: [JALMBench (arXiv 2505.17568)](https://www.arxiv.org/pdf/2505.17568)
- `bon_audio.go` - Best-of-N audio augmentation sampling
  - Black-box algorithm for audio jailbreaks
  - Source: [OpenReview](https://openreview.net/forum?id=yougZBoUY3)

**Phase 2 Success Criteria:**
- [ ] 4 new attack packages (`rag/`, `agentic/mcp/`, `agentic/browser/`, `audio/`)
- [ ] 15 new Go attack modules
- [ ] Integration tests for each new attack surface
- [ ] Python template equivalents for ML pipeline
- [ ] Documentation for each new attack category
- Estimated scope: ~20 new Go files, ~15 new YAML templates

### Research Insights: Phase 2

**Best Practices (from OWASP Agentic Research):**
- RAG attacks should include both retrieval-time poisoning (embedding manipulation) and generation-time poisoning (context injection)
- MCP attacks: Test all 4 MCP transport types (stdio, SSE, HTTP, WebSocket) -- attack surfaces differ by transport
- Browser agent attacks should test CSS-hidden, font-color-matched, zero-width Unicode, and HTML comment injection vectors
- Audio attacks require dedicated worker pool with separate memory/CPU limits (audio processing is CPU-intensive)

**Performance (from Performance Oracle):**
- Audio payload processing (WAV/MP3 manipulation) requires dedicated goroutine pool -- don't share with text attack workers
- RAG embedding generation (768/1536-dim vectors) is computationally expensive -- batch embed operations, cache results
- MCP tool enumeration should be cached per-server (tools don't change mid-session)
- Browser agent attacks may need headless browser dependency -- make this optional, not a core requirement

**Security (from Security Sentinel):**
- `filesystem_escape.go` creates self-risk: the tool implementing filesystem escape attacks could itself be exploited via the same vectors. **Mitigation:** Run filesystem escape tests in isolated sandbox (Docker or chroot), never against the host filesystem.
- `mcp_supply_chain.go` involves downloading/executing untrusted MCP servers. **Mitigation:** Network isolation, read-only filesystem mounts, timeout enforcement.
- Audio deps (if using Python bridge for audio processing) have supply chain risk -- pin versions, verify checksums.
- China-based provider endpoints (DeepSeek) used in RAG embedding attacks have data residency implications -- document in README.

**Architecture (from Architecture Strategist):**
- `src/attacks/agentic/` subtree is 3 levels deep (`agentic/mcp/`, `agentic/browser/`) vs. flat pattern everywhere else. Consider flattening to `src/attacks/mcp/` and `src/attacks/browser/` for consistency, or accept the nesting as intentional grouping.
- RAG attacks need a dual-provider pattern: one provider for embedding generation, another as the attack target. The `AttackConfig` should support `EmbeddingProvider` and `TargetProvider` fields.

**Agent-Native (from Agent-Native Reviewer):**
- Individual attack execution API endpoint needed (currently only scan-level at `src/api/router.go:249-253`). Without this, agents cannot invoke specific attack techniques programmatically.
- Attack results should include structured feedback fields: `success_indicators []string`, `confidence float64`, `suggested_followup string` for agent-driven adaptation loops.

**OWASP Mapping (from OWASP Research):**
- RAG poisoning maps to ASI04 (Supply Chain) AND ASI06 (Memory Poisoning) -- dual mapping needed
- MCP attacks map to ASI02 (Tool Misuse) AND ASI04 (Supply Chain) AND ASI05 (Code Execution) -- triple mapping
- Browser agent attacks primarily map to ASI01 (Goal Hijack) with secondary ASI09 (Trust Exploitation)
- Audio attacks currently have NO ASI mapping -- propose ASI01 (Goal Hijack via audio prompt injection)

**References:**
- [Hidden-in-Plain-Text RAG benchmark](https://arxiv.org/html/2601.10923v1) -- standardized RAG attack evaluation
- [DeepTeam RAG testing](https://docs.confident-ai.com/docs/red-teaming-rag) -- existing RAG testing tool
- [Promptfoo OWASP Agentic plugins](https://www.promptfoo.dev/docs/red-team/owasp-agentic/) -- integration opportunity
- [MAESTRO 7-layer threat model](https://cloudsecurityalliance.org/research/artifacts/maestro-multi-agent-security-threat-model) -- CSA framework for agentic AI

---

#### Phase 3: Reasoning Model Exploitation & Adaptive Attacks

Advanced techniques requiring more sophisticated implementation.

**3.1 Reasoning Model Attacks** (`src/attacks/reasoning/`)
- `autonomous_jailbreak.go` - Reasoning-model-as-attacker
  - Use DeepSeek-R1, Gemini 2.5 Flash, Grok 3 Mini as autonomous adversaries
  - 97.14% overall jailbreak success
  - Source: [Nature Communications](https://www.nature.com/articles/s41467-026-69010-1)
- `cot_exploitation.go` - Chain-of-thought manipulation
  - Exploit visible CoT reasoning transparency
  - Manipulate step-by-step thinking similar to phishing
  - Source: [Trend Micro](https://www.trendmicro.com/en_us/research/25/c/exploiting-deepseek-r1.html), [HiddenLayer](https://hiddenlayer.com/innovation-hub/deepsht-exposing-the-security-risks-of-deepseek-r1)
- `reasoning_loop_exploit.go` - Infinite reasoning loops and resource exhaustion
  - Target reasoning models' extended thinking time

**3.2 Tool-Use Interface Attacks** (`src/attacks/agentic/tool_use/`)
- `imist.go` - Transform harmful queries into tool invocations
  - Exploit function-calling interface
  - Source: [arXiv 2601.05466](https://arxiv.org/html/2601.05466v1)
- `agent_exploitation.go` - AI coding editor attacks (AIShellJack pattern)
  - 75-88% execution rate, 71.5% privilege escalation
  - Source: [IEEE S&P 2026](https://arxiv.org/html/2511.05797v1)

**3.3 Adaptive Defense Bypass** (`src/attacks/adaptive/`)
- `defense_bypass_optimizer.go` - Systematic defense bypass optimization
  - Gradient descent, RL, random search, human-guided exploration
  - Target: CaMel (67% neutralization), LLM Salting (3% ASR), and other defenses
  - Source: OpenAI/Anthropic/DeepMind joint research 2025
- `camel_bypass.go` - CaMeL-specific bypass techniques
  - Target capability-based access control limitations
  - Exploit policy definition gaps
  - Source: [arXiv 2503.18813](https://arxiv.org/abs/2503.18813)
- `salt_resistance.go` - Anti-salting techniques
  - Dynamic attack generation that doesn't rely on static pre-computed prompts
  - Target per-deployment fine-tuning variations
  - Source: [Sophos CAMLIS 2025](https://news.sophos.com/en-us/2025/10/21/getting-salty-with-llms-sophosai-unveils-new-defense-against-jailbreaking-at-camlis-2025/)

**3.4 DiffusionAttacker** (`src/attacks/adaptive/diffusion_attacker.go`)
- Leverage diffusion models to iteratively refine adversarial prompts
- Source: EMNLP 2025
- Note: May require optional Python ML dependency for diffusion model

**Phase 3 Success Criteria:**
- [ ] 9 new Go attack modules
- [ ] Reasoning model orchestration working with at least 2 provider APIs
- [ ] Adaptive optimizer framework with pluggable strategies
- [ ] Defense-specific bypass modules for CaMel and LLM Salting
- Estimated scope: ~12 new Go files, ~10 new YAML templates

### Research Insights: Phase 3

**Best Practices (from Go Security Testing Patterns):**
- Autonomous jailbreak module (`autonomous_jailbreak.go`) creates an uncontrolled feedback loop: a reasoning model generates attacks, evaluates results, and refines -- potentially indefinitely. **Mandatory:** Human-in-the-loop gating. Implement `--max-autonomous-turns` (default 10), require explicit `--allow-autonomous` flag, and log every generated payload for audit.
- DiffusionAttacker requires Python ML dependency (diffusion model). Implement as optional Python bridge via `os/exec` subprocess, NOT as a core Go dependency. Use the existing Go/Python bridge pattern from `ml/`.
- Defense bypass optimizer needs convergence detection: if success rate hasn't improved in N iterations, stop. Also needs wall-time limit (default 30 minutes) to prevent runaway optimization.

**Performance (from Performance Oracle):**
- Reasoning model API calls have 10-100x longer latency than standard chat completions (extended thinking time). **Mitigation:** Separate timeout configuration for reasoning providers; don't count reasoning time against rate limits.
- Recursive sub-agent spawning (`recursive_spawn_abuse.go` in Phase 6 uses similar patterns) without limits caused $38K LangGraph incident. **Mitigation:** Agent tree depth limit (default 3), agent tree breadth limit (default 10), total agent count limit (default 50).
- Adaptive optimizer with gradient descent/RL strategies is CPU-intensive. Run optimization in dedicated goroutine pool with `runtime.LockOSThread()` for consistent performance.

**Security (from Security Sentinel):**
- `autonomous_jailbreak.go` is the highest-risk module in the entire plan. A reasoning model generating novel jailbreaks could produce genuinely harmful content that gets cached in attack data pipeline. **Mitigations:**
  1. All generated payloads must be logged to audit trail before execution
  2. Output filtering: reject generated payloads that match known harmful content patterns
  3. Session isolation: each autonomous run gets its own SQLite database, not shared with main attack DB
  4. Rate limiting: max 1 autonomous session at a time per installation
- `cot_exploitation.go` reads reasoning traces which may contain PII or sensitive information from the model's training data. Ensure PII redaction is applied to CoT output.
- `defense_bypass_optimizer.go` could be misused to optimize attacks against production systems. Add prominent warning, require `--i-understand-risks` flag, and log target information.

**Architecture (from Architecture Strategist):**
- Phase 3 requires the attack composition/chaining API that Phase 0 defines. Specifically, `defense_bypass_optimizer.go` needs to compose arbitrary attack techniques, evaluate results, and iterate -- this is the `AttackPipeline` pattern.
- Reasoning model attacks need access to the `ReasoningProvider` optional interface from Phase 0.4. Without it, there's no way to access CoT traces.
- The adaptive optimizer framework should use the Strategy pattern with pluggable search strategies (gradient descent, RL, random search, human-guided). Each strategy implements `NextCandidate(history) AttackConfig`.

**Edge Cases:**
- Reasoning models may refuse to generate attack payloads even when instructed to do so for security testing. The autonomous jailbreak module should handle refusals gracefully and not count them as "failures" in the optimization loop.
- CoT exploitation requires models that expose thinking traces. As of Feb 2026, only DeepSeek-R1, Gemini 2.5, and Grok 3 expose this. Claude and GPT have internal CoT that is not accessible.
- CaMeL bypass targeting capability-based access control assumes the target uses CaMeL. Need capability detection before attempting bypass.

**References:**
- [LangGraph $38K incident](https://arxiv.org/html/2510.23883v1) -- recursive spawning cost explosion
- [CaMel capability-based defense](https://arxiv.org/abs/2503.18813) -- understanding what to bypass
- [DeepTeam reasoning exploitation](https://docs.confident-ai.com/docs/red-teaming-reasoning) -- existing tooling

---

#### Phase 4: Model Adapters & Provider Updates

Update provider infrastructure for new model families.

**4.1 New Provider Adapters** (`src/provider/`)
- Update OpenAI adapter for GPT-5/5.1/5.2 (400K context, adaptive reasoning)
- Update Anthropic adapter for Claude 4.5 Sonnet/Opus (agent stability, tool use)
- Update Google adapter for Gemini 3 Pro ("Deep Think" mode)
- Add Meta/Llama adapter for Llama 4 Scout/Maverick (10M context, MoE)
- Add DeepSeek adapter for V3.2 (sparse attention, MIT license)
- Add Alibaba adapter for Qwen3 235B (reasoning model)
- Add xAI adapter for Grok 3/3 Mini (reasoning capabilities)

**4.2 Model-Specific Attack Profiles** (`templates/model_profiles/`)
- Reasoning model profiles (R1, o1, o3, Gemini Deep Think, Qwen3)
  - CoT exploitation sequences, reasoning manipulation patterns
- Long-context profiles (Llama 4 Scout 10M, GPT-5.2 400K)
  - Many-shot amplification calibrated to context window size
- MoE profiles (Llama 4, DeepSeek V3)
  - Expert routing manipulation patterns
- Agentic profiles (Claude 4.5, GPT-5)
  - Tool-use exploitation, function calling attack sequences

**4.3 Benchmark Success Rate Database**
- Store per-model expected success rates from published benchmarks:
  - Claude: 42.8% WASR (JailbreakBench)
  - GPT-5: 55.9% WASR
  - Gemini: 59.5% WASR
  - Per-technique ASR data from Nature Communications, Unit 42, Anthropic research

**Phase 4 Success Criteria:**
- [ ] 7 provider adapters updated/created
- [ ] Model-specific attack profiles for each model family
- [ ] Benchmark database populated with published success rates
- [ ] Python harness updated to test against new models
- Estimated scope: ~10 modified Go files, ~7 new profile YAML files

### Research Insights: Phase 4

**Best Practices (from Go Security Testing Patterns):**
- Provider adapters should use the factory + middleware chain pattern already established in `src/provider/`
- Hierarchical rate limiting: global limiter + per-provider limiter using `golang.org/x/time/rate`. Current `distributed_rate_limiter.go` has single `DefaultLimit: 100` -- needs per-provider config:
  - OpenAI: 10,000 RPM (tier 5)
  - Anthropic: 4,000 RPM
  - Google: 1,500 RPM
  - DeepSeek: variable (often lower)
  - Qwen (Alibaba): TBD (check at implementation time)
  - xAI (Grok): TBD
- Model-specific attack profiles should be YAML-driven, not Go code. Load profiles from `templates/model_profiles/*.yaml` rather than hardcoding in `profile_*.go` files. This allows users to add custom profiles without recompiling.

**Performance (from Performance Oracle):**
- Benchmark database writes should use batch inserts with proper indexing on `(model, technique, timestamp)` for efficient querying
- Provider response caching (three-tier strategy):
  1. **Provider response cache**: Cache identical API calls (same messages + config) -- TTL 1 hour
  2. **Augmentation cache**: Cache BoN augmentation results (same base payload + augmentation params) -- TTL 24 hours
  3. **Success rate cache**: Cache per-model per-technique success rates -- TTL 7 days
- Current `redis_cluster_cache.go` `MaxValueSize` of 1MB will be breached by long-context model responses. Increase to 10MB for long-context providers.

**Security (from Security Sentinel):**
- China-based provider endpoints (DeepSeek, Qwen/Alibaba) have data residency implications:
  - Attack payloads sent to these endpoints may be subject to Chinese data regulations
  - Document prominently in README and provider configuration
  - Consider adding `--data-residency-warning` flag that requires acknowledgment
- Alibaba Cloud SDK has had supply chain incidents. Pin exact versions, verify checksums.
- API keys for 7 providers increase the credential attack surface. Ensure all provider configs use the vault system from Phase 0.4.

**Architecture (from Architecture Strategist):**
- Provider adapters need capability detection: query the model and return supported `ModelCapability` flags. This enables attack modules to auto-select techniques based on model capabilities (e.g., skip audio attacks for text-only models).
- Go/Python parity tiers should be explicit:
  - **Template parity**: Static payloads work identically in both. YAML templates shared.
  - **API parity**: Stateful attacks (multi-turn, adaptive) available via Go API and Python bridge.
  - **Python-only**: ML-heavy features (DiffusionAttacker, bandit optimizer training) remain Python-only.

**Edge Cases:**
- GPT-5.2's 400K context window and Llama 4 Scout's 10M context require streaming payload construction -- cannot buffer full payload in memory
- MoE model expert routing is not exposed via standard APIs -- expert routing manipulation attacks may need white-box access (local model only)
- Benchmark success rates from published papers use specific evaluation criteria that may not match LLMrecon's success detection. Document the evaluation methodology difference.

**References:**
- [golang.org/x/time/rate](https://pkg.go.dev/golang.org/x/time/rate) -- hierarchical rate limiting
- [Provider rate limit documentation](https://platform.openai.com/docs/guides/rate-limits) -- per-provider limits
- [Alibaba Cloud SDK security](https://github.com/aliyun/alibaba-cloud-sdk-go/security) -- supply chain considerations

---

#### Phase 5: OWASP Compliance & Documentation

**5.1 OWASP Agentic Top 10 2026 Mapping**
- Add new compliance mapping alongside existing LLM Top 10 2025
- Map each new attack technique to applicable ASI categories:
  - ASI01 (Goal Hijack) ← Crescendo, Skeleton Key, Bad Likert Judge, Browser agent attacks
  - ASI02 (Tool Misuse) ← iMIST, MCP tool poisoning, Agent exploitation
  - ASI03 (Privilege Abuse) ← Agent exploitation (71.5% privilege escalation), MCP supply chain
  - ASI04 (Supply Chain) ← MCP supply chain, RAG poisoning
  - ASI05 (Code Execution) ← MCP filesystem escape, Agent exploitation
  - ASI06 (Memory Poisoning) ← RAG poisoning, Many-shot, Context manipulation
  - ASI07 (Inter-Agent Comms) ← MCP sampling injection, Inter-agent attacks
  - ASI08 (Cascading Failures) ← Reasoning loops, Cascading agent failures
  - ASI09 (Trust Exploitation) ← Bad Likert Judge, Immersive World
  - ASI10 (Rogue Agents) ← Autonomous reasoning model jailbreaking

**5.2 Documentation Updates**
- Update CLAUDE.md with new attack categories
- Update `harness_config.json` with new technique entries
- Create technique reference cards for each new attack
- Update OWASP compliance report generator (`cmd/compliance-report/`)

**5.3 Test Coverage**
- Unit tests for all new modules
- Integration tests for multi-turn attack flows
- Provider integration tests for new model adapters
- OWASP compliance validation tests

**Phase 5 Success Criteria:**
- [ ] Full OWASP Agentic 2026 mapping complete
- [ ] All attack techniques mapped to both LLM 2025 and Agentic 2026
- [ ] Documentation updated
- [ ] Test coverage >80% for new code
- Estimated scope: ~5 modified Go files, ~3 new test files, documentation updates

### Research Insights: Phase 5

**Best Practices (from OWASP Agentic Research):**
- **70+ concrete test cases** across ASI01-ASI10 (7-8 per category) identified from existing tools and benchmarks:
  - ASI01 (Goal Hijack): Indirect prompt injection via retrieved docs, system prompt override, context window manipulation, role confusion, attention steering, task redirection, benign-to-malicious escalation
  - ASI02 (Tool Misuse): Unvalidated tool parameters, tool description manipulation, excessive tool permissions, chained tool exploitation, tool output injection, shadow tool invocation, parameter type confusion
  - ASI03 (Privilege Abuse): Horizontal privilege escalation, vertical privilege escalation, delegation chain abuse, permission inheritance exploitation, cross-tenant access, API key scope abuse, session privilege persistence
  - ASI04 (Supply Chain): Malicious plugin injection, dependency confusion, update channel poisoning, marketplace trust abuse, transitive dependency exploitation, build pipeline injection, signed package impersonation
  - ASI05 (Code Execution): Sandbox escape, command injection via tool, code generation exploitation, eval injection, deserialization attacks, container breakout, file system access via code
  - ASI06 (Memory Poisoning): RAG data poisoning, conversation history manipulation, persistent memory injection, context window pollution, embedding space attacks, memory retrieval manipulation, cross-session contamination
  - ASI07 (Inter-Agent Comms): Message tampering, agent impersonation, protocol downgrade, shared state poisoning, broadcast injection, relay manipulation, schema exploitation
  - ASI08 (Cascading Failures): Recursive loop induction, resource exhaustion chain, error propagation amplification, timeout cascade, deadlock induction, state corruption propagation, monitoring evasion during cascade
  - ASI09 (Trust Exploitation): Implicit trust assumption abuse, trust boundary confusion, trust transitivity exploitation, social engineering via agent, authority impersonation, compliance theater, trust score manipulation
  - ASI10 (Rogue Agents): Deceptive alignment detection, goal misalignment, reward hacking, emergent collusion, sandbox escape with deception, monitoring evasion, alignment faking

- **MITRE ATLAS cross-reference**: Map each ASI test case to ATLAS tactics/techniques. October 2025 ATLAS update added 14 agent-specific techniques. Structure:
  ```yaml
  test_case:
    id: ASI01-TC01
    title: "Indirect prompt injection via retrieved document"
    asi_category: ASI01
    atlas_tactics: [AML.TA0005, AML.TA0043]
    atlas_techniques: [AML.T0051, AML.T0054]
    maestro_layers: [L5-Agent, L6-Ecosystem]
  ```

- **Existing tools to integrate rather than reimplement**:
  - **DeepTeam** (Confident AI): Already implements 40+ test cases for RAG, tool use, and agent safety. Consider integration via Python bridge.
  - **Promptfoo**: Has OWASP Agentic plugins since Dec 2025. Consider YAML template compatibility.
  - **Agent Security Bench** (ICLR 2025): 10 agent-to-agent scenarios with ASI mapping. Consider importing test scenarios.
  - **AgentThreatBench** (UK AI Safety Institute): Agent-specific threat benchmark. Reference for evaluation methodology.

**Performance (from Performance Oracle):**
- OWASP compliance report generation should be incremental -- don't regenerate the full report for each test run. Cache per-technique results and only regenerate the summary.
- Test coverage validation should run in CI, not at test execution time (expensive).

**Architecture (from Architecture Strategist):**
- OWASP mapping should be bidirectional: given an ASI category, list all applicable attack techniques; given a technique, list all mapped ASI categories. Implement as a lookup table in YAML, not code.
- Compliance report generator at `cmd/compliance-report/` needs update for the dual-framework (LLM 2025 + Agentic 2026) requirement. Consider a unified report format with sections for each framework.

**References:**
- [OWASP Agentic Top 10 2026](https://genai.owasp.org/resource/owasp-top-10-for-agentic-applications-for-2026/) -- official framework
- [MITRE ATLAS October 2025 update](https://atlas.mitre.org/) -- 14 new agent-specific techniques
- [MAESTRO threat model](https://cloudsecurityalliance.org/research/artifacts/maestro-multi-agent-security-threat-model) -- 7-layer model
- [Agent Security Bench](https://proceedings.iclr.cc/paper_files/paper/2025/file/5750f91d8fb9d5c02bd8ad2c3b44456b-Paper-Conference.pdf) -- ICLR 2025
- [DeepTeam documentation](https://docs.confident-ai.com/docs/red-teaming) -- integration reference

---

#### Phase 6: Multi-Agent & Orchestrator Exploitation

Attacks targeting LLM-controlling-LLM architectures (OpenClaw, CrewAI, LangGraph, AutoGen, etc.) where an orchestrator agent delegates tasks to sub-agents with varying privilege levels.

**6.1 Inter-Agent Delegation Attacks** (`src/attacks/agentic/multi_agent/`)
- `delegation_escalation.go` - Exploit trust boundaries between orchestrator and sub-agents
  - Feed low-privilege agent a malformed request that tricks higher-privilege agent into unauthorized actions
  - Real-world: ServiceNow AI assistant privilege bypass via delegation
  - Source: [ScienceDirect - Protocol Exploits](https://www.sciencedirect.com/science/article/pii/S2405959525001997)
- `toxic_agent_flow.go` - Chain prompt injections across agent communication channels
  - Poisoned output from one agent becomes trusted input for the next
  - "Toxic Agent Flow" exploit pattern from GitHub MCP servers
  - Source: [arXiv 2506.23260](https://arxiv.org/abs/2506.23260)
- `recursive_spawn_abuse.go` - Exploit recursive sub-agent spawning for resource exhaustion
  - Real-world: LangGraph incident burned $38K in 4 hours via recursive spawning
  - Source: [Multi-Agent Security Survey](https://arxiv.org/html/2510.23883v1)

**6.2 Malicious Skill/Plugin Injection** (`src/attacks/agentic/skill_injection/`)
- `skill_poisoning.go` - Upload malicious skills/plugins to agent marketplaces
  - OpenClaw: ~900 malicious skills found (20% of total packages), auto-deployed via scripts
  - Skills inherit system-wide permissions once loaded
  - Source: [Kaspersky OpenClaw Analysis](https://www.kaspersky.com/blog/openclaw-vulnerabilities-exposed/55263/)
- `skill_takeover_chain.go` - Chain skill loading → persistent compromise
  - Indirect prompt injection via URL summarization → update config files → permanent compromise
  - Source: [Penligent OpenClaw Research](https://www.penligent.ai/hackinglabs/the-openclaw-prompt-injection-problem-persistence-tool-hijack-and-the-security-boundary-that-doesnt-exist/)

**6.3 Agent Persistence & Hijack** (`src/attacks/agentic/persistence/`)
- `agent_config_rewrite.go` - Exploit agents that can modify their own configuration
  - Inject instructions that cause the agent to rewrite its system prompt or config files
  - Once rewritten, compromise persists across sessions
  - Source: [Penligent](https://www.penligent.ai/hackinglabs/the-openclaw-prompt-injection-problem-persistence-tool-hijack-and-the-security-boundary-that-doesnt-exist/)
- `credential_harvest.go` - Extract API keys, OAuth tokens, SSH credentials from agents
  - OpenClaw: agents inherit access to API keys, OAuth tokens, SSH creds, browser sessions
  - Source: [Cubic Security Audit](https://www.cubic.dev/blog/we-found-and-fixed-critical-security-vulnerabilities-in-openclaw)
- `rce_chain.go` - Multi-step exploitation chains (prompt injection → tool access → RCE)
  - 1-click RCE: single webpage visit → full host takeover via agent
  - Source: [depthfirst CVE-2026-25253](https://depthfirst.com/post/1-click-rce-to-steal-your-moltbot-data-and-keys)

**6.4 Deceptive Agent Alignment** (`src/attacks/agentic/deception/`)
- `deceptive_alignment.go` - Agent appears aligned but acts against objectives
  - Detection challenges even for dedicated "ObserverAI" monitoring agents
  - Source: [Agentic AI Security Survey](https://arxiv.org/html/2510.23883v1)
- `agent_collusion.go` - Multiple compromised agents cooperating against the orchestrator
  - Emergent collusion behavior in multi-agent systems
  - Maps to OWASP ASI10 (Rogue Agents)

**6.5 Framework-Specific Attack Profiles**
- `profile_openclaw.go` - OpenClaw-specific attack sequences
  - 512 known vulnerabilities, 8 critical, ~10K+ exposed instances
  - Sub-agent queue lane isolation bypass
  - Source: [SiliconANGLE](https://siliconangle.com/2026/02/09/tens-thousands-openclaw-systems-exposed-due-misconfiguration-known-exploits/), [VentureBeat](https://venturebeat.com/security/openclaw-agentic-ai-security-risk-ciso-guide/)
- `profile_crewai.go` - CrewAI-specific attacks (no RBAC, task-scoping limitations)
- `profile_langgraph.go` - LangGraph-specific attacks (state machine manipulation)
- `profile_autogen.go` - AutoGen-specific attacks (Docker sandbox escape)

**Phase 6 Assessment: Fits within LLMrecon**

This belongs in LLMrecon (not a separate tool) because:
1. Core attack primitives are prompt injection and trust boundary violations — LLMrecon's domain
2. OWASP Agentic Top 10 2026 (Phase 5) directly maps to these scenarios
3. Existing `src/attacks/agentic/` package provides natural home
4. Same ML pipeline can optimize multi-agent attack strategies

**What would NOT fit** (separate tool territory):
- Infrastructure scanning of deployed OpenClaw/CrewAI instances for misconfigurations
- Network-level interception of inter-agent protocols
- Automated malicious skill deployment to marketplaces (offensive, not defensive testing)

**Phase 6 Success Criteria:**
- [ ] 11 new Go attack modules across 4 sub-packages
- [ ] Framework-specific profiles for OpenClaw, CrewAI, LangGraph, AutoGen
- [ ] Integration with OWASP Agentic 2026 mapping (Phase 5)
- [ ] Delegation escalation and toxic agent flow demonstrated in test harness
- Estimated scope: ~15 new Go files, ~8 new YAML templates

### Research Insights: Phase 6

**Best Practices (from OWASP Agentic Research):**
- Framework-specific profiles should be YAML-driven, not Go code. Replace proposed `profile_openclaw.go` etc. with `templates/framework_profiles/openclaw.yaml`. Benefits:
  - Users can add custom framework profiles without recompiling
  - Profiles can be updated independently of code releases
  - YAML profiles are testable via template validation (Phase 5)
- Agent collusion detection is an active research area with no mature solutions. Implement detection heuristics (anomalous inter-agent message patterns, unexpected tool invocations) rather than claiming formal guarantees.
- Multi-agent attacks require orchestration of multiple LLM sessions simultaneously. This is fundamentally different from single-model attacks and needs its own execution engine (not reusing existing `MultiTurnOrchestrator`).

**Performance (from Performance Oracle):**
- Recursive sub-agent spawning attacks (`recursive_spawn_abuse.go`) need strict resource limits:
  - Agent tree depth limit: default 3 (prevents exponential fan-out)
  - Agent tree breadth limit: default 10 (caps agents per level)
  - Total agent count limit: default 50 (hard ceiling)
  - Per-agent memory limit: 256MB (prevents single-agent OOM)
  - Session cost ceiling: $50 (prevents cost bomb scenarios)
- Multi-agent attack sessions are long-running (minutes, not seconds). Need streaming results, progress callbacks, and graceful cancellation.
- Framework-specific attack profiles may need framework SDKs as optional dependencies. Use lazy loading -- don't import OpenClaw SDK unless OpenClaw profile is selected.

**Security (from Security Sentinel):**
- `credential_harvest.go` tests for API key extraction. The test payloads themselves should NOT contain real API keys. Use synthetic/placeholder keys in test fixtures.
- `rce_chain.go` implements actual RCE chain exploitation. **This is the highest blast-radius module.** Mitigations:
  1. Must run in Docker container with `--network=none`
  2. Filesystem must be read-only except `/tmp`
  3. No access to host process namespace
  4. Automatic cleanup of any spawned processes on test completion
- `skill_poisoning.go` involves creating malicious skills. Ensure test skills are clearly marked as test artifacts and cannot accidentally be deployed.

**Agent-Native (from Agent-Native Reviewer):**
- Multi-turn orchestration is currently inaccessible to external agents (runs as internal goroutine in `MultiTurnOrchestrator.StartSession()`). For Phase 6, agents need to:
  1. Start a multi-agent attack session via API
  2. Observe inter-agent messages in real-time (WebSocket/SSE)
  3. Inject messages into the agent communication channel
  4. Terminate sessions programmatically
- Attack composition API needed: combine Phase 1 techniques with Phase 6 orchestration (e.g., Crescendo escalation within a delegation attack).
- Consider MCP server interface for LLMrecon itself: expose attack capabilities as MCP tools that other agents can invoke.

**Architecture (from Architecture Strategist):**
- Phase 6 depends heavily on Phase 0 (shared `AttackModule` interface) and Phase 3 (reasoning model integration). Do not start Phase 6 until both are complete.
- `src/attacks/agentic/` subtree will have 4 sub-packages after Phase 6: `mcp/`, `browser/`, `multi_agent/`, `skill_injection/`, `persistence/`, `deception/`, `tool_use/`. This is 7 sub-packages under one parent -- consider whether this is intentional grouping or excessive nesting.
- Deceptive alignment detection (`deceptive_alignment.go`) is research-grade, not production-ready. Label it as "experimental" and gate behind `--experimental` flag.

**References:**
- [OpenClaw 512 vulnerabilities analysis](https://siliconangle.com/2026/02/09/tens-thousands-openclaw-systems-exposed-due-misconfiguration-known-exploits/) -- vulnerability landscape
- [LangGraph recursive spawning incident](https://arxiv.org/html/2510.23883v1) -- cost explosion precedent
- [Multi-agent defense pipeline](https://arxiv.org/html/2509.14285v4) -- defensive patterns
- [AgentThreatBench](https://www.gov.uk/government/publications/agentthreatbench) -- UK AI Safety Institute benchmark

---

## Missing Attack Categories (Identified by Security Review)

Six attack categories not covered in Phases 1-6 that should be added:

1. **Structured Output / JSON Mode Exploitation** -- Manipulate JSON schema constraints to bypass safety filters. Models forced to output valid JSON may comply with harmful requests to satisfy schema requirements. Add to Phase 1.
2. **System Prompt Extraction / Leakage** -- Techniques to extract the system prompt from deployed models. Already partially covered by existing extraction module but needs dedicated techniques for newer models. Add to Phase 1.
3. **Fine-Tuning / Alignment Tax Attacks** -- Exploit fine-tuned models where safety training has been diluted by domain-specific fine-tuning. Source: [arXiv 2310.03693](https://arxiv.org/abs/2310.03693). Add to Phase 3.
4. **Embedding Space Adversarial Attacks** -- Craft inputs in embedding space (not token space) that steer model behavior. Requires white-box access. Add to Phase 2 (RAG section, since embeddings are involved).
5. **Multimodal Fusion Confusion** -- Exploit inconsistencies between how models process text vs. image vs. audio inputs. Contradictory information across modalities causes unpredictable behavior. Add to Phase 2 (Audio section).
6. **Model Merging / Mixture Exploitation** -- Target vulnerabilities in merged/mixture models (e.g., Llama 4 Maverick's 128-expert MoE). Expert routing manipulation, dormant expert activation. Add to Phase 4 (MoE profiles).

## Alternative Approaches Considered

1. **Python-only implementation**: Rejected because the Go enterprise component needs parity with Python ML component for production use
2. **External tool integration for audio attacks**: Considered using external audio processing tools, but decided to implement within Go for consistency, with optional Python ML bridge for diffusion models
3. **Single monolithic update**: Rejected in favor of phased approach to maintain stability and allow incremental testing

## Acceptance Criteria

### Functional Requirements
- [ ] All 19 new attack categories + 11 multi-agent modules implemented and testable
- [ ] Framework-specific attack profiles for OpenClaw, CrewAI, LangGraph, AutoGen
- [ ] 7 provider adapters updated for new model families
- [ ] OWASP Agentic Top 10 2026 compliance mapping complete
- [ ] Python harness updated with new templates
- [ ] Go enterprise component updated with new attack modules

### Non-Functional Requirements
- [ ] No regression in existing 130+ attack technique tests
- [ ] Build passes `go vet`, `go test ./...`, and gosec clean
- [ ] No new code scanning alerts introduced
- [ ] Attack modules follow existing plugin architecture patterns

### Quality Gates
- [ ] Unit test coverage >80% for new code
- [ ] All YAML templates validated against schema
- [ ] OWASP mapping reviewed for completeness
- [ ] Provider adapters tested against at least one model each

## Success Metrics

- Number of new attack techniques: 19 categories + 11 multi-agent modules, ~50 specific variants
- OWASP coverage: Both LLM Top 10 2025 and Agentic Top 10 2026 fully mapped
- Model coverage: 7 new model families supported
- Defense bypass: CaMel and LLM Salting bypass modules functional

## Dependencies & Prerequisites

- Existing codebase on branch `fix/codeql-go-build` with all 695 security alerts resolved
- **Phase 0 must complete before Phase 1 begins** (shared interface, common types, security fixes)
- **Phases 0 + 1 + 2 must complete before Phase 3** (reasoning attacks depend on provider extensions and attack composition)
- **Phases 0-5 must complete before Phase 6** (multi-agent attacks compose techniques from all prior phases)
- Go 1.23+ toolchain (per `go.mod`)
- Python 3.10+ with numpy, pandas, requests, rich
- Access to provider APIs for integration testing (OpenAI, Anthropic, Google, DeepSeek, xAI, Alibaba)
- Optional: Ollama for local model testing
- Optional: Docker for sandboxed execution of filesystem escape and RCE chain tests
- Optional: Redis cluster for distributed rate limiting and caching (production scale)
- New Go dependencies: `golang.org/x/time/rate` (hierarchical rate limiting), `golang.org/x/crypto/argon2` (KDF fix)

## Risk Analysis & Mitigation

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| **CRITICAL: BoN cost bomb** | Very High | High | Per-run cost ceiling ($10 default), per-session ceiling ($50), `--max-cost` CLI flag |
| **CRITICAL: Autonomous jailbreak feedback loop** | Very High | Medium | Human-in-the-loop gating, `--max-autonomous-turns 10`, `--allow-autonomous` flag, audit logging |
| **CRITICAL: Recursive spawn DoS** | Very High | Medium | Agent tree depth limit (3), breadth limit (10), total count limit (50), session cost ceiling |
| **CRITICAL: Filesystem escape self-risk** | High | Medium | Run filesystem escape tests in Docker sandbox with `--network=none`, read-only FS |
| **HIGH: Credential leakage via vault** | High | Medium | Mask `Value` in `ListCredentials()`, fix weak KDF (SHA-256 → argon2id), remove hardcoded salt |
| **HIGH: Type fragmentation blocks scaling** | High | High | Phase 0 prerequisite: consolidate 6 duplicated types before adding 30+ modules |
| **HIGH: Many-shot memory exhaustion** | High | Medium | Streaming payload construction, increase `MaxMemoryUsage` 100MB → 2GB |
| Provider API changes break adapters | High | Medium | Abstract provider interface, version detection, per-provider rate config |
| Audio attack module requires heavy deps | Medium | High | Make audio attacks optional, lazy-load, dedicated worker pool |
| DiffusionAttacker requires ML infra | Medium | High | Optional Python bridge via `os/exec`, not core Go dependency |
| MCP protocol evolves rapidly | Medium | High | Protocol version flexibility, transport-type-specific testing |
| Reasoning model APIs differ significantly | Medium | Medium | Optional `ReasoningProvider` interface, capability detection |
| China-based data residency (DeepSeek/Qwen) | Medium | High | Prominent documentation, `--data-residency-warning` flag |
| No tests for existing attack modules | Medium | High | Phase 0 test infrastructure, retroactive test coverage for existing modules |
| Agent-native API gaps block Phases 3 & 6 | High | High | Add attack-execution and attack-chaining API endpoints in Phase 0/1 |

## Future Considerations

- **Video modality attacks**: As video-input models mature (GPT-5 video understanding)
- **Federated learning attacks**: As federated LLM training becomes more common
- **Constitutional AI bypass**: As more models adopt CAI-style training
- **Quantum-resistant attacks**: Future consideration for cryptographic components
- **Real-time attack adaptation**: ML-driven real-time attack strategy optimization using bandit feedback

## References & Research

### Internal References
- Existing attack modules: `src/attacks/` (multimodal, orchestration, persistence, extraction, evasion)
- Provider infrastructure: `src/provider/` (anthropic, openai, core)
- OWASP mapping: `src/cmd/vulnerability_detection.go`, `cmd/compliance-report/`
- ML pipeline: `ml/data/attack_data_pipeline.py`, `ml/agents/multi_armed_bandit.py`
- Test harness: `llmrecon_harness.py`, `llmrecon_2025.py`
- Templates: `templates/`

### External References - Attack Techniques
1. Autonomous Reasoning Model Jailbreaking: [Nature Communications](https://www.nature.com/articles/s41467-026-69010-1)
2. Best-of-N Jailbreaking: [arXiv 2412.03556](https://arxiv.org/html/2412.03556v1)
3. Bad Likert Judge: [Palo Alto Unit 42](https://unit42.paloaltonetworks.com/multi-turn-technique-jailbreaks-llms/)
4. MetaBreak: [arXiv 2510.10271](https://arxiv.org/abs/2510.10271)
5. Crescendo: [arXiv 2404.01833](https://arxiv.org/abs/2404.01833)
6. Skeleton Key: [Microsoft Security Blog](https://www.microsoft.com/en-us/security/blog/2024/06/26/mitigating-skeleton-key-a-new-type-of-generative-ai-jailbreak-technique/)
7. Many-Shot Jailbreaking: [Anthropic Research](https://www.anthropic.com/research/many-shot-jailbreaking)
8. iMIST: [arXiv 2601.05466](https://arxiv.org/html/2601.05466v1)
9. PoisonedRAG: [USENIX Security 2025](https://www.usenix.org/system/files/usenixsecurity25-zou-poisonedrag.pdf)
10. AudioJailbreak: [arXiv 2505.14103](https://arxiv.org/abs/2505.14103)
11. MCP Attacks: [Unit 42](https://unit42.paloaltonetworks.com/model-context-protocol-attack-vectors/), [Practical DevSecOps](https://www.practical-devsecops.com/mcp-security-vulnerabilities/)
12. Browser Agent Attacks: [OpenAI Atlas](https://openai.com/index/hardening-atlas-against-prompt-injection/), [Anthropic](https://www.anthropic.com/research/prompt-injection-defenses)
13. AIShellJack: [IEEE S&P 2026](https://arxiv.org/html/2511.05797v1)
14. CoT Exploitation: [Trend Micro](https://www.trendmicro.com/en_us/research/25/c/exploiting-deepseek-r1.html)
15. Agent Attacks Q4 2025: [eSecurity Planet](https://www.esecurityplanet.com/artificial-intelligence/ai-agent-attacks-in-q4-2025-signal-new-risks-for-2026/)

### External References - Defenses
16. CaMel: [arXiv 2503.18813](https://arxiv.org/abs/2503.18813), [GitHub](https://github.com/google-research/camel-prompt-injection)
17. LLM Salting: [Sophos CAMLIS 2025](https://news.sophos.com/en-us/2025/10/21/getting-salty-with-llms-sophosai-unveils-new-defense-against-jailbreaking-at-camlis-2025/)

### External References - Frameworks & Threat Models
18. OWASP LLM Top 10 2025: [OWASP GenAI](https://genai.owasp.org/resource/owasp-top-10-for-llm-applications-2025/)
19. OWASP Agentic Top 10 2026: [OWASP GenAI](https://genai.owasp.org/resource/owasp-top-10-for-agentic-applications-for-2026/)
20. JailbreakBench: [ResearchGate](https://www.researchgate.net/publication/397201287_JailbreakBench_An_Open_Robustness_Benchmark_for_Jailbreaking_Large_Language_Models)
37. MITRE ATLAS (Oct 2025 update): [MITRE](https://atlas.mitre.org/) -- 15 tactics, 66 techniques, 46 sub-techniques, 14 agent-specific
38. MAESTRO 7-Layer Threat Model: [CSA](https://cloudsecurityalliance.org/research/artifacts/maestro-multi-agent-security-threat-model)
39. Fine-Tuning Alignment Tax: [arXiv 2310.03693](https://arxiv.org/abs/2310.03693)

### External References - Benchmarks & Evaluation
21. Investigator Agents: [Transluce](https://transluce.org/jailbreaking-frontier-models)
22. Multi-Turn Jailbreak Simplicity: [arXiv 2508.07646](https://arxiv.org/html/2508.07646v1)
34. AgentThreatBench (UK AI Safety Institute): [Gov.UK](https://www.gov.uk/government/publications/agentthreatbench)
35. DeepTeam (Confident AI): [Documentation](https://docs.confident-ai.com/docs/red-teaming)
36. Promptfoo OWASP Agentic: [Promptfoo Docs](https://www.promptfoo.dev/docs/red-team/owasp-agentic/)

### External References - Multi-Agent & Orchestrator Security
23. Protocol Exploits Survey: [arXiv 2506.23260](https://arxiv.org/abs/2506.23260), [ScienceDirect](https://www.sciencedirect.com/science/article/pii/S2405959525001997)
24. Agentic AI Security Survey: [arXiv 2510.23883](https://arxiv.org/html/2510.23883v1)
25. OpenClaw MAESTRO Threat Model: [Substack](https://kenhuangus.substack.com/p/openclaw-threat-model-maestro-framework)
26. OpenClaw Vulnerabilities (Kaspersky): [Kaspersky](https://www.kaspersky.com/blog/openclaw-vulnerabilities-exposed/55263/)
27. OpenClaw Enterprise Exploitation (Bitdefender): [Bitdefender](https://businessinsights.bitdefender.com/technical-advisory-openclaw-exploitation-enterprise-networks)
28. OpenClaw Prompt Injection Persistence: [Penligent](https://www.penligent.ai/hackinglabs/the-openclaw-prompt-injection-problem-persistence-tool-hijack-and-the-security-boundary-that-doesnt-exist/)
29. OpenClaw Critical Vulnerabilities (Cubic): [Cubic](https://www.cubic.dev/blog/we-found-and-fixed-critical-security-vulnerabilities-in-openclaw)
30. OpenClaw Exposure Analysis: [SiliconANGLE](https://siliconangle.com/2026/02/09/tens-thousands-openclaw-systems-exposed-due-misconfiguration-known-exploits/), [VentureBeat](https://venturebeat.com/security/openclaw-agentic-ai-security-risk-ciso-guide/), [Cisco](https://blogs.cisco.com/ai/personal-ai-agents-like-openclaw-are-a-security-nightmare)
31. Moltbot RCE (CVE-2026-25253): [depthfirst](https://depthfirst.com/post/1-click-rce-to-steal-your-moltbot-data-and-keys)
32. Agent Security Bench (ICLR 2025): [ICLR Proceedings](https://proceedings.iclr.cc/paper_files/paper/2025/file/5750f91d8fb9d5c02bd8ad2c3b44456b-Paper-Conference.pdf)
33. Multi-Agent Defense Pipeline: [arXiv 2509.14285](https://arxiv.org/html/2509.14285v4)

### External References - Implementation Patterns
40. ProjectDiscovery nuclei (registry + self-registration): [GitHub](https://github.com/projectdiscovery/nuclei)
41. Go errgroup bounded concurrency: [pkg.go.dev](https://pkg.go.dev/golang.org/x/sync/errgroup)
42. Go x/time/rate hierarchical limiting: [pkg.go.dev](https://pkg.go.dev/golang.org/x/time/rate)
43. CVE-2024-43405 nuclei signature bypass: [GitHub Advisory](https://github.com/projectdiscovery/nuclei/security/advisories)
44. OWASP Cryptographic Failures (A02:2021): [OWASP](https://owasp.org/Top10/A02_2021-Cryptographic_Failures/)

### Codebase-Specific Findings (from Deepening)

**Files requiring changes before Phase 1:**
- `src/provider/config/encryption.go:89-92` -- Replace bare SHA-256 KDF with argon2id
- `src/security/vault/vault.go:124` -- Replace hardcoded salt with random salt generation
- `src/security/vault/vault.go:456-471` -- Mask `Value` field in `ListCredentials()` output
- `src/security/vault/vault.go:612-622` -- Replace non-cryptographic `GenerateCredentialID`
- `src/provider/anthropic/provider.go` -- Enforce TLS for API base URLs
- `src/provider/middleware/logging.go` -- Enable `redactPII: true` by default

**Files requiring capacity updates:**
- `src/performance/memory_pool.go:119` -- `MaxMemoryUsage` 100MB → 2GB for many-shot mode
- `src/performance/redis_cluster_cache.go:352` -- `MaxValueSize` 1MB → 10MB for long-context
- `src/performance/concurrency_engine.go` -- `MaxWorkers` NumCPU*10 → NumCPU*25 for multi-agent
- `src/performance/distributed_rate_limiter.go` -- Single `DefaultLimit: 100` → per-provider config

**Duplicated types to consolidate (Phase 0.2):**
- `Provider` interface: `src/attacks/injection/interfaces.go`, `src/attacks/orchestration/multi_turn.go`, `src/attacks/multimodal/multimodal_attacks.go`
- `Logger` interface: Same 3 files (byte-for-byte identical)
- `Message` struct: Same 3 files
- `randInt()`: 6 packages
- `generateAttackID()`: 4 implementations
- `contains()`: 3 implementations

**API gaps (Agent-Native):**
- `src/api/router.go:249-253` -- Only scan-level endpoints; need individual attack execution endpoint
- `src/attacks/orchestration/multi_turn.go:StartSession()` -- Internal goroutine, not API-accessible
- No OpenAPI spec for attack techniques/strategies
- ML optimizer at `ml/agents/multi_armed_bandit.py` has no API surface
