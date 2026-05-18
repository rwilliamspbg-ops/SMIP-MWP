# Quick Start

This quick-start shows how to build and run the forwarder in both stub and AF_XDP modes for simple verification.

Prerequisites
- Go 1.20+ (or the version in `go_version.txt`)
- For AF_XDP mode: kernel with AF_XDP support, libbpf/clang if compiling XDP programs
- Optional: MoonGen or TRex for traffic generation

Build
```bash
# Stub mode (fast, no kernel deps)
go build -o bin/mohawk-stub ./cmd/mohawk-node

# AF_XDP mode (requires withafxdp tag)
go build -tags=withafxdp -o bin/mohawk-xdp ./cmd/mohawk-node
```

Run (stub)
```bash
./bin/mohawk-stub --config examples/config-stub.yaml
```

Run (AF_XDP)
```bash
sudo ./bin/mohawk-xdp --config examples/config-xdp.yaml
```

Smoke test
- Use `nc` or a simple packet generator to send/receive traffic and verify logs.
- For AF_XDP, verify the interface is bound and program attached (e.g., via `ip -d link show dev <iface>`).

Where to look next
- `benchmarks/` for benchmark artifacts and the new `MOONGEN_TRex_BENCHPLAN.md` for hardware validation guidance.
- `OPTIMIZATION_ROADMAP.md` for prioritized optimization items (Tier 1/2/3).

