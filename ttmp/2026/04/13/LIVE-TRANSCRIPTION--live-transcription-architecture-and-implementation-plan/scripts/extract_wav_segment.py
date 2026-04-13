#!/usr/bin/env python3
"""Extract a time range from a PCM WAV file without transcoding.

Usage:
  python3 extract_wav_segment.py \
    --input input.wav \
    --output clip.wav \
    --start 0 \
    --duration 120
"""

from __future__ import annotations

import argparse
import wave
from pathlib import Path


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", required=True, type=Path)
    parser.add_argument("--output", required=True, type=Path)
    parser.add_argument("--start", required=True, type=float)
    parser.add_argument("--duration", required=True, type=float)
    args = parser.parse_args()

    if args.start < 0:
        raise SystemExit("--start must be >= 0")
    if args.duration <= 0:
        raise SystemExit("--duration must be > 0")

    args.output.parent.mkdir(parents=True, exist_ok=True)

    with wave.open(str(args.input), "rb") as src:
        params = src.getparams()
        frame_rate = src.getframerate()
        start_frame = int(args.start * frame_rate)
        frame_count = int(args.duration * frame_rate)
        src.setpos(start_frame)
        frames = src.readframes(frame_count)

        with wave.open(str(args.output), "wb") as dst:
            dst.setparams(params)
            dst.writeframes(frames)

    print(args.output)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
