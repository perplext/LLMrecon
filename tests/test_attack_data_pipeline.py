"""Tests for ml.data.attack_data_pipeline — v0.9.0 migration / redaction /
bandit reward filter.

Run via:
    python -m pytest tests/test_attack_data_pipeline.py -v
or:
    python tests/test_attack_data_pipeline.py    # unittest fallback

No third-party deps beyond the pipeline's own (numpy / pandas, which the
top-level module imports). Tests use sqlite3 directly and tempfile for
isolated databases.
"""

import os
import sqlite3
import sys
import tempfile
import unittest
from pathlib import Path

# Make project root importable when tests/ is invoked directly.
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from ml.data.attack_data_pipeline import (  # noqa: E402
    BANDIT_REWARD_OUTCOMES,
    OUTCOME_REFUSED,
    OUTCOME_SKIPPED,
    OUTCOME_SUCCESS,
    _migrate_v090,
    _redact_sensitive_keys,
)


# ---------------------------------------------------------------------------
# Fixture helpers
# ---------------------------------------------------------------------------


def _create_legacy_db(path: Path, rows):
    """Create a v0.8.0-shaped attacks table at `path` (no outcome /
    parent_run_id columns) and insert `rows` (each a dict with keys
    matching the legacy column names)."""
    conn = sqlite3.connect(path)
    conn.execute(
        """
        CREATE TABLE attacks (
            attack_id TEXT PRIMARY KEY,
            timestamp TEXT NOT NULL,
            attack_type TEXT NOT NULL,
            target_model TEXT NOT NULL,
            provider TEXT NOT NULL,
            payload TEXT NOT NULL,
            technique_params TEXT,
            obfuscation_level REAL,
            status TEXT NOT NULL,
            response TEXT,
            response_time REAL,
            tokens_used INTEGER,
            success_indicators TEXT,
            detection_score REAL,
            semantic_similarity REAL,
            session_id TEXT,
            user_id TEXT,
            campaign_id TEXT,
            features TEXT,
            created_at TEXT DEFAULT CURRENT_TIMESTAMP
        )
        """
    )
    for r in rows:
        conn.execute(
            """
            INSERT INTO attacks (
                attack_id, timestamp, attack_type, target_model, provider,
                payload, status, response_time, tokens_used,
                technique_params, obfuscation_level, response,
                success_indicators, detection_score, semantic_similarity, features
            ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            """,
            (
                r["attack_id"],
                r["timestamp"],
                r["attack_type"],
                r["target_model"],
                r["provider"],
                r["payload"],
                r["status"],
                r.get("response_time", 0.0),
                r.get("tokens_used", 0),
                "{}",
                0.0,
                "",
                "[]",
                0.0,
                0.0,
                "{}",
            ),
        )
    conn.commit()
    conn.close()


# ---------------------------------------------------------------------------
# Credential redaction
# ---------------------------------------------------------------------------


class TestRedactSensitiveKeys(unittest.TestCase):
    def test_top_level_key_match(self):
        out = _redact_sensitive_keys({"api_key": "sk-1234", "model": "gpt-4"})
        self.assertEqual(out["api_key"], "[REDACTED]")
        self.assertEqual(out["model"], "gpt-4")

    def test_case_insensitive(self):
        out = _redact_sensitive_keys({"API_KEY": "x", "Auth_Token": "y", "name": "z"})
        self.assertEqual(out["API_KEY"], "[REDACTED]")
        self.assertEqual(out["Auth_Token"], "[REDACTED]")
        self.assertEqual(out["name"], "z")

    def test_substring_match(self):
        out = _redact_sensitive_keys(
            {
                "secret_phrase": "x",
                "user_password_hash": "y",
                "session_auth_id": "z",
                "ordinary_field": "ok",
            }
        )
        self.assertEqual(out["secret_phrase"], "[REDACTED]")
        self.assertEqual(out["user_password_hash"], "[REDACTED]")
        self.assertEqual(out["session_auth_id"], "[REDACTED]")
        self.assertEqual(out["ordinary_field"], "ok")

    def test_nested_dict(self):
        out = _redact_sensitive_keys(
            {"params": {"auth_token": "x", "endpoint": "https://example.com"}}
        )
        self.assertEqual(out["params"]["auth_token"], "[REDACTED]")
        self.assertEqual(out["params"]["endpoint"], "https://example.com")

    def test_list_of_dicts(self):
        out = _redact_sensitive_keys(
            [{"key": "v1"}, {"name": "v2"}, {"password": "v3"}]
        )
        self.assertEqual(out[0]["key"], "[REDACTED]")
        self.assertEqual(out[1]["name"], "v2")
        self.assertEqual(out[2]["password"], "[REDACTED]")

    def test_passthrough_for_non_dict_non_list(self):
        # Strings, ints, None, etc. pass through unchanged.
        for v in ["secret_string_value", 42, None, 3.14, True]:
            self.assertEqual(_redact_sensitive_keys(v), v)

    def test_does_not_mutate_input(self):
        original = {"api_key": "sk-1234", "model": "gpt-4"}
        snapshot = dict(original)
        _ = _redact_sensitive_keys(original)
        self.assertEqual(original, snapshot, "input dict was mutated")


