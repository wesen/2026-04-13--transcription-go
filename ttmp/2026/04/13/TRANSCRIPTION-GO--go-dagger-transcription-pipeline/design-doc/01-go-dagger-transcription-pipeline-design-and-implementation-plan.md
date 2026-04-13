---
Title: "Go + Dagger Transcription Pipeline - Design and Implementation Plan"
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
    - Path: ../../../../../2026-04-09--screencast-studio/ttmp/2026/04/13/TRANSCRIPT-PIPELINE--setting-up-an-analysis-pipeline-for-transcripts/design-doc/02-dagger-docker-handoff.md
      Note: Original handoff specification with Python reference and Docker considerations
    - Path: ../../../../../2026-04-09--screencast-studio/ttmp/2026/04/13/TRANSCRIPT-PIPELINE--setting-up-an-analysis-pipeline-for-transcripts/scripts/transcript_db.py
      Note: Working Python pipeline - chunked ASR
    - Path: ../../../../../go-go-golems/glazed/cmd/build-web/main.go
      Note: Reference Dagger Go implementation with CacheVolume pattern and local fallback
    - Path: ../../../../../corporate-headquarters/smailnail/cmd/build-web/main.go
      Note: Second reference Dagger Go build with CacheVolume
ExternalSources: []
Summary: "Revised design: Go CLI handles ffmpeg conversion and all output formatting locally. A long-running Python ASR server runs as a Dagger Service, keeping the Nemotron model in memory between transcriptions. Go sends audio chunks over HTTP, gets back JSON word timestamps, then writes SQLite/SRT/VTT/TXT entirely in Go."
LastUpdated: 2026-04-13T00:00:00Z
WhatFor: "Reproducible, fast audio transcription pipeline. The ASR server stays warm so repeated transcriptions skip model loading. All format logic lives in Go."
WhenToUse: "When you need to transcribe audio files using NVIDIA Nemotron ASR from a Go toolchain, with minimal cold-start overhead on repeated runs."
---

# Go + Dagger Transcription Pipeline — Design and Implementation Plan

> **Revision 2** — Architecture revised based on three simplifications:
> 1. **FFmpeg in Go** instead of a separate Dagger container
> 2. **Long-running ASR server** in the Python container (model stays in memory)
> 3. **All output formatting in Go** (SRT, VTT, TXT, SQLite) — Python only returns raw word timestamps

---

## 1. Executive Summary

A **Go CLI** (`transcribe`) that uses **Dagger** to manage a long-running **Python ASR microservice** inside a container. The pipeline:

1. **Converts** audio to 16 kHz mono WAV locally via Go (shelling out to ffmpeg).
2. **Starts** a Dagger `Service` — a Python FastAPI server with Nemotron 0.6B loaded in memory.
3. **Sends** 60-second audio chunks over HTTP to the server, receives JSON word-level timestamps.
4. **Formats** output entirely in Go: SQLite DB, SRT, VTT, TXT — with filler word filtering.

The ASR server **stays warm** between transcriptions. On repeated runs, model loading (~30 seconds) only happens once. The Go CLI tunnels to the Dagger service via `Host.Tunnel()`.

### Key Properties

| Property | Detail |
|----------|--------|
| **Runtime** | ~5 min for 30-min audio (cold); ~4 min (warm, model already loaded) |
| **Model** | NVIDIA Nemotron 0.6B ASR (~1.2 GB, cached in Dagger `CacheVolume`) |
| **Host deps** | Go 1.21+, Dagger v0.20+, ffmpeg (system) |
| **Output** | SQLite DB + SRT/VTT/TXT (all generated in Go) |
| **Architecture** | Go CLI ↔ HTTP ↔ Dagger Service (Python ASR server) |

---

## 2. Problem Statement and Scope

### 2.1 Problem

The existing Python-based transcription pipeline works but has operational problems:

1. **Fragile local environment**: Requires Python 3.13.2 with a specific venv containing torch, torchaudio, NeMo, soundfile, omegaconf, and pytorch-lightning.
2. **Not composable**: Cannot be called from Go code or integrated into a larger Go toolchain.
3. **Model reload overhead**: Each invocation reloads the 1.2 GB model (~30 seconds wasted).
4. **Output logic entangled with inference**: The Python script mixes ASR inference, SQLite storage, SRT formatting, and filler filtering — making it hard to extend output formats.

