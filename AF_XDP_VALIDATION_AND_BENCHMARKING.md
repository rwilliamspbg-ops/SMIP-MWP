# AF_XDP Hardware Validation & Benchmarking Suite

**Target Date**: June 10, 2026  
**Goal**: Validate 3-5 Gbps forwarding with <500ns per-packet latency

---

## Quick Start: Hardware Validation

### Prerequisites Check

```bash
# 1. Verify kernel version (5.10+)
uname -r

# 2. Install AF_XDP dependencies
sudo apt-get update
sudo apt-get install -y libbpf-dev clang llvm libelf-dev ethtool iproute2

# 3. Set up hugepages (4096 x 2MB = 8GB)
echo 4096 | sudo tee /proc/sys/vm/nr_hugepages
grep Huge /proc/meminfo

# 4. Check NIC driver (must support AF_XDP: ixgbe, i40e, mlx5, etc.)
ethtool -i <interface> | grep driver

# 5. List RX/TX queues
ethtool -l <interface>
```

### Single Interface Validation

```bash
# Run preflight checks
./scripts/test_xdp.sh --iface <interface>

# Build with AF_XDP support
go build -tags=withafxdp ./cmd/mohawk-node

# Start forwarder (dry-run first)
./mohawk-node --iface <interface> --dry-run=true --metrics-addr=:9090

# Run with AF_XDP (requires root)
sudo ./mohawk-node --iface <interface> --dry-run=false \
  --frames=4096 --frame-size=2048 --batch-size=64 --workers=4
```

### Benchmarking

```bash
# Run full benchmark suite with pprof profiles
./scripts/bench.sh --pprof

# Run specific crypto/session benchmarks
go test -tags=withafxdp ./internal/crypto -bench . -benchmem

# Run datapath benchmarks
go test -tags=withafxdp ./internal/datapath/afxdp -bench . -benchmem
```

---

## Validation Stages

### Stage 1: Unit Test Validation (5 min)

```bash
go test -tags=withafxdp ./... -v -short
```

**Success Criteria**:
- ✅ All unit tests pass
- ✅ No panics or segfaults
- ✅ ML-KEM keys verified (non-random)

### Stage 2: Micro-benchmark Validation (10 min)

```bash
go test -tags=withafxdp ./internal/crypto -bench=BenchmarkEncryptInPlace -benchmem -count=5
go test -tags=withafxdp ./internal/datapath/afxdp -bench=Benchmark -benchmem -count=3
```

**Success Criteria**:
- ✅ EncryptInPlace: <1 µs per packet
- ✅ Packet pool: <100 ns pool Get/Put
- ✅ HKDF cache hit: <1 µs

**Expected Output**:
```
BenchmarkEncryptInPlace-8    500000   2500 ns/op   1000 B/op   2 allocs/op
BenchmarkPacketPool-8     10000000    100 ns/op      0 B/op   0 allocs/op
```

### Stage 3: Integration Validation (30 sec per test)

```bash
# Run in stub mode (no AF_XDP required)
go test -tags=!withafxdp ./internal/datapath/afxdp -v -run=IntegrationTest

# With AF_XDP (requires kernel + NIC)
sudo go test -tags=withafxdp ./internal/datapath/afxdp -v -run=IntegrationTest
```

**Success Criteria**:
- ✅ Forwarder starts and runs
- ✅ Graceful shutdown on signal
- ✅ No resource leaks (check /proc/[pid]/fd)

### Stage 4: Hardware Forwarding Validation (5-10 min)

**Setup** (requires 2 hosts or 2 NICs):

```bash
# Host A (forwarder)
sudo ./mohawk-node --iface eth0 --dry-run=false \
  --frames=4096 --frame-size=2048 --batch-size=64 \
  --workers=0 --metrics-addr=:9090

# Host B (traffic generator) — send packets at 1 Gbps
pktgen -i eth1 -d <host-a-ip> -s 1500 -r 1000000

# Monitor metrics
watch -n 1 'curl -s localhost:9090/metrics | grep -E "rx|tx|latency"'
```

**Success Criteria**:
- ✅ RX packets incrementing at expected rate
- ✅ TX packets ≈ RX packets (no drops)
- ✅ Latency <5ms (p99)
- ✅ CPU utilization <60% per core

### Stage 5: Throughput Ramp Test (20 min)

Gradually increase traffic load and measure throughput scaling:

```bash
for rate in 100 250 500 1000 2000 5000 10000 Mbps; do
  pktgen -i eth1 -d <host-a-ip> -s 1500 -r $rate
  sleep 60
  curl -s localhost:9090/metrics | grep -E "rate|latency" >> results_$rate.txt
done
```

**Expected Output**:
- 100 Mbps: ~500ns/pkt, CPU 5%
- 250 Mbps: ~500ns/pkt, CPU 12%
- 500 Mbps: ~500ns/pkt, CPU 25%
- 1 Gbps: ~500ns/pkt, CPU 50%
- 3 Gbps: ~500ns/pkt, CPU 70% (target for Week 4)
- 5+ Gbps: Requires eBPF steering + multi-queue tuning

---

## Performance Targets & Acceptance Criteria

### Week 3 Targets (May 28-Jun 3)

