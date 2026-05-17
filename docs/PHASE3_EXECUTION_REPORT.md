# Phase 3 Execution Report — Hardening & Verification

Date: 2026-05-17
Author: SMIP-MWP automation

Summary
-------

This document consolidates the short-run Phase 3 execution artifacts produced on 2026-05-17, summarizes findings, and outlines recommended next steps toward full Phase 3 acceptance (sustained 10 Gbps validation, formal verification, and security audit).

Artifacts and commands run
--------------------------

- AF_XDP preflight (compile-only):

  ```bash
  ./scripts/test_xdp.sh --run-go-test --output benchmarks/xdp-preflight-local.txt
  ```

- Package-level benchmarks (profiles produced):

  ```bash
  go test ./internal/datapath/afxdp -bench . -benchmem -run '^$' -count=1 -cpuprofile=benchmarks/afxdp-cpu.prof -memprofile=benchmarks/afxdp-mem.prof
  go test ./internal/crypto -bench . -benchmem -run '^$' -count=1 -cpuprofile=benchmarks/crypto-cpu.prof -memprofile=benchmarks/crypto-mem.prof
  ```

- Artifacts produced (workspace paths):

  - `benchmarks/xdp-preflight-local.txt`
  - `benchmarks/afxdp-cpu.prof`
  - `benchmarks/afxdp-mem.prof`
  - `benchmarks/crypto-cpu.prof`
  - `benchmarks/crypto-mem.prof`

Key results (short-run)
-----------------------

- Preflight: kernel 6.8.0 (>=5.10) — PASS. Interface `eth0` detected — PASS. `ethtool` missing — WARN (limits NIC capability detection).

- AF_XDP hot-path (short-run):
  - `BenchmarkPacketAllocate-4`: 1e9 ops, ~0.3344 ns/op — indicates optimized zero-alloc path for packet allocation.
  - `BenchmarkPacketPool-4`: 26.1M ops, ~43.95 ns/op, 24 B/op, 1 alloc/op — pool path is low-cost but still allocs observed.
  - `BenchmarkRunXDPLoop_NoCrypto-4`: 637k ops, ~1,909 ns/op, 544 B/op, 8 allocs/op — per-packet processing baseline in this environment.

- Crypto primitives (short-run):
  - `NewHybridSession_Cached`: ~620 ns/op
  - `NewHybridSession_Uncached`: ~785 ns/op
  - `EncryptInPlace`: ~944 ns/op
  - `DecryptInPlace`: ~713 ns/op

Interpretation
--------------

- The short-run results show performant crypto primitives and low-allocation paths for packet allocation and pooling. The `RunXDPLoop_NoCrypto` timings highlight that per-packet overhead (descriptor handling, ring ops, and remaining allocations) is the primary optimization target for throughput gains.

- The missing `ethtool` on the test host prevented automated detection of NIC driver capabilities and offload features; this should be added to the preflight checklist for future runs.

Gaps vs Phase 3 acceptance
---------------------------

- Phase 3 acceptance criteria include sustained 10 Gbps throughput and p99 GC pause <100µs. These require:
  - Dedicated hardware (XDP-capable NICs) and isolated test environment.
  - Multi-queue forwarder configuration and NUMA-aware CPU affinity.
  - Longer sustained runs (1–24 hours) to detect degradation and GC behavior.

Recommendations / Next actions
-----------------------------

1. Install `ethtool` on test hosts and re-run `./scripts/test_xdp.sh` to collect full NIC/driver info and offload capabilities.
2. Implement and prioritize Phase 2 critical items before full Phase 3 sustained runs:
   - Descriptor reuse (UMEM lifecycle, RX→TX descriptor chain)
   - Multi-queue forwarder and CPU affinity
   - eBPF steering to improve queue distribution
3. Run a staged sustained throughput test (recommendation):
   - Stage A: 30m run, multi-queue enabled, gather pprof and Prometheus metrics
   - Stage B: 2–4h run to observe GC and resource drift
   - Stage C: 24h run for full acceptance validation
4. For each sustained run, collect and archive:
   - CPU and mem profiles (`.prof`), Prometheus metrics scrape, pcap (if needed), and `xdp-preflight` report
5. Produce a formal Phase 3 report including hardware specs, run IDs, measured throughput/latency histograms (p50/p99/p99.9), and any mitigation actions taken.
6. Engage formal verification and security audit teams once sustained-run metrics meet throughput and latency targets.

PR note
-------

This report (and associated docs updates) is intended to be delivered as a PR so reviewers can link the bench artifacts and propose additional test host hardware or configuration changes.

Appendix: quick pprof commands
-----------------------------

```bash
go tool pprof -http=:8080 benchmarks/afxdp-cpu.prof
go tool pprof -http=:8080 benchmarks/crypto-cpu.prof
```