### 2.2 Scope

**In scope:**
- Go CLI that takes audio input, produces SQLite + SRT/VTT/TXT output
- Dagger Service running a Python FastAPI server with Nemotron ASR
- Model caching via Dagger `CacheVolume` (survives container restarts)
- Audio conversion via Go (shell out to host ffmpeg)
- All output formatting in Go (SRT, VTT, TXT, SQLite, filler filtering)
- Server lifecycle management (start, health check, reuse across runs)

**Out of scope:**
- GPU acceleration (future)
- Real-time streaming transcription
- Multi-speaker diarization

---

## 3. Current-State Analysis

### 3.1 Working Python Pipeline (Reference)

The reference implementation at `TRANSCRIPT-PIPELINE/scripts/transcript_db.py` (486 lines) does everything in Python:

```
audio-mix.wav → ffmpeg → 16kHz WAV → 60s chunks → Nemotron ASR → SQLite → SRT
```

**Key observations for the revised design:**
- The Python code has three distinct responsibilities: (a) audio preprocessing, (b) ASR inference with word timestamps, (c) output formatting. These can be cleanly separated.
- The ASR inference part is the only thing that *must* run in Python (NeMo dependency).
- Audio preprocessing is trivially done by shelling out to ffmpeg from Go.
- Output formatting (SRT, VTT, TXT, SQLite) is straightforward string/file manipulation — no Python needed.

### 3.2 Dagger Service API (v0.20.5)

Verified in the Dagger Go SDK at `dagger.io/dagger@v0.20.5`:

```go
// Turn a container into a long-running service
service := container.AsService()

// Tunnel from host to service
tunnel := client.Host().Tunnel(service)

// Get the endpoint URL
addr, _ := tunnel.Endpoint(ctx, dagger.ServiceEndpointOpts{Port: 8000})
// addr = "127.0.0.1:34567" (random host port forwarded to container:8000)
```

This is the key API that enables the server architecture. The Go CLI can:
1. Define a container with the Python server
2. Convert it to a `Service` via `AsService()`
3. Create a host tunnel via `Host().Tunnel()`
4. Get the tunnel endpoint and send HTTP requests to it

### 3.3 Dagger Pattern Reference

From `go-go-golems/glazed/cmd/build-web/main.go` — the existing Dagger pattern:

```go
client, err := dagger.Connect(ctx)
pnpmStore := client.CacheVolume("repo-ui-pnpm-store")
ctr := client.Container().From("node:22").
    WithMountedCache("/pnpm/store", pnpmStore).
    WithDirectory("/src", webDir).
    WithExec([]string{"pnpm", "install"}).
    WithExec([]string{"pnpm", "build"})
```

**What we reuse:** `CacheVolume` for pip + HuggingFace caches, `WithDirectory` for mounting scripts.
**What's new:** `AsService()` + `Host().Tunnel()` for the long-running server pattern.

---

## 4. Revised Architecture

