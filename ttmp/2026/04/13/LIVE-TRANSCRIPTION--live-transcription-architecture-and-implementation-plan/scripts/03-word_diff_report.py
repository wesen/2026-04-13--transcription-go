#!/usr/bin/env python3
"""Generate a concrete word-level diff report between two transcript SQLite DBs.

Usage:
  python3 03-word_diff_report.py \
    --batch-db out-batch-clip-000-120/transcript.db \
    --live-db out-live-clip-000-120/transcript.db
"""

from __future__ import annotations

import argparse
import re
import sqlite3
from dataclasses import dataclass
from difflib import SequenceMatcher
from pathlib import Path


@dataclass
class Word:
    raw: str
    norm: str
    start: float
    end: float


def normalize(word: str) -> str:
    return re.sub(r"^[\W_]+|[\W_]+$", "", word.lower())


def load_words(path: Path) -> list[Word]:
    conn = sqlite3.connect(str(path))
    try:
        rows = conn.execute(
            "SELECT word, start_time, end_time FROM words ORDER BY start_time, id"
        ).fetchall()
        return [Word(raw=w, norm=normalize(w), start=s, end=e) for w, s, e in rows]
    finally:
        conn.close()


def render_window(words: list[Word], start: int, end: int) -> str:
    return " ".join(w.raw for w in words[start:end])


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--batch-db", required=True, type=Path)
    parser.add_argument("--live-db", required=True, type=Path)
    args = parser.parse_args()

    batch = load_words(args.batch_db)
    live = load_words(args.live_db)

    sm = SequenceMatcher(
        a=[w.norm for w in batch],
        b=[w.norm for w in live],
        autojunk=False,
    )

    print(f"# Word diff report\n")
    print(f"- batch db: `{args.batch_db}`")
    print(f"- live db: `{args.live_db}`")
    print(f"- batch words: `{len(batch)}`")
    print(f"- live words: `{len(live)}`")
    print(f"- delta: `{len(live) - len(batch)}`\n")

    opcodes = [op for op in sm.get_opcodes() if op[0] != "equal"]
    print(f"- mismatch regions: `{len(opcodes)}`\n")

    for idx, (tag, i1, i2, j1, j2) in enumerate(opcodes, start=1):
        b0, b1 = max(0, i1 - 8), min(len(batch), i2 + 8)
        l0, l1 = max(0, j1 - 8), min(len(live), j2 + 8)
        print(f"## Region {idx}: {tag}")
        print(
            f"- batch range: `{i1}:{i2}` ({i2 - i1} words)"
            f" time~`{batch[i1].start if i1 < len(batch) else batch[-1].end:.2f}s`"
        )
        print(
            f"- live range: `{j1}:{j2}` ({j2 - j1} words)"
            f" time~`{live[j1].start if j1 < len(live) else live[-1].end:.2f}s`"
        )
        print(f"- batch words: `{render_window(batch, b0, b1)}`")
        print(f"- live words: `{render_window(live, l0, l1)}`\n")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
