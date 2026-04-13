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

---

## Step 2: Architecture Revision — Server Pattern + Go Output Formatting

The user proposed three simplifications that significantly improve the architecture: (1) do ffmpeg conversion in Go instead of a separate Dagger container, (2) run a long-running ASR server in the Python container to keep the model in memory, and (3) handle all output formatting (SRT, VTT, TXT, SQLite) in Go. I verified the Dagger Go SDK (v0.20.5) supports `Container.AsService()` + `Host.Tunnel()` for the server pattern, and rewrote the design doc from scratch.

### Prompt Context

**User prompt (verbatim):** "can we do the ffmpeg work in go itself? simplifying the setup. Also, would we run a server in the transcription container to make it easier to keep model in memory and all? and then we can save to different output formats in go itself."

**Assistant interpretation:** Three architectural simplifications: eliminate the ffmpeg container, run the Python ASR as a long-running HTTP server (not fire-and-forget), and move all output formatting to Go.

**Inferred user intent:** Reduce complexity (fewer containers), improve performance (model stays warm), and make output format extensibility a Go concern (no Python changes needed for new formats).

### What I did

1. Investigated Dagger Go SDK v0.20.5 for service support — found `Container.AsService()`, `Host.Tunnel()`, `Service.Endpoint()` in `dagger.gen.go`
2. Checked Go audio library ecosystem — concluded shelling out to ffmpeg is the pragmatic choice
3. Designed the FastAPI server API: `GET /health`, `POST /transcribe`, `POST /transcribe/full`
4. Rewrote the entire design doc (v2) with new architecture
5. Added Go implementation sketches for all components: Dagger service lifecycle, ASR HTTP client, ffmpeg conversion, SRT/VTT/TXT/SQLite formatters
6. Added v1 vs v2 comparison table

### Why

The three changes address real pain points:
- **FFmpeg in Go**: One fewer container to manage. ffmpeg is ~5s, not worth a container.
- **ASR server**: Model loading is ~30s. Keeping it in memory means batch jobs don't pay that cost repeatedly.
- **Go output formatting**: Makes adding VTT, TXT, or custom formats trivial — no Python changes needed. Also makes the output logic testable without containers.

### What worked

- The Dagger service API is exactly what we need. `AsService()` + `Host.Tunnel()` creates a tunnel from the host to the container. The Go CLI can just use `http.Get`/`http.Post` to talk to the Python server.
- The Python server script becomes much simpler: ~100 lines of FastAPI + NeMo, no SQLite, no SRT, no filler detection.
- Clean separation of concerns: Python = inference only, Go = orchestration + formatting.

### What didn't work

- No issues during design. The Dagger SDK had all the APIs we needed.

### What I learned

1. `Container.AsService()` turns a Dagger container into a long-running service. Dagger manages the lifecycle.
2. `Host.Tunnel(service)` creates a network tunnel from the host machine to the service, returning a new `Service` with host-reachable endpoint.
3. `Service.Endpoint(ctx)` returns the host:port that the Go CLI can connect to.
4. Pure Go audio libraries (go-audio/wav) exist but don't handle format decoding for non-WAV inputs. Shell out to ffmpeg for universal format support.

### What was tricky to build

The main design question was the Go-Python communication contract. Options considered:
- **Files on disk** (v1): Python writes SQLite + SRT, Go reads them. Simple but couples Go to Python's output format.
- **HTTP JSON API** (v2 chosen): Python returns word-level timestamps as JSON. Go handles everything else. Clean but requires HTTP server in the container.
- **gRPC**: Overkill for 2 endpoints.

Chose HTTP JSON for simplicity and debuggability (can curl the server directly for testing).

### What warrants a second pair of eyes

1. **`Host.Tunnel()` reliability**: This is a relatively new Dagger feature. Need to verify it works on Linux with the installed Dagger v0.20.0 engine + v0.20.5 SDK.
2. **Server chunking strategy**: Should `/transcribe/full` handle chunking server-side (simpler Go client) or should the Go client send individual chunks (more control)? Current design: server handles chunking.
3. **SQLite schema compatibility**: If the Go-generated SQLite needs to match the Python pipeline's schema exactly, we need to verify the schema is replicated correctly.

### What should be done in the future

