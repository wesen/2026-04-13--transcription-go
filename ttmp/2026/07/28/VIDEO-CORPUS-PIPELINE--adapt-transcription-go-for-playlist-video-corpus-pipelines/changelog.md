# Changelog

## 2026-07-28

- Initial workspace created


## 2026-07-28

Completed direct current-state analysis and authored the intern architecture/implementation guide, corpus schema/API contract, and operator playbook.

### Related Files

- /home/manuel/worktrees/2026-07-28--transcription-go-video-pipeline/ttmp/2026/07/28/VIDEO-CORPUS-PIPELINE--adapt-transcription-go-for-playlist-video-corpus-pipelines/design-doc/01-intern-guide-playlist-video-corpus-transcription-architecture-and-implementation.md — Primary intern-facing design
- /home/manuel/worktrees/2026-07-28--transcription-go-video-pipeline/ttmp/2026/07/28/VIDEO-CORPUS-PIPELINE--adapt-transcription-go-for-playlist-video-corpus-pipelines/reference/01-corpus-database-and-pipeline-api-contracts.md — Implementation contracts


## 2026-07-28

Validated all Go tests and ticket metadata, then uploaded Video Corpus Transcription Intern Guide.pdf to /ai/2026/07/28/VIDEO-CORPUS-PIPELINE.

### Related Files

- /home/manuel/worktrees/2026-07-28--transcription-go-video-pipeline/ttmp/2026/07/28/VIDEO-CORPUS-PIPELINE--adapt-transcription-go-for-playlist-video-corpus-pipelines/reference/02-investigation-diary.md — Validation and publication evidence


## 2026-07-28

Added implementation-phase tasks; the design deliverable is complete but the active ticket remains open for product work.


## 2026-07-28

Implemented corpus pipeline end to end: Phase 0 fixes, internal/corpus package, transcribe corpus CLI, manifest converter, and validated with Southwell video 019 (467 words, 24 chunks, resume verified).

### Related Files

- cmd/transcribe/corpus.go — Corpus command group
- internal/corpus/store.go — Atomic commit and resume planner
- scripts/convert_media_manifest.py — Manifest converter


## 2026-07-28

Streamed multipart uploads via io.Pipe and started the full 36-video Southwell corpus run in tmux.

### Related Files

- internal/asr/client.go — Streaming upload implementation


## 2026-07-28

Video 001 (58 min) completed: 6850 words, 360 chunks, all DB invariants pass. Full corpus run ongoing in tmux.

