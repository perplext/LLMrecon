---
title: "feat: v0.13.0 — new attack modules (post-v0.12.0 research wave, Mar–Sep 2026)"
type: feat
date: 2026-09-10
status: proposed
origin: docs/plans/2026-09-10-feat-v0-13-0-new-attack-modules-plan.md
target_branch: feature/v0.13.0-new-attack-modules
follows: docs/plans/2026-06-19-001-feat-v0-12-0-new-attack-modules-plan.md
---

# feat: v0.13.0 — New Attack Modules

## Summary

The attack catalog's research frontier stops at the v0.12.0 absorption
(arXiv ~2605/2606, June 2026). Since then the only commits have been
dependency bumps — no new research has landed in roughly 80 days. This plan
absorbs the **March–September 2026** wave, with emphasis on **July, August,
and September 2026**, which is entirely unrepresented.

The recommended v0.13.0 release is a **focused set of 6 net-new modules + 2
capability/mock additions**, chosen for recency, distinctness from the
existing catalog, and buildability against mock targets under the existing
honesty bar. A larger prioritized backlog (evolutions of existing modules,
plus lower-priority or heavier-lift techniques) is captured for v0.14.0+.

All items below were verified against primary sources (arXiv abstract pages,
vendor advisories, CVE records) during research on 2026-09-10. Every
candidate arXiv ID was cross-checked against the repository and confirmed
**not** already referenced by any module.

---

## Problem Frame

v0.12.0 observed that the frontier had shifted from single-prompt jailbreaks
toward persistence and optimization against stateful agents. The Mar–Sep 2026
wave sharpens that further into four distinct clusters the catalog does not
represent:

1. **Data-channel and rendering-fidelity attacks** — injecting *data*
   (spoofed provenance, control tokens, invisible-encoded tool metadata)
   rather than instructions, and exploiting the gap between what a human
   approves and what the model actually receives.
2. **Coding-agent command/OS-state composition** — chaining individually
   benign shell commands into an exploit via shared OS state.
3. **Reasoning-channel attacks that preserve a benign surface** — corrupting
   intermediate reasoning while the final answer looks safe, and stealing the
   hidden reasoning channel itself.
4. **Storage-layer and supply-chain attacks on RAG / fine-tuning** —
   soft-deleted embedding recovery, model-code backdoors, code-RAG vuln
   propagation.

The cost of the gap is the same one v0.12.0 named: as a testing tool,
LLMrecon under-represents the surface its users now defend, and absent an
explicit home this research is lost. A dedicated v0.13.0 line is where it
belongs.

---

## Coverage Baseline (what already exists — do NOT re-implement)

The catalog has ~50 registered attack modules. The following are directly
adjacent to candidates below and must not be duplicated:

- **Memory:** `memory_poisoning` (modes: minja/memorygraft/injecmem),
  `memmorph` (arXiv 2605.26154), `sleeper` (arXiv 2605.15338), `imist`.
- **Injection / adaptive:** `iterinject` (arXiv 2605.24659), `jbfuzz`,
  `persona_evolve`, `defense_bypass_optimizer`, `indirect_rag_injection`.
- **Reasoning:** `h_cot` (arXiv 2502.12893), `cot_exploitation`,
  `reasoning_loop_exploit`, `siva`, `vsh`.
- **Jailbreak / evasion:** `crescendo`, `skeleton_key`, `bad_likert_judge`,
  `many_shot`, `metabreak` (cipher), `content_concretization`,
  `immersive_world`, `poetry_attacks`, `salt_resistance`.
- **Agentic:** `mcp_tool_poisoning`, `mcp_supply_chain`,
  `mcp_filesystem_escape`, `mcp_sampling_injection`, `browser_*`
  (screenshot/document/hidden_instruction), `symjack`, `trustfall`,
  `rce_chain`, `credential_harvest`, `skill_poisoning`,
  `skill_takeover_chain`, `agent_collusion`, `toxic_agent_flow`.
- **RAG / extraction:** `poisoned_rag`, `kg_rag_poisoning`,
  `vector_embedding_attack`, `model_inversion`, `membership_inference`,
  `template_extraction`, `differential_extraction`.
