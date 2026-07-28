---
Title: Investigation Diary
Ticket: VIDEO-CORPUS-PIPELINE
Status: active
Topics:
    - transcription
    - asr
    - video
    - playlist
    - corpus
    - pipeline
    - sqlite
    - go
    - dagger
    - nemo
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/worktrees/2026-07-28--transcription-go-video-pipeline/ttmp/2026/07/28/VIDEO-CORPUS-PIPELINE--adapt-transcription-go-for-playlist-video-corpus-pipelines/analysis/01-current-system-and-video-corpus-gap-analysis.md
      Note: Evidence-based current-state analysis
    - Path: /home/manuel/worktrees/2026-07-28--transcription-go-video-pipeline/ttmp/2026/07/28/VIDEO-CORPUS-PIPELINE--adapt-transcription-go-for-playlist-video-corpus-pipelines/design-doc/01-intern-guide-playlist-video-corpus-transcription-architecture-and-implementation.md
      Note: Main intern-facing architecture and implementation guide
    - Path: /home/manuel/worktrees/2026-07-28--transcription-go-video-pipeline/ttmp/2026/07/28/VIDEO-CORPUS-PIPELINE--adapt-transcription-go-for-playlist-video-corpus-pipelines/reference/01-corpus-database-and-pipeline-api-contracts.md
      Note: Manifest, state, schema, and Go interface contracts
    - Path: /home/manuel/worktrees/2026-07-28--transcription-go-video-pipeline/ttmp/2026/07/28/VIDEO-CORPUS-PIPELINE--adapt-transcription-go-for-playlist-video-corpus-pipelines/playbook/01-playlist-corpus-operator-playbook.md
      Note: Planned operating and recovery procedures
ExternalSources: []
Summary: Chronological record of worktree creation, ticket setup, code analysis, architecture design, documentation, validation, and publication.
LastUpdated: 2026-07-28T00:00:00Z
WhatFor: Preserve decisions, failures, evidence, and continuation instructions for the video corpus adaptation.
WhenToUse: Read before resuming implementation or reviewing why the design has its current boundaries.
---

# Diary

## Goal

Create an isolated worktree and docmgr ticket, analyze the existing Go/Dagger/Nemotron implementation, and produce a clear intern-facing architecture, design, implementation, API, schema, and operator guide for adapting it to a resumable multi-video corpus pipeline.

## Step 1: Create the isolated worktree and ticket

The source repository had an unrelated uncommitted change in `internal/server/dagger.go`. Rather than mixing that state into the new work, I created a new branch and adjacent Git worktree from committed `HEAD` (`89fb5db`). The video-corpus ticket and its documentation are therefore isolated from the original checkout's local modification.

The ticket was created with separate analysis, design, reference, diary, and playbook documents. This structure lets the main guide remain readable while implementation-level schemas and operating commands live in focused supporting documents.

### Prompt Context

**User prompt (verbatim):** "ok, let's start a new worktree in that 2026-04-13--transcription-go directory, a crew a new docmgr ticket there to adapt the functionality to work with our video pipeline."

**Assistant interpretation:** Create a Git worktree from the transcription-go repository, branch it for video-pipeline adaptation, and create a docmgr ticket inside that worktree.

**Inferred user intent:** Keep the adaptation isolated and documented so implementation can proceed without disturbing the existing stable transcription checkout.

**Commit (docs):** 9f31f7e8123a48c2a2044482a751b6345562d71d — "Create video corpus pipeline design ticket"

### What I did
- Inspected source repository status, branches, worktrees, and current commit.
- Preserved the uncommitted source-checkout change without copying it.
- Created `/home/manuel/worktrees/2026-07-28--transcription-go-video-pipeline`.
- Created branch `feature/video-pipeline-corpus` from `89fb5db`.
- Added project vocabulary for `video`, `sqlite`, `pipeline`, `corpus`, and `playlist`.
- Created ticket `VIDEO-CORPUS-PIPELINE` with five focused documents and five tasks.
- Committed only the new ticket and vocabulary changes.

### Why
- A worktree provides clean branch isolation while preserving access to the original repository and history.
- The source checkout was dirty, so branching from committed `HEAD` avoided accidentally capturing unrelated work.
- Separate analysis/design/reference/playbook documents improve intern onboarding and future implementation review.

