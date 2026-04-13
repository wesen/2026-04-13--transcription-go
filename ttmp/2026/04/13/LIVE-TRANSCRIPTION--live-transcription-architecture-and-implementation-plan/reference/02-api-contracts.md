---
Title: Live transcription API contracts
Ticket: LIVE-TRANSCRIPTION
Status: active
Topics:
    - go
    - dagger
    - asr
    - streaming
    - websocket
    - transcription
    - audio
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/transcribe/live.go
      Note: |-
        Replay-oriented live CLI flags reflect the current Phase 1 source strategy
        API contract doc now reflects WS as the default live transport
    - Path: internal/asr/client.go
      Note: Implemented Go-side batch and chunk HTTP contracts
    - Path: internal/asr/client_test.go
      Note: Contract tests for multipart full/chunk uploads
    - Path: internal/live/metrics.go
      Note: Phase 1 live runner now emits a machine-readable replay summary artifact
    - Path: internal/live/replay_source.go
      Note: Phase 1 simulated live source now uses replayed WAV input rather than directory watching
    - Path: internal/live/runner.go
      Note: Summary artifact is now refreshed incrementally during replay
    - Path: internal/live/sinks.go
      Note: Phase 1 live runner can now persist committed transcript outputs to console/text/subtitle/sqlite artifacts
    - Path: internal/live/stream_receiver.go
      Note: Current Go WS receiver semantics reflected in the API contract
    - Path: internal/live/stream_sender.go
      Note: Current Go WS sender pacing/flush semantics reflected in the API contract
    - Path: internal/live/wsclient.go
      Note: Current Go WS event handling reflected in the API contract
    - Path: out-live-ws-clip-000-015/live-summary.json
      Note: WS smoke-run evidence referenced by the API contract
    - Path: out-live-ws-clip-000-120-fix/live-summary.json
      Note: Corrected 120s WS validation evidence referenced in the API contract
    - Path: out-live-ws-clip-000-120-fix/transcript.db
      Note: Corrected 120s WS timestamp/word-count comparison evidence referenced in the API contract
    - Path: server/live_decoder.py
      Note: |-
        Current buffered WS decoder behavior and constraints reflected in the API contract
        Current WS decoder pts-anchoring behavior reflected in the API contract
    - Path: server/live_sessions.py
      Note: Current WS session registry and structured error behavior reflected in the API contract
    - Path: server/server.py
      Note: |-
        Implemented FastAPI batch/chunk endpoints and shared helper flow
        Implemented WS event handling and endpoint semantics reflected in the API contract
    - Path: ttmp/internal/asr/client.go
      Note: Go-side HTTP contract for full-file and chunk uploads
    - Path: ttmp/server/server.py
      Note: Python-side FastAPI contract for batch and chunk transcription
ExternalSources: []
Summary: Reference document for the current batch and chunk HTTP contracts and the planned WebSocket streaming protocol for true real-time transcription.
LastUpdated: 2026-04-13T00:00:00Z
WhatFor: Give implementers a concrete request/response schema reference for Phase 1 near-live mode and the future session-oriented streaming transport.
WhenToUse: Use when implementing or reviewing client/server protocol changes for live transcription.
---










# Live transcription API contracts

## Purpose

This document captures the protocol shape for three layers of the system:

1. the existing batch API,
2. the implemented chunk API used for near-live mode,
3. the planned WebSocket streaming API for the real production architecture.

For Phase 1 simulated live testing, the current recommended source is a prerecorded WAV replayed on a synthetic timeline rather than a filesystem chunk watcher. That source choice shapes the current client-side runner, but does not change the transport contracts below.

Current operational stance:

- **default live transport:** `ws`
- **fallback/debug transport:** `chunk`

The intent is to make the transport boundary explicit so the Go and Python implementations can evolve without ambiguity.

---

## 1. Current batch HTTP contract

### Endpoint

```text
POST /transcribe/full
```

### Transport

`multipart/form-data`

### Request fields

- `file`: uploaded WAV file
- `chunk_size`: integer seconds per internal processing chunk

### Response

```json
{
  "words": [
    {"word": "hello", "start": 0.12, "end": 0.44},
    {"word": "world", "start": 0.45, "end": 0.81}
  ],
  "total_duration": 1672.431,
  "chunk_count": 28,
  "word_count": 4226
}
```

### Semantics

- The request is batch-oriented and uploads the full converted WAV.
- The server chunks internally and returns a single final transcript.
- Word-level timestamps are required and are already part of the contract.

---

## 2. Phase 1 near-live chunk HTTP contract

### Endpoint

