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
    - Path: cmd/transcribe/batch.go
      Note: Step 2 extracted batch implementation (commit 23f88e5d123a9482b0ca39519876eded5788f119)
    - Path: cmd/transcribe/live.go
      Note: Step 2 live command scaffold (commit 23f88e5d123a9482b0ca39519876eded5788f119)
    - Path: cmd/transcribe/main.go
      Note: Batch CLI flow inspected during live-architecture mapping
    - Path: cmd/transcribe/root.go
      Note: Step 2 CLI mode split and command wiring (commit 23f88e5d123a9482b0ca39519876eded5788f119)
    - Path: internal/asr/client.go
      Note: |-
        Current full-file API client inspected to identify the live transport gap
        Step 3 chunk client and shared multipart helpers (commit d14a887adafe235d1a6ebbd4e519f8e479252cbd)
    - Path: internal/asr/client_test.go
      Note: Step 3 client contract tests (commit d14a887adafe235d1a6ebbd4e519f8e479252cbd)
    - Path: internal/live/runner.go
      Note: Step 2 live runner scaffold (commit 23f88e5d123a9482b0ca39519876eded5788f119)
    - Path: internal/live/source.go
      Note: Step 2 audio source contract scaffold (commit 23f88e5d123a9482b0ca39519876eded5788f119)
    - Path: internal/server/dagger.go
      Note: Warm Dagger service boundary inspected as live-mode foundation
    - Path: internal/server/helpers.go
      Note: Step 2 shared server startup/path helpers (commit 23f88e5d123a9482b0ca39519876eded5788f119)
    - Path: server/server.py
      Note: |-
        Current full-file chunking implementation inspected for service-boundary refactoring
        Step 3 chunk API and Form-based multipart contract (commit d14a887adafe235d1a6ebbd4e519f8e479252cbd)
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

## Step 2: Split the CLI into explicit batch/live paths and scaffold the live package

The first implementation step focused on changing the repository shape without changing the working batch behavior. That meant separating the CLI into explicit `batch` and `live` paths, extracting shared server bootstrap helpers, and creating the first `internal/live` package boundaries so future tasks would have a stable place to land.

This was intentionally a structural step rather than a feature-completion step. The goal was to reduce the chance that early live-transcription work would destabilize the already validated batch pipeline.

### Prompt Context

**User prompt (verbatim):** "Ok, now work on it task by task, committing at appropriate intervals, and keeping a detailed diary"

**Assistant interpretation:** Begin implementing the live-transcription ticket incrementally, keep the work sliced into reviewable commits, and document each meaningful step in detail.

**Inferred user intent:** Make steady engineering progress toward live transcription while preserving the ability to review, validate, and continue the work later without losing context.

**Commit (code):** 23f88e5d123a9482b0ca39519876eded5788f119 — "Refactor CLI for batch/live transcription modes"

### What I did

- Rewrote `cmd/transcribe/main.go` to delegate to a new root-command constructor.
- Added `cmd/transcribe/root.go` with:
  - root command,
  - explicit `batch` subcommand,
  - explicit `live` subcommand.
- Added `cmd/transcribe/batch.go` and moved the previous batch implementation there.
- Kept the root invocation behaving like batch mode so existing usage still works.
- Added `cmd/transcribe/live.go` with a scaffolded live command and future-facing flags:
  - `--chunk-dir`
  - `--session-id`
  - `--overlap-seconds`
  - `--live-format`
- Added `internal/server/helpers.go` with:
  - `ResolveServerDir(...)`
  - `StartDefault(...)`
  - `DefaultPort`
- Added `internal/live/source.go`, `internal/live/accumulator.go`, and `internal/live/runner.go` as the first live package skeleton.
- Ran:
  - `gofmt -w cmd/transcribe/*.go internal/server/*.go internal/live/*.go`
  - `go test ./... -count=1`
  - `go run ./cmd/transcribe --help`
  - `go run ./cmd/transcribe live --help`

### Why

The ticket’s first implementation phase called for coexistence, not replacement. The batch path already works and should remain the stable reference implementation while live mode is built beside it.

### What worked

- The CLI now has a clean place to add live behavior without mixing it into the batch orchestration code.
- Existing batch behavior was preserved at the root command, which reduces migration friction.
- The `internal/live` package now exists, so later tasks can add real implementations without another structural refactor.
- Formatting and the full Go test suite passed after the refactor.

### What didn't work

- N/A in the sense of a failing build or test run.
- One design tension appeared immediately: the task asked for explicit `batch` and `live` modes, but the repository also already had a working root command. I resolved that by keeping the root invocation as batch mode while still adding explicit subcommands.

### What I learned

- This repository benefits from early package-boundary work because the original batch CLI concentrated most orchestration in a single file.
- Preserving the root batch invocation is a low-cost way to get explicit mode separation without making existing usage clumsier.

### What was tricky to build

