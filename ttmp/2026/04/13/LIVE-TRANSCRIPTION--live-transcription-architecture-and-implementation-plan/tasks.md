# Tasks

## Guiding rule

- Preserve the current batch transcription path as the stable reference implementation.
- Build chunked near-live mode only as a proving phase.
- Treat session-oriented streaming transcription as the real target architecture.

---

## Phase 0 — Preparation and coexistence with batch mode

- [x] 0.1 Split CLI modes clearly
  - Add explicit `batch` and `live` subcommands (or equivalent mode separation)
  - Keep current batch behavior unchanged
  - Ensure existing end-to-end batch run still works
  - Files:
    - `cmd/transcribe/main.go`

- [x] 0.2 Extract shared server bootstrap helpers
  - Refactor current Dagger service startup into reusable helpers usable by both batch and live paths
  - Ensure service startup/health-check behavior remains identical
  - Files:
    - `internal/server/dagger.go`
    - possibly new `internal/server/helpers.go`

- [x] 0.3 Define live package boundaries
  - Create `internal/live/` package scaffold
  - Introduce core types:
    - `AudioChunk`
    - `AudioSource`
    - `TranscriptAccumulator`
    - `LiveRunner`
  - Files:
    - `internal/live/source.go`
    - `internal/live/accumulator.go`
    - `internal/live/runner.go`

- [x] 0.4 Write API contract reference doc
  - Add a dedicated reference doc describing chunk API and WebSocket event schemas
  - Ticket doc output, not app code
  - Files:
    - `ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/reference/02-api-contracts.md`

---

## Phase 1 — Chunked near-live mode (proving phase)

### 1A. Server-side chunk API

- [x] 1.1 Add `POST /transcribe/chunk`
  - Accept a small WAV chunk upload
  - Accept metadata:
    - `session_id`
    - `chunk_index`
    - `chunk_start`
    - `overlap_seconds` (optional)
    - `is_final_chunk` (optional)
  - Return word-level timestamps for that chunk
  - Files:
    - `server/server.py`

- [x] 1.2 Refactor server transcription helpers
  - Extract reusable helper for “transcribe one chunk and return words”
  - Reuse `_extract_words(...)`
  - Avoid duplicating logic between `/transcribe/full` and `/transcribe/chunk`
  - Files:
    - `server/server.py`

- [x] 1.3 Add server-side progress/diagnostic logging
  - Log `session_id`, `chunk_index`, `chunk_start`, `processing_ms`
  - Make logs readable during long runs
  - Files:
    - `server/server.py`

### 1B. Go client for chunk API

- [x] 1.4 Add chunk response model
  - Define Go struct for chunk-transcription response
  - Preserve word-level timestamps
  - Files:
    - `internal/asr/client.go`

- [x] 1.5 Add `TranscribeChunk(...)`
  - Upload a small chunk to `/transcribe/chunk`
  - Include timing/session metadata in request
  - Keep `TranscribeFull(...)` intact
  - Files:
    - `internal/asr/client.go`

- [x] 1.6 Add chunk-client tests
  - Mock server responses
  - Verify request fields and response decoding
  - Files:
    - `internal/asr/client_test.go`

### 1C. Live source abstraction

- [x] 1.7 Define `AudioSource` interface
  - Source should produce ordered `AudioChunk` values
  - Include `SessionID`, `Sequence`, `Start`, `Duration`, and PCM/WAV payload
  - Files:
    - `internal/live/source.go`

- [x] 1.8 Implement WAV-backed replay source as the primary simulated live input
  - Read a prerecorded WAV file and emit fixed-duration frames/chunks on a simulated timeline
  - Support both real-time pacing and accelerated replay for tests
  - Make this the main Phase 1 source because it maps naturally onto the future WebSocket sender loop
  - Files:
    - `internal/live/replay_source.go`

- [ ] 1.9 Make chunk-directory ingestion optional/deferred
  - Do not treat directory watching as the default near-live path
  - Only add it later if an external recorder integration actually needs it
  - Files:
    - `internal/live/chunk_dir_source.go` (only if later justified)

### 1D. Client-side accumulation and rolling output

- [x] 1.10 Implement initial transcript accumulator
  - Maintain:
    - committed words
    - pending words
    - last finalized time
  - Support overlap dedupe for chunk responses
  - Files:
    - `internal/live/accumulator.go`

