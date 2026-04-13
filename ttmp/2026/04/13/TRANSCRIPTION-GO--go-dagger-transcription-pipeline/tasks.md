# Tasks

## TODO

- [ ] 8. End-to-end test with rabbit-hole recording

## DONE

- [x] 1. Go module setup + deps (go-audio, dagger, cobra)
- [x] 2. Pure Go audio conversion (`internal/convert`) + tests
- [x] 3. Python ASR server (`server/server.py`)
- [x] 4. Dagger service lifecycle (`internal/server`) — start, tunnel, health check, stop
- [x] 5. ASR HTTP client (`internal/asr`) — POST /transcribe/full, parse JSON words
- [x] 6. Output formatters (`internal/output`) — SQLite, SRT, VTT, TXT, filler filtering
- [x] 7. CLI entrypoint (`cmd/transcribe/main.go`) — wire everything together
- [x] 9. Makefile + .gitignore
