# SMIP-MWP

[![CI](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/ci.yml/badge.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/ci.yml)
[![Benchmarks](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/benchmarks.yml/badge.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/benchmarks.yml)
[![Lean Build](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/lean4.yml/badge.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/lean4.yml)
[![Formal Verification](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/lean-formalization-gate.yml/badge.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/lean-formalization-gate.yml)
[![Generated Up-to-date Check](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/check-generated-up-to-date.yml/badge.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/check-generated-up-to-date.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/rwilliamspbg-ops/SMIP-MWP)](https://goreportcard.com/report/github.com/rwilliamspbg-ops/SMIP-MWP)
[![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org)
[![License: AGPL-3.0](https://img.shields.io/badge/license-AGPL--3.0-blue.svg)](LICENSE)
[![Sponsor](https://img.shields.io/badge/Sponsor%20Me-GitHub%20Sponsors-ea4aaa)](https://github.com/sponsors/rwilliamspbg-ops)

SMIP-MWP is a formally verified, high-performance transport and routing protocol stack. The repository combines a Go control plane, AF_XDP fast-path plumbing, hybrid cryptographic session setup, and Lean-based verification gates so the implementation, generated artifacts, and formal model stay aligned.

## What this repository is

- A protocol and forwarding stack with a verified core and a production-oriented fast path
- An AF_XDP dataplane for low-copy packet I/O on Linux hosts with the `withafxdp` build tag
- A benchmarked crypto and routing surface with artifacts checked into `benchmarks/`
- A repo that treats generated code, verification artifacts, and implementation changes as one delivery surface

## Architecture at a glance

1. Control plane and protocol orchestration live in `cmd/mohawk-node`.
2. AF_XDP datapath components live under `internal/datapath/afxdp` and are exercised through `withafxdp`.
3. Hybrid session crypto and hardening live under `internal/crypto`.
4. Routing, transport, metrics, and runner helpers are organized under `internal/` and `scripts/`.
5. Formal verification workflows are gated by Lean and GitHub Actions so generated artifacts stay in sync with source.

For the operational docs first, start with [docs/QUICK_START.md](docs/QUICK_START.md), [docs/USAGE.md](docs/USAGE.md), and [docs/PERFORMANCE.md](docs/PERFORMANCE.md).

## Setup Walkthrough

### 1. Install prerequisites

- Go 1.25 or the version declared in [go_version.txt](go_version.txt)
- `clang`, `llvm`, `libbpf-dev`, and `libelf-dev` for AF_XDP builds
- Linux kernel 5.10+ for AF_XDP work, with 6.x preferred

### 2. Clone and build

```bash
go build ./...
```

To build the AF_XDP-enabled binary, include the build tag:

```bash
go build -tags=withafxdp ./cmd/mohawk-node
```

### 3. Run the basic checks

```bash
go test ./... -v
go vet ./...
```

For an AF_XDP host preflight and compile-only validation, use:

```bash
./scripts/test_xdp.sh --run-go-test
```

### 4. Use the quick start

The quickest path through the repo is documented in [docs/QUICK_START.md](docs/QUICK_START.md). It covers stub mode, AF_XDP mode, and the minimum smoke checks for each.

## Testing and Benchmarking

Use the standard benchmark runner for repeatable local measurements and artifact capture:

```bash
./scripts/bench.sh
```

Add profiling when you want CPU and memory data:

```bash
./scripts/bench.sh --pprof
```

For a tighter AF_XDP-oriented host validation flow, use the tuning helper:

```bash
./scripts/max_throughput_run.sh --iface eth0 --role receiver --generator moongen \
  --queues 16 --hugepages 4096 --auto-pin --cpu-start 2 --duration 120 --benchtime 30
```

That script prepares the host, runs preflight checks, applies conservative tuning, captures benchmark artifacts, and prints the traffic-generator commands needed for sender-side testing.

### Performance snapshot

The current measured baseline is documented in [docs/PERFORMANCE.md](docs/PERFORMANCE.md). In short:

- Canonical synthetic benchmark: about **1611 ns/op** in the latest documented run
- Best tuned development sweep: about **1487 ns/op** with `CRYPTO_WORKERS=1` and `CRYPTO_BATCH_SIZE=4`
- AF_XDP loop benchmark in this environment: about **1.9 µs/op** for the crypto-free loop path

These are development numbers, not line-rate claims. For hardware validation, use dedicated hosts and a generator such as MoonGen or TRex, then compare the resulting profiles and throughput with the artifacts in [benchmarks/](benchmarks/).

## Documentation Map

- [docs/USAGE.md](docs/USAGE.md) for runtime flags, prerequisites, and canonical commands
- [docs/PERFORMANCE.md](docs/PERFORMANCE.md) for benchmark interpretation and hardware validation guidance
- [benchmarks/README.md](benchmarks/README.md) for artifact policy and benchmark hygiene
- [docs/runner/README.md](docs/runner/README.md) for self-hosted XDP runner setup
- [docs/archive/ARCHIVED_DOCS.md](docs/archive/ARCHIVED_DOCS.md) for legacy planning material

## Contributing

Contributions are welcome, but this repository expects changes to be treated with the same rigor as the protocol itself. Start with [CONTRIBUTING.md](CONTRIBUTING.md) for branch, PR, and review expectations.

## Sponsorship

Use the Sponsor button above or the GitHub Sponsors link for the repo owner. Sponsorship helps support the work needed to keep the protocol, benchmarks, and formalization aligned.

## License

This project is licensed under AGPL-3.0. See [LICENSE](LICENSE).
