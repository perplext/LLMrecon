---
title: "feat: v0.11.0 — the stabilization release"
type: feat
date: 2026-05-04
brainstorm: docs/brainstorms/2026-05-04-v0-11-0-stabilization-brainstorm.md
follows: docs/plans/2026-05-02-feat-v0-10-0-phased-execution-plan.md
issues_in_scope: [166, 168, 170, 171, 174, 177, 198, 222]
new_issues_to_file:
  - "honesty: update/verification.go silently skips checksum verification"
  - "honesty: triage 4 build-ignored cmd/*.go files (sign/owasp/reporting/template_security)"
  - "honesty: triage 48 production TODOs"
  - "tests: cover src/template (76 files, 0 tests)"
  - "tests: cover src/bundle (42 files, 0 tests)"
  - "tests: cover src/security (46 files, 0 tests)"
  - "tests: cover src/api (17 files, 0 tests, REST surface)"
  - "docs: attic obsolete versioned docs (V0_2_0_*, V0_3_0_*, etc)"
  - "docs: fold 5 overlapping quickstart variants into one"
  - "docs: regen DOCUMENTATION-INDEX.md"
  - "cli: attack run --provider=openai / --provider=anthropic plumbing"
reviewers: [dhh, kieran, simplicity]
---

# v0.11.0 — The Stabilization Release

## Thesis

v0.10.0 was the **honesty release**: stop the binary from lying.
v0.11.0 is the **stabilization release**: make what's there work,
prove it works with tests, and stop carrying dead weight.

## Single ordering rule

**Honesty cleanup ships before tests.** A test against a function
that's silently lying about success is a test you have to rewrite
when you fix the bug. Fix the bugs first, then write the tests.

Corollary: doc consolidation also waits for honesty cleanup, because
some of the docs that need atticking are "this feature is half-
implemented" notes that become unnecessary once the feature is real
or removed.

## Honesty invariant (non-negotiable, repeated from v0.10.0)

Three checks at release time. If any fails, hold the tag:

1. **No "not implemented" success.** No code path returns
   `(success-shaped value, nil)` while a TODO or stub is in the body.
2. **No package > 15 files without tests.** Every load-bearing
   subsystem has at least integration-style tests over its public
   surface.
3. **No top-level doc claims a feature the binary doesn't ship.** The
   v0.10.0 #175 README smoke test extends to `docs/` entry-point pages.

## Scope

Effort estimates assume single-author cadence with normal interruptions.

| Phase | Items | Effort |
|---|---|---|
| 1. Honesty cleanup | verification.go, build-ignored cmd files, top TODOs | 5 days |
| 2. CLI provider wiring | #166 follow-up | 5 days |
| 3. Test coverage | template, bundle, security, api | 10 days |
| 4. Doc consolidation | attic + fold + regen index | 3 days |
| 5. Deferred research | #168 Purger, #170 embed fitness, #171 MCTS | 8 days |
| 6. CI hygiene | #198 Node 24, #222 gosec 2.26 | 2 days |
| 7. Release | rc1 + tag | 1 day |

**Total: 34 days of focused work (~7–9 weeks at single-author cadence).**

## Phase 1 — honesty cleanup (ships first)

### 1a. `update/verification.go` checksum verification

**Problem**: `src/update/verification.go:53,56` returns success while
two TODOs explicitly say "implement manifest checksum verification"
and "implement checksum verification using component checksums". This
is the same honesty class v0.10.0 #174 fixed for `update apply`.

**Fix**: implement using the existing `Components` field on
`PackageManifest` (per the TODO comment), OR if the manifest design
genuinely doesn't carry the data, error honestly with a typed
`ErrVerificationUnsupported` until the manifest format catches up.

**Acceptance**: `update verify` either verifies and returns success
deterministically OR returns a non-nil error. Never both succeeds and
skips silently.

### 1b. Triage 4 build-ignored `cmd/*.go` files

`src/cmd/{sign,owasp,reporting,template_security}.go` carry
`//go:build ignore` since 2025-08-26. Each gets one of three
dispositions:

- **Re-enable + finish** if there's a real consumer or the v0.10.0
  release notes promised it (e.g., `bundle create --sign` cosign
  flow listed under #177's deferred items).
- **Attic** to `attic/v0-7-0-cmd-disabled/` with a README, mirroring
  v0.10.0 #177's pattern with the bundle disabled files.
- **Document removal** in the CHANGELOG.

Pre-triage suspicion (subject to change after reading each file):
- `sign.go` → re-enable + finish (cosign integration is real follow-up)
- `owasp.go` → attic (commands are obsolete; `attack run` covers the
  same ground via the registered modules)
- `reporting.go` → attic (no consumer; `attack run --emit-jsonl` is
  the canonical reporting surface)
- `template_security.go` → re-enable if it's still wanted, else
  attic

**Acceptance**: zero `//go:build ignore` files in `src/cmd/`.

### 1c. Top production TODOs

48 production TODOs total. Triage by impact:

- **Honesty-class** (returns success while incomplete): fix or convert
  to typed error.
- **Feature-class** (a feature is announced but partial): finish or
  delete advertisement.
- **Comment-noise** (note-to-self that's been there 6+ months without
  movement): delete the TODO; write the actual fact.

Specifically targeted in scope:
- `src/bundle/delta.go:438,540` — patch application + tar.gz delta
  bundles. Believed unused; verify by checking if any consumer calls
  `applyPatch` / `compressDelta`. If unused → attic the delta path.
- `src/bundle/export.go:755,760,837,842,847` — key management,
  encryption, configurable defaults. Same audit: who calls?
- `src/update/customization.go:144` — "actual update logic here". Audit.
- `src/update/check.go:97` — "Add template and module versions". This
  one is feature-class; finish it (small).
- `src/cmd/package.go:56` — `// TODO: Uncomment when rootCmd is
  properly defined`. The `rootCmd` situation has been resolved for
  releases. Either uncomment or remove the dead command file.

**Acceptance**: zero TODOs in `src/cmd/`, `src/update/`, `src/bundle/`
production paths. The `// TODO: Make configurable` style ones in
deeper subsystems can stay if they describe genuine future work; they
go through code review.

## Phase 2 — CLI provider wiring (`attack run --provider=openai`)

**Closes**: v0.10.0 #166 follow-up — adapters exist (Tier A + B
shipped); `attack run` only accepts `--provider=mock`.

### Pieces

1. **Provider construction from CLI flags**. `--provider=openai`
   reads `OPENAI_API_KEY` from env (or `--api-key` flag) and
   constructs a `core.OpenAIProvider`. Same for Anthropic.
2. **Bridge promotion at the call site**. `attack run` already
   receives a `common.Provider`; the constructor needs to call
   `bridge.WrapCore(coreProvider)` and pass the result through.
3. **Config file support**. `~/.LLMrecon.yaml` (already used by other
   commands) reads `providers.openai.api_key` etc. so the operator
   doesn't have to pass `--api-key` every time.
4. **Tests**. End-to-end with mock HTTP server (httptest) covering
   provider construction, bridge promotion, and capability surface.
   Live API tests gated by `RUN_INTEGRATION=1` + the relevant
   `_API_KEY` env var.

### Acceptance

- `llmrecon attack run --module=h_cot --provider=openai
  --metadata=i_understand_risks=true` runs against `o4-mini` (or
  whatever model the operator configured) and returns a typed
  `AttackResult` with `Outcome=Success/Refused/Skipped`.
- Same for `--provider=anthropic` against `claude-opus-4-5`.
- H-CoT against Anthropic deterministically emits `SkipSignatureGated`.
- Failure modes are distinct: missing API key → `SkipPreconditionFailed`,
  rate limit → retried via `core.RetryableQuery` then
  `SkipProviderError` if exhausted, content policy refusal →
  `OutcomeRefused`.

## Phase 3 — test coverage on load-bearing packages

