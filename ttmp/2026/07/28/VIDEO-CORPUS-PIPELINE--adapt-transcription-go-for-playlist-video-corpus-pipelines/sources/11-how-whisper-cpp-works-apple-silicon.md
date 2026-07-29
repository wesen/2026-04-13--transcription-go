OpenAI released Whisper as a Python library requiring a GPU server to run at reasonable speeds. 
Twelve months later, people were running it in real-time on a MacBook Air. The bridge between those 
two realities is **whisper.cpp** - a C/C++ reimplementation by Georgi Gerganov that makes local 
speech recognition practical on consumer hardware.

This is the engine inside LexaWrite and most other offline Mac dictation apps. Here’s how it 
actually works.

## The Problem Whisper.cpp Solves

OpenAI’s original Whisper is a PyTorch model. Running it requires:

- Python runtime
- PyTorch (hundreds of megabytes)
- CUDA-capable GPU (NVIDIA only) for reasonable speed
- Significant RAM overhead from Python’s memory management

On a Mac with Apple Silicon, none of this applies. There’s no NVIDIA GPU, no CUDA. Python adds 
overhead. The original Whisper code is a non-starter for a lightweight desktop app.

Whisper.cpp solves this by reimplementing Whisper’s entire inference pipeline in pure C/C++ with:

- Zero Python dependency
- Metal GPU acceleration for Apple Silicon
- Optimized memory layout for ARM64
- Model quantization to reduce size and speed up inference
- Single-file deployment (one executable, one model file)

## Whisper’s Architecture (Simplified)

Whisper is an **encoder-decoder transformer** - the same fundamental architecture as GPT, but 
designed for audio instead of text.

### Step 1: Audio Preprocessing

Raw audio from your microphone arrives as a waveform - amplitude values over time. Before the model 
can process it:

1. **Resample to 16kHz mono** - Whisper expects exactly 16,000 samples per second, single channel. 
Mac microphones typically capture at 44.1kHz or 48kHz stereo, so conversion is needed.
2. **Compute log-mel spectrogram** - the raw waveform is converted into a visual representation of 
frequency over time (80 mel-frequency bins). This is what the model actually “sees.”
3. **Pad or chunk to 30 seconds** - Whisper processes audio in 30-second windows. Shorter audio is 
padded with silence. Longer audio is split into chunks.

The spectrogram conversion happens in whisper.cpp’s `whisper_pcm_to_mel()` function - pure C, no 
library dependencies, highly optimized.

### Step 2: Encoder

The encoder is a stack of transformer blocks that processes the mel spectrogram and produces a 
sequence of hidden states - a compressed representation of “what was said.”

For the Small model:

- 12 encoder layers
- 768-dimensional hidden states
- 12 attention heads
- ~120M parameters (half the total model)

The encoder runs once per 30-second chunk. On Apple Silicon with Metal, this is the most 
compute-intensive step - and where GPU acceleration matters most.

### Step 3: Decoder

The decoder generates text tokens one at a time, attending to both the encoder output (the audio 
representation) and previously generated tokens (the text so far).

For the Small model:

- 12 decoder layers
- 768-dimensional hidden states
- 12 attention heads
- ~120M parameters

The decoder runs autoregressively - each token requires a forward pass. A 30-second audio clip 
might produce 50-100 tokens, so the decoder runs 50-100 times. This is why decoder speed matters so 
much.

### Step 4: Token-to-Text

The decoder outputs token IDs from Whisper’s vocabulary (~51,865 tokens including multilingual 
tokens, timestamps, and special tokens). These are mapped back to text strings and concatenated to 
produce the final transcription.

## Why Metal Makes It Fast

Apple Silicon chips (M1, M2, M3, M4) have a unified memory architecture - the CPU and GPU share the 
same physical RAM. This eliminates the biggest bottleneck in traditional GPU computing: copying 
data between CPU memory and GPU memory.

In a traditional setup (NVIDIA GPU):

```plaintext
CPU RAM → [copy over PCIe bus] → GPU VRAM → compute → [copy back] → CPU RAM
```

On Apple Silicon:

```plaintext
Unified Memory → GPU compute → result is already accessible by CPU
```

No copies. No bus latency. The model weights sit in unified memory and are directly accessible by 
the Metal GPU shaders.

Whisper.cpp’s Metal backend (`ggml-metal.m`) compiles compute shaders that run matrix 
multiplications on Apple’s GPU cores. The key operations:

- **Matrix multiplication** (the bulk of transformer computation) - runs on GPU
- **Softmax** - runs on GPU
- **Layer normalization** - runs on GPU
- **Mel spectrogram computation** - runs on CPU (I/O bound, not worth GPU overhead)

On an M1 MacBook Air, the Small model encoder processes a 30-second chunk in about 1.5 seconds on 
Metal vs. ~6 seconds on CPU-only. A 4x speedup from the GPU alone.

## Quantization: Shrinking Models Without Losing Quality

Whisper’s original models use 32-bit floating point (FP32) weights. Whisper.cpp supports 
quantized formats that reduce precision to save memory and increase speed:

