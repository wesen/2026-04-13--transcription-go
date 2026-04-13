# Changelog

## 2026-04-13

- Initial workspace created
- Added primary design doc: live transcription architecture, design, and implementation guide
- Mapped the current batch pipeline to concrete files (`cmd/transcribe`, `internal/server`, `internal/asr`, `server/server.py`, `internal/output`)
- Recommended chunked near-live mode as the proving phase and a session-oriented streaming API as the correct long-term production solution
- Prepared ticket for validation and reMarkable upload

## 2026-04-13

Created live-transcription research ticket, mapped the current batch system, and wrote a detailed intern-facing design/implementation guide recommending a session-oriented streaming architecture as the production target.

### Related Files

- /home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/design-doc/01-live-transcription-architecture-design-and-implementation-guide.md — Primary deliverable
- /home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/reference/01-investigation-diary.md — Chronological design reasoning and evidence


## 2026-04-13

Expanded LIVE-TRANSCRIPTION tasks into a detailed phased implementation plan for true real-time streaming transcription, covering coexistence with batch mode, chunked proving phase, WebSocket session transport, transcript-state management, testing, and production hardening.

### Related Files

- /home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/tasks.md — Detailed actionable task breakdown for live streaming implementation


## 2026-04-13

Step 2: split the CLI into explicit batch/live paths, extracted shared server helpers, and scaffolded the internal/live package (commit 23f88e5d123a9482b0ca39519876eded5788f119).

### Related Files

- /home/manuel/code/wesen/2026-04-13--transcription-go/cmd/transcribe/root.go — Root/batch/live command wiring
- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/server/helpers.go — Shared server bootstrap helpers


## 2026-04-13

Step 3: added the near-live chunk transcription API, Go chunk client, contract tests, and API-contract reference doc (commit d14a887adafe235d1a6ebbd4e519f8e479252cbd).

### Related Files

- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/asr/client.go — Chunk request/response support
- /home/manuel/code/wesen/2026-04-13--transcription-go/server/server.py — Chunk endpoint and shared normalization/transcription helpers
- /home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/reference/02-api-contracts.md — Documented implemented and planned protocols


## 2026-04-13

Step 4: re-prioritized Phase 1 to use a WAV-backed replay source as the main simulated live input, deferring chunk-directory ingestion unless a real integration later requires it.

### Related Files

- /home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/reference/01-investigation-diary.md — Recorded the source-strategy pivot and rationale
- /home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/tasks.md — Phase 1 source strategy updated to match the eventual WebSocket architecture


## 2026-04-13

Step 5: implemented a WAV-backed replay source, overlap-aware Phase 1 accumulator, console sink, and a runnable transcribe-live chunk loop over the new /transcribe/chunk API (commit b5aadd7617d06c596af6d165db58d94e7f26754d).

### Related Files

- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/accumulator.go — Overlap-aware committed transcript accumulation
- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/replay_source.go — Simulated live replay source
- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/runner.go — Live runner now uploads replayed chunks


## 2026-04-13

Step 6: fixed Python startup churn by caching pip install behind requirements.txt and aligning Lightning pins with NeMo, then successfully validated the live replay path in tmux through real /transcribe/chunk processing (commit 180dfdae073da2ba9469062b5ca0efcf45b7fbd4).

### Related Files

- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/server/dagger.go — Improved Dagger dependency-layer caching for iterative server work
- /home/manuel/code/wesen/2026-04-13--transcription-go/server/requirements.txt — Aligned Python pins to reduce resolver backtracking


## 2026-04-13

Clarified in the diary why Python resolver backtracking appeared during live work: a latent dependency issue was exposed when server-side code edits invalidated the old Dagger install layer cache.

### Related Files

- /home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/reference/01-investigation-diary.md — Added explicit explanation of the cache-vs-resolution behavior


## 2026-04-13

Step 7: added rolling live output sinks for SRT/VTT/TXT/SQLite plus --output-dir support, so the replay-driven live path now persists committed transcript state instead of only logging to console (commit 0fc41bf9f2a2664eadbaaa1aed66e6a8f30ffa53).

### Related Files

- /home/manuel/code/wesen/2026-04-13--transcription-go/cmd/transcribe/live.go — Live output-dir CLI flag
- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/sqlite_sink.go — Live SQLite artifact generation
- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/subtitle_sink.go — Live subtitle/text artifact generation


## 2026-04-13

Step 8: added live replay metrics collection and a live-summary.json artifact with chunk count, committed words, average latency, and throughput fields (commit 6560072a7d7c567ba3e0408d9a0a99494a21eeaf).

### Related Files

- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/metrics.go — Replay metrics summary implementation
- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/runner.go — Runner now writes metrics summary at completion


## 2026-04-13

