---
Title: "Apple Silicon Optimization Research Summary"
Ticket: VIDEO-CORPUS-PIPELINE
Status: active
Topics:
    - asr
    - transcription
    - apple-silicon
    - metal
    - optimization
DocType: reference
Intent: long-term
Summary: Research into Apple Metal GPU acceleration options for Nemotron ASR and alternative speech recognition models on Apple Silicon.
LastUpdated: 2026-07-28T00:00:00Z
---

# Apple Silicon Optimization Research Summary

## Current state

Our Nemotron 0.6B ASR model runs on CPU inside a Dagger/Docker container using `python:3.11-slim-bookworm` on x86 emulation. The server explicitly sets `asr_model.cpu().eval()`, so even if PyTorch MPS were available in the container, the model wouldn't use it.

On the Mac M1 Max (mimimi-2.local), video 032 (985.9s audio) took approximately 6 minutes of ASR time (~2.7x realtime). On the Linux host (Ryzen), video 001 (3475s audio) took ~12 minutes (~4.8x realtime). The Mac is slower per-chunk (~6s/chunk vs ~10s/chunk on Linux), but the Mac runs inside Docker x86 emulation.

## Three optimization paths identified

### Path 1: Nemotron with PyTorch MPS (native Python, not Docker)

Source: `01-nemotron-asr-mlx-github.md`, `02-apple-metal-pytorch.md`

There is a community project (`199-biotechnologies/nemotron-asr-mlx`) that ports Nemotron Speech Streaming ASR to Apple Silicon via MLX. This runs natively on macOS using the Neural Engine, with full WER parity vs fp32.

A simpler approach: run the existing NeMo server natively on macOS (not in Docker) and change `asr_model.cpu().eval()` to `asr_model.to("mps").eval()`. PyTorch 2.x supports MPS (Metal Performance Shaders) on Apple Silicon. This would require:
- Python 3.11+ natively on macOS (not in Docker)
- PyTorch with MPS support (`pip install torch` on macOS gives MPS-enabled build)
- NeMo toolkit installed natively
- Change one line in `server.py`: `asr_model.to("mps").eval()` instead of `asr_model.cpu().eval()`

Risk: NeMo may not fully support MPS for all operations (FastConformer-RNNT has custom CUDA kernels). Some ops may fall back to CPU, but the overall pipeline could still be faster.

### Path 2: whisper.cpp with Metal (most mature Apple Silicon ASR)

Sources: `03-whisper-cpp-github.md`, `04-whisper-cpp-benchmark-mac.md`, `07-faster-whisper-vs-whisper-cpp-2026.md`, `08-whisper-benchmark-apple-silicon.md`, `11-how-whisper-cpp-works-apple-silicon.md`

whisper.cpp runs OpenAI Whisper models natively on Apple Silicon with full Metal GPU acceleration:
- `large-v3` model achieves 2-3x realtime on M3/M4 with Metal
- On M1 Max, expect ~1.5-2x realtime for large-v3
- Builds with `-DGGML_METAL=ON` flag
- No Docker/Python overhead — pure C/C++ with Metal shaders
- Word-level timestamps supported via `--output-words`

This would require adding a new `Transcriber` implementation that calls whisper.cpp as a subprocess or via its C API, replacing the Nemotron/Dagger backend. The Go CLI already has a `Transcriber` interface, so this is architecturally clean.

Benchmarks from sources suggest whisper.cpp large-v3 on M1 Max Metal would process 78 hours of audio in roughly 39-52 hours, vs our current ~15+ hours per machine with Nemotron CPU. However, running whisper.cpp natively avoids Docker overhead entirely.

### Path 3: WhisperKit (Apple CoreML)

Source: `10-whisperkit-vs-whisper-cpp.md`

Argmax WhisperKit uses Apple's CoreML framework for even deeper Apple Silicon optimization. It may edge ahead of whisper.cpp on Apple devices due to deeper CoreML integration. However, it's a Swift package, which adds a language boundary to our Go CLI.

## Recommendation

**Short-term**: Run Nemotron natively on macOS with PyTorch MPS (`asr_model.to("mps")`). This is a one-line change to `server.py` and avoids Docker x86 emulation overhead entirely. If MPS has partial op support, the fallback is automatic.

**Medium-term**: Add a whisper.cpp `Transcriber` implementation. This gives the best Apple Silicon performance, native Metal GPU usage, and no Python/Docker dependency. The `Transcriber` interface in `internal/corpus/transcriber.go` makes this a clean adapter.

**Parallelization**: Once we have Metal-accelerated ASR on the Mac, we can:
1. Run the Mac and Linux in parallel (already doing this)
2. On the Mac, run multiple whisper.cpp processes in parallel (Metal supports concurrent execution)
3. Use the Mac's Neural Engine via MLX for a third parallel path if the nemotron-asr-mlx project is viable

## Key files in this sources/ folder

| File | Content |
|------|---------|
| `01-nemotron-asr-mlx-github.md` | Community MLX port of Nemotron to Apple Silicon |
| `02-apple-metal-pytorch.md` | Apple's official PyTorch Metal guide |
| `03-whisper-cpp-github.md` | whisper.cpp repo with Metal support |
| `04-whisper-cpp-benchmark-mac.md` | whisper.cpp speed/accuracy benchmarks on Apple Silicon |
| `06-apple-silicon-stt-guide.md` | Real-world STT guide for Apple Silicon |
| `07-faster-whisper-vs-whisper-cpp-2026.md` | 2026 comparison of ASR tools |
| `08-whisper-benchmark-apple-silicon.md` | M1-M4 whisper benchmark results |
| `09-reddit-nemotron-apple-silicon.md` | Reddit discussion of Nemotron on Apple Silicon |
| `10-whisperkit-vs-whisper-cpp.md` | WhisperKit vs whisper.cpp comparison |
| `11-how-whisper-cpp-works-apple-silicon.md` | Technical deep dive into whisper.cpp on Apple Silicon |
| `12-nvidia-nemotron-hub-github.md` | NVIDIA Nemotron developer hub |