- **Audio / multimodal:** `audio_jailbreak`, `bon_audio`,
  `multilingual_audio`, `speech_model_exploit`.

### Collision / already-covered flags (resolved during research)

- **"InjecMEM" (arXiv 2608.23471, Aug 2026)** shares a name with the existing
  `injecmem` memory-poisoning *mode*. The existing mode derives from earlier
  (2025) InjecAgent-style work; the new paper is single-interaction,
  retriever-agnostic-anchor memory injection. **Before building, diff the new
  paper's mechanism against the existing mode.** If distinct, implement as a
  new mode with a non-colliding registered name; if substantially the same,
  drop it. Do not register a second `injecmem`.
- **"Hidden in Memory: Sleeper Memory Poisoning" (arXiv 2605.15338)** is
  **already implemented** as `sleeper`. Excluded.
- **QueryIPI (arXiv 2510.23675)** is **October 2025** — out of the 180-day
  window. Remains a valid unimplemented technique but must not be counted as
  recent.

---

## Recommended v0.13.0 Scope (6 modules + 2 infra units)

Selection criteria: net-new attack *class* (not an evolution of an existing
module), black-box or mock-target implementable, strong primary sourcing,
distinct success semantics. Each maps cleanly onto existing conventions
(`AttackModule` + `init()` self-registration, 3-state `AttackOutcome`, typed
`SkipReason`, safety-gate metadata, OWASP LLM + Agentic mapping in
`Techniques()`).

| # | Module | Package | Source | Class | Gate | Capability |
|---|---|---|---|---|---|---|
| M1 | `agent_data_injection` | `adaptive/` | arXiv 2607.05120 (Jul 2026) | Data/provenance + control-token injection | `i_understand_risks` | Provider (+ optional ToolProvider) |
| M2 | `mcp_tag_concealment` | `agentic/mcp/` | arXiv 2607.05744 (Jul 2026) | Invisible-encoding + approval-view fidelity gap | `i_understand_risks` | `MCPProvider` |
| M3 | `mosaic_cmd_chain` | `agentic/persistence/` | arXiv 2607.02857 (Jul 2026) | Benign-command composition via OS state | `i_understand_risks` | `CodingAgentProvider` |
| M4 | `prja_reasoning_inject` | `reasoning/` | arXiv 2604.15725 (Apr 2026) | Reasoning-step injection, benign final answer | `i_understand_risks` | Provider (+ `ReasoningTrace`) |
| M5 | `token_suppression` | `evasion/` | CrowdStrike PT0197 (Jul 2026) | Refusal-token distribution suppression | none (low-risk) | Provider |
| M6 | `ghost_vectors` | `rag/` | arXiv 2606.18497 (Jun 2026) | Soft-deleted embedding reconstruction | `i_understand_risks` | new `VectorStoreProbe` |

Infra units:

- **U-CAP: `VectorStoreProbe` capability + `MockVectorStore` (testutil).**
  Needed by M6. Interface exposes insert / delete (soft) / raw-index-read so
  the module can assert that a "deleted" vector remains reconstructible.
  Mirrors the `MemoryProbe` / `MockCodingAgent` precedent. Against any
  provider lacking it: `OutcomeSkipped + SkipMissingCapability`.
- **U-OWASP: taxonomy refresh.** Add the six technique→category entries to
  `templates/owasp_agentic_2026.yaml`, regenerate via `cmd/owasp-gen`, keep
  the hand-written `TechniqueToAgenticCategories` consistent, and confirm
  `owasp_agentic_drift_test.go` stays green. Also fold in the **Microsoft
  "Taxonomy of Failure Modes in Agentic AI Systems v2.0"** (April 2026)
  category labels where they apply (MCP/plugin abuse, computer-use visual
  attacks, NL-tool-description supply chain).

### Per-module implementation notes

