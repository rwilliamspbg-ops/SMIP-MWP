Hardware Validation Runbook

Objective
- Validate line-rate throughput and latency on bare-metal with AF_XDP-enabled NICs.
- Produce sustained (30–120s, and 1–24h soak) zero-drop runs and p99/p999 latency measurements.

Prerequisites (DUT)
- Kernel: 5.10+ (6.x preferred)
- libbpf-dev, clang, llvm, libelf-dev
- Hugepages configured (e.g., 4096)
- NICs: ixgbe, i40e, mlx5 recommended
- Traffic generator host with MoonGen or TRex and appropriate NICs

Host prep (DUT)
- Disable CPU frequency scaling and turbo
- Isolate CPUs (use `isolcpus` or systemd CPUSet)
- Configure IRQ affinity for NIC queues
- Reserve hugepages and mount hugetlbfs

Example host tuning snippet (as root)

```bash
# reserve hugepages
echo 4096 > /proc/sys/vm/nr_hugepages
# pin IRQs to CPUs - use provided helper script
./scripts/max_throughput_run.sh --iface eth0 --role receiver --generator moongen \
  --queues 16 --hugepages 4096 --auto-pin --cpu-start 2 --duration 120 --benchtime 30
```

Run the forwarder (AF_XDP)

```bash
# build with AF_XDP support
go build -tags=withafxdp -o bin/mohawk-node ./cmd/mohawk-node
# run pinned to CPU range (example using taskset)
sudo taskset -c 2-9 ./bin/mohawk-node --config /etc/mohawk/config-xdp.yaml
```

MoonGen quick-run (traffic generator host)

```bash
# Example: send 10 Gbps using example script
sudo ./build/MoonGen benchmarks/moongen/l3_load_latency.lua --dev0 0 --dev1 1 --rate 10000 --duration 600
```

TRex quick-run

```bash
# Start server/daemon on generator host and run stateless profile
sudo ./t-rex-64 -i
sudo ./trex-console -f my_profile.yaml -m 10000
```

Profiling and data collection
- Collect `pprof` CPU and memory profiles from DUT during runs (use `./scripts/bench.sh --pprof` or `go test` harness)
- Collect MoonGen latency outputs, TRex stats, and DUT `sar`/`top`/`perf` samples
- Save raw pcap if needed for packet-level debugging

Soak testing
- For soak runs (1–24h), run traffic at target sustainable rate (no-drop) and monitor p99/p999 latency and drops

QUIC comparison plan (quic-go)
- Run a `quic-go` server on DUT and a client on generator host or separate host
- Use `wrk` or `wrk2` with a quic plugin, or custom client using `quic-go` load generator
- Measure throughput and latency under same CPU pinning and NIC configs

Example quic-go run
```bash
# server
./quic-server --addr :4433 --cert cert.pem --key key.pem
# client (simple loader)
./quic-client --addr <DUT>:4433 --concurrency 100 --requests 100000
```

Notes
- Record kernel/NIC driver versions, NIC firmware, and any ethtool settings
- Repeat runs with different batch sizes and `BatchSizeMin`/`BatchSizeMax` to find sweet spot

