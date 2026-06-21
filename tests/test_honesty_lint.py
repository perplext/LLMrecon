"""Tests for scripts/honesty_lint.py (#233)."""
import os
import sys

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "scripts"))

import honesty_lint as hl  # noqa: E402


def test_parse_added_lines_numbers_added_lines():
    diff = (
        "+++ b/src/foo.go\n"
        "@@ -10,1 +10,3 @@\n"
        " context line\n"
        "+added one\n"
        "+added two\n"
    )
    added = hl.parse_added_lines(diff)
    assert added == {"src/foo.go": {11, 12}}


def test_removed_lines_do_not_advance_new_counter():
    diff = (
        "+++ b/src/foo.go\n"
        "@@ -1,2 +1,2 @@\n"
        "-old line\n"
        "+new line\n"
        " unchanged\n"
    )
    added = hl.parse_added_lines(diff)
    assert added == {"src/foo.go": {1}}


def test_flags_added_return_nil_after_todo():
    files = {"src/foo.go": ["func A() error {", "\t// TODO: do it", "\treturn nil", "}"]}
    added = {"src/foo.go": {2, 3}}
    v = hl.find_violations(files, added)
    assert len(v) == 1
    assert v[0][0] == "src/foo.go" and v[0][1] == 3


def test_return_nil_nil_variant_flagged():
    files = {"src/foo.go": ["// FIXME later", "return nil, nil"]}
    assert len(hl.find_violations(files, {"src/foo.go": {1, 2}})) == 1


def test_honesty_allow_suppresses():
    files = {"src/foo.go": ["// TODO", "return nil // honesty:allow reviewed"]}
    assert hl.find_violations(files, {"src/foo.go": {1, 2}}) == []


def test_preexisting_line_not_added_is_ignored():
    files = {"src/foo.go": ["// TODO", "return nil"]}
    # The return-nil line (2) is NOT in the added set -> not a new regression.
    assert hl.find_violations(files, {"src/foo.go": {1}}) == []


def test_todo_too_far_above_is_ignored():
    files = {"src/foo.go": ["// TODO"] + ["x"] * hl.LOOKBACK + ["return nil"]}
    last = len(files["src/foo.go"])
    assert hl.find_violations(files, {"src/foo.go": {last}}) == []


def test_excludes_test_files_and_attic():
    assert hl.is_scanned("src/foo.go") is True
    assert hl.is_scanned("src/foo_test.go") is False
    assert hl.is_scanned("attic/x.go") is False
    assert hl.is_scanned("README.md") is False


def test_bundled_self_test_passes():
    assert hl._self_test() == 0
