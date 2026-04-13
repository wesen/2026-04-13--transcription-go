# Tasks

## TODO

- [ ] Investigate 22-word delta versus reference transcript (`4248` reference vs `4226` Go pipeline output)
- [ ] Add server-side or client-visible chunk progress metrics beyond the current heartbeat logs if finer progress reporting is needed

## DONE

- [x] 1. Go module setup + deps (go-audio, dagger, cobra)
- [x] 2. Pure Go audio conversion (`internal/convert`) + tests
- [x] 3. Python ASR server (`server/server.py`)
- [x] 4. Dagger service lifecycle (`internal/server`) — final fix: run uvicorn via `AsService(...Args...)`, start the host tunnel, and use a random host endpoint
- [x] 5. ASR HTTP client (`internal/asr`) — POST /transcribe/full, parse JSON words
- [x] 6. Output formatters (`internal/output`) — SQLite, SRT, VTT, TXT, filler filtering
- [x] 7. CLI entrypoint (`cmd/transcribe/main.go`) — wire everything together
- [x] 8. End-to-end test with rabbit-hole recording
- [x] 9. Makefile + .gitignore

## E2E Result Summary

- Input: `/home/manuel/code/wesen/2026-04-09--screencast-studio/recordings/rabbit-hole-2026-04-10--2/audio-mix.wav`
- Converted audio duration: `1664.8s` (~27.7 min)
- Chunk size: `60s`
- Output words: `4226`
- Reference words: `4248`
- Delta: `-22`
- Outputs written: `out/transcript.srt`, `out/transcript.db`