Step 9: validated mid-run live artifacts in tmux (transcript.srt and transcript.db) and updated the live runner to refresh live-summary.json after each chunk for better observability (commit 3ff211a09fdc4d2bb2f9d557d153fd8138cc0c17).

### Related Files

- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/runner.go — Incremental summary artifact refresh
- /home/manuel/code/wesen/2026-04-13--transcription-go/out-live-e2e/transcript.db — Observed rolling SQLite artifact during replay validation
- /home/manuel/code/wesen/2026-04-13--transcription-go/out-live-e2e/transcript.srt — Observed rolling subtitle artifact during replay validation


## 2026-04-13

Step 10: added a ticket-local transcript DB comparison script and used it for the first partial live-vs-reference comparison against the in-progress out-live-e2e transcript.db.

### Related Files

- /home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/01-compare_transcript_dbs.py — Reusable live-vs-reference SQLite comparison helper


## 2026-04-13

Step 11: recorded the completed full-run live replay results (3845 vs 4248 words, delta -403) and established a fast 120s clipped-WAV iteration workflow, which yielded a much smaller same-pipeline live-vs-batch delta of -21 words (302 vs 323).

### Related Files

- /home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/02-extract_wav_segment.py — Fast iteration helper for clipped WAV subsets



## 2026-04-13

Step 12: inspected the actual ordered words in the 120-second live and batch transcript DBs, wrote a dedicated ticket report for the mismatch regions, and normalized all ticket helper scripts to numbered `01-` / `02-` / `03-` names.

### Related Files

- /home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/reference/03-word-analysis-report.md — Word-level analysis of missing/substituted/compressed regions on the 120-second subset
- /home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/03-word_diff_report.py — Reusable direct word-sequence diff helper
- /home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/01-compare_transcript_dbs.py — Numbered comparison helper name used by the ticket going forward
- /home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/02-extract_wav_segment.py — Numbered clipped-WAV helper name used by the ticket going forward


## 2026-04-13

Step 13: introduced an explicit partial/final transcript-state model in Go, refactored the accumulator and sinks around `TranscriptEvent` / `TranscriptState`, and added tests for partial revisions, finalization, duplicate filtering, and out-of-order event rejection (commit 64bb79c).

### Related Files

- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/types.go — Transport-neutral transcript event/state definitions for current chunk mode and future WebSocket mode
- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/accumulator.go — Explicit pending vs committed transcript handling with monotonic finalization and sequence validation
- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/runner.go — Current chunk replay path now feeds the accumulator through transcript events
- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/console_sink.go — Console sink can surface committed additions and pending preview text
- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/subtitle_sink.go — Durable subtitle/text artifacts now derive from committed state only
- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/sqlite_sink.go — Durable SQLite artifacts now derive from committed state only
- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/accumulator_test.go — Added coverage for repeated partial revisions, promotion to final, overlap filtering, and out-of-order rejection


## 2026-04-13

Step 14: added the first server-side WebSocket/session scaffold with a live session registry, buffered per-session decoder, `WS /transcribe/stream`, structured session errors, cleanup policy, and lightweight Python tests (commit e84dc41).

### Related Files

- /home/manuel/code/wesen/2026-04-13--transcription-go/server/live_sessions.py — Server-side live session registry, lifecycle, and structured sequence/session errors
- /home/manuel/code/wesen/2026-04-13--transcription-go/server/live_decoder.py — Buffered per-session decoder scaffold used by the initial WS implementation
- /home/manuel/code/wesen/2026-04-13--transcription-go/server/server.py — FastAPI now exposes `WS /transcribe/stream` beside the existing batch/chunk endpoints
- /home/manuel/code/wesen/2026-04-13--transcription-go/server/live_sessions_test.py — Python tests for decoder/session invariants without loading the ASR model


## 2026-04-13

Step 15: added the Go WebSocket live transport, including client connect/start/audio/flush/stop handling, sender/receiver loops, a transport-aware live runner, replay PCM16 support, a lightweight WS integration test, and a real 15s replay-driven WS smoke validation run (commit cd945df).

### Related Files

- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/wsclient.go — Go WebSocket client for the session-oriented live API
- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/stream_sender.go — Replay PCM16 sender loop with per-chunk flush/finalization pacing
- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/stream_receiver.go — WS message receiver that converts server events into transcript events
- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/runner.go — Live runner now supports both chunk and WS transports
- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/replay_source.go — Replay source now exposes raw PCM16 bytes for the WS transport
- /home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/wsclient_test.go — Lightweight Go WS integration test
- /home/manuel/code/wesen/2026-04-13--transcription-go/out-live-ws-clip-000-015/live-summary.json — Runtime evidence from a successful 15-second replay-driven WS smoke run
