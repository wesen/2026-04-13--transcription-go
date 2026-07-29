Compare local speech recognition models by speed, accuracy, language coverage, and Apple Silicon 
performance.

Every local speech model we ship or have benchmarked — what each is good at, how fast it runs on 
Apple Silicon, and which one to pick for your language. All numbers below come from our published 
benchmark articles.

3 topic clusters10 guides and comparisons

## Local Speech Models at a Glance (2026)

The short version: on a Mac, use **Parakeet V3** for English and European languages (it's the 
default in Whisper Notes), **SenseVoice** for Chinese, Japanese, Korean, and Cantonese, and 
**Whisper Large-v3 Turbo** for everything else in its ~100-language long tail.

| Model | Best for | Accuracy | Speed on Apple Silicon | Languages | In Whisper Notes |
| --- | --- | --- | --- | --- | --- |
| Parakeet V3 (NVIDIA, 0.6B) | English & European languages | 6.32% English WER (our FLEURS test) | 
~10× faster than Whisper Turbo | 25 | Yes — default on Mac |
| Whisper Large-v3 Turbo (OpenAI, 809M) | Widest language coverage | 7.83% English WER (our FLEURS 
test) | ~5× faster than Large V3 | ~100 | Yes — optional download (~1.6 GB) |
| SenseVoice Small (FunAudioLLM) | Chinese, Japanese, Korean, Cantonese | Purpose-built for CJK | 
27-min Chinese podcast in 13.83 s (M4 Pro) | CJK focus | Yes — since v1.5.0 |
| Whisper Large V3 (OpenAI, 1.55B) | Reference baseline | Near-identical to Turbo | Baseline 
(slowest here) | ~100 | No — Turbo supersedes it |
| Voxtral (Mistral) | Speech + language understanding | See our benchmark article | Server-class 
models | Multilingual | No — covered in blog |

WER = word error rate; lower is better. Our numbers: FLEURS benchmark plus in-house tests on an M4 
Pro running Whisper Notes — full methodology in the linked benchmark articles below. All models 
listed as "in Whisper Notes" are included in the one-time price.

4

## Highest-traffic model benchmarks

Start here when the searcher is comparing local ASR model speed, decoder layers, silence 
hallucinations, and Apple Silicon performance.

Guide 1

### [Whisper Transcription Guide](https://whispernotes.app/blog/whisper-transcription)

What Whisper transcription is, every way to run it (API, CLI, apps), accuracy by model size, and 
how to run it offline.

Read more

Guide 2

### [Parakeet V3 vs Whisper](https://whispernotes.app/blog/parakeet-v3-default-mac-model)

A benchmark page for Parakeet V3, speed, silence hallucinations, and Apple Silicon transcription.

Read more

Guide 3

### [Whisper Large V3 Turbo vs V3](https://whispernotes.app/blog/introducing-whisper-large-v3-turbo)

A benchmark of Whisper Large V3 Turbo against V3, including speed, decoder layers, and Apple 
Silicon trade-offs.

Read more

Guide 4

### [Mistral Voxtral Models](https://whispernotes.app/blog/introducing-mistral-voxtral-models)

An overview of Voxtral speech models and how they compare with local transcription options.

Read more

3

## Model-specific local transcription pages

These pages connect model searches to practical Mac, iPhone, and multilingual transcription 
workflows.

Guide 1

### [MacWhisper Alternative](https://whispernotes.app/mac-whisper)

Explains how Whisper Notes uses the same local Whisper engine while adding Fn key dictation and 
iPhone capture.

Read more

Guide 2

### [SenseVoice for CJK 
Transcription](https://whispernotes.app/blog/sensevoice-fastest-cjk-transcription)

A model-focused article for Chinese, Japanese, and Korean transcription speed and accuracy.

Read more

Guide 3

### [Best Whisper App for iPhone and Mac](https://whispernotes.app/whisper-app)

A platform-focused guide for people searching for a practical Whisper app rather than a model paper.

Read more

3

## Use models in real workflows

Where these models actually get used: iPhone recording, Mac Fn-key dictation, and end-to-end 
offline transcription workflows.

Guide 1

### [Voice to Text Offline](https://whispernotes.app/voice-to-text-offline)

How local models power iPhone recording, Mac Fn key dictation, and offline voice-to-text workflows.

Read more

Guide 2

### [Offline Transcription Guides](https://whispernotes.app/offline-transcription)

The broader guide hub for users who arrived through a model search but need an end-to-end workflow.

Read more

Guide 3

### [Offline Speech to Text Complete 
Guide](https://whispernotes.app/blog/offline-speech-to-text-complete-guide)

A complete guide to local transcription, privacy, speed, supported devices, and when offline speech 
to text beats cloud transcription.

Read more
