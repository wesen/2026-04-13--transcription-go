---
Title: Go + Dagger Transcription Pipeline - Design and Implementation Plan
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
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../2026-04-09--screencast-studio/ttmp/2026/04/13/TRANSCRIPT-PIPELINE--setting-up-an-analysis-pipeline-for-transcripts/design-doc/02-dagger-docker-handoff.md
      Note: Original handoff specification with Python reference and Docker considerations
    - Path: ../../../../../../../2026-04-09--screencast-studio/ttmp/2026/04/13/TRANSCRIPT-PIPELINE--setting-up-an-analysis-pipeline-for-transcripts/scripts/transcript_db.py
      Note: Working Python pipeline - chunked ASR
    - Path: ../../../../../../../corporate-headquarters/smailnail/cmd/build-web/main.go
      Note: Second reference Dagger Go build with CacheVolume
    - Path: ../../../../../../../go-go-golems/glazed/cmd/build-web/main.go
      Note: Reference Dagger Go implementation with CacheVolume pattern and local fallback
    - Path: ttmp/2026-04-09--screencast-studio/ttmp/2026/04/13/TRANSCRIPT-PIPELINE--setting-up-an-analysis-pipeline-for-transcripts/design-doc/02-dagger-docker-handoff.md
      Note: Original handoff doc specifying requirements and Python reference implementation
    - Path: ttmp/2026-04-09--screencast-studio/ttmp/2026/04/13/TRANSCRIPT-PIPELINE--setting-up-an-analysis-pipeline-for-transcripts/scripts/transcribe_to_srt.py
      Note: Simpler Python transcription script (no chunking, OOM on long audio)
    - Path: ttmp/2026-04-09--screencast-studio/ttmp/2026/04/13/TRANSCRIPT-PIPELINE--setting-up-an-analysis-pipeline-for-transcripts/scripts/transcript_db.py
      Note: Working Python transcription pipeline with chunked ASR, SQLite storage, and SRT export
    - Path: ttmp/corporate-headquarters/smailnail/cmd/build-web/main.go
      Note: Another reference Dagger Go build program with CacheVolume pattern
    - Path: ttmp/go-go-golems/glazed/cmd/build-web/main.go
      Note: Reference Dagger Go program for container-based builds with cache volumes and local fallback
ExternalSources: []
Summary: 'Design and implementation plan for a self-contained Go CLI that uses Dagger to orchestrate audio transcription in containers: audio preprocessing (ffmpeg), NVIDIA Nemotron ASR inference (Python/NeMo in a CPU container), and output collection (SQLite DB + SRT files). No local Python or CUDA deps required.'
LastUpdated: 2026-04-13T14:35:21.20952146-04:00
WhatFor: Reproducible, zero-local-dep audio transcription pipeline that can be invoked from Go code and integrated into larger Go-based tooling.
WhenToUse: When you need to transcribe audio files using NVIDIA Nemotron ASR without installing Python, torch, or NeMo locally.
---





# Go + Dagger Transcription Pipeline — Design and Implementation Plan

## 1. Executive Summary

This document designs a **self-contained Go CLI** (`transcribe`) that uses **Dagger** to orchestrate audio transcription entirely inside containers. The pipeline:

1. **Accepts** any audio file the host ffmpeg can read.
2. **Preprocesses** it to 16 kHz mono WAV inside a Dagger container (ffmpeg image).
3. **Transcribes** it using NVIDIA Nemotron Speech Streaming 0.6B ASR inside a Python container (CPU-only, 60-second chunks to avoid OOM).
4. **Produces** a SQLite database with word-level timestamps, plus configurable SRT exports (with filler filtering).

The user needs **only Go and Dagger** installed locally — no Python, torch, NeMo, or CUDA required. The model download (~1.2 GB) is cached across runs via a Dagger `CacheVolume`.

### Key Properties

| Property | Detail |
|----------|--------|
| **Runtime** | ~5 min for 30-min audio on CPU |
| **Model** | NVIDIA Nemotron 0.6B ASR (~1.2 GB, cached) |
| **Dependencies** | Go 1.21+, Dagger v0.20+ (host); Python 3.11+ with torch/NeMo (container) |
| **Output** | SQLite DB (`transcript.db`) + SRT files |
| **Pattern** | Follows go-go-golems `cmd/build-web` Dagger pattern |

