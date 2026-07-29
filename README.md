# transcription-go

A Go-orchestrated audio transcription pipeline with support for single-file batch transcription, live streaming transcription, and resumable playlist-scale corpus transcription. Uses NVIDIA Nemotron (via Dagger container) on Linux/CUDA and NVIDIA Parakeet TDT 0.6B v3 (via whisper.cpp) on Apple Silicon with Metal GPU acceleration.

## Features

- **Batch transcription** — transcribe a single WAV file to SRT, VTT, TXT, and SQLite
- **Live transcription** — streaming WebSocket-based live transcription with session management
- **Corpus transcription** — playlist-scale batch processing with resume, provenance, and full-text search
- **Metal GPU backend** — Parakeet TDT 0.6B v3 on Apple Silicon via whisper.cpp (98x realtime on M1 Max)
- **Nemotron backend** — NVIDIA Nemotron 0.6B via Dagger-managed FastAPI service (CUDA or CPU)
- **SQLite storage** — word-level timestamps, FTS5 search, chunk derivation, export tracking
- **Multi-format export** — SRT, VTT, TXT with per-video output directories

## Requirements

### Go backend (orchestration)

- Go 1.25+
- ffmpeg and ffprobe (for corpus audio chunking and duration detection)
- Dagger CLI (only for Nemotron backend; not needed for Metal backend)

### Nemotron backend (Linux/CUDA)

- Docker (for Dagger container execution)
- NVIDIA GPU with CUDA support (optional — CPU mode works at ~5x realtime)
- Dagger CLI

### Metal backend (Apple Silicon)

