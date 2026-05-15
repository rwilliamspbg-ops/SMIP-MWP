# Actionable optimization suggestions — pprof CPU profiles

## Crypto (`internal/crypto`)
- Hotspots: `EncryptInPlace` / AEAD Seal, HKDF key derivation helpers.
- Suggestions:
  - Reduce allocations in AEAD paths: prefer in-place sealing where possible and reuse buffers across calls.
  - Avoid short-lived slices and header copies; use a single pre-allocated frame buffer per worker.
  - Move HKDF-derived key setup out of the per-packet path (cache derived keys for session duration).
  - If available, enable CPU crypto extensions or use accelerated AEAD implementations (Go stdlib or assembly optimized libs).
  - Benchmark with larger batch sizes to amortize per-call overhead.

## Datapath (`internal/datapath/afxdp`)
- Hotspots: packet processing loop, buffer/slice operations, metric recording.
- Suggestions:
  - Minimize per-packet allocations: reuse frame buffers, and avoid creating new slices or headers per packet.
  - Batch metric updates or use lock-free counters to reduce contention/overhead in hot loops.
  - Inline critical hot-path functions and reduce interface indirection in the loop body.
  - Consider sampling histograms or using aggregated per-worker histograms to reduce per-packet metric cost.
  - For AF_XDP path, prefer zero-copy UMEM and descriptor reuse (reuse descriptors instead of reallocating frames).

## Next diagnostic steps
- Generate flamegraphs (SVG) via `go tool pprof -http` or `-web` on a machine with `dot` available to visually inspect stack traces.
- Run `go test -c` and `pprof -http` on representative hardware under load to validate hotspots.
- For each suggested change, create microbenchmarks to measure impact.