---

## 2. Problem Statement and Scope

### 2.1 Problem

The existing Python-based transcription pipeline works but has several operational problems:

1. **Fragile local environment**: Requires Python 3.13.2 with a specific venv containing torch, torchaudio, NeMo, soundfile, omegaconf, and pytorch-lightning. Dependency conflicts are frequent (e.g., pyenv 3.11.3 fails with torch version mismatches).
2. **Not reproducible**: The pipeline relies on a specific venv at `/home/manuel/code/wesen/patreon/videos/002-sqrt/venv/` — not portable.
3. **Not composable**: Cannot be called from Go code or integrated into a larger Go-based toolchain.
4. **No caching of model downloads**: The 1.2 GB Nemotron model re-downloads if the HF cache is lost.

### 2.2 Scope

**In scope:**
- Go CLI `cmd/transcribe` that takes audio input, produces SQLite + SRT output
- Dagger-based container orchestration (audio conversion + ASR)
- Model caching via Dagger `CacheVolume`
- SRT export with filler word filtering
- Local Python fallback (if Dagger unavailable)

**Out of scope:**
- GPU acceleration (CPU-only for now, GPU support as future enhancement)
- Real-time streaming transcription
- Web UI or API server
- Multi-speaker diarization

---

## 3. Current-State Analysis

### 3.1 Working Python Pipeline

The reference implementation lives at:
- **Main pipeline**: `TRANSCRIPT-PIPELINE/scripts/transcript_db.py` (486 lines)
- **Simple transcription**: `TRANSCRIPT-PIPELINE/scripts/transcribe_to_srt.py` (156 lines)
- **Query tool**: `TRANSCRIPT-PIPELINE/scripts/query_transcript.py` (228 lines)

**Pipeline flow** (from `transcript_db.py`):

```
audio-mix.wav
    ↓  ffmpeg -ar 16000 -ac 1
audio-mix_16k_mono.wav
    ↓  60s chunks, Nemotron 0.6B ASR
Word-level timestamps (start, end, text, is_filler)
    ↓  SQLite INSERT
transcript.db (words, chunks, chunk_words, srt_exports, srt_segments)
    ↓  Configurable export
*.srt (full, no-fillers, short-segments, custom filters)
```

### 3.2 Key Dependencies

From the working venv at `patreon/videos/002-sqrt/venv/`:

| Package | Version | Purpose |
|---------|---------|---------|
| torch | 2.11.0+cpu | Tensor operations |
| torchaudio | (bundled) | Audio I/O |
| nemo_toolkit[asr] | git@e9906001 | ASR model + inference |
| soundfile | 0.13.1 | Audio metadata |
| omegaconf | 2.3.0 | Config management |
| pytorch-lightning | 2.6.1 | Model framework |
| ffmpeg | (system) | Audio preprocessing |

### 3.3 Dagger Pattern Reference

The go-go-golems `cmd/build-web` pattern (Glazed, Smailnail) provides a proven template:

```go
// From glazed/cmd/build-web/main.go
func buildWithDagger(webPath, outPath, pnpmVersion, builderImage string) error {
    ctx := context.Background()
    client, err := dagger.Connect(ctx)
    // ...
    pnpmStore := client.CacheVolume("repo-ui-pnpm-store")
    ctr := client.Container().From(builderImage).
        WithMountedCache("/pnpm/store", pnpmStore).
        WithDirectory("/src", webDir).
        WithExec([]string{"pnpm", "install"}).
        WithExec([]string{"pnpm", "build"})
    _, err = ctr.Directory("/src/dist").Export(ctx, outPath)
    return err
}
```

**Key patterns to reuse:**
1. `dagger.Connect(ctx)` with optional `dagger.WithLogOutput(os.Stdout)`
2. `client.CacheVolume("name")` for caching large downloads
3. `.WithMountedCache()` to mount cache into container
4. Local fallback when Dagger unavailable
5. `findRepoRoot()` to locate working directory

### 3.4 Observed Issues from Previous Work

