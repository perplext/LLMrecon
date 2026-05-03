#!/usr/bin/env bash
# verify-drift.sh — assert derived artifacts match their sources of truth.
#
# Three checks:
#   1. OWASP compliance codegen — `go generate ./...` produces no diff in
#      src/compliance/owasp_agentic_generated.go.
#   2. Go-version pins — every go-version: '1.X' in .github/workflows/*.yml
#      and FROM golang:1.X in Dockerfile matches go.mod's `go 1.X` directive.
#   3. OWASP YAML ↔ code id resolution — every attack_techniques.id and
#      technique_index entry in templates/owasp_agentic_2026.yaml resolves
#      to either a registered module name or a known TechniqueInfo ID.
#      (v0.10.0 #179)
#
# Used by:
#   - `make verify-drift` (local)
#   - .github/workflows/ci.yml verify-drift job (CI; emits ::error annotations)
#
# Set CI=1 (auto-set by GitHub Actions) to emit clickable annotations.
# Otherwise emits plain text diagnostics.

set -euo pipefail

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

emit_error() {
    local file=$1
    local line=$2
    local message=$3
    if [ -n "${CI:-}" ]; then
        echo "::error file=${file},line=${line}::${message}"
    else
        echo "FAIL: ${file}:${line}: ${message}"
    fi
}

# ---------------------------------------------------------------------------
# Check 1 — OWASP compliance codegen up-to-date
# ---------------------------------------------------------------------------

echo "→ OWASP compliance codegen up-to-date"
go generate ./... > /dev/null
if ! git diff --exit-code -- src/compliance/owasp_agentic_generated.go > /dev/null; then
    emit_error \
        "src/compliance/owasp_agentic_generated.go" \
        1 \
        "out of sync with templates/owasp_agentic_2026.yaml. Run 'go generate ./...' locally and commit."
    exit 1
fi
echo "  OK"

# ---------------------------------------------------------------------------
# Check 2 — Go-version pins match go.mod
# ---------------------------------------------------------------------------

echo "→ Go-version pins match go.mod"
EXPECTED=$(awk '/^go [0-9]+\.[0-9]+/ {print $2; exit}' go.mod | cut -d. -f1-2)
if [ -z "$EXPECTED" ]; then
    emit_error "go.mod" 1 "could not extract Go version directive"
    exit 1
fi
echo "  go.mod requires Go ${EXPECTED}"

FAIL=0

# Workflow files: any literal go-version: '1.X' must match. Workflows that
# use go-version-file: 'go.mod' have no literal pin, so they're skipped here
# automatically (the grep returns nothing).
for f in .github/workflows/*.yml; do
    [ -e "$f" ] || continue
    while IFS= read -r line; do
        actual=$(echo "$line" | grep -oE "[0-9]+\.[0-9]+" | head -1)
        if [ -n "$actual" ] && [ "$actual" != "$EXPECTED" ]; then
            line_num=$(grep -n "$line" "$f" 2>/dev/null | head -1 | cut -d: -f1)
            emit_error "$f" "${line_num:-1}" \
                "go-version pin '${actual}' does not match go.mod Go ${EXPECTED}. Use go-version-file: 'go.mod' instead."
            FAIL=1
        fi
    done < <(grep "^[[:space:]]*go-version:[[:space:]]*['\"]\?[0-9]" "$f" || true)
done

# Dockerfile FROM golang:X.Y must match.
if [ -f Dockerfile ] && grep -qE "^FROM golang:[0-9]+\.[0-9]+" Dockerfile; then
    actual=$(grep -oE "golang:[0-9]+\.[0-9]+" Dockerfile | head -1 | cut -d: -f2)
    if [ "$actual" != "$EXPECTED" ]; then
        line_num=$(grep -n "^FROM golang:" Dockerfile | head -1 | cut -d: -f1)
        emit_error "Dockerfile" "${line_num:-1}" \
            "FROM golang:${actual} does not match go.mod Go ${EXPECTED}"
        FAIL=1
    fi
fi

if [ "$FAIL" -eq 0 ]; then
    echo "  All pins match."
fi

# ---------------------------------------------------------------------------
# Check 3 — OWASP YAML id resolution (v0.10.0 #179)
# ---------------------------------------------------------------------------
#
# Delegates to a Go test that imports the src/attacks/all barrel to
# populate attacks.DefaultRegistry, then walks the YAML's
# attack_techniques.id entries asserting each resolves. The test itself
# emits clear failure messages with "did you mean…?" hints.

echo "→ OWASP YAML ↔ code id resolution"
if ! go test ./src/compliance/ -run TestOWASPYAML -count=1 > /tmp/yaml-drift-test.log 2>&1; then
    cat /tmp/yaml-drift-test.log
    if [ -n "${CI:-}" ]; then
        echo "::error file=templates/owasp_agentic_2026.yaml::OWASP YAML id resolution failed; see test output above"
    fi
    FAIL=1
fi
if [ "$FAIL" -eq 0 ]; then
    echo "  All YAML ids resolve."
fi

exit "$FAIL"