The tricky part was defining “explicit subcommands” without regressing the established batch workflow. A strict switch to `transcribe batch` only would have been clean architecturally, but would also have changed the user-facing entrypoint immediately. Keeping the root command as a batch alias preserved the stable behavior while still giving live mode its own namespace.

### What warrants a second pair of eyes

- Whether the live runner scaffold should remain in `internal/live/runner.go` or eventually be split into multiple orchestration files earlier than planned
- Whether the root-command-as-batch-alias behavior is desirable long term, or whether the CLI should eventually require explicit subcommands once live mode is mature

### What should be done in the future

- Implement the chunk API and matching Go client next
- Add a dedicated API-contract reference doc to the ticket
- Replace the placeholder live runner and accumulator implementations with real behavior in subsequent phases

### Code review instructions

Start here:

- `/home/manuel/code/wesen/2026-04-13--transcription-go/cmd/transcribe/root.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/cmd/transcribe/batch.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/cmd/transcribe/live.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/server/helpers.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/runner.go`

Validation commands:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go

gofmt -w cmd/transcribe/*.go internal/server/*.go internal/live/*.go
go test ./... -count=1
go run ./cmd/transcribe --help
go run ./cmd/transcribe live --help
```

### Technical details

Key symbols added in this step:

- `newRootCmd()`
- `runBatch(...)`
- `runLive(...)`
- `server.ResolveServerDir(...)`
- `server.StartDefault(...)`
- `live.AudioChunk`
- `live.AudioSource`
- `live.LiveRunner`

## Step 3: Add the chunk transcription API, Go client, and protocol documentation

The second implementation step moved from structure into the first actual near-live transport. I added a dedicated `POST /transcribe/chunk` endpoint on the Python server, a matching Go client method and tests, and a reference document in the ticket that now records the implemented HTTP contracts and the planned streaming WebSocket protocol.

This step is important because it is the first place where the live-transcription design becomes executable rather than purely conceptual. It also locked in the rule that live mode must preserve word-level timestamps, not just plain text.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Continue the live-transcription implementation by completing the next concrete tasks and keep preserving reviewable checkpoints in the diary and git history.

**Inferred user intent:** See the design ticket converted into real protocol and code increments, with enough documentation to onboard the next engineer or resume later.

**Commit (code):** d14a887adafe235d1a6ebbd4e519f8e479252cbd — "Add chunk transcription API and Go client"

### What I did

- Extended `server/server.py` with:
  - `POST /transcribe/chunk`
  - `_normalize_wav(...)`
  - `_transcribe_chunk_file(...)`
- Refactored `POST /transcribe/full` to reuse the new normalization/transcription helper path.
- Added diagnostic logging for:
  - full-transcription chunk progress,
  - chunk-transcription requests,
  - processing time.
- Fixed multipart metadata parsing in FastAPI by declaring non-file fields with `Form(...)`.
- Expanded `internal/asr/client.go` with:
  - `ChunkRequest`
  - `ChunkResponse`
  - `TranscribeChunk(...)`
  - shared multipart upload/JSON decode helpers.
- Added `internal/asr/client_test.go` with tests for:
  - `TranscribeChunk(...)`
  - `TranscribeFull(...)`
- Added a new ticket reference doc:
  - `reference/02-api-contracts.md`
- Ran:
  - `gofmt -w internal/asr/client.go internal/asr/client_test.go`
  - `go test ./... -count=1`
  - `python3 -m py_compile server/server.py`

### Why

The chunk API is the smallest real near-live transport that fits the long-term design. It gives the Go side a proper live-facing contract while still keeping the Python service simple enough to validate incrementally.

### What worked

- The new server endpoint preserves word-level timestamps and emits chunk metadata that the Go side will need later.
- The Go client now supports both batch and chunk modes through a shared upload helper path.
- The new tests exercised the HTTP contract and confirmed the multipart field names/values.
- Python syntax validation and the full Go test suite passed.

### What didn't work

- The first cut of the FastAPI chunk endpoint used plain typed parameters for multipart metadata instead of `Form(...)`. That would have caused the server to look for query parameters rather than multipart form fields when files and metadata are sent together.
- I caught and fixed that before finalizing the step by changing:
  - `chunk_size`
  - `session_id`
  - `chunk_index`
  - `chunk_start`
  - `overlap_seconds`
  - `is_final_chunk`
  to `Form(...)` declarations.

### What I learned

- The multipart/form contract is worth documenting immediately because both FastAPI and Go are permissive enough that incorrect field plumbing can look plausible until runtime.
- Writing the Go client tests early is especially useful here because the server and client now have a second transport contract beyond the original batch endpoint.

### What was tricky to build

The tricky part here was making the chunk API feel like a real new transport rather than a one-off helper. If the endpoint had been added without shared normalization/transcription helpers or without explicit metadata fields, it would have been harder to evolve into a proper live runner later. Refactoring the Python side at the same time kept the batch and chunk paths aligned.

### What warrants a second pair of eyes

- Whether server-side normalization via `ffmpeg` is the right long-term behavior for chunk mode, or whether live inputs should be normalized entirely on the Go side before upload
- Whether the chunk response needs additional fields soon, such as server-side queue/backlog timing or confidence values
- Whether `partial` should remain a hard-coded `false` field in Phase 1 or be omitted until true partial semantics exist

### What should be done in the future

- Implement the chunk-directory source and replay source
- Add the live runner wiring on top of the new chunk client
- Begin the accumulator/deduplication logic needed to merge overlapping chunk results into a stable transcript

### Code review instructions

Start here:

- `/home/manuel/code/wesen/2026-04-13--transcription-go/server/server.py`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/asr/client.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/asr/client_test.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/reference/02-api-contracts.md`

Validation commands:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go

gofmt -w internal/asr/client.go internal/asr/client_test.go
go test ./... -count=1
python3 -m py_compile server/server.py
```

### Technical details

Implemented Phase 1 contract shape:

```json
{
  "session_id": "session-123",
  "chunk_index": 7,
  "chunk_start": 12.5,
  "chunk_duration": 1.5,
  "words": [{"word": "hello", "start": 12.5, "end": 12.8}],
  "partial": false,
  "processing_ms": 187
}
```

Important implementation details:

- FastAPI multipart form fields are declared with `Form(...)`
- word-level timestamps remain required in the live-facing chunk contract
- the Go client now shares multipart upload logic between full-file and chunk requests

## Step 4: Pivot the Phase 1 live source toward WAV replay instead of directory watching

After the first near-live transport work landed, the next design choice was about the source model for Phase 1. The original task list included a chunk-directory watcher as the first live input source, but that now looks like the wrong default because the long-term destination is a WebSocket streaming transport, not a filesystem-driven integration.

The user’s suggestion was the right simplification: use a prerecorded WAV as a simulated streaming source, emit timed frames/chunks from it, and make that replay source the main proving path. That keeps the implementation aligned with the eventual streaming sender loop and avoids building a source mode we may never actually use.

### Prompt Context

**User prompt (verbatim):** "i think we should just simulate with taking a wav and pretending its a streaming source, I don't think we'll actually use the chunk directory approach, since we will go to websocket ultimately"

**Assistant interpretation:** Re-prioritize the Phase 1 source strategy so the primary near-live input is a replayed WAV file rather than a watched chunk directory.

**Inferred user intent:** Avoid investing in an intermediate source abstraction that does not pull its weight, and keep the proving phase closely aligned with the eventual WebSocket streaming architecture.

### What I did

- Updated the ticket task list so that:
  - WAV-backed replay becomes the primary simulated live source
  - chunk-directory ingestion is explicitly deferred/optional
  - live CLI planning now centers on replay input and pacing controls
  - replay-based validation is framed around deterministic WAV slicing rather than directory fixtures

### Why

A WAV replay source exercises the parts of the system that actually matter for the long-term design:

- timeline-aware chunk/frame emission,
- pacing,
- transport boundaries,
- transcript accumulation,
- output stability.

A chunk-directory watcher mostly exercises filesystem coordination, which is not on the critical path for the eventual streaming architecture.

### What worked

- The task plan was easy to simplify without invalidating the work already done.
- This decision makes the next implementation step clearer: build a replay source that can later feed either chunk HTTP uploads or WebSocket audio frames.

### What didn't work

- N/A as a code/build issue.
- Conceptually, this exposed that the earlier task list still contained one source mode that was more “possible” than “strategically important”.

### What I learned

- The right proving source is the one that best approximates the final transport semantics, not necessarily the one that is easiest to sketch in isolation.
- WAV replay is a better bridge to the WebSocket model because it naturally produces ordered, timestamped frames/chunks under pacing control.

### What was tricky to build

The tricky part here is architectural prioritization rather than implementation mechanics. Directory watching can look attractive because it is concrete and easy to reason about, but it risks pulling the design toward a file-ingestion mindset when the real target is session-oriented media streaming. The fix is to privilege the source model that matches the future sender loop.

### What warrants a second pair of eyes

- Whether the replay source should emit WAV-backed chunks first or PCM frames directly
- Whether the first replay path should target the HTTP chunk API or whether it should immediately be shaped to resemble the future WebSocket sender abstraction
- What pacing controls are most useful for development versus CI replay tests

### What should be done in the future

- Implement the WAV replay source next
- Add CLI flags for replay input and pacing
- Use the replay source as the main validation harness for both chunk mode and the later WebSocket transport

### Code review instructions

Review the updated tasks in:

- `/home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/tasks.md`

Confirm that the new priority order matches the architecture goal: simulated streaming from WAV first, transport-aligned abstractions second, directory watching only if later needed.

### Technical details

The task changes were:

- old primary source:
  - chunk-directory watcher
- new primary source:
  - WAV-backed replay source with pacing control
- deferred source:
  - chunk-directory ingestion, only if a real external integration later requires it