From the TRANSCRIPT-PIPELINE investigation diary:

1. **OOM on full audio**: `transcribe_to_srt.py` crashes with exit code 137 on 27.7-min audio. Only `transcript_db.py` (60s chunking) works.
2. **Python version sensitivity**: Only Python 3.13.2 venv works; 3.11.3 has torch/torchvision conflicts.
3. **Heavy dependencies**: NeMo + torch install is ~3 GB and takes 10+ minutes.
4. **Model download**: 1.2 GB on first run, stored in `~/.cache/huggingface/`.

---

## 4. Gap Analysis

| Gap | Impact | Solution in This Design |
|-----|--------|------------------------|
| No containerized pipeline | Requires exact Python env locally | Python container built once, cached |
| No model caching strategy | Redownloads 1.2 GB model | Dagger `CacheVolume` for HF cache |
| No Go entrypoint | Can't compose from Go toolchain | `cmd/transcribe/main.go` CLI |
| No structured output contract | Ad-hoc file I/O | Defined output directory with SQLite + SRT |
| No Dockerfile versioning | Reproducibility risk | Pinned Python base + requirements.txt in repo |
| No GPU option | CPU-only, slow | Extensible architecture for GPU container variant |

---

## 5. Proposed Architecture

### 5.1 High-Level Design

```
┌─────────────────────────────────────────────────────────────┐
│  cmd/transcribe (Go CLI)                                     │
│                                                              │
│  ┌──────────────┐  ┌─────────────────┐  ┌───────────────┐  │
│  │ ffmpeg       │  │ Python NeMo     │  │ Output        │  │
│  │ Container    │──│ Container       │──│ Collection    │  │
│  │ (audio →16k) │  │ (ASR chunks)    │  │ (DB + SRT)    │  │
│  └──────────────┘  └─────────────────┘  └───────────────┘  │
│         ↑                  ↑                    ↓           │
│    audio.wav         CacheVolume          ./out/            │
│    (host mount)      (HF model)       transcript.db        │
│                                        *.srt                │
└─────────────────────────────────────────────────────────────┘
```

### 5.2 Project Structure

```
transcription-go/
├── cmd/
│   └── transcribe/
│       └── main.go            # CLI entrypoint (cobra)
├── internal/
│   ├── pipeline/
│   │   └── pipeline.go        # Dagger orchestration logic
│   ├── python/
│   │   └── runner.go          # Python container setup + exec
│   └── ffmpeg/
│       └── convert.go         # Audio conversion container
├── scripts/
│   ├── transcribe.py          # Python transcription script (runs in container)
│   └── requirements.txt       # Pinned Python deps for container
├── ttmp/                      # docmgr workspace
├── go.mod
├── go.sum
├── Makefile
└── .gitignore
```

### 5.3 Container Strategy

#### Container 1: Audio Conversion (ffmpeg)

- **Base image**: `jrottenberg/ffmpeg:6-alpine` (lightweight, ~80 MB)
- **Purpose**: Convert any audio to 16 kHz mono WAV
- **Input**: Host audio file mounted via Dagger
- **Output**: Converted WAV in container filesystem

#### Container 2: ASR Transcription (Python + NeMo)

- **Base image**: `python:3.11-slim-bookworm` (~150 MB)
- **Purpose**: Install deps, run Nemotron ASR, produce SQLite + SRT
- **Cache**: `CacheVolume("transcription-hf-cache")` mounted at `/root/.cache/huggingface`
- **Cache**: `CacheVolume("transcription-pip-cache")` mounted at `/root/.cache/pip`
- **Input**: Converted WAV from Container 1 (passed as Dagger directory/artifact)
- **Output**: SQLite DB + SRT files exported to host

**Critical**: pip install of torch + NeMo is expensive (~3 GB, 5+ minutes). Use a Dagger `CacheVolume` for pip cache and a separate one for the HuggingFace model cache. On warm cache, the only cost is the ASR inference time.

### 5.4 The Python Script (`scripts/transcribe.py`)

The containerized Python script is a simplified version of `transcript_db.py` that:

