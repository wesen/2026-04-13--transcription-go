---
Title: Live transcription operator playbook
Ticket: LIVE-TRANSCRIPTION
Status: active
Topics:
    - go
    - dagger
    - asr
    - websocket
    - transcription
    - tmux
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/transcribe/live.go
      Note: CLI flags and WS-default transport behavior used by the operator playbook
    - Path: internal/live/metrics.go
      Note: live-summary.json output described by the playbook
    - Path: internal/live/runner.go
      Note: Transport-aware live runner invoked by the playbook commands
    - Path: internal/live/wsclient.go
      Note: Go WS transport used by the default live path described in the playbook
    - Path: out-live-ws-clip-000-120-fix/live-summary.json
      Note: Example corrected WS artifact referenced by the playbook
    - Path: server/server.py
      Note: Python ASR/WS server started by the live workflow described in the playbook
    - Path: ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/01-compare_transcript_dbs.py
      Note: |-
        Quick transcript DB comparison helper for post-run checks
        Comparison helper used in the playbook
    - Path: ttmp/cmd/transcribe/live.go
      Note: CLI entrypoint and flags for live replay operation
    - Path: ttmp/internal/live/metrics.go
      Note: live-summary.json output used for runtime monitoring
    - Path: ttmp/internal/live/runner.go
      Note: Transport-aware live runner for ws and chunk modes
    - Path: ttmp/internal/live/wsclient.go
      Note: Go WebSocket live transport used by the default live path
    - Path: ttmp/out-live-ws-clip-000-120-fix/live-summary.json
      Note: Example corrected WS validation artifact showing expected summary fields
    - Path: ttmp/server/server.py
      Note: Python ASR server exposing HTTP and WS live endpoints
ExternalSources: []
Summary: Practical runbook for starting, monitoring, debugging, and comparing live transcription runs, with WebSocket as the default live transport and chunk transport retained as a fallback/debug mode.
LastUpdated: 2026-04-13T00:00:00Z
WhatFor: Give operators and developers a fast, practical procedure for running live transcription in tmux and understanding the resulting artifacts.
WhenToUse: Use when starting a live replay run, monitoring progress, checking artifacts, or comparing output quality across transports.
---


# Live transcription operator playbook

## Status summary

Current recommended live path:

- **default transport:** `ws`
- **fallback/debug transport:** `chunk`

Current recommendation:

- use **tmux** for any non-trivial run
- prefer **WS transport** unless you are deliberately comparing against the older chunk path
- use short clipped WAV subsets for fast debugging
- rely on `live-summary.json`, `transcript.db`, and the tmux/log output for live monitoring

---

## 1. Start a short WS replay in tmux

Example 120-second run:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go

python3 ./ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/02-extract_wav_segment.py \
  --input /home/manuel/code/wesen/2026-04-09--screencast-studio/recordings/rabbit-hole-2026-04-10--2/audio-mix.wav \
  --output /tmp/transcription-live-clip-000-120.wav \
  --start 0 \
  --duration 120

mkdir -p logs
SESSION="live-ws-clip-120s"
LOG="logs/${SESSION}-$(date +%Y%m%d-%H%M%S).log"

tmux new-session -d -s "$SESSION" \
  "cd /home/manuel/code/wesen/2026-04-13--transcription-go && \
   go run ./cmd/transcribe live \
     -i /tmp/transcription-live-clip-000-120.wav \
     -o ./out-live-ws-clip-000-120 \
     --transport ws \
     --live-format console,db,txt \
     --chunk-duration 5 \
     --overlap-seconds 0.5 \
     --replay-speed 0 2>&1 | tee $LOG"
