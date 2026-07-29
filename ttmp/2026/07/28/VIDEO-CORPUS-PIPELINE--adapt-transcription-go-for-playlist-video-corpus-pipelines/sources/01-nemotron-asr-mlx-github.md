## nemotron-asr-mlx

[![nemotron-asr-mlx](https://github.com/199-biotechnologies/nemotron-asr-mlx/raw/main/banner.png)](h
ttps://github.com/199-biotechnologies/nemotron-asr-mlx/blob/main/banner.png)

**NVIDIA Nemotron ASR on Apple Silicon. 112x realtime. Pure MLX.**

[![PyPI 
version](https://camo.githubusercontent.com/5f6e324db45e0225fbdf7c9a8157609c4b378d13bfc7a8c4d20c83c2
886f8736/68747470733a2f2f696d672e736869656c64732e696f2f707970692f762f6e656d6f74726f6e2d6173722d6d6c7
82e737667)](https://pypi.org/project/nemotron-asr-mlx/) 
[![license](https://camo.githubusercontent.com/a8ac1dd1c547cbd515747d96fa4a33e71ded6e570ae17fcc3f103
46d06e8231f/68747470733a2f2f696d672e736869656c64732e696f2f707970692f6c2f6e656d6f74726f6e2d6173722d6d
6c782e737667)](https://github.com/199-biotechnologies/nemotron-asr-mlx/blob/main/LICENSE) [![python 
version](https://camo.githubusercontent.com/859e65b1e2a0f6ad95a4f887333d3a51733fa716d987ec4cdc5c197f
1a3604cf/68747470733a2f2f696d672e736869656c64732e696f2f707970692f707976657273696f6e732f6e656d6f74726
f6e2d6173722d6d6c782e737667)](https://www.python.org/)

---

93 minutes of audio transcribed in under a minute on an M-series Mac. No GPU drivers, no CUDA, no 
Docker. Just `pip install` and go.

This is a native [MLX](https://github.com/ml-explore/mlx) port of [NVIDIA's Nemotron-ASR 
0.6B](https://huggingface.co/nvidia/nemotron-asr-speech-streaming-en-0.6b) — the cache-aware 
streaming conformer that processes each audio frame exactly once. No sliding windows, no 
recomputation, no rewinding. State lives in fixed-size ring buffers so latency stays flat no matter 
how long you talk.

## Requirements

- Apple Silicon Mac (M1/M2/M3/M4)
- Python 3.10+
- [ffmpeg](https://ffmpeg.org/) installed and on PATH (for audio loading)

## Install

```
pip install nemotron-asr-mlx
```

Model weights (~1.2 GB) download automatically on first run from 
[HuggingFace](https://huggingface.co/dboris/nemotron-asr-mlx).

## Quick Start

```
from nemotron_asr_mlx import from_pretrained

model = from_pretrained("dboris/nemotron-asr-mlx")
result = model.transcribe("meeting.wav")
print(result.text)    # full transcription string
print(result.tokens)  # list of BPE token IDs

# Optional: beam search for maximum accuracy (slower)
result = model.transcribe("meeting.wav", beam_size=4)

# Maximum accuracy: beam search + ILM subtraction
result = model.transcribe("meeting.wav", beam_size=4, ilm_scale=0.15)
```

`transcribe()` accepts a file path (any format ffmpeg supports: wav, mp3, flac, m4a, ogg, opus, 
webm, mp4, etc.) or a numpy array of float32 PCM samples at 16 kHz.

It returns a `StreamEvent` with these fields:

| Field | Type | Description |
| --- | --- | --- |
| `text` | `str` | Full transcription text |
| `text_delta` | `str` | New text (same as `text` in batch mode) |
| `tokens` | `list[int]` | BPE token IDs |
| `is_final` | `bool` | Always `True` in batch mode |

## CLI

```
nemotron-asr transcribe meeting.wav                 # transcribe a file
nemotron-asr transcribe recording.mp3               # any format ffmpeg supports
nemotron-asr transcribe meeting.wav --beam-size 4   # beam search (slower, lower WER)
nemotron-asr transcribe meeting.wav --beam-size 4 --ilm-scale 0.15  # + ILM subtraction
nemotron-asr listen                                 # stream from microphone
```

## Benchmark

### Official WER (Open ASR Leaderboard datasets)

Evaluated on the standard [Open ASR 
Leaderboard](https://huggingface.co/spaces/hf-audio/open_asr_leaderboard) datasets. Machine: Apple 
M4 Max, 16-core, 64 GB.

| Dataset | WER | NVIDIA ref | RTFx |
| --- | --- | --- | --- |
| LibriSpeech test-clean | **2.70%** | 2.31% | 112x |
| LibriSpeech test-other | **5.52%** | 4.75% | 98x |
| TED-LIUM v3 | **6.27%** | 4.50% | 122x |

NVIDIA reference numbers are from 
[nemotron-asr-speech-streaming-en-0.6b](https://huggingface.co/nvidia/nemotron-asr-speech-streaming-
en-0.6b) at 1120ms chunk size (PyTorch, A100 GPU). Our MLX port runs in batch mode on Apple Silicon.

v0.2.0 improvements: mel frontend parity fixes (periodic Hann window, center-padded STFT) + 
blank-frame skipping decoder reduced WER from 2.79% to 2.70% and increased speed from 76x to 112x 
realtime.

Run the evaluation yourself:

```
pip install datasets jiwer torchcodec
python eval_wer.py librispeech-clean librispeech-other tedlium
```

### Speed benchmark

Measured on LibriSpeech test-clean, Apple M4 Max, 64 GB.

| Content | Duration | Inference | Speed |
| --- | --- | --- | --- |
| Short utterances (~3s each, 50 samples) | 4.1 min | 2.7s | **90x** RT |
| Sentences (~10s each, 50 samples) | 9.7 min | 4.6s | **128x** RT |
| Paragraphs (~30s each, 20 samples) | 10.0 min | 5.3s | **113x** RT |
| Long-form (14 min single file) | 14.3 min | 8.2s | **105x** RT |
| Full test-clean (2,620 samples) | 5.4 hours | 173s | **112x** RT |

618.5M parameters. 3.4 GB peak GPU memory. Model loads in 0.1s after first download.

## Why this exists

Most "streaming" ASR on Mac is either (a) Whisper with overlapping windows reprocessing the same 
audio over and over, or (b) cloud APIs adding network latency to every utterance. Nemotron's 
cache-aware conformer is architecturally different:

- **Each frame processed once** — state carried forward in fixed-size ring buffers, not recomputed
- **Constant memory** — no growing KV caches, no memory spikes on long recordings
- **Native Metal** — no PyTorch, no ONNX, no bridge layers. Direct MLX on Apple GPU
- **112x realtime** — an hour of audio in 32 seconds

## Architecture

FastConformer encoder (24 layers, 1024-dim) with 8x depthwise striding subsampling. RNNT decoder 
with 2-layer LSTM prediction network and joint network. Per-layer-group attention context windows 
`[[70,13], [70,6], [70,1], [70,0]]` for progressive causal restriction. Greedy decoding with 
blank-frame skipping (batched joint network evaluation skips ~90% of silent frames). Optional beam 
search with n-gram LM shallow fusion and ILM (Internal Language Model) subtraction.

Based on [Cache-aware Streaming Conformer](https://arxiv.org/abs/2312.17279) and the 
[NeMo](https://github.com/NVIDIA/NeMo) toolkit.

## Live Demo

A browser-based demo with live mic transcription. Mic is captured in the terminal via sounddevice; 
the browser displays the transcript.

```
pip install websockets sounddevice
python demo/server.py
```

Open [http://localhost:8765](http://localhost:8765/) and click Record.

## Weight Conversion

If you have a `.nemo` checkpoint and want to convert it yourself:

```
pip install torch safetensors pyyaml  # conversion deps only
nemotron-asr convert model.nemo ./output_dir
```

Produces `config.json` + `model.safetensors`. Conversion deps are not needed for inference.

## Dependencies

Deliberately minimal:

- `mlx` — Apple's ML framework
- `huggingface-hub` — model download
- `numpy` — mel spectrogram
- `librosa` — mel filterbank (optional, improves accuracy)
- `sounddevice` — mic access (for live streaming)
- `websockets` — live demo server (optional)
- `typer` — CLI