1. Takes CLI args: `--input <wav>` `--output-dir <dir>` `--no-fillers` `--model <name>`
2. Does NOT do ffmpeg conversion (already done in Container 1)
3. Processes in 60-second chunks (OOM prevention)
4. Stores words in SQLite at `<output-dir>/transcript.db`
5. Exports SRT files to `<output-dir>/`
6. Writes a JSON manifest of outputs for the Go caller to parse

```python
# scripts/transcribe.py — runs inside Dagger container
#!/usr/bin/env python3
"""
Transcription script for containerized execution.
Reads 16kHz mono WAV, produces SQLite + SRT outputs.
"""
import argparse, json, os, sqlite3, subprocess, sys

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", required=True, help="16kHz mono WAV path")
    parser.add_argument("--output-dir", required=True, help="Output directory")
    parser.add_argument("--model", default="nvidia/nemotron-speech-streaming-en-0.6b")
    parser.add_argument("--chunk-size", type=int, default=60, help="Seconds per chunk")
    parser.add_argument("--export-variants", action="store_true", help="Export SRT variants")
    args = parser.parse_args()
    
    os.makedirs(args.output_dir, exist_ok=True)
    # ... (chunked ASR, SQLite storage, SRT export)
    
    # Write manifest for Go caller
    manifest = {
        "db_path": os.path.join(args.output_dir, "transcript.db"),
        "srt_files": [...],
        "word_count": len(all_words),
        "duration": duration,
    }
    with open(os.path.join(args.output_dir, "manifest.json"), "w") as f:
        json.dump(manifest, f, indent=2)
    
if __name__ == "__main__":
    main()
```

### 5.5 Go CLI Interface

```bash
# Basic usage
transcribe --input recording.wav --output-dir ./out/

# With options
transcribe \
  --input recording.wav \
  --output-dir ./out/ \
  --model nvidia/nemotron-speech-streaming-en-0.6b \
  --chunk-size 60 \
  --export-variants \
  --verbose

# Output:
#   out/transcript.db        — SQLite with word-level timestamps
#   out/full.srt             — Full transcript
#   out/no_fillers.srt       — Filler words removed
#   out/manifest.json        — Machine-readable output manifest
```

### 5.6 Dagger Pipeline Flow (Go)

```go
func runPipeline(ctx context.Context, inputPath, outputDir string, opts Options) error {
    client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
    if err != nil {
        return fmt.Errorf("connect dagger: %w", err)
    }
    defer client.Close()

    // Step 1: Audio conversion in ffmpeg container
    converted := convertAudio(client, inputPath)

    // Step 2: Transcription in Python container
    hfCache := client.CacheVolume("transcription-hf-cache")
    pipCache := client.CacheVolume("transcription-pip-cache")

    output := client.Container().
        From("python:3.11-slim-bookworm").
        WithMountedCache("/root/.cache/huggingface", hfCache).
        WithMountedCache("/root/.cache/pip", pipCache).
        WithDirectory("/pipeline/scripts", client.Host().Directory("scripts")).
        WithDirectory("/pipeline/input", converted).
        WithWorkdir("/pipeline").
        WithExec([]string{"pip", "install", "--no-cache-dir", "-r", "scripts/requirements.txt"}).
        WithExec([]string{"python", "scripts/transcribe.py",
            "--input", "/pipeline/input/audio_16k_mono.wav",
            "--output-dir", "/pipeline/output",
            "--export-variants"}).
        Directory("/pipeline/output")

    // Step 3: Export results to host
    _, err = output.Export(ctx, outputDir)
    return err
}
```

---

## 6. Key Design Decisions

### D1: Two-Container Pipeline vs. Single Container

**Decision**: Use two containers (ffmpeg + Python).

**Rationale**: 
- ffmpeg image is ~80 MB and purpose-built; combining it with the ~1.5 GB Python image would waste bandwidth on rebuilds.
- The ffmpeg step is stateless and fast (< 5 seconds); the Python step is stateful (model cache) and slow (minutes).
- Separation allows independent caching and rebuilding.
- Alternative: Use `python:3.11` + install ffmpeg via apt. Rejected because it adds apt overhead on every container build.

### D2: Inline Python Script vs. Pre-built Docker Image

**Decision**: Mount the Python script from host into the container each run. Do NOT pre-build a custom Docker image.