1. Implement Phase 1 and validate against existing transcript data
2. Test `Host.Tunnel()` on the actual machine — this is the biggest unknown
3. Benchmark cold vs warm server startup
4. Consider adding a `--daemon` mode for persistent server

### Code review instructions

**Files to review:**
- `design-doc/01-go-dagger-transcription-pipeline-design-and-implementation-plan.md` — revised v2 design doc

**How to validate:**
```bash
docmgr doctor --ticket TRANSCRIPTION-GO
```

### Technical details

**Dagger Service API (from dagger.gen.go v0.20.5):**
```go
// Turn container into a service
service := container.AsService()

// Tunnel from host to service
tunnel := client.Host().Tunnel(service)

// Get host-reachable endpoint
addr, _ := tunnel.Endpoint(ctx, dagger.ServiceEndpointOpts{Port: 8000})
// e.g., "127.0.0.1:34567"
```

**FastAPI server endpoints:**
```
GET  /health              → {"status": "ok", "model_loaded": true}
POST /transcribe          → {"words": [{"word": "hello", "start": 0.32, "end": 0.78}]}
POST /transcribe/full     → {"words": [...], "total_duration": 1664.8, "word_count": 4248}
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

**SQLite schema (from transcript_db.py):**
```sql
words(id, word, start_time, end_time, is_filler, is_removed, confidence, chunk_id)
chunks(id, start_time, end_time, text, word_count, source_type)
chunk_words(chunk_id, word_id, position)
srt_exports(id, filename, config, segment_count, word_count)
srt_segments(id, export_id, sequence_num, start_time, end_time, text)
```

---

## Step 3: Eliminate ffmpeg — Pure Go Audio Conversion

The user asked to remove the ffmpeg dependency entirely and do WAV conversion natively in Go. I investigated the Go audio library ecosystem, found `go-audio/wav` (decoder/encoder) and `oov/audio/resampler` (opus-tools quality resampler ported to pure Go), and verified them with a benchmark against ffmpeg on the actual rabbit-hole recording. Pure Go streaming conversion is **2.9s vs ffmpeg's 3.5s** with identical output.

### Prompt Context

**User prompt (verbatim):** "can we do the wav manipulation / conversion in go natively, that way we don't depend on ffmpeg at all."

**Assistant interpretation:** Remove the ffmpeg host dependency. Do WAV → 16kHz mono conversion entirely in Go using pure Go libraries.

**Inferred user intent:** Zero host dependencies beyond Go + Dagger. The binary should be self-contained.

### What I did

1. Investigated Go audio libraries:
   - `github.com/go-audio/wav` v1.1.0 — WAV decoder + encoder (ReadSeeker-based)
   - `github.com/go-audio/audio` v1.0.0 — `IntBuffer`, `Float32Buffer`, `Format` types
   - `github.com/oov/audio/resampler` — Opus-tools resampler ported to pure Go, supports `ProcessFloat64` per-channel
2. Wrote a test program in `/tmp/go-audio-test/` that:
   - Reads WAV via `wav.NewDecoder(r).PCMBuffer(buf)` (streaming in 100ms chunks)
   - Downmixes stereo to mono (average channels)
   - Resamples 48kHz → 16kHz via `resampler.New(1, 48000, 16000, 0)`
   - Writes output via `wav.NewEncoder(w, 16000, 16, 1, 1)`
3. Benchmarked against ffmpeg on the rabbit-hole recording (305 MB, 27.7 min, stereo 48kHz)
4. Verified all screencast recordings are WAV PCM 16-bit stereo 48kHz — pure WAV input is sufficient
5. Updated design doc to v3: replaced ffmpeg conversion with pure Go, updated comparison table, risks

### What worked

- **Streaming approach**: 2.9s (Go streaming) vs 3.5s (ffmpeg). Faster AND no subprocess.
- **File size match**: 53,272,044 bytes (Go) vs 53,272,078 bytes (ffmpeg) — 34 byte diff in WAV headers only.
- **Constant memory**: Streaming in 100ms chunks means ~6.4 KB buffer regardless of input size.
- **The `oov/audio/resampler`** is a direct port of the opus-tools resampler — high quality, pure Go, no CGO.

### What didn't work

- **`FullPCMBuffer()` approach**: Loading the entire file into memory took 101 seconds and used ~1.5 GB RAM. Switching to streaming (`PCMBuffer` with fixed-size buffer) fixed both issues.
- **`resampler.BestQuality`** constant doesn't exist — quality is an int (0 = default). Minor API doc issue.
- **go-audio/mp3, go-audio/flac**: These modules don't exist on the Go module proxy (404 from GitHub). Only WAV is available.

### What I learned

1. `go-audio/wav` supports streaming read/write via `PCMBuffer(buf)` + `Encoder.Write(buf)`. The `FullPCMBuffer()` method loads everything at once — avoid for large files.
2. `oov/audio/resampler` operates per-channel: `ProcessFloat64(channelIndex, in, out)`. For mono, channelIndex=0.
3. The output buffer for resampling needs to be oversized: `int(float64(len(in)) * outRate/inRate) + 256`. The `+256` prevents buffer overflows.
4. All screencast recordings are WAV PCM 16-bit stereo 48kHz — no MP3/M4A/FLAC in sight. The WAV-only limitation of go-audio is perfectly fine.

### What was tricky to build

The `FullPCMBuffer()` trap: the first implementation loaded the entire 305 MB WAV into a `[]int` array (79908000 samples × 8 bytes = ~610 MB), then converted to `[]float64` for resampling (another ~610 MB). This was 101 seconds of allocation + GC pressure. Switching to streaming with 100ms chunks (4800 samples per chunk) reduced memory to ~40 KB and time to 2.9s.

Also: the resampler has internal state (filter history). You must reuse the same `resampler.New()` instance across all chunks — creating a new one per chunk would lose phase continuity and introduce clicks.

### What warrants a second pair of eyes

1. **Resampling quality**: The Go resampler produces byte-identical output size to ffmpeg. But is the audio quality actually equivalent? Should verify by running transcription on both outputs and comparing word counts.
2. **Downmix strategy**: Currently averaging channels. For voice audio this is fine, but is there a case where one channel has the voice and the other doesn't?
3. **Resampler quality parameter**: Using quality=0. Should we use a higher quality setting? The opus-tools docs suggest 0-10 range.

### What should be done in the future

1. Test transcription with Go-converted audio vs ffmpeg-converted audio — verify word counts match
2. Add format detection (check file header) and clear error message for non-WAV inputs
3. Consider adding `go-audio` format decoders if MP3/FLAC support becomes needed

### Code review instructions

**Test program**: `/tmp/go-audio-test/main.go` (can be deleted, but useful for reference)

**How to validate:**
```bash
cd /tmp/go-audio-test && go build -o convert_test .
time ./convert_test \
  /home/manuel/code/wesen/2026-04-09--screencast-studio/recordings/rabbit-hole-2026-04-10--2/audio-mix.wav \
  /tmp/test-output.wav
