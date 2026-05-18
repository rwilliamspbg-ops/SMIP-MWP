# AF_XDP Phase 2 Optimization Guide: Achieving 3-5 Gbps

**Target Date**: June 10, 2026  
**Goal**: Validate 3-5 Gbps forwarding with <500ns per-packet latency on real hardware

---

## Executive Summary

Phase 1 is complete ✅. Phase 2 requires three parallel workstreams to achieve the 3-5 Gbps target:

1. **Hardware Validation Setup** (May 28-Jun 3)
   - Prepare test environment (2 hosts or traffic generator)
   - Run baseline benchmarks
   - Measure current throughput and identify bottlenecks

2. **Datapath Optimization** (Jun 4-7)
   - Profile with pprof to find hotspots
   - Implement multi-queue worker spawning
   - Tune UMEM allocation and descriptor reuse

3. **Performance Tuning** (Jun 8-10)
   - GC profiling and optimization
   - Cache locality improvements
   - Final validation against acceptance criteria

---

## Workstream 1: Hardware Validation Setup (May 28-Jun 3)

### 1.1 Environment Preparation

**Checklist**:
```bash
# 1. Kernel version (5.10+ required; 6.x preferred)
uname -r                                    # Target: 6.1+

# 2. Install AF_XDP dependencies
sudo apt-get update
sudo apt-get install -y libbpf-dev clang llvm libelf-dev ethtool iproute2 \
  linux-tools-generic bpftool

# 3. Set up hugepages (4096 x 2MB = 8GB)
echo 4096 | sudo tee /proc/sys/vm/nr_hugepages
grep Huge /proc/meminfo                     # Verify: 4096 HugePages_Free

# 4. NIC capability check
ethtool -i eth0 | grep driver               # Must be ixgbe, i40e, mlx5, etc.
ethtool -k eth0                             # Check offload capabilities
ethtool -l eth0                             # Note RX/TX queue counts

# 5. CPU affinity preparation
nproc                                       # Note CPU count (for worker tuning)
```

### 1.2 Baseline Benchmarks (Local/CI)

```bash
# Crypto benchmarks (10-30 sec)
cd /workspaces/SMIP-MWP
go test -bench=BenchmarkEncryptInPlace -benchmem ./internal/crypto/... -count=5

# Expected: ~2.5 µs/op (or faster with hardware acceleration)
```

**Record baseline**:
- [ ] EncryptInPlace latency (ns/op)
- [ ] Session creation latency (ns/op)
- [ ] Memory allocations (B/op)

### 1.3 Traffic Generator Setup

**Option A**: Use existing tools
```bash
# Install pktgen-dpdk or similar
sudo apt-get install -y pktgen

# Or use netcat for simple throughput testing
```

**Option B**: Create test script
```bash
#!/bin/bash
# Generate sustained UDP traffic at target rate
iperf3 -c <forwarder-ip> -u -b 1G -t 60 -J > results.json
```

### 1.4 Monitoring Setup

**Option A**: Prometheus metrics
```bash
# Start forwarder with metrics enabled
sudo ./mohawk-node --iface eth0 --dry-run=false --metrics-addr=:9090

# Monitor in real-time
watch -n 1 'curl -s localhost:9090/metrics | grep -E "rx|tx|latency"'
```

**Option B**: Simple packet counting
```bash
# In another terminal
watch -n 1 'ethtool -S eth0 | grep -E "rx|tx"'
```

---

## Workstream 2: Datapath Optimization (Jun 4-7)

### 2.1 Profile Current Implementation

**Run with pprof**:
```bash
# Build forwarder
go build -tags=withafxdp ./cmd/mohawk-node

# Run with profiling enabled (requires pprof port)
# Extend main.go to include pprof:
import _ "net/http/pprof"

# Profile CPU for 30s under load
go tool pprof -http=:8080 \
  http://localhost:6060/debug/pprof/profile?seconds=30

# Analyze results:
# - Top 5 functions taking most CPU time
# - Look for locks (runtime.selectgo, runtime.unlock2)
# - Identify allocation hotspots
```

**Expected Profile Output**:
```
Flat    Cum     Function
15%     15%     runtime.selectgo           (hot synchronization)
12%     27%     runtime.unlock2
8%      35%     runtime.nanotime
6%      41%     afxdp.(*Forwarder).RunXDPBatchLoop
```

### 2.2 Multi-Queue Worker Integration

**Files to Modify**: `internal/datapath/afxdp/forwarder_xdp.go`

**Current State**:
```go
func (f *Forwarder) Start(ctx context.Context) {
    num := f.cfg.NumWorkers
    if num <= 0 {
        num = runtime.NumCPU()
    }
    // Spawns workers; each worker needs dedicated socket + UMEM
}
```