### 4.1 Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│  Go CLI (cmd/transcribe)                                        │
│                                                                  │
│  ┌────────────┐     ┌─────────────────────────────────┐         │
│  │ ffmpeg     │     │ Dagger Service (Python FastAPI)  │         │
│  │ (local)    │     │                                  │         │
│  │ audio →    │     │  ┌──────────────────────────┐   │         │
│  │ 16kHz WAV  │────→│  │ Nemotron 0.6B (in mem)   │   │         │
│  └────────────┘     │  │                          │   │         │
│                      │  │ POST /transcribe         │   │         │
│  ┌────────────┐     │  │   ← audio chunk (WAV)    │   │         │
│  │ Output     │     │  │   → word timestamps JSON  │   │         │
│  │ Formatter  │←────│  └──────────────────────────┘   │         │
│  │ (Go)       │     │                                  │         │
│  │            │     │  CacheVolumes:                   │         │
│  │ - SQLite   │     │    /root/.cache/huggingface      │         │
│  │ - SRT      │     │    /root/.cache/pip              │         │
│  │ - VTT      │     └─────────────────────────────────┘         │
│  │ - TXT      │             ↑ Host.Tunnel                       │
│  └────────────┘             ↓                                    │
│                         127.0.0.1:RANDOM                         │
│                              ↓                                   │
│                     ./out/transcript.db                          │
│                     ./out/full.srt                               │
│                     ./out/no_fillers.vtt                         │
│                     ./out/transcript.txt                         │
└─────────────────────────────────────────────────────────────────┘
```

### 4.2 Project Structure

```
transcription-go/
├── cmd/
│   └── transcribe/
│       └── main.go                # CLI entrypoint (cobra)
├── internal/
│   ├── asr/
│   │   ├── client.go              # HTTP client for the ASR server
│   │   └── client_test.go
│   ├── convert/
│   │   ├── ffmpeg.go              # Shell out to ffmpeg for audio conversion
│   │   └── ffmpeg_test.go
│   ├── output/
│   │   ├── sqlite.go              # SQLite writer (word-level storage)
│   │   ├── srt.go                 # SRT formatter
│   │   ├── vtt.go                 # WebVTT formatter
│   │   ├── txt.go                 # Plain text formatter
│   │   ├── filler.go              # Filler word detection + filtering
│   │   └── output_test.go
│   └── server/
│       └── dagger.go              # Dagger Service lifecycle management
├── server/
│   ├── server.py                  # FastAPI ASR server (runs in container)
│   └── requirements.txt           # Pinned Python deps
├── ttmp/                          # docmgr workspace
├── go.mod
├── go.sum
├── Makefile
└── .gitignore
```

### 4.3 Component Responsibilities

| Component | Language | Responsibility |
|-----------|----------|---------------|
| `cmd/transcribe` | Go | CLI flags, orchestration, server lifecycle |
| `internal/convert` | Go | Shell out to ffmpeg: `ffmpeg -ar 16000 -ac 1` |
| `internal/server` | Go | Start Dagger service, tunnel, health check, stop |
| `internal/asr` | Go | HTTP client: chunk audio, POST to server, parse JSON |
| `internal/output` | Go | SQLite storage, SRT/VTT/TXT formatting, filler filtering |
| `server/server.py` | Python | Load Nemotron model once, serve `/transcribe` endpoint |

---

## 5. The ASR Server (`server/server.py`)

The Python container runs a **FastAPI server** that loads the Nemotron model once on startup and serves transcription requests.

### 5.1 Server API

```
GET  /health                  → {"status": "ok", "model_loaded": true}
POST /transcribe              → word-level timestamps
     Body (multipart/form-data):
       file: <WAV audio chunk>
       chunk_start: <float, seconds offset for timestamp adjustment>
     Response (JSON):
       {
         "words": [
           {"word": "hello", "start": 0.32, "end": 0.78, "confidence": 0.98},
           ...
         ],
         "duration": 58.4,
         "chunk_index": 0
       }
POST /transcribe/full         → full file transcription (server handles chunking)
     Body (multipart/form-data):
       file: <WAV audio, any length>
       chunk_size: <int, default 60>
     Response (JSON):
       {
         "words": [...],
         "total_duration": 1664.8,
         "chunk_count": 28,
         "word_count": 4248
       }
```

### 5.2 Server Implementation Sketch

```python
# server/server.py
from fastapi import FastAPI, UploadFile
import nemo.collections.asr as nemo_asr
from omegaconf import OmegaConf, open_dict
import soundfile as sf
import tempfile, subprocess

app = FastAPI()
asr_model = None
time_stride = None

@app.on_event("startup")
async def load_model():
    global asr_model, time_stride
    asr_model = nemo_asr.models.ASRModel.from_pretrained(
        "nvidia/nemotron-speech-streaming-en-0.6b"
    )
    asr_model = asr_model.cpu().eval()
    decoding_cfg = asr_model.cfg.decoding
    with open_dict(decoding_cfg):
        decoding_cfg.preserve_alignments = True
        decoding_cfg.compute_timestamps = True
        decoding_cfg.word_separator = " "
        asr_model.change_decoding_strategy(decoding_cfg)
    time_stride = 8 * asr_model.cfg.preprocessor.window_stride