### What worked
- `git worktree add -b feature/video-pipeline-corpus ... HEAD` succeeded.
- The worktree was clean after creation.
- Docmgr created and recognized the new ticket and documents.
- The initial documentation commit was focused.

### What didn't work
- No implementation was attempted in this step.
- The user's phrase “in that directory” could imply nesting a worktree inside the source repository. Git worktrees are safer as adjacent directories, so the worktree was placed under `/home/manuel/worktrees/` and the source repository was left intact.

### What I learned
- The original checkout's local `internal/server/dagger.go` change must be reconciled separately before any future merge if it overlaps this branch.
- The repository already has mature `TRANSCRIPTION-GO` and `LIVE-TRANSCRIPTION` tickets, so the new ticket should build on those rather than repeat their history.

### What was tricky to build
- The main sharp edge was preserving uncommitted source state. Creating the worktree from `HEAD` made the new branch reproducible but intentionally excludes that local change.

### What warrants a second pair of eyes
- Review the original checkout's uncommitted Dagger change before merging this feature branch.
- Confirm the chosen ticket and branch names match project conventions.

### What should be done in the future
- Implement the design in this worktree.
- Merge or rebase only after deciding what to do with the original checkout's local Dagger change.

### Code review instructions
- Run `git worktree list` from the source repository.
- Run `git status --short --branch` in both the source checkout and new worktree.
- Review commit `9f31f7e` to confirm it contains only ticket scaffolding and vocabulary.

### Technical details

```text
source: /home/manuel/code/wesen/2026-04-13--transcription-go
worktree: /home/manuel/worktrees/2026-07-28--transcription-go-video-pipeline
branch: feature/video-pipeline-corpus
base: 89fb5db03205e9a3d746405d6d605e7de9d26ec0
```

## Step 2: Analyze the current implementation and design the corpus pipeline

The current code was inspected directly. The analysis confirmed that the working architecture should be extended rather than replaced: Go already owns orchestration and output, Dagger already keeps Nemotron reproducible, FastAPI already returns timed words, and the output package already creates words/chunks/exports for one file. The missing layer is corpus identity, warm-service multi-item orchestration, atomic revisions, resume state, and corpus-wide search.

The resulting documentation defines one canonical corpus database while retaining per-video exports. It treats raw word timestamps as evidence, chunks as derived sentence/display fragments, and exports as rebuildable artifacts. It also records concrete correctness gaps in the current implementation that should be fixed before a large run.

### Prompt Context

**User prompt (verbatim):** "Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Perform a deep local analysis and write a comprehensive technical onboarding/design package with prose, bullets, pseudocode, diagrams, schemas, API contracts, exact file references, an operator playbook, and reMarkable delivery.

**Inferred user intent:** Give a new intern enough system context and concrete implementation detail to build the adaptation correctly without rediscovering prior architecture and failure modes.

### What I did
- Read `cmd/transcribe/root.go` and `batch.go` to map CLI and lifecycle.
- Read `internal/convert/convert.go` to establish the WAV boundary and memory model.
- Read `internal/asr/client.go` to map HTTP APIs and multipart behavior.
- Read `internal/server/dagger.go` to verify model/cache/tunnel lifecycle.
- Read `server/server.py` to document model configuration, chunking, timestamps, and live APIs.
- Read output types, formatters, SQLite writer, and tests.
- Read existing ticket history and streaming architecture documentation.
- Grounded the design in the Southwell manifest and observed 37/36/1 source state.
- Wrote the gap analysis, API/schema contract, intern guide, and operator playbook.

### Why
- The corpus design must be based on current code behavior, not only older aspirational documents.
- A shared schema and interface contract reduces ambiguity before implementation begins.
- The operator playbook forces the design to address interruption, failure, export repair, and verification.

### What worked
- The existing architecture has clean reusable seams: converter, ASR client, service manager, formatters, and SQLite driver.
- The Southwell smoke run already proved the single-file model path.
- The earlier Nemotron database design maps naturally to corpus-level `videos -> revisions -> words -> chunks`.

