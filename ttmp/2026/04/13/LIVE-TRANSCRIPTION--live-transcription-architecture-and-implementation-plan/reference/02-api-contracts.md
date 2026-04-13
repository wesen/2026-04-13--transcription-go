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
    - Path: internal/asr/client.go
      Note: Implemented Go-side batch and chunk HTTP contracts
    - Path: internal/asr/client_test.go
      Note: Contract tests for multipart full/chunk uploads
    - Path: server/server.py
      Note: Implemented FastAPI batch/chunk endpoints and shared helper flow
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

- The Python server normalizes the uploaded audio chunk to `16kHz mono PCM16 WAV` before inference.
- The Go client preserves the exact metadata fields above.
- FastAPI form fields must be declared with `Form(...)` when sent alongside file uploads.

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
  "source": "chunk-dir|stdin|mic|system-audio"
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

---

## 4. Transcript-state invariants

These invariants apply across chunk and streaming modes.

- Word-level timestamps are required.
- Final words are authoritative and persistence-grade.
- Partial words may be shown to users, but should not be treated as durable transcript state.
- Client-side committed transcript state must be monotonic in time.
- SRT/VTT/SQLite outputs should derive from committed/final words only.

---

## 5. File references

- Go client:
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/asr/client.go`
- Python server:
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/server/server.py`
- Main design doc:
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/design-doc/01-live-transcription-architecture-design-and-implementation-guide.md`