@app.get("/health")
async def health():
    return {"status": "ok", "model_loaded": asr_model is not None}

@app.post("/transcribe")
async def transcribe_chunk(file: UploadFile, chunk_start: float = 0.0):
    # Save uploaded chunk to temp file
    with tempfile.NamedTemporaryFile(suffix=".wav") as tmp:
        tmp.write(await file.read())
        tmp.flush()
        hypotheses = asr_model.transcribe([tmp.name], return_hypotheses=True, batch_size=1)
    
    words = extract_words(hypotheses, chunk_start)
    return {"words": words, "chunk_index": 0}

@app.post("/transcribe/full")
async def transcribe_full(file: UploadFile, chunk_size: int = 60):
    # Save full audio
    with tempfile.NamedTemporaryFile(suffix=".wav", delete=False) as tmp:
        tmp.write(await file.read())
        tmp.flush()
        info = sf.info(tmp.name)
        duration = info.duration
    
    all_words = []
    for start in range(0, int(duration), chunk_size):
        end = min(start + chunk_size + 2, duration)
        chunk_tmp = tempfile.mktemp(suffix=".wav")
        subprocess.run(["ffmpeg", "-y", "-i", tmp.name, "-ss", str(start),
                       "-t", str(end - start), "-c", "copy", chunk_tmp],
                      capture_output=True)
        hyps = asr_model.transcribe([chunk_tmp], return_hypotheses=True, batch_size=1)
        all_words.extend(extract_words(hyps, start))
        os.remove(chunk_tmp)
    
    return {
        "words": all_words,
        "total_duration": duration,
        "chunk_count": len(range(0, int(duration), chunk_size)),
        "word_count": len(all_words)
    }

def extract_words(hypotheses, chunk_start):
    # Extract word timestamps from hypothesis (same logic as transcript_db.py)
    words = []
    if hypotheses and hypotheses[0] and hasattr(hypotheses[0], 'timestamp'):
        hyp = hypotheses[0]
        if hyp.timestamp and 'word' in hyp.timestamp:
            for stamp in hyp.timestamp['word']:
                word = stamp.get('word', '')
                word_start = stamp.get('start_offset', 0) * time_stride + chunk_start
                word_end = stamp.get('end_offset', 0) * time_stride + chunk_start
                if word.strip():
                    words.append({
                        "word": word.strip(),
                        "start": round(word_start, 3),
                        "end": round(word_end, 3),
                    })
    return words
```

### 5.3 Requirements (`server/requirements.txt`)

```
--extra-index-url https://download.pytorch.org/whl/cpu
torch==2.11.0
torchaudio
soundfile==0.13.1
omegaconf==2.3.0
pytorch-lightning==2.6.1
fastapi==0.115.0
uvicorn[standard]==0.34.0
python-multipart==0.0.20
git+https://github.com/NVIDIA/NeMo.git@e99060017bafe5e6a7443c486ee259f81b9049e8#egg=nemo_toolkit[asr]
```

---

## 6. Go Implementation Details

### 6.1 Dagger Service Lifecycle

```go
// internal/server/dagger.go
package server

type ASRServer struct {
    client   *dagger.Client
    service  *dagger.Service
    endpoint string
}

func Start(ctx context.Context) (*ASRServer, error) {
    client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
    if err != nil {
        return nil, fmt.Errorf("connect dagger: %w", err)
    }

    hfCache := client.CacheVolume("transcription-hf-cache")
    pipCache := client.CacheVolume("transcription-pip-cache")

    scriptsDir := client.Host().Directory("server",
        dagger.HostDirectoryOpts{Exclude: []string{"__pycache__"}})

    // Build the container with deps installed
    ctr := client.Container().
        From("python:3.11-slim-bookworm").
        WithMountedCache("/root/.cache/huggingface", hfCache).
        WithMountedCache("/root/.cache/pip", pipCache).
        WithEnvVariable("PNPM_HOME", "/pnpm"). // not needed, remove
        WithDirectory("/app", scriptsDir).
        WithWorkdir("/app").
        WithExec([]string{"pip", "install", "-r", "requirements.txt"}).
        WithExposedPort(8000).
        WithExec([]string{"uvicorn", "server:app", "--host", "0.0.0.0", "--port", "8000"})

    // Convert to a long-running Dagger service
    service := ctr.AsService()

    // Tunnel from host to the service
    tunnel := client.Host().Tunnel(service)
    endpoint, err := tunnel.Endpoint(ctx, dagger.ServiceEndpointOpts{Port: 8000})
    if err != nil {
        return nil, fmt.Errorf("get tunnel endpoint: %w", err)
    }

    svc := &ASRServer{
        client:   client,
        service:  service,
        endpoint: endpoint,
    }

    // Wait for server to be ready (health check)
    if err := svc.waitReady(ctx); err != nil {
        svc.Stop()
        return nil, err
    }

    return svc, nil
}

