# Production Readiness Checklist

This document captures the finalization steps to move SMIP-MWP to production-ready
status for AF_XDP-enabled forwarding.

Prerequisites
- Dedicated bare-metal hosts with 40/100Gbps NICs
- Kernel 5.10+ (6.x preferred)
- Root privileges for XDP attach and hugepage config

Steps
1. CI Green
   - All GitHub Actions workflows pass: `ci.yml`, `benchmarks.yml`.
2. Host Prep
   - Install required packages: `libbpf-dev clang llvm libelf-dev ethtool iproute2`.
   - Reserve hugepages: `vm.nr_hugepages=4096` (example).
   - Set NIC queue count and pin IRQs to worker CPUs.
3. Build
   - `go test ./... -tags=withafxdp -run '^$'` (compile check)
   - `./scripts/release.sh --out-dir=dist` to build release artifacts.
4. Deploy as systemd
   - Install `deploy/mohawk-node.service` and enable it.
5. Performance Validation
   - Run `./scripts/max_throughput_run.sh` with `--auto-pin` and MoonGen/Trex to validate sustained throughput (30–120s runs).
   - Capture `benchmarks/*-cpu.prof`, `/proc/interrupts`, `dmesg` logs.
6. Observability & Soak
   - Verify Prometheus metrics endpoint and run a 24-hour soak with real traffic.
7. Security & Hardening
   - Run security audit, enable seccomp/AppArmor as appropriate, verify handshake properties.
8. Release
   - Tag release, publish artifacts, and update deployment runbooks.

Rollback plan
- If attach or runtime failures occur, stop the service and collect logs: `journalctl -u mohawk-node -b`.

Acceptance criteria
- CI green
- Benchmarks show >=80% line-rate for target NICs (validated artifacts)
- No critical errors in 24-hour soak
- Handshake and routing invariants verified