| Format | Bits per Weight | Size (Small) | Speed Impact | Accuracy Impact |
| --- | --- | --- | --- | --- |
| FP32 | 32 | 466MB | Baseline | Baseline |
| FP16 | 16 | 233MB | ~1.3x faster | Negligible |
| Q8\_0 | 8 | ~120MB | ~1.8x faster | Very small |
| Q5\_1 | 5 | ~85MB | ~2.2x faster | Small |
| Q4\_0 | 4 | ~70MB | ~2.5x faster | Noticeable |

Most apps (including LexaWrite) use FP16 or Q8\_0 - the sweet spot where file size is halved and 
speed improves with virtually no accuracy loss.

The quantization is done offline - you download an already-quantized model file. At inference time, 
whisper.cpp’s GGML library handles the dequantization arithmetic in its matrix multiplication 
kernels.

## The Inference Pipeline (What Happens When You Speak)

Here’s the exact sequence when you hold Fn and speak in LexaWrite:

1. **Audio capture** - AVAudioEngine records at the mic’s native sample rate (typically 48kHz)
2. **Format conversion** - Audio is resampled to 16kHz mono Float32 PCM in real-time via an 
AVAudioConverter
3. **Buffer accumulation** - PCM samples accumulate in a ring buffer while you speak
4. **Release Fn** - recording stops, the audio buffer is finalized
5. **Minimum length check** - if less than 0.5 seconds, abort (Whisper needs enough audio context)
6. **Mel spectrogram** - `whisper_pcm_to_mel()` converts the PCM buffer to log-mel features
7. **Encoder forward pass** - Metal GPU processes the spectrogram through 12 transformer layers
8. **Decoder loop** - generates text tokens one at a time until an end-of-text token
9. **Token assembly** - tokens are mapped to text strings
10. **Post-processing** - custom dictionary replacements, style matching
11. **Paste** - text is placed on clipboard and Cmd+V is simulated into the foreground app

Steps 6-9 (the actual Whisper inference) take about 1-3 seconds for the Small model on Apple 
Silicon for a typical dictation of 5-15 seconds. The total end-to-end latency from releasing Fn to 
seeing text is typically under 2 seconds.

## Performance by Hardware

Real-world transcription speed for 30 seconds of audio:

| Chip | Small Model | Medium Model | Large Model |
| --- | --- | --- | --- |
| M1 (8-core GPU) | ~3s | ~12s | ~28s |
| M1 Pro (16-core GPU) | ~2s | ~7s | ~18s |
| M2 (10-core GPU) | ~2.5s | ~9s | ~22s |
| M3 (10-core GPU) | ~2s | ~7s | ~16s |
| M3 Pro (18-core GPU) | ~1.5s | ~5s | ~12s |
| M4 (10-core GPU) | ~1.8s | ~6s | ~14s |
| M4 Pro (20-core GPU) | ~1.2s | ~4s | ~9s |

*Approximate values. Actual performance varies with audio content, background load, and thermal 
conditions.*

The takeaway: for the Small model, every Apple Silicon Mac transcribes faster than real-time. The 
model is essentially “instant” for typical dictation lengths (5-15 seconds).

## Why Not CoreML?

Apple provides CoreML as the standard framework for running ML models on Apple Silicon. Several 
apps use CoreML-converted Whisper models instead of whisper.cpp. Why does LexaWrite use whisper.cpp?

**Advantages of whisper.cpp:**

- More model size options (CoreML conversions are typically limited to a few sizes)
- Faster iteration (whisper.cpp updates within days of upstream Whisper changes)
- Better quantization support
- Community-driven optimizations
- Cross-platform (if we ever target other platforms)

**Advantages of CoreML:**

- Tighter integration with Apple’s Neural Engine (ANE)
- Potentially better power efficiency for sustained workloads
- Simpler deployment for App Store apps

In practice, whisper.cpp with Metal is fast enough that the CoreML advantages don’t justify the 
tradeoffs. And whisper.cpp’s model flexibility (5 sizes, multiple quantization levels) is a 
significant user-facing feature.

## The Open Source Advantage

Whisper.cpp is MIT-licensed open source. This means:

1. **Anyone can audit the code** - verify that audio processing happens locally
2. **Bugs get fixed quickly** - a large community of contributors
3. **Performance improves continuously** - new Metal optimizations land regularly
4. **No vendor lock-in** - if Gerganov stopped maintaining it, someone else could fork it

As of 2026, whisper.cpp has over 36,000 GitHub stars and hundreds of contributors. It’s one of 
the most actively maintained ML inference libraries in the open source ecosystem.

## What’s Next for Local Speech Recognition

The trajectory is clear: local inference is getting faster, more efficient, and more accurate.

- **Whisper v3 and beyond** - OpenAI continues improving Whisper’s accuracy
- **Apple Silicon improvements** - each M-series generation adds GPU cores and bandwidth
- **Better quantization** - new formats (like GGML’s k-quants) reduce model size further with 
minimal accuracy loss
- **Distillation** - smaller models trained to match larger ones’ accuracy (Distil-Whisper is an 
active research direction)

The gap between local and cloud speech recognition narrows with every hardware and software 
generation. For the Small model on Apple Silicon, we’re already at the point where the bottleneck 
is human speaking speed, not machine processing speed.

Your MacBook is a speech recognition powerhouse. Whisper.cpp is the software that unlocks it.

[Try LexaWrite](#download) - whisper.cpp running on your Mac’s GPU, nothing more.
