# Changelog

All notable changes to LLMrecon are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project
follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

This changelog was started with v0.9.0; earlier history lives in `git log`.

## [Unreleased]

### Added

- **MCTS-Explore selection strategy for jbfuzz (#171).** New opt-in seed
  selector via metadata `selection=mcts_explore`, alongside the unchanged
  default `ucb1_restart`. Builds a Monte-Carlo search tree (each node a prompt,
  children its mutated variants) with a UCT tree policy — exploitation plus a
  depth-weighted exploration bonus — progressive widening, and reward
  backpropagation from each expanded leaf to the root, per GPTFuzzer (Yu et al.,
  USENIX Security 2024). Deterministic under a fixed `rng_seed`. The result
  records `selection` and `node_count`. Selection is now behind a `selector`
  abstraction; UCB1+restart behavior (and RNG call order) is preserved exactly.
- **Embedding-fitness for jbfuzz / persona_evolve (#170).** The evolutionary
  engines' opt-in `fitness=embedding` scorer is now implemented (was a stub
  returning "not implemented"). It scores goal-relevance as the cosine
  similarity between the target response and the operator objective, embedded
  via a local Ollama-style `/api/embeddings` endpoint (option b), and blends
  that with the refusal heuristic. Configure with metadata `embedding_endpoint`
  (default `http://localhost:11434/api/embeddings`) and `embedding_model`
  (default `nomic-embed-text`). Opting in without a reachable endpoint yields a
  clean `OutcomeSkipped` + `SkipPreconditionFailed`, never a crash; a transient
  mid-run embed failure degrades to the heuristic. Covered by unit tests against
  an in-process mock endpoint and an opt-in `RUN_INTEGRATION` smoke test against
  a real Ollama model.
- **`common.Purger` provider interface + `attack purge` command (#168).**
  Automated cleanup of memory-poisoning implants, succeeding v0.9.0's manual
  `CleanupHint` workflow. Providers that own a purgeable memory store implement
  `Purger.Purge(ctx, recordIDs)`; the memory-poisoning modules (`minja`,
  `memorygraft`, `injecmem`) now report `purger_available` in their result
  metadata, and `llmrecon attack purge --provider <name> --record-ids <ids>`
  (or `--result <emit-jsonl-file>`) rolls back the injection. Providers without
  the capability get a friendly error pointing back to the manual `CleanupHint`.
  Includes an in-memory reference `Purger` (`testutil.MockMemoryProvider`) and an
  inject→verify-present→purge→verify-absent smoke test.

## [0.10.0] - 2026-05-03

The v0.10.0 release is **the honesty release**. Every code path that
previously printed "success" while doing nothing now either does the
work for real, returns a typed error, or is removed along with the
docs that advertised it. No CLI command is more capable today than
v0.9.0; many are *less* capable, in the sense that paths that used to
fake-succeed now fail-fast with actionable messages.

The single acceptance criterion for v0.10.0: **no code path returns
"not implemented" while printing or returning success.**

### CLI surface — real, end-to-end

`attack list` and `attack run` enumerate and execute every registered
attack module. Before v0.10.0 the registry was empty at runtime — every
`init()`-registered module since v0.7.0 was build-only because no
binary linked the attack packages. Fixed via `src/attacks/all` barrel
import.

```
./llmrecon attack list                                 # 50+ modules
./llmrecon attack list --json                          # CI / scorecard
./llmrecon attack run --module=jbfuzz --provider=mock \
    --metadata=allow_experimental=true                 # ends with typed AttackResult
```

A `--provider=mock` is the only provider in v0.10.0; OpenAI and
Anthropic are wired but capability adapters drive
`SkipMissingCapability` until a future release wires the bridge into
the `attack run` CLI path. (#173)

### Provider capability adapters (#166)

OpenAI and Anthropic adapters now implement `core.ImageProvider` and
`core.ReasoningProvider`; the bridge package promotes them into
`common.ImageProvider` and `common.ReasoningProvider` via a 2×2
wrapper-type matrix. SIVA, VSH, and H-CoT modules' v0.9.0
type-assertion gates now find a capable provider against real OpenAI
and Anthropic targets instead of always emitting
`SkipMissingCapability`.

OpenAI:
- Vision via Chat Completions multimodal content parts
  (data: URLs for inline bytes, URL refs verbatim, `Detail` honored).
- Reasoning via the Responses API (`include=["reasoning.encrypted_content"]`,
  `reasoning.summary="detailed"`); empty-summary case (o3 omits >90%)
  surfaces as `len(trace.Steps)==0` so H-CoT short-circuits to
  `SkipReasoningTraceEmpty` after retry budget.

Anthropic:
- Vision via Messages API content blocks; `Detail` hint dropped (no
  equivalent in API).
- Reasoning via extended thinking (`thinking.type="enabled"`,
  `budget_tokens=10000`). `ReasoningTraceIsSigned() bool` returns
  `true` so the bridge surfaces `Signed=true`; H-CoT short-circuits
  to `SkipSignatureGated` rather than wasting attempts on
  thinking-text mutation that the API would silently discard.

### Capability gates for agentic/audio modules (#176)

Three new capability interfaces — `MCPProvider`, `BrowserProvider`,
`AudioProvider` — gate 13 modules (4 mcp + 2 tool_use + 3 browser + 4
audio) that previously text-simulated their advertised modality
against any provider. Default behavior: `OutcomeSkipped +
SkipMissingCapability` against a non-matching provider.

Operators who explicitly want the legacy text-simulation behavior
pass `Metadata["mode"]="text_simulation"`; modules then fall back
to plain `Query` AND tag the result with `Metadata["mode"]` and
`Metadata["true_modality"]` so downstream consumers (compliance
scorecards, bandit reward) can filter simulated runs out of
real-attack aggregations.

### `update apply` — atomic-replace + Tier 1 honesty (#174)

Tier 1: every "not implemented" stub in `src/update/` and
`src/cmd/update_apply.go` returns a non-nil error so the CLI exits
non-zero. The previous behavior was to print success while writing
nothing to disk — the worst kind of bug for a security tool.

Tier 2: `--experimental` flag opts into the new atomic-replace path.
Strategy:

1. Stage: extract bundle into a sibling tmp dir on the SAME
   filesystem as dest (so `os.Rename` is syscall-atomic).
2. Validate: caller-supplied `Validate` runs on staged dir.
3. Backup: `os.Rename(dest, dest+".bak."+ts)` — atomic.
4. Apply: `os.Rename(staged, dest)` — atomic.
5. Cleanup: optional, controlled by `--backup` flag.

Kill -9 at any point leaves the operator with EITHER the old install
OR the new one — never a half-extracted state.
`RecoverFromInterruptedApply` finds and restores from the most-recent
`.bak` sibling when a kill -9 lands between steps 3 and 4.

ZIP extraction guards: zip-slip rejection (per-entry path
re-verified via `filepath.Rel`), symlink skipping, per-entry
size cap (100 MiB), total cap (1 GiB), file mode AND'd with `0o755`.

### Bundle round-trip (#177)

`bundle verify` and `bundle import` are restored against the live
`src/bundle` API. The five `.disabled` cmd files (2,175 lines, written
against an early API draft that no longer exists) are atticked under
`attic/v0-7-0-bundle-disabled/` with a README. Aspirational commands
(`publish`, `sync`, `registry`) stay atticked — they depend on a
bundle-registry concept the codebase never finished and are out of
scope for v0.10.0.

`offline_bundle_cli.go` validation-level switch fixed: previously
parsed 7 levels but only handled 2; remaining 5 (basic, standard,
strict, manifest, compatibility) now route through the standard
`BundleValidator.Validate(level)` interface.

### Python ↔ Go JSONL bridge (#181)

`attack run --emit-jsonl=<path>` writes one JSONL row per
`AttackResult`. `python -m ml.data.ingest` reads JSONL and inserts
into the v0.9.0 `attacks.db` schema with field-by-field equality.
Credential redaction at the Python boundary scrubs `OPENAI_API_KEY`,
`ANTHROPIC_API_KEY`, generic bearer tokens, and any value matching
the v0.9.0 `_SENSITIVE_KEY_PATTERN`.

```
./llmrecon attack run --module=jbfuzz --provider=mock \
    --metadata=allow_experimental=true \
    --emit-jsonl=- | python3 -m ml.data.ingest
```

### Drift-detection CI (#169 + #179)

`scripts/verify-drift.sh` runs on every PR and asserts:

1. OWASP compliance codegen up-to-date (`go generate ./...` produces
   no diff against `src/compliance/owasp_agentic_generated.go`).
2. Go-version pins match `go.mod` across Dockerfiles, CI workflows,
   and any in-repo references.
3. Every YAML technique ID in `templates/owasp_agentic_2026.yaml`
   resolves to a Go constant in `src/compliance`.

The v0.9.0 Docker push break was caused by a Go-pin drift (1.24 in
Dockerfile vs 1.25 in `go.mod`); this CI guard prevents recurrence.

### Dead RBAC cleanup (#180)

The v0.2.0 RBAC + MFA + auth subsystem under `src/security/access/`
returned "not implemented" from every constructor. Four CLI commands
(`access_control`, `audit`, `auth`, `user`) sat as `.disabled` files.
No consumer ever wired the framework end-to-end. Removed: 121 files,
~24,000 lines. If auth/RBAC returns it'll be in v0.11.0+ as a fresh
design rather than a partial revival.

### `common.Provider` bridge shim (#167)

`bridge.WrapCore(core.Provider) → common.Provider` translates between
the heavier core LLM-API surface and the minimal surface attack
modules consume. Selects from a 2×2 wrapper-type matrix based on the
underlying provider's capabilities (`coreAdapter`, `imageAdapter`,
`reasoningAdapter`, `imageReasoningAdapter`). Four concrete types
instead of one struct with always-present methods because Go's type
assertion is structural — a single struct with `QueryWithImages`
would always satisfy `common.ImageProvider` regardless of underlying
support, defeating the gate that attack modules use to detect
capability presence.

### README pass (#175)

Six fictional CLI examples removed from the README. The first
Go-side quick-start example was `./llmrecon scan --provider openai
--model gpt-4 --owasp` — there is no `scan` command that takes those
flags. Replaced with the actually-working v0.10.0 surface (`attack
list`, `attack run`, `bundle verify`, `update apply --experimental`).

`src/cmd/readme_smoke_test.go` pins documented command paths against
the live Cobra tree so future README drift fails CI.

### Migration notes

- **No data migration needed.** v0.9.0 SQLite schema unchanged.
- **`attack run` provider value**: today only `mock` is wired through
  the CLI. Real providers (`openai`, `anthropic`) require the bridge
  package's `WrapCore` to be invoked from the CLI — that's planned
  for v0.10.1 along with `--api-key` / config-file wiring.
- **`update apply` operators**: the `--experimental` flag is
  required to opt into the atomic-replace path. Without it the apply
  path errors out with the v0.10.0 honesty message. One release
  cycle of opt-in before the default flips.
- **`bundle import` and `bundle verify`** require an EXTRACTED bundle
  directory, not a `.tar.gz` / `.zip` archive. Extract with `tar xf`
  or `unzip` first.

### Deferred to v0.11.0

- **#168 Purger interface** — automated cleanup of memory-poisoning
  implants. Depends on a new CLI subcommand surface; no operator
  pulling on it.
- **#170 Embedding fitness** — opt-in fitness function for
  `jbfuzz`/`persona_evolve`; current heuristic remains the default.
- **#171 MCTS-Explore** — opt-in selection algorithm for `jbfuzz`;
  default UCB1+restart unchanged. Pure engine R&D.
- **#174 Tier 3** — binary self-replace + signature verification for
  `update apply`. Out of scope; needs separate planning cycle.
- **`bundle create --sign`** (cosign integration), **`bundle create
  --source=github`** / **`--source=gitlab`** — Tier B/C of #177.
- **OpenAI/Anthropic provider wiring through the CLI** — adapters
  exist (#166), but `attack run` only accepts `--provider=mock`
  today. v0.10.1 will land the wiring + `--api-key`/config plumbing.

## [0.9.0] - 2026-05-02

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
