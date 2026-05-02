# Changelog

All notable changes to LLMrecon are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project
follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

This changelog was started with v0.9.0; earlier history lives in `git log`.

## [Unreleased]

## [0.9.0] — 2026-05-02

The v0.9.0 release adds **6 Go attack modules and 5 attack templates**
covering Q4 2025 – Q2 2026 LLM and agentic-AI attack research, plus the
infrastructure to make outcomes machine-comparable across the
ML/bandit/compliance stack: a typed `AttackOutcome` taxonomy, optional
provider capabilities, an idempotent ML schema migration, and an
OWASP-Agentic codegen.

### New attack modules

| Module | Path | Source | OWASP Agentic | Safety gate |
|---|---|---|---|---|
| `minja` / `memorygraft` / `injecmem` | `src/attacks/memory/poisoning.go` | arXiv 2503.03704 / 2512.16962 / OpenReview QVX6hcJ2um | ASI06 (+ ASI10 for memorygraft) | `i_understand_risks=true` |
| `h_cot` | `src/attacks/reasoning/h_cot.go` | arXiv 2502.12893, 2510.26418 | ASI01 | `i_understand_risks=true` + `common.ReasoningProvider` |
| `siva` | `src/attacks/multimodal/siva.go` | arXiv 2602.08136 | ASI01 | `common.ImageProvider` |
| `vsh` | `src/attacks/multimodal/vsh.go` | ScienceDirect S0031320325010520 | ASI01 | `common.ImageProvider` |
| `jbfuzz` | `src/attacks/adaptive/jbfuzz.go` | arXiv 2503.08990 | ASI01 | `allow_experimental=true` |
| `persona_evolve` | `src/attacks/adaptive/persona_evolve.go` | arXiv 2507.22171 | ASI01 + ASI09 | `allow_experimental=true` |

### New attack templates

| Template | Source | Substitution |
|---|---|---|
| `templates/cot_hijack_prepend.json` | arXiv 2510.26418 | static |
| `templates/echoleak_chain.json` | arXiv 2509.10540 / CVE-2025-32711 | `{{HARMFUL_INSTRUCTION}}` / `{{EXFIL_URL}}` |
| `templates/reverse_captcha.json` | arXiv 2603.00164 | `{{HARMFUL_INSTRUCTION}}` |
| `templates/system_prompt_surface.json` | arXiv 2603.25056 | static |
| `templates/genetic_persona_seeds.json` | corpus for `persona_evolve` | n/a (consumed by GA engine) |

`template_loader.py` rejects templates with unfilled `{{PLACEHOLDER}}`
fields at load time so a typo can never silently send the literal
template to a target.

### `AttackOutcome` 3-state taxonomy

`src/attacks/common/types.go` introduces a typed outcome enum that every
v0.9.0 module returns. Earlier modules keep their `Success bool` API;
the new field is additive.

```go
type AttackOutcome string
const (
    OutcomeSuccess AttackOutcome = "success"   // attack landed
    OutcomeRefused AttackOutcome = "refused"   // ran fully, target resisted
    OutcomeSkipped AttackOutcome = "skipped"   // didn't run (capability/gate/budget/error)
)
```

Skipped runs additionally set a typed `SkipReason`:

| `SkipReason` | When |
|---|---|
| `SkipMissingCapability` | Provider doesn't implement the optional interface the module needs. |
| `SkipGateBlocked` | Safety-gate metadata flag is missing (`i_understand_risks` / `allow_experimental`). |
| `SkipBudgetExceeded` | Engine exhausted query/wall-clock/generation budget without crossing the success threshold. |
| `SkipProviderError` | Provider call failed in a way that isn't a logical refusal (network, 5xx, parse). |
| `SkipPreconditionFailed` | Operator config missing required field (corpus path, payload, etc). |
| `SkipModelRefusedImage` | Multimodal module's post-check matched an image-blind indicator. |
| `SkipReasoningTraceEmpty` | Reasoning provider returned no steps after the retry budget. |
| `SkipSignatureGated` | Anthropic-style cryptographically-signed reasoning trace; mutation would be silently discarded. |
| `SkipNoMutationTarget` | H-CoT couldn't locate a safety-recognition step to hijack. |
| `SkipMemoryNotRetained` | `MemoryProbe` reported the target is stateless. |

