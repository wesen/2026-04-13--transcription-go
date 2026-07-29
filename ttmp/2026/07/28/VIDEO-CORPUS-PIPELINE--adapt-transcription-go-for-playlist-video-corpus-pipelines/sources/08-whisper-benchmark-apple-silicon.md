## How fast is Whisper on Apple Silicon Macs?

**OpenAI Whisper runs faster than realtime on every shipping Apple Silicon Mac, with the smaller 
models (tiny, base, small) clearing 10× realtime on an M2 Pro and beyond, and large-v3 hitting 
roughly 2 to 3× realtime on an M3 or M4 with Metal acceleration enabled.** That means a one-minute 
clip transcribes in 20 seconds with the highest-accuracy model, and a base-size model finishes the 
same minute in under 6 seconds on modern silicon.

The practical takeaway: Apple Silicon has changed what counts as a "live" dictation model. Three 
years ago, only `tiny` and `base` were realistic for press-and-hold dictation. Today, most users 
can comfortably run `small` for general dictation and `medium` for high-accuracy work without 
noticing latency.

## Methodology

**These numbers are illustrative ranges drawn from public whisper.cpp benchmarks plus internal 
JustVoice testing on a fleet of Apple Silicon Macs throughout early 2026.** Actual figures will 
vary with audio characteristics (clean speech vs noisy, monologue vs multi-speaker), thread count, 
thermal state, and model quantization. We have rounded conservatively rather than cherry-picking 
best-case runs.

The setup we used as a reference point:

- **Engine**: whisper.cpp ~1.6.x with the Metal backend enabled (`WHISPER_METAL=1`)
- **Audio**: a 60-second mono 16 kHz clip of clean English speech (a podcast monologue), looped to 
make a 10-minute reference for the longer models
- **Quantization**: q5\_0 / q5\_1 quantized weights for tiny/base/small (the defaults in most Mac 
apps); fp16 for medium and large-v3
- **Repetitions**: median of 5 runs after a warmup run; chips at idle thermals
- **Reported metric**: realtime factor (RTF). "9× realtime" means the model transcribes 60 seconds 
of audio in roughly 6.7 seconds.
We deliberately did not include batched inference, speculative decoding, or distil-whisper 
variants. Those tricks improve numbers significantly but most shipping Mac apps don't use them yet, 
and the goal here is "what should I expect when I open a dictation app today."

## Whisper model sizes and memory footprint

**The five canonical Whisper checkpoints span 75 MB to 2.9 GB on disk and roughly 1 GB to 10 GB of 
working RAM in fp16.** Quantized variants (q5, q8) cut both numbers by 30 to 60% with a small 
accuracy hit that is usually invisible for dictation.

| Model | Parameters | File size (fp16) | RAM headroom (fp16) | Quantized size |
| --- | --- | --- | --- | --- |
| `tiny` | 39M | 75 MB | ~1.0 GB | ~40 MB (q5) |
| `base` | 74M | 142 MB | ~1.5 GB | ~80 MB (q5) |
| `small` | 244M | 466 MB | ~2.5 GB | ~250 MB (q5) |
| `medium` | 769M | 1.5 GB | ~5.0 GB | ~800 MB (q5) |
| `large-v3` | 1550M | 2.9 GB | ~10 GB | ~1.6 GB (q5) |

Two underappreciated points. First, RAM headroom is not just the weights, Whisper allocates KV 
cache and Mel spectrogram buffers that scale with audio length. Pushing `large-v3` on an 8 GB Mac 
is technically possible but you will swap. Second, quantized models load and unload from disk 
roughly twice as fast, which matters for cold-start latency in press-and-hold dictation.

## Speed by model size on M1 / M2 Pro / M3 / M4

**The realtime factor scaling is roughly linear with GPU core count, not CPU count.** A base M3 
with 10 GPU cores beats a base M2 (8 GPU cores) at every model size, and an M4 Pro narrows the gap 
to large-v3 by roughly 35% over an M2 Pro.

Numbers below are RTF (×realtime). Higher is better. "9×" means a 60-second clip finishes in ~6.7 
seconds.

| Model | M1 (8-core GPU) | M2 Pro (16-core GPU) | M3 (10-core GPU) | M4 (10-core GPU) |
| --- | --- | --- | --- | --- |
| `tiny` (q5) | ~24× | ~38× | ~30× | ~38× |
| `base` (q5) | ~14× | ~22× | ~18× | ~24× |
| `small` (q5) | ~6× | ~11× | ~9× | ~12× |
| `medium` (fp16) | ~2.5× | ~5× | ~3.5× | ~5× |
| `large-v3` (fp16) | ~1.0× | ~2.5× | ~1.8× | ~2.6× |

Three observations the table makes obvious:
1. **M2 Pro still beats base M3 for Whisper.** The 16-core GPU in the M2 Pro is the floor for 
serious large-v3 work. If you want to live on `large-v3` and you're choosing between a used M2 Pro 
and a new base M3, the M2 Pro is the better Whisper machine.
2. **The base M3 → M4 jump is roughly 25 to 30% on the larger models.** Most of that comes from 
improved Neural Engine throughput and faster unified memory, not raw GPU core count.
3. **`large-v3` finally clears realtime on the base chip.** On M1, large-v3 was effectively a 
transcribe-then-wait model. On M3 it's barely usable for live; on M4 it's fine for most dictation 
lengths under a minute.

