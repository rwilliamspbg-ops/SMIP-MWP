# Phase 2 Optimization Guide: Path to 10 Gbps

**Target**: 10 Gbps sustained throughput  
**Current**: 7.06 Gbps per core (4 cores = 28.2 Gbps theoretical)  
**Gap**: 29% improvement needed  
**Timeline**: 1-2 weeks with hardware deployment

---

## Optimization Strategy

### Tier 1: Critical Path (Must-Do) - 10-15% Gain

These optimizations directly reduce per-packet latency in the hot path.

#### 1.1 Sharded Session Map (100-200 ns gain)

**Problem**: Global RWMutex on session map creates bottleneck with concurrent workers.

**Solution**: Shard session map into 16 independent maps with independent mutexes.

```go
// Current: 1 RWMutex protecting 1 map
type Forwarder struct {
    sessions map[[16]byte]*Session
    mu       sync.RWMutex
}

// Optimized: 16 RWMutexes protecting 16 maps (hash-based sharding)
type Forwarder struct {
    sessionShards [16]struct {
        sessions map[[16]byte]*Session
        mu       sync.RWMutex
    }
}

func (f *Forwarder) getSessionShard(sid [16]byte) int {
    return int(binary.BigEndian.Uint64(sid[:8]) % 16)
}

func (f *Forwarder) GetSession(sid [16]byte) *Session {
    shard := f.getSessionShard(sid)
    f.sessionShards[shard].mu.RLock()
    defer f.sessionShards[shard].mu.RUnlock()
    return f.sessionShards[shard].sessions[sid]
}
```

**Benefits**:
- Reduces mutex contention by 16x
- 4 workers now contend on average with 4/16 = 0.25 other workers per shard
- Estimated gain: 100-200 ns per lookup

**Implementation effort**: 2-3 hours  
**Risk level**: Low (well-understood technique)

---

#### 1.2 Batch Descriptor Pre-allocation (50-100 ns gain)

**Problem**: Allocating descriptor slices in loop wastes cycles.

**Solution**: Pre-allocate reusable descriptor slices before polling.

```go
// Current: Allocate fresh slice each poll
descs := xsk.Receive(numRx)  // This allocates a new slice

// Optimized: Pre-allocate and reuse
type batchContext struct {
    descriptors [256]*XDPDescriptor  // Pre-allocated array
}

func (f *Forwarder) RunXDPBatchLoop(...) {
    batch := &batchContext{} // Created once
    
    for {
        // Receive returns slice pointing into batch.descriptors
        descs := batch.Receive(numRx)  // No allocation
    }
}
```

**Implementation approach**:
1. Create `batchContext` with pre-allocated descriptor array
2. Modify `xsk.Receive()` to populate pre-allocated array
3. Eliminate `make()` call from hot path

**Benefits**:
- Remove allocation from hot path
- ~50 ns per batch saved (64 descriptors per batch)
- ~0.8 ns per packet improvement

**Implementation effort**: 1-2 hours  
**Risk level**: Low (compile-time verified)

---

#### 1.3 Reduce Session Lookup Latency (50-100 ns gain)

**Problem**: RWMutex lock/unlock cycles on every packet.

**Solution**: Cache hot session pointers in thread-local storage.

```go
// Per-worker thread-local cache
type workerContext struct {
    sessionCache [8]*Session      // Cache last 8 accessed sessions
    sessionCacheHits uint64       // Metrics
}

func (wc *workerContext) getSession(sid [16]byte, f *Forwarder) *Session {
    // Quick check: is this the most recent session?
    if wc.lastSessionID == sid && wc.lastSession != nil {
        wc.sessionCacheHits++
        return wc.lastSession
    }
    
    // Miss: go to forwarder map
    sess := f.GetSession(sid)
    if sess != nil {
        wc.lastSessionID = sid
        wc.lastSession = sess
    }
    return sess
}
```

**Benefits**:
- Typical patterns access same session repeatedly
- 90%+ cache hit rate expected
- Avoids mutex on cache hit

**Implementation effort**: 2-3 hours  
**Risk level**: Low (cache misses gracefully degrade)

---

### Tier 2: Scaling Optimizations (5-10% Gain)

These optimizations improve overall system throughput by better resource utilization.

#### 2.1 Batch Size Tuning

**Current**: Batch size = 64, batch interval = 10ms

**Optimization**: Adaptive batch sizing based on load.

```go
// Low load: process quickly (10ms)
// High load: larger batches (128-256)

batchSize := calculateDynamicBatchSize(rxCount)
// If rxCount > threshold: 128, else: 64
```

**Benefits**:
- Better amortization of poll/transmit overhead
- Maintains low latency under low load
- Increases throughput under high load

**Implementation effort**: 2-3 hours  
**Risk level**: Medium (tuning parameters needed)

---

#### 2.2 Reduce Poll Interval Under Load

**Current**: Fixed 10ms poll interval

**Optimization**: Adaptive polling frequency.

```go
// If we received packets last round: poll sooner (1ms)
// If we got no packets: wait longer (50ms)

pollInterval := 10 * time.Millisecond
if lastRxCount > 0 {
    pollInterval = 1 * time.Millisecond
} else {
    pollInterval = 50 * time.Millisecond
}
```

**Benefits**:
- Responds faster to traffic bursts
- Reduces latency under bursty loads

**Implementation effort**: 1-2 hours  
**Risk level**: Low

---

### Tier 3: Advanced Optimizations (10-15% Gain)

These require kernel/hardware integration but offer significant gains.

#### 3.1 eBPF Hardware Steering (10-15% gain)

**Problem**: Software queue selection adds latency.