| Metric | Target | Status |
|--------|--------|--------|
| **Throughput** | 1-3 Gbps | 🔴 Not validated |
| **Per-packet latency** | <1 µs (p99) | 🔴 Not validated |
| **CPU utilization** | <80% per core | 🔴 Not validated |
| **GC pause (p99)** | <100 µs | 🔴 Not validated |
| **Linear queue scaling** | 2x queues = ~2x throughput | 🔴 Not validated |

### Week 4 Targets (Jun 4-10)

| Metric | Target | Status |
|--------|--------|--------|
| **Throughput** | 3-5 Gbps sustained | 🔴 Not validated |
| **Per-packet latency** | <500 ns (p99) | 🔴 Not validated |
| **CPU utilization** | <75% per core | 🔴 Not validated |
| **Queue scaling** | Linear: N queues = N x throughput | 🔴 Not validated |
| **Handshake latency** | <50 ms | 🔴 Not validated |

### Phase 3 Targets (Aug 5)

| Metric | Target | Status |
|--------|--------|--------|
| **Throughput** | 10 Gbps sustained, 24-hour burn | 🔴 Not validated |
| **Per-packet latency** | <500 ns (p99) + <1ms added vs plain | 🔴 Not validated |
| **GC pause (p99)** | <100 µs | 🔴 Not validated |
| **Handshake** | <50 ms (all trials) | 🔴 Not validated |
| **Zero routing flaps** | Formal proof + conformance | 🔴 Not validated |

---

## Instrumentation & Monitoring

### Prometheus Metrics (Active)

**Endpoint**: `http://localhost:9090/metrics`

Key metrics to track:

```
# Packet counters
smip_afxdp_rx_packets_total{worker_id="0"}
smip_afxdp_tx_packets_total{worker_id="0"}
smip_afxdp_dropped_packets_total{worker_id="0"}
smip_afxdp_crypto_errors_total

# Latency histograms (if implemented)
smip_afxdp_processing_latency_seconds
smip_afxdp_handshake_duration_seconds

# Handshake events
smip_afxdp_handshakes_total
```

### Profiling

```bash
# CPU profile (30s)
go tool pprof -http=:8080 http://localhost:9091/debug/pprof/profile?seconds=30

# Memory profile
go tool pprof -http=:8080 http://localhost:9091/debug/pprof/heap

# Goroutine analysis
go tool pprof -http=:8080 http://localhost:9091/debug/pprof/goroutine
```

---

## Optimization Checklist for Week 3

- [ ] Run benchmark suite: `./scripts/bench.sh --pprof`
- [ ] Review CPU profiles (RunXDPBatchLoop hotspots?)
- [ ] Check allocations: `BenchmarkPacketPool` should show 0 allocs/op
- [ ] Verify descriptor reuse: no malloc in packet loop
- [ ] Test multi-queue: spawn 2-4 workers, measure scaling
- [ ] GC tuning: monitor pause times under load
- [ ] eBPF steering: implement BPF classifier for RX direction

---

## Optimization Checklist for Week 4

- [ ] Achieve 3-5 Gbps baseline on target hardware
- [ ] Validate latency <500ns (p99) for single core
- [ ] Confirm linear queue scaling (2x cores ≈ 2x throughput)
- [ ] Profile hotspots with pprof, address top 3
- [ ] GC pause <100µs (p99) under sustained load
- [ ] Document results in `benchmarks/VALIDATION_REPORT.md`

---

## Troubleshooting

### "withafxdp not compiled in"

```bash
# Rebuild with AF_XDP support
go build -tags=withafxdp ./cmd/mohawk-node
```

### "permission denied" when running forwarder

```bash
# AF_XDP operations require root or CAP_NET_ADMIN
sudo ./mohawk-node --iface eth0 --dry-run=false
```

### "interface not found" or "unsupported driver"

```bash
# Check NIC driver and AF_XDP capability
ethtool -i eth0
# Must be one of: ixgbe, i40e, mlx5, etc.

# If not AF_XDP capable, use a different NIC or VM passthrough
```

### Slow handshake (>50ms)

```bash
# Profile handshake latency
go test -bench=BenchmarkHybridHandshake -benchmem ./internal/crypto
# If slow, check HKDF cache hit rate
```

### Packet drops at high throughput

```bash
# Increase UMEM frame count and batch size
./mohawk-node --iface eth0 --frames=8192 --batch-size=128

# Tune CPU affinity if multi-queue isn't spreading load
numactl -C 0-3 ./mohawk-node ...
```

---

## Phase 2 Completion Checklist

- [x] Phase 1: ML-KEM integration
- [x] Phase 1: Real handshake state machine
- [x] main.go updated with forwarder startup
- [ ] AF_XDP descriptor reuse validated (3-5 Gbps baseline)
- [ ] Multi-queue worker spawning working
- [ ] eBPF steering program implemented (optional for 3-5 Gbps)
- [ ] Benchmarking suite run and documented
- [ ] <500ns per-packet latency measured
- [ ] Linear queue scaling verified
- [ ] All validation tests passing

---

## Next Steps

1. **Immediate** (May 17-20): Run benchmark suite on dev machine
2. **Week 3** (May 28-Jun 3): Deploy on target hardware, achieve 3-5 Gbps
3. **Week 4** (Jun 4-10): Optimize for 10 Gbps target, profile hotspots
4. **Phase 3** (Jun 11+): Formal verification, security audit

