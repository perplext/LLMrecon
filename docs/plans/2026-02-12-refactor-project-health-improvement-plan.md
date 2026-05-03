---
title: "refactor: LLMrecon Project Health Improvement Plan"
type: refactor
date: 2026-02-12
---

# LLMrecon Project Health Improvement Plan

## Overview

Comprehensive audit of the LLMrecon project revealing **26 actionable issues** across testing, security, CI/CD, documentation, and code quality. The project has strong architectural foundations and impressive attack module breadth, but critical gaps in testing infrastructure, version consistency, and security hygiene undermine reliability. This plan prioritizes fixes by impact and risk.

## Current Scorecard

| Area | Grade | Key Issue |
|------|-------|-----------|
| **Go Build** | B | Compiles, but 3 packages fail `go vet` |
| **Test Coverage** | F | `*_test.go` in .gitignore; ~1.2% coverage |
| **Python Health** | C | 1 broken file, missing deps, no pytest |
| **CI/CD** | C- | No `go test` in CI, disabled security checks |
| **Documentation** | D+ | Broken links, version chaos, wrong emails |
| **Security Hygiene** | D | Hardcoded admin password, unsafe deserialization, race conditions |
| **Dependencies** | B+ | Mostly current; gym deprecated, sqlite3/CGO conflict |
| **Architecture** | A- | Well-organized 186 packages, clean separation |

---

## Phase 1: Critical Fixes (Blocking)

### 1.1 Remove `*_test.go` from .gitignore
- **File:** `.gitignore:32`
- **Problem:** ALL Go test files are excluded from version control. This is the single most impactful bug.
- **Fix:** Remove the `*_test.go` line, commit existing test files
- **Impact:** Enables test tracking for the entire Go codebase

### 1.2 Fix Go vet format-string errors
- **Files:** `src/bundle/import.go:906`, `src/repository/cache_manager.go:130,194,266,334`, `src/ui/` (73+ instances)
- **Problem:** Non-constant format strings in `fmt.Errorf()` / `fmt.Sprintf()` cause build failures under `go vet`
- **Fix:** Wrap dynamic strings with `"%s"` format specifier
- **Example:**
  ```go
  // Before (broken)
  fmt.Errorf(errMsg)
  // After (fixed)
  fmt.Errorf("%s", errMsg)
  ```

### 1.3 Fix RWMutex pass-by-value race conditions
- **Files:** `src/security/access/api/middleware.go` (8 instances), `src/security/access/api/server.go` (5 instances), `src/automated/chain/exploit_chain_builder.go:1239`
- **Problem:** `sync.RWMutex` copied by value instead of passed by pointer, causing race conditions in the security subsystem
- **Fix:** Change struct fields and constructors to use pointer types

### 1.4 Fix hardcoded admin password
- **File:** `src/security/access/manager.go:629`
- **Problem:** Plaintext `"admin"` as PasswordHash in AccessControlManager initialization
- **Fix:** Use bcrypt or argon2id hashing, load from environment or config

### 1.5 Add `go test` to CI
- **File:** `.github/workflows/ci.yml`
- **Problem:** CI runs `go vet` and `gofmt` but never runs `go test`
- **Fix:** Add `go test ./...` step to CI workflow

### 1.6 Pin Trivy action versions
- **Files:** `.github/workflows/ci.yml:97`, `.github/workflows/release.yml:245`
- **Problem:** `aquasecurity/trivy-action@master` is a supply chain attack vector
- **Fix:** Pin to specific stable tag (e.g., `@v0.30.0`)

---

## Phase 2: High Priority (Correctness & Safety)

### 2.1 Unify version references
- **Problem:** Version claimed differently everywhere:

| Location | Version |
|----------|---------|
| README.md badge | v0.7.1 |
| RELEASE.md | v0.8.0 |
| Makefile VERSION | 0.1.0 |
| SECURITY.md | 1.0.x |
| .github/SECURITY.md | 0.7.x |

- **Fix:** Create single `VERSION` file at root, derive all others. Update README badge, Makefile, both SECURITY.md files.

