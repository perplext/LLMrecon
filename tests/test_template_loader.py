"""Tests for template_loader.py — placeholder substitution and rejection.

Run via:
    python -m pytest tests/test_template_loader.py -v
or:
    python tests/test_template_loader.py    # unittest fallback

No third-party deps; uses unittest.
"""

import os
import sys
import unittest

# Make project root importable when tests/ is invoked directly.
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from template_loader import (  # noqa: E402
    UnfilledPlaceholderError,
    find_unfilled_placeholders,
    parse_define,
    parse_defines,
    substitute_placeholders,
    validate_no_unfilled,
)


class TestSubstitution(unittest.TestCase):
    def test_substitutes_simple_string(self):
        self.assertEqual(
            substitute_placeholders("Hello {{NAME}}", {"NAME": "world"}),
            "Hello world",
        )

    def test_leaves_unfilled_placeholders_in_place(self):
        self.assertEqual(
            substitute_placeholders("Hello {{NAME}}", {}),
            "Hello {{NAME}}",
        )

    def test_substitutes_nested_dict(self):
        tpl = {
            "prompt": "Send to {{URL}}",
            "variations": [
                {"prompt": "Also try {{URL}}/alt", "indicators": ["x"]},
            ],
        }
        out = substitute_placeholders(tpl, {"URL": "https://example.com"})
        self.assertEqual(out["prompt"], "Send to https://example.com")
        self.assertEqual(out["variations"][0]["prompt"], "Also try https://example.com/alt")

    def test_does_not_mutate_input(self):
        tpl = {"prompt": "Hi {{X}}"}
        substitute_placeholders(tpl, {"X": "there"})
        self.assertEqual(tpl["prompt"], "Hi {{X}}")

    def test_non_string_leaves_pass_through(self):
        tpl = {"count": 5, "enabled": True, "tags": [1, 2, 3], "label": "{{X}}"}
        out = substitute_placeholders(tpl, {"X": "y"})
        self.assertEqual(out["count"], 5)
        self.assertEqual(out["enabled"], True)
        self.assertEqual(out["tags"], [1, 2, 3])
        self.assertEqual(out["label"], "y")

    def test_recursive_substitution_expands_nested_placeholders(self):
        # If subs[A] contains {{B}}, it should be expanded too.
        out = substitute_placeholders(
            "{{A}}",
            {"A": "alpha-{{B}}", "B": "beta"},
        )
        self.assertEqual(out, "alpha-beta")

    def test_recursion_does_not_loop_forever(self):
        # Cyclic substitutions terminate without raising.
        out = substitute_placeholders(
            "{{A}}",
            {"A": "{{B}}", "B": "{{A}}"},
        )
        # Either {{A}} or {{B}} — point is no infinite loop.
        self.assertIn("{{", out)


class TestPlaceholderDetection(unittest.TestCase):
    def test_finds_placeholders_in_string(self):
        self.assertEqual(
            find_unfilled_placeholders("Hi {{NAME}} from {{ORG}}"),
            ["NAME", "ORG"],
        )

    def test_dedupes_and_sorts(self):
        self.assertEqual(
            find_unfilled_placeholders("{{B}} {{A}} {{B}} {{A}}"),
            ["A", "B"],
        )

    def test_walks_nested_structures(self):
        tpl = {
            "prompt": "Use {{URL}}",
            "variations": [
                {"prompt": "{{HARMFUL_INSTRUCTION}}"},
            ],
        }
        self.assertEqual(
            find_unfilled_placeholders(tpl),
            ["HARMFUL_INSTRUCTION", "URL"],
        )

    def test_empty_when_all_filled(self):
        self.assertEqual(find_unfilled_placeholders("nothing here"), [])

    def test_lowercase_braces_ignored(self):
        # Only [A-Z][A-Z0-9_]* identifiers count; lowercase braces in natural
        # text shouldn't be flagged.
        self.assertEqual(find_unfilled_placeholders("see {{example}} below"), [])

    def test_partial_braces_ignored(self):
        self.assertEqual(find_unfilled_placeholders("{X}}, {{X}, {{x}}"), [])


class TestValidation(unittest.TestCase):
    def test_passes_when_no_placeholders(self):
        validate_no_unfilled({"prompt": "all filled"}, source="t.json")

    def test_raises_with_source_and_keys(self):
        try:
            validate_no_unfilled(
                {"prompt": "Send {{HARMFUL_INSTRUCTION}} to {{EXFIL_URL}}"},
                source="echoleak_chain.json",
            )
            self.fail("expected UnfilledPlaceholderError")
        except UnfilledPlaceholderError as e:
            self.assertEqual(e.source, "echoleak_chain.json")
            self.assertEqual(e.placeholders, ["EXFIL_URL", "HARMFUL_INSTRUCTION"])
            msg = str(e)
            self.assertIn("echoleak_chain.json", msg)
            self.assertIn("{{HARMFUL_INSTRUCTION}}", msg)
            self.assertIn("{{EXFIL_URL}}", msg)
            self.assertIn("-D KEY=VALUE", msg)


class TestParseDefine(unittest.TestCase):
    def test_simple(self):
        self.assertEqual(parse_define("URL=https://example.com"), ("URL", "https://example.com"))

    def test_value_may_contain_equals(self):
        self.assertEqual(parse_define("Q=k=v&z=2"), ("Q", "k=v&z=2"))

    def test_rejects_missing_equals(self):
        with self.assertRaises(ValueError):
            parse_define("BARE")

    def test_rejects_empty_key(self):
        with self.assertRaises(ValueError):
            parse_define("=value")

    def test_rejects_lowercase_key(self):
        with self.assertRaises(ValueError):
            parse_define("url=x")

    def test_rejects_key_starting_with_digit(self):
        with self.assertRaises(ValueError):
            parse_define("1KEY=x")

    def test_rejects_key_with_embedded_braces(self):
        # Regression: re.match would accept this because "{{A}}" is a valid
        # prefix; fullmatch requires the entire wrapped string to match.
        with self.assertRaises(ValueError):
            parse_define("A}}B=x")

    def test_rejects_key_with_trailing_garbage(self):
        with self.assertRaises(ValueError):
            parse_define("KEY!=x")


class TestParseDefines(unittest.TestCase):
    def test_collects_multiple(self):
        self.assertEqual(
            parse_defines(["A=1", "B=2"]),
            {"A": "1", "B": "2"},
        )

    def test_rejects_duplicate_keys(self):
        with self.assertRaises(ValueError):
            parse_defines(["A=1", "A=2"])

    def test_empty(self):
        self.assertEqual(parse_defines([]), {})


if __name__ == "__main__":
    unittest.main()
