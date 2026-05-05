---
title: "feat: v0.11.0 — the stabilization release"
type: feat
date: 2026-05-04
revision: v2 (post-review, 2026-05-04)
brainstorm: docs/brainstorms/2026-05-04-v0-11-0-stabilization-brainstorm.md
follows: docs/plans/2026-05-02-feat-v0-10-0-phased-execution-plan.md
issues_in_scope: [166, 174, 177, 198, 222]
issues_deferred_to_v0_12_0: [168, 170, 171]
new_issues_to_file:
  - "honesty: update/verification.go silently skips checksum verification"
  - "honesty: stale 'v1: mock only' help text in src/cmd/attack.go (current bug, not v0.11 work)"
  - "honesty: triage 4 build-ignored cmd/*.go files (sign/owasp/reporting/template_security)"
  - "honesty: attic src/cmd/package.go (262-line orphan, never registered)"
  - "honesty: attic src/bundle/delta.go (dead-by-induction; CompressDelta is no-op)"
  - "tests: cover src/template (76 files, 0 tests)"
  - "tests: cover src/bundle (42 files, 0 tests)"
  - "tests: cover src/security (46 files, 0 tests, public-function checklist)"
  - "docs: attic obsolete versioned docs + fold quickstarts (one PR)"
  - "ci: lint for 'return nil' adjacent to 'TODO' (catch the next honesty regression)"
reviewers: [dhh-rails-reviewer, kieran-rails-reviewer, code-simplicity-reviewer]
---

# v0.11.0 — The Stabilization Release

## Thesis

v0.10.0 was the **honesty release**: stop the binary from lying.
v0.11.0 is the **stabilization release**: make what's there work,
prove it works with tests, and stop carrying dead weight.

## Single ordering rule

> A test against a function that's silently lying about success is a
> test you have to rewrite when you fix the bug.

Honesty cleanup ships first. Tests come after. New code (CLI provider
wiring) ships after the tests so it lands on a tested baseline, not a
moving target.

Refinement (per review): for TODOs whose reachability is unknown, a
30-minute probing test is faster than a code-reading audit. Let the
test answer "is this a live lie or dead code?" then fix or delete.

## Honesty invariant (release-blocking)

Three checks at release time. If any fails, hold the tag.

1. **No "not implemented" success** anywhere on a CLI-reachable path.
2. **Every package modified during v0.11.0 has tests on its modified
   surface.** Scoped to this release; the tree-wide quantifier was
   wishful.
3. **The README and `docs/quickstart.md` smoke tests pass.** The
   v0.10.0 #175 pattern, extended to one more entry-point doc.

## Scope (4 phases, 25 days focused work)

Reviewers convergent on cutting Phase 5 (deferred research) and the
ceremony phases (Phase 6 + 7 from the v1 plan). Theme conflict:
"stabilization release" + "8 days of opt-in research features" can't
both be true. Cut.

Effort assumes single-author cadence. **25 working days = ~7 weeks
calendar with the rc soak as wall time, not work time.**

| Phase | Items | Days |
|---|---|---|
| 1. Honesty cleanup | verification.go, attack.go help text, build-ignored cmds, named TODOs, CI honesty lint | 5 |
| 2. Test coverage | template, bundle, security (public-function checklist) | 14 |
| 3. CLI provider wiring | finish what's already started in attack.go | 3 |
| 4. Doc consolidation | attic + fold quickstarts + smoke test (one PR) | 3 |

Soak window (overlapping with rc1): #198 Node 24 actions bump + #222
gosec 2.26 cleanup. ~2 days total, no phase header.

## Phase 1 — honesty cleanup (5 days)

### 1a. `update/verification.go` checksum verification

**Problem**: returns success while skipping manifest + component
checksum verification (two TODOs at lines 53, 56).

**Fix**: implement using the existing `Components` field on
`PackageManifest`, OR error honestly with a typed
`ErrVerificationUnsupported` until the manifest format catches up.

**Acceptance gate** (per Kieran):