### 2.2 Fix Dockerfile Go version mismatch
- **File:** `Dockerfile`
- **Problem:** Uses `golang:1.23-alpine` but `go.mod` requires Go 1.24.0
- **Fix:** Update to `golang:1.24-alpine`

### 2.3 Resolve CGO_ENABLED=0 / sqlite3 conflict
- **Problem:** `github.com/mattn/go-sqlite3` requires CGO, but Dockerfile and release workflow use `CGO_ENABLED=0`
- **Options:**
  - (a) Replace `go-sqlite3` with pure-Go `modernc.org/sqlite` (recommended)
  - (b) Enable CGO in Docker/release builds

### 2.4 Fix broken documentation references
- **Missing files referenced in README/CLAUDE.md:**
  - `SECURITY_UPDATE_v0.7.1.md` (linked from README:362, deleted in cleanup)
  - `requirements.txt` at root (README:98 says `pip install -r requirements.txt`)
  - `verify_2025_features.py` (README:202, deleted in cleanup)
  - `demo.sh` and `harness_config.json` (CLAUDE.md references, deleted in cleanup)
- **Fix:** Update README and CLAUDE.md to remove dead links, point to correct paths

### 2.5 Standardize contact information
- **Problem:** Three different security emails and doc URLs:
  - `security@LLMrecon.org` / `security@llmrecon.io` / `security@llmrecon.ai`
  - `docs.llmrecon.io` / `docs.llmrecon.ai` / `llmrecon.com`
- **Fix:** Pick one domain (llmrecon.com) and standardize everywhere

### 2.6 Fix Python llmrecon_indirect.py syntax error
- **File:** `llmrecon_indirect.py:414-418`
- **Problem:** Invalid f-string syntax with ternary operators missing parentheses
- **Fix:** Wrap ternary expressions in parentheses inside f-strings

### 2.7 Enable CodeQL SARIF uploads
- **File:** `.github/workflows/codeql-analysis.yml`
- **Problem:** All 3 CodeQL jobs have `upload: false` so findings never reach GitHub Security tab
- **Fix:** Remove `upload: false` lines

### 2.8 Remove gosec -nosec and -no-fail flags
- **File:** `.github/workflows/go-security.yml:33`
- **Problem:** `-nosec` disables inline overrides, `-no-fail` hides all findings
- **Fix:** Remove both flags, triage existing findings properly

### 2.9 Update OWASP template directory names
- **Problem:** Template directories use 2023 naming (`llm02-insecure-output`, `llm10-model-theft`) but project claims OWASP 2025 compliance
- **Fix:** Rename to match 2025 category names. Add missing templates for LLM05 and LLM08.

---

## Phase 3: Medium Priority (Quality & Maintenance)

### 3.1 Clean up .disabled files
- **Problem:** 21 `.disabled` files in `src/` (e.g., `src/cmd/access_control.go.disabled`)
- **Fix:** Use Go build tags for optional features, or delete if truly unused

### 3.2 Remove duplicate version/ package
- **Problem:** `src/version/consolidated/` is byte-for-byte duplicate of `src/version/` (~1,300 lines)
- **Fix:** Delete `consolidated/` directory, update any imports

### 3.3 Remove dead modules/ directory
- **Problem:** `modules/` at root has 3 small Go files from earlier incarnation of `src/`
- **Fix:** Delete `modules/` directory

### 3.4 Add go.work to .gitignore
- **Problem:** `go.work` and `go.work.sum` committed — these are local dev files
- **Fix:** Add to .gitignore, remove from tracking

### 3.5 Create pytest infrastructure
- **Problem:** Python tests are standalone scripts, not automated test suites
- **Fix:** Add `pytest.ini` or `pyproject.toml`, `conftest.py`, convert existing tests

### 3.6 Replace unsafe deserialization in ML code
- **Files:** `ml/transfer/cross_model_transfer.py:804`, `ml/storage/model_storage.py:228`
- **Problem:** Unsafe deserialization allows arbitrary code execution
- **Fix:** Replace with `torch.load(weights_only=True)` or JSON/safetensors

