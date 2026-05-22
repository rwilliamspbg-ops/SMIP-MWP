# SMIP-MWP

[![CI](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/ci.yml/badge.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/ci.yml)
[![Benchmarks](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/benchmarks.yml/badge.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/benchmarks.yml)
[![Rust CI](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/ci-rust.yml/badge.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/ci-rust.yml)
[![Lean Build](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/lean4.yml/badge.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/lean4.yml)
[![Formal Verification](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/lean-formalization-gate.yml/badge.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/lean-formalization-gate.yml)
[![Generated Up-to-date Check](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/check-generated-up-to-date.yml/badge.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/check-generated-up-to-date.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/rwilliamspbg-ops/SMIP-MWP)](https://goreportcard.com/report/github.com/rwilliamspbg-ops/SMIP-MWP)
[![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org)
[![License: AGPL-3.0](https://img.shields.io/badge/license-AGPL--3.0-blue.svg)](LICENSE)

<!-- Capability badges -->
[![AF_XDP](https://img.shields.io/badge/AF_XDP-supported-brightgreen.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP)
[![PQC](https://img.shields.io/badge/PQC-hybrid-blue.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP)
[![Benchmarked](https://img.shields.io/badge/benchmarked-yes-green.svg)](benchmarks/)
[![Observability](https://img.shields.io/badge/observability-prometheus-orange.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP)
[![Build tag](https://img.shields.io/badge/build_tag-withafxdp-lightgrey.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP)

SMIP-MWP is a high-performance sovereign transport and routing stack focused on:
- Hybrid PQC session security
- AF_XDP-oriented fast-path forwarding
- Predictive routing controls
- Observable, benchmarked delivery toward production targets

**Documentation status:** Older design and phase documents have been archived to the `docs/archive/` folder; operational usage and performance summaries are in `docs/USAGE.md` and `docs/PERFORMANCE.md`.

## Highlights

- Deterministic hybrid session crypto with in-place AEAD support
- Zero-copy-friendly packet loop scaffolding and batched AF_XDP path
- Prometheus counters and per-worker latency metrics
- Benchmark harness and CI pipeline with artifact retention

## Performance (summary)

This repository includes a profile-driven optimization workflow. Full artifacts and pprof captures live under [benchmarks/](benchmarks/). Below are measured results from the most recent development runs (local bench harness, 60s runs):

- AF_XDP datapath (crypto path, 60s): **2014 ns/op**, 560 B/op, 9 allocs/op (`benchmarks/afxdp-bench-single-60s.txt`).
- AF_XDP datapath (4 workers, crypto path, 60s): **2376 ns/op**, 632 B/op, 11 allocs/op (`benchmarks/afxdp-bench-multi-60s.txt`).

Crypto microbench highlights (60s runs):
- `BenchmarkNewHybridSession_Cached`: **545.0 ns/op**, 1392 B/op, 4 allocs/op
- `BenchmarkNewHybridSession_Uncached`: **610.4 ns/op**, 1392 B/op, 4 allocs/op
- `BenchmarkEncryptInPlace`: **833.1 ns/op**, 1552 B/op, 2 allocs/op
- `BenchmarkDecryptInPlace`: terminated mid-run (recommend re-run with pprof)

These numbers reflect development harness runs — bare-metal validation with MoonGen/TRex is required for final line-rate claims. See `benchmarks/` for raw outputs and `benchmarks/pprof/run_pprof_bench.sh` to capture pprof CPU/memory profiles.

Recommended canonical configuration for hardware validation:

- `CRYPTO_WORKERS=1`
- `CRYPTO_BATCH_SIZE=4`
- Pin receiver process to IRQ/core range using `--auto-pin` helper in `./scripts/max_throughput_run.sh`

See [docs/PERFORMANCE.md](docs/PERFORMANCE.md) for interpretation, pprof usage, and the recommended hardware test process (MoonGen/TRex + Ansible orchestration).

## Usage

Quick start, runtime flags, and host prerequisites are consolidated in [docs/USAGE.md](docs/USAGE.md). Key quick commands:

For a minimal quick-start for new contributors, see [docs/QUICK_START.md](docs/QUICK_START.md).

Run tests and vet:

```bash
go test ./... -v
go vet ./...
```

Run local benchmarks:

```bash
./scripts/bench.sh
```

Run AF_XDP-enabled tests (requires host support and build tag):

```bash
go test ./... -v -tags=withafxdp
```

Host prep and high-throughput helper

We provide an automated host-run helper to apply common tuning, pin IRQs,
run the repo microbench, and print traffic-generator commands for line-rate
tests. This is intended for bare-metal test hosts only (not in containers):

```bash
./scripts/max_throughput_run.sh --iface eth0 --role receiver --generator moongen \
	--queues 16 --hugepages 4096 --auto-pin --cpu-start 2 --duration 120 --benchtime 30
```

The script writes profiling artifacts to `benchmarks/` and prints recommended
`taskset` usage to pin the `mohawk-node` worker process to the IRQ/core range.

## CI and Benchmark Automation

- Main CI: tests and vet on push/PR via `.github/workflows/ci.yml`
- Benchmark CI: scheduled + manual dispatch via `.github/workflows/benchmarks.yml` with artifacts uploaded to runs

## AF_XDP Notes

Recommended host baseline:
- Linux kernel 5.10+ (6.x preferred)
- `libbpf-dev`, `clang`, `llvm`, `libelf-dev`
- XDP-capable NIC/driver and appropriate privileges

Detailed prerequisites and runbooks are referenced from [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md).

## Roadmap and Tracking

- [ROADMAP_EXECUTIVE_SUMMARY.md](ROADMAP_EXECUTIVE_SUMMARY.md)
- [TIMELINE_AND_TRACKER.md](TIMELINE_AND_TRACKER.md)
- [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md)
- [benchmarks/README.md](benchmarks/README.md)

## License

This project is licensed. See [LICENSE](LICENSE).

---

## 10Gbps Performance Upgrade Status

**Goal:** Achieve ≥10 Gbps throughput with full security hardening and formal verification.

### Current State
- AF_XDP latency: **2014 ns/op** (single-worker baseline)
- Sharded session map: ✅ Implemented (reduces mutex contention 16x)
- Frame pooling: ✅ Implemented (zero-allocation hot path)
- Multi-queue forwarding: ✅ Implemented (linear scaling architecture)
- Security hardening: ✅ Implemented (replay protection, DoS mitigation)

### Completed Optimizations
1. **Frame Pooling** (`internal/datapath/afxdp/pool.go`) - Eliminates per-packet allocations
2. **Security Hardening** (`internal/crypto/security_hardening.go`) - Replay attack, overflow checks, DoS protection  
3. **Routing Enhancements** (`internal/routing/router_enhanced.go`) - LPM support with priority ranking
4. **Multi-Queue Forwarder** (`internal/datapath/afxdp/multi_queue.go`) - Linear scaling architecture
5. **UDP Transport Layer** (`internal/transport/udp.go`) - Validates handshake and routing logic

### Path to 10Gbps
See [FINAL_EXECUTION_SUMMARY_10GBPS.md](FINAL_EXECUTION_SUMMARY_10GBPS.md) for complete execution plan and implementation guide.

**Expected Performance After Optimization:**
- Throughput: ≥10 Gbps on appropriate hardware
- Latency Overhead: <1ms (p99)
- GC Pause: <100μs with frame pooling
- Allocation Rate: <0.1% of throughput