```go
// src/update/verification_test.go
func TestVerifyUpdate_DetectsTamperedComponent(t *testing.T) {
    // construct UpdatePackage with a deliberately-wrong checksum;
    // assert err != nil and result.Success == false.
}
func TestVerifyUpdate_AcceptsCorrectChecksums(t *testing.T) {
    // construct UpdatePackage with correct checksums;
    // assert err == nil and result.Success == true.
}
```

Without these gates, "deterministic" is a vibe.

### 1b. Triage 4 build-ignored `cmd/*.go` files (ORDERED BEFORE 1c)

`src/cmd/{sign,owasp,reporting,template_security}.go` carry
`//go:build ignore` since 2025-08-26. **This phase ships before 1c**
because TODOs inside build-ignored files would otherwise block 1c's
acceptance.

Per-file dispositions (verify-then-decide, don't pre-commit):

- `sign.go` → re-enable + finish OR attic. The cosign-integration
  story under #177 deferred items was the original consumer.
- `owasp.go` → attic. `attack run` covers the same surface via
  registered modules.
- `reporting.go` → attic. `attack run --emit-jsonl` is the canonical
  reporting surface.
- `template_security.go` → attic unless a real consumer surfaces.

**Acceptance**: zero `//go:build ignore` files in `src/cmd/`. Add a
CI grep guard so the rule sticks.

### 1c. Named TODO list (no quantifier acceptance)

Per Simplicity's correction: don't promise "zero TODOs in three
directories." Promise specific named-TODO resolution.

| File:Line | Disposition |
|---|---|
| `src/update/verification.go:53,56` | Fix (1a) |
| `src/update/customization.go:144` | Probing test → fix or delete |
| `src/cmd/attack.go:58,60,78,82,105` | Fix stale "v1: mock only" help text — current honesty bug |
| `src/cmd/package.go:56,132,187` | Attic the file (262-line orphan, never registered) |
| `src/bundle/delta.go:438,540` | Attic the delta path (dead by induction; verify by test before deletion) |
| `src/update/check.go:97` | Finish (small — add template + module versions) |

Remaining TODOs in deeper subsystems get filed as a single tracking
issue for v0.12.0. Don't grep-hunt the tree.

### 1d. CI honesty lint (DHH addition)

Add `scripts/verify-honesty.sh` (called from CI):

```bash
# Flag any function body containing both a TODO and a 'return nil'.
git grep -B5 -A0 "return nil" -- '*.go' | \
  awk '/TODO/{p=1} /return nil/{if(p)print; p=0}'
```

Treat any match as a CI warning the first month, then a build error.
Catches the next "I'll get to it" honesty regression at PR time.

## Phase 2 — test coverage (14 days)

Reality check: 4 packages × ~30% coverage in 10 days was optimistic
(per Kieran's audit of file counts and fixture costs). Bump to 14
days, drop `src/api/` from scope (REST surface is unused per the
v0.10.0 #180 audit).

**Goal**: load-bearing functions have tests. Not a coverage number.

### `src/template/` (76 files)

- `loader.go` — manifest parsing, schema validation, error paths.
- `validator.go` — template-format validation against schema.
- `cache.go` — read-after-write, eviction, concurrent access.
- `repository/` — file-system + in-memory implementations.

Estimated 12–16 test files, ~30 cases. 5 days.

### `src/bundle/` (42 files)

- `bundle.go::OpenBundle` / `loader.go::LoadBundle` — round-trip with
  known-good fixture; reject malformed manifest.
- `signature.go::VerifyBundle` / `VerifyBundleChecksums` — happy +
  tampered.
- `validator.go::DefaultBundleValidator` — each level (basic,
  standard, strict) routes to the right internal checks.
- `compression.go::DecompressBundle` — zip + tar.gz with the
  zip-slip protections from v0.10.0 #174 Tier 2.

Estimated 8–10 test files, ~25 cases. 4 days.

### `src/security/` (46 files) — checklist, not coverage %

Statement coverage is the wrong metric for crypto. A test suite
hitting `keystore.Generate()` happy paths can clear 30% while never
touching key rotation under concurrent reads, cipher constructors
with corrupted ciphertext, cert pinning with **resumed** TLS
sessions, or vault encryption-at-rest round-trip.

**Acceptance**: every public function in `keystore`, `vault`,
`communication/cert_pinning`, `prompt` has at least one happy-path
test AND at least one adversarial-input test. No `if err != nil`
branch on a crypto operation goes untested.

Including the cert pinning fix from #222 G123 (TLS session resumption
bypass).

Estimated 10–12 test files, ~30 cases. 5 days.

### Acceptance for the phase

- Each package has tests in its immediate directory; `go test
  ./src/<pkg>/...` is non-trivial.
- Modified-package coverage check passes (honesty invariant #2).
- All tests pass under `-race`.

## Phase 3 — CLI provider wiring (3 days)

**What's already done** (per Kieran's audit):

- `src/cmd/attack.go:273-312` has the `--provider=openai` and
  `--provider=anthropic` switch.
- `bridge.WrapCore(p)` is in the path.
- Env-key reading via `OPENAI_API_KEY` / `ANTHROPIC_API_KEY`.
- Model fallback via env vars.
- `src/cmd/attack_test.go` (297 lines) covers wiring.

**What's actually missing**:

1. **`--api-key` flag**. Today only env vars work; flag is promised
   in the help text but not registered.
2. **Failure-mode httptest matrix**. The current tests cover happy
   wiring; not the failure modes the v0.10.0 #166 acceptance promised:
   - missing API key → `SkipPreconditionFailed`
   - rate limit → `RetryableQuery` retry → `SkipProviderError` if
     exhausted
   - content policy refusal → `OutcomeRefused`
   - H-CoT against Anthropic → `SkipSignatureGated`
3. **Stale help text** ("v1: mock only") — covered in 1c.

### Acceptance

- All four httptest scenarios pass against fakes.
- Help text matches reality.
- `--api-key` flag works and takes precedence over env.

NOT IN SCOPE (per DHH + Simplicity): config file (`~/.LLMrecon.yaml`
`providers.openai.api_key`). Env + flag covers every documented use
case. Add when an operator asks.

## Phase 4 — doc consolidation (3 days, one PR)

Per Simplicity: 4a/4b/4c/4d sub-numbering was ceremony. One PR
covers:

- **Attic 15 obsolete versioned docs** to `attic/v0-2-and-v0-3-docs/`
  with a README:
  - `V0_2_0_*` (3 files), `V0_3_0_*` (3 files)
  - `REVISED_ROADMAP.md`, `STABILIZATION_PLAN.md`,
    `STRATEGY_ASSESSMENT.md`
  - `EARLY_ADOPTER_*.md` (2 files), `DEPLOYMENT_STATUS.md`
  - `compilation_fixes*.md` (2 files)
  - `access_control_roadmap.md` (RBAC was deleted in v0.10.0 #180)
  - For any URL referenced from old release notes: leave a forwarding
    stub.

- **Fold 5 overlapping quickstarts into `docs/quickstart.md`**:
  `GETTING_STARTED.md`, `installation.md`, `QUICK-START-REFERENCE.md`,
  `QUICK_REFERENCE.md` → fold into `quickstart.md`. Delete or stub.

- **Regenerate `DOCUMENTATION-INDEX.md`** OR delete it. If hand-
  curated → kill (it'll rot). If auto-generatable from `docs/`
  contents grouped by topic → keep with a build hook.

- **Extend the README smoke test to `docs/quickstart.md`**. Every
  command documented in quickstart.md must resolve via
  `rootCmd.Find`.

### Acceptance (per Simplicity-corrected invariant #3)

- `docs/` has 60–70 markdown files, down from ~85 top-level / 110
  recursive.
- Quickstart smoke test passes in `src/cmd/readme_smoke_test.go`.

## Soak-window housekeeping (during rc1, no phase header)

These are routine maintenance, not release-blocking work. They land
during the rc soak via standalone PRs:

- **#198 — Node.js 24 actions bump.** 7 action pins across 4 workflow
  files. ~1 day.
- **#222 — gosec 2.26.x cleanup.** 1 real G123 fix in
  `cert_pinning.go` (covered by Phase 2 security tests anyway), ~14
  `#nosec` annotations on G703 false positives, 4 `#nosec G101`
  annotations, 1 stylistic G118. ~1 day.

## Release

- Cut `v0.11.0-rc1` from main once Phases 1–4 complete.
- Two-week soak. Wall time, not work time.
- Soak-window housekeeping (Node 24, gosec) lands during soak.
- Cut `v0.11.0` from rc1 + any RC fixes.

## Risks worth tracking (one paragraph, not a table)

If Phase 1 surfaces work that exceeds its 5-day budget, we drop
`src/security/` from Phase 2 and ship Phases 1+2 (template + bundle
only) + 3 + 4. We do NOT extend the release calendar. Per DHH:
"file follow-up issues for v0.12.0 rather than expanding scope." If
honesty cleanup uncovers more lies than expected, that's the whole
point of the phase, not a risk.

## v0.12.0 stub plan (DHH addition)

**v0.12.0 — the coverage release.** Targets:

- Phase 5 from this plan's v1: #168 Purger interface, #170
  embedding-fitness opt-in, #171 MCTS-Explore selection.
- Bring `src/api/`, `src/reporting/`, `src/repository/`, `src/version/`
  to the same coverage bar v0.11.0 set for `template`/`bundle`/
  `security`.
- Resolve the v0.12.0-tracked TODO sweep from v0.11.0 Phase 1c.
- #174 Tier 3 binary self-replace (deferred from v0.10.0).

This stub is the discipline that makes the v0.11.0 deferrals real.
Without it, "we'll get to that next" never lands.

## Out of scope (explicitly)

- New attack modules. Bottleneck is real-provider wiring, not
  invention.
- New OWASP categories. 2026 update is in.
- Auth/RBAC restoration. v0.10.0 #180 removed with intent.
- Distributed coordination, monitoring dashboards, v0.2.0-era
  infrastructure. None has tests; none has a documented operator.
  Either revive in v0.12.0+ with a real consumer, or stay atticked.
- Config-file support for provider credentials.
- `src/api/` test coverage (deferred to v0.12.0; surface is unused).

## v1 → v2 changelog

Reviewers (DHH, Kieran, Simplicity) converged on multiple cuts.
Captured here so the rationale survives if anyone restores a v1 item:

- **Phase 5 dropped entirely.** Theme conflict — "stabilization
  release" + opt-in research features can't both be true.
  Three deferred items moved to v0.12.0 stub.
- **Phase 6 + 7 absorbed into soak-window housekeeping.** Two
  unrelated dependency bumps + a release checklist were not real
  phases.
- **Phase 4 sub-numbering collapsed into one PR.** 4a/4b/4c/4d
  shared review surface and risk; the breakdown was ceremony.
- **Phase 2 (now Phase 3) re-scoped from 5 days to 3.** Already half-
  done in `src/cmd/attack.go`; plan now targets only what's missing.
- **Phase 3 (now Phase 2) re-scoped to 14 days from 10**, dropped
  `src/api/` (REST surface unused; defer to v0.12.0).
- **`src/security/` acceptance changed from "30% coverage" to
  "public-function checklist."** Statement coverage is the wrong
  metric for crypto.
- **Phase 2 (CLI wiring) reordered after Phase 2 (tests)** so we
  test against a baseline not a moving target.
- **Phase 1 added 1d** — CI honesty lint (`return nil` near `TODO`).
- **Phase 1c reframed** from "zero TODOs in three directories" to a
  named-TODO list with explicit dispositions.
- **Honesty invariant #2 scoped** to "modified packages" not
  tree-wide.
- **Honesty invariant #3 scoped** to README + quickstart smoke
  tests, not all docs.
- **Effort math restated honestly**: 25 days focused = 7 weeks
  calendar with rc soak as wall time.
- **v0.12.0 stub paragraph added** so deferrals are real.

## References

- v0.10.0 plan: `docs/plans/2026-05-02-feat-v0-10-0-phased-execution-plan.md`
- v0.10.0 release: tag `v0.10.0`.
- v0.11.0 brainstorm: `docs/brainstorms/2026-05-04-v0-11-0-stabilization-brainstorm.md`
- Plan reviews: this PR's review thread.
