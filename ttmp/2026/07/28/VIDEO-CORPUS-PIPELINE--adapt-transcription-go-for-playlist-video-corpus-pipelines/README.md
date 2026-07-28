# VIDEO-CORPUS-PIPELINE

Design and implementation workspace for adapting `transcription-go` from one-file transcription to a resumable playlist-wide video corpus pipeline.

## Read in this order

1. `analysis/01-current-system-and-video-corpus-gap-analysis.md`
2. `design-doc/01-intern-guide-playlist-video-corpus-transcription-architecture-and-implementation.md`
3. `reference/01-corpus-database-and-pipeline-api-contracts.md`
4. `playbook/01-playlist-corpus-operator-playbook.md`
5. `reference/02-investigation-diary.md`

## Main decision

Use one warm Nemotron service per corpus run and one canonical corpus SQLite database. Preserve timed words as evidence; derive chunks, search indexes, and SRT/VTT/TXT exports.

## Immediate implementation phases

- Correct current formatter and legacy SQLite issues.
- Add normalized manifest parsing and validation.
- Add corpus schema, fingerprinting, planning, and atomic revisions.
- Add warm-service sequential runner and resumability.
- Add export repair, status, and corpus search.
- Validate on Southwell video 019, interruption/resume, then the full accessible corpus.
