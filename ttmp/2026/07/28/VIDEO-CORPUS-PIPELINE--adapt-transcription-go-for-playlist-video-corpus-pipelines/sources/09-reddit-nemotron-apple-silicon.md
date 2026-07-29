NVIDIA released Nemotron-3.5-ASR-Streaming-0.6B last month. Same FastConformer + RNN-T family as 
their earlier English-only streaming model, but trained on 40 language-locales and conditioned by a 
prompt kernel that injects a one-hot language slot into every encoded frame — so the same 600M 
weights serve every language. Ported it to Apple Silicon end-to-end:

- **CoreML INT8** bundle that routes through the Apple Neural Engine (encoder + RNN-T decoder + 
joint network)
- **MLX bf16 / 8-bit / 4-bit** bundles for GPU-resident inference
	WER (FLEURS test, 50 samples/language, M5 Pro):
	| lang | fp32 NeMo source | CoreML INT8 (ANE) | MLX bf16 | MLX 4-bit |
	| --- | --- | --- | --- | --- |
	| en\_us | 9.33 | 9.59 | 10.36 | 15.98 |
	| de\_de | 10.22 | 10.41 | 10.87 | 14.96 |
	| fr\_fr | 11.13 | 12.18 | 11.62 | 15.85 |
	| ar\_eg | 13.27 | 13.37 | 13.76 | 20.88 |
	| hi\_in | 5.26 | 4.42 | 5.36 | 8.13 |
	| ja\_jp | 16.97 \* | 17.66 \* | 17.33 \* | 19.56 \* |

\*char-level scoring, matching NVIDIA's CJK methodology.

INT8 palletization, MLX bf16, and MLX 8-bit all stay within ±0.3 pp WER of the fp32 PyTorch 
reference. MLX 4-bit costs ~6 pp on average in exchange for the smallest disk (473 MB) and 
streaming RSS (747 MB) — useful when disk or RAM is the binding constraint.

Architecture preserved — exported through `EncDecRNNTBPEModelWithPrompt` (restored from 
`EncDecHybridRNNTCTCBPEModelWithPrompt` since the prompt-RNNT target class hasn't been released 
yet):

audio → mel (NeMo FilterbankFeatures-equivalent) → 24-layer cache-aware FastConformer encoder 
(1024 hidden) → prompt kernel: Linear(1152→2048) → ReLU → Linear(2048→1024) (folds 
one-hot language\_mask into every encoded frame) → RNN-T: 2-layer LSTM predictor (640 hidden) + 
joint over 13 087 BPE

Streaming caches — attention KV \[24, 1, 56, 1024\], depthwise conv \[24, 1, 1024, 8\], mel 
`pre_cache` — flow chunk-to-chunk so context survives the 320 ms boundaries. Streaming RTF on M5 
Pro is 0.068 with the CoreML INT8 bundle (p50 chunk latency 18.6 ms, p99 23.4 ms).

Bit-identical Swift↔Python WER on 5 of 6 languages — to validate the Apple-side numbers I 
ported Whisper's `BasicTextNormalizer` + `EnglishTextNormalizer` + the English number-words state 
machine to Swift. The lone 0.04 pp Hindi delta traces to a single ANE non-determinism sample, not a 
port bug.

Bundles (Apache 2.0 SDK; bundles carry NVIDIA's eval license, linked on each model card):

- 
[https://huggingface.co/aufklarer/Nemotron-3.5-ASR-Streaming-0.6B-CoreML-INT8](https://huggingface.c
o/aufklarer/Nemotron-3.5-ASR-Streaming-0.6B-CoreML-INT8)
- 
[https://huggingface.co/aufklarer/Nemotron-3.5-ASR-Streaming-0.6B-MLX-bf16](https://huggingface.co/a
ufklarer/Nemotron-3.5-ASR-Streaming-0.6B-MLX-bf16)
- 
[https://huggingface.co/aufklarer/Nemotron-3.5-ASR-Streaming-0.6B-MLX-8bit](https://huggingface.co/a
ufklarer/Nemotron-3.5-ASR-Streaming-0.6B-MLX-8bit)
- 
[https://huggingface.co/aufklarer/Nemotron-3.5-ASR-Streaming-0.6B-MLX-4bit](https://huggingface.co/a
ufklarer/Nemotron-3.5-ASR-Streaming-0.6B-MLX-4bit)

Repo: [https://github.com/soniqo/speech-swift](https://github.com/soniqo/speech-swift) Guide: 
[https://soniqo.audio/guides/nemotron](https://soniqo.audio/guides/nemotron)

---

## Comments

> **Plane-Fun-6105** · 
[2026-06-04](https://reddit.com/r/nvidia/comments/1twu0ow/nvidia_nemotron35_asr_streaming_ported_to_
apple/opr0le9/) · 2 points
> 
> damn those latency numbers on m5 pro are actually pretty solid for real-time stuff
> 
> > **Fabulous\_Tip\_8539** · 
[2026-06-10](https://reddit.com/r/nvidia/comments/1twu0ow/nvidia_nemotron35_asr_streaming_ported_to_
apple/oqrvu4d/) · 1 point
> > 
> > Not only on the M5, but also on the iPhone 15 Pro, we can experience real-time functionality. 
I’ve just migrated it into iOS.

> **roninXpl** · 
[2026-06-11](https://reddit.com/r/nvidia/comments/1twu0ow/nvidia_nemotron35_asr_streaming_ported_to_
apple/or1qbkv/) · 1 point
> 
> Very cool library. Here are some of my test results, HTH:
> 
> | **Engine** | **Outcome** |
> | --- | --- |
> | Nemotron 3.5 ASR | Poor — drops content, mangles terms, leaks <en-US> tokens, missed 
"relying". Streaming chunking loses to offline on this overlapped audio, as predicted. |
> | Qwen3-ASR 0.6B | Unstable — loops/degenerates on >15 s spans. One win: short --context 
priming → "American Cyanamid" (a term no current engine gets). |
> | Qwen3-ASR 1.7B | Inference hangs (never produces output) — broken on this setup. |
> | Omnilingual 300M / 7B (CTC) | Phonetic noise at any size ("aerian"=American, "henivix 
pefore"=Zenivex E4). Got "relying" once. MLX backend can't even window >10 s audio. |

> **DiGiqr** · 
[2026-06-05](https://reddit.com/r/nvidia/comments/1twu0ow/nvidia_nemotron35_asr_streaming_ported_to_
apple/opvrni3/) · 1 point
> 
> Is this that model that produces completely gibberish?
> 
> > **ivan\_digital** · 
[2026-06-05](https://reddit.com/r/nvidia/comments/1twu0ow/nvidia_nemotron35_asr_streaming_ported_to_
apple/opvrt0x/) · 1 point
> > 
> > Why gibberish? It is transcription model, pretty tested. Any particular example?
> > 
> > > **DiGiqr** · 
[2026-06-05](https://reddit.com/r/nvidia/comments/1twu0ow/nvidia_nemotron35_asr_streaming_ported_to_
apple/opxpj2l/) · 0 points
> > > 
> > > Oh ok, in that case it was some different *new* model from Nvidia. I've seen some video and 
report but I don't remember what exactly it was.