### What didn't work
- A delegated analysis was briefly started before the user said “no subagents.” It was stopped immediately and none of its output was used. All analysis and documentation in this ticket were produced directly from local file inspection.
- The current source has several issues discovered during analysis:
  - multipart uploads buffer the full WAV in memory;
  - Dagger/model lifecycle is inside one-file `runBatch`;
  - trailing subtitle segments can end at zero;
  - writer errors are ignored;
  - legacy SQLite chunk `word_count` is always zero because `split` is a placeholder;
  - full ASR results do not carry confidence or source chunk index;
  - overlap deduplication is not explicit.

### What I learned
- The adaptation is primarily a lifecycle and persistence project, not a new ASR project.
- Source availability and transcript processing state must be separate; members-only is not an ASR failure.
- A per-video database is useful as an artifact but insufficient as the canonical corpus model.
- Export failure must not invalidate a successful transcript revision.

### What was tricky to build
- The design must balance revision history with implementation simplicity. Immutable successful revisions plus an active revision pointer preserve provenance without requiring event sourcing.
- ASR happens outside transactions, but transcript activation must be atomic. The attempt/revision protocol solves this by committing one complete result in a short transaction.
- The source manifest is yt-dlp-derived, but internal business logic should not depend on yt-dlp's unstable schema. A normalized versioned manifest is the boundary.

### What warrants a second pair of eyes
- Whether old transcript revisions should be retained indefinitely or garbage-collected.
- Whether Phase 1 should store source chunks from the server before overlap deduplication.
- Whether the model artifact revision can be pinned more strongly than requirements/model name.
- Whether an existing migration framework should be introduced or an embedded schema is sufficient initially.
- Whether video demux should remain entirely upstream.

### What should be done in the future
- Implement Phase 0 correctness fixes with regression tests.
- Add `internal/corpus` manifest/store/runner packages.
- Add the `transcribe corpus` command group.
- Validate with a fake transcriber, then Southwell video 019, interruption/resume, and the full corpus.

### Code review instructions
- Start with the analysis document for current facts and gaps.
- Read the main guide Sections 3–12 for boundaries, lifecycle, schema, and phases.
- Review the reference contract before writing store code.
- Use the operator playbook to assess whether the proposed interface is operationally complete.
- Validate file references and run the existing tests with `make test`.

### Technical details

```text
Current unit of work: one WAV file
Proposed unit of work: one corpus run with per-video atomic revisions
Canonical evidence: manifest identities, hashes, pipeline fingerprint, word timestamps
Derived data: chunks, FTS, SRT, VTT, TXT, future summaries/embeddings
Immediate source: 37 playlist entries, 36 available, 1 members-only
```

## Step 3: Honor the no-subagent constraint

After the user explicitly requested no subagents and asked whether I would perform the analysis myself, I stopped the in-progress delegated read-only analysis. I then inspected the implementation and authored the ticket documents directly.

### Prompt Context

**User prompt (verbatim):** "no subagents"

**Assistant interpretation:** Stop all delegated work and perform the remaining analysis and documentation directly.

**Inferred user intent:** Keep ownership and reasoning in the primary agent session rather than distributing the task.

### What I did
- Requested stop for the delegated run.
- Removed a generated untracked Python `__pycache__` directory.
- Checked that no delegated context artifact was present or used.
- Continued direct code inspection and writing.

### Why
- The user's execution constraint is authoritative.
- Avoiding mixed authorship makes the resulting analysis traceable to this session's direct evidence.

### What worked
- Delegated work was stopped before its output was incorporated.
- The worktree remained clean except for intended ticket documents.

### What didn't work
- Starting delegation before the user stated this preference created avoidable overhead.

### What I learned
- For this project, continue all subsequent work in the primary session unless the user explicitly changes the constraint.

### What was tricky to build
- There was no technical complexity beyond ensuring no partial generated artifacts remained in the worktree.

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- Do not launch subagents for this ticket without explicit permission.

### Code review instructions
- Confirm `git status` contains only intended ticket changes.

## Step 4: Validate, commit, and publish the design bundle

The repository's complete Go test suite passed without source changes, docmgr reported no ticket problems, and the documentation diff passed Git whitespace validation. The complete ticket bundle was then rendered and uploaded to the ticket-specific reMarkable directory.