```text
POST /transcribe/chunk
```

### Transport

`multipart/form-data`

### Request fields

- `file`: uploaded WAV chunk
- `session_id`: logical live session identifier
- `chunk_index`: monotonically increasing integer
- `chunk_start`: absolute start time in the overall recording timeline, seconds
- `overlap_seconds`: overlap window with the previous chunk, seconds
- `is_final_chunk`: whether this is the last chunk in the near-live run

### Response

```json
{
  "session_id": "session-123",
  "chunk_index": 7,
  "chunk_start": 12.5,
  "chunk_duration": 1.5,
  "words": [
    {"word": "hello", "start": 12.50, "end": 12.80},
    {"word": "world", "start": 12.81, "end": 13.10}
  ],
  "partial": false,
  "processing_ms": 187
}
```

### Semantics

- `chunk_start` defines the absolute timeline base used for emitted word timestamps.
- `words` must preserve word-level timestamps.
- `partial` is currently always `false` for Phase 1 chunk mode.
- overlap handling is primarily a client-side accumulator concern.

### Current implementation notes

- The HTTP chunk contract is still implemented and supported.
- It is now primarily a fallback/debug comparison path rather than the recommended default live transport.
- The Python server normalizes the uploaded audio chunk to `16kHz mono PCM16 WAV` before inference.
- The Go client preserves the exact metadata fields above.
- FastAPI form fields must be declared with `Form(...)` when sent alongside file uploads.
- The current Phase 1 live runner can now persist committed transcript state to:
  - console output,
  - `transcript.srt`,
  - `transcript.vtt`,
  - `transcript.txt`,
  - `transcript.db`
  under the configured live `--output-dir`.
- The current Phase 1 live runner also writes `live-summary.json` under `--output-dir`, capturing replay/session metrics such as chunk count, committed word count, average server processing time, average end-to-end latency, and effective audio-seconds-per-wall-second throughput.
- `live-summary.json` is refreshed during the replay as chunks are processed, not only at final completion.

---

## 3. Planned streaming WebSocket contract

This is the target protocol for true real-time transcription.

### Endpoint

```text
WS /transcribe/stream
```

## 3.1 Client → server events

### `start`

```json
{
  "type": "start",
  "session_id": "session-123",
  "sample_rate": 16000,
  "channels": 1,
  "format": "pcm_s16le",
  "source": "replay_wav|stdin|mic|system-audio"
}
```

### `audio`

```json
{
  "type": "audio",
  "sequence": 42,
  "pts": 84.000,
  "duration": 0.500,
  "pcm16_base64": "..."
}
```

### `flush`

```json
{"type": "flush"}
```

### `stop`

```json
{"type": "stop"}
```

## 3.2 Server → client events

### `started`

```json
{
  "type": "started",
  "session_id": "session-123"
}
```

### `partial`

```json
{
  "type": "partial",
  "session_id": "session-123",
  "sequence": 42,
  "text": "hello wor",
  "words": [
    {"word": "hello", "start": 84.22, "end": 84.51}
  ]
}
```

### `final_words`

```json
{
  "type": "final_words",
  "session_id": "session-123",
  "up_to_time": 84.90,
  "words": [
    {"word": "hello", "start": 84.22, "end": 84.51},
    {"word": "world", "start": 84.52, "end": 84.90}
  ]
}
```

### `stopped`

```json
{
  "type": "stopped",
  "session_id": "session-123",
  "word_count": 927,
  "duration": 302.4
}
```

### `error`

```json
{
  "type": "error",
  "session_id": "session-123",
  "code": "invalid-sequence",
  "message": "expected sequence 41, got 44"
}
```

### Current server-side implementation notes

The Python service now has an initial session-oriented WebSocket implementation under:

- `server/server.py`
- `server/live_sessions.py`
- `server/live_decoder.py`

Current behavior/constraints:

- `start` creates a server-side session registry entry keyed by `session_id`.
- `audio` appends base64-decoded `pcm_s16le` bytes into a per-session buffered decoder.
- `partial` events are emitted from the currently buffered region once enough unfinalized audio has accumulated.
- `flush` emits `final_words` for the current buffered region and advances the session's finalized timeline.
- `stop` performs a final flush, emits `stopped`, and removes the session.
- The current implementation only supports:
  - `sample_rate=16000`
  - `channels=1`
  - `format="pcm_s16le"`
- Session sequencing is enforced on the server; out-of-order `audio.sequence` values produce structured `error` events.
- Sessions are cleaned up on:
  - idle timeout (`LIVE_SESSION_IDLE_TIMEOUT_SECONDS`, default `300`)
  - explicit `stop`
  - broken WebSocket connection
  - app shutdown