```

Notes:

- `--transport ws` is explicit above for clarity, but current `transcribe live` now defaults to WS.
- `--transport chunk` is still available for fallback/debug comparison.

---

## 2. Poll a run without blocking yourself

Capture current pane output:

```bash
tmux capture-pane -pt live-ws-clip-120s | tail -n 80
```

Attach interactively:

```bash
tmux attach -t live-ws-clip-120s
```

Tail the log directly:

```bash
tail -f logs/live-ws-clip-120s-YYYYMMDD-HHMMSS.log
```

List tmux sessions:

```bash
tmux list-sessions
```

Kill a session if needed:

```bash
tmux kill-session -t live-ws-clip-120s
```

---

## 3. What healthy runtime output looks like

Healthy startup usually includes:

- `Tunnel established at 127.0.0.1:...`
- `Health check passed (attempt 1)`
- `ASR server ready at 127.0.0.1:...`

Healthy WS replay usually includes lines like:

- `Preview: Welcome back to the Go Golems lab.`
- `Committed +7 words: Welcome back to the Go Golems lab.`
- `Processed live chunk transport=ws seq=0 ...`

On completion, look for:

- `Live replay complete: session=... transport=ws ...`

---

## 4. Output artifacts to inspect

Given `-o ./out-live-ws-clip-000-120`, the important files are:

- `out-live-ws-clip-000-120/live-summary.json`
- `out-live-ws-clip-000-120/transcript.db`
- `out-live-ws-clip-000-120/transcript.txt`
- optional subtitle outputs if requested via `--live-format srt,vtt`

### Quick summary check

```bash
cat out-live-ws-clip-000-120/live-summary.json
```

Important fields:

- `chunks_processed`
- `committed_words`
- `effective_audio_seconds`
- `elapsed_seconds`
- `audio_seconds_per_wall_second`
- `average_server_processing_ms`
- `average_end_to_end_ms`

### Quick DB check

```bash
python3 - <<'PY'
import sqlite3
conn = sqlite3.connect('out-live-ws-clip-000-120/transcript.db')
try:
    print(conn.execute('select count(*), min(start_time), max(end_time) from words').fetchone())
finally:
    conn.close()
PY
```

---

## 5. Compare WS output against another baseline

### WS vs batch baseline

```bash
python3 ./ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/01-compare_transcript_dbs.py \
  --live-db out-live-ws-clip-000-120/transcript.db \
  --reference-db out-batch-clip-000-120/transcript.db \
  --summary-json out-live-ws-clip-000-120/live-summary.json
```

### WS vs chunk fallback run

```bash
python3 ./ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/01-compare_transcript_dbs.py \
  --live-db out-live-ws-clip-000-120/transcript.db \
  --reference-db out-live-clip-000-120/transcript.db \
  --summary-json out-live-ws-clip-000-120/live-summary.json
```

---

## 6. When to use chunk transport intentionally

Use `--transport chunk` only when you want to:

- compare WS and chunk behavior on the same input
- debug WS-specific protocol/session issues
- preserve an older fallback path during transition

Example:

```bash
go run ./cmd/transcribe live \
  -i /tmp/transcription-live-clip-000-120.wav \
  -o ./out-live-chunk-clip-000-120 \
  --transport chunk \
  --live-format console,db,txt \
  --chunk-duration 5 \
  --overlap-seconds 0.5 \
  --replay-speed 0
```

---

## 7. Common failure patterns

### Startup takes a long time

Expected causes:

- Dagger engine/service startup
- Python model load
- cold HuggingFace/pip caches

Healthy signs:

- repeated Dagger startup logs
- eventual `ASR server ready at ...`

### WS timestamps drift past the clip duration

This used to happen before the WS decoder was anchored to incoming `pts`.

Quick check:

```bash
python3 - <<'PY'
import sqlite3
conn = sqlite3.connect('out-live-ws-clip-000-120/transcript.db')
try:
    print(conn.execute('select min(start_time), max(end_time) from words').fetchone())
finally:
    conn.close()
PY
```

If `max(end_time)` is much larger than the known clip duration, something is wrong.

### No progress after server startup

Check:

- tmux pane for `Processed live chunk ...`
- `live-summary.json` existence and refresh behavior
- `transcript.db` existence and word count

---

## 8. Current practical recommendation

Use this default mental model:

- **batch** = stable offline reference
- **ws live** = primary live path
- **chunk live** = fallback/debug comparison path

That is the current operator stance for this repository.
