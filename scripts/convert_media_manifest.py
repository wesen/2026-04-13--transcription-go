#!/usr/bin/env python3
"""Convert the Southwell media-manifest.json into a normalized transcription manifest."""
from __future__ import annotations

import json
import sys
from pathlib import Path

MEMBERS_ONLY_ID = "y8WC6ngApmU"


def main() -> None:
    media_dir = Path(sys.argv[1]) if len(sys.argv) > 1 else Path.home() / "Movies/richard-southwell-category-theory-for-beginners"
    manifest_path = media_dir / "media-manifest.json"
    output_path = media_dir / "transcription-manifest.json"

    raw = json.loads(manifest_path.read_text())
    items = []
    for entry in raw.get("items", []):
        info = {}
        meta_path = Path(entry["path"]).with_suffix(".info.json")
        if meta_path.exists():
            info = json.loads(meta_path.read_text())
        source_id = entry.get("youtube_id") or info.get("id", "")
        if not source_id:
            continue
        position = entry.get("playlist_index") or info.get("playlist_index", 0)
        title = entry.get("title") or info.get("title", source_id)
        source_url = info.get("webpage_url") or f"https://www.youtube.com/watch?v={source_id}"
        audio_path = str(media_dir / "audio" / f"{Path(entry['path']).stem}.wav")
        availability = "available"
        if source_id == MEMBERS_ONLY_ID:
            availability = "members_only"
            audio_path = ""
        elif not Path(audio_path).exists():
            availability = "missing"
            audio_path = ""
        duration = 0.0
        for stream in entry.get("probe", {}).get("streams", []):
            if stream.get("codec_type") == "audio":
                duration = float(stream.get("duration", 0))
                break
        if duration == 0:
            duration = float(entry.get("probe", {}).get("format", {}).get("duration", 0))

        items.append({
            "source_id": source_id,
            "position": position,
            "title": title,
            "source_url": source_url,
            "media_path": entry["path"],
            "audio_path": audio_path,
            "media_sha256": entry.get("sha256", ""),
            "audio_sha256": "",
            "duration_seconds": round(duration, 3),
            "availability": availability,
            "metadata": {} if availability == "available" else {"reason": availability},
        })

    # Ensure the members-only item is present even if not in media-manifest.
    existing_ids = {i["source_id"] for i in items}
    if MEMBERS_ONLY_ID not in existing_ids:
        items.append({
            "source_id": MEMBERS_ONLY_ID,
            "position": 35,
            "title": "Category Theory For Beginners (members-only)",
            "source_url": f"https://www.youtube.com/watch?v={MEMBERS_ONLY_ID}",
            "availability": "members_only",
            "duration_seconds": 0,
            "metadata": {"reason": "channel membership required"},
        })

    items.sort(key=lambda x: x["position"])

    manifest = {
        "schema": "transcription-video-corpus/v1",
        "corpus": {
            "id": "richard-southwell-category-theory-for-beginners",
            "title": "Category Theory For Beginners",
            "source_url": "https://www.youtube.com/playlist?list=PLCTMeyjMKRkoS699U0OJ3ymr3r01sI08l",
        },
        "items": items,
    }
    output_path.write_text(json.dumps(manifest, indent=2, ensure_ascii=False) + "\n")
    print(f"wrote {output_path} with {len(items)} items")


if __name__ == "__main__":
    main()