file /tmp/test-output.wav
# Expected: RIFF (little-endian) data, WAVE audio, Microsoft PCM, 16 bit, mono 16000 Hz
```

### Technical details

**Go audio conversion stack:**
```
github.com/go-audio/wav       v1.1.0  — WAV decoder (ReadSeeker) + encoder (WriteSeeker)
github.com/go-audio/audio     v1.0.0  — IntBuffer, Format, Float32Buffer types
github.com/oov/audio/resampler         — Opus-tools resampler, pure Go
```

**Streaming conversion flow:**
```
Input WAV (48kHz, stereo, 16-bit)
    ↓ wav.Decoder.PCMBuffer (100ms chunks)
IntBuffer{Data: []int, Format: {NumChannels: 2, SampleRate: 48000}}
    ↓ Average channels
[]float64 (mono, 48kHz)
    ↓ resampler.ProcessFloat64(0, in, out)
[]float64 (mono, 16kHz)
    ↓ Clamp + convert to int
IntBuffer{Data: []int, Format: {NumChannels: 1, SampleRate: 16000}}
    ↓ wav.Encoder.Write
Output WAV (16kHz, mono, 16-bit)
```

**Benchmark results:**
```
Input:  audio-mix.wav (305 MB, 27.7 min, 48kHz stereo 16-bit PCM)
Output: 16kHz mono 16-bit PCM WAV (53.3 MB)

Go streaming:   2.9s  → 53,272,044 bytes
ffmpeg:          3.5s  → 53,272,078 bytes
Go FullPCMBuf:  101.5s → 53,272,044 bytes (DO NOT USE)
```
