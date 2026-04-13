---
Title: Word-level analysis report for live vs batch transcript output
Ticket: LIVE-TRANSCRIPTION
Status: active
Topics:
    - go
    - dagger
    - asr
    - streaming
    - transcription
    - sqlite
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: out-batch-clip-000-120/transcript.db
      Note: Batch baseline subset DB inspected directly in the word analysis report
    - Path: out-live-clip-000-120/transcript.db
      Note: Live replay subset DB inspected directly in the word analysis report
    - Path: ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/01-compare_transcript_dbs.py
      Note: Count/coverage comparison helper used for full-run and subset comparisons
    - Path: ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/02-extract_wav_segment.py
      Note: WAV slicing helper used to create the 120-second fast-iteration subset
    - Path: ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/03-word_diff_report.py
      Note: |-
        Word-by-word diff report generator used for the 120-second analysis
        Reusable script used to produce the ticket's word-level diff evidence
    - Path: ttmp/out-batch-clip-000-120/transcript.db
      Note: 120s batch transcript database used as same-pipeline baseline
    - Path: ttmp/out-live-clip-000-120/transcript.db
      Note: 120s live replay transcript database used for direct word-level inspection
    - Path: ttmp/out-live-e2e/transcript.db
      Note: Full-run live replay transcript database used for final count-level comparison
ExternalSources: []
Summary: Detailed word-level analysis of the 120-second live replay vs batch baseline, plus context from the completed full-run comparison. Focuses on what words are actually missing, substituted, or duplicated in the current live path.
LastUpdated: 2026-04-13T00:00:00Z
WhatFor: Explain the current live-vs-batch word delta with direct evidence from the transcript databases.
WhenToUse: Use when debugging the remaining word-loss and transcript-quality differences in the live replay path.
---


# Word-level analysis report for live vs batch transcript output

## Executive summary

I compared the live replay transcript database directly against a same-pipeline batch transcript database on a 120-second clipped WAV, and also recorded the completed full-run comparison for context.

The high-level findings are:

- the **completed full run** ended at `3845` live words vs `4248` words in the older reference DB (`-403`),
- the **120-second same-pipeline subset** ended at `302` live words vs `323` batch words (`-21`),
- direct word inspection on the 120-second subset shows the current live gap is **not one single failure mode**,
- the remaining difference is a mix of:
  - genuine missing function/content words,
  - transcription substitutions,
  - phrase compression,
  - some places where the live output actually removes duplicated junk that the batch output kept.

That last point matters: the raw word-count delta overstates how much worse the live output is, because some of the “missing” words are repeated artifacts in the batch baseline rather than obviously desirable words.

## Inputs and method

### Full-run context

Completed full replay artifacts:

- `out-live-e2e/transcript.db`
- `out-live-e2e/live-summary.json`

Full-run comparison command:

```bash
python3 ./ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/01-compare_transcript_dbs.py \
  --live-db out-live-e2e/transcript.db \
  --reference-db /home/manuel/code/wesen/2026-04-09--screencast-studio/ttmp/2026/04/13/TRANSCRIPT-PIPELINE--setting-up-an-analysis-pipeline-for-transcripts/sources/audio_transcript.db \
  --summary-json out-live-e2e/live-summary.json
```

### Fast 120-second subset

Created with:

```bash
python3 ./ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/02-extract_wav_segment.py \
  --input /home/manuel/code/wesen/2026-04-09--screencast-studio/recordings/rabbit-hole-2026-04-10--2/audio-mix.wav \
  --output /tmp/transcription-live-clip-000-120.wav \
  --start 0 \
  --duration 120
```

Batch baseline:

- `out-batch-clip-000-120/transcript.db`
- `323` words

Live replay output:

- `out-live-clip-000-120/transcript.db`
- `302` committed words

Word-level diff command:

```bash
python3 ./ttmp/2026/04/13/LIVE-TRANSCRIPTION--live-transcription-architecture-and-implementation-plan/scripts/03-word_diff_report.py \
  --batch-db out-batch-clip-000-120/transcript.db \
  --live-db out-live-clip-000-120/transcript.db
```

## Count-level results

### Full run

