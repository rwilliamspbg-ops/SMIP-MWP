# SMIP-MWP

[![CI](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/ci.yml/badge.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/ci.yml)
[![Benchmarks](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/benchmarks.yml/badge.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/benchmarks.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/rwilliamspbg-ops/SMIP-MWP)](https://goreportcard.com/report/github.com/rwilliamspbg-ops/SMIP-MWP)
[![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

SMIP-MWP is a high-performance sovereign transport and routing stack focused on:
- Hybrid PQC session security
- AF_XDP-oriented fast-path forwarding
- Predictive routing controls
- Observable, benchmarked delivery toward production targets

## Current Delivery Snapshot

Completed in this repository:
- Deterministic hybrid session crypto with in-place AEAD support
- Zero-copy-friendly packet loop scaffolding and batched AF_XDP path
- Prometheus counters and per-worker latency metrics
- Benchmark harness and benchmark CI workflow with artifacts

In progress:
- Full AF_XDP hardware validation path (withafxdp on production-like hosts)
- AF_XDP prerequisites and runtime runbook hardening

## Quick Start

Run tests and vet:

```bash
go test ./... -v
go vet ./...
```

Run local benchmarks:

```bash
./scripts/bench.sh
```

Run benchmarks and collect pprof artifacts:

```bash
./scripts/bench.sh --pprof
```

Run a selected benchmark command through the standardized runner:

```bash
./scripts/bench.sh -- go test ./internal/crypto -bench . -benchmem -run ^$ -count=1
```

## CI and Benchmark Automation

- Main CI: tests and vet on push/PR via `.github/workflows/ci.yml`
- Benchmark CI: scheduled weekly + manual dispatch via `.github/workflows/benchmarks.yml`
- Benchmark artifacts are uploaded per OS runner (`ubuntu-latest`, `macos-latest`) and retained for 14 days

## AF_XDP Notes

AF_XDP integration is build-tagged behind `withafxdp`:

```bash
go test ./... -v -tags=withafxdp
```

Recommended host baseline:
- Linux kernel 5.10+ (6.x preferred)
- `libbpf-dev`, `clang`, `llvm`, `libelf-dev`
- XDP-capable NIC/driver and appropriate privileges

Detailed prerequisites and operator runbook:
- [IMPLEMENTATION_PLAN.md#af_xdp-prerequisites-runbook](IMPLEMENTATION_PLAN.md#af_xdp-prerequisites-runbook)

## Roadmap and Tracking

- [ROADMAP_EXECUTIVE_SUMMARY.md](ROADMAP_EXECUTIVE_SUMMARY.md)
- [TIMELINE_AND_TRACKER.md](TIMELINE_AND_TRACKER.md)
- [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md)
- [benchmarks/README.md](benchmarks/README.md)

## License

This project is MIT licensed. See [LICENSE](LICENSE).