**Required Enhancements**:
1. Each worker needs separate UMEM and XDPSocket allocation
2. Distributed across queues (worker 0→queue 0, worker 1→queue 1, etc.)
3. Per-worker metrics aggregation

**Code Pattern**:
```go
func (f *Forwarder) Start(ctx context.Context) {
    num := f.cfg.NumWorkers
    if num <= 0 {
        num = runtime.NumCPU()
    }
    
    workerCtx, cancel := context.WithCancel(ctx)
    f.workersCancel = cancel
    
    SpawnPerCPUWorkers(workerCtx, num, &f.workersWG, func(wctx context.Context, id int) {
        // Each worker:
        // 1. Allocates UMEM
        umem, _ := NewUMEM(f.cfg.NumFrames, f.cfg.FrameSize)
        
        // 2. Creates socket for assigned queue
        queueID := id % NumQueues  // Distribute across available queues
        sock, _ := NewXDPSocket(f.cfg.Interface, queueID, umem)
        
        // 3. Runs batch loop
        f.RunXDPBatchLoop(wctx, sock, umem, id)
    })
}
```

**Acceptance Criteria**:
- [ ] N workers spawn (N = NumCPU)
- [ ] Each worker locked to own CPU core
- [ ] Each has dedicated UMEM + socket
- [ ] Per-worker metrics reported independently
- [ ] Linear throughput scaling: 2x workers ≈ 2x throughput

### 2.3 UMEM & Descriptor Optimization

**Already Implemented**:
- ✅ UMEM frame allocation (async/xdp)
- ✅ Descriptor batching in RunXDPBatchLoop
- ✅ In-place packet processing
- ✅ Packet pool (pktPool) for fallback paths

**Tuning Parameters** (in order of impact):

| Parameter | Current | Tuning Range | Impact |
|-----------|---------|--------------|--------|
| NumFrames | 4096 | 2048-16384 | UMEM memory; larger = more buffering |
| FrameSize | 2048 | 1024-4096 | MTU + headroom; tune for typical packet size |
| BatchSize | 64 | 32-256 | RX/TX batch; larger = better amortization |
| NumWorkers | runtime.NumCPU | 1-N | Linear scaling up to NIC queue count |

**Tuning Guide**:
```bash
# Conservative (1-2 Gbps baseline)
./mohawk-node --iface eth0 --frames=4096 --frame-size=2048 --batch-size=64 --workers=4

# Aggressive (3-5 Gbps target)
./mohawk-node --iface eth0 --frames=8192 --frame-size=2048 --batch-size=128 --workers=8

# Maximum (10 Gbps attempt)
./mohawk-node --iface eth0 --frames=16384 --frame-size=2048 --batch-size=256 --workers=16
```

---

## Workstream 3: Performance Tuning (Jun 8-10)

### 3.1 GC Optimization

**Current State**:
- ✅ HKDF caching implemented (unbounded)
- ✅ Packet pooling active
- Need: Bounded LRU cache to prevent unbounded memory growth

**Future Improvement** (if needed):
```go
// In internal/crypto/hybrid.go
const MaxCachedSessions = 10000

type lruCache struct {
    mu    sync.Mutex
    cache map[[32]byte]hkdfCacheEntry
    order []hkdfCacheEntry  // LRU tracking
}

func (c *lruCache) Put(key [32]byte, val hkdfCacheEntry) {
    // Add with LRU eviction if full
}
```

**Validation**:
```bash
# Profile GC behavior under load
go test -bench=BenchmarkEncryptInPlace -cpuprofile=cpu.prof \
  -memprofile=mem.prof ./internal/crypto/...

# Analyze:
go tool pprof cpu.prof   # Look for GC time percentage
go tool pprof mem.prof   # Check allocation patterns
```

### 3.2 Cache Locality Improvements

**Review with pprof**:
```bash
# Run benchmark suite with profiling
./scripts/bench.sh --pprof

# Analyze memory layout:
go tool pprof -http=:8080 benchmarks/bench-*-mem.prof

# Look for:
# - Large allocations in hot path
# - Frequent small allocations (coalesce these)
# - Cache line conflicts
```

**Common Optimizations**:
- Pre-allocate buffers at startup (not per-packet)
- Align hot data structures to cache lines (64 bytes)
- Reduce pointer chasing in packet processing loop

### 3.3 Final Validation Testing

**Test Sequence** (Jun 8-10):

```bash
# Day 1: Functionality validation
go test -tags=withafxdp ./... -v

# Day 2: Performance baseline at increasing load
for rate in 500M 1G 2G 3G 5G 10G; do
  iperf3 -c <target> -u -b $rate -t 30 > results_${rate}.txt
  sleep 10
done

# Day 3: Production readiness
# - 24-hour burn test (optional)
# - Latency distribution analysis
# - Document all findings
```

