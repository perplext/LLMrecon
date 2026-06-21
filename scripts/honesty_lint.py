#!/usr/bin/env python3
"""Honesty lint — catch the next "TODO + return nil" regression (#233).

The v0.10.0/v0.11.0 honesty doctrine forbids a code path that claims success
(`return nil`) while leaving the work unfinished (`// TODO`). This lint flags a
`return nil` added within a few lines after a `TODO`/`FIXME` comment.

It is **diff-scoped**: it only inspects lines *added* in the current change, so
it catches new regressions without requiring the existing tree to be cleaned
first. A reviewed, intentional case can be suppressed with a trailing
`// honesty:allow <reason>` on the `return nil` line.

Usage:
    python3 scripts/honesty_lint.py [--base <ref>]
    python3 scripts/honesty_lint.py --self-test
"""
from __future__ import annotations

import argparse
import os
import re
import subprocess
import sys

LOOKBACK = 5  # how many lines above the `return nil` to scan for a TODO
RETURN_NIL = re.compile(r"^return nil(, nil)?$")
TODO = re.compile(r"//.*\b(TODO|FIXME)\b", re.IGNORECASE)
ALLOW = "honesty:allow"

EXCLUDE_PREFIXES = ("attic/",)
EXCLUDE_SUFFIXES = ("_test.go",)


def is_scanned(path: str) -> bool:
    if not path.endswith(".go"):
        return False
    if any(path.startswith(p) for p in EXCLUDE_PREFIXES):
        return False
    if any(path.endswith(s) for s in EXCLUDE_SUFFIXES):
        return False
    return True


def parse_added_lines(diff_text: str) -> dict[str, set[int]]:
    """Parse `git diff` output into {path: {new_line_numbers_added}}."""
    added: dict[str, set[int]] = {}
    path: str | None = None
    new_lineno = 0
    for line in diff_text.splitlines():
        if line.startswith("+++ b/"):
            path = line[6:]
            continue
        if line.startswith("@@"):
            # @@ -a,b +c,d @@  -> next added line is numbered c
            m = re.search(r"\+(\d+)", line.split("@@")[1])
            new_lineno = int(m.group(1)) if m else 0
            continue
        if path is None:
            continue
        if line.startswith("+") and not line.startswith("+++"):
            added.setdefault(path, set()).add(new_lineno)
            new_lineno += 1
        elif line.startswith("-") and not line.startswith("---"):
            pass  # removed line — does not advance the new-file counter
        else:
            new_lineno += 1  # context line
    return added


def find_violations(
    files: dict[str, list[str]], added: dict[str, set[int]]
) -> list[tuple[str, int, str]]:
    """Return [(path, lineno, reason)] for added `return nil` near a TODO."""
    violations: list[tuple[str, int, str]] = []
    for path, lines in files.items():
        if not is_scanned(path):
            continue
        added_here = added.get(path, set())
        for i, raw in enumerate(lines):
            lineno = i + 1
            if lineno not in added_here:
                continue
            stripped = raw.strip()
            if not RETURN_NIL.match(stripped):
                continue
            if ALLOW in raw:
                continue
            # Look back for a TODO/FIXME within LOOKBACK lines.
            start = max(0, i - LOOKBACK)
            window = lines[start:i]
            if any(TODO.search(w) for w in window):
                todo_rel = next(
                    (start + j + 1 for j, w in enumerate(window) if TODO.search(w)), lineno
                )
                violations.append(
                    (path, lineno, f"`return nil` claims success but a TODO/FIXME sits at line {todo_rel}")
                )
    return violations


def _git(args: list[str]) -> str:
    return subprocess.run(
        ["git", *args], capture_output=True, text=True, check=True
    ).stdout


def run(base: str) -> int:
    merge_base = _git(["merge-base", base, "HEAD"]).strip() or base
    diff = _git(["diff", f"{merge_base}...HEAD"])
    added = parse_added_lines(diff)
    files: dict[str, list[str]] = {}
    for path in added:
        if is_scanned(path) and os.path.exists(path):
            with open(path, encoding="utf-8") as fh:
                files[path] = fh.read().splitlines()
    violations = find_violations(files, added)
    if not violations:
        print("honesty-lint: clean — no new TODO+return-nil regressions.")
        return 0
    print("honesty-lint: found honesty regressions in this change:\n")
    for path, lineno, reason in violations:
        print(f"  {path}:{lineno}: {reason}")
    print(
        "\nEither finish the work, return a typed error, or — if this is a "
        "reviewed exception — add `// honesty:allow <reason>` to the return line."
    )
    return 1


def _self_test() -> int:
    diff = (
        "+++ b/src/foo.go\n"
        "@@ -1,2 +1,6 @@\n"
        " func A() error {\n"
        "+\t// TODO: implement A\n"
        "+\treturn nil\n"
        "+}\n"
        "+func B() error { return errors.New(\"x\") }\n"
    )
    added = parse_added_lines(diff)
    assert added["src/foo.go"] == {2, 3, 4, 5}, added
    files = {
        "src/foo.go": [
            "func A() error {",
            "\t// TODO: implement A",
            "\treturn nil",
            "}",
            'func B() error { return errors.New("x") }',
        ]
    }
    v = find_violations(files, added)
    assert len(v) == 1 and v[0][1] == 3, v

    # Suppression marker silences it.
    files["src/foo.go"][2] = "\treturn nil // honesty:allow stub kept intentionally"
    assert find_violations(files, added) == [], "honesty:allow should suppress"

    # Pre-existing (not-added) return nil is ignored.
    assert find_violations(files, {"src/foo.go": {2}}) == [], "non-added line ignored"

    # Test files and attic are excluded.
    assert find_violations({"x_test.go": ["//TODO", "return nil"]}, {"x_test.go": {1, 2}}) == []
    assert find_violations({"attic/x.go": ["//TODO", "return nil"]}, {"attic/x.go": {1, 2}}) == []

    print("honesty_lint self-test: PASS")
    return 0


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--base", default=os.environ.get("HONESTY_BASE", "origin/main"))
    ap.add_argument("--self-test", action="store_true")
    args = ap.parse_args()
    if args.self_test:
        return _self_test()
    return run(args.base)


if __name__ == "__main__":
    sys.exit(main())
