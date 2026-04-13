# Tasks

## TODO

- [ ] Validate the ticket with `docmgr doctor --ticket LIVE-TRANSCRIPTION --stale-after 30`
- [ ] Upload the design-doc bundle to reMarkable and verify the remote listing
- [ ] If implementation begins immediately, add a dedicated API-contract reference doc for chunk and WebSocket message schemas
- [ ] If implementation begins immediately, add a replay-test playbook for pseudo-live validation

## DONE

- [x] Create ticket workspace `LIVE-TRANSCRIPTION`
- [x] Add primary design doc
- [x] Add investigation diary
- [x] Map current repository architecture against the batch pipeline
- [x] Write phased implementation plan for live transcription
- [x] Recommend the long-term production architecture (session-oriented streaming service)

## Proposed implementation phases

- [x] Phase 0 — preserve batch mode and refactor for coexistence
- [x] Phase 1 — add chunk-oriented near-live mode
- [x] Phase 2 — formalize transcript accumulator and live persistence semantics
- [x] Phase 3 — introduce true streaming session API over WebSocket
- [x] Phase 4 — production hardening