- live words: `3845`
- reference words: `4248`
- delta: `-403`
- live chunk count: `352`
- reference chunk count: `196`
- coverage delta: `+0.14s`

### 120-second subset

- live words: `302`
- batch words: `323`
- delta: `-21`
- live chunks processed: `27`
- live chunk count in DB: `25`
- batch chunk count in DB: `16`
- coverage delta: `+0.04s`

### Interpretation

The 120-second subset is the better debugging target because it compares the current live path against the current batch path under the same codebase and finishes quickly. The `-21` delta is small enough that direct inspection is practical and meaningful.

## Direct word-diff findings

The 120-second diff produced:

- `32` non-equal regions
- opcode mix:
  - `22` deletes
  - `5` inserts
  - `5` replaces

Affected token counts in the normalized sequence diff:

- batch-side words affected: `35`
- live-side words affected: `14`

That pattern means the live output is usually **shorter/compressed**, not simply equally long with lots of substitutions.

## Main mismatch classes

## 1. Live output drops some repeated/junky words that batch kept

Examples:

### Opening branding region

Batch:

```text
Welcome back to the Go Go Golems lab ...
```

Live:

```text
Welcome back to the Go Golems lab.
```

### Rabbit-hole repetition region

Batch:

```text
... rabbit hole. So So my my today's today's rabbit rabbit hole. hole ...
```

Live:

```text
... rabbit hole. hole. my today's rabbit hole ...
```

Interpretation:

Some of the live word-count deficit is actually the live path **collapsing duplicated noise** that the batch baseline preserved. This means the `-21` delta should not be interpreted as 21 unquestionably good words being lost.

## 2. Live output has genuine recognition/substitution errors

Examples:

### “Golems” → “Columns”

Batch:

```text
... Go Go Golems lab today ...
```

Live:

```text
... Go Columns lab. Today ...
```

### “rabbit hole in” → “rabbit holeing”

Batch:

```text
... rabbit hole in with LLMs ...
```

Live:

```text
... rabbit holeing with LLMs ...
```

### “trust that” → “try”

Batch:

```text
... I don't really trust that it was really fast ...
```

Live:

```text
... I don't really try. it. was really fast ...
```

Interpretation:

These are not accumulator-only effects. They are actual recognition/tokenization differences across the chunked live path.

## 3. Live output loses small function-word bridges

There are many one-word deletions in otherwise aligned phrases. Examples include missing words like:

- `uh`
- `what`
- `do`
- `but`
- `end`
- `the`
- `random`
- `so`
- `you`

Representative windows:

### Around “what you can do with LLMs”

Batch:

```text
... what you can do with LLMs, what can tools do that are out there, but also just pure ...
```

Live:

```text
... you can with LLMs, can tools do that are out there, also just pure ...
```

### Around “end of the day”

Batch:

```text
... at the end of the day, right? So let's imagine ...
```

Live:

```text
... at the of day, right? let's imagine ...
```

### Around “something random”

Batch:

```text
... do something random like I have a certain task ...
```

Live:

```text
... do something. like I have a certain task ...
```

Interpretation:

The live path is frequently shaving off small connector words. This is the strongest evidence that the current live pipeline is still slightly over-compressing content at chunk boundaries or during dedupe/finalization.

## 4. Live output also introduces a few extra junk tokens

Examples:

### “M I” artifact

Batch:

```text
... it's an LLM I don't really see what I have ...
```

Live:

```text
... it's an LLM I. M I don't really see what I have ...
```

### “P is” artifact

Batch:

```text
... get computers to do square roots ...
```

Live:

```text
... get computers to P is to do square roots ...
```

### “That one” insertion

Batch:

```text
... pretty stupid one. It's uh I want to do ...
```

Live:

```text
... pretty stupid one. That one it's uh I want to do ...
```

Interpretation:

The live path is not simply “missing” words; it is also occasionally fabricating short junk insertions. That suggests the live chunk path is producing slightly different local hypotheses around chunk transitions.

## Region-by-region notes

These are the most obviously useful debugging windows from the 120s diff.

## Region A — 0s to ~14s

Symptoms:

- missing duplicate `Go`
- `Golems` becomes `Columns`
- live inserts `Slab`

