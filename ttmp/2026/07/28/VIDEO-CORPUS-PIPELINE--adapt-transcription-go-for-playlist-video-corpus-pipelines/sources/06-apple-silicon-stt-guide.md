## The Problem

I recently recorded a brainstorming session (15MB m4a, ~20 minutes) and wanted to transcribe it 
locally. OpenAI’s Whisper seemed like the obvious choice — open source, free, runs locally, no 
privacy concerns.

My machine: MacBook Pro M4 Pro with 24GB unified memory.

## First Attempt: CPU Whisper

I used the Homebrew-installed `whisper` CLI with the `turbo` model:

```bash
whisper recording.m4a \
  --model turbo \
  --language en \
  --initial_prompt "Context about the recording topic." \
  --output_dir ./docs \
  --output_format txt
```

Then I waited… for 2.5 hours. CPU temperature spiked, fans went full blast, `top` showed over 
1000% CPU usage.

**The output directory was empty. Not a single word.**

The issue is clear: OpenAI’s Python Whisper defaults to CPU on macOS and completely ignores the 
M4 Pro’s GPU.

## The Solution: mlx-whisper

Apple Silicon Macs have a dedicated ML framework called [MLX](https://github.com/ml-explore/mlx), 
built by Apple specifically for their chips (think PyTorch but for Apple Silicon). The community 
built [mlx-whisper](https://github.com/ml-explore/mlx-examples/tree/main/whisper) on top of it, 
leveraging Metal GPU acceleration.

### Installation

```bash
# macOS Homebrew Python doesn't allow direct pip install, use pipx
pipx install mlx-whisper
```

### Usage

```bash
mlx_whisper \
  --model mlx-community/whisper-turbo \
  --language en \
  --initial-prompt "Context about the recording topic." \
  --output-dir ./docs \
  --output-format txt \
  recording.m4a
```

Result: **Transcription completed in a few minutes** with good quality, handling mixed 
English-Mandarin segments well.

### Performance Comparison

| Method | Time | CPU Usage | Result |
| --- | --- | --- | --- |
| whisper (CPU) | 2.5+ hours | \>1000% | No output |
| mlx-whisper (GPU) | ~3 minutes | Normal | Full transcript |

That’s roughly a **50x difference**. If you’re on Apple Silicon, mlx-whisper is non-negotiable.

## About the Turbo Model

Whisper turbo’s full name is `large-v3-turbo` — `turbo` is just the shorthand. It was created 
by pruning large-v3’s 32 decoder layers down to 4, then fine-tuning:

- Parameters: 809M (between medium at 769M and large at 1550M)
- Speed: ~8x faster than large
- Accuracy: Only ~1-2% WER increase
- Multilingual only, no English-only variant

For most use cases, turbo offers the **best bang for your buck**.

## Open-Source STT Landscape in 2026

I did some research and found that Whisper hasn’t been updated in a while:

- **2022-09**: Whisper initial release
- **2023-11**: large-v3 released
- **2024-10**: large-v3-turbo released (last update)

OpenAI has shifted focus to closed-source API models (`gpt-4o-transcribe`). The open-source Whisper 
is essentially in maintenance mode.

Meanwhile, other open-source models have surpassed Whisper:

| Model | WER | Highlights | M4 Pro Compatible? |
| --- | --- | --- | --- |
| **NVIDIA Canary Qwen 2.5B** | 5.63 | Current open-source best | ⚠️ Requires NeMo, poor Mac 
support |
| **IBM Granite Speech 8B** | 5.85 | Enterprise-grade | ⚠️ 8B model too large, 24GB barely 
enough |
| Whisper Large V3 | 7.4 | Most mature ecosystem | ✅ via mlx-whisper |
| Whisper Turbo | 7.75 | Fast | ✅ via mlx-whisper |
| **NVIDIA Parakeet TDT 1.1B** | ~8.0 | Ultra-low latency | ✅ MLX version available |

For Apple Silicon users, **mlx-whisper (turbo)** remains the most practical choice — mature 
ecosystem, easy setup. If you need lower latency, **Parakeet TDT** also has an MLX version worth 
trying.

## Cache Cleanup Reminder

Whisper model caches eat up significant disk space. I found 6 different model versions cached on my 
machine, totaling **11.3GB**. Remember to clean up after use:

```bash
# Check cache sizes
du -sh ~/.cache/huggingface/hub/models--*whisper*
du -sh ~/.cache/whisper/

# Remove unused models
rm -rf ~/.cache/huggingface/hub/models--unused-model
rm -rf ~/.cache/whisper/
```

## Takeaways

1. On Apple Silicon Macs, **always use mlx-whisper** — vanilla Whisper on CPU is unusably slow
2. `turbo` (aka large-v3-turbo) is the most cost-effective model available
3. While Whisper’s open-source development has stalled, the MLX ecosystem keeps it the most 
convenient option on Mac
4. Clean up model caches after use — easily saves 10GB+ of disk space
