"""ml.data.ingest — JSONL → SQLite ingest for the v0.10.0 #181 bridge.

Reads JSONL emitted by `llmrecon attack run --emit-jsonl=<path>` and
inserts each entry into the attacks SQLite database via direct SQL
(bypassing the threaded AttackDataPipeline.collect() queue, which is
designed for live data collection rather than batch ingestion).

Run via:
    python -m ml.data.ingest --from-jsonl <path>
    python -m ml.data.ingest --from-jsonl -        # stdin

The ingest is idempotent at the row level: `INSERT OR REPLACE` keys on
attack_id, so re-ingesting the same JSONL file is a no-op (overwrites
identical rows). Useful when retrying a partial ingest.

JSONL schema (the wire format the Go side emits):

    {
        "provider": "openai",
        "model": "gpt-4o-mini",
        "result": {
            "id": "...",
            "timestamp": "2026-05-03T...",
            "technique": "jbfuzz",
            "payload": "...",
            "response": "...",
            "outcome": "skipped",            # v0.9.0 outcome taxonomy
            "skip_reason": "budget_exceeded",
            "skip_detail": "...",
            "success": false,
            "confidence": 0.0,
            "attempt_count": 8,
            "duration": 1234567890,           # nanoseconds
            "tokens_used": 0,
            "cost_usd": 0.0,
            "metadata": {...},
            "success_indicators": [...],
            "parent_run_id": "..."
        }
    }
"""

import argparse
import json
import logging
import sqlite3
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Dict, Iterable

# Reuse the credential-redaction helper + DB schema constants from the
# pipeline module so this ingest stays in sync with the v0.9.0 schema.
from ml.data.attack_data_pipeline import (
    AttackDataPipeline,
    _redact_sensitive_keys,
)

logger = logging.getLogger(__name__)


def _open_jsonl_lines(path: str) -> Iterable[str]:
    """Yield non-empty lines from a JSONL file. Path '-' reads stdin."""
    if path == "-":
        for line in sys.stdin:
            line = line.strip()
            if line:
                yield line
        return

    p = Path(path)
    with p.open("r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if line:
                yield line


def _flatten_entry(entry: Dict[str, Any]) -> Dict[str, Any]:
    """Flatten a Go-emitted JSONL entry into the row shape the SQLite
    schema expects. The Go envelope is `{provider, model, result}`;
    the SQLite columns expect provider + target_model at the top level
    plus the result fields underneath.
    """
    if "result" not in entry:
        raise ValueError("JSONL entry missing 'result' key — not from `attack run --emit-jsonl`")
    result = entry["result"]

    return {
        "attack_id": result.get("id", ""),
        # timezone-aware UTC; .utcnow() is deprecated in Python 3.12+.
        "timestamp": result.get("timestamp", datetime.now(timezone.utc).isoformat()),
        "attack_type": result.get("technique", "unknown"),
        "target_model": entry.get("model", "unknown"),
        "provider": entry.get("provider", "unknown"),
        "payload": result.get("payload", ""),
        "technique_params": result.get("metadata", {}),
        # Status mapping: outcome→status for backward compat with the
        # legacy AttackStatus enum. The v0.9.0 outcome column is the
        # canonical signal.
        "status": _outcome_to_status(result.get("outcome", "")),
        "outcome": result.get("outcome", ""),
        "parent_run_id": result.get("parent_run_id"),
        "response": result.get("response", ""),
        "response_time": _ns_to_seconds(result.get("duration", 0)),
        "tokens_used": result.get("tokens_used", 0),
        "success_indicators": result.get("success_indicators") or [],
        "detection_score": 0.0,
        "semantic_similarity": 0.0,
        "obfuscation_level": result.get("confidence", 0.0),
        "session_id": None,
        "user_id": None,
        "campaign_id": None,
        # Features intentionally empty: extracted live by the pipeline,
        # not reconstructable from JSONL.
        "features": {},
    }


def _outcome_to_status(outcome: str) -> str:
    """Map the v0.9.0 outcome enum to the legacy AttackStatus column.
    Best-effort backward compat — outcome is the canonical signal."""
    return {
        "success": "success",
        "refused": "failed",
        "skipped": "failed",
    }.get(outcome, "failed")


def _ns_to_seconds(ns: Any) -> float:
    """Convert Go's time.Duration (nanoseconds as int) to seconds."""
    try:
        return float(ns) / 1e9
    except (TypeError, ValueError):
        return 0.0


def ingest_jsonl(jsonl_path: str, db_path: Path) -> int:
    """Read JSONL from jsonl_path, write rows into db_path's attacks
    table. Returns the number of rows inserted/replaced.

    Triggers AttackDataPipeline init on the db_path to ensure the v0.9.0
    schema (outcome + parent_run_id columns) is present before insert —
    a fresh DB or a pre-v0.9.0 DB on this path will be migrated
    automatically.
    """
    # Initialize the pipeline against this DB path; this runs the v0.9.0
    # migration if needed. We don't actually use the pipeline for
    # writes — direct SQL is cleaner for batch ingest.
    db_path.parent.mkdir(parents=True, exist_ok=True)
    AttackDataPipeline({"storage_path": str(db_path.parent)})

    conn = sqlite3.connect(db_path)
    inserted = 0
    try:
        for line in _open_jsonl_lines(jsonl_path):
            try:
                entry = json.loads(line)
            except json.JSONDecodeError as e:
                logger.warning("skipping malformed JSONL line: %s", e)
                continue
            try:
                row = _flatten_entry(entry)
            except ValueError as e:
                logger.warning("skipping entry: %s", e)
                continue

            # Apply credential redaction at the ingest boundary, mirroring
            # what _store_attack_data does on the live-collection path.
            safe_params = _redact_sensitive_keys(row["technique_params"])

            conn.execute(
                """INSERT OR REPLACE INTO attacks (
                    attack_id, timestamp, attack_type, target_model, provider,
                    payload, technique_params, obfuscation_level, status,
                    response, response_time, tokens_used, success_indicators,
                    detection_score, semantic_similarity, session_id, user_id,
                    campaign_id, features, outcome, parent_run_id
                ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)""",
                (
                    row["attack_id"],
                    row["timestamp"],
                    row["attack_type"],
                    row["target_model"],
                    row["provider"],
                    row["payload"],
                    json.dumps(safe_params),
                    row["obfuscation_level"],
                    row["status"],
                    row["response"],
                    row["response_time"],
                    row["tokens_used"],
                    json.dumps(row["success_indicators"]),
                    row["detection_score"],
                    row["semantic_similarity"],
                    row["session_id"],
                    row["user_id"],
                    row["campaign_id"],
                    json.dumps(row["features"]),
                    row["outcome"],
                    row["parent_run_id"],
                ),
            )
            inserted += 1
        conn.commit()
    finally:
        conn.close()
    return inserted


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Ingest JSONL emitted by `llmrecon attack run --emit-jsonl` into the attacks SQLite database",
    )
    parser.add_argument(
        "--from-jsonl",
        required=True,
        help="path to JSONL file, or '-' for stdin",
    )
    parser.add_argument(
        "--db-path",
        default="data/attacks/attacks.db",
        help="target SQLite database (default: data/attacks/attacks.db)",
    )
    args = parser.parse_args()

    logging.basicConfig(level=logging.INFO, format="%(message)s")

    n = ingest_jsonl(args.from_jsonl, Path(args.db_path))
    print(f"Ingested {n} row(s) into {args.db_path}")


if __name__ == "__main__":
    main()
