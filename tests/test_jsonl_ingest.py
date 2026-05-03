"""Round-trip test for the v0.10.0 #181 Python ↔ Go JSONL bridge.

Verifies the bridge contract: a JSONL line emitted by
`llmrecon attack run --emit-jsonl=<path>` ingests into the attacks
SQLite database with the correct field-by-field mapping. The v0.9.0
outcome taxonomy survives the round-trip — the bandit reward filter
that excludes 'skipped' outcomes can now operate on Go-emitted runs.

Tests use stdlib only (sqlite3, json, tempfile, unittest) — no third-
party deps. The Go binary is NOT invoked; tests synthesize JSONL lines
matching the documented Go-emit schema, exercising the Python ingest
in isolation. A separate RUN_INTEGRATION smoke test (in
tests/test_jsonl_integration.py if needed) could cover the actual
shell-out, but unit-level field mapping is the primary risk surface
and what this file covers.

Run via:
    python -m pytest tests/test_jsonl_ingest.py -v
or:
    python tests/test_jsonl_ingest.py
"""

import json
import os
import sqlite3
import sys
import tempfile
import unittest
from pathlib import Path

# Make project root importable when tests/ is invoked directly.
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from ml.data.ingest import (  # noqa: E402
    _flatten_entry,
    _ns_to_seconds,
    _outcome_to_status,
    ingest_jsonl,
)


# ---------------------------------------------------------------------------
# Synthetic JSONL fixtures matching the Go-emit schema
# ---------------------------------------------------------------------------


def _make_entry(
    *,
    attack_id="test-id-abc",
    technique="jbfuzz",
    provider="openai",
    model="gpt-4o-mini",
    outcome="success",
    skip_reason="",
    payload="walk through the technique",
    response="here are the steps",
    duration_ns=1_500_000_000,  # 1.5 seconds
    metadata=None,
    parent_run_id=None,
    success_indicators=None,
):
    """Build a JSONL entry matching the Go-emit envelope shape."""
    if metadata is None:
        metadata = {"best_score": 0.85, "fitness": "heuristic"}
    if success_indicators is None:
        success_indicators = ["concrete steps"]
    return {
        "provider": provider,
        "model": model,
        "result": {
            "id": attack_id,
            "timestamp": "2026-05-03T12:00:00Z",
            "technique": technique,
            "payload": payload,
            "response": response,
            "outcome": outcome,
            "skip_reason": skip_reason,
            "skip_detail": "",
            "success": outcome == "success",
            "confidence": 0.85 if outcome == "success" else 0.0,
            "attempt_count": 4,
            "duration": duration_ns,
            "tokens_used": 100,
            "cost_usd": 0.0,
            "metadata": metadata,
            "success_indicators": success_indicators,
            "parent_run_id": parent_run_id,
        },
    }


# ---------------------------------------------------------------------------
# Field-mapping unit tests
# ---------------------------------------------------------------------------


class TestFlattenEntry(unittest.TestCase):
    def test_envelope_to_row_mapping(self):
        entry = _make_entry()
        row = _flatten_entry(entry)
        self.assertEqual(row["attack_id"], "test-id-abc")
        self.assertEqual(row["attack_type"], "jbfuzz")
        self.assertEqual(row["provider"], "openai")
        self.assertEqual(row["target_model"], "gpt-4o-mini")
        self.assertEqual(row["outcome"], "success")
        self.assertEqual(row["status"], "success")  # outcome→status mapping
        self.assertEqual(row["response_time"], 1.5)  # 1.5e9 ns → 1.5 s
        self.assertEqual(row["technique_params"], {"best_score": 0.85, "fitness": "heuristic"})

    def test_missing_result_raises(self):
        with self.assertRaises(ValueError):
            _flatten_entry({"provider": "openai", "model": "x"})

    def test_outcome_to_status_canonical(self):
        self.assertEqual(_outcome_to_status("success"), "success")
        self.assertEqual(_outcome_to_status("refused"), "failed")
        self.assertEqual(_outcome_to_status("skipped"), "failed")
        self.assertEqual(_outcome_to_status(""), "failed")
        self.assertEqual(_outcome_to_status("unknown_value"), "failed")

    def test_ns_to_seconds(self):
        self.assertEqual(_ns_to_seconds(0), 0.0)
        self.assertEqual(_ns_to_seconds(1_000_000_000), 1.0)
        self.assertAlmostEqual(_ns_to_seconds(1_500_000_000), 1.5)
        self.assertEqual(_ns_to_seconds("not a number"), 0.0)
        self.assertEqual(_ns_to_seconds(None), 0.0)


