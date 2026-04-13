---
Title: Investigation diary
Ticket: LIVE-TRANSCRIPTION
Status: active
Topics:
    - go
    - dagger
    - asr
    - streaming
    - websocket
    - transcription
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/transcribe/main.go
      Note: Batch CLI flow inspected during live-architecture mapping
    - Path: internal/asr/client.go
      Note: Current full-file API client inspected to identify the live transport gap
    - Path: internal/server/dagger.go
      Note: Warm Dagger service boundary inspected as live-mode foundation
    - Path: server/server.py
      Note: Current full-file chunking implementation inspected for service-boundary refactoring
    - Path: ttmp/cmd/transcribe/main.go
      Note: Current batch CLI flow inspected while designing live-mode coexistence
    - Path: ttmp/internal/asr/client.go
      Note: Current full-file request contract inspected to identify the gap to chunk and streaming APIs
    - Path: ttmp/internal/server/dagger.go
      Note: Current Dagger service/tunnel lifecycle inspected as the basis for the future live runtime
    - Path: ttmp/server/server.py
      Note: Current FastAPI model startup and full-file chunking logic inspected to define the future live service boundary
ExternalSources: []
Summary: Chronological diary for the live-transcription design ticket, capturing how the current repository was mapped and why the long-term recommendation is a session-oriented streaming API.
LastUpdated: 2026-04-13T00:00:00Z
WhatFor: Record the design investigation so future engineers can understand how the recommendation was derived.
WhenToUse: Use when continuing or reviewing the live transcription design work.
---


# Investigation diary

## Goal

Capture the architecture-mapping and design reasoning behind the new live-transcription ticket so a future engineer can pick up the work without reconstructing the current system from scratch.

## Step 1: Map the current batch pipeline and identify the real architectural gap

The first task was not to invent a live architecture from scratch, but to understand exactly what the current repository already does well. That matters because the right live design should reuse the hard-won parts of the current system — especially the Dagger service lifecycle and the Go-owned output layer — rather than replacing them.

The most important conclusion from the current-state review is that the repository is already structurally close to live transcription, but its external contract is still batch-oriented. The gap is therefore not “make ASR work”; it is “introduce session state, rolling audio ingestion, and partial/final transcript semantics without disturbing the current working batch path”.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket and write a detailed implementation plan for the correct solution. 

reate a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable.

We also want to ultimately have the nice long-term path for real production use."

**Assistant interpretation:** Create a new ticket dedicated to live transcription, map the existing repository carefully, and write a very detailed intern-facing design/implementation guide that recommends the correct production architecture rather than a quick hack.

**Inferred user intent:** Preserve the live-transcription design work as a durable engineering deliverable that can guide future implementation and onboarding, and publish it to both the repository ticket workspace and reMarkable.

### What I did

- Created a new ticket workspace:
  - `LIVE-TRANSCRIPTION`
- Added two documents:
  - primary design doc
  - investigation diary
- Inspected and line-mapped the current implementation files:
  - `cmd/transcribe/main.go`
  - `internal/server/dagger.go`
  - `internal/asr/client.go`
  - `server/server.py`
  - `internal/convert/convert.go`
  - `internal/output/format.go`
  - `internal/output/sqlite.go`
- Compared the current system against the requirements of a real live system.
- Wrote the design doc with a phased recommendation:
  - preserve batch mode,
  - add chunked near-live as the proving phase,
  - aim at a session-oriented WebSocket streaming architecture for production use.

### Why

The user explicitly asked for the “correct solution” and also explicitly said the long-term path should be suitable for real production use. That rules out stopping at a purely chunk-polling workaround. A detailed design guide therefore needed to distinguish:

- what is expedient,
- what is a good proving phase,
- what is the real target architecture.

### What worked

- The current repository structure is clean enough to support a strong design recommendation.
- The Dagger service/tunnel architecture already gives the repository a warm ASR service boundary, which is exactly the operational foundation that live mode needs.
- The Go output layer is already distinct from the Python inference layer, which makes the transcript-state design easier to evolve.

### What didn't work

- The initial write of the design doc was long enough to cut through the middle of a Mermaid block and had to be completed in a second pass.
- The ticket generator’s default index/tasks/diary documents were skeletal and needed to be rewritten to make the ticket continuation-friendly.

### What I learned

- The current repository’s biggest live-mode limitation is not performance; it is protocol shape.
- The easiest wrong answer is to repeatedly reuse `/transcribe/full` and call that “live”. That may work as an experiment, but it is not the right production design.
- The cleanest architecture is one that preserves batch mode and adds a separate live path with explicit session and result semantics.

### What was tricky to build

The trickiest part of the design work was avoiding a false binary between “quick hack” and “perfect streaming system”. In practice the right answer is phased:

1. define the production target clearly,
2. choose an intermediate implementation phase that proves the product behavior,
3. make sure Phase 1 code is structured so it can evolve into the final system instead of being thrown away.

That is what led to the recommendation of chunked near-live mode first, but only as a deliberate step toward a session-oriented streaming architecture.

### What warrants a second pair of eyes

- Whether the current Nemotron runtime exposes a sufficiently clean incremental/streaming decode model for the final WebSocket design
- Whether live persistence should reuse the current SQLite schema directly or introduce explicit pending/final state tables
- Whether direct audio capture should ever live in this repo, or remain outside via chunk-directory or pipe-based source integrations

### What should be done in the future

- Run `docmgr doctor` and fix any vocabulary or metadata issues
- Upload the completed ticket bundle to reMarkable
- If implementation starts soon, add a dedicated API-contract reference doc and a replay-test playbook

### Code review instructions

Start with the design doc, then compare its claims against these files:

- `/home/manuel/code/wesen/2026-04-13--transcription-go/cmd/transcribe/main.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/server/dagger.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/asr/client.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/server/server.py`

Validation approach:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go
rg -n "TranscribeFull|AsService|Tunnel|/transcribe/full|BuildSegments|WriteSQLite" cmd internal server -S
```

Confirm that the current implementation is indeed batch-oriented and that the live design recommendations correspond to the actual current boundaries.

### Technical details

The key file-backed observations used in the design were:

- batch CLI orchestration:
  - `cmd/transcribe/main.go:64-138`
- Dagger service + host tunnel lifecycle:
  - `internal/server/dagger.go:27-91`
- full-file multipart upload contract:
  - `internal/asr/client.go:63-115`
- full-file server endpoint with internal chunking:
  - `server/server.py:58-108`
- Go-owned subtitle formatting:
  - `internal/output/format.go:16-109`
- Go-owned SQLite writing:
  - `internal/output/sqlite.go:10-145`