func (s *ASRServer) waitReady(ctx context.Context) error {
    for i := 0; i < 60; i++ { // 60 second timeout
        resp, err := http.Get(fmt.Sprintf("http://%s/health", s.endpoint))
        if err == nil && resp.StatusCode == 200 {
            return nil
        }
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(time.Second):
        }
    }
    return fmt.Errorf("server not ready after 60s")
}

func (s *ASRServer) Endpoint() string { return s.endpoint }

func (s *ASRServer) Stop() {
    s.client.Close()
}
```

### 6.2 Audio Conversion (Go)

```go
// internal/convert/ffmpeg.go
package convert

func To16kMono(inputPath, outputPath string) error {
    cmd := exec.Command("ffmpeg", "-y", "-i", inputPath,
        "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", outputPath)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}
```

No Dagger container needed. ffmpeg is ubiquitous on development machines and handles every audio format.

### 6.3 ASR Client (Go)

```go
// internal/asr/client.go
package asr

type Word struct {
    Word       string  `json:"word"`
    Start      float64 `json:"start"`
    End        float64 `json:"end"`
    Confidence float64 `json:"confidence,omitempty"`
}

type TranscribeResponse struct {
    Words       []Word  `json:"words"`
    Duration    float64 `json:"duration"`
    ChunkIndex  int     `json:"chunk_index"`
}

func TranscribeFile(ctx context.Context, endpoint, audioPath string) ([]Word, error) {
    body, buf := &bytes.Buffer{}, &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    
    part, _ := writer.CreateFormFile("file", filepath.Base(audioPath))
    f, _ := os.Open(audioPath)
    io.Copy(part, f)
    f.Close()
    writer.Close()

    req, _ := http.NewRequestWithContext(ctx, "POST",
        fmt.Sprintf("http://%s/transcribe/full", endpoint), body)
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result TranscribeResponse
    json.NewDecoder(resp.Body).Decode(&result)
    return result.Words, nil
}
```

### 6.4 Output Formatting (Go)

All formatting happens in Go. The Python server only returns raw word timestamps.

```go
// internal/output/srt.go
func WriteSRT(words []Word, outputPath string, opts SRTOptions) error {
    segments := buildSegments(words, opts)
    f, err := os.Create(outputPath)
    // ... SRT format: index, timestamp range, text
}

// internal/output/vtt.go
func WriteVTT(words []Word, outputPath string, opts VTTOptions) error {
    // WebVTT format: WEBVTT header, timestamp range, text
}

// internal/output/txt.go
func WriteTXT(words []Word, outputPath string, opts TXTOptions) error {
    // Plain text with optional paragraph breaks
}

// internal/output/sqlite.go
func WriteSQLite(words []Word, outputPath string) error {
    // Create SQLite with words, chunks, chunk_words tables
    // Same schema as Python pipeline for compatibility
}

// internal/output/filler.go
var FillerWords = map[string]bool{
    "um": true, "uh": true, "er": true, "ah": true,
    "like": true, "you know": true, "sort of": true, "kind of": true,
}

