---
title: "v0.11.0 — the stabilization release"
date: 2026-05-04
status: brainstorm
follows: docs/plans/2026-05-02-feat-v0-10-0-phased-execution-plan.md
---

# v0.11.0 brainstorm — the stabilization release

## Working thesis

v0.10.0 was the **honesty release**: stop the binary from lying. v0.11.0
should be the **stabilization release**: make what's there work,
prove it works with tests, and stop carrying dead weight.

Concretely, three pressures converge:

1. **Latent honesty bugs** — `update/verification.go` silently skips
   checksum verification with a TODO. We thought v0.10.0 caught these;
   it didn't. There are ~48 production TODOs and four build-ignored
   `cmd/*.go` files that need triage (fix / delete / document).
2. **Coverage gap** — 20 packages have **zero tests**, including
   `template` (76 files), `bundle` (42), `security` (46), `reporting`
   (28). These are load-bearing subsystems and they have no safety net.
3. **Doc rot** — 102 markdown files in `docs/`, including obsolete
   `V0_2_0_*` / `V0_3_0_*` / `STABILIZATION_PLAN.md` / `REVISED_ROADMAP.md`
   from prior release cycles. Five overlapping quickstart variants.
   New users see all of it.

## Audit findings (from 2026-05-04 sanity-check pass)

### Honesty gaps that survived v0.10.0

- `src/update/verification.go:53,56` — two TODOs around manifest +
  component checksum verification. Code currently *returns success*
  while skipping verification. Same bug class v0.10.0 #174 fixed for
  `update apply` (which now errors honestly); the verification path
  was missed.
- `src/update/customization.go:144` — "Implement actual update logic
  here". If reachable, same problem.
- `src/bundle/delta.go:438,540` — incomplete patch application + tar.gz
  compression for delta bundles. Means the delta-bundle code path
  exists in tree but doesn't work.
- `src/bundle/export.go:755,760` — "Implement key management",
  "Implement encryption". Same shape — code path looks live, isn't.
- `src/bundle/manifest_enhanced.go:325` — "Implement proper YAML
  parsing". Either the simple parsing covers it or it doesn't; needs
  a real test against real bundles.

### Build-ignored `cmd/*.go` files (since 2025-08-26)

- `src/cmd/sign.go` — bundle signing (referenced by #177's deferred
  cosign-integration item)
- `src/cmd/owasp.go` — `owasp test` and `owasp vulnerability` commands;
  this was the command the v0.9.0 README documented as the entry point
- `src/cmd/reporting.go` — report generation (the README docs called
  for this with `--format pdf` etc.)
- `src/cmd/template_security.go` — template security audit

These either need to come back (rewrite against current API like #177
did with bundle round-trip), or move to `attic/` with a tracking note
like #180 did with the RBAC framework. Carrying them as
`//go:build ignore` is the worst option — they show up in `git grep`,
confuse readers, and rot further.

### Coverage gap by package size

| Package | Files | Tests | Risk |
|---|---:|---:|---|
| `src/template/` | 76 | 0 | High — template loader/validator powers attack data |
| `src/testing/` | 60 | 0 | Medium — fixtures, doesn't run in prod |
| `src/security/` | 46 | 0 | High — crypto, vault, prompt protection |
| `src/bundle/` | 42 | 0 | High — air-gapped distribution |
| `src/reporting/` | 28 | 0 | Medium — output formatters |
| `src/utils/` | 19 | 0 | Low — helpers |
| `src/ui/` | 17 | 0 | Low — terminal UI |
| `src/api/` | 17 | 0 | High — REST API surface |
| `src/version/` | 16 | 0 | Medium — version parsing |
| `src/repository/` | 15 | 0 | Medium — abstraction layer |

13 more packages each with <15 files. We can't reasonably bring 20
packages from 0% to thoroughly tested in one release; pick the
load-bearing ones (template, security, bundle, api) and write
enough integration tests to catch regressions.

### Documentation rot

- **15 obsolete versioned docs** to attic: `V0_2_0_*`, `V0_3_0_*`,
  `REVISED_ROADMAP.md`, `STABILIZATION_PLAN.md`, `STRATEGY_ASSESSMENT.md`,
  `EARLY_ADOPTER_*.md`, `DEPLOYMENT_STATUS.md`, `compilation_fixes*.md`,
  `access_control_roadmap.md`.
- **5 overlapping entry-point docs**: `GETTING_STARTED.md`,
  `installation.md`, `quickstart.md`, `QUICK-START-REFERENCE.md`,
  `QUICK_REFERENCE.md`. Pick one, fold the others in.
- **`DOCUMENTATION-INDEX.md` exists** but doesn't reflect the
  current set. Needs a regen.
- **Several `*_summary.md` / `*_implementation.md` doublets** suggest
  doc-as-postmortem patterns that don't age well — squash into
  canonical specs or attic.

### Already-known carry-overs from v0.10.0

