#!/usr/bin/env python3
"""Compare a live replay transcript SQLite database against a reference transcript DB.

Usage:
  python3 compare_transcript_dbs.py \
    --live-db /path/to/live/transcript.db \
    --reference-db /path/to/reference/audio_transcript.db \
    [--summary-json /path/to/live-summary.json]
"""

from __future__ import annotations

import argparse
import json
import sqlite3
from pathlib import Path
from typing import Any


def fetch_db_stats(path: Path) -> dict[str, Any]:
    conn = sqlite3.connect(str(path))
    conn.row_factory = sqlite3.Row
    try:
        cur = conn.cursor()
        word_count = cur.execute("SELECT COUNT(*) FROM words").fetchone()[0]
        filler_count = cur.execute("SELECT COUNT(*) FROM words WHERE is_filler = 1").fetchone()[0]
        chunk_count = cur.execute("SELECT COUNT(*) FROM chunks").fetchone()[0]
        row = cur.execute(
            "SELECT MIN(start_time) AS min_start, MAX(end_time) AS max_end FROM words"
        ).fetchone()
        first_words = [r[0] for r in cur.execute(
            "SELECT word FROM words ORDER BY start_time ASC, id ASC LIMIT 10"
        ).fetchall()]
        last_words = [r[0] for r in cur.execute(
            "SELECT word FROM words ORDER BY end_time DESC, id DESC LIMIT 10"
        ).fetchall()][::-1]
        return {
            "path": str(path),
            "word_count": word_count,
            "filler_count": filler_count,
            "chunk_count": chunk_count,
            "min_start": row["min_start"],
            "max_end": row["max_end"],
            "first_words": first_words,
            "last_words": last_words,
        }
    finally:
        conn.close()


def load_json(path: Path | None) -> dict[str, Any] | None:
    if not path:
        return None
    return json.loads(path.read_text())


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--live-db", required=True, type=Path)
    parser.add_argument("--reference-db", required=True, type=Path)
    parser.add_argument("--summary-json", type=Path)
    args = parser.parse_args()

    live = fetch_db_stats(args.live_db)
    reference = fetch_db_stats(args.reference_db)
    summary = load_json(args.summary_json)

    report = {
        "live": live,
        "reference": reference,
        "delta": {
            "word_count": live["word_count"] - reference["word_count"],
            "chunk_count": live["chunk_count"] - reference["chunk_count"],
            "coverage_seconds": (live["max_end"] or 0) - (reference["max_end"] or 0),
        },
        "live_summary": summary,
    }

    print(json.dumps(report, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