func FilterFillers(words []Word) []Word {
    var result []Word
    for _, w := range words {
        if !FillerWords[strings.ToLower(strings.Trim(w.Word, ".,!?"))] {
            result = append(result, w)
        }
    }
    return result
}
```

### 6.5 CLI Wiring

```go
// cmd/transcribe/main.go
func main() {
    var input, outputDir, format string
    var noFillers bool
    
    rootCmd := &cobra.Command{
        Use:   "transcribe --input audio.wav --output-dir ./out/",
        Short: "Transcribe audio using NVIDIA Nemotron ASR",
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx := context.Background()
            
            // 1. Convert audio (local ffmpeg)
            convertedPath := filepath.Join(outputDir, "audio_16k_mono.wav")
            if err := convert.To16kMono(input, convertedPath); err != nil {
                return fmt.Errorf("convert audio: %w", err)
            }
            
            // 2. Start ASR server (Dagger service)
            server, err := asrserver.Start(ctx)
            if err != nil {
                return fmt.Errorf("start server: %w", err)
            }
            defer server.Stop()
            
            // 3. Transcribe (HTTP to server)
            words, err := asr.TranscribeFile(ctx, server.Endpoint(), convertedPath)
            if err != nil {
                return fmt.Errorf("transcribe: %w", err)
            }
            
            // 4. Write outputs (all in Go)
            output.WriteSQLite(words, filepath.Join(outputDir, "transcript.db"))
            
            if noFillers {
                words = output.FilterFillers(words)
            }
            
            for _, fmt := range strings.Split(format, ",") {
                switch fmt {
                case "srt":
                    output.WriteSRT(words, filepath.Join(outputDir, "transcript.srt"), opts)
                case "vtt":
                    output.WriteVTT(words, filepath.Join(outputDir, "transcript.vtt"), opts)
                case "txt":
                    output.WriteTXT(words, filepath.Join(outputDir, "transcript.txt"), opts)
                }
            }
            
            return nil
        },
    }
    
    rootCmd.Flags().StringVar(&input, "input", "", "Input audio file")
    rootCmd.Flags().StringVar(&outputDir, "output-dir", "./out", "Output directory")
    rootCmd.Flags().StringVar(&format, "format", "srt", "Output formats: srt,vtt,txt")
    rootCmd.Flags().BoolVar(&noFillers, "no-fillers", false, "Remove filler words")
    rootCmd.Execute()
}
```

---

## 7. Key Design Decisions

### D1: FFmpeg in Go (not a Dagger container) ✅ REVISED

**Decision**: Shell out to host ffmpeg from Go.

**Rationale**:
- ffmpeg is universally available on Linux development machines.
- The conversion is trivial: `ffmpeg -ar 16000 -ac 1` — < 5 seconds.
- Eliminates an entire Dagger container, reducing complexity.
- No pure-Go alternative handles all audio formats (MP3, M4A, FLAC, etc.) without CGO.
- The Go `exec.Command` call is simpler than defining + managing a Dagger container.

**Tradeoff**: Requires ffmpeg on the host. Acceptable for a developer tool.

### D2: Long-Running ASR Server (Dagger Service) ✅ REVISED

**Decision**: Run the Python ASR as a FastAPI server using `Container.AsService()` + `Host.Tunnel()`.

**Rationale**:
- **Model stays in memory**: Nemotron 0.6B takes ~30 seconds to load and ~4 GB RAM. Keeping it loaded avoids this overhead on every run.
- **Batch friendliness**: Transcribe multiple files without restarting the container.
- **Clean separation**: Python only does ASR inference. Go handles everything else.
- **Dagger API supports it**: `AsService()` + `Tunnel()` is a first-class Dagger pattern (verified in SDK v0.20.5).

**Tradeoff**: Container stays running between invocations. Memory usage is ~4 GB while the server runs. The Go CLI manages the lifecycle.

### D3: All Output Formatting in Go ✅ REVISED

**Decision**: SRT, VTT, TXT, and SQLite generation all happen in Go.

**Rationale**:
- **Single responsibility**: Python = ASR inference only. Go = everything else.
- **Extensibility**: Adding VTT, TXT, or custom formats is pure Go — no Python changes needed.
- **Testability**: Format logic can be unit-tested without containers.
- **No SQLite in container**: The Python container doesn't need `sqlite3`. It just returns JSON.
- **Simpler Python**: The server script is ~100 lines of FastAPI + NeMo — no SQLite, no SRT logic, no filler detection.

### D4: CacheVolume for Model + Pip deps

**Decision**: Two Dagger `CacheVolume`s — one for HuggingFace model cache, one for pip cache.

**Rationale**:
- First run: pip install (~3 GB, 5+ min) + model download (~1.2 GB, 2 min).
- Subsequent runs (warm cache): pip install is instant (cached), model is instant (cached).
- The model stays in memory while the server runs; the cache ensures fast restarts.

### D5: Local Fallback

**Decision**: If Dagger is unavailable, fall back to starting the Python server directly.

**Rationale**: Matches the `cmd/build-web` pattern. Graceful degradation for environments without Docker.

---

## 8. Comparison: v1 vs v2 Architecture

| Aspect | v1 (Original) | v2 (Revised) |
|--------|---------------|---------------|
| FFmpeg conversion | Dagger container | Go `exec.Command` (local ffmpeg) |
| ASR execution | Fire-and-forget `WithExec` | Long-running FastAPI server (`AsService`) |
| Model loading | Every run | Once, stays in memory |
| Output formatting | Python (in container) | Go (locally) |
| Output formats | SRT only | SRT, VTT, TXT, SQLite |
| Filler filtering | Python | Go |
| SQLite generation | Python | Go |
| Go-Python contract | Files on disk | HTTP JSON API |
| Containers | 2 (ffmpeg + Python) | 1 (Python server only) |
| Complexity | Medium | Lower (less container coordination) |
| Batch performance | Cold each time | Warm after first run |

---

## 9. Phased Implementation Plan

### Phase 1: Core Pipeline (MVP)

**Goal**: `go run ./cmd/transcribe --input audio.wav --output-dir ./out/` produces transcript.db + SRT.

**Files:**

| File | Purpose |
|------|---------|
| `go.mod` | Module: `github.com/go-go-golems/transcription-go` |
| `cmd/transcribe/main.go` | CLI entrypoint |
| `internal/convert/ffmpeg.go` | Shell out to ffmpeg |
| `internal/server/dagger.go` | Dagger service lifecycle |
| `internal/asr/client.go` | HTTP client for `/transcribe/full` |
| `internal/output/srt.go` | SRT formatter |
| `internal/output/sqlite.go` | SQLite writer |
| `internal/output/filler.go` | Filler word detection |
| `server/server.py` | FastAPI ASR server |
| `server/requirements.txt` | Python deps |

**Steps:**

1. `go mod init github.com/go-go-golems/transcription-go`
2. Add deps: `dagger.io/dagger@v0.20.5`, `github.com/spf13/cobra`, `modernc.org/sqlite`
3. Write `server/server.py` — FastAPI with `/health` + `/transcribe/full`
4. Write `internal/server/dagger.go` — start Dagger service with tunnel
5. Write `internal/convert/ffmpeg.go` — audio conversion
6. Write `internal/asr/client.go` — HTTP transcription client
7. Write `internal/output/` — SRT + SQLite + filler filtering
8. Wire `cmd/transcribe/main.go`
9. Test with rabbit-hole recording

**Estimated effort**: 3–4 hours

### Phase 2: Additional Formats and Options

**Additions:**
- `--format srt,vtt,txt` (multiple, comma-separated)
- `internal/output/vtt.go` — WebVTT formatter
- `internal/output/txt.go` — plain text formatter
- `--no-fillers` flag
- `--chunk-size` flag
- `--verbose` with progress reporting

**Estimated effort**: 1–2 hours

### Phase 3: Server Management and Batch

**Additions:**
- `transcribe daemon start` / `transcribe daemon stop` — persistent server
- Batch mode: `transcribe --input-dir ./recordings/`
- Server auto-stop after idle timeout
- `--keep-alive` flag to keep server running between invocations

**Estimated effort**: 2–3 hours

### Phase 4: Performance and GPU (Future)

- Pre-built Docker image (skip pip install on cold start)
- GPU container variant
- Parallel chunk processing (send chunks concurrently)
- Streaming progress via chunk-level callbacks

---

## 10. Test Strategy

### 10.1 Unit Tests (No containers needed)

- `internal/convert/ffmpeg_test.go`: Mock exec.Command, verify args
- `internal/output/srt_test.go`: Known words → expected SRT output
- `internal/output/vtt_test.go`: Known words → expected VTT output
- `internal/output/sqlite_test.go`: Known words → SQLite with correct schema
- `internal/output/filler_test.go`: Verify filler detection

### 10.2 Integration Tests (Requires Dagger)

- Start server, transcribe 10-second test clip, verify word count
- Health check endpoint returns `model_loaded: true`
- Server handles multiple requests without restart

### 10.3 Validation Test

- Transcribe full rabbit-hole-2 recording
- Compare word count against known value (4,248 words)
- Verify SQLite schema matches Python pipeline output

### 10.4 Test Audio

```
/home/manuel/code/wesen/2026-04-09--screencast-studio/recordings/rabbit-hole-2026-04-10--2/audio-mix.wav
```

---

## 11. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| NeMo pip install fails in container | Medium | High | Pin exact git commit; test in clean container |
| `Host.Tunnel()` doesn't reach host | Low | High | Fallback: allocate host port, bind directly |
| Server startup slow (pip + model load) | High | Low | CacheVolumes + health check with timeout |
| ffmpeg not on host | Low | Medium | Document requirement; detect and error clearly |
| Python 3.11 vs 3.13 compatibility | Medium | Medium | Test both; pin to whichever works |
| Server OOM on long audio chunks | Low | Medium | Server handles chunking internally (60s chunks) |

---

## 12. Alternatives Considered

### A1: Pure Go (No Python)

Use whisper.cpp via CGO, or a Go-native ASR model.

- **Pros**: No containers, no Python
- **Cons**: Different model quality; CGO complexity; no Nemotron
- **Verdict**: Interesting future option. Current design uses best-available ASR model.

### A2: Pre-built Docker Image

Build image once, push to registry, pull in Dagger.

- **Pros**: Faster cold start
- **Cons**: Registry management, CI overhead
- **Verdict**: Good Phase 4 optimization. `CacheVolume` handles the interim.

### A3: gRPC Instead of HTTP/REST

Use gRPC for the Go-Python communication.

- **Pros**: Type-safe, streaming support
- **Cons**: Adds protobuf dependency, complexity overkill for two endpoints
- **Verdict**: HTTP JSON is sufficient. The API surface is tiny (2 endpoints).

### A4: Fire-and-Forget (v1 Original)

Keep the original v1 architecture (no server).

- **Pros**: Simpler Dagger code (no service lifecycle)
- **Cons**: Model reload on every run (~30s), can't batch, output formatting in Python
- **Verdict**: v2 is strictly better. The server adds minimal complexity and significant performance gains.

---

## 13. Open Questions

1. **Python version**: Test NeMo on Python 3.11 (Docker default). Fall back to 3.13 if needed.
2. **Tunnel reliability**: Does `Host.Tunnel()` work reliably on Linux? Need to verify.
3. **Idle timeout**: Should the server auto-stop after N minutes of inactivity? Or always stay running?
4. **Batch performance**: How much faster is warm-start vs cold-start? Need benchmarks.
5. **SQLite compatibility**: Should the Go-generated SQLite use the exact same schema as the Python pipeline for compatibility?

---

## 14. References

| Ref | Location | Purpose |
|-----|----------|---------|
| Original handoff doc | `TRANSCRIPT-PIPELINE/design-doc/02-dagger-docker-handoff.md` | Requirements and Python reference |
| Working Python pipeline | `TRANSCRIPT-PIPELINE/scripts/transcript_db.py` | Chunked ASR + SQLite + SRT |
| Glazed build-web | `go-go-golems/glazed/cmd/build-web/main.go` | Dagger Go pattern reference |
| Smailnail build-web | `corporate-headquarters/smailnail/cmd/build-web/main.go` | Dagger Go pattern reference |
| Go-web-dagger skill | `~/.pi/agent/skills/go-web-dagger-pnpm-build/` | Dagger build pattern documentation |
| Dagger SDK v0.20.5 | `dagger.io/dagger@v0.20.5/dagger.gen.go` | `AsService()`, `Host.Tunnel()`, `Service.Endpoint()` |
| Test audio | `screencast-studio/recordings/rabbit-hole-2026-04-10--2/audio-mix.wav` | Validation test input |
| Existing transcript DB | `TRANSCRIPT-PIPELINE/sources/audio_transcript.db` | Expected output reference |