- #166 follow-up — `attack run --provider=openai` / `--provider=anthropic`
  CLI plumbing. Bridge promotion exists; CLI side missing.
- #168 Purger interface — automated cleanup of memory-poisoning implants.
- #170 Embedding-fitness opt-in for `jbfuzz`/`persona_evolve`.
- #171 MCTS-Explore selection strategy.
- #174 Tier 3 — binary self-replace + signature verification for
  `update apply`.
- #177 Tier B/C — `bundle create --sign` (cosign), `bundle create
  --source=github`/`--source=gitlab`.
- #198 Node.js 24 actions bump (June 2026 hard cutoff).
- #222 gosec 2.26 cleanup pass + bump (1 real G123 finding + ~14
  annotations).

## Theme candidates

### A — pure stabilization (recommended)

Theme: *make what's already in the tree work, prove it with tests,
attic the rest.* No new attack modules, no new techniques. Every PR
either fixes a TODO, adds tests for an untested package, or removes
code/docs.

- Pros: huge confidence multiplier on every subsequent release.
  Reviewers can trust the tree. CI catches regressions.
- Cons: zero "new" features. Hard to communicate as a release if
  marketing matters.

### B — stabilization + selective new features

Pure stabilization plus the three deferred research items (Purger /
embedding fitness / MCTS) and CLI provider wiring. Headline message:
"now you can actually run attacks against OpenAI/Anthropic."

- Pros: real-provider story closes a gap from v0.10.0; deferred
  research items have been waiting a release cycle.
- Cons: scope creep risk. The deferred items have been "1-week each"
  estimates that historically grow.

### C — full clearout

A + B + completing all 48 TODOs + bringing every package above some
coverage floor (say 30%). 

- Pros: leaves the tree in genuinely shippable state.
- Cons: probably 8–12 weeks of work. Risks v0.11.0 stalling like
  v0.4.0–v0.6.0 did per the legacy roadmap docs.

## Recommendation

**Theme B with hard scope discipline.** Lock the v0.11.0 list now and
defer everything else to v0.12.0 *during planning*, not at the end of
the release.

Specifically:

1. **Honesty cleanup** (1 week) — finish the v0.10.0 unfinished
   business: `update/verification.go` checksum verification (fix or
   error honestly), build-ignored `cmd/*.go` files (re-enable or
   attic), the highest-impact production TODOs.
2. **Test coverage** (2 weeks) — bring `template`, `bundle`, `security`,
   `api` from 0% to "covers the happy path + failure modes". Not 80%
   coverage; "the load-bearing functions have tests."
3. **Doc consolidation** (½ week) — attic 15 obsolete files, fold 5
   quickstarts into one, regen `DOCUMENTATION-INDEX.md`.
4. **CLI provider wiring** (1 week) — `attack run --provider=openai`
   / `--provider=anthropic` end-to-end. Closes the v0.10.0 #166 loop.
5. **Deferred research** (1–2 weeks total) — #168 Purger, #170
   embedding fitness, #171 MCTS, all opt-in.
6. **CI hygiene** (½ week) — #198 Node 24 actions, #222 gosec 2.26.x.

**Total: 5–6 weeks of focused work, ~7–9 weeks with normal interruptions.**

## Honesty invariant

Same as v0.10.0, restated for stabilization:

> No code path returns "not implemented" while printing or returning
> success. No package over 15 files lacks at least one
> integration-style test of its public surface. No top-level doc
> claims a feature the binary doesn't ship.

Three checks at the end of the release. If any fails, hold the tag.

## Single ordering rule

**Honesty cleanup ships first.** A test you write against a function
that's silently lying about success is a test you have to rewrite
when you fix the bug. Fix the bugs, then write the tests.

Corollary: the doc consolidation pass also waits for honesty cleanup,
because some of the docs that need atticking are "this feature is
half-implemented" notes that become unnecessary once the feature is
either real or removed.

## Out of scope

Explicitly NOT in v0.11.0:

- New attack modules. v0.10.0 + v0.9.0 added 50+ across 9 packages.
  The bottleneck is wiring them through to real providers, not
  inventing more.
- New OWASP categories. The 2026 update is already in.
- Auth/RBAC restoration. v0.10.0 #180 deleted the framework with
  intent — bring it back when there's a real consumer.
- Distributed coordination, monitoring dashboards, the production-
  scale infrastructure from v0.2.0. None of it has tests; none of it
  has a documented operator running it. Either revive with intent or
  attic.

## Open questions for planning phase

- Do we ship v0.11.0-rc1 first to surface integration issues with the
  new test suite? (Recommendation: yes for this release, given the
  scope of test additions.)
- Does theme B's CLI provider wiring also need adding `attack run
  --emit-jsonl` consumption to the Python side, or is that already
  covered by v0.9.0 #181? (Believe yes via `python -m ml.data.ingest`,
  worth confirming.)
- For the bundle delta path (`src/bundle/delta.go`), is the partial
  implementation worth completing, or should the whole delta concept
  attic? No operator using it that we know of.