## Which Whisper model should I use on my Mac?

**For everyday dictation, `small` is the right answer on every Apple Silicon Mac shipping today; 
drop to `base` only on M1 / M1 Air with 8 GB RAM, and reach for `medium` or `large-v3` only when 
accuracy on technical vocabulary is the constraint.**

Concrete picks per chip:

- **M1 / M1 Air (8 GB)**: `base` (q5). `small` works but eats your RAM budget if you're also 
running Chrome and an IDE.
- **M1 Pro / M2 / M2 Air (16 GB)**: `small` (q5). Comfortable headroom, faster than realtime by a 
wide margin.
- **M2 Pro / M3 Pro / M4 Pro**: `medium` (fp16) for general use, `large-v3` (q5) for code and 
technical content.
- **M3 Max / M4 Max / Ultra**: `large-v3` always. The accuracy gain on proper nouns, code 
identifiers, and accented speech is meaningful and the speed cost is invisible.
A non-obvious tip: many users assume "bigger model = better dictation." For short, conversational 
dictation (the bread and butter of voice typing), `small` and `medium` produce nearly identical 
output, and `small` returns 2 to 3× faster. The tail-end accuracy of `large-v3` mostly shows up on 
rare proper nouns, low-quality audio, and non-English passages. If your hotkey-to-text latency 
matters more than catching every Czech surname, stay on `small`.

## GPU vs CPU: how much does Metal help?

**Enabling Metal (Apple's GPU compute API) on Apple Silicon yields a 30 to 60% speedup on Whisper 
inference over CPU-only execution, with the gap widening as model size increases.** On `tiny` the 
difference is barely felt; on `large-v3` it's the difference between usable and not.

A representative comparison on M2 Pro:

| Model | CPU only (8 perf cores) | Metal (16 GPU cores) | Speedup |
| --- | --- | --- | --- |
| `tiny` | ~28× | ~38× | 1.36× |
| `base` | ~15× | ~22× | 1.47× |
| `small` | ~6.5× | ~11× | 1.69× |
| `medium` | ~2.8× | ~5× | 1.79× |
| `large-v3` | ~1.3× | ~2.5× | 1.92× |

Two practical notes. First, Metal also reduces CPU pressure significantly, which matters if you're 
dictating into Cursor while a dev server and Docker are running on the same machine. Second, the 
Neural Engine (ANE) is *not* automatically used by whisper.cpp, there are ANE-accelerated forks 
(e.g. WhisperKit by Argmax) that can squeeze another 1.3 to 1.8× on top of Metal numbers, 
particularly on M3 and M4. If you see a Mac dictation app advertising "Neural Engine acceleration," 
that's what they mean.

## Accuracy: WER by model size

**Word Error Rate (WER) is the standard accuracy metric for speech recognition; lower is better. On 
clean English speech, Whisper's WER drops from roughly 9% on `tiny` to under 4% on `large-v3`, with 
most of the improvement coming between `base` and `small`.**

Approximate WER on the LibriSpeech test-clean benchmark, drawn from the original Whisper paper plus 
subsequent community testing of `large-v3`:

| Model | WER (clean English) | WER (noisy / accented) |
| --- | --- | --- |
| `tiny` | ~9.0% | ~22% |
| `base` | ~6.5% | ~17% |
| `small` | ~4.5% | ~12% |
| `medium` | ~3.8% | ~9% |
| `large-v3` | ~3.2% | ~7% |

The "noisy / accented" column is the more interesting one. The accuracy gap between `small` and 
`large-v3` on clean speech is about 1.3 percentage points, most users won't notice. On noisy or 
accented audio the gap widens to 5 points, which absolutely is noticeable. If you dictate from 
cafes, take meeting notes with multiple speakers, or speak English as a second language, the larger 
model genuinely earns its weight.

## What this means for live dictation

**For press-and-hold dictation, the core JustVoice workflow, total perceived latency is 
hotkey-press-to-text-appearing, which equals audio length plus inference time plus a fixed ~150 ms 
of pipeline overhead. With `small` on an M3, a 5-second utterance lands in roughly 5.7 seconds 
total; with `base`, in roughly 5.4 seconds.**

That's why model choice matters less than it looks for short dictation. The audio length itself 
dominates. Where model choice *does* matter is for longer dictation (30+ seconds), file 
transcription, and dictation into low-quality microphones, the cases where inference time is 
actually your bottleneck.

JustVoice runs Whisper locally on your Mac's GPU with Metal acceleration, ships quantized variants 
of every model size, and exposes the model picker so you can dial in the speed/accuracy trade-off 
for your hardware. [See how it works](https://justvoice.ai/features), or [download it 
here](https://justvoice.ai/download). If you're comparing it to cloud-only options like Wispr Flow, 
[we wrote a side-by-side](https://justvoice.ai/alternatives/wispr-flow).

For deeper reading on the Whisper architecture itself, the original [OpenAI Whisper 
paper](https://cdn.openai.com/papers/whisper.pdf) is the canonical reference, and 
[whisper.cpp](https://github.com/ggerganov/whisper.cpp) is the C++ port that powers most Mac apps 
shipping local Whisper today.
