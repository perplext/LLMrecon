---
title: "feat: v0.12.0 — new attack modules (MemMorph, Sleeper, IterInject, SymJack, TrustFall)"
type: feat
date: 2026-06-19
status: completed
origin: docs/brainstorms/2026-06-19-v0-12-0-new-attack-modules-requirements.md
target_branch: feature/v0.12.0-new-attack-modules
follows: docs/plans/2026-05-04-feat-v0-11-0-stabilization-plan.md
---

# feat: v0.12.0 — New Attack Modules

## Summary

Add five May–June 2026 attack modules — MemMorph + Hidden Sleeper (memory), IterInject (feedback-optimized indirect injection), and SymJack + TrustFall (coding-agent RCE) — plus a `CodingAgentProvider` capability and a testutil mock coding-agent target. All five land in packages already wired into the registry barrel, follow the existing AttackModule + 3-state-outcome + capability-gate conventions, and meet the v0.10/v0.11 honesty bar: real `Execute()`, tests on modified surface, typed skips, no fabricated success.

---

## Problem Frame

The catalog stops at the v0.9.0 research absorption (~arXiv 2602/2603). The last 30 days produced a distinct wave of agent-targeting techniques (arXiv 2605.xxxxx + named Adversa disclosures) that the catalog does not represent, all clustering on agent statefulness and trust UX rather than model gullibility. v0.11.0 stabilization deliberately cut new research to protect its tag, so this work belongs on a separate v0.12.0 branch built on the stabilized baseline. See origin: `docs/brainstorms/2026-06-19-v0-12-0-new-attack-modules-requirements.md`.

---

## Resolved Planning Questions

The origin doc deferred five questions. All five are resolved from codebase patterns:

1. **Coding-agent capability — shared or distinct?** → **One** `CodingAgentProvider` interface in `src/attacks/common/capabilities.go`. SymJack and TrustFall target the same archetype (a coding agent with an approval/trust surface); splitting would force a target to implement two interfaces to be exercised by either. Mirrors the narrow `MCPProvider`/`BrowserProvider` precedent and reuses `MissingCapabilitySkip`.
2. **IterInject optimizer — reuse target or separate model?** → **Rule-based diagnoser in pure Go** (no model, per the paper) + **LLM optimizer that reuses the target provider** as a self-refinement loop, consistent with the single-`provider` `Execute` signature and runnable against `--provider=mock`. A separate optimizer-provider capability is deferred (see Scope Boundaries). Module lives in `adaptive/`, **not** the legacy `injection/` package — `injection/` is a deliberately-unregistered import-orphan (per the `src/attacks/all` barrel comment), while `adaptive/` houses the sibling engines (`jbfuzz`, `persona_evolve`) and is already barrel-wired.
3. **Mock coding-agent target — testutil or provider?** → **`testutil/`**, as a `MockCodingAgent` double implementing `CodingAgentProvider`, alongside the existing `MockProvider`/`MockLLMServer`. The RCE pair executes end-to-end in module `_test.go` against this double; the CLI `attack run --provider=mock` path honestly emits `SkipMissingCapability` (text mock lacks the capability) until a real coding-agent adapter exists.
4. **Disguise-record / external-content formats** → Follow the `buildInjectPayload(mode, …)` per-mode template pattern in `src/attacks/memory/poisoning.go`. Exact MemMorph record templates and Sleeper external-content formats are grounded in the full paper bodies at implementation time (only abstracts grounded so far — see Deferred questions).
5. **OWASP mapping — hand-written or generated?** → `templates/owasp_agentic_2026.yaml` is canonical: add new technique entries there, regenerate the generated map via `cmd/owasp-gen`, and keep the hand-written `TechniqueToAgenticCategories` consistent. The `owasp_agentic_drift_test.go` suite enforces YAML↔registry consistency.

---

## High-Level Technical Design

*Directional guidance for review, not implementation specification. The implementing agent should treat it as context, not code to reproduce.*

Dependency graph of implementation units:

```
U1 CodingAgentProvider capability ──┐
U2 MockCodingAgent (testutil) ──────┼──> U6 SymJack ──┐
                                    └──> U7 TrustFall ─┤
U3 MemMorph (memory mode) ───────────────────────────┤
U4 Hidden Sleeper (memory mode) ─────────────────────┼──> U8 OWASP mapping + regen ──> U9 Integration smoke tests
U5 IterInject (adaptive engine) ─────────────────────┘
```

U1/U2 are the only net-new shared infrastructure; U3–U7 are independent modules that can be built in parallel once their dependencies exist; U8 and U9 are cross-cutting closeouts.

---

## Requirements Traceability

| Origin R-IDs | Implementation Unit(s) |
|---|---|
| R1–R4 (MemMorph) | U3 |
| R5–R8 (Sleeper) | U4 |
| R9–R12 (IterInject) | U5 |
| R13, R15, R16 (SymJack) | U1, U2, U6 |
| R14, R15, R16 (TrustFall) | U1, U2, U7 |
| R17 (mock target) | U2 |
| R18 (registry/CLI) | U3–U7 (auto-register; barrel already imports these packages) |
| R19 (OWASP mapping) | U8 |
| R20 (tests / integration) | every unit's test scenarios + U9 |

---

## Implementation Units

### U1. CodingAgentProvider capability interface

**Goal:** Add the optional capability that SymJack/TrustFall type-assert against, plus its skip helper.
**Requirements:** R16; enables R13, R14.
**Dependencies:** none.
**Files:** `src/attacks/common/capabilities.go` (extend), `src/attacks/common/capabilities_test.go` (extend or add).
**Approach:** Define `CodingAgentProvider` exposing the approval/trust surface needed by both RCE modules: (a) request approval for a described operation, returning both *what the user is shown* and *what the operation actually resolves to* (symlink resolution) — this is the SymJack misrepresentation signal; (b) a folder-trust action that returns the project-defined MCP paths auto-executed on accept — the TrustFall signal. Keep the interface narrow and behavior-only; no transport detail. Reuse `SkipMissingCapability` via the existing `MissingCapabilitySkip(moduleName, "common.CodingAgentProvider")` helper. Evaluate whether `SkipNoMutationTarget` ("no_safety_step_to_hijack") fits the "agent has no approval step" sub-case; add a new typed `SkipReason` only if neither existing reason fits.
**Patterns to follow:** `MCPProvider`, `BrowserProvider`, `AudioProvider` (narrow capability interfaces); `MissingCapabilitySkip` / `MarkTextSimulation` helpers in the same file.
**Test scenarios:**
- A provider implementing `CodingAgentProvider` satisfies the type assertion; a plain `Provider` does not.
- `MissingCapabilitySkip` for this capability yields `OutcomeSkipped` + `SkipMissingCapability` naming `common.CodingAgentProvider`.
- If a new `SkipReason` is added, it serializes to its snake_case JSON string like the existing enum.
**Verification:** package compiles; capability assertion and skip helper covered by tests.

### U2. MockCodingAgent test double

**Goal:** A controllable coding-agent target so U6/U7 execute end-to-end in tests.
**Requirements:** R17.
**Dependencies:** U1.
**Files:** `src/attacks/testutil/testutil.go` (extend) or `src/attacks/testutil/coding_agent.go` (new), `src/attacks/testutil/coding_agent_test.go`.
**Approach:** Implement `MockCodingAgent` satisfying `CodingAgentProvider`. Model a minimal in-memory filesystem with symlink resolution: an approval can be configured so the *shown* destination differs from the *resolved* destination (symlink into a simulated MCP-config dir) — letting U6 assert the misrepresentation and the resulting config write. Model folder-trust so that accepting trust on a cloned-repo fixture "executes" attacker-controlled project MCP paths, recorded for U7 to assert. No real filesystem or process execution — pure in-memory simulation with recorded effects. Provide knobs to simulate a *refusing* agent (approval shows true destination / trust does not auto-execute) so modules can produce `refused`, not just `success`.
**Patterns to follow:** `MockProvider`, `MockLLMServer`, `DefaultAttackConfig` in `src/attacks/testutil/testutil.go` — constructor + recorded-calls + queued-behavior style.
**Test scenarios:**
- Approval configured with a symlinked destination resolves to the MCP-config target; the shown vs resolved paths differ.
- Approval configured honestly resolves shown == resolved (the refuse case).
- Folder-trust on a fixture repo with a project MCP path records auto-execution; trust-disabled fixture records none.
- Recorded effects are queryable after the operation (call-count / last-effect accessors).
**Verification:** mock compiles and its own tests pass; effects are observable for downstream module tests.