The ticket remains active because product implementation has not begun. The documentation and contracts are complete enough to guide implementation, but this step does not claim that the proposed `transcribe corpus` commands exist yet.

### Prompt Context

**User prompt (verbatim):** (see Step 2)

**Assistant interpretation:** Validate and publish the requested intern-facing documentation after completing the direct analysis.

**Inferred user intent:** Make the design both reviewable in Git and available as an offline reMarkable handoff.

**Commit (docs):** f289174a6402b4ef8e7a7b028165ca1075ee601d — "Document video corpus transcription architecture"

### What I did
- Ran `go test ./... -count=1`; all packages passed.
- Ran `docmgr doctor --ticket VIDEO-CORPUS-PIPELINE --stale-after 30`; all checks passed.
- Ran `git diff --check` before committing.
- Committed the analysis, guide, contracts, playbook, diary, index, tasks, and changelog.
- Uploaded the six-document bundle with reMarkable ToC depth 2.

### Why
- Tests verify the existing baseline remains healthy even though this phase changes documentation only.
- Docmgr validation catches broken metadata and related-file references.
- The reMarkable bundle provides the requested intern reading copy while Git remains the editable source.

### What worked
- All Go tests passed.
- Docmgr reported `✅ All checks passed`.
- Upload returned: `OK: uploaded Video Corpus Transcription Intern Guide.pdf -> /ai/2026/07/28/VIDEO-CORPUS-PIPELINE`.

### What didn't work
- N/A for validation and upload.

### What I learned
- The documentation renders as one coherent bundle when ordered index, analysis, design, API contract, playbook, and diary.

### What was tricky to build
- The playbook describes a future command surface. It is clearly labeled as proposed so operators do not attempt commands before implementation.

### What warrants a second pair of eyes
- Inspect Mermaid diagrams and wide SQL/code blocks on the reMarkable display for readability.
- Review the schema and interface contracts before implementation freezes them.

### What should be done in the future
- Begin Phase 0 correctness fixes and tests in this worktree.
- Update the guide and playbook whenever implemented flags or schema details diverge from the proposal.

### Code review instructions
- Review commits `9f31f7e` and `f289174`.
- Run `go test ./... -count=1` and `docmgr doctor --ticket VIDEO-CORPUS-PIPELINE --stale-after 30`.
- Start implementation from the main guide's Phase 0 and the reference contract's required invariants.

## Step 5: Implement the corpus pipeline and validate with one video

The design was implemented end to end and validated with Southwell video 019 (RZPeFGg84Ng). The implementation covers Phase 0 correctness fixes, the full corpus package (manifest, fingerprint, store, runner, transcriber, export, search), the `transcribe corpus` command group, a manifest converter script, and a normalized manifest for the 37-entry Southwell corpus.

### Prompt Context

**User prompt (verbatim):** "Implement as far as you can test with one video."

**Assistant interpretation:** Build the designed corpus pipeline in the worktree and validate it with one real video through the warm Nemotron service, including resume and export repair.

**Inferred user intent:** Prove the architecture works end to end on real audio before scaling to the full 36-video corpus.

**Commits (code):**
- 557bdeb — "Fix trailing segment timestamps, propagate writer errors, fix chunk word counts" (Phase 0)
- 5aaff02 — "Implement corpus pipeline: manifest, store, runner, exports, search, CLI"
- 9f38c0a — "Sort words by start time for overlap chunk boundaries, add manifest converter"

### What I did
- Phase 0: fixed trailing segment end timestamp (was 0, now uses last word end); propagated SRT/VTT/TXT writer errors in batch.go; fixed legacy SQLite chunk word_count (replaced placeholder `split` with `strings.Fields` and corrected count from linked rows); added regression tests.
- Implemented `internal/corpus` package: manifest loading/validation, pipeline fingerprinting, SQLite store with full schema (corpora, videos, transcript_attempts, transcript_revisions, words, chunks, chunk_words, exports, chunk_fts), resume planner, atomic commit protocol, export coordinator, FTS5 search, sequential warm-service runner.
- Implemented `cmd/transcribe/corpus.go` with `run`, `status`, `search`, and `export` subcommands.
- Wrote `scripts/convert_media_manifest.py` to convert the yt-dlp media manifest into the normalized transcription manifest.
- Generated and validated the 37-item Southwell manifest (36 available, 1 members-only).
- Ran dry-run: 37 total, 36 available, 1 unavailable, 36 pending.
- Ran real transcription of video 019: 467 words, 24 chunks, 171.5s, committed revision 1.
- Generated SRT/VTT/TXT exports from committed rows.
- Verified resume: rerun performed zero ASR calls, zero service starts.
- Verified export repair: deleted SRT, regenerated via `corpus export` without Dagger.
- Verified search: "category theory" returned timestamped chunks with deep links.