**Goal**: bring `template`, `bundle`, `security`, `api` from 0% to
"covers the happy path + the failure modes the operator would notice".
Not 80% coverage; "the load-bearing functions have tests."

### Per-package targets

#### `src/template/` (76 files, currently 0 tests)

- `loader.go` — manifest parsing, schema validation, error paths
  (malformed YAML, missing required fields, version mismatch).
- `validator.go` — template-format validation against schema.
- `cache.go` — read-after-write, eviction, concurrent access.
- `repository/` — file-system-backed and in-memory implementations.

Estimated 8–12 test files, ~30 test cases.

#### `src/bundle/` (42 files, 0 tests)

- `bundle.go::OpenBundle` / `loader.go::LoadBundle` — round-trip with
  a known-good fixture; rejects malformed manifest.
- `signature.go::VerifyBundle` / `VerifyBundleChecksums` — happy +
  tampered cases.
- `validator.go::DefaultBundleValidator` — each level (basic,
  standard, strict) hits the right internal checks.
- `compression.go::DecompressBundle` — zip + tar.gz, with the same
  zip-slip protections we added in v0.10.0 #174 Tier 2.

Estimated 6–8 test files, ~25 test cases.

#### `src/security/` (46 files, 0 tests)

- `communication/cert_pinning.go` — pin verification on fresh +
  resumed sessions (the G123 finding from #222).
- `keystore/` — key generation, retrieval, rotation, export/import.
- `prompt/` — injection-protection rule matching, threshold-based
  blocking.
- `vault/manager.go` — secret read/write, encryption-at-rest.

Estimated 8–10 test files, ~30 test cases. Higher density because
this is crypto-adjacent.

#### `src/api/` (17 files, 0 tests, REST surface)

- Auth middleware (if any survived #180).
- Per-endpoint happy + error paths.
- Request validation, JSON shape, error response shape.

Estimated 4–6 test files, ~20 test cases.

### Phase 3 acceptance

- Each of the four packages has at least one test file in its
  immediate directory (so `go test ./src/<pkg>/...` is non-trivial).
- `go test -cover ./src/template/... ./src/bundle/... ./src/security/...
  ./src/api/...` reports >= 30% statement coverage on each.
- CI runs the new tests with `-race` like the existing suite.

## Phase 4 — doc consolidation

### 4a. Attic obsolete versioned docs

Move to `attic/v0-2-and-v0-3-docs/` with a README:

- `docs/V0_2_0_MONITORING_CHECKLIST.md`
- `docs/V0_2_0_PERFORMANCE_REPORT.md`
- `docs/V0_2_0_STABILIZATION_PROGRESS.md`
- `docs/V0_3_0_REFINED_PLAN.md`
- `docs/V0_3_0_ROADMAP.md`
- `docs/V0_3_0_TODO_SUMMARY.md`
- `docs/REVISED_ROADMAP.md`
- `docs/STABILIZATION_PLAN.md`
- `docs/STRATEGY_ASSESSMENT.md`
- `docs/EARLY_ADOPTER_COMMUNICATIONS.md`
- `docs/EARLY_ADOPTER_ONBOARDING.md`
- `docs/DEPLOYMENT_STATUS.md`
- `docs/compilation_fixes_summary.md`
- `docs/compilation_fixes.md`
- `docs/access_control_roadmap.md` (RBAC was deleted in v0.10.0 #180)

### 4b. Fold 5 quickstarts into one

Pick `docs/quickstart.md` as the canonical quickstart. Fold the
unique content from `GETTING_STARTED.md`, `installation.md`,
`QUICK-START-REFERENCE.md`, `QUICK_REFERENCE.md` into it. Delete the
others or stub-redirect.

### 4c. Regenerate `DOCUMENTATION-INDEX.md`

Auto-generate from current `docs/` contents grouped by topic
(getting-started, architecture, attack-modules, bundles, updates,
testing, internals).

### 4d. README pass for `docs/` entry-point claims

The v0.10.0 #175 README smoke test catches CLI drift in the README.
Extend the same idea to `docs/quickstart.md` — every command
documented there must resolve via `rootCmd.Find`.

### Phase 4 acceptance

- `docs/` has 60–70 markdown files, down from 102.
- `docs/quickstart.md` is the single entry point linked from README.
- `DOCUMENTATION-INDEX.md` is current and references existing files.
- `docs/quickstart.md` smoke test added to `src/cmd/readme_smoke_test.go`.

## Phase 5 — deferred research items

These have been pending one release cycle. Each is opt-in (operator
must explicitly enable); none changes existing default behavior.

### #168 — Purger interface

Automated cleanup of memory-poisoning implants (the `minja` /
`memorygraft` / `injecmem` modules). Adds a `common.Purger` interface;
modules that implement `Cleanup(ctx, recordIDs)` get called from the
CLI's `attack run --purge-after` flag.

Estimated 3 days.

### #170 — Embedding-fitness opt-in

`jbfuzz` and `persona_evolve` currently use heuristic fitness
(refusal-keyword absence + objective-word match). Embedding fitness
trades a 50-100ms cosine-similarity call for a more semantic signal.
Opt-in via `--metadata=fitness=embedding`.

Estimated 3 days.

### #171 — MCTS-Explore selection strategy

Replaces UCB1+restart with Monte Carlo Tree Search exploration in
`jbfuzz`. Opt-in via `--metadata=selection=mcts`.

Estimated 2 days.

## Phase 6 — CI hygiene

### #198 — Node.js 24 actions bump

7 action pins across 4 workflow files. Bump to Node-24-compatible
versions, or set `FORCE_JAVASCRIPT_ACTIONS_TO_NODE24=true` workflow-
level. Hard cutoff is June 2026 (forced) / September 2026 (removed).

Estimated 1 day.

### #222 — gosec 2.26 cleanup pass

19 findings to address before bumping the scanner:

- 1 real G123 (TLS session resumption / `VerifyPeerCertificate` in
  `cert_pinning.go`) — fix.
- ~14 `#nosec` annotations on G703 false positives where the path
  is already `filepath.Clean`'d.
- 4 `#nosec G101` annotations on attack-template strings.
- 1 G118 stylistic on context.Background in a goroutine.

Estimated 1 day.

## Phase 7 — Release

- Cut `v0.11.0-rc1` from main once phases 1–4 complete.
- Two-week soak in CI / available for community testing.
- Phases 5 + 6 land during soak (no behavior change to existing
  default paths).
- Cut `v0.11.0` from rc1 + any RC fixes.

## Risks worth tracking

| Risk | Mitigation |
|---|---|
| Honesty cleanup uncovers more hidden lies | Time-box to 5 days; if more found, file follow-up issues for v0.12.0 rather than expanding scope |
| Test additions catch real bugs that block release | Expected; treat each as a separate PR (may displace some Phase 5 items) |
| `attack run --provider=openai` plumbing turns out to need #167 bridge changes | Bridge already promotes; CLI plumbing is purely command-side. Risk: low. |
| Doc consolidation breaks external links | Add redirect stubs at old paths for the canonical-quickstart pick |

## Out of scope (explicitly)

- New attack modules. Bottleneck is wiring through existing 50+
  modules to real providers, not inventing more.
- New OWASP categories. 2026 update is in.
- Auth/RBAC restoration. v0.10.0 #180 removed with intent — bring back
  when there's a real consumer.
- Distributed coordination, monitoring dashboards, production-scale
  infrastructure from v0.2.0. None has tests; none has a documented
  operator. Either revive with intent in v0.12.0+ or leave atticked.

## References

- v0.10.0 plan: `docs/plans/2026-05-02-feat-v0-10-0-phased-execution-plan.md`
- v0.10.0 release: tag `v0.10.0`, commit `a4170b0` (or current main HEAD).
- v0.11.0 brainstorm: `docs/brainstorms/2026-05-04-v0-11-0-stabilization-brainstorm.md`
- v0.10.0 sanity audit: this plan's frontmatter `new_issues_to_file` list.