### U3. MemMorph memory module

**Goal:** Disguised-record memory poisoning that biases the target's autonomous tool selection.
**Requirements:** R1, R2, R3, R4.
**Dependencies:** none (extends existing memory scaffold).
**Files:** `src/attacks/memory/poisoning.go` (add `memmorph` mode + payload builder + tool-selection classifier) or `src/attacks/memory/memmorph.go` (new file, same package), `src/attacks/memory/memmorph_test.go`.
**Approach:** Register a `memmorph` mode reusing the gate→probe→inject→classify scaffold. Gate on `i_understand_risks`; require `MemoryProbe`; on stateless target emit `SkipMemoryNotRetained`. Inject a small default set (3) of records disguised as technical facts / incident reports / operational policies (new payload builder mirroring `buildInjectPayload`). Success classifier differs from `minja`: success = the target *autonomously selects the attacker-preferred tool* on a later neutral task (not an indicator-string match); refusal = it selects the legitimate tool. Emit `CleanupHint` + `injected_record_ids`; implement the `Cleaner` hook if cleanup is mechanical.
**Patterns to follow:** `MemoryPoisoningModule` mode dispatch, `buildInjectPayload`, `CleanupHint`/`injected_record_ids` metadata in `src/attacks/memory/poisoning.go`.
**Test scenarios:**
- `Covers AE1.` Missing `i_understand_risks` → `OutcomeSkipped` + `SkipGateBlocked`, no queries.
- `Covers AE2.` Stateless provider (no `MemoryProbe`) → `SkipMissingCapability`; `MemoryProbe` returning `(false,nil)` → `SkipMemoryNotRetained`.
- Default injects exactly 3 disguised records; record IDs surface in `injected_record_ids`.
- Target that selects the attacker tool → `OutcomeSuccess`; target that selects the legitimate tool → `OutcomeRefused`.
- `ProbeMemory` error → `SkipProviderError` (not `success=false`).
**Verification:** appears in `attack list`; runs against a memory-capable mock to a typed result; tests pass.

### U4. Hidden Sleeper memory module

**Goal:** Plant a dormant fabricated memory via external content; verify cross-session re-emergence.
**Requirements:** R5, R6, R7, R8.
**Dependencies:** none (extends memory scaffold; uses `SessionProvider` like memorygraft).
**Files:** `src/attacks/memory/sleeper.go` (new, same package), `src/attacks/memory/sleeper_test.go`.
**Approach:** Register a `sleeper` mode. Gate on `i_understand_risks`; require both `MemoryProbe` and `SessionProvider` (cross-session). The inject phase delivers *external content* (poisoned document/page/repo text the target ingests into memory) rather than a direct instruction. Verification happens in a **fresh** session via `NewSession`: success only when the dormant memory both re-emerges and steers an attacker-intended action; if it does not re-emerge → `OutcomeRefused`. Reuse memorygraft's cross-session trigger flow.
**Patterns to follow:** memorygraft mode in `src/attacks/memory/poisoning.go` (`SessionProvider`, `NewSession`, trigger-in-fresh-session).
**Test scenarios:**
- `Covers AE1.` Missing gate → `SkipGateBlocked`.
- `Covers AE2.` Missing `SessionProvider` → `SkipMissingCapability` naming the session requirement.
- `Covers AE4.` Memory planted in session A, does not re-emerge in fresh session B → `OutcomeRefused`, not `success`.
- Memory re-emerges in session B and steers the attacker action → `OutcomeSuccess`; verification ran in a different session ID than injection.
- `NewSession` error → `SkipProviderError`.
**Verification:** appears in `attack list`; cross-session flow exercised against a session-capable mock; tests pass.