**Solution**: Let XDP BPF program redirect packets to correct queue.

```c
// XDP program in kernel
int xdp_prog(struct xdp_md *ctx) {
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;
    
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return XDP_DROP;
    
    // Calculate queue based on flow hash
    u32 flow_hash = hash(eth->src);
    u32 queue_idx = flow_hash % num_queues;
    
    return bpf_redirect_map(&tx_queues, queue_idx, 0);
}
```

**Benefits**:
- Eliminates software steering
- ~100-200 ns saved per packet
- Better cache locality

**Implementation effort**: 8-12 hours  
**Risk level**: Medium (kernel/eBPF expertise required)

---

#### 3.2 Zero-Copy TX Buffers

**Current**: Some memory copies in transmit path.

**Solution**: Direct frame transmission without intermediate buffering.

**Benefits**:
- Reduces memory bandwidth
- Faster frame TX

**Implementation effort**: 4-6 hours  
**Risk level**: Medium

---

## Optimization Roadmap

### Week 1: Implement Tier 1 (All 3 optimizations)
```
Monday-Tuesday:    Sharded session map + testing
Tuesday-Wednesday: Batch pre-allocation + testing
Wednesday:         Session caching + testing
Thursday:          Integration & performance validation
Friday:            Hardware preparation / buffer time
```

**Expected gain**: 15-20% (7.06 Gbps/core → 8.4+ Gbps/core)

### Week 2: Hardware Testing + Tier 2 (as needed)
```
Monday-Wednesday:  Hardware deployment & validation
Thursday-Friday:   Batch size tuning & adaptive polling
```

**Expected gain**: Additional 5-10% if implemented

### Week 3: Tier 3 (if needed for 10 Gbps)
```
Full week:         eBPF steering implementation & validation
```

**Expected gain**: 10-15% additional

---

## Performance Validation Plan

### Baseline Measurement
```bash
# Current performance snapshot
go test -bench=. -benchmem ./internal/crypto
go test -bench=. -benchmem ./internal/datapath/afxdp

# Expected output:
# - Crypto: 1.7 µs per packet
# - Batch: 1.8 µs per packet
# - Total: 3.5 µs per packet
```

### After Tier 1 Optimization
```bash
# Measure improvement
# Target: 3.5 µs → 2.8 µs (20% improvement)
# Per-core throughput: 7.06 Gbps → 8.8 Gbps
```

### Hardware Testing Protocol
```
1. Baseline run (5 minutes at 1 Gbps)
   - Measure: throughput, latency (p50/p99/p99.9)
   
2. Ramp test (10-30 Gbps in 1 Gbps increments)
   - Identify saturation point
   - Measure CPU utilization per core
   
3. Sustained run at target (10 Gbps for 1 hour)
   - Stability validation
   - Thermal validation
   
4. Load spike test (0 → 15 Gbps → 0)
   - Measure jitter response
   - Validate graceful degradation
```

---

## Optimization Checklist

### Tier 1 - Critical Path
- [ ] Sharded session map implementation
- [ ] Pre-allocated batch descriptors
- [ ] Thread-local session cache
- [ ] Tier 1 performance validation
- [ ] Integration tests passing

### Tier 2 - Scaling
- [ ] Adaptive batch sizing
- [ ] Adaptive poll interval
- [ ] Load testing with varying traffic patterns
- [ ] Performance validation

### Tier 3 - Advanced (if needed)
- [ ] eBPF program implementation
- [ ] Kernel integration testing
- [ ] Hardware steering validation

### Deployment
- [ ] Code review & approval
- [ ] Comprehensive benchmarking
- [ ] Stress testing (12+ hour runs)
- [ ] Production deployment

---

## Expected Results

| Optimization | Latency Gain | Throughput Gain | Effort |
|---|---|---|---|
| Sharded map | 100-200 ns | +3% | 2-3h |
| Batch pre-alloc | 50-100 ns | +1% | 1-2h |
| Session cache | 50-100 ns | +1% | 2-3h |
| **Tier 1 Total** | **200-400 ns** | **+5%** | **5-8h** |
| Batch tuning | 100-200 ns | +3% | 2-3h |
| Adaptive poll | 50-100 ns | +1% | 1-2h |
| **Tier 2 Total** | **150-300 ns** | **+4%** | **3-5h** |
| eBPF steering | 100-200 ns | +5% | 8-12h |
| **Tier 3 Total** | **100-200 ns** | **+5%** | **8-12h** |

**Combined expected gain**: 450-900 ns (13-26% improvement)

---

## Success Metrics

### Minimum Success (7.5+ Gbps)
- All Tier 1 optimizations complete
- Validated on hardware
- Stable for 1-hour duration

### Target Success (10+ Gbps)
- All Tier 1 + Tier 2 optimizations complete
- Validated on hardware
- Stable for 1-hour duration
- Per-core throughput: 2.5+ Mbps

### Maximum Stretch (10+ Gbps with low latency)
- All Tiers + eBPF steering
- p99 latency < 500 ns
- Zero packet drops under 10 Gbps load

---

## Risk Mitigation

### Performance Regressions
- Keep baseline benchmarks before each optimization
- Run full test suite after each change
- Compare benchmarks: `git diff <baseline>`

### Production Issues
- Stress test thoroughly (48 hour runs)
- Graceful degradation on overload
- Metrics/monitoring for real-time validation

### Hardware Compatibility
- Test on multiple NIC models (ixgbe, i40e, mlx5)
- Fallback to software steering if eBPF unavailable
- Support both older and newer kernel versions

---

This roadmap provides a clear path from 7 Gbps to 10+ Gbps with well-understood optimizations and measurable progress targets.

