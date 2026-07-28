---
Title: Playlist Corpus Operator Playbook
Ticket: VIDEO-CORPUS-PIPELINE
Status: active
Topics:
    - transcription
    - asr
    - video
    - playlist
    - corpus
    - pipeline
    - sqlite
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/Movies/richard-southwell-category-theory-for-beginners/media-manifest.json
      Note: Source media inventory for the immediate corpus
    - Path: /home/manuel/worktrees/2026-07-28--transcription-go-video-pipeline/cmd/transcribe/batch.go
      Note: |-
        Current working single-file command used for pre-implementation smoke tests
        Current smoke command implementation
    - Path: /home/manuel/worktrees/2026-07-28--transcription-go-video-pipeline/internal/server/dagger.go
      Note: |-
        Service startup and health behavior operators will observe
        Service logs and health behavior used in operations
ExternalSources: []
Summary: Preflight, execution, monitoring, recovery, validation, and search procedures for the proposed corpus command.
LastUpdated: 2026-07-28T00:00:00Z
WhatFor: Give operators a safe and repeatable procedure for running and resuming the playlist transcription pipeline.
WhenToUse: Use before a real corpus run, after interruption, or when validating exports and database completeness.
---


# Playlist Corpus Operator Playbook

## Status of this playbook

The corpus commands below are the intended interface and are not implemented at ticket creation time. Until implementation, use `transcribe batch` only for individual smoke files. Do not mistake the existing ticket-local shell loop for a warm-service corpus run.

## Paths for the Southwell corpus

```bash
export REPO=/home/manuel/worktrees/2026-07-28--transcription-go-video-pipeline
export MEDIA=/home/manuel/Movies/richard-southwell-category-theory-for-beginners
export MANIFEST=$MEDIA/transcription-manifest.json
export DB=$MEDIA/corpus-nemotron.db
export OUT=$MEDIA/transcripts-nemotron
```

## Preflight

### 1. Confirm repository and branch

```bash
cd "$REPO"
git status --short --branch
```

Expected branch:

```text
feature/video-pipeline-corpus
```

Do not run from the original checkout when implementation testing should remain isolated.

### 2. Confirm disk capacity

```bash
df -h "$MEDIA"
du -sh "$MEDIA"
```

The existing media directory already contains videos and normalized WAV derivatives. Leave margin for the corpus database, temporary uploads, and exports.

### 3. Validate the manifest without starting the model

```bash
./transcribe corpus run \
  --manifest "$MANIFEST" \
  --database "$DB" \
  --output-dir "$OUT" \
  --dry-run
```

Expected Southwell planning totals at the beginning:

```text
manifest=37 available=36 unavailable=1
```

The exact pending/complete counts depend on prior runs. Item 35 is members-only and must remain visible as unavailable.

### 4. Run repository tests

```bash
make test
```

For focused corpus work:

```bash
go test ./internal/corpus ./internal/asr ./internal/output -count=1
```

## Build

```bash
make build
./transcribe --help
./transcribe corpus --help
```

## Smoke before full execution

Use video 019 because it is short and already has a known successful Nemotron result.

```bash
./transcribe corpus run \
  --manifest "$MANIFEST" \
  --database "$DB" \
  --output-dir "$OUT" \
  --source-id RZPeFGg84Ng \
  --format srt,vtt,txt \
  --server-dir "$REPO/server" \
  --verbose
```

Validate:

```bash
sqlite3 "$DB" '
SELECT v.source_id, v.processing_state, r.word_count
FROM videos v
LEFT JOIN transcript_revisions r ON r.id = v.active_revision_id
WHERE v.source_id = "RZPeFGg84Ng";
'

find "$OUT" -path '*RZPeFGg84Ng*' -type f -maxdepth 3 -print
```

The current reference run produced 467 words, but do not hard-code that as an eternal model-quality invariant unless model and dependency revisions are pinned.

## Full run

Run long work under tmux so terminal/session interruption does not kill it accidentally:

```bash
tmux new-session -d -s southwell-corpus \
  "cd '$REPO' && ./transcribe corpus run \
    --manifest '$MANIFEST' \
    --database '$DB' \
    --output-dir '$OUT' \
    --format srt,vtt,txt \
    --server-dir '$REPO/server' \
    --verbose 2>&1 | tee '$MEDIA/corpus-nemotron.log'"
```

Inspect without attaching:

