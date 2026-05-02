"""Placeholder substitution and validation for v0.9.0 attack templates.

Several v0.9.0 templates (echoleak_chain.json, reverse_captcha.json) ship
structural scaffolds with {{HARMFUL_INSTRUCTION}} / {{EXFIL_URL}} placeholders
rather than literal weaponized payloads. This keeps the repo from acting as a
GitHub-searchable working-payload library for indirect-injection chains.

The functions here:

* substitute_placeholders(obj, subs)  - walk an arbitrarily nested template
  dict/list/str and replace {{KEY}} occurrences with subs[KEY].
* find_unfilled_placeholders(obj)     - return a sorted list of any
  {{...}} occurrences still present after substitution.
* validate_no_unfilled(obj, source)   - raise UnfilledPlaceholderError listing
  what wasn't substituted and where to provide it.

Usage from llmrecon_harness.py:

    from template_loader import substitute_placeholders, validate_no_unfilled
    raw = json.load(template_file)
    substituted = substitute_placeholders(raw, cli_substitutions)
    validate_no_unfilled(substituted, source=template_file.name)
    template = AttackTemplate(substituted)

This module has no third-party deps so it can be vendored or stand alone.
"""

import re
from typing import Any, Dict, List, Mapping, Sequence, Set


# Match {{IDENTIFIER}} where IDENTIFIER is a non-empty sequence of A-Z, 0-9, _.
# Restricted to uppercase + underscore + digits to keep accidental natural-text
# braces (e.g., a JSON example in a description) out of the match.
PLACEHOLDER_PATTERN = re.compile(r"\{\{([A-Z][A-Z0-9_]*)\}\}")


class UnfilledPlaceholderError(ValueError):
    """Raised when a template still contains {{...}} after substitution."""

    def __init__(self, source: str, placeholders: Sequence[str]) -> None:
        self.source = source
        self.placeholders = list(placeholders)
        joined = ", ".join(f"{{{{{p}}}}}" for p in self.placeholders)
        super().__init__(
            f"template {source!r} has unfilled placeholders: {joined}. "
            f"Provide them via the harness CLI -D KEY=VALUE flag, an "
            f"--instruction-file, or substitution dict."
        )


def substitute_placeholders(obj: Any, subs: Mapping[str, str]) -> Any:
    """Return a copy of *obj* with every {{KEY}} replaced by ``subs[KEY]``.

    *obj* may be a string, dict, list, or any combination thereof; non-str
    leaves (numbers, booleans, None) are returned unchanged.

    Substitutions are applied repeatedly per string (not just once), so a
    substitution value that itself contains a placeholder will be expanded.
    A maximum recursion depth of 10 guards against runaway cycles.
    """
    if isinstance(obj, str):
        return _substitute_string(obj, subs)
    if isinstance(obj, Mapping):
        return {k: substitute_placeholders(v, subs) for k, v in obj.items()}
    if isinstance(obj, list):
        return [substitute_placeholders(v, subs) for v in obj]
    return obj


def _substitute_string(s: str, subs: Mapping[str, str]) -> str:
    last = None
    cur = s
    for _ in range(10):
        if last == cur:
            break
        last = cur
        cur = PLACEHOLDER_PATTERN.sub(
            lambda m: subs.get(m.group(1), m.group(0)),
            cur,
        )
    return cur


def find_unfilled_placeholders(obj: Any) -> List[str]:
    """Return a sorted list of unique {{KEY}} identifiers still in *obj*.

    Walks dicts and lists recursively; only inspects string leaves.
    """
    found: Set[str] = set()
    _collect(obj, found)
    return sorted(found)


def _collect(obj: Any, found: Set[str]) -> None:
    if isinstance(obj, str):
        found.update(PLACEHOLDER_PATTERN.findall(obj))
        return
    if isinstance(obj, Mapping):
        for v in obj.values():
            _collect(v, found)
        return
    if isinstance(obj, list):
        for v in obj:
            _collect(v, found)


def validate_no_unfilled(obj: Any, source: str) -> None:
    """Raise UnfilledPlaceholderError if *obj* contains any {{...}}.

    *source* is included in the error message (typically the template
    filename) so operators can locate the offending file quickly.
    """
    placeholders = find_unfilled_placeholders(obj)
    if placeholders:
        raise UnfilledPlaceholderError(source, placeholders)


def parse_define(arg: str) -> tuple:
    """Parse ``KEY=VALUE`` strings as accepted by ``-D`` CLI flags.

    Empty KEY or whitespace-only KEY is rejected. VALUE may contain ``=``;
    only the first ``=`` separates key from value.
    """
    if "=" not in arg:
        raise ValueError(f"-D expects KEY=VALUE, got {arg!r}")
    key, value = arg.split("=", 1)
    key = key.strip()
    if not key:
        raise ValueError(f"-D KEY may not be empty (from {arg!r})")
    if not PLACEHOLDER_PATTERN.fullmatch("{{" + key + "}}"):
        # fullmatch (not match) — match anchors only at the start, so a
        # malformed key like "A}}B" would slip through because "{{A}}" is a
        # valid prefix of "{{A}}B}}".
        raise ValueError(
            f"-D KEY {key!r} must match [A-Z][A-Z0-9_]* "
            f"(uppercase letters, digits, underscores; starting with a letter)"
        )
    return key, value


def parse_defines(args: Sequence[str]) -> Dict[str, str]:
    """Parse a sequence of KEY=VALUE strings into a dict, rejecting duplicates."""
    out: Dict[str, str] = {}
    for arg in args:
        key, value = parse_define(arg)
        if key in out:
            raise ValueError(f"-D {key} specified twice")
        out[key] = value
    return out
