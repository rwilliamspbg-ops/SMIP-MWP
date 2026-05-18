# Changelog

All notable changes to this project are documented in this file.

## Unreleased (2026-05-18)

- Tier-1 performance optimizations implemented and merged (PR #15):
  - Sharded session map (16 shards) to reduce RWMutex contention
  - Per-worker session cache (hot-session circular buffer)
  - Pre-allocated descriptor reuse in AF_XDP poll loop
  - Per-worker packet pools replacing hot `sync.Pool` usage
- Measured development harness results (60s runs):
  - AF_XDP single-worker (crypto): 2014 ns/op, 560 B/op, 9 allocs/op
  - AF_XDP multi-worker (4 workers, crypto): 2376 ns/op, 632 B/op, 11 allocs/op
  - Crypto microbench highlights: `NewHybridSession_Cached` 545 ns/op; `EncryptInPlace` 833 ns/op
- Added: `benchmarks/pprof/run_pprof_bench.sh`, `benchmarks/HARDWARE_VALIDATION_RUNBOOK.md`, TRex/MoonGen templates and QUIC run scripts.