### U5. IterInject adaptive engine

**Goal:** Feedback-optimized indirect prompt injection with a rule-based diagnoser + self-refinement optimizer under `EngineBudget`.
**Requirements:** R9, R10, R11, R12.
**Dependencies:** none (sibling of jbfuzz/persona_evolve).
**Files:** `src/attacks/adaptive/iterinject.go` (new), `src/attacks/adaptive/iterinject_test.go`, optional `templates/iterinject_seeds/` for disguise seeds.
**Approach:** Gate on `allow_experimental`. Build the feedback loop: a **rule-based diagnoser** (pure Go) produces a structured outcome label per attempt; the **optimizer** refines the payload conditioned on the full attempt history by querying the **target provider** (self-refinement — see Resolved Question 2); a synthesis step derives new disguise seeds from failure patterns. Drive iteration with `common.EngineBudget` (`budgetFromConfig` + `Clamp()`), surface clamped knobs in `Metadata["budget_clamped"]`, honor `EarlyStopOnSuccess`. On budget exhaustion without a landing payload → `OutcomeSkipped` + `SkipBudgetExceeded`; on landing → `OutcomeSuccess` with the winning payload + iteration count.
**Patterns to follow:** `src/attacks/adaptive/jbfuzz.go` (budget loop, `Clamp`, `budget_clamped` metadata, deterministic `rng_seed`, seed-corpus loading, trajectory/early-stop); `persona_evolve.go` generational loop.
**Test scenarios:**
- `Covers AE1.` Missing `allow_experimental` → `SkipGateBlocked`, no queries.
- `Covers AE3.` Budget exhausts before any payload lands → `OutcomeSkipped` + `SkipBudgetExceeded` (never `success=false`).
- `Covers AE6.` Config exceeding a hard ceiling → clamped; clamped knob names appear in `Metadata["budget_clamped"]`.
- A mock that "lands" at iteration N → `OutcomeSuccess` with winning payload and iteration count recorded.
- Deterministic `rng_seed` → reproducible iteration sequence.
- Diagnoser labels map correctly from canned mock responses (landed vs refused vs partial).
**Verification:** appears in `attack list`; runs against `--provider=mock` to a typed result respecting budget; tests pass.

### U6. SymJack coding-agent RCE module

**Goal:** Model approval-prompt symlink misrepresentation → MCP-config write.
**Requirements:** R13, R15, R16.
**Dependencies:** U1, U2.
**Files:** `src/attacks/agentic/persistence/symjack.go` (new), `src/attacks/agentic/persistence/symjack_test.go`.
**Approach:** Gate on `i_understand_risks`; require `CodingAgentProvider` (else `SkipMissingCapability`). Construct the SymJack scenario: present a benign-looking operation (e.g., file copy) whose destination resolves through a symlink into the agent's MCP-config dir. Success = the misrepresented write actually lands in the config target (observed via the mock's recorded effects); refusal = the agent surfaces the true destination / declines. Run end-to-end against `MockCodingAgent` in tests.
**Patterns to follow:** `src/attacks/agentic/persistence/rce_chain.go` and `agent_config_rewrite.go` (module shape, `CategoryPersistence`, gate handling).
**Test scenarios:**
- `Covers AE1.` Missing gate → `SkipGateBlocked`.
- `Covers AE2.` Provider without `CodingAgentProvider` → `SkipMissingCapability`.
- `Covers AE5.` Symlinked destination resolves to MCP-config dir and the write lands → `OutcomeSuccess`.
- Honest mock (shown == resolved) → `OutcomeRefused`.
**Verification:** appears in `attack list`; against text `--provider=mock` emits `SkipMissingCapability`; against `MockCodingAgent` in tests, executes end-to-end; tests pass.

### U7. TrustFall coding-agent RCE module