---

## Key Metrics to Track

### During Optimization

| Metric | Tool | Target | Frequency |
|--------|------|--------|-----------|
| Throughput (pps) | ethtool -S | 1G+ pps | Per test |
| CPU utilization | top/htop | <80% per core | Real-time |
| Latency (p99) | tcpdump + analysis | <1 µs | Per test |
| GC pause time | pprof | <100 µs | Memory profile |
| Memory usage | RSS | <1 GB | Real-time |

### Success Thresholds

```
Week 3 (Jun 3):  1-3 Gbps    @ <80% CPU
Week 4 (Jun 10): 3-5 Gbps    @ <75% CPU
Phase 3 (Jul 8): 10 Gbps     @ <75% CPU (sustained 24h)
```

---

## Common Bottlenecks & Solutions

### Bottleneck 1: Lock Contention

**Symptom**: High `runtime.selectgo` in pprof

**Solution**:
- Pre-lock RWMutex for entire batch (already done in RunXDPLoop)
- Use atomic operations instead of locks where possible
- Consider lock-free session table (advanced)

### Bottleneck 2: Memory Allocations

**Symptom**: High allocation rate in pprof

**Solution**:
- Increase pktPool size
- Pre-allocate buffers at startup
- Use object pooling for temporary structures

### Bottleneck 3: Single-Core Saturation

**Symptom**: Throughput plateaus at 1 Gbps with multi-queue not scaling

**Solution**:
- Verify per-worker isolation (separate UMEM + socket)
- Check CPU affinity is working (Linux: taskset or numactl)
- Profile single worker in isolation

### Bottleneck 4: NIC Driver Overhead

**Symptom**: Flat response after 5-8 Gbps

**Solution**:
- Implement eBPF steering (redirect to queues in kernel)
- Use driver-specific optimizations (mlx5 has accelerated datapath)
- Consider newer NIC driver version

---

## Fallback Plan (If 3-5 Gbps Not Achievable)

If hardware validation shows <3 Gbps:

1. **Check prerequisites**: Verify hugepages, driver version, CPU scaling
2. **Profile bottlenecks**: Use pprof to identify top time consumers
3. **Optimize identified hotspot**: Apply targeted fix
4. **Measure impact**: Re-benchmark after each change
5. **Document findings**: Record in `benchmarks/OPTIMIZATION_LOG.md`

**Acceptable Alternative Targets**:
- Phase 2 (Jun 10): 1-2 Gbps (proven, stable)
- Phase 3 (Aug 5): 3-5 Gbps (with eBPF + additional tuning)

---

## Deliverables for Phase 2 Completion

### Week 3 (Jun 3) Deliverables
- [ ] Hardware validation environment set up
- [ ] Baseline benchmarks collected and documented
- [ ] Multi-queue worker integration complete
- [ ] 1-3 Gbps throughput validated
- [ ] Per-worker metrics functional

### Week 4 (Jun 10) Deliverables
- [ ] 3-5 Gbps sustained throughput achieved
- [ ] <500 ns per-packet latency (p99) validated
- [ ] GC pause <100 µs (p99) confirmed
- [ ] Linear queue scaling verified (2x cores ≈ 2x throughput)
- [ ] Comprehensive benchmarking report (`benchmarks/VALIDATION_REPORT.md`)
- [ ] All Phase 2 acceptance criteria met ✅

---

## Quick Reference: Build & Test Commands

```bash
# Build without AF_XDP (CI/development)
go build ./cmd/mohawk-node
go test ./... -short

# Build with AF_XDP (hardware)
go build -tags=withafxdp ./cmd/mohawk-node
go test -tags=withafxdp ./... -short

# Run forwarder
sudo ./mohawk-node --iface eth0 --dry-run=false --workers=4

# Benchmark
go test -bench=. -benchmem ./internal/crypto/...
./scripts/bench.sh --pprof

# Profile running forwarder (in another terminal)
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

---

## Phase Completion Checklist

- [x] Phase 1: ML-KEM integration
- [x] Phase 1: Handshake state machine  
- [x] Phase 1: main.go forwarder startup
- [ ] Phase 2: Multi-queue worker integration
- [ ] Phase 2: Hardware baseline validation (1-3 Gbps)
- [ ] Phase 2: Optimization & profiling complete
- [ ] Phase 2: 3-5 Gbps target achieved
- [ ] Phase 2: Latency & scaling validated

---

**Next Step**: Run `./scripts/bench.sh --pprof` and analyze results to identify optimization priorities.