- Apple M1 or later
- [whisper.cpp](https://github.com/ggml-org/whisper.cpp) built with Metal support
- Parakeet TDT 0.6B v3 GGUF model file

## Build

```bash
make build
# or
go build -o transcribe ./cmd/transcribe
```

For Apple Silicon (cross-compile from Linux):

```bash
GOOS=darwin GOARCH=arm64 go build -o transcribe-darwin ./cmd/transcribe
```

## Quick start

### Single file transcription (Nemotron via Dagger)

```bash
./transcribe -i input.wav -o ./out --format srt,vtt,txt,db
```

This starts a Dagger container with the Nemotron ASR server, transcribes the audio file in 60-second chunks with 2-second overlap, and writes SRT, VTT, TXT, and SQLite output to `./out/`.

### Corpus transcription (Nemotron via Dagger)

```bash
./transcribe corpus run \
  --manifest manifest.json \
  --database corpus.db \
  --output-dir exports/ \
  --verbose
```

### Corpus transcription (Parakeet Metal on Apple Silicon)

```bash
./transcribe corpus run \
  --manifest manifest.json \
  --database corpus.db \
  --metal-gpu \
  --metal-backend parakeet \
  --metal-binary ~/code/whisper/whisper.cpp/build/bin/parakeet-cli \
  --metal-model ~/code/whisper/whisper.cpp/models/ggml-parakeet-tdt-0.6b-v3-f16.bin \
  --output-dir exports/ \
  --verbose
```

## Setting up the Metal backend (Apple Silicon)

### 1. Build whisper.cpp with Metal

```bash
git clone https://github.com/ggml-org/whisper.cpp.git
cd whisper.cpp
cmake -B build -DGGML_METAL=ON
cmake --build build -j
```

This produces `build/bin/parakeet-cli` and `build/bin/whisper-cli`.

### 2. Download the Parakeet model

```bash
cd whisper.cpp
pip3 install huggingface_hub
python3 -c "from huggingface_hub import hf_hub_download; \
  hf_hub_download('ggml-org/parakeet-GGUF', 'ggml-parakeet-tdt-0.6b-v3-f16.bin', local_dir='models')"
```

Or download Whisper large-v3-turbo as an alternative:

```bash
sh ./models/download-ggml-model.sh large-v3-turbo
```

### 3. parakeet-cli JSON output patch

The upstream `parakeet-cli` only supports plain text output. The corpus pipeline requires JSON output with word-level timestamps. A patched `parakeet-cli.cpp` is included at:

```
ttmp/2026/07/28/VIDEO-CORPUS-PIPELINE--adapt-transcription-go-for-playlist-video-corpus-pipelines/scripts/parakeet-cli-json-output.cpp
```

To apply it:

```bash
cp ttmp/.../scripts/parakeet-cli-json-output.cpp ~/code/whisper.cpp/examples/parakeet-cli/parakeet-cli.cpp
cd ~/code/whisper.cpp
cmake --build build --target parakeet-cli
```

The patch adds the `-oj` (output JSON) flag with BPE subword accumulation. Without this patch, the Metal backend cannot produce word-level timestamps.

### 4. Verify

```bash
~/code/whisper.cpp/build/bin/parakeet-cli \
  -m ~/code/whisper.cpp/models/ggml-parakeet-tdt-0.6b-v3-f16.bin \
  -f test.wav -oj -of /tmp/test -np
cat /tmp/test.json | python3 -m json.tool | head -20
```

## Corpus pipeline usage

### Manifest format

The manifest is a JSON file declaring the corpus:

```json
{
  "corpus_id": "my-lecture-series",
  "videos": [
    {
      "source_id": "dQw4w9WgXcQ",
      "title": "Lecture 1: Introduction",
      "playlist_position": 1,
      "audio_path": "/path/to/audio/01.wav",
      "duration_seconds": 3600.0,
      "availability": "available"
    }
  ]
}
```

Fields:

| Field | Required | Description |
|-------|----------|-------------|
| `source_id` | yes | Unique video identifier (e.g. YouTube ID) |
| `title` | yes | Video title |
| `playlist_position` | no | Position in playlist |
| `audio_path` | yes (if available) | Path to normalized 16kHz mono WAV |
| `duration_seconds` | yes | Audio duration in seconds |
| `availability` | yes | `available`, `members_only`, `private`, `deleted`, `missing`, `unknown` |

A manifest converter script is provided at `scripts/convert_media_manifest.py` for converting a media index SQLite database to the manifest format.

### Commands

#### `corpus run` — transcribe pending videos

```bash
./transcribe corpus run \
  --manifest manifest.json \
  --database corpus.db \
  --output-dir exports/ \
  --format srt,vtt,txt \
  --source-id VIDEO_ID \
  --retry-failed \
  --dry-run \
  --verbose
```

Flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--manifest` | (required) | Path to manifest JSON |
| `--database` | (required) | Path to corpus SQLite database |
| `--output-dir` | (none) | Directory for per-video exports |
| `--format` | `srt,vtt,txt` | Export formats (comma-separated) |
| `--source-id` | (all) | Only process these source IDs (repeatable) |
| `--retry-failed` | false | Retry failed videos |
| `--dry-run` | false | Plan only; do not transcribe |
| `--fail-fast` | false | Stop on first failure |
| `--metal-gpu` | false | Use Metal GPU backend instead of Nemotron/Dagger |
| `--metal-backend` | (required with `--metal-gpu`) | `parakeet` or `whisper` |
| `--metal-binary` | (required with `--metal-gpu`) | Path to `parakeet-cli` or `whisper-cli` |
| `--metal-model` | (required with `--metal-gpu`) | Path to GGUF model file |
| `--chunk-size` | 60 | Seconds per ASR chunk (Nemotron only) |
| `--verbose` | false | Verbose output |

#### `corpus status` — print corpus state

```bash
./transcribe corpus status \
  --manifest manifest.json \
  --database corpus.db
```

Output:

```
corpus=my-lecture-series total=37 available=25 unavailable=12 pending=0 transcribing=0 complete=25 failed=0 stale=0
```

#### `corpus search` — full-text search

```bash
./transcribe corpus search \
  --manifest manifest.json \
  --database corpus.db \
  --query "functor" \
  --limit 10
```

Returns matching chunks with video title, timestamp range, YouTube deep link, and surrounding text.

#### `corpus export` — regenerate exports without ASR

```bash
./transcribe corpus export \
  --manifest manifest.json \
  --database corpus.db \
  --output-dir exports/ \
  --format srt,vtt,txt \
  --source-id VIDEO_ID
```

Regenerates SRT/VTT/TXT files from committed transcripts without re-running ASR.

### Resume behavior

If a corpus run is interrupted (process crash, machine sleep, network failure), restarting with the same database and manifest skips completed videos and resumes pending ones:

- **Complete** videos → skipped (or export-only if `--output-dir` is set)
- **Failed** videos → skipped unless `--retry-failed` is set
- **Transcribing** videos → treated as stale, reset to pending
- **Pending** videos → transcribed

Each transcription attempt records its pipeline fingerprint (SHA-256 of model name, decoding parameters, chunk size, audio contract). Changing the ASR model or configuration produces a new revision rather than overwriting existing work.

### Long audio handling (Metal backend)

Parakeet's Metal GPU encoder runs out of memory on audio longer than ~5000 seconds. The Metal transcriber automatically splits long audio into 3600-second (1-hour) chunks using ffmpeg, transcribes each chunk independently, and merges word timestamps with time offsets. This is transparent — no manual chunking is needed.

### Running in tmux

For long corpus runs, use tmux to survive terminal disconnects and macOS sleep:

```bash
tmux new-session -d -s corpus "./transcribe corpus run \
  --manifest manifest.json \
  --database corpus.db \
  --metal-gpu \
  --metal-backend parakeet \
  --metal-binary ~/code/whisper.cpp/build/bin/parakeet-cli \
  --metal-model ~/code/whisper.cpp/models/ggml-parakeet-tdt-0.6b-v3-f16.bin \
  --output-dir exports/ \
  --verbose 2>&1 | tee corpus.log"

# check progress
tmux capture-pane -t corpus -p | tail -10
```

### Direct SQLite search

If the audio files are not present on the machine running the search (e.g. the database was copied from another machine), the Go CLI search will fail on manifest validation. Direct SQLite queries work regardless:

```bash
sqlite3 corpus.db "
SELECT v.title, substr(c.text, 1, 80), ROUND(c.start_time, 1)
FROM chunk_fts fts
JOIN chunks c ON c.id = fts.rowid
JOIN transcript_revisions r ON c.revision_id = r.id
JOIN videos v ON v.id = r.video_id
WHERE chunk_fts MATCH 'functor'
LIMIT 10;"
```

## Batch transcription

For single-file transcription without corpus management:

```bash
./transcribe batch \
  -i input.wav \
  -o ./out \
  --format srt,vtt,txt,db \
  --no-fillers \
  --verbose
```

Flags:

| Flag | Default | Description |
|------|---------|-------------|
| `-i, --input` | (required) | Input audio file (WAV) |
| `-o, --output-dir` | `./out` | Output directory |
| `-f, --format` | `srt,db` | Output formats: srt, vtt, txt, db |
| `--no-fillers` | false | Remove filler words (um, uh, etc.) |
| `--chunk-size` | 60 | Seconds per transcription chunk |
| `--server-dir` | `server` | Path to Python server directory |
| `-v, --verbose` | false | Show Dagger logs |

## Live transcription

For streaming live audio transcription via WebSocket:

```bash
./transcribe live \
  -i input.wav \
  -o ./out-live \
  --live-format console,srt,vtt,txt,db \
  --transport ws \
  --replay-speed 1.0 \
  --chunk-duration 2 \
  --session-id my-session
```

Flags:

| Flag | Default | Description |
|------|---------|-------------|
| `-i, --input` | (required) | Input WAV file to replay as simulated live source |
| `-o, --output-dir` | `./out-live` | Output directory |
| `--live-format` | `console` | Output formats: console, srt, vtt, txt, db |
| `--transport` | `ws` | Transport: `ws` (WebSocket) or `chunk` (debug) |
| `--replay-speed` | 1.0 | Replay speed (1.0 = realtime, 0 = no pacing) |
| `--chunk-duration` | 2 | Replay chunk duration in seconds |
| `--overlap-seconds` | 0 | Expected chunk overlap |
| `--session-id` | (auto) | Stable session identifier |

## Architecture

```
transcription-go/
├── cmd/transcribe/
│   ├── main.go          # Entry point
│   ├── root.go          # Root + batch + live commands
│   ├── batch.go         # Batch transcription logic
│   ├── corpus.go        # Corpus subcommands (run, status, search, export)
│   └── live.go          # Live transcription logic
├── internal/
│   ├── asr/             # HTTP client for Nemotron ASR server
│   ├── convert/         # Pure-Go audio conversion
│   ├── corpus/          # Corpus pipeline (manifest, store, runner, export, search)
│   ├── live/            # Live transcription session management
│   ├── metal/           # Metal GPU transcriber (parakeet-cli / whisper-cli)
│   ├── output/          # SRT, VTT, TXT, SQLite formatting
│   └── server/          # Dagger service management
├── server/
│   ├── server.py        # FastAPI Nemotron ASR server
│   └── requirements.txt # Python dependencies
├── scripts/
│   └── convert_media_manifest.py  # Media index → manifest converter
└── ttmp/                # Docmgr ticket documentation
    └── 2026/07/28/VIDEO-CORPUS-PIPELINE--*/
        ├── analysis/    # Gap analysis
        ├── design-doc/  # Architecture and implementation guide
        ├── playbook/    # Operator playbook
        ├── reference/   # API contracts, diary
        ├── scripts/     # parakeet-cli JSON patch
        └── sources/     # ASR research sources
```

### Transcriber interface

The ASR backend is abstracted behind a Go interface:

```go
type Transcriber interface {
    Transcribe(ctx context.Context, audioPath string, opts TranscribeOptions) (Transcription, error)
}
```

Two implementations:

- **HTTPTranscriber** — sends audio chunks to a Dagger-hosted Nemotron server
- **MetalTranscriber** — shells out to `parakeet-cli` or `whisper-cli` with Metal GPU

The runner does not know which backend it is using. Adding a new ASR backend requires only implementing this interface and adding a CLI flag.

### SQLite schema

The corpus database has 12 tables:

| Table | Purpose |
|-------|---------|
| `corpora` | Corpus metadata |
| `videos` | Per-video state and metadata |
| `transcript_attempts` | Each transcription attempt with provenance |
| `transcript_revisions` | Committed transcripts (UNIQUE on video + fingerprint) |
| `words` | Word-level timestamps (normalized text, start/end, filler flag) |
| `chunks` | Derived chunks for FTS5 indexing |
| `chunk_words` | Junction table linking chunks to words |
| `chunk_fts` | FTS5 virtual table for full-text search |
| `exports` | Export records with content SHA-256 |

## Testing

```bash
make test
# or
go test ./... -count=1
```

## Benchmark

Corpus transcription of 69 hours of video lectures (25 videos):

| Backend | Hardware | Speed | Wall-clock | Words |
|---------|----------|-------|------------|-------|
| Parakeet TDT 0.6B v3 (Metal) | M1 Max | 98x realtime | 42 min | 540,840 |
| Nemotron 0.6B (Dagger/CPU) | RTX 3060 | 4.8x realtime | ~14 hr (extrapolated) | 69,266 (11 videos) |

## Project documentation

Detailed documentation is in the docmgr tickets:

- `ttmp/2026/04/13/TRANSCRIPTION-GO--go-dagger-transcription-pipeline/` — original single-file pipeline design
- `ttmp/2026/07/28/VIDEO-CORPUS-PIPELINE--adapt-transcription-go-for-playlist-video-corpus-pipelines/` — corpus pipeline design, implementation diary, API contracts, and operator playbook

## License

See the repository for license information.