# ---------------------------------------------------------------------------
# Migration
# ---------------------------------------------------------------------------


class TestMigrateV090(unittest.TestCase):
    def setUp(self):
        self.tmpdir = tempfile.mkdtemp()
        self.db_path = Path(self.tmpdir) / "attacks.db"

    def tearDown(self):
        # Best-effort cleanup; tests may create backup files too.
        for p in Path(self.tmpdir).iterdir():
            try:
                p.unlink()
            except OSError:
                pass
        os.rmdir(self.tmpdir)

    def test_adds_outcome_and_parent_run_id_columns(self):
        _create_legacy_db(self.db_path, [])
        conn = sqlite3.connect(self.db_path)
        try:
            _migrate_v090(conn, self.db_path)
            cols = {r[1] for r in conn.execute("PRAGMA table_info(attacks)").fetchall()}
        finally:
            conn.close()
        self.assertIn("outcome", cols)
        self.assertIn("parent_run_id", cols)

    def test_backfills_outcome_from_status(self):
        rows = [
            {
                "attack_id": "a1",
                "timestamp": "2026-01-01T00:00:00",
                "attack_type": "t",
                "target_model": "m",
                "provider": "p",
                "payload": "x",
                "status": "success",
            },
            {
                "attack_id": "a2",
                "timestamp": "2026-01-01T00:00:00",
                "attack_type": "t",
                "target_model": "m",
                "provider": "p",
                "payload": "x",
                "status": "detected",
            },
            {
                "attack_id": "a3",
                "timestamp": "2026-01-01T00:00:00",
                "attack_type": "t",
                "target_model": "m",
                "provider": "p",
                "payload": "x",
                "status": "blocked",
            },
        ]
        _create_legacy_db(self.db_path, rows)
        conn = sqlite3.connect(self.db_path)
        try:
            _migrate_v090(conn, self.db_path)
            outcomes = dict(conn.execute("SELECT attack_id, outcome FROM attacks").fetchall())
        finally:
            conn.close()
        # status='success' → 'success'; everything else → 'refused'.
        self.assertEqual(outcomes["a1"], OUTCOME_SUCCESS)
        self.assertEqual(outcomes["a2"], OUTCOME_REFUSED)
        self.assertEqual(outcomes["a3"], OUTCOME_REFUSED)

    def test_idempotent_on_second_run(self):
        # Pre-migration row counts and post-migration columns must remain
        # stable across repeated invocations.
        rows = [
            {
                "attack_id": f"a{i}",
                "timestamp": "2026-01-01T00:00:00",
                "attack_type": "t",
                "target_model": "m",
                "provider": "p",
                "payload": "x",
                "status": "success" if i % 2 == 0 else "blocked",
            }
            for i in range(5)
        ]
        _create_legacy_db(self.db_path, rows)
        conn = sqlite3.connect(self.db_path)
        try:
            _migrate_v090(conn, self.db_path)
            first_outcomes = dict(conn.execute("SELECT attack_id, outcome FROM attacks").fetchall())
            first_count = conn.execute("SELECT COUNT(*) FROM attacks").fetchone()[0]
            # Run a second time — must not change anything.
            _migrate_v090(conn, self.db_path)
            second_outcomes = dict(conn.execute("SELECT attack_id, outcome FROM attacks").fetchall())
            second_count = conn.execute("SELECT COUNT(*) FROM attacks").fetchone()[0]
        finally:
            conn.close()
        self.assertEqual(first_outcomes, second_outcomes, "second run changed outcomes")
        self.assertEqual(first_count, second_count, "second run changed row count")

    def test_does_not_overwrite_existing_outcome(self):
        # Rows that already have an outcome value must keep it after a
        # subsequent migration call.
        _create_legacy_db(self.db_path, [
            {
                "attack_id": "a1",
                "timestamp": "2026-01-01T00:00:00",
                "attack_type": "t",
                "target_model": "m",
                "provider": "p",
                "payload": "x",
                "status": "blocked",
            },
        ])
        conn = sqlite3.connect(self.db_path)
        try:
            _migrate_v090(conn, self.db_path)
            # Manually set the outcome to skipped for this row, then
            # re-run migration — which should NOT reset our value.
            conn.execute("UPDATE attacks SET outcome = ? WHERE attack_id = ?",
                         (OUTCOME_SKIPPED, "a1"))
            conn.commit()
            _migrate_v090(conn, self.db_path)
            outcome = conn.execute("SELECT outcome FROM attacks WHERE attack_id = ?",
                                   ("a1",)).fetchone()[0]
        finally:
            conn.close()
        self.assertEqual(outcome, OUTCOME_SKIPPED)

    def test_creates_partial_indexes(self):
        _create_legacy_db(self.db_path, [])
        conn = sqlite3.connect(self.db_path)
        try:
            _migrate_v090(conn, self.db_path)
            indexes = {r[0] for r in conn.execute(
                "SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = 'attacks'"
            ).fetchall()}
        finally:
            conn.close()
        self.assertIn("idx_attacks_outcome", indexes)
        self.assertIn("idx_attacks_parent_run_id", indexes)

    def test_creates_backup_when_altering(self):
        _create_legacy_db(self.db_path, [])
        conn = sqlite3.connect(self.db_path)
        try:
            _migrate_v090(conn, self.db_path)
        finally:
            conn.close()
        # A .bak.{timestamp} sibling must exist.
        backups = [
            p for p in Path(self.tmpdir).iterdir()
            if p.name.startswith("attacks.db.bak.")
        ]
        self.assertEqual(len(backups), 1, f"expected 1 backup; saw {[p.name for p in backups]}")

    def test_no_backup_when_already_migrated(self):
        # Running migration on a DB that already has the columns should
        # NOT create a fresh backup (avoids backup-file proliferation).
        _create_legacy_db(self.db_path, [])
        conn = sqlite3.connect(self.db_path)
        try:
            _migrate_v090(conn, self.db_path)  # First pass — backup created.
        finally:
            conn.close()

        # Delete any backups so we can test the second-run no-backup path
        # in isolation.
        for p in Path(self.tmpdir).iterdir():
            if p.name.startswith("attacks.db.bak."):
                p.unlink()

        conn = sqlite3.connect(self.db_path)
        try:
            _migrate_v090(conn, self.db_path)  # Second pass — no backup expected.
        finally:
            conn.close()

        backups = [
            p for p in Path(self.tmpdir).iterdir()
            if p.name.startswith("attacks.db.bak.")
        ]
        self.assertEqual(
            len(backups), 0,
            "second migration created an unnecessary backup",
        )


