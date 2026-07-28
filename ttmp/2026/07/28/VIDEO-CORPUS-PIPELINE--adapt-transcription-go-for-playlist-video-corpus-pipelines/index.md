---
Title: Adapt Transcription Go for Playlist Video Corpus Pipelines
Ticket: VIDEO-CORPUS-PIPELINE
Status: active
Topics:
    - transcription
    - asr
    - audio
    - video
    - playlist
    - corpus
    - pipeline
    - sqlite
    - go
    - dagger
    - nemo
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/worktrees/2026-07-28--transcription-go-video-pipeline/cmd/transcribe/batch.go
      Note: Existing one-file pipeline being adapted
    - Path: /home/manuel/worktrees/2026-07-28--transcription-go-video-pipeline/internal/output/sqlite.go
      Note: Existing word/chunk persistence baseline
    - Path: /home/manuel/worktrees/2026-07-28--transcription-go-video-pipeline/server/server.py
      Note: Nemotron inference service reused by the proposed corpus runner
ExternalSources:
    - https://www.youtube.com/playlist?list=PLCTMeyjMKRkoS699U0OJ3ymr3r01sI08l
Summary: Design ticket for a warm-service, resumable, corpus-wide video transcription pipeline built around the existing Go/Dagger/Nemotron implementation.
LastUpdated: 2026-07-28T00:00:00Z
WhatFor: Coordinate analysis, design, implementation contracts, and operations for the corpus adaptation.
WhenToUse: Start here when reviewing or resuming VIDEO-CORPUS-PIPELINE.
---

# Adapt Transcription Go for Playlist Video Corpus Pipelines

## Overview

The existing tool reliably transcribes one WAV file into timed words, subtitle/text exports, and a per-file SQLite database. This ticket designs the next layer: import an ordered video-corpus manifest, reuse one warm Nemotron service across many videos, persist immutable per-video transcript revisions in one corpus database, resume failures safely, regenerate exports without ASR, and search across videos by timestamp.

The immediate validation corpus is Richard Southwell's *Category Theory For Beginners*: 37 manifest entries, 36 accessible local videos/audio files, and one members-only entry that must remain explicitly represented.

## Key documents

1. [Current System and Video Corpus Gap Analysis](analysis/01-current-system-and-video-corpus-gap-analysis.md)
2. [Intern Guide: Playlist Video Corpus Transcription Architecture and Implementation](design-doc/01-intern-guide-playlist-video-corpus-transcription-architecture-and-implementation.md)
3. [Corpus Database and Pipeline API Contracts](reference/01-corpus-database-and-pipeline-api-contracts.md)
4. [Investigation Diary](reference/02-investigation-diary.md)
5. [Playlist Corpus Operator Playbook](playbook/01-playlist-corpus-operator-playbook.md)

## Core design

```text
versioned manifest
  -> corpus import and resume plan
  -> one warm Dagger/Nemotron service
  -> sequential per-video inference
  -> atomic transcript revisions
  -> words -> chunks -> FTS and exports
```

## Current status

Documentation and implementation contracts are ready. Product code has not yet been changed on this branch. The first implementation phase should fix known formatter/legacy SQLite correctness issues and extract reusable single-file behavior before adding `internal/corpus`.

## Branch and worktree

```text
branch: feature/video-pipeline-corpus
worktree: /home/manuel/worktrees/2026-07-28--transcription-go-video-pipeline
base: 89fb5db
```

## Tasks and history

- [tasks.md](tasks.md)
- [changelog.md](changelog.md)
