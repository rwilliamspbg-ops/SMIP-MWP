# Performance Summary

This file provides an executive summary of measured performance artifacts and guidance for interpreting benchmark outputs produced by the project's benchmark runner and CI.

Where artifacts live

- Raw and processed benchmark outputs are in `benchmarks/` with timestamps and pprof profiles when enabled. Use those artifacts for reproducible analysis.

Interpreting results

- Throughput: look for tx/rx rates printed by the runner or extracted from pprof traces.
- Latency: per-worker histograms and Prometheus summaries are captured by the benchmark harness where enabled.
- Pprof: use `go tool pprof` to load CPU and memory profiles from `benchmarks/*-cpu.prof` and `*-mem.prof`.

Example pprof usage

```bash
go tool pprof -http=:8080 benchmarks/bench-codespaces-*-cpu.prof
```

Reporting

- For CI-driven benchmark runs, reference the workflow run that generated the artifacts and include the run ID/timestamp when reporting performance.

Line-rate testing guidance

Reaching 10–50 Gbps requires dedicated test hardware and a line-rate generator
such as MoonGen or TRex. Quick checks with `iperf3` are useful for sanity but
are unlikely to saturate 25–50Gbps.

Recommended process:

1. Prepare both sender and receiver hosts using `./scripts/max_throughput_run.sh`.
2. Reserve hugepages (2MB) and set NIC queue counts to match available worker
	cores (use `ethtool -L <iface> combined <n>`).
3. Auto-pin NIC IRQs with `--auto-pin` and pin the `mohawk-node` worker process
	to the same CPU range via `taskset`.
4. Generate traffic from the sender using MoonGen/TRex; tune packet size and
	per-core rates to hit the target Gbps.
5. Collect `benchmarks/*-cpu.prof`, `/proc/interrupts`, `/proc/irq/*/smp_affinity`,
	and `dmesg` for post-run analysis.

Profile analysis

Load CPU profiles with `go tool pprof -http=:8080` and identify hot loops
in the AF_XDP path (look for `RunXDPLoop`, `PrepareForPacket`, and allocation
sites). Target optimizations at:

- reducing per-packet allocations (increase frame reuse / pooling)
- minimizing lock contention in routing/forwarder tables
- aligning queue counts with workers and pinning to avoid cross-core cache
  thrashing

If you provide pprof profiles and machine details (NIC model, CPU, kernel), we
can analyze hotspots and recommend precise code changes to approach 25–50Gbps.
