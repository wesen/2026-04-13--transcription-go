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
      Note: |-
        Step 2 live command scaffold (commit 23f88e5d123a9482b0ca39519876eded5788f119)
        Step 5 replay-oriented live CLI flags and runner wiring (commit b5aadd7617d06c596af6d165db58d94e7f26754d)
        Step 7 output-dir support for live replay (commit 0fc41bf9f2a2664eadbaaa1aed66e6a8f30ffa53)
        Step 15 live CLI transport selection flag
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
    - Path: internal/live/accumulator.go
      Note: |-
        Step 5 overlap-aware Phase 1 accumulator (commit b5aadd7617d06c596af6d165db58d94e7f26754d)
        Step 13 explicit pending vs committed state handling and sequence validation
    - Path: internal/live/accumulator_test.go
      Note: |-
        Step 5 overlap-deduplication tests (commit b5aadd7617d06c596af6d165db58d94e7f26754d)
        Step 13 transcript-state tests for partial/final promotion and out-of-order rejection
    - Path: internal/live/console_sink.go
      Note: |-
        Step 5 incremental committed-word console output (commit b5aadd7617d06c596af6d165db58d94e7f26754d)
        Step 13 console sink consumes transcript state snapshots
    - Path: internal/live/metrics.go
      Note: Step 8 live replay metrics collection and summary artifact (commit 6560072a7d7c567ba3e0408d9a0a99494a21eeaf)
    - Path: internal/live/metrics_test.go
      Note: Step 8 metrics summary tests (commit 6560072a7d7c567ba3e0408d9a0a99494a21eeaf)
    - Path: internal/live/replay_source.go
      Note: |-
        Step 5 WAV-backed simulated live source (commit b5aadd7617d06c596af6d165db58d94e7f26754d)
        Step 15 replay source now exposes PCM16 bytes for WS transport
    - Path: internal/live/replay_source_test.go
      Note: Step 5 replay-source chunking tests (commit b5aadd7617d06c596af6d165db58d94e7f26754d)
    - Path: internal/live/runner.go
      Note: |-
        Step 2 live runner scaffold (commit 23f88e5d123a9482b0ca39519876eded5788f119)
        Step 5 runnable live chunk loop over replay input (commit b5aadd7617d06c596af6d165db58d94e7f26754d)
        Step 13 current chunk replay path now feeds TranscriptEvent values
        Step 15 transport-aware live runner for chunk and ws modes
    - Path: internal/live/sinks.go
      Note: Step 7 live sink construction from requested formats (commit 0fc41bf9f2a2664eadbaaa1aed66e6a8f30ffa53)
    - Path: internal/live/source.go
      Note: Step 2 audio source contract scaffold (commit 23f88e5d123a9482b0ca39519876eded5788f119)
    - Path: internal/live/sqlite_sink.go
      Note: |-
        Step 7 rolling SQLite artifact sink (commit 0fc41bf9f2a2664eadbaaa1aed66e6a8f30ffa53)
        Step 13 sqlite sink consumes committed state only
    - Path: internal/live/stream_receiver.go
      Note: Step 15 WS result receiver loop converting server events into transcript events
    - Path: internal/live/stream_sender.go
      Note: Step 15 WS audio sender loop with per-chunk flush/finalization pacing
    - Path: internal/live/subtitle_sink.go
      Note: |-
        Step 7 rolling subtitle/text artifact sink (commit 0fc41bf9f2a2664eadbaaa1aed66e6a8f30ffa53)
        Step 13 subtitle/text sink consumes committed state only
    - Path: internal/live/types.go
      Note: Step 13 transport-neutral transcript event/state definitions for future WS work
    - Path: internal/live/wsclient.go
      Note: Step 15 Go WebSocket client implementation for the live API
    - Path: internal/live/wsclient_test.go
      Note: Step 15 lightweight Go WS integration test
    - Path: internal/server/dagger.go
      Note: |-
        Warm Dagger service boundary inspected as live-mode foundation
        Step 6 dependency-layer caching fix for iterative live validation (commit 180dfdae073da2ba9469062b5ca0efcf45b7fbd4)
    - Path: internal/server/helpers.go
      Note: Step 2 shared server startup/path helpers (commit 23f88e5d123a9482b0ca39519876eded5788f119)
    - Path: out-live-e2e/transcript.db
      Note: Step 9 real persisted-output validation artifact inspected during live replay
    - Path: out-live-e2e/transcript.srt
      Note: Step 9 real persisted-output validation artifact inspected during live replay
    - Path: out-live-ws-clip-000-015-fix/transcript.db
      Note: Step 16 corrected short WS smoke validation with proper end time
    - Path: out-live-ws-clip-000-015/live-summary.json
      Note: Step 15 successful 15s replay-driven WS smoke run metrics
    - Path: out-live-ws-clip-000-120-fix/live-summary.json
      Note: Step 16 corrected 120s WS metrics used for comparison evidence
    - Path: out-live-ws-clip-000-120-fix/transcript.db
      Note: Step 16 corrected 120s WS transcript DB used for batch/chunk comparison
    - Path: server/live_decoder.py
      Note: |-
        Step 14 buffered per-session decoder used by the initial WS transport
        Step 16 fix for pts-anchored WS timestamping over overlapped replay chunks
    - Path: server/live_sessions.py
      Note: |-
        Step 14 server-side live session registry and cleanup policy
        Step 16 session path now passes pts into the buffered decoder
    - Path: server/live_sessions_test.py
      Note: |-
        Step 14 lightweight Python tests for session and decoder invariants
        Step 16 Python tests cover pts-anchored buffered decoding
    - Path: server/requirements.txt
      Note: Step 6 Python pin alignment to reduce resolver backtracking (commit 180dfdae073da2ba9469062b5ca0efcf45b7fbd4)
    - Path: server/server.py
      Note: |-
        Current full-file chunking implementation inspected for service-boundary refactoring
        Step 3 chunk API and Form-based multipart contract (commit d14a887adafe235d1a6ebbd4e519f8e479252cbd)
        Step 14 FastAPI WebSocket endpoint for session-oriented live transcription
    - Path: ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/01-compare_transcript_dbs.py
      Note: Step 10 partial live-vs-reference comparison helper
    - Path: ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/02-extract_wav_segment.py
      Note: Step 11 clipped-WAV fast-iteration helper
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

## Step 5: Implement the WAV replay source, near-live accumulator, and runnable live chunk loop

This step turned the live-transcription ticket from mostly protocol and scaffolding work into a runnable Phase 1 system. The new `transcribe live` path now accepts a prerecorded WAV file, converts it to the repository’s standard 16kHz mono format, replays it as timed chunks, uploads each chunk to the new `/transcribe/chunk` endpoint, accumulates overlap-aware transcript words, and prints incremental committed transcript output to the console.

This is still not the final WebSocket streaming architecture, but it is now a real simulated live pipeline rather than a design placeholder. Just as importantly, the chosen source model — replaying a WAV on a timeline — aligns much better with the eventual sender loop than the earlier chunk-directory idea would have.

### Prompt Context

**User prompt (verbatim):** (same as Step 4)

**Assistant interpretation:** Continue implementing the next live-transcription tasks using the newly agreed WAV-replay source model rather than a filesystem-driven source.

**Inferred user intent:** Turn the design pivot into working code that exercises the transport, timing, and accumulation boundaries the future WebSocket system will need.

**Commit (code):** b5aadd7617d06c596af6d165db58d94e7f26754d — "Add WAV replay source and runnable live chunk pipeline"

### What I did

- Extended `cmd/transcribe/live.go` to support replay-oriented live flags:
  - `--input`
  - `--chunk-duration`
  - `--overlap-seconds`
  - `--replay-speed`
  - `--session-id`
  - `--live-format`
- Extended `internal/live/source.go` so emitted chunks now carry `EmittedAt` timing in addition to sequence/timeline metadata.
- Added `internal/live/replay_source.go`:
  - reads a WAV file,
  - slices it into fixed-duration chunks with overlap,
  - writes temporary chunk WAVs,
  - emits them on a simulated timeline,
  - supports pacing control through `ReplaySpeed`.
- Added `internal/live/console_sink.go` for incremental committed-word console output.
- Replaced the placeholder accumulator in `internal/live/accumulator.go` with a Phase 1 overlap-aware accumulator that:
  - tracks committed words,
  - keeps monotonic finalization state,
  - suppresses obvious overlapping duplicates.
- Added tests:
  - `internal/live/replay_source_test.go`
  - `internal/live/accumulator_test.go`
- Rewrote `internal/live/runner.go` so the live runner now:
  - validates replay input,
  - converts the input WAV to 16k mono,
  - starts the Dagger ASR service,
  - replays chunks through `/transcribe/chunk`,
  - feeds the accumulator,
  - prints incremental transcript output,
  - logs chunk-level timing/progress.
- Ran:
  - `gofmt -w cmd/transcribe/live.go internal/live/*.go`
  - `go test ./... -count=1`
  - `go run ./cmd/transcribe live --help`

### Why