- [x] 1.11 Implement overlap dedupe heuristics
  - Drop near-duplicate words across adjacent chunks
  - Use timing tolerance and textual comparison
  - Add tests for repeated overlap windows
  - Files:
    - `internal/live/accumulator.go`
    - `internal/live/accumulator_test.go`

- [x] 1.12 Implement rolling console output
  - Print live transcript updates without flooding the terminal excessively
  - Distinguish pending from committed output if useful
  - Files:
    - `internal/live/console_sink.go`

- [x] 1.13 Implement rolling subtitle sink
  - Emit committed transcript into rolling SRT/VTT-friendly segments
  - Do not finalize subtitle text from non-final words
  - Files:
    - `internal/live/subtitle_sink.go`
    - maybe `internal/output/format.go`

- [x] 1.14 Implement optional live SQLite sink
  - Persist only committed words
  - Decide whether pending words stay in memory only for Phase 1
  - Files:
    - `internal/live/sqlite_sink.go`
    - maybe `internal/output/sqlite.go`

### 1E. Live CLI wiring

- [x] 1.15 Add `transcribe live` CLI command
  - Add flags for:
    - `--input` or another explicit replay WAV source flag
    - `--session-id`
    - `--overlap-seconds`
    - `--live-format`
    - replay pacing controls (real-time vs accelerated)
  - Files:
    - `cmd/transcribe/main.go`
    - possibly `cmd/transcribe/live.go`

- [x] 1.16 Add live runner
  - Wire together:
    - Dagger service startup
    - source
    - chunk client
    - accumulator
    - sinks
  - Files:
    - `internal/live/runner.go`

### 1F. Phase 1 validation

- [x] 1.17 Create replay-based validation fixtures
  - Use prerecorded WAV inputs and deterministic frame/chunk slicing for simulated live tests
  - Prefer a single WAV replay flow over directory-driven chunk fixtures
  - Files:
    - `ttmp/.../scripts/` or repo test fixtures dir
    - `internal/live/testdata/`

- [x] 1.18 Measure near-live latency
  - Capture:
    - chunk creation time
    - request send time
    - response receipt time
    - transcript commit time
  - Files:
    - `internal/live/runner.go`
    - maybe `internal/live/metrics.go`

- [x] 1.19 Compare near-live output against batch output
  - Confirm word counts and transcript quality are within acceptable range
  - Validate overlap logic does not cause runaway duplication
  - Output to ticket docs
  - Current evidence:
    - full 27.7m live replay vs older reference DB: `3845` vs `4248` words (`-403`)
    - fast 120s subset live replay vs same-pipeline batch DB: `302` vs `323` words (`-21`)
    - direct word-level analysis report: `reference/03-word-analysis-report.md`

---

## Phase 2 — Formal transcript-state model

- [x] 2.1 Introduce explicit partial vs final state model
  - Add result/event types that distinguish preview text from committed words
  - Files:
    - `internal/live/types.go`
    - `internal/live/accumulator.go`

- [x] 2.2 Define transcript finalization rules
  - Document exactly when a word moves from pending to committed
  - Ensure finalization is monotonic in time
  - Files:
    - `internal/live/accumulator.go`
    - ticket reference docs

- [x] 2.3 Refactor rolling outputs to depend only on committed state
  - Console may show pending state
  - SRT/VTT/SQLite should use committed state only
  - Files:
    - `internal/live/console_sink.go`
    - `internal/live/subtitle_sink.go`
    - `internal/live/sqlite_sink.go`

- [x] 2.4 Add transcript-state tests
  - Cases:
    - repeated partial revisions
    - partial → final promotion
    - overlap-induced duplicates
    - out-of-order chunk arrival rejection/handling
  - Files:
    - `internal/live/accumulator_test.go`

---

## Phase 3 — True real-time streaming transport (production target)

### 3A. Protocol and session model

- [x] 3.1 Define WebSocket message schema
  - Client events:
    - `start`
    - `audio`
    - `flush`
    - `stop`
  - Server events:
    - `started`
    - `partial`
    - `final_words`
    - `stopped`
    - `error`
  - Files:
    - `ttmp/.../reference/02-api-contracts.md`
    - `server/` protocol helpers
    - `internal/live/` protocol types

- [x] 3.2 Introduce server-side session registry
  - Track active sessions by `session_id`
  - Manage lifecycle and cleanup
  - Files:
    - `server/live_sessions.py` (new)
    - `server/server.py`

