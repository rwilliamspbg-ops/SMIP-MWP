PR #6 — WIP: AF_XDP test tweaks, benchmarks, and perf changes
=============================================================

Branch: resume/wip-stash-1778809093

Summary
-------
This PR continues WIP AF_XDP test improvements and adds benchmark artifacts plus a set
of performance-focused changes and microbenchmarks to evaluate impact.

Key changes
- Implemented a small HKDF-derived key cache in `internal/crypto` to avoid
  repeated HKDF work when `NewHybridSession` is called with identical inputs.
- Added `pktPool` (`sync.Pool`) to `afxdp.Forwarder` and used it in the conservative
  `RunXDPLoop` fallback path to reduce per-packet allocations when constructing
  new packets (allocations moved to pooled buffers when possible).
- Batched per-worker Prometheus metric updates in `internal/datapath/afxdp/metrics.go`:
  atomic per-worker counters with a short-interval flusher (10ms) that aggregates
  into Prometheus counters, reducing overhead in the hot packet loop.
- Added microbenchmarks:
  - `internal/crypto/bench_cache_test.go` (NewHybridSession cached vs uncached)
  - `internal/datapath/afxdp/bench_pool_test.go` (allocation vs `sync.Pool`)

Results and observations
- Microbenchmarks (local Codespaces run) show:
  - `BenchmarkNewHybridSession_Cached` and `Uncached` in the ~600–800 ns/op range; cached
    path similar but slightly faster on average — HKDF is not the dominant cost here;
    AEAD setup and key handling still contribute significantly.
  - `BenchmarkPacketPool` shows ~40–50 ns/op for pooled buffers vs negligible for trivial
    allocation microbenchmark (the allocation benchmark is optimized away on this host);
    real-world benefit will be more visible under load and when allocations previously
    caused GC pressure.
- `pprof` CPU profiles saved to `benchmarks/` and `pprof-top` text reports added.

Actionable recommendations
- For crypto: move HKDF-derived key derivation to session lifecycle (done via cache),
  and consider a bounded LRU cache for long-lived processes. Prefer in-place AEAD and
  reuse AEAD instances when possible.
- For datapath: reuse UMEM frames and minimize per-packet header copies (pktPool helps
  in fallback paths). Reduce per-packet metric overhead (batched counters help).

Files added in branch
- `benchmarks/*` — captured benchmark outputs, profiles, and pprof reports
- `internal/crypto/bench_cache_test.go`
- `internal/datapath/afxdp/bench_pool_test.go`
- Modifications to: `internal/crypto/hybrid.go`, `internal/datapath/afxdp/*` (pktPool, loop changes),
  `internal/datapath/afxdp/metrics.go` (batched counters)

Suggested next steps
1. Run full `withafxdp` benchmarks on target hardware and compare profiles.
2. Replace unbounded HKDF cache with an LRU capped at N sessions.
3. Consider increasing metric flusher interval for production (e.g., 1s) to reduce Prometheus churn.
4. Implement descriptor reuse and zero-copy TX path for AF_XDP forwarder.

If you'd like, I can post this summary as a PR comment on #6 or further refine any section.
