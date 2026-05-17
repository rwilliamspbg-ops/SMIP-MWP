# Usage

This document summarizes quick start, runtime flags, and host prerequisites for running and evaluating SMIP-MWP.

Quick start

- Run unit tests and static checks:

```bash
go test ./... -v
go vet ./...
```

- Run the standard benchmark runner (captures artifacts and optional pprof):

```bash
./scripts/bench.sh --pprof
```

Build tags

- AF_XDP functionality is gated by the `withafxdp` build tag. Use `-tags=withafxdp` when running tests or building the fast-path components.

Host prerequisites (AF_XDP)

- Linux kernel 5.10+ (6.x preferred)
- `libbpf-dev`, `clang`, `llvm`, `libelf-dev`
- XDP-capable NIC/driver and appropriate privileges

Where to find more

- Performance summaries: [docs/PERFORMANCE.md](docs/PERFORMANCE.md)
- Benchmark artifacts and run data: [benchmarks/](benchmarks/)