- [x] 3.3 Define per-session decoder state object
  - Hold incremental audio buffer / state
  - Expose methods:
    - append audio
    - decode incrementally
    - flush
    - close
  - Files:
    - `server/live_decoder.py` (new)
    - `server/live_sessions.py`

### 3B. Server-side streaming API

- [x] 3.4 Add `WS /transcribe/stream`
  - Accept session start event
  - Receive audio frames continuously
  - Emit partial and final result events
  - Files:
    - `server/server.py`

- [x] 3.5 Add structured error handling for live sessions
  - Invalid event types
  - Sequence errors
  - Session-not-found
  - Decode errors
  - Files:
    - `server/server.py`
    - `server/live_sessions.py`

- [x] 3.6 Add session cleanup policy
  - Idle timeout
  - Stop event cleanup
  - Broken connection cleanup
  - Files:
    - `server/live_sessions.py`

### 3C. Go streaming client

- [ ] 3.7 Add WebSocket live client
  - Connect/disconnect lifecycle
  - Send control/audio messages
  - Receive result events
  - Files:
    - `internal/live/wsclient.go`

- [ ] 3.8 Add audio-frame sender loop
  - Push audio frames at real-time or replay pace
  - Include timestamps/sequence numbers
  - Files:
    - `internal/live/stream_sender.go`

- [ ] 3.9 Add result-event receiver loop
  - Decode partial/final events
  - Feed accumulator
  - Files:
    - `internal/live/stream_receiver.go`

- [ ] 3.10 Make live runner transport-agnostic
  - Support both:
    - chunk transport
    - WebSocket transport
  - Files:
    - `internal/live/runner.go`

### 3D. Phase 3 validation

- [ ] 3.11 Add WebSocket integration tests
  - Start session
  - Send sample frames
  - Observe partial/final events
  - Stop cleanly
  - Files:
    - `internal/live/wsclient_test.go`
    - Python-side tests if added

- [ ] 3.12 Add replay harness for streaming mode
  - Feed prerecorded chunks/frames as if live
  - Compare resulting transcript to batch baseline
  - Files:
    - `internal/live/replay_source.go`
    - test harness scripts

- [ ] 3.13 Verify word-level timestamp quality in streaming mode
  - Confirm final emitted words retain reliable `start` and `end`
  - Compare against batch output quality where possible

---

## Phase 4 — Production hardening

- [ ] 4.1 Add per-session structured logging
  - Include:
    - `session_id`
    - `sequence`
    - `event_type`
    - `processing_ms`
    - `queue_depth`
  - Files:
    - `server/` live session modules
    - `internal/live/` runner/client

- [ ] 4.2 Add metrics collection
  - Measure:
    - partial latency
    - finalization latency
    - real-time factor
    - dropped frames
    - reconnects
  - Files:
    - `internal/live/metrics.go`
    - `server/` metrics hooks

- [ ] 4.3 Add backpressure/buffer policy
  - Define what happens when sender outruns decoder
  - Bound memory usage explicitly
  - Files:
    - `server/live_sessions.py`
    - `server/live_decoder.py`

- [ ] 4.4 Add reconnect/resume policy
  - Decide whether reconnect creates a new session or resumes an existing one
  - Document guarantees clearly
  - Files:
    - `internal/live/wsclient.go`
    - protocol docs

- [ ] 4.5 Add operational playbook
  - Start/stop/debug live service
  - Replay test procedure
  - Latency diagnosis checklist
  - Files:
    - `ttmp/.../playbooks/01-live-transcription-operator-playbook.md`

---

## Cross-cutting requirements

- [ ] Preserve word-level timestamps in all live result paths
  - chunk mode responses
  - partial event previews if possible
  - final words definitely

- [ ] Preserve batch transcription output parity as much as practical
  - Live mode should be comparable against batch mode, not a separate unverifiable system

- [ ] Keep Dagger service warm and reusable
  - Avoid regressions that restart model load unnecessarily between live sessions

- [ ] Keep all user-visible formatting logic in Go
  - Python should remain inference/service focused

---

## Deliverables checklist

- [ ] Detailed API contract doc for chunk + WebSocket modes
- [ ] Near-live chunk mode implemented and validated
- [ ] Transcript accumulator with explicit partial/final semantics
- [ ] Streaming WebSocket transport implemented end-to-end
- [ ] Replay-based test harness for deterministic validation
- [ ] Operational playbook and latency/quality validation notes
- [ ] Ticket docs updated with findings after each phase