The bandit-relevant invariant: **`OutcomeSkipped` rows are excluded from
reward aggregation**. Skipped reflects engine/capability state, not
target behavior; including them would bias the bandit toward (or
against) targets that simply lack capabilities. The Python pipeline
exposes this filter via `AttackDataPipeline.get_bandit_rewards()`, the
canonical single point of truth for the rule.

### New optional provider capabilities

In `src/attacks/common/capabilities.go`, mirroring the
`common.Provider` minimal-interface pattern. Modules type-assert at
`Execute()` entry and emit `OutcomeSkipped + SkipMissingCapability`
when the assertion fails.

| Interface | Method | Used by |
|---|---|---|
| `ImageProvider` | `QueryWithImages(ctx, prompt, images, opts)` | `siva`, `vsh` |
| `SessionProvider` | `SessionID()`, `NewSession(ctx)` | `memorygraft` (cross-session verification) |
| `MemoryProbe` | `ProbeMemory(ctx) (retains bool, err error)` | All memory-poisoning modes |
| `ReasoningProvider` | `QueryWithReasoning(ctx, msgs, opts) (resp, ReasoningTrace, err)` | `h_cot` |
| `Cleaner` (on modules, not providers) | `Cleanup(ctx, ids)` | future v0.10.0 cleanup helper |

`ReasoningTrace` carries `Steps []string` and `Signed bool`. The Signed
flag is the v0.9.0 Anthropic limitation: modifying the text of a signed
thinking block on round-trip is silently discarded by the API, so
`h_cot` short-circuits to `SkipSignatureGated` rather than wasting a
re-injection round-trip.

`ImagePayload` is constructor-only (`NewImagePayloadBytes` /
`NewImagePayloadURL`) — direct struct literals fail at compile time
because the fields are unexported. Constructors validate MIME type,
size cap (`MaxImagePayloadBytes = 5 MiB`), and detail enum.

### Retry helper with typed transient/permanent errors

`src/provider/core/retry.go` (`RetryableQuery`) provides a generic retry
loop wrapping any `func(ctx) (T, error)`. The error taxonomy is
deliberately small:

- `TransientError` — rate limit, 5xx, network — retry with exponential
  backoff, jitter, optional `Retry-After` honor, ctx-aware sleep.
- `PermanentError` — auth, content-policy, schema mismatch — surface
  immediately without retry.
- Any other error type — surface immediately. Buggy provider returns
  must not absorb the retry budget.

ctx cancellation always wins: cancelling mid-loop returns `ctx.Err()`,
never the previous transient.

### OWASP Agentic 2026 codegen

`cmd/owasp-gen` derives `GeneratedTechniqueToAgenticCategories` from
`templates/owasp_agentic_2026.yaml`. v0.9.0 ships the generator
side-by-side with the existing hand-written
`TechniqueToAgenticCategories`; v0.10.0 will switch the runtime lookup
to the generated map and add a `go generate ./... && git diff
--exit-code` drift check in CI.

The YAML's `attack_techniques` list under each `ASIxx` category is the
canonical source. The 8 v0.9.0 techniques (h_cot, siva, vsh, jbfuzz,
persona_evolve, minja, memorygraft, injecmem) are mapped, with
`technique_index` providing the reverse lookup.

### ML data pipeline schema migration

`ml/data/attack_data_pipeline.py` adds two columns to the `attacks`
table:

- `outcome TEXT` — typed promotion of the legacy `status` column.
- `parent_run_id TEXT` — links generations of evolutionary engines.

Plus partial indexes (`idx_attacks_outcome`,
`idx_attacks_parent_run_id`) and a one-time backfill: `outcome =
'success' if status='success' else 'refused'`.