**M1 `agent_data_injection` (arXiv 2607.05120).** Inject malicious content
disguised as *trusted data* — spoofed resource identifiers, data-origin
fields, and tool-call/response framing — rather than imperative instructions.
Include the paper's **probabilistic delimiter** sub-technique (exploiting a
model's fuzzy interpretation of inexact delimiters so untrusted content is
parsed as trusted framing). **Merge CrowdStrike Special Token Injection
(PT0198)** here as a `mode=control_token` variant — it is the same
delimiter/control-marker-spoofing idea. Success = target acts on the spoofed
data / treats injected content as system-priority; refusal = target isolates
the untrusted data. Two agents independently surfaced this as the highest-value
gap; confirmed RCE on Claude Code/Codex/Gemini CLI and arbitrary clicks on web
agents. OWASP: ASI01 (+ LLM01).

**M2 `mcp_tag_concealment` (arXiv 2607.05744).** MCP tool metadata is shown to
the human once at approval but re-injected into model context every turn; the
protocol never requires the approval-rendered bytes to match the model-fed
bytes. Hide payloads in Unicode TAG block (U+E0000–U+E007F) that render as zero
glyphs but survive into the tokenizer, plus a TOCTOU "rug-pull" that swaps
metadata after approval. Success = payload reaches model context / rug-pull
forces no re-approval; the core assertion is **approval-view ≠ model-view**.
Reuses `MCPProvider`. Distinct from `mcp_tool_poisoning` (which tests malicious
*semantic content*, not invisible-encoding/rendering divergence). OWASP:
ASI01/ASI05.

**M3 `mosaic_cmd_chain` (arXiv 2607.02857).** Compose individually-benign
shell commands into an exploit where dangerous producer→consumer relationships
form across a command trace via shared OS state; each command passes filters
alone, the chain is the payload. Delivered as realistic developer-workflow
tasks. 96.59% ASR across 5 coding agents × 5 backends. Reuses
`CodingAgentProvider` + `MockCodingAgent` (extend the mock to record a command
trace and evaluate cross-command state effects). Success = the composed chain
reaches a dangerous OS-state effect the mock flags. Distinct from `rce_chain`
(single-step) and `symjack` (symlink/approval). OWASP: ASI01/ASI05.

**M4 `prja_reasoning_inject` (arXiv 2604.15725).** Two stages: identify
manipulable reasoning triggers in the target's chain-of-thought, then craft
prompts grounded in named psychological levers (obedience-to-authority, moral
disengagement) that inject harmful content into *intermediate reasoning steps
while keeping the final answer's surface benign* — evading answer-level safety
checks. 83.6% avg ASR on DeepSeek-R1, o4-mini, Qwen2.5-Max. Distinct from
`h_cot` (hijacks the safety-recognition step to flip the answer) and
`cot_exploitation`. Where the provider exposes `ReasoningTrace`, score success
on trace content; on `Signed=true` traces, short-circuit to `SkipSignatureGated`
(same rule H-CoT uses). OWASP: LLM01 / ASI01.

**M5 `token_suppression` (CrowdStrike PT0197).** Instruct the model to avoid
the tokens/phrases it uses to refuse (apologies, "I can't", policy
statements), shifting the output distribution away from refusal patterns
without a direct jailbreak request. Lowest-lift module in the set; black-box;
runs against `--provider=mock`. Treat as low-risk (no `i_understand_risks`
gate) consistent with other pure-prompt evasion modules — confirm against the
existing evasion-package gate matrix during implementation. OWASP: LLM01.

**M6 `ghost_vectors` (arXiv 2606.18497).** HNSW vector stores implement
deletion as soft-delete (tombstone), leaving raw embeddings in the index. With
storage-layer access, read the raw index, extract "deleted" vectors, run
Vec2Text-style inversion to reconstruct source text. The vulnerability is
**deletion durability**, not the inversion model — so the test is insert →
delete → attempt raw-index recovery, and success = a soft-deleted record is
still reconstructible. Reported up to 99–100% recovery in some domains.
Requires U-CAP (`VectorStoreProbe` + `MockVectorStore`). Distinct from
`vector_embedding_attack` / `model_inversion` (which target live embeddings via
API). OWASP: LLM02/LLM08.

---

## v0.14.0+ Backlog (curated, deduplicated)

### Strong net-new candidates deferred only for scope