- Partial preview decode currently uses a buffered file-based decode behind the session boundary rather than true model-state streaming. This is a transport-correct first implementation, not yet the final optimized streaming decoder.

---

## 4. Transcript-state invariants

These invariants apply across chunk and streaming modes.

- Word-level timestamps are required.
- Final words are authoritative and persistence-grade.
- Partial words may be shown to users, but should not be treated as durable transcript state.
- Client-side committed transcript state must be monotonic in time.
- SRT/VTT/SQLite outputs should derive from committed/final words only.

### Current Go-side implementation notes

The Go live path now has an explicit transport-neutral transcript-state model under `internal/live/`:

- `TranscriptEvent` distinguishes `partial` from `final_words`
- `TranscriptState` exposes `Committed`, `Pending`, and `LastFinalTime`
- the accumulator rejects out-of-order numbered events
- partial updates replace the pending preview state
- final updates append only monotonic, non-duplicate words to committed state
- after finalization, sinks derive durable artifacts from committed words only; console output may additionally surface pending preview text

This is intentionally aligned with the planned WebSocket event model so the chunk-based proving path and the future streaming path can share the same accumulator semantics.

### Current Go-side WebSocket transport notes

The Go live path now also has an initial WebSocket transport implementation under `internal/live/`, and this is now the default `transcribe live` transport:

- `WSLiveClient` handles connect/start/audio/flush/stop plus inbound JSON messages
- `SendAudioFrames(...)` pushes replay PCM16 chunks over the WS transport
- `ReceiveResultEvents(...)` converts WS server messages into `TranscriptEvent` values
- `LiveRunner` now supports both `chunk` and `ws` transports via `--transport`

Current behavior/constraints:

- replay input chunks now carry both `WAVPath` and raw `PCM16` bytes so they can feed either transport
- the initial WS sender flushes after each replay chunk and waits for the corresponding finalization before advancing to the next chunk
- partial WS events update pending preview state, but durable artifacts still only use committed/final words
- buffered WS decoder spans are now anchored to incoming audio `pts`, so replay chunk overlap does not accumulate into artificial timestamp drift
- the current WS path is transport-correct and session-oriented, but it intentionally preserves chunk-level pacing/finalization to keep validation and attribution simple while the decoder remains buffered server-side

### Current WS smoke-run evidence

A short replay-driven WS smoke run completed successfully in tmux using:

```bash
go run ./cmd/transcribe live \
  -i /tmp/transcription-live-clip-000-015.wav \
  -o ./out-live-ws-clip-000-015 \
  --transport ws \
  --live-format console,db,txt \
  --chunk-duration 5 \
  --overlap-seconds 0.5 \
  --replay-speed 0
```

Observed artifacts/results:

- initial output dir: `out-live-ws-clip-000-015/`
- corrected output dir after timestamp-anchoring fix: `out-live-ws-clip-000-015-fix/`
- corrected artifacts: `transcript.db`, `transcript.txt`, `live-summary.json`
- corrected `transcript.db` coverage:
  - `min_start=2.24`
  - `max_end=15.02`
- corrected `live-summary.json`:
  - `chunks_processed=4`
  - `committed_words=23`
  - `effective_audio_seconds=15.0`
  - `average_server_processing_ms=852.75`
  - `average_end_to_end_ms=3233.25`
- runtime evidence included:
  - `Preview: Welcome back to the Go Golems lab.`
  - `Committed +7 words: Welcome back to the Go Golems lab.`
  - `Processed live chunk transport=ws seq=0 ...`

### Current 120s WS comparison evidence

A corrected 120-second replay-driven WS run completed with:

- output dir: `out-live-ws-clip-000-120-fix/`
- `transcript.db` coverage: `120.12s`
- `live-summary.json`:
  - `chunks_processed=27`
  - `committed_words=318`
  - `effective_audio_seconds=120.0`
  - `average_server_processing_ms=974.148`
  - `average_end_to_end_ms=3887.852`
- comparison against the same-pipeline batch baseline:
  - WS live words: `318`
  - batch words: `323`
  - delta: `-5`
  - coverage delta: `+0.04s`
- comparison against the earlier HTTP chunk-live run:
  - WS live words: `318`
  - HTTP chunk-live words: `302`
  - delta: `+16`

---

## 5. File references

- Go client:
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/asr/client.go`
- Python server:
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/server/server.py`
- Main design doc:
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/design-doc/01-live-transcription-architecture-design-and-implementation-guide.md`
