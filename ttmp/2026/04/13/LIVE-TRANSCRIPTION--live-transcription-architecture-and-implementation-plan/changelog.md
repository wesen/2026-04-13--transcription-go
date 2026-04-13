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