### Why
- The design required proof that the warm-service, atomic-revision, resume architecture works on real audio before scaling.
- Video 019 is the shortest Southwell lecture (171.5s) and already had a known 467-word reference.

### What worked
- All 8 Go test packages pass, including 8 new corpus tests.
- `gofmt -l cmd internal` returns no diffs.
- `go vet ./...` clean.
- DB integrity_check and foreign_key_check pass.
- `transcript_revisions.word_count` (467) equals `COUNT(words)` (467).
- Zero chunks with mismatched `word_count`.
- Final SRT cue ends at 00:02:50,879 (nonzero).
- Resume: attempt count stayed at 2 (no new ASR attempt); summary showed `will_transcribe=0 service_start=false`.
- Export repair regenerated the deleted SRT without starting Dagger.
- Search returned 5 ranked hits with YouTube deep links.
- `docmgr doctor` passes.

### What didn't work
- First real run failed validation: overlap chunking produced out-of-order words at boundaries (word 159 at 61.120s before word 158 at 61.920s). Fixed by sorting words by start time before validation and storage. The first failed attempt is preserved in `transcript_attempts` as evidence.
- The existing `corpus.db` from the caption indexing had a different schema; the Nemotron corpus uses a separate `corpus-nemotron.db`.
- `SetMaxOpenConns(1)` caused a deadlock with prepared statements inside transactions; removed it.

### What I learned
- Nemotron's 2-second overlap naturally produces out-of-order words at chunk boundaries. Sorting by start time before storage is the correct fix; the alternative of tighter validation would reject valid overlap words.
- The existing `corpus.db` name collision is a real operational hazard; the operator playbook documents the separate database path.
- Prepared statements within transactions on a single-connection pool deadlock in modernc.org/sqlite.

### What was tricky to build
- The transcriber needs the service endpoint, which only exists after the service starts. The runner uses a `TranscriberFactory` function that receives the endpoint, rather than a pre-built `Transcriber`.
- The FTS5 external content table (`content='chunks'`) requires manual `INSERT INTO chunk_fts (rowid, text)` during the same transaction that creates the chunk row.
- The manifest converter must handle unavailable items (members-only) even when they have no media file or entry in the media manifest.

### What warrants a second pair of eyes
- The word sort is a stable insertion sort (O(n²)); for 467 words this is fine, but for hour-long lectures with thousands of words, consider `sort.SliceStable`.
- The first failed attempt is preserved; a cleanup policy for abandoned/failed attempts should be defined.
- The corpus database has no migration framework; schema changes require manual handling.
- Streaming multipart uploads are not yet implemented; the current client still buffers the full WAV in memory.

### What should be done in the future
- Run the full 36-video Southwell corpus.
- Stream multipart uploads for long lectures.
- Add per-word confidence and source chunk index from the server response.
- Add a migration framework or versioned schema.
- Add a cleanup policy for old attempts and retired revisions.

### Code review instructions
- Review commits 557bdeb, 5aaff02, 9f38c0a on branch `feature/video-pipeline-corpus`.
- Run `go test ./... -count=1` and `gofmt -l cmd internal`.
- Run `./transcribe corpus run --manifest ... --dry-run` to inspect planning.
- Inspect `corpus-nemotron.db` with the queries in the operator playbook.
- Verify resume by running twice and checking attempt count.

### Technical details

```text
Video: RZPeFGg84Ng (019, My New Category Theory Book)
Words: 467
Chunks: 24
Duration: 171.456s
Attempts: 2 (1 failed validation, 1 succeeded)
Exports: SRT (3327B), VTT (3272B), TXT (2520B)
Database: corpus-nemotron.db
Final SRT cue end: 00:02:50,879
```

