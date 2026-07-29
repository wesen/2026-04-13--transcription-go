**Update 2026-07-15:** [Round 2 is 
live](https://get-inscribe.com/blog/parakeet-moss-apple-speech-benchmark.html), adding NVIDIA 
Parakeet TDT v2/v3 (the comparison most readers asked for) and MOSS-Transcribe-Diarize, measured on 
the identical corpus and scoring.

## The result, up front

Apple's new SpeechAnalyzer is the most accurate on-device speech engine we tested. It beat every 
Whisper model we shipped at the time, including Whisper Small, on both the clean and the noisy half 
of LibriSpeech, while running roughly three times faster than Small. And the API it replaces, 
SFSpeechRecognizer, came last on clean speech: behind even Whisper Tiny, a 40MB model.

| Engine | test-clean WER | test-other WER | Model size |
| --- | --- | --- | --- |
| **Apple SpeechAnalyzer** (iOS/macOS 26) | **2.12%** | **4.56%** | system |
| Whisper Small (WhisperKit CoreML) | 3.74% | 7.95% | ~460MB |
| Whisper Base | 5.42% | 12.51% | ~140MB |
| Whisper Tiny | 7.88% | 17.04% | ~40MB |
| Apple SFSpeechRecognizer (legacy) | 9.02% | 16.25% | system |

Lower is better: WER is word error rate, the percentage of words an engine substitutes, drops, or 
invents. LibriSpeech test-clean is 2,620 utterances of clean read speech; test-other is 2,939 
harder, noisier utterances. Every engine ran fully on-device on an Apple M2 Pro (32GB, macOS 
26.5.1).

Apple SpeechAnalyzer2.12%

Whisper Small3.74%

Whisper Base5.42%

Whisper Tiny7.88%

SFSpeechRecognizer (legacy)9.02%

## Why we ran this

With iOS 26 and macOS 26, Apple replaced SFSpeechRecognizer with a new API, SpeechAnalyzer and 
SpeechTranscriber. It published no accuracy figures for either one. So every developer deciding 
whether to migrate, and everyone comparing Apple's built-in recognition against Whisper, has been 
guessing.

At the time of this benchmark we shipped both Apple engines and three Whisper models side by side 
in [Inscribe](https://get-inscribe.com/index.html), a private on-device AI workspace, which put us 
in an unusual position: we could run all five through identical production code paths on the same 
machine and the same audio. So we did.

## Should you migrate off SFSpeechRecognizer?

Yes. This is the clearest result in the data. The new API cuts word error rate by 3.5 to 4x on the 
same audio: from 9.02% to 2.12% on clean speech, and from 16.25% to 4.56% on noisy speech. There is 
no accuracy trade-off to weigh; the new API wins everywhere we measured, and it produces 
punctuated, cased text where the legacy engine's output is rougher.

Put differently: an hour-long meeting transcribed with the legacy API contains roughly four times 
as many wrong words as the same meeting through SpeechAnalyzer. If your app still uses 
SFSpeechRecognizer for anything longer than a voice command, the migration is worth it on accuracy 
alone.

## SpeechAnalyzer vs Whisper

The more surprising result: Apple's new engine also beat Whisper Small, the largest model we 
shipped, by a comfortable margin on both splits, at roughly a third of Whisper Small's compute time 
per second of audio. For English, on Apple hardware, the built-in engine is now the strongest 
on-device option we can measure.

Whisper keeps two real advantages. It covers far more languages (SpeechTranscriber supports around 
30 locales), and it runs anywhere, not just on Apple platforms with OS 26. But for English 
transcription on a current iPhone or Mac, the days of Whisper being the automatic accuracy pick are 
over.

We changed our own product on this result: Inscribe's Auto engine now prefers SpeechAnalyzer for 
the languages it supports. (It went further after [round 
2](https://get-inscribe.com/blog/parakeet-moss-apple-speech-benchmark.html): as of Inscribe 1.8.9 
the Whisper models are retired entirely, replaced by NVIDIA Parakeet as the fallback engine.) 
Shipping a benchmark and ignoring it in your own defaults would be a strange kind of honesty.

## Speed

All five engines ran comfortably faster than real time: between roughly 12x and 40x on the M2 Pro, 
meaning an hour of audio transcribes in about 1.5 to 5 minutes on-device. SpeechAnalyzer was about 
3x faster than Whisper Small per second of audio while beating it on accuracy. We are deliberately 
not printing a precise per-engine timing table yet: the accuracy runs shared the machine with a 
development workload, which does not affect WER but does add noise to timing. We will update this 
page with timings from a dedicated idle run.

## Methodology, and why you can check it

A benchmark from a company that sells one of the engines should be treated with suspicion. Ours has 
two properties designed for that suspicion.

### The Whisper column is reproducible against OpenAI's own numbers

We used LibriSpeech precisely because OpenAI published Whisper's WER on it. If our harness measured 
Whisper correctly, our numbers should land on theirs. They do, on all six measurements:

| Engine / split | Ours | OpenAI published | Delta |
| --- | --- | --- | --- |
| Whisper Tiny, test-clean | 7.88% | 7.6% | +0.28 |
| Whisper Base, test-clean | 5.42% | 5.0% | +0.42 |
| Whisper Small, test-clean | 3.74% | 3.4% | +0.34 |
| Whisper Tiny, test-other | 17.04% | 16.9% | +0.14 |
| Whisper Base, test-other | 12.51% | 12.4% | +0.11 |
| Whisper Small, test-other | 7.95% | 7.6% | +0.35 |

The small, consistent positive offset (a slightly stricter text normalizer plus CoreML 
quantization) is what honest reproduction looks like; random error would scatter in both 
directions. Since the same corpus, normalizer, and scorer produced the Apple columns, the numbers 
nobody else can check inherit the validation from the numbers anyone can.

### The raw transcripts are public

Every per-utterance hypothesis for both Apple engines is downloadable below, next to the reference 
text and per-utterance WER. Disagree with our normalization? Rescore it yourself.

- [summary.json](https://get-inscribe.com/data/speech-benchmark/summary.json) - all ten 
measurements, machine-readable (3KB)
- 
[raw-transcripts-apple.json.gz](https://get-inscribe.com/data/speech-benchmark/raw-transcripts-apple
.json.gz) - SpeechAnalyzer, all 5,559 utterances (620KB)
- 
[raw-transcripts-legacy.json.gz](https://get-inscribe.com/data/speech-benchmark/raw-transcripts-lega
cy.json.gz) - SFSpeechRecognizer, all 5,559 utterances (620KB)

### Details that decide whether a WER number means anything

- **Same production code paths.** Each engine ran through the exact code Inscribe users get, not a 
lab harness with different buffering or settings.
- **Text normalization.** LibriSpeech references are uppercase, unpunctuated, with numbers spelled 
out; modern engines emit punctuation and digits. Both sides pass through the same normalizer 
(casing, punctuation, digits-to-words, contractions), mirroring OpenAI's English normalizer. Score 
raw text and you punish engines for formatting nicely rather than for mishearing.
- **Corpus WER, not averaged WER.** Total errors divided by total reference words, so short 
utterances are not over-weighted.
- **Fully on-device, verified.** SFSpeechRecognizer sends audio to Apple's servers by default. We 
forced on-device recognition and made the harness refuse to run rather than silently fall back to 
the cloud, both because a cloud result would invalidate the comparison and because we were not 
going to upload 5,559 utterances from a privacy product.
- **Failures counted, not hidden.** An engine returning nothing scores 100% WER for that utterance. 
It happened once in 27,795 transcriptions (legacy, test-other).

## What building this taught us about our own app

The benchmark found a shipping bug in Inscribe. Our Apple-engine file import fed audio to 
SpeechAnalyzer and closed the input stream, but never called 
`finalizeAndFinishThroughEndOfInput()`. Without that call the analyzer never delivers its final 
results, and the import hangs forever. It had gone unnoticed because our Auto setting preferred 
Whisper. The fix shipped the same day, and it is part of why we publish the harness details: 
measuring your own product carefully has a way of finding the things you were not looking for.

## Limitations

- **English only.** LibriSpeech is English read speech. These numbers say nothing about the 100+ 
languages Whisper supports that SpeechTranscriber does not.
- **Read audiobook speech, not meetings.** LibriSpeech is the standard, comparable corpus, which is 
why we started with it. Accented, far-field, and multi-speaker meeting audio is the obvious 
follow-up.
- **One machine.** M2 Pro, macOS 26.5.1. Accuracy should transfer across Apple Silicon; speed will 
vary by chip.
- **Whisper via WhisperKit CoreML.** Quantized on-device conversions, the same builds Inscribe 
shipped at the time (since 1.8.9 the app ships NVIDIA Parakeet instead). Reference GPU 
implementations may differ slightly, which the validation table quantifies.

## What this means if you just want good transcription

If you are on a current iPhone or Mac, the best on-device transcription engine for English is 
already in the operating system, and the private option is no longer the compromise option. 
Inscribe uses exactly the engines measured here: SpeechAnalyzer where it supports your language, 
Whisper where it does not, all fully on-device, nothing uploaded. The benchmark is not separate 
from the product; it is how we decide what the product does.
