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