**Goal:** Model folder-trust auto-execution of project-defined MCP servers.
**Requirements:** R14, R15, R16.
**Dependencies:** U1, U2.
**Files:** `src/attacks/agentic/persistence/trustfall.go` (new), `src/attacks/agentic/persistence/trustfall_test.go`.
**Approach:** Gate on `i_understand_risks`; require `CodingAgentProvider`. Construct a cloned-repo fixture carrying an attacker-controlled project MCP path; the module exercises the folder-trust action and checks whether accepting trust auto-executes the attacker path. Success = attacker path executes on the trust default; refusal = trust does not auto-execute project MCP. End-to-end against `MockCodingAgent` in tests.
**Patterns to follow:** same persistence-module shape as U6; `MockCodingAgent` trust knobs from U2.
**Test scenarios:**
- `Covers AE1.` Missing gate → `SkipGateBlocked`.
- `Covers AE2.` Missing capability → `SkipMissingCapability`.
- Trust-enabled fixture with project MCP path → attacker path executes → `OutcomeSuccess`.
- Trust-disabled fixture → no auto-execution → `OutcomeRefused`.
**Verification:** appears in `attack list`; CLI text-mock run skips; mock-agent test run executes end-to-end; tests pass.

### U8. OWASP Agentic mapping + regeneration