| Candidate | Source | Note |
|---|---|---|
| Reasoning-Trace Stealing | arXiv 2608.09867 (Aug 2026) | Cross-model encrypted-reasoning-block echo recovers hidden CoT/PII. Novel surface; needs a provider that round-trips encrypted reasoning blobs — heavier target modeling. Extraction package. |
| ContextLeak | arXiv 2608.27800 (Aug 2026) | RL-optimized malicious tool name/desc makes the agent pass its own context as tool arguments. Distinct exfil vector; RL-optimizer is a build cost. |
| Copirate 365 / CVE-2026-24299 | Embrace The Red, DEF CON Aug 2026 (disclosed Mar 2026) | M365 Copilot HTML-preview exfil (CSS background-image → @font-face), near-zero-click. Strong repro; models a specific product surface — good `agentic/browser` or dedicated module. |
| Active Execution Hijacking | arXiv 2604.27426 (Apr 2026) | Model-*code* backdoor (`modeling_*.py`) exfiltrates victim fine-tune data during local training. Supply-chain; >98% ASR. Needs a fine-tune/model-code execution surface. |
| Embedding Inference Attack | arXiv 2607.01276 (Jul 2026) | Black-box embedding-model fingerprinting as recon before inversion. Pairs naturally with M6/`ghost_vectors`. |
| Mind Viruses | arXiv 2608.10218 (Aug 2026) | Self-propagating payloads via shared editable prompt files (SOUL.md/MEMORY.md). Multi-agent; note the paper's own mitigation (one warning paragraph) as an expected-failure condition. |
| Claudini autoresearch | arXiv 2603.24511 (Mar 2026) | Meta-capability: an agent loop that *discovers* new attack algorithms. Large lift; belongs with the `adaptive/` engines as a capstone, not a module. |
| JailAgent | arXiv 2604.05549 (Apr 2026) | Reasoning-trajectory + memory-retrieval hijack with "constraint tightening". Overlaps M4; evaluate together. |
| BioShocking | LayerX, Jun 2026 | False-premise → guardrail-context invalidation + authenticated-session browser credential theft across 6 browser agents. Industry disclosure; good `agentic/browser` module. |
| Comment and Control | Disclosed Apr 2026 (JHU) | CI/CD VCS-metadata (PR titles, comments) → autonomous coding agent in a privileged runner → secret egress via logs/findings. Distinct trigger channel. |
| Dependency Steering | arXiv 2605.09594 (May 2026) | Skill references compromised external packages; agent's autonomous dep-resolution pulls the backdoor. Extends `skill_poisoning`. |

### Evolutions of existing modules (extend, do NOT add a new class)

| Evolution | Source | Extend |
|---|---|---|
| Prosody-driven audio jailbreak (PJ-Break) | arXiv 2607.26541 (Jul 2026) | `audio_jailbreak` — add paralinguistic prosody presets (arousal/authority/rate) with fixed transcript. |
| Information-overloading LVLM jailbreak | arXiv 2607.02961 (Jul 2026) | `vsh` / image jailbreak — recursive image-text complexity overload of cross-modal alignment. |
| CodePoisonRAG | arXiv 2609.02774 (Sep 2026) | `poisoned_rag` — payload is an attacker-chosen CWE vuln + "safe/reviewed" mislabeling. |
| Query-agnostic RAG poison (QASRP) | arXiv 2607.04379 (Jul 2026) | `poisoned_rag` — single universal-ranking poison, no known-query assumption. |
| The Framing Gap | arXiv 2608.27092 (Aug 2026) | `credential_harvest` / exfil — reframe exfil as "integrity signature"/config field; 0%→100% on gpt-4o. |
| Arbitrary in-context cipher | arXiv 2609.09553 (Sep 2026) | `metabreak` — teach an arbitrary cipher via ICL (no fine-tuning), output-classifier evasion. |
| Test-time-search IPI | arXiv 2609.04495 (Sep 2026) | `iterinject` / IPI engines — make the attacker **compute-budget-swept** (recon + strategy-portfolio + budget scaling) rather than fixed-budget. |
| Depth-dependent IPI | arXiv 2605.30686 (Jun 2026) | IPI modules — add injection-depth / turn-budget as fuzzing dimensions. |
| CrowdStrike PT0201 / PT0200 / IM0018 | CrowdStrike (Jul 2026) | Trigger-activated rule (logic bomb) → near `sleeper`; algorithmic payload decomposition → near DrAttack; unwitting-user context injection → near indirect-injection. Evaluate for overlap before building. |