The migration:

- Is **idempotent**: subsequent invocations no-op on the `ALTER`
  branches via `PRAGMA table_info` introspection.
- Backs up the database via SQLite's **Online Backup API**
  (`Connection.backup`), not `shutil.copy2`. WAL-mode databases keep
  recently-committed transactions in `-wal`/`-shm` sidecar files; a
  filesystem copy can produce an inconsistent snapshot.
- Runs at every pipeline init via `_init_database`, but only creates
  a `.bak.{UTC-timestamp}` file on the first run that detects either
  column is missing — avoids backup proliferation.

### Credential redaction in the ML pipeline

`_redact_sensitive_keys()` recursively scrubs dict keys matching
`(?i)(key|token|secret|password|auth)` in `technique_params` and
`features` before `INSERT`. Matching values become `"[REDACTED]"`.
Operator-supplied Metadata frequently leaks API keys / OAuth tokens /
session bearers when captured verbatim into analytics; this closes
that path at the storage layer.

### Safety-gate flag matrix

| Flag | Modules requiring it |
|---|---|
| `i_understand_risks=true` | `minja`, `memorygraft`, `injecmem`, `h_cot`, `rce_chain` |
| `allow_experimental=true` | `jbfuzz`, `persona_evolve`, `autonomous_jailbreak`, deceptive-alignment, agent-collusion |
| `allow_autonomous=true` | `autonomous_jailbreak` (per v0.8.0) |

Modules emit `OutcomeSkipped + SkipGateBlocked` when the corresponding
flag is missing, never silent `Success=false`.

### Anthropic signature limitation (H-CoT)

Per arXiv 2510.26418, modifying the text of an Anthropic
extended-thinking block on round-trip is silently discarded by the
API; only the original text is replayed back to the model. The H-CoT
module detects this via `ReasoningTrace.Signed=true` and short-circuits
to `SkipSignatureGated` rather than wasting a re-injection round-trip
that would have no effect.

For Anthropic targets, the `cot_hijack_prepend.json` template (Phase 2)
is the in-tree alternative — a static template-prepend that doesn't
require live trace mutation.

### Integration test cost notes

The `RUN_INTEGRATION` smoke tests added in this release run against a
local `MockLLMServer` and never touch a real provider. Operators
running them against real providers (uncapped budgets, real keys)
should expect ~$3–$8 per `RUN_INTEGRATION` run on production-tier
models — most of that cost comes from the GA engines (`jbfuzz`,
`persona_evolve`) which can issue thousands of queries per run. The
default budget caps in `common.HardMaxQueries` (5000) and
`common.HardMaxWallClockSeconds` (1800) bound the worst case.

### Other

- 7 v0.9.0 attack-technique entries added to
  `templates/owasp_agentic_2026.yaml` and the hand-written
  `TechniqueToAgenticCategories` map.
- `src/provider/core/retry.go` — top-of-loop ctx check now returns
  `ctx.Err()` unconditionally (operator cancellation trumps a previous
  transient).
- Test data: `src/attacks/multimodal/testdata/{siva_source.png,
  vsh_scenic.png}` (procedurally generated, ~2 KB total).

### Migration notes for v0.8.0 → v0.9.0

- **No breaking API changes.** Existing modules continue to compile
  and produce `Success bool` results unchanged.
- **DB migration is automatic** on the first `AttackDataPipeline`
  init that opens a pre-v0.9.0 database. A backup is created.
- **Safety-gate flags are required** to use v0.9.0 modules with
  network effects; absence emits a clean Skipped result rather than
  a runtime error.
- **The ML bandit reward filter excludes Skipped rows.** If your
  consumer aggregates raw `success_rate`, switch to
  `get_bandit_rewards()` for the canonical filter.

[Unreleased]: https://github.com/perplext/LLMrecon/compare/v0.9.0...HEAD
[0.9.0]: https://github.com/perplext/LLMrecon/releases/tag/v0.9.0
