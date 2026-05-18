# SMIP-MWP

[![CI](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/ci.yml/badge.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/ci.yml)
[![Benchmarks](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/benchmarks.yml/badge.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/benchmarks.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/rwilliamspbg-ops/SMIP-MWP)](https://goreportcard.com/report/github.com/rwilliamspbg-ops/SMIP-MWP)
[![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

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

This repository includes an iterative, profile-driven optimization workflow. Representative artifacts and full pprof captures live under [benchmarks/](benchmarks/). Key measured results from the most recent development runs (local CI / bench harness) follow — use these as a canonical baseline for upcoming hardware runs.

- Final canonical synthetic benchmark (30s, dev host): **1611 ns/op** (~620k packets/sec), `benchmarks/final-canonical-cpu.prof`
- Best local measured result during tuning: **1487 ns/op** (~672k pps) (config: `CRYPTO_WORKERS=1`, `CRYPTO_BATCH_SIZE=4`, pre-warmed sessions)
- Memory: ~312 B/op, 6 allocs/op (hot-path stable)

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

This project is MIT licensed. See [LICENSE](LICENSE).