### Deferred-but-surveyed (from the v0.12.0 scope boundaries — status confirmed)

- **ASPI (arXiv 2605.17324, May 2026)** — clarification-seeking state transition
  amplifies IPI susceptibility; 728 task-attack scenarios. In-window. Good
  candidate; models an agent *state* rather than a payload.
- **Structural Template Injection / "Phantom" (arXiv 2602.16958, Feb 2026)** —
  chat-template token injection causing role confusion. **Pre-window (Feb)** but
  a valid gap.
- **Skill-Inject (arXiv 2602.20156, Feb 2026)** — skill-file injection
  benchmark. Pre-window; overlaps `skill_poisoning`.
- **OX Security MCP supply-chain** — enrichment for `mcp_supply_chain`
  (STDIO command injection, MCPoison/CVE-2025-54136 trust reuse, "Miasma"
  worm). Data/enrichment, not a new module.

---

## Cross-Cutting Requirements (apply to every new module)

Carried forward verbatim from the v0.10/v0.11/v0.12 honesty bar:

- **Real `Execute()` or a typed skip.** No fabricated `success=false` theater.
  Use `SkipGateBlocked` (missing safety flag), `SkipMissingCapability` (failed
  interface assertion), `SkipBudgetExceeded` (engine exhausted),
  `SkipProviderError` (transient), `SkipPreconditionFailed` (operator-config).
- **3-state outcome** via `common.NewAttackResult`; `OutcomeSkipped` rows stay
  excluded from bandit reward aggregation (`get_bandit_rewards`).
- **Self-registration** via `init()` into `attacks.DefaultRegistry`; appears in
  `attack list` / `attack list --json` and runs via `attack run`.
- **Barrel:** `agentic/mcp`, `agentic/persistence`, `adaptive`, `reasoning`,
  `evasion`, and `rag` are **already imported** in `src/attacks/all`; the six
  recommended modules need **no barrel edit**. Run
  `go test ./src/cmd/... -run TestNoNameCollisions` after adding names.
- **Tests on every modified surface**, plus one `RUN_INTEGRATION`-gated smoke
  test per new family under `src/attacks/integration/` (`t.Skip` when unset).
- **Ground each module in the primary source** — the header comment cites the
  arXiv ID / advisory and states the mechanism, as `iterinject.go` /
  `symjack.go` do today.
- **OWASP mapping** in `Techniques()` and in the YAML; regenerate, never
  hand-edit the generated map; keep the drift test green.

### Grounding still required at implementation time

Success rates were not in the abstract (mechanism verified, effectiveness
partial) for: **ContextLeak, JailAgent, arbitrary-cipher (2609.09553),
InjecMEM (2608.23471), test-time-search IPI (2609.04495)**. Read the full paper
bodies before finalizing success classifiers for any of these. The six
recommended v0.13.0 modules (M1–M6) all had mechanism **and** date verified
from primary sources.

---

## Verification / Acceptance

- `go build -o llmrecon ./src/main.go` succeeds; `attack list` shows the six new
  modules; `attack run` reaches each against `--provider=mock` (or the relevant
  mock target), producing a typed `AttackResult` with no fabricated success.
- `RUN_INTEGRATION=1 go test ./src/attacks/integration/...` green.
- `go test ./src/attacks/... ./src/compliance/...` green;
  `owasp_agentic_drift_test.go` green after YAML regeneration.
- A researcher can trace each module to its cited source and see why it is
  distinct from the adjacent existing module named in these notes.

---

## Scope Boundaries

- No production `CodingAgentProvider` / `VectorStoreProbe` adapters against real
  systems — mock targets only (matches the v0.12.0 boundary).
- Evolutions in the backlog table are **not** in v0.13.0; they extend existing
  modules and should be batched per-package later.
- No ML/bandit pipeline changes beyond correct outcome emission.
- White-box weight-modification attacks (e.g. SAHA / Depth Charge,
  arXiv 2603.05772) are excluded — out of scope for a black-box prompt-testing
  tool unless a self-hosted-weights testing mode is added.