The repo needed a Phase 1 path that simulates real-time transcription without dragging in an irrelevant source mechanism. WAV replay is the simplest source that still exercises:

- chunk timing,
- overlap handling,
- transport round trips,
- transcript accumulation,
- user-visible incremental output.

### What worked

- The live CLI now exposes replay-oriented controls rather than the earlier placeholder chunk-directory idea.
- The replay source and accumulator both have unit tests.
- The runner is now functional enough to drive the chunk API end to end once given a real WAV input and an available Dagger environment.
- The full Go test suite passed after the implementation landed.

### What didn't work

- The first formatting pass failed because `internal/live/console_sink.go` had an extra closing brace.
- Exact command:

```bash
gofmt -w cmd/transcribe/live.go internal/live/*.go
```

- Exact error:

```text
internal/live/console_sink.go:31:1: expected declaration, found '}'
```

- I fixed the stray brace, reran formatting, and then reran the full test suite successfully.

### What I learned

- The replay source is already a better conceptual bridge to the future WebSocket sender loop than a chunk-directory source would have been.
- Even a simple Phase 1 accumulator needs explicit monotonicity and duplicate-suppression rules or the overlap path becomes noisy immediately.
- Carrying an `EmittedAt` timestamp on replay chunks is useful even before a full metrics subsystem exists, because it gives a straightforward end-to-end latency measurement point.

### What was tricky to build

The trickiest part was deciding what the replay source should emit. For the future streaming path, raw PCM frames would be the cleanest transport payload, but the current Phase 1 chunk API accepts WAV uploads. I resolved that by making the replay source emit temporary WAV chunk files now while keeping the chunk metadata transport-neutral enough (`SessionID`, `Sequence`, `Start`, `Duration`, `EmittedAt`) that the same source model can later feed a PCM/WebSocket sender abstraction.

The second tricky part was overlap handling in the accumulator. Phase 1 is still based on chunk-local final outputs, so the dedupe logic is necessarily heuristic. I kept it deliberately simple and explicit: reject empty words, reject words that end before the last committed finalization time, and reject near-identical words with very close timing.

### What warrants a second pair of eyes

- Whether the replay source should eventually emit PCM frames directly and have the chunk API path wrap them only as an adapter
- Whether the current duplicate-suppression tolerances are too aggressive or too conservative on real transcripts with overlap
- Whether `transcribe live` should start supporting non-console sinks in the next phase, or keep those deferred until transcript-state semantics are richer

### What should be done in the future

- Run an actual end-to-end `transcribe live` replay against a real WAV input in `tmux`
- Add a replay-based validation fixture or scripted harness for deterministic comparisons
- Evolve the runner/transport abstraction so the replay source can later feed the WebSocket client directly

### Code review instructions

Start here:

- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/replay_source.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/runner.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/accumulator.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/cmd/transcribe/live.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/replay_source_test.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/accumulator_test.go`

Validation commands:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go

gofmt -w cmd/transcribe/live.go internal/live/*.go
go test ./... -count=1
go run ./cmd/transcribe live --help
```

### Technical details

The replay source currently works as:

```text
input wav
  -> pure-Go conversion to 16k mono wav
  -> replay source slices fixed-duration overlapping chunks
  -> each chunk written as temp wav
  -> chunk uploaded to POST /transcribe/chunk
  -> chunk words fed to Phase 1 accumulator
  -> newly committed words printed to console
```

Current live-specific flags:

- `--input`
- `--chunk-duration`
- `--overlap-seconds`
- `--replay-speed`
- `--session-id`
- `--live-format`

## Step 6: Fix Dagger/Python startup churn and validate the live replay path in tmux

The next real issue showed up only once I tried to run the new live path against a real WAV in tmux. The first replay attempt did not fail in transcription logic; it got bogged down in a long `pip install -r requirements.txt` resolver cycle inside the Dagger container because the current graph made the Python dependency-install step depend on the whole server source directory, and the existing Lightning pin also appears to have been poorly aligned with NeMo’s own dependency range.

This step therefore had two parts: first, improve the Dagger graph and requirements pins so server startup becomes practical again after code changes; second, rerun the real replay in tmux and confirm that the system gets through health-check, starts processing chunk uploads, and emits incremental transcript output with timing metrics.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Keep advancing the implementation in practical, reviewable steps, and address real runtime blockers as they appear.

**Inferred user intent:** Move beyond static code changes into actual usable live-transcription behavior, and fix operational problems instead of leaving them as hand-waved future work.

**Commit (code):** 180dfdae073da2ba9469062b5ca0efcf45b7fbd4 — "Optimize ASR service dependency caching and pin Python deps"

### What I did

- Started a real live replay run in `tmux`:
  - session: `live-replay-e2e`
  - command shape:
    - `go run ./cmd/transcribe live -i .../audio-mix.wav --chunk-duration 5 --overlap-seconds 0.5 --replay-speed 0`
- Observed the first run spending a long time in `pip install -r requirements.txt` with resolver backtracking.
- Killed the blocked session and investigated the Dagger build graph.
- Updated `internal/server/dagger.go` so:
  - the `pip install -r requirements.txt` layer depends on a minimal bootstrap directory containing only `requirements.txt`,
  - the full mutable server source is mounted afterward.
- Updated `server/requirements.txt` to reduce resolver churn:
  - pinned `torchaudio==2.11.0` to match `torch==2.11.0`
  - changed `pytorch-lightning==2.6.1` to `pytorch-lightning==2.4.0`
  - added `lightning==2.4.0`
- Reran the real tmux replay after the fix.
- Captured successful runtime evidence showing:
  - health-check success,
  - live chunk POSTs succeeding,
  - committed transcript output appearing chunk by chunk,
  - end-to-end chunk timings being logged.

### Why

A live path that only works after an unpredictable multi-minute dependency backtracking cycle is not usable for iteration. The runtime blocker needed to be fixed before the live runner could be meaningfully validated.

### What worked

- The revised Dagger graph now isolates the `pip install` cache behind `requirements.txt` instead of invalidating it on every server code change.
- Aligning the Lightning pins with NeMo’s `<=2.4.0` range eliminated the worst resolver thrash during startup.
- The retried tmux run succeeded in reaching real live replay processing.
- Captured runtime evidence from the successful retry includes:

```text
2026/04/13 17:25:06 Tunnel established at 127.0.0.1:43479, waiting for health check...
2026/04/13 17:25:06 Health check passed (attempt 1)
2026/04/13 17:25:06 ASR server ready at 127.0.0.1:43479
2026/04/13 17:25:07 Committed +7 words: Welcome back to the Go Golems lab.
2026/04/13 17:25:07 Processed live chunk seq=0 start=0.000s duration=5.000s server_words=8 committed_total=7 committed_added=7 processing_ms=1207 end_to_end=1.218s
```

and later:

```text
2026/04/13 17:26:00 Processed live chunk seq=18 start=81.000s duration=5.000s server_words=17 committed_total=201 committed_added=17 processing_ms=1294 end_to_end=2.677s
```

### What didn't work

- The first tmux replay run hit long pip resolver backtracking instead of reaching inference quickly.
- The relevant commands/log context were:

```bash
tmux new-session -d -s live-replay-e2e "cd /home/manuel/code/wesen/2026-04-13--transcription-go && go run ./cmd/transcribe live -i /home/manuel/code/wesen/2026-04-09--screencast-studio/recordings/rabbit-hole-2026-04-10--2/audio-mix.wav --chunk-duration 5 --overlap-seconds 0.5 --replay-speed 0 2>&1 | tee logs/live-replay-e2e-20260413-171638.log"
```

with characteristic log lines like:

```text
INFO: pip is looking at multiple versions of lightning to determine which version is compatible with other requirements.
INFO: This is taking longer than usual. You might need to provide the dependency resolver with stricter constraints to reduce runtime.
```

### What I learned

- The Dagger graph shape matters a lot for Python-heavy services: mounting the full source directory before dependency installation made iterative startup much more expensive than necessary.
- The previous `pytorch-lightning==2.6.1` pin was a poor fit with NeMo’s transitive `lightning<=2.4.0,>2.2.1` expectation and contributed to resolver churn.
- The Phase 1 live path is now real enough to expose transcript-quality issues and chunk-latency behavior under actual runtime conditions, which is exactly where we want to be at this stage.

### What was tricky to build

The tricky part was recognizing that the apparent “live replay is slow” problem was not actually about the replay logic or the chunk API. It was an operational cache/dependency issue at the container-build boundary. That distinction matters because fixing the wrong layer would have led to wasted work in the runner while the true bottleneck remained in Python environment setup.

### What warrants a second pair of eyes

- Whether the new Lightning pins (`2.4.0`) are the best long-term constraint set or just the first stable one
- Whether the NeMo VCS requirement should remain as-is or be replaced with an explicit pinned wheel/version once confidence is higher
- Whether the repeated NeMo/Lhotse warnings during chunk transcription indicate a future tuning opportunity for lower-latency live settings

### What should be done in the future