**Rationale**:
- Following the `cmd/build-web` pattern: source is mounted, deps installed per-run (with pip cache).
- Avoids Docker registry management and image versioning.
- The pip `CacheVolume` makes warm runs fast (deps already cached).
- The HF `CacheVolume` avoids re-downloading the 1.2 GB model.
- If build time becomes problematic, a pre-built image can be introduced later without changing the API.

### D3: SQLite as Output Format

**Decision**: Output SQLite database as the primary transcript store, with SRT as export format.

**Rationale**:
- Proven by the existing Python pipeline.
- SQLite is self-contained, queryable, and supports the word/chunk/export schema.
- Enables re-exporting SRTs with different filters without re-transcribing.
- Easy to consume from Go via `modernc.org/sqlite` (pure Go, no CGO needed).

### D4: Manifest File for Go-Python Contract

**Decision**: The Python script writes a `manifest.json` with output paths and metadata.

**Rationale**:
- Provides a structured contract between the containerized Python and the Go host.
- Go can parse the manifest to report results and decide next steps.
- More robust than parsing stdout or assuming output file names.

### D5: Local Fallback

**Decision**: If Dagger is unavailable, fall back to running Python directly.

**Rationale**:
- Follows the `cmd/build-web` pattern (Dagger-first, local fallback).
- Useful during development or on systems without Docker.
- Requires Python + deps installed locally (not always available, but graceful degradation).

---

## 7. Phased Implementation Plan

### Phase 1: Minimal Working Pipeline (MVP)

**Goal**: `go run ./cmd/transcribe --input audio.wav --output-dir ./out/` works end-to-end.

**Files to create:**

| File | Purpose |
|------|---------|
| `go.mod` | Module definition |
| `cmd/transcribe/main.go` | CLI entrypoint with cobra flags |
| `internal/pipeline/pipeline.go` | Dagger pipeline orchestration |
| `scripts/transcribe.py` | Container Python script (adapted from `transcript_db.py`) |
| `scripts/requirements.txt` | Pinned Python dependencies |
| `Makefile` | `build`, `run`, `test` targets |

**Steps:**

1. Initialize Go module: `go mod init github.com/go-go-golems/transcription-go`
2. Add dependencies: `dagger.io/dagger`, `github.com/spf13/cobra`
3. Write `cmd/transcribe/main.go` — parse flags, call pipeline
4. Write `internal/pipeline/pipeline.go` — two-stage Dagger pipeline
5. Write `scripts/transcribe.py` — extracted from `transcript_db.py`, simplified for container use
6. Write `scripts/requirements.txt` — pinned deps from working venv
7. Test with rabbit-hole recording audio file
8. Validate output matches existing `audio_transcript.db`

**Estimated effort**: 2–3 hours

### Phase 2: Polish and Options

**Goal**: Production-ready CLI with all export options.

**Additions:**
- `--format srt|vtt|txt` output format selection
- `--no-fillers` flag
- `--chunk-size` flag
- `--model` flag for alternative ASR models
- `--verbose` flag for Dagger log output
- Progress reporting (parse Python stdout for chunk progress)
- Proper error handling and cleanup

**Estimated effort**: 1–2 hours

### Phase 3: Integration and Library API

**Goal**: Usable as a Go library, not just a CLI.

**Additions:**
- `internal/pipeline` becomes `pkg/pipeline` (exported API)
- `pipeline.Transcribe(ctx, input, opts)` returns structured result
- Option to keep containers running for batch processing
- Integration tests with test audio

**Estimated effort**: 2–3 hours

### Phase 4: Performance and GPU (Future)

**Potential enhancements:**
- Pre-built Docker image with deps pre-installed (avoid pip install time)
- GPU container variant (`nvidia/cuda` base image)
- Parallel chunk processing
- Streaming progress via Dagger events

---

## 8. Test Strategy

### 8.1 Unit Tests

- `internal/pipeline/pipeline_test.go`: Mock Dagger client, verify container setup
- `cmd/transcribe/main_test.go`: CLI flag parsing

### 8.2 Integration Tests

