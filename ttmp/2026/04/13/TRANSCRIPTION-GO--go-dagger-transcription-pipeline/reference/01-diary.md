---
Title: "Diary"
Ticket: TRANSCRIPTION-GO
Status: active
Topics:
    - transcription
    - audio
    - dagger
    - docker
    - go
    - nemo
    - asr
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../2026-04-09--screencast-studio/ttmp/2026/04/13/TRANSCRIPT-PIPELINE--setting-up-an-analysis-pipeline-for-transcripts/design-doc/02-dagger-docker-handoff.md
      Note: Source handoff document with pipeline specs and Python reference
    - Path: ../../../../../go-go-golems/glazed/cmd/build-web/main.go
      Note: Reference Dagger Go implementation
    - Path: ../../../../../corporate-headquarters/smailnail/cmd/build-web/main.go
      Note: Second Dagger Go reference implementation
    - Path: /home/manuel/.pi/agent/skills/go-web-dagger-pnpm-build/SKILL.md
      Note: Dagger build pattern skill
ExternalSources: []
Summary: "Investigation diary for the Go + Dagger transcription pipeline ticket."
LastUpdated: 2026-04-13T00:00:00Z
WhatFor: "Record the investigation and design process for the Go + Dagger transcription pipeline."
WhenToUse: "When continuing work on the TRANSCRIPTION-GO ticket."
---

# Diary

## Goal

This diary documents the investigation, analysis, and design work for creating a self-contained Go + Dagger transcription pipeline that containerizes the existing Python-based NVIDIA Nemotron ASR pipeline.

---

## Step 1: Source Material Review and Ticket Setup

Read the original handoff document (`TRANSCRIPT-PIPELINE/design-doc/02-dagger-docker-handoff.md`) which provided a comprehensive spec for the transcription pipeline: audio preprocessing via ffmpeg, NVIDIA Nemotron 0.6B ASR with chunked processing, SQLite storage with word-level timestamps, and SRT export with filler filtering. Also studied two reference Dagger Go implementations (Glazed and Smailnail `cmd/build-web/main.go`) to understand the established pattern for Dagger-based builds in Go.

### Prompt Context

**User prompt (verbatim):** "Create a new docmgr ticket to create a self-contained go + docker (using dagger) pipeline to run the transcription. Read ../2026-04-09--screencast-studio/ttmp/2026/04/13/TRANSCRIPT-PIPELINE--setting-up-an-analysis-pipeline-for-transcripts/design-doc/02-dagger-docker-handoff.md to get you started, then do your own analysis. You can also look at the go web dagger build skill for more information about using dagger from within go, and also ../go-go-golems/glazed/cmd/build-web for a concrete example. Then write your analysis and implementation plan in the ticket and upload to remarkable. Keep a diary as you work."

**Assistant interpretation:** Create a docmgr ticket for a Go + Dagger transcription pipeline, read the existing design doc and reference implementations, do independent analysis, write a comprehensive design doc, and upload to remarkable.

**Inferred user intent:** Produce an actionable design doc that can be handed off for implementation, backed by evidence from the existing Python pipeline and established Dagger patterns.

### What I did

1. Read the handoff doc (`02-dagger-docker-handoff.md`) — full pipeline specification
2. Read the go-web-dagger-pnpm-build skill — Dagger Go pattern documentation
3. Read Glazed `cmd/build-web/main.go` (4136 bytes) — concrete Dagger Go program with cache volumes and local fallback
4. Read Smailnail `cmd/build-web/main.go` — similar pattern with CacheVolume
5. Read the full Python pipeline (`transcript_db.py`, 486 lines) — chunked ASR, SQLite schema, SRT export
6. Read `transcribe_to_srt.py` (156 lines) — simpler transcription (known to OOM on long audio)
7. Read the TRANSCRIPT-PIPELINE investigation diary — learned about dependency fragility and OOM issues
8. Created docmgr ticket `TRANSCRIPTION-GO` with topics: transcription, audio, dagger, docker, go, nemo, asr
9. Created design doc and diary documents
10. Wrote comprehensive design doc covering architecture, implementation plan, risks, and alternatives

### Why

