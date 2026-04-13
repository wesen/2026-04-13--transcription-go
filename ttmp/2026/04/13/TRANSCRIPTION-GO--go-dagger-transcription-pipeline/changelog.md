# Changelog

## 2026-04-13

- Initial workspace created


## 2026-04-13

Created ticket, wrote design doc with full architecture analysis, implementation plan, and risk assessment. Reviewed 5 reference files (handoff doc, 2 Dagger Go implementations, 2 Python pipeline scripts).


## 2026-04-13

Revised architecture (v2): ffmpeg in Go, long-running ASR server via Dagger Service, all output formatting in Go. Verified Dagger SDK v0.20.5 service API (AsService, Host.Tunnel, Service.Endpoint). Added v1 vs v2 comparison table.


## 2026-04-13

v3: Eliminated ffmpeg dependency. Pure Go audio conversion (go-audio/wav + oov/audio/resampler). Benchmarked: 2.9s Go vs 3.5s ffmpeg on rabbit-hole recording. Zero host deps beyond Go + Dagger.