### 3.7 Migrate gym to gymnasium
- **File:** `ml/requirements.txt`
- **Problem:** OpenAI Gym deprecated in 2022, replaced by Gymnasium
- **Fix:** Change `gym>=0.26.0` to `gymnasium>=0.29.0`, update imports

### 3.8 Create CODEOWNERS file
- **Fix:** Add `.github/CODEOWNERS` with ownership matrix

### 3.9 Replace fake code review workflow
- **File:** `.github/workflows/claude-code-review.yml`
- **Problem:** Always posts "All checks passed" regardless of code quality
- **Fix:** Either remove or replace with meaningful automated checks

### 3.10 Create root requirements.txt
- **Problem:** README references `pip install -r requirements.txt` but file doesn't exist at root
- **Fix:** Create root file or update README to point to correct locations

---

## Phase 4: Nice to Have (Polish)

### 4.1 Add missing CI workflows
- Lint workflow (golangci-lint, flake8, black, mypy)
- Integration test workflow (separate from unit tests)
- Docker image security scan
- Dependabot auto-merge for patch updates

### 4.2 Enhance release workflow
- SBOM generation (syft or in-toto)
- Artifact signing (cosign)
- Provenance attestation (SLSA framework)
- Changelog verification

### 4.3 Add Dependabot Docker scanning
```yaml
- package-ecosystem: "docker"
  directory: "/"
  schedule:
    interval: "weekly"
```

### 4.4 Address code smells
- 168 ignored errors (`_ = ...`) across Go codebase
- 28 `panic()` calls in `src/` — replace with error returns
- 78 TODO/FIXME comments — convert to GitHub issues

### 4.5 Pin Python dependencies
- **Problem:** `ml/requirements.txt` uses `>=` with no upper bounds
- **Fix:** Pin exact versions or use compatible release (`~=`) specifier

---

## What Works Well (Keep)

- **Architecture:** Clean layered design (CLI > API > Business Logic > Repository)
- **Attack breadth:** 45+ modules covering FlipAttack, DrAttack, PAP, MCP, RAG, multi-agent
- **OWASP Agentic 2026:** Forward-looking compliance mapping
- **Go build succeeds:** 241K lines across 186 packages compile cleanly
- **Landing page:** Professional 3-theme site on llmrecon.com
- **Security scanning infra:** Gosec, Trivy, CodeQL, govulncheck, Dependabot all configured
- **Release workflow:** Cross-platform builds with checksums, Docker multi-arch
- **Python ML components:** Well-designed bandit optimizer, data pipeline, DQN agent
- **Template system:** YAML/JSON attack templates with variations and metadata

---

## Implementation Order

```
Phase 1 (Week 1): Critical fixes — unblock testing, fix security, fix CI
Phase 2 (Week 2): High priority — version unification, docs, Python fixes
Phase 3 (Week 3-4): Medium — cleanup, pytest, dependency updates
Phase 4 (Ongoing): Polish — enhanced CI, release hardening, code smells
```

## Acceptance Criteria

- [x] `go test ./...` runs in CI and passes
- [x] `go vet ./...` passes with zero errors
- [x] All version references agree on current version
- [x] No broken links in README, RELEASE.md, or CLAUDE.md
- [x] CodeQL findings upload to GitHub Security tab
- [x] All GitHub Actions pinned to specific versions (no @master/@beta)
- [x] Python `llmrecon_indirect.py` imports without error
- [x] No hardcoded credentials in codebase
- [x] Single contact email domain used everywhere

## References

- PR #82: Repo cleanup (130+ files removed)
- PR #83: Year references + CodeQL v4 upgrade
- PR #84: Pages workflow fix
- OWASP LLM Top 10 2025: https://owasp.org/www-project-top-10-for-large-language-model-applications/
- OWASP Agentic Top 10 2026: https://owasp.org/www-project-agentic-ai-threats/
- Go vet documentation: https://pkg.go.dev/cmd/vet
