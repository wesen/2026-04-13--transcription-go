SpeakUp is powered by [whisper.cpp](https://github.com/ggerganov/whisper.cpp) — a C/C++ port of 
OpenAI's Whisper speech recognition model, optimized for Apple Silicon through Metal GPU 
acceleration. This post breaks down how it performs across different Macs, model sizes, and 
real-world dictation scenarios.

## What Is whisper.cpp?

In September 2022, OpenAI released Whisper — an open-source speech recognition model trained on 
680,000 hours of multilingual audio. It was a breakthrough in accuracy, but the original Python 
implementation required a dedicated GPU server to run at practical speeds.

Georgi Gerganov's whisper.cpp changed that. Written in C/C++ with no dependencies, it brought 
Whisper to consumer hardware. The critical addition for Mac users was Metal GPU support, which 
offloads the heavy matrix multiplications to Apple Silicon's GPU cores. This made real-time 
on-device transcription not just possible, but fast.

SpeakUp integrates whisper.cpp directly into a native Swift application, managing the audio 
pipeline, model loading, and text output. When you press the hotkey and speak, whisper.cpp is doing 
the work on your Mac's GPU.

## Performance Across Apple Silicon Generations

We measure performance using the "real-time factor" (RTF) — the ratio of processing time to audio 
duration. An RTF of 0.1 means a 30-second clip processes in 3 seconds. Lower is better.

The following benchmarks use the "small" model (244M parameters) with Metal GPU enabled, processing 
a 30-second English audio clip:

- **M1 (8-core GPU):** RTF ~0.10 — 30s audio in ~3.0s
- **M1 Pro (16-core GPU):** RTF ~0.07 — 30s audio in ~2.1s
- **M2 (10-core GPU):** RTF ~0.08 — 30s audio in ~2.4s
- **M2 Pro (19-core GPU):** RTF ~0.05 — 30s audio in ~1.5s
- **M3 (10-core GPU):** RTF ~0.06 — 30s audio in ~1.8s
- **M3 Pro (18-core GPU):** RTF ~0.04 — 30s audio in ~1.2s
- **M4 (10-core GPU):** RTF ~0.05 — 30s audio in ~1.5s
- **M4 Pro (20-core GPU):** RTF ~0.03 — 30s audio in ~0.9s

Every Apple Silicon Mac from the base M1 onward can run the small model faster than real time. The 
Pro and Max variants with additional GPU cores see proportional speedups. For the base M1 — the 
oldest and least powerful Apple Silicon chip — a 30-second dictation still processes in about 3 
seconds. That is fast enough for seamless real-time use.

## Model Sizes: Speed vs Accuracy Tradeoffs

Whisper comes in several model sizes, each balancing speed against transcription quality. Here is 
how they compare on an M2 Mac with Metal GPU:

- **Tiny** (39M params, ~75 MB) — RTF: 0.02. Extremely fast but noticeably less accurate. Good 
for quick notes where speed matters more than precision. Word Error Rate (WER) on English: ~8%.
- **Base** (74M params, ~142 MB) — RTF: 0.04. A step up in accuracy with minimal speed cost. WER: 
~5.5%.
- **Small** (244M params, ~466 MB) — RTF: 0.08. The sweet spot for most users. Excellent accuracy 
with fast processing. WER: ~3.4%. This is what SpeakUp uses by default.
- **Medium** (769M params, ~1.5 GB) — RTF: 0.20. Higher accuracy for difficult audio (heavy 
accents, noisy environments). WER: ~2.9%. Available in SpeakUp for users who prioritize accuracy.
- **Large-v3** (1.55B params, ~3 GB) — RTF: 0.45. The most accurate Whisper model. WER: ~2.5%. 
Usable on Macs with 16 GB+ RAM, but the processing time makes it better suited for transcribing 
recordings than real-time dictation.

SpeakUp defaults to the small model because it delivers professional-grade accuracy while keeping 
processing well under real-time on every Apple Silicon Mac. Users can switch to the medium model in 
settings for scenarios where maximum accuracy is needed.

## Accuracy: How whisper.cpp Compares to Cloud Services

Word Error Rate (WER) is the standard metric for speech recognition accuracy. It measures the 
percentage of words that are inserted, deleted, or substituted incorrectly. Lower is better.

On standard English benchmarks (LibriSpeech test-clean), here is how whisper.cpp with the small 
model compares:

- **whisper.cpp (small):** ~3.4% WER
- **Google Speech-to-Text:** ~3.0% WER
- **Microsoft Azure Speech:** ~3.2% WER
- **Apple Dictation (on-device):** ~5.5% WER
- **whisper.cpp (medium):** ~2.9% WER

The small model is within 0.4 percentage points of Google's cloud service, while running entirely 
on your local hardware with zero latency and complete privacy. The medium model actually 
outperforms most cloud offerings. [Apple's built-in 
dictation](https://getspeakup.app/vs/apple-dictation/) uses a much smaller on-device model, which 
explains its significantly higher error rate.

For German, whisper.cpp's advantage is even more pronounced. Whisper was trained on substantial 
German audio data, and the small model achieves ~4.2% WER on German benchmarks — competitive with 
or better than most cloud services for German speech.

## Memory Usage and Low Memory Mode

On-device inference requires loading the model into memory. Here is the approximate RAM usage for 
each model on Apple Silicon (using Metal GPU, where the model is loaded into unified memory shared 
between CPU and GPU):

- **Tiny:** ~200 MB
- **Base:** ~350 MB
- **Small:** ~900 MB
- **Medium:** ~2.5 GB
- **Large-v3:** ~5 GB

For Macs with 8 GB of unified memory, the small model runs comfortably with room for other 
applications. SpeakUp includes a Low Memory mode that keeps the model unloaded when not actively 
dictating, freeing memory between sessions. When you press the hotkey, the model loads in under a 
second on most Macs, so the delay is imperceptible.

On 16 GB Macs, you can run the medium model alongside a full workload without memory pressure. The 
8 GB base MacBook Air — the most constrained Apple Silicon Mac — handles the small model 
without issue, which is why SpeakUp chose it as the default.

## Why Metal GPU Matters

Speech recognition models perform billions of floating-point operations per inference pass. CPUs 
can handle this, but slowly. Apple Silicon's GPU cores, accessed through the Metal framework, are 
designed for exactly this kind of parallel computation.

The difference is dramatic. On an M2, the small model with CPU-only inference has an RTF of 
approximately 0.35 — meaning a 30-second clip takes about 10.5 seconds to process. With Metal GPU 
enabled, the same clip processes in 2.4 seconds. That is a **4.4x speedup**, and it is the 
difference between dictation that feels laggy and dictation that feels instantaneous.

Metal acceleration also offloads the work from CPU cores, meaning your Mac stays responsive while 
dictating. You can dictate and your other apps continue running smoothly because the heavy 
computation is happening on the GPU.

whisper.cpp's Metal backend takes advantage of Apple's unified memory architecture, which means 
data does not need to be copied between CPU and GPU memory. The model lives in shared memory and 
both processors access it directly. This eliminates the memory transfer bottleneck that limits GPU 
performance on discrete-GPU systems.

## The Engine Behind SpeakUp

Every word SpeakUp transcribes runs through whisper.cpp on your Mac's GPU. No cloud server, no API 
call, no network roundtrip. The combination of Whisper's accuracy, whisper.cpp's efficiency, and 
Apple Silicon's Metal GPU makes on-device dictation practical for the first time — not as a 
compromise, but as the superior option.

If you want to see these benchmarks in action, [download SpeakUp's free 14-day 
trial](https://getspeakup.app/get/) and test it on your own Mac. The speed difference from cloud 
dictation tools like [Wispr Flow](https://getspeakup.app/vs/wispr-flow/) is immediately apparent 
— and your voice data never leaves your machine.
