---
Title: Live transcription architecture and implementation plan
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
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
  - Path: ../../../../../cmd/transcribe/main.go
    Note: Current batch CLI flow that live mode must preserve rather than replace
  - Path: ../../../../../internal/server/dagger.go
    Note: Current Dagger service and host tunnel lifecycle; foundation for live mode
  - Path: ../../../../../internal/asr/client.go
    Note: Current full-file request contract that needs new live-facing siblings
  - Path: ../../../../../server/server.py
    Note: Current FastAPI/Nemotron service; starting point for chunk and streaming APIs
ExternalSources: []
Summary: "Research ticket for designing the correct live transcription solution: chunked near-live as an intermediate phase, and a session-oriented streaming API as the real production architecture."
LastUpdated: 2026-04-13T00:00:00Z
WhatFor: "Orient engineers and track the design/implementation work required to add live transcription to the current Go+Dagger Nemotron system."
WhenToUse: "Use when planning or implementing live/near-live transcription in this repository, or when onboarding engineers to the architecture and phased execution plan."
---

# Live transcription architecture and implementation plan

## Overview

This ticket captures the design and implementation plan for adding live transcription to the current `transcription-go` repository. The current system already provides the key foundation — a warm, Dagger-managed ASR service and a Go-owned output pipeline — but it is still batch-oriented. This ticket explains the current architecture, the missing pieces for live mode, and the recommended path to get from the current batch system to a production-ready streaming system.

The primary recommendation is deliberate: implement a chunked near-live path first to prove behavior and latency, but treat a session-oriented streaming API over WebSocket as the actual long-term production solution.

## Key Links

- Design doc: [design-doc/01-live-transcription-architecture-design-and-implementation-guide.md](./design-doc/01-live-transcription-architecture-design-and-implementation-guide.md)
- Diary: [reference/01-investigation-diary.md](./reference/01-investigation-diary.md)
- Tasks: [tasks.md](./tasks.md)
- Changelog: [changelog.md](./changelog.md)

## Current status

Current status: **active**

What has been done:

- ticket workspace created
- primary design doc written
- architecture mapped against current repository files
- phased implementation strategy defined

What remains:

- validate docmgr doctor cleanly
- upload bundle to reMarkable
- optionally expand with API-contract or playbook subdocs if implementation begins soon

## Main conclusion

The correct long-term design is **not** repeated reuse of the current `/transcribe/full` endpoint. The correct production shape is:

- explicit live sessions,
- explicit partial vs final result semantics,
- explicit streaming transport,
- explicit client-side transcript accumulation and finalization.

## Structure

- `design-doc/` — primary analysis and implementation guide
- `reference/` — diary and future API contracts / quick references
- `playbooks/` — future operator or QA procedures
- `scripts/` — future replay/probe helpers
- `sources/` — supporting copied artifacts if needed later