```bash
tmux capture-pane -pt southwell-corpus:0 | tail -100
```

Or attach:

```bash
tmux attach -t southwell-corpus
```

## Monitoring queries

### Corpus state counts

```sql
SELECT processing_state, COUNT(*)
FROM videos
GROUP BY processing_state
ORDER BY processing_state;
```

### Latest attempts

```sql
SELECT
  v.playlist_position,
  v.source_id,
  a.attempt_number,
  a.state,
  a.word_count,
  a.processing_ms,
  a.error_class,
  a.error_message
FROM transcript_attempts a
JOIN videos v ON v.id = a.video_id
ORDER BY a.id DESC
LIMIT 20;
```

### Validate active revisions

```sql
SELECT COUNT(*) AS invalid_active_revision_links
FROM videos v
JOIN transcript_revisions r ON r.id = v.active_revision_id
WHERE r.video_id != v.id;
```

Expected: `0`.

### Validate word counts

```sql
SELECT r.id, r.word_count, COUNT(w.id) AS actual
FROM transcript_revisions r
LEFT JOIN words w ON w.revision_id = r.id
GROUP BY r.id
HAVING r.word_count != COUNT(w.id);
```

Expected: no rows.

## Interruption and resume

Stop only when necessary:

```bash
tmux send-keys -t southwell-corpus:0 C-c
```

Then inspect status:

```bash
./transcribe corpus status --database "$DB"
```

Resume:

```bash
./transcribe corpus run \
  --manifest "$MANIFEST" \
  --database "$DB" \
  --output-dir "$OUT" \
  --format srt,vtt,txt \
  --server-dir "$REPO/server"
```

Completed videos with matching source and pipeline fingerprints must be skipped. A stale `running` attempt from the interrupted process should be marked abandoned before retry.

## Retrying failures

List failures:

```bash
./transcribe corpus status --database "$DB" --state failed
```

Retry all failed items:

```bash
./transcribe corpus run \
  --manifest "$MANIFEST" \
  --database "$DB" \
  --retry-failed
```

Retry selected source IDs:

```bash
./transcribe corpus retry \
  --database "$DB" \
  --source-id ID1 \
  --source-id ID2
```

Do not retry `members_only`, `private`, or `deleted` entries as ASR failures. Fix acquisition first and import a new manifest.

## Export repair

If transcript rows exist but files were deleted:

```bash
./transcribe corpus export \
  --database "$DB" \
  --output-dir "$OUT" \
  --format srt,vtt,txt
```

This must not start Dagger or call Nemotron.

## Search validation

```bash
./transcribe corpus search \
  --database "$DB" \
  --query "Yoneda lemma" \
  --limit 10
```

Every result must include:

- source ID;
- playlist position and title;
- start/end seconds;
- transcript excerpt;
- source URL with timestamp;
- active revision/pipeline identity.

Open a local video timestamp with a suitable player or use the source deep link to spot-check alignment.

## Failure triage

### Service fails before health

Inspect Dagger output for dependency installation, model download, Uvicorn startup, or port/tunnel errors. No per-video attempt should begin before health passes.

### HTTP request consumes too much memory

Confirm whether the streaming multipart implementation is in use. The old client builds the entire WAV body in a `bytes.Buffer` and is unsuitable for corpus-scale long files.

### Timestamps exceed video duration

Treat as a failed validation. Inspect chunk overlap and timestamp offsets; do not clamp and commit silently.

### Final SRT cue ends at zero

This indicates the trailing-segment bug in `internal/output/format.go` was not fixed or the export is stale. Regenerate only after correcting and testing the segment builder.

### Database says complete but exports are missing

Run export repair. Do not retranscribe.

### Database integrity error

```bash
sqlite3 "$DB" 'PRAGMA integrity_check; PRAGMA foreign_key_check;'
```

Stop writes before attempting manual repair. Preserve a copy of the database and logs.

## Completion audit

```bash
./transcribe corpus status --database "$DB"
sqlite3 "$DB" 'PRAGMA integrity_check; PRAGMA foreign_key_check;'
```

Acceptance for the Southwell source:

- 37 source entries represented;
- 36 available entries complete, or failures explicitly listed;
- one members-only entry unavailable;
- no running attempts left;
- word/revision counts agree;
- requested exports have hashes and files;
- search returns timestamped source evidence;
- a no-change rerun schedules zero ASR jobs.