The user needs a containerized, reproducible transcription pipeline that can be invoked from Go. The existing Python pipeline works but is fragile (specific Python version, heavy deps, no caching strategy). Dagger provides the containerization layer; Go provides the CLI/library entrypoint.

### What worked

- The handoff doc was extremely detailed — contained exact Python code, dependency versions, schema definitions, and Docker considerations. This made the analysis straightforward.
- The Glazed/Smailnail Dagger patterns are directly applicable: `CacheVolume` for pip/HF caches, `WithExec` for pipeline steps, local fallback pattern.
- The TRANSCRIPT-PIPELINE diary provided critical context about what failed (OOM on full audio, Python version sensitivity).

### What didn't work

- No issues encountered during analysis phase. All source material was accessible and well-structured.

### What I learned

1. The pipeline has a natural two-stage boundary: ffmpeg (fast, stateless) → Python ASR (slow, stateful). This maps cleanly to two Dagger containers.
2. The pip install is the biggest performance concern (~3 GB, 5+ minutes cold). `CacheVolume` for pip cache is essential.
3. The HF model cache (~1.2 GB) is the second concern. Another `CacheVolume` for `/root/.cache/huggingface`.
4. The Python script needs to be adapted: remove the ffmpeg conversion step (done in Container 1), add CLI args for container use, write a manifest.json for the Go host.
5. The `cmd/build-web` pattern is very close to what we need — just swap `pnpm install + pnpm build` for `pip install + python transcribe.py`.

### What was tricky to build

No implementation yet — this is the analysis and design phase. The trickiest design decision was whether to use a pre-built Docker image or mount scripts + install deps each run. Chose the mount approach to match the existing pattern and avoid registry overhead, with a note to revisit in Phase 4.

### What warrants a second pair of eyes

1. **Python version choice**: I proposed 3.11 (more common in Docker) but the working venv uses 3.13.2. Need to verify NeMo compatibility.
2. **Container separation**: Two containers vs one (ffmpeg + Python). The two-container approach is cleaner but adds complexity.
3. **Output contract**: The manifest.json approach for Go-Python communication. Is this over-engineered vs. just using known output filenames?

### What should be done in the future

1. Implement Phase 1 (MVP) and validate against existing transcript data
2. Test with Python 3.11 in Docker — if NeMo fails, document and switch to 3.13
3. Measure cold vs. warm cache performance
4. Consider pre-built Docker image if pip install remains too slow

### Code review instructions

**Files to review:**
- `design-doc/01-go-dagger-transcription-pipeline-design-and-implementation-plan.md` — the full design doc

**How to validate:**
```bash
docmgr doc list --ticket TRANSCRIPTION-GO
docmgr doctor --ticket TRANSCRIPTION-GO
```

### Technical details

**Dagger Go SDK pattern (from Glazed):**
```go
client, err := dagger.Connect(ctx)
pnpmStore := client.CacheVolume("repo-ui-pnpm-store")
ctr := client.Container().From("node:22").
    WithMountedCache("/pnpm/store", pnpmStore).
    WithDirectory("/src", webDir).
    WithExec([]string{"pnpm", "install"}).
    WithExec([]string{"pnpm", "build"})
_, err = ctr.Directory("/src/dist").Export(ctx, outPath)
```

**Our adaptation:**
```go
hfCache := client.CacheVolume("transcription-hf-cache")
pipCache := client.CacheVolume("transcription-pip-cache")
ctr := client.Container().From("python:3.11-slim-bookworm").
    WithMountedCache("/root/.cache/huggingface", hfCache).
    WithMountedCache("/root/.cache/pip", pipCache).
    WithDirectory("/pipeline/scripts", scriptsDir).
    WithExec([]string{"pip", "install", "-r", "scripts/requirements.txt"}).
    WithExec([]string{"python", "scripts/transcribe.py", ...})
```

**SQLite schema (from transcript_db.py):**
```sql
words(id, word, start_time, end_time, is_filler, is_removed, confidence, chunk_id)
chunks(id, start_time, end_time, text, word_count, source_type)
chunk_words(chunk_id, word_id, position)
srt_exports(id, filename, config, segment_count, word_count)
srt_segments(id, export_id, sequence_num, start_time, end_time, text)
```