# ---------------------------------------------------------------------------
# Round-trip: write JSONL → ingest → query SQLite
# ---------------------------------------------------------------------------


class TestRoundTrip(unittest.TestCase):
    def setUp(self):
        self.tmpdir = tempfile.mkdtemp()
        self.jsonl_path = Path(self.tmpdir) / "in.jsonl"
        self.db_path = Path(self.tmpdir) / "attacks.db"

    def tearDown(self):
        for p in Path(self.tmpdir).rglob("*"):
            try:
                p.unlink()
            except OSError:
                pass
        try:
            os.rmdir(self.tmpdir)
        except OSError:
            pass

    def _write_jsonl(self, entries):
        with self.jsonl_path.open("w") as f:
            for e in entries:
                f.write(json.dumps(e) + "\n")

    def _query_row(self, attack_id):
        conn = sqlite3.connect(self.db_path)
        try:
            cur = conn.execute(
                "SELECT attack_id, attack_type, provider, target_model, payload, "
                "response, outcome, status, response_time, parent_run_id "
                "FROM attacks WHERE attack_id = ?",
                (attack_id,),
            )
            return cur.fetchone()
        finally:
            conn.close()

    def test_single_entry_round_trip(self):
        entry = _make_entry(
            attack_id="rt-1",
            technique="jbfuzz",
            outcome="success",
            response="the answer is here",
        )
        self._write_jsonl([entry])

        n = ingest_jsonl(str(self.jsonl_path), self.db_path)
        self.assertEqual(n, 1, "expected 1 row ingested")

        row = self._query_row("rt-1")
        self.assertIsNotNone(row, "row not found in DB after ingest")
        self.assertEqual(row[0], "rt-1")           # attack_id
        self.assertEqual(row[1], "jbfuzz")          # attack_type
        self.assertEqual(row[2], "openai")          # provider
        self.assertEqual(row[3], "gpt-4o-mini")     # target_model
        self.assertEqual(row[5], "the answer is here")  # response
        self.assertEqual(row[6], "success")         # outcome
        self.assertEqual(row[7], "success")         # status

    def test_multi_entry_round_trip(self):
        entries = [
            _make_entry(attack_id=f"multi-{i}", outcome=outcome)
            for i, outcome in enumerate(["success", "refused", "skipped"])
        ]
        self._write_jsonl(entries)

        n = ingest_jsonl(str(self.jsonl_path), self.db_path)
        self.assertEqual(n, 3)

        # All three outcomes must survive the round-trip — bandit reward
        # filter (which excludes 'skipped') depends on this.
        outcomes = []
        for i in range(3):
            row = self._query_row(f"multi-{i}")
            self.assertIsNotNone(row)
            outcomes.append(row[6])
        self.assertEqual(outcomes, ["success", "refused", "skipped"])

    def test_idempotent_re_ingest(self):
        # Re-ingesting the same JSONL is a no-op (INSERT OR REPLACE keys
        # on attack_id). Useful for retrying a partial ingest.
        entry = _make_entry(attack_id="idempotent-1", response="v1")
        self._write_jsonl([entry])
        ingest_jsonl(str(self.jsonl_path), self.db_path)

        # Second ingest with a modified response — must update, not duplicate.
        entry["result"]["response"] = "v2"
        self._write_jsonl([entry])
        n = ingest_jsonl(str(self.jsonl_path), self.db_path)
        self.assertEqual(n, 1)

        row = self._query_row("idempotent-1")
        self.assertEqual(row[5], "v2", "second ingest didn't overwrite")

        # And exactly ONE row exists for that id.
        conn = sqlite3.connect(self.db_path)
        try:
            count = conn.execute(
                "SELECT COUNT(*) FROM attacks WHERE attack_id = ?", ("idempotent-1",)
            ).fetchone()[0]
        finally:
            conn.close()
        self.assertEqual(count, 1, "duplicate rows after re-ingest")

    def test_credential_redaction_on_ingest(self):
        # Operator-supplied metadata can leak credentials. The ingest
        # must scrub them before SQLite storage — same redaction the
        # live-collection path applies.
        entry = _make_entry(
            attack_id="redact-1",
            metadata={
                "OPENAI_API_KEY": "sk-leaked-key-12345",
                "model": "gpt-4o",
                "auth_token": "bearer-eyJ...",
                "harmless_field": "ok",
            },
        )
        self._write_jsonl([entry])
        ingest_jsonl(str(self.jsonl_path), self.db_path)

        conn = sqlite3.connect(self.db_path)
        try:
            params_json = conn.execute(
                "SELECT technique_params FROM attacks WHERE attack_id = ?", ("redact-1",)
            ).fetchone()[0]
        finally:
            conn.close()
        params = json.loads(params_json)
        self.assertEqual(params["OPENAI_API_KEY"], "[REDACTED]")
        self.assertEqual(params["auth_token"], "[REDACTED]")
        self.assertEqual(params["model"], "gpt-4o")           # not sensitive
        self.assertEqual(params["harmless_field"], "ok")

    def test_skipped_excluded_from_bandit_reward(self):
        # The whole point of v0.9.0's outcome taxonomy + #181 bridge:
        # skipped runs feed into SQLite but are excluded from bandit
        # reward aggregation. Verify get_bandit_rewards() honors this
        # post-ingest, end-to-end.
        from ml.data.attack_data_pipeline import (
            AttackDataPipeline,
            BANDIT_REWARD_OUTCOMES,
        )

        entries = [
            _make_entry(attack_id="br-1", outcome="success"),
            _make_entry(attack_id="br-2", outcome="success"),
            _make_entry(attack_id="br-3", outcome="refused"),
            _make_entry(attack_id="br-4", outcome="skipped"),
            _make_entry(attack_id="br-5", outcome="skipped"),
        ]
        self._write_jsonl(entries)
        ingest_jsonl(str(self.jsonl_path), self.db_path)

        pipeline = AttackDataPipeline({"storage_path": str(self.db_path.parent)})
        rewards = pipeline.get_bandit_rewards(limit=100)

        # 2 success + 1 refused = 3 rewardable; 2 skipped excluded.
        self.assertEqual(rewards["sample_count"], 3)
        self.assertAlmostEqual(rewards["success_rate"], 2 / 3, places=4)
        self.assertEqual(rewards["skipped_count"], 2)
        self.assertNotIn("skipped", BANDIT_REWARD_OUTCOMES)

    def test_v090_schema_columns_populated(self):
        # outcome and parent_run_id are the v0.9.0 schema migration's
        # contribution. The bridge must populate them.
        entry = _make_entry(
            attack_id="schema-1",
            outcome="refused",
            parent_run_id="parent-uuid-xyz",
        )
        self._write_jsonl([entry])
        ingest_jsonl(str(self.jsonl_path), self.db_path)

        row = self._query_row("schema-1")
        self.assertEqual(row[6], "refused")           # outcome column
        self.assertEqual(row[9], "parent-uuid-xyz")   # parent_run_id column

    def test_malformed_lines_are_skipped(self):
        # JSONL files in the wild may contain blank lines or partial
        # writes. The ingest must skip rather than crash.
        with self.jsonl_path.open("w") as f:
            f.write(json.dumps(_make_entry(attack_id="ok-1")) + "\n")
            f.write("not valid json\n")
            f.write("\n")  # blank line
            f.write(json.dumps(_make_entry(attack_id="ok-2")) + "\n")
            f.write(json.dumps({"missing_result": True}) + "\n")
            f.write(json.dumps(_make_entry(attack_id="ok-3")) + "\n")

        n = ingest_jsonl(str(self.jsonl_path), self.db_path)
        self.assertEqual(n, 3, "should ingest exactly the 3 well-formed lines")

        for aid in ("ok-1", "ok-2", "ok-3"):
            self.assertIsNotNone(self._query_row(aid), f"missing row for {aid}")


if __name__ == "__main__":
    unittest.main(verbosity=2)