**Goal:** Map the five new techniques to OWASP Agentic 2026 categories without drift.
**Requirements:** R19.
**Dependencies:** U3, U4, U5, U6, U7 (technique names must exist).
**Files:** `templates/owasp_agentic_2026.yaml` (add technique entries), `src/compliance/owasp_agentic_generated.go` (regenerated via `cmd/owasp-gen`), `src/compliance/owasp_agentic.go` (keep hand-written map consistent if still the runtime lookup).
**Approach:** Add entries to the canonical YAML — `memmorph`→ASI06, `sleeper`→ASI06/ASI10, `iterinject`→ASI01, `symjack`/`trustfall`→ASI01 (plus any agent-specific categories the technique warrants). Regenerate the generated map. Confirm whether runtime `TechniqueToAgenticCategories` reads the hand-written or generated map (CLAUDE.md notes v0.10.0 planned the switch) and keep whichever is authoritative correct; the drift test enforces consistency.
**Patterns to follow:** existing entries in `templates/owasp_agentic_2026.yaml`; `cmd/owasp-gen/main.go`; `src/compliance/owasp_agentic_drift_test.go`.
**Test scenarios:**
- Each new technique name resolves to at least one OWASP Agentic category via the runtime lookup.
- `owasp_agentic_drift_test.go` passes (YAML ↔ generated/registry consistent) — regenerate, don't hand-edit the generated file.
- `go generate ./... && git diff --exit-code` (or the repo's drift check) is clean after regeneration.
**Verification:** drift test green; new techniques appear in compliance mapping.

### U9. Integration smoke tests + registry collision check

**Goal:** One end-to-end smoke test per new family; confirm registry wiring.
**Requirements:** R20, R18.
**Dependencies:** U3–U8.
**Files:** `src/attacks/integration/integration_test.go` (extend), no barrel change expected (`memory`, `adaptive`, `agentic/persistence` already imported in `src/attacks/all`).
**Approach:** Add a smoke test per family (memory: memmorph/sleeper; adaptive: iterinject; persistence: symjack/trustfall) following the existing `RUN_INTEGRATION`-gated, `t.Skip`-when-unset pattern. Run the name-collision test to confirm no duplicate registrations. Verify `attack list` enumerates all five and `attack run` reaches each. Confirm the barrel needs no edit (all three target packages already imported); if any new package were introduced, append it alphabetically — but none is.
**Patterns to follow:** `src/attacks/integration/integration_test.go` (`RUN_INTEGRATION` gate, `t.Skip`); `TestNoNameCollisions` referenced in the `src/attacks/all` barrel comment.
**Test scenarios:**
- `RUN_INTEGRATION` unset → all new smoke tests `t.Skip` (CI silent by default).
- `RUN_INTEGRATION=1` against `MockLLMServer`/mocks → each family runs to a typed outcome.
- `attack list --json` includes `memmorph`, `sleeper`, `iterinject`, `symjack`, `trustfall`.
- Name-collision test passes (no duplicate registry names).
**Verification:** `go build -o llmrecon ./src/main.go` succeeds; `attack list` shows 5 new modules; `RUN_INTEGRATION=1 go test ./src/attacks/integration/...` green; full `go test ./src/attacks/... ./src/compliance/...` green.

---

## Key Technical Decisions

- **IterInject → `adaptive/`, not `injection/`** (origin said `injection/`): the barrel deliberately omits `injection/` as a legacy unregistered orphan; `adaptive/` houses the sibling engines and is barrel-wired. Correct placement also means zero barrel edits.
- **Single `CodingAgentProvider` capability** for both RCE modules: same target archetype; avoids forcing a target to implement two interfaces. (Resolved Q1.)
- **IterInject optimizer reuses the target provider** (self-refinement) with a pure-Go diagnoser: fits the single-`provider` `Execute` signature, runs against mock, stays honest. Separate optimizer-provider deferred. (Resolved Q2.)
- **Mock coding-agent lives in `testutil/`**; CLI runs honestly skip until a real adapter exists. (Resolved Q3.)
- **Memory modules extend the existing `MemoryPoisoningModule` mode scaffold** rather than new structs — matches how minja/memorygraft/injecmem coexist; MemMorph supplies a distinct tool-selection success classifier.
- **YAML is canonical for OWASP mapping**; regenerate, never hand-edit the generated map; drift test enforces. (Resolved Q5.)
- **No barrel edits required** — all three target packages are already imported in `src/attacks/all`.

---

## System-Wide Impact

- **Registry / CLI:** five new modules surface in `attack list` / `attack run`. No new provider; RCE pair skips on text providers by design.
- **Compliance:** new technique→category entries; drift test must stay green.
- **ML / bandit:** new modules emit the existing 3-state taxonomy; `skipped` rows already excluded from reward aggregation — no pipeline change needed.
- **Honesty invariant:** every new `Execute()` does real work or returns a typed skip; every modified package gains tests on its modified surface.

---

## Scope Boundaries

- Not in v0.11.0; built on a v0.12.0 branch off the stabilized baseline; must not modify stabilization scope.
- No real coding-agent provider adapter (Claude Code / Cursor / Copilot bridges) — only the testutil mock.
- No compliance/taxonomy refresh (Microsoft red-team v2.0 failure modes, OX MCP supply-chain enrichment) — captured in `scratch-new-attacks-2026-06.md`.
- Other surveyed techniques excluded: Copirate 365 (CVE-2026-24299), ASPI (2605.17324), QueryIPI (2510.23675), Structural Template Injection (2602.16958), Skill-Inject (2602.20156).

### Deferred to Follow-Up Work

- A separate optimizer-provider capability for IterInject (beyond target self-refinement).
- A production `CodingAgentProvider` adapter so SymJack/TrustFall run via the CLI against real coding agents.

---

## Dependencies / Assumptions

- `src/attacks/common` capabilities (`MemoryProbe`, `SessionProvider`, `Cleaner`, `MissingCapabilitySkip`) and `EngineBudget` + hard ceilings are present and reusable (verified by reading `src/attacks/common/capabilities.go` and `src/attacks/memory/poisoning.go`).
- `src/attacks/all` already imports `memory`, `adaptive`, `agentic/persistence` (verified).
- `cmd/owasp-gen` regenerates `src/compliance/owasp_agentic_generated.go` from `templates/owasp_agentic_2026.yaml` (verified files exist; regenerate during U8).

---

## Outstanding Questions

### Deferred to Implementation

- [Affects U3, U4][Needs research] Exact MemMorph disguised-record templates and Sleeper external-content formats — confirm against the full paper bodies (arXiv 2605.26154 / 2605.15338); only abstracts grounded at plan time.
- [Affects U1][Technical] Whether the "agent has no approval step" sub-case reuses `SkipNoMutationTarget` or warrants a new typed `SkipReason` — decide when implementing the mock's refuse path.
- [Affects U8][Technical] Whether runtime `TechniqueToAgenticCategories` reads the hand-written or generated map in the current tree — confirm and keep the authoritative one correct.
- [Affects U5][Technical] Whether IterInject's self-refinement optimizer prompt should be capped to a sub-budget of total queries so diagnosis+optimization don't starve the attack attempts — tune against the budget loop during implementation.
