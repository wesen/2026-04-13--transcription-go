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

