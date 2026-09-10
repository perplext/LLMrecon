---
title: "chore: repo hygiene — purge committed build artifacts & test output from HEAD and history"
type: chore
date: 2026-09-10
status: proposed
target_branch: chore/repo-hygiene-history-cleanup
---

# chore: Repo Hygiene — Committed Artifacts Cleanup (HEAD + History)

## Summary

A scan of the working tree and full git history found build artifacts, stale
test output, and a hardcoded developer path committed to the repository. **No
live secrets were found** — every API-key / private-key pattern in the tree is
a test fake, a mock-provider string, or an attack-template payload, so **no
credential rotation is required**. This plan removes the offending content from
**both HEAD and all history** and closes the `.gitignore` gaps that let it in.

> History rewriting is destructive and public-repo-facing. It rewrites every
> commit SHA, requires a force-push, and breaks existing clones, forks, and
> open PRs. Execute only with explicit go-ahead and after the coordination
> steps below. This plan documents the procedure; it does not authorize the
> force-push on its own.

---

## Findings

### A. Tracked in HEAD — should not be

1. **~13 compiled Mach-O (arm64) binaries at the repo root, ~100 MB+.**
   `llmrecon` (44 MB), `demo`, `downgrade_example`, `enhanced_prompt_protection`,
   `execution-benchmark`, `hash_example`, `offline-bundle`, `owasp-gen`,
   `owasp-mock-test`, `static_file_handler_example`,
   `static_file_handler_standalone`, `template_security_standalone`,
   `validate-manifest`. These are build outputs of `go build`. `.gitignore`
   covers `compliance-report`, `*.exe`, `*.so`, `*.dylib`, `*.test`, but not
   these platform executables.

2. **`security_reports/` — 10 files** including `all_models_test_results.json`
   (64 KB). These are test-run outputs. `.gitignore` already lists
   `*_results.json` and `*_report.json`, so they contradict the repo's own
   ignore policy (committed before the rule / force-added). The JSON also embeds
   a third party's scraped email address.

3. **Hardcoded developer home path** in
   `examples/testing/test_with_ml_integration.py:192`:
   `/Users/nconsolo/claude-code/llmrecon/data/attacks/`. Leaks a local
   username/path; should be a relative or configurable path.

### B. History-only — already removed from HEAD, still clonable

4. `data/attacks/attacks.db` (SQLite DB), removed in commit
   `2ab6226` ("Remove 130+ unnecessary files"), still in history.
5. `llmrecon_2025.log`, `llmrecon_harness.log`, still in history.

### Not an issue (verified, no action)

- `sk-ant-test-fake-key-not-used` in `attack_test.go`, the `sk-...` / `AKIA...`
  strings in `openai_mock_provider.go`, and the `BEGIN RSA PRIVATE KEY` lines in
  `templates/.../knowledge-extraction.yaml` are all intentional test/mock/
  template content. No real credentials in tree or history.

---

## Plan

### Phase 1 — Stop the bleeding (HEAD only, non-destructive, no history rewrite)

1. `git rm --cached` the 13 root binaries and the `security_reports/` tree
   (keep working-tree copies if desired; only untrack).
2. Extend `.gitignore`:
   - Add the named root binaries (or better, build all binaries into a
     gitignored `bin/` and update `Makefile` / `go build -o` targets).
   - Add `security_reports/`.
3. Fix `examples/testing/test_with_ml_integration.py:192` to a relative path
   (e.g. `./data/attacks/`) or a config value.
4. Commit on a normal branch, open a PR. This alone stops future recommits and
   removes the artifacts from HEAD. Safe, reversible, no coordination needed.

### Phase 2 — History purge (destructive; requires go-ahead + coordination)

Rewrites all history to drop the artifacts from every commit. Recommended tool:
`git filter-repo` (not the deprecated `filter-branch`).

**Pre-flight (required before any rewrite):**

- Confirm with the user; this is a public repo (`perplext/LLMrecon`, ~280
  commits).
- Full backup: `git clone --mirror` the repo to a safe location first.
- Enumerate open PRs and forks — all open PRs must be merged or closed; forks
  will diverge and need manual re-sync. Coordinate with any collaborators.
- Do the rewrite on a fresh `--mirror` clone, verify, then force-push.

**Paths to purge (globs):**

```
# run from a fresh: git clone --mirror git@github.com:perplext/LLMrecon.git
git filter-repo \
  --path data/attacks/attacks.db \
  --path llmrecon_2025.log \
  --path llmrecon_harness.log \
  --path security_reports/ \
  --path demo --path downgrade_example --path enhanced_prompt_protection \
  --path execution-benchmark --path hash_example --path llmrecon \
  --path offline-bundle --path owasp-gen --path owasp-mock-test \
  --path static_file_handler_example --path static_file_handler_standalone \
  --path template_security_standalone --path validate-manifest \
  --invert-paths
```

Notes:
- `--invert-paths` keeps everything EXCEPT the listed paths.
- The root binary names are also valid at their historical paths (they lived at
  repo root throughout); confirm with
  `git log --all --diff-filter=A --name-only -- <path>` per path before running.
- If any binary was ever committed under a different path, add that path too —
  `filter-repo` matches exact paths/globs.

**Post-rewrite:**

- Verify size drop: `git count-objects -vH` before/after; confirm the 44 MB
  `llmrecon` blob and the `.db` are gone (`git rev-list --all --objects | grep`).
- Force-push the mirror: `git push --force --mirror`.
- Have every collaborator re-clone fresh (old clones cannot fast-forward).
- Consider GitHub's cached-view note: rewritten blobs may persist in GitHub's
  cache/PR views for a while; contact GitHub Support if a specific blob must be
  purged from their cache. (Not needed here — no secrets — but relevant if a
  secret is ever found.)

### Phase 3 — Guardrails (prevent recurrence)

- Add a pre-commit hook or CI check that rejects staged binaries (ELF/Mach-O)
  and files matching the existing `_results.json` / `_report.json` /`*.db`
  ignore rules. The repo already has a `ci(honesty)` diff-scoped lint (#233) —
  extend that job.
- Move all `go build -o` targets in the `Makefile` / docs to a gitignored
  `bin/` directory so the default build never lands a binary in a tracked path.

---

## Verification / Acceptance

- Phase 1: `git ls-files` shows no root Mach-O binaries and no `security_reports/`
  files; `.gitignore` covers both; the Python path is relative; PR green.
- Phase 2 (if executed): `git rev-list --all --objects | grep -E 'llmrecon$|attacks.db|security_reports/'` returns nothing; repo size materially reduced;
  force-push complete; collaborators re-cloned.
- Phase 3: CI rejects a test commit that stages a binary or a `*_results.json`.

## Scope Boundaries / Risk

- Phase 1 is safe and can ship immediately.
- **Phase 2 must not run without explicit user go-ahead** and the pre-flight
  coordination — it is irreversible for downstream clones and rewrites public
  history. No secrets are exposed, so there is no urgency forcing the rewrite;
  it is a size/cleanliness improvement, and that trade-off is the user's call.
