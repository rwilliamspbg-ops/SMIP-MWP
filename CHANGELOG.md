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

- Integrations and tooling updates (merged from `rust-rewrite/add-lake-generate-hkdf`):
  - Add a Lake `exe` target `GenerateHKDFVectors` to run the Lean vector generator via `lake exe GenerateHKDFVectors`.
  - CI now installs `elan` and runs the Lake `GenerateHKDFVectors` exe before generating and validating Go test vectors.
  - Generator (`tools/leanvecgen`) updated to derive HKDF vectors from the production Go implementation to ensure generated tests match runtime behavior.
  - Lean sources: small fixes to `CryptoHandshakeSpec` and `GenerateHKDFVectors` so the Lean generator builds under Lake.

- Performance and Rust scaffold (merged from `perf/afxdp-bench-allocs2`):
  - Added extensive AF_XDP benchmarking artifacts and PProf outputs under `benchmarks/`.
  - Introduced a `rust_rewrite/` scaffold with AF_XDP, crypto, routing and wire Rust crates and initial benchmark and integration test harnesses.
  - Added formalization checklist and additional Lean specification modules under `formal/lean4`.