# ---------------------------------------------------------------------------
# Bandit reward filter — exercised via direct SQL because instantiating the
# full AttackDataPipeline brings in numpy/pandas/threading that aren't
# necessary for testing the filter contract.
# ---------------------------------------------------------------------------


class TestBanditRewardFilter(unittest.TestCase):
    """The bandit reward filter's central invariant is `outcome IN
    ('success', 'refused')` — skipped is NEVER in the population. These
    tests assert that invariant via the constants themselves (which the
    pipeline method reads from) AND via a direct SQL query that mirrors
    the method's filter."""

    def test_canonical_constants(self):
        # The constants are the contract. Anything that diverges is a bug.
        self.assertEqual(BANDIT_REWARD_OUTCOMES, ("success", "refused"))
        self.assertEqual(OUTCOME_SUCCESS, "success")
        self.assertEqual(OUTCOME_REFUSED, "refused")
        self.assertEqual(OUTCOME_SKIPPED, "skipped")
        # And the central rule: skipped is NOT in the reward set.
        self.assertNotIn(OUTCOME_SKIPPED, BANDIT_REWARD_OUTCOMES)

    def test_filter_excludes_skipped_at_sql_level(self):
        # Build a tiny migrated DB with one of each outcome; verify a
        # filter query returns only success + refused.
        tmpdir = tempfile.mkdtemp()
        db_path = Path(tmpdir) / "attacks.db"
        try:
            rows = [
                {"attack_id": "s1", "status": "success"},
                {"attack_id": "r1", "status": "blocked"},
                {"attack_id": "k1", "status": "success"},  # will become 'skipped'
            ]
            _create_legacy_db(db_path, [
                {**r,
                 "timestamp": "2026-01-01T00:00:00",
                 "attack_type": "t",
                 "target_model": "m",
                 "provider": "p",
                 "payload": "x"}
                for r in rows
            ])
            conn = sqlite3.connect(db_path)
            try:
                _migrate_v090(conn, db_path)
                # Override one row to skipped.
                conn.execute("UPDATE attacks SET outcome = ? WHERE attack_id = ?",
                             (OUTCOME_SKIPPED, "k1"))
                conn.commit()

                # The filter the pipeline uses.
                placeholders = ",".join("?" * len(BANDIT_REWARD_OUTCOMES))
                cursor = conn.execute(
                    f"SELECT attack_id FROM attacks WHERE outcome IN ({placeholders})",
                    BANDIT_REWARD_OUTCOMES,
                )
                ids = sorted(r[0] for r in cursor.fetchall())
            finally:
                conn.close()

            self.assertEqual(ids, ["r1", "s1"], f"skipped row leaked into reward set: {ids}")
        finally:
            for p in Path(tmpdir).iterdir():
                try:
                    p.unlink()
                except OSError:
                    pass
            os.rmdir(tmpdir)


if __name__ == "__main__":
    unittest.main(verbosity=2)
