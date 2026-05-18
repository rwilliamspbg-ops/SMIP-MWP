# MoonGen / TREx Benchmark Plan

Objective
- Validate real hardware throughput and latency for AF_XDP forwarder.
- Reproduce sustained 1–24h runs, p99/p999 latencies, and zero-drop operation at target rates.

Test Matrix
- NICs: ixgbe (Intel 82599), i40e, mlx5 (Mellanox/ConnectX)
- Links: 25G, 40G, 100G where available
- Drivers: stock kernel drivers for each NIC family
- Workloads:
  - Uniform 1500B packets (TCP/UDP/RAW)
  - Small-packet mix (64B/128B/256B)
  - Mixed realistic flows (flows with stateful sessions)
  - Bursty traffic (on/off patterns with 10ms–1s bursts)
- Duration: short runs (5–10min) + long soak (1–24h)

Metrics to collect
- Throughput (pps and Gbps)
- Packet loss (per-second and total)
- Latency: median, p95, p99, p999 (use MoonGen latency test or hardware timestamping)
- CPU utilization per core
- Memory usage and GC if applicable
- pprof CPU + heap profiles at baseline and under load
- Prometheus metrics exposed by the forwarder

Test Steps (MoonGen)
1. Prepare DUT: disable offloads that interfere with AF_XDP, pin IRQs, set CPU isolations.
2. Build the forwarder with `-tags=withafxdp` and deploy to DUT.
3. Configure MoonGen script for chosen NIC pair and traffic pattern.
4. Run ramp tests: 1G -> 3G -> 5G -> 10G -> target.
5. Record metrics and pprof snapshots at each step.
6. For soak runs, run continuous traffic at target for 1–24h and verify zero-drop.

Commands (example MoonGen run)
```bash
# On traffic generator host (MoonGen):
sudo ./build/MoonGen ./examples/l3-load-latency.lua --dev 0 --dev 1 --rate 10000

# Build forwarder on DUT:
go build -tags=withafxdp -o bin/mohawk-node ./cmd/mohawk-node
# Run on DUT pinned to CPUs with systemd/unit (example):
./bin/mohawk-node --config config.yaml
```

TRex notes
- Use TRex for high-rate packet streams when MoonGen scripting is insufficient for the desired patterns.
- Use TRex stateless for raw throughput and stateful for flow mixes.

Reporting
- Store raw outputs and summarized CSVs per run.
- Attach pprof CPU/profile outputs and relevant traces.
- Produce an executive summary with: max sustainable throughput (no-drop), p99 latency at that throughput, and CPU/core efficiency.

Failure modes & remediation
- Drops at high rates: check RX/TX ring sizes, NUMA affinity, XDP program attachment, and driver IRQ balance.
- High latency: examine batch sizes, CPU steal, interrupt coalescing, and packet steering.

Requirements for reproducibility
- Kernel version and distro
- NIC firmware and driver versions
- MoonGen/TRex commit hashes
- DUT CPU model and core/NUMA layout
- Exact command lines and config files