- Run pipeline on test audio (10-second clip from rabbit-hole recording)
- Verify SQLite output has expected schema and row count
- Verify SRT output is parseable
- Verify manifest.json is valid
- Verify filler filtering removes expected words

### 8.3 Validation Test

- Run on full rabbit-hole-2 recording
- Compare output `transcript.db` word count against known value (4,248 words)
- Compare SRT output against existing `no_fillers.srt` from TRANSCRIPT-PIPELINE

### 8.4 Test Audio

Available at:
```
/home/manuel/code/wesen/2026-04-09--screencast-studio/recordings/rabbit-hole-2026-04-10--2/audio-mix.wav
```

---

## 9. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| NeMo pip install fails in container | Medium | High | Pin exact git commit; test in clean container |
| Nemotron model download fails/breaks | Low | High | CacheVolume + retry logic; fallback to local model path |
| OOM during transcription | Low | Medium | 60s chunking proven to work; make chunk size configurable |
| Dagger version incompatibility | Low | Medium | Pin dagger.io/dagger to tested version |
| Container build slow on first run | High | Low | CacheVolume for pip; document warm-cache experience |
| Python 3.11 vs 3.13 compatibility | Medium | Medium | Test with 3.11 first (more common in Docker); fall back to 3.13 if needed |

---

## 10. Alternatives Considered

### A1: Pre-built Docker Image

Build a Docker image with all Python deps pre-installed, push to registry, pull in Dagger.

- **Pros**: Faster cold start (no pip install each run)
- **Cons**: Requires Docker registry, image versioning, CI pipeline for image builds
- **Verdict**: Defer to Phase 4. The `CacheVolume` approach is sufficient for now.

### A2: Python Dagger SDK

Use Dagger's Python SDK instead of Go.

- **Pros**: Direct access to NeMo/torch in the same language
- **Cons**: Doesn't compose with Go toolchain; user needs Python locally
- **Verdict**: Rejected — the whole point is a Go-native pipeline.

### A3: Whisper.cpp via CGO

Use whisper.cpp compiled as a C library, called from Go via CGO.

- **Pros**: No container needed, native Go
- **Cons**: Different model (whisper vs Nemotron), CGO complexity, GPU support harder
- **Verdict**: Interesting future option but different model quality characteristics.

### A4: Subprocess Python

Shell out to Python from Go (no Dagger).

- **Pros**: Simple
- **Cons**: Requires local Python + all deps; fragile
- **Verdict**: This is the current state. Dagger eliminates this dependency.

---

## 11. Open Questions

1. **Python 3.11 vs 3.13**: The working venv uses 3.13.2. Docker `python:3.11-slim` is more common. Need to test if NeMo works on 3.11.
2. **Warm-cache performance**: How fast is a re-run with pip + HF caches warm? Expect < 1 min for setup + transcription time.
3. **Container size**: Python + torch + NeMo install is ~3 GB in the container. Is this acceptable for Docker Desktop users?
4. **Batch processing**: Should the CLI support multiple input files in one invocation?
5. **Output format extensibility**: Should VTT/TXT be added now or later?

---

## 12. References

| Ref | Location | Purpose |
|-----|----------|---------|
| Original handoff doc | `TRANSCRIPT-PIPELINE/design-doc/02-dagger-docker-handoff.md` | Requirements and Python reference |
| Working Python pipeline | `TRANSCRIPT-PIPELINE/scripts/transcript_db.py` | Chunked ASR + SQLite + SRT |
| Glazed build-web | `go-go-golems/glazed/cmd/build-web/main.go` | Dagger Go pattern reference |
| Smailnail build-web | `corporate-headquarters/smailnail/cmd/build-web/main.go` | Dagger Go pattern reference |
| Go-web-dagger skill | `~/.pi/agent/skills/go-web-dagger-pnpm-build/` | Dagger build pattern documentation |
| TRANSCRIPT-PIPELINE diary | `TRANSCRIPT-PIPELINE/reference/02-investigation-diary.md` | Previous work context |
| Test audio | `screencast-studio/recordings/rabbit-hole-2026-04-10--2/audio-mix.wav` | Validation test input |
| Existing transcript DB | `TRANSCRIPT-PIPELINE/sources/audio_transcript.db` | Expected output reference |