- Let the current tmux replay continue and inspect later transcript quality/throughput deeper into the file
- Record final replay metrics once the run completes or enough representative chunks have been processed
- Consider a follow-up tuning pass on NeMo/Lhotse runtime warnings if they materially affect throughput

### Code review instructions

Start here:

- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/server/dagger.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/server/requirements.txt`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/runner.go`

Validation commands:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go

gofmt -w internal/server/dagger.go
go test ./... -count=1

# inspect the active replay session
tmux capture-pane -pt live-replay-e2e:0 | tail -120
```

### Technical details

Active successful replay validation context:

- `tmux` session: `live-replay-e2e`
- log file: `logs/live-replay-e2e-20260413-172135-pins.log`
- replay command:

```bash
go run ./cmd/transcribe live \
  -i /home/manuel/code/wesen/2026-04-09--screencast-studio/recordings/rabbit-hole-2026-04-10--2/audio-mix.wav \
  --chunk-duration 5 \
  --overlap-seconds 0.5 \
  --replay-speed 0
```

### Clarification: why the resolver/backtracking blocker appeared now even though earlier runs worked

This was most likely a **latent dependency-resolution issue that had been masked by cache**, not a brand-new runtime regression in the transcription code itself.

The key mechanics were:

1. **Earlier successful runs had a warm Dagger graph**
   - The previous container graph mounted the full `server/` directory before `pip install -r requirements.txt`.
   - As long as that source snapshot stayed stable enough, Dagger could reuse the already-built layer and startup looked fine.

2. **The pip cache is not the same as a fully cached dependency-resolution result**
   - `/root/.cache/pip` helps with downloaded wheels and artifacts.
   - It does **not** eliminate the need for pip to re-run dependency resolution if the install layer is invalidated.

3. **Live-transcription work changed the Python-side server files**
   - We changed `server/server.py` to add `/transcribe/chunk` and related helpers.
   - Because the old Dagger graph made `pip install` depend on the whole `server/` directory, those code changes invalidated the dependency-install layer even though `requirements.txt` had not meaningfully changed yet.

4. **Once pip had to resolve again, a real version-tension surfaced**
   - We had `pytorch-lightning==2.6.1` pinned.
   - NeMo was pulling a `lightning<=2.4.0,>2.2.1` constraint transitively.
   - `torchaudio` was also unpinned while `torch==2.11.0` was pinned.
   - That combination was enough to trigger noticeable resolver backtracking.

So the practical answer is:

> it worked before because cache was hiding the problem, and it became visible now because recent Python/server edits forced pip to resolve dependencies again.

This is also why the fix had two parts instead of one:

- **graph fix** — make the Dagger `pip install` layer depend on `requirements.txt`, not the entire mutable server tree
- **pin fix** — align the Python dependency versions so pip has a simpler, less contradictory solve to perform when it really does need to run again

## Step 7: Add rolling live output sinks and output-directory support

With the replay runner and chunk loop working, the next weakness was obvious: the live path could only print transcript updates to the console. That was useful for initial debugging, but not enough for comparing near-live behavior with the existing batch pipeline. To make the live mode reviewable and measurable, I added output sinks so committed transcript state can now be written to the same family of artifacts the batch mode already produces.

This step does not yet introduce richer partial/final semantics; it still works from the Phase 1 committed transcript state. But it makes the live path much more practical because it can now leave behind SRT/VTT/TXT/SQLite artifacts under a configurable output directory.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Continue iterating on the live-transcription implementation by making the current replay-driven pipeline more usable and easier to validate.

**Inferred user intent:** Move from “it prints some live words” toward a near-live mode that produces durable artifacts suitable for comparison, inspection, and future testing.

**Commit (code):** 0fc41bf9f2a2664eadbaaa1aed66e6a8f30ffa53 — "Add live transcript output sinks and output-dir support"

### What I did

- Added `--output-dir` to `transcribe live` and defaulted it to `./out-live`.
- Added `internal/live/sinks.go` to build live sinks from the requested `--live-format` list.
- Updated `internal/live/console_sink.go` so it implements the shared sink interface.
- Added `internal/live/subtitle_sink.go`:
  - rewrites `transcript.srt`, `transcript.vtt`, or `transcript.txt` from committed words.
- Added `internal/live/sqlite_sink.go`:
  - rewrites `transcript.db` from committed words.
- Updated `internal/live/runner.go` so after each chunk it now:
  - updates the accumulator,
  - pushes committed transcript state through all requested sinks.
- Added tests:
  - `internal/live/subtitle_sink_test.go`
  - `internal/live/sqlite_sink_test.go`
- Ran:
  - `gofmt -w cmd/transcribe/live.go internal/live/*.go`
  - `go test ./... -count=1`
  - `go run ./cmd/transcribe live --help`

### Why

The live path needs artifacts, not just logs. Without persisted outputs, it is difficult to:

- compare live vs batch behavior,
- inspect transcript quality later,
- integrate future validation scripts,
- reuse the current work for ticket evidence.

### What worked

- The live CLI now has an explicit output directory.
- The live runner can persist the committed transcript state to the same kinds of formats the batch path uses.
- The sink tests passed, and the full Go test suite still passed.
- The sink structure also creates a cleaner extension point for future partial/final-aware output behavior.

### What didn't work

- N/A as a build/test failure in this step.
- One implementation tradeoff is deliberate: both subtitle and SQLite sinks currently rewrite full artifacts from committed state, rather than performing a more incremental append/update strategy.

### What I learned

- Even before Phase 2 transcript semantics are implemented, the sink abstraction already improves the architecture because it separates “how transcript state is accumulated” from “how transcript state is rendered/persisted”.
- Rewriting whole artifacts per chunk is acceptable for Phase 1 and much simpler than trying to optimize persistence too early.

### What was tricky to build

The tricky design question was whether to make the sinks truly incremental or to regenerate outputs from the committed transcript state after each chunk. For Phase 1, regeneration is the safer choice because it avoids subtle state drift between the accumulator and the rendered artifacts. Once richer partial/final semantics exist, more efficient incremental update strategies can be considered.

### What warrants a second pair of eyes

- Whether rewriting the full SQLite database after each chunk is acceptable for the expected Phase 1 workload
- Whether the live subtitle segmentation should later diverge from the batch segmentation rules once partial/final semantics are introduced
- Whether `--live-format` should continue to accept `console` mixed with file outputs, or whether a more explicit sink configuration would be clearer long term

### What should be done in the future

- Validate the new sink behavior in a real `tmux` live replay run using `--output-dir`
- Add a replay-comparison workflow that inspects the generated `transcript.db` and subtitle/text artifacts against a batch baseline
- Introduce latency metrics and summary output next so live runs are easier to characterize quantitatively

### Code review instructions

Start here:

- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/sinks.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/subtitle_sink.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/sqlite_sink.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/runner.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/cmd/transcribe/live.go`

Validation commands:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go

gofmt -w cmd/transcribe/live.go internal/live/*.go
go test ./... -count=1
go run ./cmd/transcribe live --help
```

### Technical details

Current supported live output formats:

- `console`
- `srt`
- `vtt`
- `txt`
- `db`

Current behavior:

- sinks are driven from committed words only
- subtitle/text sinks rebuild transcript artifacts from committed words
- the SQLite sink rebuilds `transcript.db` from committed words
- all file outputs are written under `--output-dir`

## Step 8: Add latency/throughput metrics and a live summary artifact

Once the live runner could produce durable transcript artifacts, the next missing piece was quantitative measurement. Console logs showed per-chunk timings, but there was no structured summary of how a whole replay run performed. That made it harder to compare runs, spot regressions, or include concise evidence in the ticket. To fix that, I added a small metrics collector and a `live-summary.json` artifact written into the live output directory at the end of a run.

This is still Phase 1 telemetry rather than a full metrics subsystem, but it closes an important gap: the near-live runner now produces both transcript artifacts and a machine-readable performance summary.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Keep iterating on the live implementation so it becomes easier to validate and compare, not just easier to run.

**Inferred user intent:** Have the live path emit enough structured evidence that performance and behavior can be reviewed later without re-reading raw logs.

**Commit (code):** 6560072a7d7c567ba3e0408d9a0a99494a21eeaf — "Add live replay metrics and summary artifact"

### What I did

- Added `internal/live/metrics.go` with:
  - `MetricsCollector`
  - `MetricsSummary`
  - `WriteMetricsSummary(...)`
- Added `internal/live/metrics_test.go`.
- Updated `internal/live/runner.go` so it now:
  - tracks chunk-level server processing time,
  - tracks chunk-level end-to-end latency,
  - tracks effective audio coverage,
  - writes `live-summary.json` at the end of the run,
  - logs average server latency, average end-to-end latency, and throughput in audio-seconds-per-wall-second.
- Ran:
  - `gofmt -w internal/live/*.go`
  - `go test ./... -count=1`

### Why

Phase 1 live replay needed a structured summary so we can answer basic questions like:

- how many chunks were processed,
- how many words were committed,
- what the average chunk server latency was,
- what the average end-to-end latency was,
- how fast the system was relative to real time.

Without that, each run would require manual log inspection.

### What worked

- The runner now emits a summary artifact suitable for future comparison scripts or ticket evidence.
- The metrics structure is simple and testable.
- The full Go test suite still passed after the runner changes.

### What didn't work

- N/A as a compile/test issue in this step.
- One deliberate simplification remains: the throughput estimate uses effective audio coverage (`max chunk end`) rather than attempting to model overlap-adjusted semantic progress more deeply.

### What I learned

- Even a small summary artifact improves the usefulness of long-running replay tests a lot.
- Capturing metrics at the runner layer is a good Phase 1 choice because it is transport-adjacent without being coupled to the Python server internals.

### What was tricky to build

The main design choice was deciding what “throughput” should mean in a replay system with overlapping chunks. Summing raw chunk durations would overcount overlap, so the metrics collector instead tracks the maximum covered audio time (`chunk.Start + chunk.Duration`) across the session. That gives a better approximation of effective progress through the source audio.

### What warrants a second pair of eyes

- Whether the current throughput metric is the right default for later performance dashboards
- Whether chunk-level metrics should eventually be persisted in a more detailed per-chunk JSONL or SQLite metrics table
- Whether summary emission should happen only at the end of a run or also periodically during long replays

### What should be done in the future

- Run a fresh live replay with the new output sinks and metrics enabled together
- Use `live-summary.json` and `transcript.db` as the basis for the first near-live vs batch comparison notes
- Consider adding per-chunk metrics output if detailed offline analysis becomes necessary

### Code review instructions

Start here:

- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/metrics.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/runner.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/metrics_test.go`

Validation commands:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go

gofmt -w internal/live/*.go
go test ./... -count=1
```

### Technical details

The current `live-summary.json` contains fields for:

- `session_id`
- `chunks_processed`
- `committed_words`
- `effective_audio_seconds`
- `elapsed_seconds`
- `audio_seconds_per_wall_second`
- `average_server_processing_ms`
- `max_server_processing_ms`
- `average_end_to_end_ms`
- `max_end_to_end_ms`

## Step 9: Validate persisted live outputs in tmux and make the metrics summary update incrementally

After the live sinks and summary artifact were implemented, I launched a fresh `tmux` replay run that used the new `--output-dir` and persisted-output formats. The main goal of this step was to confirm that the replay runner was not only transcribing chunk by chunk, but also leaving behind useful artifacts while the run was still in progress.

That validation succeeded: while the run was active, `transcript.srt` and `transcript.db` were already present under `out-live-e2e/`, and their contents were updating as new chunks were processed. While inspecting that, I noticed one remaining observability gap: `live-summary.json` would only be written at the end of the run, which is inconvenient for long sessions. I fixed that by refreshing the summary file after each chunk as well.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Continue turning the live path into something practically inspectable and validate it with real runs, not just unit tests.

**Inferred user intent:** See the near-live system produce durable artifacts and make long-running replay sessions easier to observe while they are still executing.

**Commit (code):** 3ff211a09fdc4d2bb2f9d557d153fd8138cc0c17 — "Update live summary artifact incrementally during replay"

### What I did

- Started a new tmux replay session using the new persisted-output path:

```bash
go run ./cmd/transcribe live \
  -i /home/manuel/code/wesen/2026-04-09--screencast-studio/recordings/rabbit-hole-2026-04-10--2/audio-mix.wav \
  -o ./out-live-e2e \
  --live-format console,srt,db \
  --chunk-duration 5 \
  --overlap-seconds 0.5 \
  --replay-speed 0
```

- Verified during the run that:
  - `out-live-e2e/transcript.db` existed and was a valid SQLite database,
  - `out-live-e2e/transcript.srt` existed and contained current transcript segments.
- Captured concrete runtime evidence showing chunk progress and output persistence.
- Updated `internal/live/runner.go` so `live-summary.json` is refreshed after each processed chunk rather than only at final completion.
- Ran:
  - `gofmt -w internal/live/runner.go`
  - `go test ./... -count=1`

### Why

For long replays, end-of-run artifacts are not enough. The system should be inspectable while it is still running, especially when using `tmux` for iterative validation.

### What worked

- The persisted-output replay created live artifacts successfully.
- At inspection time, `out-live-e2e/` contained:
  - `transcript.db`
  - `transcript.srt`
- Sample SRT content looked correct structurally, for example:

```text
1
00:00:02,240 --> 00:00:05,120
Welcome back to the Go Golems lab.
```

- SQLite validation showed the file was a real database while the run was in progress:

```text
out-live-e2e/transcript.db: SQLite 3.x database ...
```

- The running session was still processing chunks successfully, e.g.:

```text
2026/04/13 17:41:36 Processed live chunk seq=22 start=99.000s duration=5.000s server_words=17 committed_total=260 committed_added=14 processing_ms=1983 end_to_end=4.022s
```

### What didn't work

- The first version of the metrics-summary behavior only wrote `live-summary.json` when the run finished.
- That was not ideal for a long-running tmux workflow because it meant the summary artifact lagged behind the actual state of the live session.
- I fixed that by writing the summary after each chunk update in addition to the final completion write.

### What I learned

- The live path is now far enough along that runtime validation produces genuinely useful artifacts mid-run, not just logs.
- Incremental artifact updates matter for operator experience almost as much as end-of-run correctness when replay sessions are long.

### What was tricky to build

The tricky part here was not the file writing itself, but deciding what “good enough observability” means in a long replay. A summary file that only appears at the end is technically correct but operationally unsatisfying. Updating it incrementally makes the system much easier to inspect without adding much complexity.

### What warrants a second pair of eyes

- Whether `live-summary.json` should include per-chunk details in a future JSONL or separate metrics file
- Whether rewriting the summary after every chunk is sufficient, or whether a periodic heartbeat/status file would be cleaner for very long runs
- Whether additional live formats (`txt`, `vtt`) should be included in the next real tmux validation run

### What should be done in the future

- Let the persisted-output tmux replay continue and inspect the resulting output directory later in the run or at completion
- Use the generated `transcript.db` and `live-summary.json` as the basis for the first batch-vs-live comparison notes
- Consider adding a small comparison script to the ticket workspace for transcript database deltas

### Code review instructions

Start here:

- `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/runner.go`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/out-live-e2e/transcript.srt`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/out-live-e2e/transcript.db`

Validation commands:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go

gofmt -w internal/live/runner.go
go test ./... -count=1
ls -lah out-live-e2e
file out-live-e2e/transcript.db
sed -n '1,40p' out-live-e2e/transcript.srt
```

### Technical details

Validation context:

- `tmux` session: `live-replay-output-e2e`
- log file: `logs/live-replay-output-e2e-20260413-173731.log`
- output dir: `out-live-e2e/`

Observed files during replay:

- `out-live-e2e/transcript.db`
- `out-live-e2e/transcript.srt`

Follow-up code change in this step:

- `internal/live/runner.go` now writes `live-summary.json` after each processed chunk.

## Step 10: Add a transcript-database comparison script and run the first partial live-vs-reference check

Now that the live replay path can persist a rolling `transcript.db`, the next practical need was a repeatable way to compare it with the reference transcript database from the earlier batch pipeline work. Rather than doing that by hand in `sqlite3`, I added a small comparison script to the ticket workspace and used it against the in-progress `out-live-e2e/transcript.db`.

Because the live replay was still running, this first comparison is intentionally partial. That is still useful: it verifies that the live output database is structurally comparable, confirms the script works, and gives a grounded snapshot of how far through the source audio the replay had progressed at inspection time.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Continue making the live-transcription work reviewable and measurable, not just runnable.

**Inferred user intent:** Have tooling around the live pipeline that makes it easier to inspect progress and compare outputs against known references.

### What I did

- Added a ticket-local comparison script:
  - `ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/01-compare_transcript_dbs.py`
- The script:
  - opens both SQLite transcript DBs,
  - reports word/filler/chunk counts,
  - reports time coverage,
  - reports first/last words,
  - optionally attaches `live-summary.json` if present.
- Ran it against:
  - live DB: `out-live-e2e/transcript.db`
  - reference DB: `/home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/TRANSCRIPT-PIPELINE--setting-up-an-analysis-pipeline-for-transcripts/sources/audio_transcript.db`

### Why

A comparison script is the fastest way to turn the current live run into concrete evidence rather than anecdotal observation. It also creates a reusable tool for the next full near-live vs batch comparison once the replay completes.

### What worked

- The script ran successfully against both databases.
- The in-progress live replay DB was valid and queryable.
- The partial comparison showed that the live run had progressed meaningfully through the source audio:

```json
{
  "live": {
    "word_count": 1152,
    "chunk_count": 106,
    "max_end": 509.12
  },
  "reference": {
    "word_count": 4248,
    "chunk_count": 196,
    "max_end": 1656.0
  }
}
```

This is exactly what we would expect from a replay that is still in progress rather than complete.

### What didn't work

- My first `chmod`/execution attempt used the wrong relative path after writing the script, so the shell could not find it.
- The follow-up locate-and-run command fixed that immediately.

### What I learned

- Even a very small comparison script is enough to turn a long-running replay into something quantitatively inspectable mid-run.
- The current live DB schema is compatible enough with the reference DB to support direct structural comparison already.

### What was tricky to build

The main subtlety was making sure the script’s output is useful even for a partial live run. Reporting raw deltas alone would be misleading mid-run, so the script also reports effective coverage (`max_end`) and first/last words to give context about how far the replay has progressed.

### What warrants a second pair of eyes

- Whether the comparison script should later report more transcript-quality signals, such as punctuation differences, average word durations, or chunk density
- Whether the eventual full comparison should be emitted as JSON only, Markdown only, or both

### What should be done in the future

- Re-run the script when the live replay completes and include `live-summary.json`
- Add a follow-up note in the ticket with the completed live-vs-reference comparison
- Optionally expand the script into a small playbook or report generator once the final comparison format is clearer

### Code review instructions

Start here:

- `/home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/01-compare_transcript_dbs.py`

Validation command:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go
python3 ./ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/01-compare_transcript_dbs.py \
  --live-db out-live-e2e/transcript.db \
  --reference-db /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/TRANSCRIPT-PIPELINE--setting-up-an-analysis-pipeline-for-transcripts/sources/audio_transcript.db
```

### Technical details

Observed partial comparison snapshot:

- live word count: `1152`
- live chunk count: `106`
- live coverage: `509.12s`
- reference word count: `4248`
- reference chunk count: `196`
- reference coverage: `1656.0s`
- delta word count: `-3096`
- delta coverage: `-1146.88s`

Interpretation: the live replay DB was valid and structurally comparable, but the run was still partial at the time of inspection.

## Step 11: Record the completed full-run replay results and establish a fast 120s subset workflow

At this point the long live replay run has completed, so the ticket finally has a real end-to-end result rather than an in-progress snapshot. I recorded the final full-run metrics and compared the completed `out-live-e2e/transcript.db` against the earlier reference transcript database. Then, following the user’s suggestion to iterate on smaller subsets for speed, I added a tiny PCM-WAV slicing helper and used it to create a 120-second clip for a much faster live-vs-batch comparison loop.

The important outcome is that the short-clip workflow already gives a clearer debugging target. The full run finished with a larger gap (`-403` words vs the older reference DB), while the 120s same-pipeline comparison only showed a `-21` word delta (`302` live vs `323` batch). That strongly suggests the smaller-subset workflow is the right way to investigate the remaining live-path loss and overlap/finalization behavior.

### Prompt Context

**User prompt (verbatim):** "ok, do it. And you can maybe use smaller subsets of the wav so that we can be faster in our iteration."

**Assistant interpretation:** Record the completed full-run results in the ticket and adopt a shorter clipped-WAV workflow so future live-vs-batch investigations iterate much faster.

**Inferred user intent:** Preserve the current results, but shift the debugging loop onto smaller reproducible clips so we can improve the live path without repeatedly paying the cost of a full 27-minute run.

**Commit (code):** f130b2488c2956413a8619015a876f195eaa5e4a — "Tools: add WAV segment extraction helper for fast iteration"

### What I did

- Confirmed the persisted full live replay had completed:
  - `out-live-e2e/transcript.db`
  - `out-live-e2e/transcript.srt`
  - `out-live-e2e/live-summary.json`
- Ran the completed live-vs-reference DB comparison:
  - live words: `3845`
  - reference words: `4248`
  - delta: `-403`
  - live coverage: `1656.14s`
  - reference coverage: `1656.0s`
- Added a ticket-local WAV slicing helper:
  - `ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/02-extract_wav_segment.py`
- Used it to create a 120s clip:
  - `/tmp/transcription-live-clip-000-120.wav`
- Ran a short batch baseline on that clip:
  - output dir: `out-batch-clip-000-120`
  - final word count: `323`
- Ran a short live replay on that same clip:
  - output dir: `out-live-clip-000-120`
  - final committed word count: `302`
  - `live-summary.json` throughput: `3.714x`
- Compared the short live clip against the short batch clip:
  - live words: `302`
  - batch words: `323`
  - delta: `-21`

### Why

The full run is necessary as a milestone, but it is too expensive for the main debugging loop. A small deterministic clip is a much better target for overlap/finalization investigation because:

- startup cost dominates less,
- turnaround is short,
- comparisons are still meaningful,
- regressions are easier to isolate.

### What worked

- The full run completed successfully and produced final metrics/artifacts.
- The 120s slicing helper worked immediately and produced a valid 22MB WAV clip.
- The short batch run completed quickly and gave a clean same-pipeline reference (`323` words).
- The short live run also completed quickly and produced a much smaller delta (`-21`) than the full-run comparison.

### What didn't work

- The first shell command after writing the slicing helper repeated an earlier mistake: I used the wrong relative path when trying to `chmod` and execute the script.
- I fixed that immediately by locating the file and rerunning the command with the correct repo-relative path.

### What I learned

- The long-form live replay is now good enough to produce a full transcript, but it is not the right vehicle for day-to-day debugging.
- The 120s same-pipeline comparison is much more actionable than the full-run comparison against the older reference DB because it isolates the current live-vs-batch behavior under the same codebase and model path.
- The smaller delta on the 120s clip suggests the remaining gap is likely tied to chunk-boundary/accumulator behavior rather than a catastrophic failure of the overall live architecture.

### What was tricky to build

The main subtlety here was choosing the right comparison baseline. The earlier full reference DB is still useful, but it is not the cleanest debugging target because it comes from an older pipeline context. The short clipped-WAV workflow is better because both sides of the comparison are generated by the current codebase under controlled conditions. That makes the resulting delta much more interpretable.

### What warrants a second pair of eyes

- Whether the 120s opening clip is the best representative debugging segment, or whether a later/more difficult portion of the recording would expose overlap issues more clearly
- Whether the current `5s` chunk duration is the right default for investigation, or whether we should compare `5s` vs `10s` chunk behavior next
- Whether the `-21` delta is mostly due to dedupe heuristics, chunk overlap, or server-side chunk segmentation behavior

### What should be done in the future

- Keep the 120s subset workflow as the default fast iteration loop
- Add one or two additional clipped segments from different parts of the recording if needed
- Investigate the `-21` subset delta next before doing more long full-run experiments

### Code review instructions

Start here:

- `/home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/02-extract_wav_segment.py`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/out-batch-clip-000-120/transcript.db`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/out-live-clip-000-120/transcript.db`
- `/home/manuel/code/wesen/2026-04-13--transcription-go/out-live-e2e/live-summary.json`

Validation commands:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go

python3 ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/02-extract_wav_segment.py \
  --input /home/manuel/code/wesen/2026-04-09--screencast-studio/recordings/rabbit-hole-2026-04-10--2/audio-mix.wav \
  --output /tmp/transcription-live-clip-000-120.wav \
  --start 0 \
  --duration 120

go run ./cmd/transcribe batch \
  -i /tmp/transcription-live-clip-000-120.wav \
  -o ./out-batch-clip-000-120 \
  -f db,txt \
  --chunk-size 60

go run ./cmd/transcribe live \
  -i /tmp/transcription-live-clip-000-120.wav \
  -o ./out-live-clip-000-120 \
  --live-format console,db,txt \
  --chunk-duration 5 \
  --overlap-seconds 0.5 \
  --replay-speed 0

python3 ./ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/01-compare_transcript_dbs.py \
  --live-db out-live-clip-000-120/transcript.db \
  --reference-db out-batch-clip-000-120/transcript.db \
  --summary-json out-live-clip-000-120/live-summary.json
```

### Technical details

Completed full-run result:

- live words: `3845`
- reference words: `4248`
- delta: `-403`
- live chunks: `352`
- reference chunks: `196`
- coverage delta: `+0.14s`
- throughput: `2.64x`

Fast 120s subset result:

- batch words: `323`
- live words: `302`
- delta: `-21`
- live chunks processed: `27`
- live throughput: `3.714x`
- average live server processing: `1166.815ms`
- average live end-to-end latency: `2353.444ms`


## Step 12: Inspect the actual live-vs-batch words directly and write a ticket report

After the fast 120-second subset workflow was in place, the next useful step was to stop reasoning abstractly about counts and instead inspect the actual transcript words in order. I exported the words from the current live and batch transcript databases, diffed them directly, and wrote a dedicated report in the ticket describing what is genuinely missing, what is substituted, and which apparent “losses” are really just removal of duplicated junk.

This step intentionally did not try to “fix” the live path yet. The purpose was to reduce ambiguity before more implementation work. The report now gives a concrete baseline for later Phase 2 and WebSocket work: some of the delta is real word loss around boundaries, some is ASR substitution noise, and some is batch-side duplication that should not necessarily be preserved.

### Prompt Context

**User prompt (verbatim):** "ok, look at the words then dude, just export them all and look at them, it's not that much. no need to be clever, just look at what's in the db right now."

**Assistant interpretation:** Export the ordered transcript words from the current live and batch SQLite databases, inspect the concrete mismatches directly, and write the findings into the ticket.

**Inferred user intent:** Stop over-abstracting the comparison and create a grounded, evidence-based explanation of the current live-vs-batch delta.

### What I did
- Exported the ordered words from:
  - `out-live-clip-000-120/transcript.db`
  - `out-batch-clip-000-120/transcript.db`
- Inspected the actual live and batch word sequences side-by-side instead of relying only on aggregate counts.
- Added a ticket-local helper script:
  - `ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/03-word_diff_report.py`
- Renamed the existing ticket helper scripts to numbered names for consistency and traceability:
  - `scripts/01-compare_transcript_dbs.py`
  - `scripts/02-extract_wav_segment.py`
  - `scripts/03-word_diff_report.py`
- Wrote a dedicated report:
  - `ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/reference/03-word-analysis-report.md`
- Updated the ticket task checklist so Phase 1 comparison work points at the new report.

### Why
- The `-21` word delta on the 120-second same-pipeline comparison was already small enough to inspect directly.
- We needed to distinguish three different cases that raw counts blur together:
  - true missing words,
  - substitutions/corruptions,
  - batch-side duplicated junk that the live path happened to collapse.

### What worked
- Direct export of the `words` table was enough to understand the current gap without any heavier tooling.
- The dedicated diff script produced clear region-by-region evidence that is now reusable for later iterations.
- The ticket now contains a concrete report rather than only count-level notes.

### What didn't work
- A follow-up shell command initially repeated the earlier relative-path mistake when trying to `chmod` the new helper script.
- I corrected that by locating the file explicitly and rerunning the command with the correct repo-relative path.

### What I learned
- The current live-vs-batch difference on the 120-second subset is not a single bug.
- Some of the missing words are real connector-word losses around chunk boundaries.
- Some apparent losses are actually beneficial compression of duplicated junk that batch preserved.
- That means later work should optimize for transcript quality and explicit finalization semantics, not just count parity.

### What was tricky to build
- The main subtlety was interpreting sequence diffs responsibly. A lower live word count does not automatically mean the live output is worse, because the batch baseline itself contains duplicated junk in a few places.
- The report therefore had to separate recognition substitutions, boundary-related omissions, and beneficial duplicate collapse rather than flatten everything into one “missing words” bucket.

### What warrants a second pair of eyes
- Whether the mismatch classification in `reference/03-word-analysis-report.md` is the right framing for later tuning work
- Whether later comparison should use additional clipped segments from harder parts of the recording
- Whether the current chunk-size/overlap defaults are still the best proving setup before moving to the WebSocket path

### What should be done in the future
- Use the 120-second subset workflow as the default inspection loop for any remaining Phase 1 quality work
- Avoid over-investing in perfect overlap heuristics now that the long-term target is session-oriented WebSocket streaming
- Reuse the word-diff report/script when validating future state-model or transport changes

### Code review instructions
- Start with:
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/reference/03-word-analysis-report.md`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/03-word_diff_report.py`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/out-live-clip-000-120/transcript.db`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/out-batch-clip-000-120/transcript.db`
- Validate with:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go

python3 ./ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/03-word_diff_report.py   --batch-db out-batch-clip-000-120/transcript.db   --live-db out-live-clip-000-120/transcript.db
```

### Technical details
- 120-second subset counts:
  - batch words: `323`
  - live words: `302`
  - delta: `-21`
- Diff shape from normalized sequence comparison:
  - mismatch regions: `32`
  - deletes: `22`
  - inserts: `5`
  - replaces: `5`
- Main classes observed:
  - dropped connector/function words
  - local substitutions (`Golems` → `Columns`, `trust that` → `try`)
  - chunk-boundary phrase compression
  - occasional live-side removal of duplicated junk that batch preserved


## Step 13: Introduce an explicit partial/final transcript-state model in Go

With the word-analysis report in place, the next useful step was not to over-tune overlap heuristics but to move the Go live path closer to the long-term WebSocket architecture. I introduced an explicit transcript-state model that separates pending preview words from committed final words, so the current chunk runner and the future WebSocket transport can share the same accumulator semantics.

This step keeps the current Phase 1 replay path working, but it changes the internal shape of the live system in a way that matters later: the accumulator now consumes transport-neutral transcript events, the sinks consume transcript state snapshots, and the tests cover partial revisions, promotion to final, overlap filtering, and out-of-order event rejection.

### Prompt Context

**User prompt (verbatim):** "ok, let's continue with the normal plan, and yes let's not make this perfect. Continue"

**Assistant interpretation:** Continue the planned live-transcription implementation without over-optimizing the temporary Phase 1 overlap heuristics, and take the next slice that best supports the future WebSocket design.

**Inferred user intent:** Keep forward momentum toward the real streaming architecture instead of spending too much time polishing the proving-phase chunk logic.

**Commit (code):** 64bb79c — "Add explicit partial/final transcript state model"

### What I did
- Added `internal/live/types.go` with transport-neutral live types:
  - `TranscriptEvent`
  - `TranscriptEventType`
  - `TranscriptState`
- Refactored `internal/live/accumulator.go` to:
  - accept explicit `partial` vs `final_words` events,
  - track `Committed`, `Pending`, and `LastFinalTime`,
  - enforce monotonic finalization,
  - reject out-of-order numbered events,
  - keep the old `ApplyFinalWords(...)` convenience path for current callers.
- Refactored live sinks so they consume transcript state rather than raw committed slices:
  - `internal/live/sinks.go`
  - `internal/live/console_sink.go`
  - `internal/live/subtitle_sink.go`
  - `internal/live/sqlite_sink.go`
- Updated the live runner to feed `TranscriptEvent` values into the accumulator and to hand `TranscriptState` snapshots to sinks.
- Expanded tests in `internal/live/accumulator_test.go` to cover:
  - repeated partial revisions,
  - partial → final promotion,
  - overlap-induced duplicate filtering,
  - out-of-order event rejection.
- Updated sink tests so durable artifacts verify they use committed words only.
- Ran:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go
gofmt -w internal/live/types.go internal/live/accumulator.go internal/live/accumulator_test.go internal/live/sinks.go internal/live/console_sink.go internal/live/subtitle_sink.go internal/live/subtitle_sink_test.go internal/live/sqlite_sink.go internal/live/sqlite_sink_test.go internal/live/runner.go
go test ./internal/live/... -count=1
go test ./... -count=1
```

### Why
- The future WebSocket path will emit partial and final transcript events explicitly.
- It is better to align the Go-side accumulator model now than to spend more time perfecting temporary overlap heuristics.
- Durable sinks should operate on committed state only, even if console output later shows preview text.

### What worked
- The refactor stayed local to `internal/live/` and did not disturb the batch pipeline.
- The live package tests now cover the state transitions we expect the future WS client/receiver loop to exercise.
- The full repo test run passed after the refactor.

### What didn't work
- During the refactor I briefly introduced a compile issue in `ConsoleSink` by trying to over-generalize the word-joining helper; the concrete `[]output.Word` helper was simpler and correct.
- I also hit a package-level helper-name collision between test code and the non-test console sink helper; renaming the test helper resolved it cleanly.

### What I learned
- The right next abstraction boundary was not “better overlap math”, it was “make state explicit”.
- Sinks become easier to reason about once they always receive the full transcript snapshot and are responsible for using only the committed portion if they are durable.
- Sequence validation belongs in the accumulator-level state model even before WebSocket transport exists, because it clarifies how out-of-order live events should be handled.

### What was tricky to build
- The main subtlety was choosing sequence semantics that are useful now without overfitting to a future transport. I needed to allow repeated partial revisions for the same sequence and a later finalization of that sequence, while still rejecting older out-of-order events. The implemented rule is intentionally simple: repeated partials for the current sequence are allowed, a final event can finalize that sequence once, and older sequences are rejected.
- Another subtle point was preserving the current replay runner behavior while changing the sink interface from raw committed words to transcript-state snapshots. The runner now bridges the current chunk API into the new event model without changing the external CLI behavior.

### What warrants a second pair of eyes
- Whether the current sequence-handling rule is the right base semantics for the future WS receiver loop
- Whether `LastFinalTime` and duplicate timing tolerances are still the right monotonic/finalization guards once real streaming results arrive
- Whether console preview logging should stay minimal or become more structured once partials are emitted for real

### What should be done in the future
- Use this state model as the base for the planned WebSocket client/server event flow
- Keep Phase 1 chunk mode stable, but avoid spending too much effort on perfect overlap tuning
- Start Phase 3 by wiring actual WebSocket protocol/session code onto `TranscriptEvent` / `TranscriptState`

### Code review instructions
- Start with:
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/types.go`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/accumulator.go`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/runner.go`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/console_sink.go`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/subtitle_sink.go`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/sqlite_sink.go`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/accumulator_test.go`
- Validate with:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go
go test ./internal/live/... -count=1
go test ./... -count=1
```

### Technical details
- New event/state types:
  - `TranscriptEventPartial`
  - `TranscriptEventFinalWords`
  - `TranscriptEvent`
  - `TranscriptState`
- Current accumulator rules:
  - partial events replace pending preview state
  - final events append monotonic non-duplicate words to committed state
  - `LastFinalTime` advances monotonically
  - durable sinks use committed state only
  - out-of-order numbered events are rejected


## Step 14: Add the first server-side WebSocket/session scaffold

With the Go-side transcript-state model in place, the next normal step was to add the server-side session boundary instead of continuing to polish the temporary chunk overlap path. I implemented a first session-oriented WebSocket API in Python: a live session registry, a per-session buffered decoder object, and a `WS /transcribe/stream` endpoint that accepts `start` / `audio` / `flush` / `stop` and emits `started` / `partial` / `final_words` / `stopped` / `error`.

This is intentionally the simplest honest version of streaming on the server. The transport and session lifecycle are now real, but the actual decoder is still buffered under the hood: incoming PCM frames are accumulated per session, partials are derived from the current buffered span, and `flush`/`stop` finalize that buffered region. That matches the ticket plan: get the protocol/session model in place first, then improve the underlying incremental decoding behavior later.

### Prompt Context

**User prompt (verbatim):** "ok, go ahead."

**Assistant interpretation:** Continue with the next planned implementation slice after the Go-side state-model work, specifically the server-side WebSocket/session architecture.

**Inferred user intent:** Keep moving along the intended live-streaming roadmap without getting bogged down in temporary proving-phase polish.

**Commit (code):** e84dc41 — "Add server-side live sessions and websocket endpoint"

### What I did
- Added `server/live_decoder.py`:
  - per-session buffered PCM16 decoder scaffold,
  - `append_audio(...)`,
  - `decode_partial()`,
  - `flush()`,
  - `close()`.
- Added `server/live_sessions.py`:
  - `LiveSessionRegistry`,
  - `LiveSession`,
  - structured `LiveSessionError` codes,
  - idle cleanup,
  - stop/broken-connection/shutdown cleanup hooks.
- Updated `server/server.py` to:
  - expose `WS /transcribe/stream`,
  - accept `start`, `audio`, `flush`, and `stop` JSON events,
  - emit `started`, `partial`, `final_words`, `stopped`, and `error` events,
  - create and destroy sessions through the registry,
  - keep the existing HTTP batch/chunk endpoints unchanged.
- Added `server/live_sessions_test.py` with lightweight Python unit tests for:
  - decoder flush/base-offset behavior,
  - out-of-order audio sequence rejection,
  - session close/session-not-found behavior.
- Validated with:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go
python3 -m unittest discover -s server -p '*_test.py'
python3 -m py_compile server/server.py server/live_decoder.py server/live_sessions.py
go test ./... -count=1
```

### Why
- The Go side now understands explicit partial/final transcript state, so the next missing architectural piece was the server-side session boundary.
- The long-term target is a real WebSocket transport with session lifecycle, not perfect HTTP-chunk overlap handling.
- A buffered decoder behind a real session protocol is a better intermediate step than inventing more client-side overlap heuristics.

### What worked
- The new Python server modules stayed isolated and did not disturb the existing batch/chunk HTTP flow.
- The session registry/decoder split made the WebSocket endpoint logic straightforward to implement.
- Lightweight Python unit tests were enough to validate the initial session/decoder invariants without pulling the ASR model into the test loop.

### What didn't work
- N/A

### What I learned
- The right first server-side streaming milestone is not “perfect incremental ASR”, it is “real session lifecycle with explicit protocol events”.
- A buffered per-session decoder already gives a useful place to hang sequence validation, cleanup policy, and future decoder-state evolution.
- The existing `_transcribe_chunk_file(...)` helper is a good bridge for this first WS version because it keeps timestamp extraction and inference behavior consistent with the current server path.

### What was tricky to build
- The main subtlety was being honest about what is and is not streaming yet. The WebSocket/session transport is real, but the decoder still re-materializes buffered audio into a WAV and transcribes that buffered span. I had to preserve that distinction clearly in the code and docs so we do not confuse “WS transport exists” with “true incremental model-state streaming already exists”.
- Another subtlety was cleanup policy. The initial implementation needed to clean up sessions in four places: idle-timeout sweeps, explicit `stop`, broken connection, and application shutdown. That policy is simple enough for now but still prevents abandoned server-side session objects from accumulating.

### What warrants a second pair of eyes
- Whether the current `partial` emission policy (buffered decode after a minimum amount of audio) is the right temporary behavior before a real incremental decoder lands
- Whether the current server-side sequence validation is strict enough or should also validate `pts` continuity more aggressively
- Whether the environment-tunable cleanup knobs (`LIVE_SESSION_IDLE_TIMEOUT_SECONDS`, `LIVE_PARTIAL_MIN_SECONDS`) are named and scoped appropriately

### What should be done in the future
- Add the Go WebSocket client and sender/receiver loops on top of this server-side WS contract
- Make the live runner transport-agnostic so chunk and WS modes can coexist during transition
- Replace the buffered decoder internals with a more truly incremental decoder/session-state implementation later

### Code review instructions
- Start with:
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/server/live_sessions.py`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/server/live_decoder.py`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/server/server.py`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/server/live_sessions_test.py`
- Validate with:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go
python3 -m unittest discover -s server -p '*_test.py'
python3 -m py_compile server/server.py server/live_decoder.py server/live_sessions.py
go test ./... -count=1
```

### Technical details
- Implemented WS endpoint:
  - `WS /transcribe/stream`
- Implemented client → server events:
  - `start`
  - `audio`
  - `flush`
  - `stop`
- Implemented server → client events:
  - `started`
  - `partial`
  - `final_words`
  - `stopped`
  - `error`
- Current constraints:
  - `sample_rate=16000`
  - `channels=1`
  - `format="pcm_s16le"`
  - buffered preview/final decode inside each session rather than true model-state streaming


## Step 15: Add the Go WebSocket client path and make the live runner transport-aware

With the server-side WebSocket/session scaffold in place, the next step was to wire a corresponding Go client path rather than leaving the WS API unused. I added a Go WebSocket client, explicit sender/receiver loops, and transport selection in the live runner so the replay source can now drive either the old HTTP chunk transport or the new session-oriented WebSocket transport.

This first Go WS path is intentionally practical rather than idealized. It sends replay PCM16 bytes over the real WS protocol, but it currently flushes once per replay chunk and waits for that chunk’s finalization before sending the next one. That keeps the sequencing and metric attribution simple while the server-side decoder is still buffered under the hood.

### Prompt Context

**User prompt (verbatim):** "go ahead"

**Assistant interpretation:** Continue with the next planned Go-side streaming slice: add the WebSocket client, sender/receiver loops, and wire the live runner so it can use the new server-side WS transport.

**Inferred user intent:** Keep progressing along the planned streaming architecture so the existing server-side WS scaffold becomes usable end-to-end from the Go CLI.

**Commit (code):** cd945df — "Add Go WebSocket live transport and transport-aware runner"

### What I did
- Added `internal/live/wsclient.go`:
  - connect/disconnect lifecycle,
  - `start`, `audio`, `flush`, `stop`,
  - WS JSON message decoding.
- Added `internal/live/stream_sender.go`:
  - sends replay PCM16 chunks over WS,
  - flushes after each replay chunk for the current implementation,
  - waits for an ack signal before advancing to the next chunk.
- Added `internal/live/stream_receiver.go`:
  - converts WS server messages into transport-neutral `TranscriptEvent` values,
  - surfaces terminal `stopped` information.
- Updated `internal/live/replay_source.go` so each replay chunk now carries raw `PCM16` bytes in addition to a temporary WAV file path.
- Updated `internal/live/types.go` so `TranscriptEvent` also carries `ProcessingMS`.
- Refactored `internal/live/runner.go` to support:
  - `chunk` transport,
  - `ws` transport,
  - shared sink/state update flow,
  - shared metrics/output handling.
- Updated `cmd/transcribe/live.go` to add:
  - `--transport chunk|ws`
- Added `internal/live/wsclient_test.go` as a lightweight Go WebSocket integration test using `httptest` + Gorilla WebSocket.
- Added the Gorilla WebSocket dependency in `go.mod` / `go.sum`.

### Why
- The server-side WS protocol is only valuable if the Go CLI can actually speak it.
- Making the runner transport-aware now lets chunk and WS modes coexist cleanly while the system transitions.
- A chunk-paced WS sender is a good intermediate step: it exercises the real transport/session semantics without forcing premature complexity into sequencing/finalization logic.

### What worked
- The Go live package tests and the full repo Go test suite passed after the refactor.
- The WS client/receiver integration test proved the basic protocol flow in-process.
- A real 15-second replay-driven WS smoke run worked through the Dagger-hosted Python service and produced persisted output artifacts.

### What didn't work
- The first WS smoke run exposed an ordering problem: the sender could push the next chunk far enough ahead that a later `partial` arrived before the previous chunk’s `final_words`, which produced an accumulator sequence error.
- I fixed this by making the first WS sender wait for a per-chunk finalization ack before moving on to the next replay chunk.

### What I learned
- The first stable WS client path does not need to be fully continuous to be useful; a chunk-paced WS sender is already enough to validate protocol correctness and runner integration.
- The transport-agnostic runner boundary is now much clearer: replay source → transport → transcript events → accumulator → sinks.
- The WS path and chunk path can now share most of the post-transport logic, which is exactly the architecture we wanted.

### What was tricky to build
- The main subtlety was sequencing. The server emits both `partial` and `final_words`, and the Go accumulator now enforces meaningful ordering. If the sender outruns the rate at which finalization events are applied, the current simple sequence model gets confused. The initial fix is deliberately conservative: flush every replay chunk and block until its final event is observed.
- Another subtle point was keeping metrics comparable across transports. The runner now associates final WS events with the corresponding replay chunks via a small finalized-chunk queue so end-to-end latency and server processing time can still be recorded in a familiar way.

### What warrants a second pair of eyes
- Whether the current per-chunk WS flush+ack strategy is the right temporary stepping stone before moving to more continuous WS streaming
- Whether the current accumulator sequence semantics will need a broader revision once multiple partial revisions per logical chunk become more common
- Whether the current replay metrics are still the right shape for comparing chunk vs WS transport behavior

### What should be done in the future
- Use the new WS path for larger replay-based validation runs
- Compare WS-mode output quality and latency against the existing chunk transport on the same clipped inputs
- Later relax the strict per-chunk flush/ack model as the server-side decoder becomes more truly incremental

### Code review instructions
- Start with:
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/wsclient.go`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/stream_sender.go`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/stream_receiver.go`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/runner.go`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/replay_source.go`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/cmd/transcribe/live.go`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/internal/live/wsclient_test.go`
- Validate with:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go
go test ./internal/live/... -count=1
go test ./... -count=1

# Optional runtime smoke test
python3 ./ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/02-extract_wav_segment.py   --input /home/manuel/code/wesen/2026-04-09--screencast-studio/recordings/rabbit-hole-2026-04-10--2/audio-mix.wav   --output /tmp/transcription-live-clip-000-015.wav   --start 0   --duration 15

go run ./cmd/transcribe live   -i /tmp/transcription-live-clip-000-015.wav   -o ./out-live-ws-clip-000-015   --transport ws   --live-format console,db,txt   --chunk-duration 5   --overlap-seconds 0.5   --replay-speed 0
```

### Technical details
- New Go WS transport files:
  - `internal/live/wsclient.go`
  - `internal/live/stream_sender.go`
  - `internal/live/stream_receiver.go`
- Runner/CLI changes:
  - `internal/live/runner.go`
  - `cmd/transcribe/live.go`
- WS smoke-run evidence (`out-live-ws-clip-000-015/live-summary.json`):
  - `chunks_processed=4`
  - `committed_words=24`
  - `effective_audio_seconds=15.0`
  - `elapsed_seconds=6.442`
  - `audio_seconds_per_wall_second=2.328`
  - `average_server_processing_ms=820`
  - `max_server_processing_ms=1020`
  - `average_end_to_end_ms=2843.75`
  - `max_end_to_end_ms=3734`


## Step 16: Fix WS timestamp anchoring and validate the corrected 120s comparison

Once the first 120-second WS replay finished, it exposed a real correctness bug: the transcript word timestamps extended out to `133.12s` on a 120-second clip. That turned out not to be an ASR issue but a transport/session accounting issue: the buffered WS decoder was advancing its timeline by cumulative buffered duration and ignoring the replay chunk’s absolute `pts`, so each overlapped chunk effectively added its overlap twice.

I fixed that by anchoring each buffered decoder span to the incoming audio `pts`. After that change, a short 15-second WS rerun showed the DB ending at `15.02s`, and a corrected 120-second WS rerun ended at `120.12s` with a much tighter comparison against batch: `318` words vs `323` (`-5`). That is a much stronger result than the earlier HTTP chunk live run on the same clip (`302` vs `323`, `-21`).

### Prompt Context

**User prompt (verbatim):** (same as Step 15)

**Assistant interpretation:** Continue the live-streaming work by validating the new WS path on longer replay inputs and fix any transport/session correctness problems that show up.

**Inferred user intent:** Use the new WS path seriously enough to expose real issues, then fix those issues rather than stopping at superficial smoke tests.

**Commit (code):** 8e2e4fc — "Fix WS timestamp anchoring for overlapped replay chunks"

### What I did
- Investigated the first 120s WS comparison and found:
  - `max_end=133.12s` on a 120s clip
  - root cause: WS buffered decode spans were using cumulative duration instead of incoming `pts`
- Updated:
  - `server/live_decoder.py`
  - `server/live_sessions.py`
  - `server/live_sessions_test.py`
- Changed the WS buffered decoder so:
  - each buffered span remembers its starting `pts`
  - partial/final decode timestamps use that `pts` anchor
  - finalized duration tracks the maximum finalized end time rather than cumulative overlap-additive duration
- Revalidated with:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go
python3 -m unittest discover -s server -p '*_test.py'
python3 -m py_compile server/server.py server/live_decoder.py server/live_sessions.py
```

- Ran a corrected 15s WS smoke replay in tmux:
  - session: `live-ws-smoke-15s-fix`
  - output dir: `out-live-ws-clip-000-015-fix`
- Verified corrected 15s DB coverage:
  - `max_end=15.02s`
- Ran a corrected 120s WS replay in tmux:
  - session: `live-ws-clip-120s-fix`
  - output dir: `out-live-ws-clip-000-120-fix`
- Compared corrected WS 120s output against:
  - `out-batch-clip-000-120/transcript.db`
  - `out-live-clip-000-120/transcript.db`

### Why
- The first WS transport path was only useful if its timestamps were actually reliable.
- Because replay chunks intentionally overlap, cumulative-duration timestamping was guaranteed to drift; the decoder needed to honor absolute `pts` to stay correct.
- Without this fix, any WS-vs-batch comparison would be misleading on timing alone.

### What worked
- The root cause was straightforward once the `133.12s` coverage number was compared against the 120s input duration.
- Anchoring decode spans to `pts` fixed the drift immediately on the 15s validation run.
- The corrected 120s WS result came out much tighter than the HTTP chunk-live result.

### What didn't work
- The first 120s WS run produced incorrect coverage because overlap was effectively double-counted at the decoder timeline layer.
- I did not preserve that run as a “good” baseline; instead I replaced it with the corrected `out-live-ws-clip-000-120-fix/` run.

### What I learned
- For replay-driven WS work, `pts` is not just metadata; it is the timeline authority.
- Even with a buffered decoder, we can preserve correct transcript timing as long as the session layer respects absolute timing anchors.
- The WS path is already outperforming the earlier HTTP chunk-live path on the same 120s clip in terms of word-count parity with batch.

### What was tricky to build
- The main subtlety was distinguishing two meanings of “progress”: cumulative finalized duration vs absolute timeline position. In overlapped replay, those are not the same thing. The decoder originally treated them as the same, which is why overlap inflated the apparent coverage. The fix was to keep absolute buffered-span start PTS separate from the tracked finalized end time.

### What warrants a second pair of eyes
- Whether the current `pts` anchoring rule is still sufficient once we move from per-chunk WS flushes to more continuous multi-frame buffered spans
- Whether the current WS accumulator semantics need any refinement when partial previews become denser or more revision-heavy
- Whether the corrected 120s WS output should now become the primary Phase 3 comparison baseline

### What should be done in the future
- Prefer the corrected `out-live-ws-clip-000-120-fix/` run over the earlier broken WS run for any future reporting
- Compare the corrected WS transcript against batch at the actual word-diff level if we need to inspect the remaining `-5`
- Continue moving WS toward more continuous decoding/finalization once this chunk-paced transport baseline is considered stable enough

### Code review instructions
- Start with:
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/server/live_decoder.py`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/server/live_sessions.py`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/server/live_sessions_test.py`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/out-live-ws-clip-000-120-fix/live-summary.json`
  - `/home/manuel/code/wesen/2026-04-13--transcription-go/out-live-ws-clip-000-120-fix/transcript.db`
- Validate with:

```bash
cd /home/manuel/code/wesen/2026-04-13--transcription-go
python3 -m unittest discover -s server -p '*_test.py'
python3 -m py_compile server/server.py server/live_decoder.py server/live_sessions.py

python3 ./ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/01-compare_transcript_dbs.py   --live-db out-live-ws-clip-000-120-fix/transcript.db   --reference-db out-batch-clip-000-120/transcript.db   --summary-json out-live-ws-clip-000-120-fix/live-summary.json
```

### Technical details
- Broken initial 120s WS run:
  - words: `350`
  - `max_end=133.12s`
- Corrected 15s WS run:
  - words: `23`
  - `max_end=15.02s`
- Corrected 120s WS run:
  - words: `318`
  - `max_end=120.12s`
  - batch words: `323`
  - delta vs batch: `-5`
  - earlier HTTP chunk-live words: `302`
  - delta vs earlier HTTP chunk-live: `+16`