## Step 6: Stream multipart uploads and start the full corpus run

After the one-video validation, two improvements were made before scaling to the full 36-video corpus: streaming multipart uploads and the full run itself.

### Prompt Context

**User prompt (verbatim):** "commit if you haven't yet, keep a detailed diary as you work, and continue"

**Assistant interpretation:** Commit any pending work, maintain the diary, and proceed with the next implementation steps.

**Inferred user intent:** Scale from the one-video proof to the full corpus while improving robustness.

**Commit (code):** b71b809 — "Stream multipart uploads via io.Pipe to avoid full-WAV memory buffering"

### What I did
- Replaced the `bytes.Buffer`-based multipart upload in `internal/asr/client.go` with an `io.Pipe`-streamed implementation. The goroutine writes the multipart form data to the pipe while the HTTP request body reads from it, so the full WAV file is never held in memory.
- Removed the unused `bytes` import.
- Verified all existing ASR tests pass with the streaming implementation.
- Checked WAV file sizes: the largest lecture (video 021) produces an 812 MB WAV file, confirming streaming is essential.
- Measured total corpus audio: 281,179 seconds (78.1 hours) across 36 available videos.
- Started the full corpus run in a tmux session (`southwell-corpus`) with all 36 available videos, skipping the already-completed video 019.
- Confirmed the service started, video 001 (US4Zr1WKD-8, 58 minutes) began transcribing, and all 35 remaining available videos show as `pending`.

### Why
- The largest WAV file is 812 MB. The old `bytes.Buffer` approach would have loaded that entire file into Go process memory before sending the HTTP request. Streaming via `io.Pipe` keeps memory bounded regardless of file size.
- The full corpus run validates the warm-service reuse across many videos, the resume planner at scale, and the sequential processing loop under real load.

### What worked
- Streaming upload tests pass; `go test ./...` all green.
- `gofmt -l` clean.
- The full corpus run started successfully in tmux.
- Video 019 was correctly skipped (resume planner showed `unchanged`).
- The Dagger service started once and is being reused across videos.
- Video 001 is actively transcribing with chunk-level progress visible.

### What didn't work
- No failures yet. The run is ongoing.

### What I learned
- 78 hours of audio on a CPU-bound ASR model is a multi-hour to multi-day operation. The sequential warm-service approach is correct for reliability but not fast.
- The Dagger cache makes service startup much faster on subsequent runs (~40s vs cold-start minutes).
- The `io.Pipe` pattern for multipart streaming is straightforward in Go but requires careful error propagation from the writer goroutine to the pipe reader.

### What was tricky to build
- The `io.Pipe` goroutine must close the pipe with an error if any write step fails, otherwise the HTTP client would hang waiting for a body that never completes.
- The pipe writer must be closed (not just the multipart writer) to signal EOF to the HTTP request body reader.

### What warrants a second pair of eyes
- The streaming goroutine error propagation path: if `io.Copy` fails mid-stream, does the server receive a partial upload and hang, or does it return an error promptly?
- Whether the 78-hour corpus should be run with GPU acceleration or a faster model in the future.
- Whether very long videos (7.4 hours) will complete within Dagger session limits.

### What should be done in the future
- Monitor the full corpus run to completion.
- Add GPU support if CPU-only processing is too slow for routine use.
- Add per-video timeout or cancellation for extremely long videos.
- Consider splitting very long videos into segments for independent processing.

### Code review instructions
- Review commit b71b809 for the `io.Pipe` streaming implementation.
- Run `go test ./internal/asr -count=1 -v` to verify streaming tests.
- Inspect the tmux session: `tmux capture-pane -pt southwell-corpus:0 | tail -50`.
- Query the corpus database for progress: `sqlite3 corpus-nemotron.db "SELECT processing_state, COUNT(*) FROM videos GROUP BY processing_state;"`

### Technical details

```text
Largest WAV: 812 MB (video 021)
Total audio: 281,179 seconds (78.1 hours)
Available videos: 36
Completed before full run: 1 (video 019)
Pending at full run start: 35
tmux session: southwell-corpus
Streaming: io.Pipe-based multipart upload
```
