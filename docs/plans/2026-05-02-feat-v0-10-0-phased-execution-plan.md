---
title: "feat: v0.10.0 — the honesty release"
type: feat
date: 2026-05-02
issues_in_scope: [166, 167, 169, 173, 174, 175, 176, 177, 178, 179, 180, 181]
issues_deferred: [168, 170, 171]
issues_partial: {174: "Tiers 1+2 in scope; Tier 3 (binary self-replace + sigverify) deferred to v0.11.0"}
reviewers: [dhh, kieran, simplicity]
---

# v0.10.0 — The Honesty Release

## Thesis

Stop the binary from lying about what it does. Currently:

- `attacks.DefaultRegistry` is empty at runtime — every attack module since v0.7.0 is build-only (#173).
- `update apply` prints success while writing nothing to disk (#174).
- 12 agentic/audio modules silently text-simulate their advertised modalities (#176).
- README's first quick-start command doesn't exist (#175).
- Python and Go halves of the codebase share branding only (#181).

v0.10.0 fixes those, plus the smaller honesty gaps surfaced by the post-release review.

## Single ordering rule

**#173 must merge before #166 / #176 / #179 / #181 / Phase-end README pass.** Everything else is independent.

Corollary: **#169 (drift-detection CI) must merge before any non-#173 PR**, so subsequent PRs are gated by the drift checks. The v0.9.0 Go-pin drift escaped because no such check existed.

## Scope

Effort estimates are honest after Kieran's review caught two systematic underestimates.

| # | Item | Effort | Notes |
|---|---|---|---|
| #169 | Drift-detection CI (compliance + Go pins) | 1 day | First, gates everything below |
| #173 | Attack registry CLI: barrel + `attack list` + `attack run --provider=mock` | 3 days | Highest leverage; unblocks 5 others |
| #167 | `common.Provider` shim around `core.Provider` | 1 day | Independent of #173 |
| #174 Tier 1 | Stop fake-success in `update apply` (return errors, exit non-zero) | ½ day | Pure honesty fix |
| #180 | Delete dead RBAC framework + 4 disabled cmd files; update CLAUDE.md | 1 day | "Wire" path was misestimated at 5 days; it's 10–14. Not in scope. |
| #178 | Close as folded into #173 (`attack run --provider=openai` is the new path) | — | No code |
| #166 | OpenAI + Anthropic adapters: `ImageProvider`, `ReasoningProvider` | **7–9 days** | Anthropic signed-trace short-circuit non-trivial; CHANGELOG callout for `SkipSignatureGated` against Claude |
| #176 | Declare `MCPProvider` / `BrowserProvider` / `AudioProvider`; gate 12 modules | 3 days | Mirror v0.9.0 `ImageProvider` pattern |
| #179 | OWASP YAML ↔ code drift CI test (folds into #169 infra) | 1 day | Imports `src/attacks/all` barrel from #173 |
| #181 | Python JSONL bridge (`attack run --emit-jsonl` + `python -m ml.data.ingest`) | 3 days | Realizes the v0.9.0 schema migration's intent |
| #174 Tier 2 | ZIP/TAR extract + atomic replace for templates + modules + backup | 5 days | Gate behind `--experimental` for one release cycle |
| #177 | Bundle: re-enable `import`/`verify`; cosign signing; GitHub/GitLab fetch | 5 days | Closes air-gapped trust loop |
| #175 | Single README pass to canonical `attack run` post-#173 | ½ day | Done at end, once |

**Total**: 31–33 working days of focused work (~6 weeks at full focus, 8–10 weeks with normal interruptions).

## Deferred to v0.11.0

| # | Why deferred |
|---|---|
| #168 Purger interface | No reported operator pain; depends on #173 + new CLI subcommand. Useful but not load-bearing for the honesty release. |
| #170 Embedding fitness | Opt-in feature; current heuristic fitness is the documented default. No operator pulling on it. |
| #171 MCTS-Explore | Opt-in selection algorithm; default UCB1+restart unchanged. Pure engine R&D. |
| #174 Tier 3 | Binary self-replace + signature verification — separate planning cycle warranted. |

## Honesty invariant (non-negotiable)

**No code path returns "not implemented" while printing or returning success.** Every existing stub either:

1. Returns a non-zero error and exits the CLI with a non-zero status, OR
2. Is implemented for real, OR
3. Is deleted along with the docs/help that advertised it.

This is the single acceptance criterion that, if violated, blocks the release.

## Per-item acceptance criteria (the ones that matter)

Most issues have testable criteria in the issue itself. The five that need extra discipline:

- **#173**: `attack list` enumerates ≥40 modules; `attack run --module=jbfuzz --provider=mock --metadata allow_experimental=true` returns a typed `AttackResult` with non-empty `Outcome`.
- **#174 Tier 2**: kill `-9` mid-apply leaves the operator with EITHER the old templates dir OR the new one — never half. Backup file (`{path}.bak.{ts}`) survives kill-9.
- **#176**: every one of the 12 affected modules emits `OutcomeSkipped + SkipMissingCapability` against today's text-only providers, OR sets `Metadata["mode"]="text_simulation"` when the operator opts in.
- **#181**: round-trip test `Go attack run → JSONL → Python ingest → SQLite read` asserts field-by-field equality; credential redaction at the Python boundary scrubs `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`, generic bearer tokens, and any value matching the v0.9.0 `_SENSITIVE_KEY_PATTERN`.
- **Barrel test (one-liner enabling everything)**: `TestNoNameCollisions` imports `src/attacks/all` and asserts `Registry.Register` panicked zero times. Cheap protection against duplicate-name footguns at binary startup.

## Risks worth tracking

| Risk | Mitigation |
|---|---|
| #174 Tier 2 atomic-replace bricks an installation | Backup before apply (mirror v0.9.0 WAL-safe pattern); gate behind `--experimental` for one release cycle |
| Phase 2 high-velocity PRs merge before #169 lands | Hard rule: no non-#173 PR merges until #169 is on `main` |
| #166 effort blows past 9 days | Descope to OpenAI-only for v0.10.0; ship Anthropic in v0.10.1 |
| #181 silently corrupts attack data on the wire | Round-trip equality test is part of the acceptance criterion |

## What this plan is not

- Not a Gantt chart or staffing model — single-author repo.
- Not an ADR factory — the four "decisions" the prior plan called out are now just choices made in this document.
- Not a stream/phase taxonomy — the only ordering rule is the one above.

## References

- v0.9.0 release: tag `v0.9.0`, commit `c2ebd47`, GitHub release page.
- v0.9.0 Phase 5 ML migration the bridge in #181 was designed for: PR #163.
- The original 480-line plan that this slim form replaces lives in `git log` for anyone who wants the long form.
- Reviewers: DHH (taste/scope), Kieran (correctness/effort), Simplicity (minimization). Their convergence on deferring #168/#170/#171 and dropping ADR ceremony is what shaped this rewrite.