Interpretation:

This is mostly recognition quality, not a dedupe bug.

## Region B — ~36s to ~41s

Symptoms:

- `what you can do with LLMs` becomes `you can with LLMs`
- `but` disappears

Interpretation:

This looks like both recognition simplification and loss of connector words.

## Region C — ~45s to ~50s

Symptoms:

- `end of the day` becomes `of day`
- `So` before `let's imagine` disappears
- `random` disappears

Interpretation:

This is one of the clearest clusters of true live-path word loss.

## Region D — ~60s to ~68s

Symptoms:

- repeated `So So my my today's today's rabbit rabbit hole` in batch
- much shorter condensed phrase in live
- live inserts `That one`

Interpretation:

This region contains both:

- likely beneficial removal of repeated garbage,
- and some live-specific insertion oddities.

## Region E — ~86s to ~92s

Symptoms:

- `I wanna learn more about how to get computers to do square roots`
- live introduces `M I` and `P is`

Interpretation:

This is a strong chunk-transition corruption region and worth targeting directly.

## Region F — ~108s to ~112s

Symptoms:

Batch:

```text
... I don't really trust that it was really fast but I'm not really sure so ...
```

Live:

```text
... I don't really try. it. was really fast, I'm really sure. what ...
```

Interpretation:

This is a real semantic degradation region, not just punctuation drift.

## Region G — ~116s to ~120s

Symptoms:

Batch:

```text
... light thinking say you know so first step of the rabbit hole
```

Live:

```text
... light thinking say, know So first step of rabbit hole.
```

Interpretation:

Another small but clean example of function-word loss (`you`, `the`).

## Overall interpretation

The current live-vs-batch word delta on the 120-second subset is best described as:

1. **partly explained by removing duplicated junk** that batch preserved,
2. **partly explained by ordinary ASR substitutions**, and
3. **partly explained by real loss of short connector words and phrase structure**.

That means the remaining problem is not a single obvious “the accumulator deletes whole chunks” bug. Instead, it looks like:

- chunk-local recognition differs from full/batch recognition,
- overlap and boundary handling still matter,
- the current live path is slightly too lossy around small words and transitions.

## Strongest current hypotheses

## Hypothesis 1: 5s chunking + 0.5s overlap is still too lossy for some phrases

The live path is doing inference on smaller independent chunks than batch, so some of the differences are probably caused by reduced context.

## Hypothesis 2: the current Phase 1 accumulator is a little too aggressive for near-equal boundary words

The accumulator rejects words if:

- `w.End <= lastFinalTime + tolerance`
- or they look near-duplicate in word+timing

This is probably harmless in many places, but it may clip legitimate short words at boundaries.

## Hypothesis 3: some of the count delta is not worth “fixing”

If batch contains duplicated garbage and live removes it, a lower live word count is not automatically worse. We should optimize for transcript quality, not only raw count parity.

## Recommended next actions

## 1. Keep the 120s subset as the default debugging loop

It is much faster and the `-21` delta is small enough to inspect directly.

## 2. Test chunk-duration sensitivity on the same 120s clip

Run the same live pipeline with:

- `--chunk-duration 5`
- `--chunk-duration 10`
- maybe `--chunk-duration 15`

and compare the resulting DBs against the same batch baseline.

## 3. Inspect boundary-word handling in the accumulator

Focus especially on regions where the live output loses:

- `what`
- `but`
- `end`
- `the`
- `you`
- `so`

## 4. Do not optimize purely for word-count equality

Some of the missing live words are duplicated artifacts in the batch baseline. The objective should be better transcript quality, not just “make the counts match”.

## Conclusion

The word-level inspection confirms that the current live path is already structurally good enough to compare directly against batch output, but it still has a small, real quality gap on the 120-second subset.

That gap is not dominated by catastrophic omissions. It is mostly made of:

- a handful of recognition substitutions,
- dropped connector words,
- chunk-boundary phrase compression,
- offset by occasional removal of duplicated junk from the batch baseline.

So the next debugging target is clear:

> use the 120-second clipped-WAV workflow to tune chunk size and boundary/dedupe behavior, rather than repeatedly rerunning the full 27-minute file.
