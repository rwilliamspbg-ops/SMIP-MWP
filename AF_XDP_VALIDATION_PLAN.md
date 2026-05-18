# AF_XDP Hardware Validation and Improvements Plan

**Date**: May 17, 2026  
**Objective**: Complete AF_XDP Hardware Validation (Phase 2) and implement all performance improvements

---

## Executive Summary

This document outlines the complete work required to:
1. Validate AF_XDP forwarding on target hardware at 3-5 Gbps (Phase 2, Week 3-4)
2. Implement performance improvements from pprof analysis
3. Meet all acceptance criteria for Phase 2 completion

**Current Status**: Phase 1 skeleton complete; Phase 2 ready to start  
**Target Completion**: June 10, 2026

---

## Phase 1 Status (May 14-27)

### Required Tasks
- [ ] ML-KEM Integration (real or interim placeholder)
- [ ] Handshake State Machine (UNINITIALIZED → AWAITING_PEER_PUBKEY → ESTABLISHED)
- [ ] Unit Tests (handshake, state machine, replay, timeout)
- [ ] Crypto integration tests passing

### Success Criteria
- ✅ Handshake completes in <50ms (LAN)
- ✅ UDP forwarding validated at 1k+ pps
- ✅ <5ms added latency vs plain UDP
- ✅ Zero packet loss under sustained load

---

## Phase 2: AF_XDP Optimization (May 28-Jun 10)

### Week 3 Tasks (May 28-Jun 3): AF_XDP Infrastructure

#### 2.1 AF_XDP Descriptor Reuse ⭐ CRITICAL
**Status**: Skeleton exists; needs production implementation

**What to implement**:
- [ ] Zero-copy UMEM frame pooling
- [ ] Descriptor reuse without malloc per packet
- [ ] RX→TX descriptor chain (avoid data copies)
- [ ] Batch descriptor filling

**Expected Gain**: 30-50% throughput improvement

**Files to modify**:
- `internal/datapath/afxdp/forwarder_xdp_umem.go` - Add UMEM lifecycle, frame reuse
- `internal/datapath/afxdp/forwarder_xdp_rxtx.go` - Descriptor reuse logic

#### 2.2 Hardware Validation Script
**Status**: Skeleton exists (`scripts/test_xdp.sh`); needs completion

**Extend with**:
- [ ] NIC driver capability checks (ixgbe, i40e, mlx5)
- [ ] Kernel AF_XDP feature detection
- [ ] Memory setup verification (hugepages, etc.)
- [ ] Hardware offload capability reporting
- [ ] Pre-run checklist validation

#### 2.3 Multi-Queue Forwarder
**Status**: Skeleton exists; needs worker pool implementation

**What to implement**:
- [ ] Per-CPU worker spawn (1 worker per queue)
- [ ] CPU affinity binding (NUMA-aware)
- [ ] Worker lifecycle management
- [ ] Per-worker metrics aggregation

**Expected Gain**: Linear scaling with CPU count (2x queues ≈ 2x throughput)

**Files to modify**:
- `internal/datapath/afxdp/worker_pool.go` - Extend pool impl

#### 2.4 eBPF Steering Program
**Status**: Stub only; needs full implementation

**Implement**:
- [ ] BPF program that classifies packets by flow (source IP)
- [ ] Attach to RX ring with XDP_REDIRECT
- [ ] Route packets to appropriate queue (hash modulo N)
- [ ] Add to `internal/datapath/afxdp/ebpf/` (new folder)

**Expected Gain**: Hardware-level load balancing (<100ns overhead)

---

### Week 4 Tasks (Jun 4-10): Optimization & Benchmarking

#### 2.5 Frame Pooling in Hot Path
**Status**: Initial pktPool exists; needs production tuning

**Extend**:
- [ ] Pre-allocate buffer pool at startup
- [ ] Tune pool size to available UMEM
- [ ] Reduce per-packet allocation overhead
- [ ] Profile allocation patterns

**Expected Gain**: 10-20% GC pressure reduction

#### 2.6 GC Optimization
**Status**: Baseline benchmarks exist; gaps in optimization

**Implement**:
- [ ] Replace unbounded session cache with LRU (max N entries)
- [ ] Use object pooling for temporary buffers
- [ ] Profile GC pause times under load

**Expected Gain**: <100μs GC pause (p99)

#### 2.7 Initial Benchmarking
**Status**: Benchmark runner (`scripts/bench.sh`) exists

**Run**:
- [ ] Full suite on target hardware (`withafxdp` tag)
- [ ] Collect CPU profiles
- [ ] Document baseline metrics
- [ ] Identify remaining hotspots

---

## Phase 3: Hardening & Verification (Jun 11+)

### 3.1 Comprehensive Benchmarks
- [ ] 24-hour sustained throughput test (target: 10 Gbps)
- [ ] Latency histogram collection (p50, p99, p99.9)
- [ ] GC pause analysis
- [ ] CPU utilization per worker

### 3.2 Formal Verification
- [ ] Lean 4 proof of routing loop-freedom
- [ ] Dynamic queue scaling proof

### 3.3 Security Audit
- [ ] Side-channel analysis
- [ ] DoS mitigation verification

---

## Performance Targets (Acceptance Criteria)

| Metric | Target | Status |
|--------|--------|--------|
| **Throughput** | 3-5 Gbps (Week 4) | 🔴 Not validated |
| **Throughput** | ≥10 Gbps (Phase 3) | 🔴 Not validated |
| **Per-packet latency** | <500ns | 🔴 Not validated |
| **Handshake** | <50ms | 🟡 Skeleton only |
| **GC pause** | <100μs (p99) | 🟡 Not profiled |
| **Linear queue scaling** | 2x queues ≈ 2x throughput | 🟡 Not tested |

---

## Recommended Optimization Sequence

### Priority 1 (Critical Path)
1. **Complete Phase 1 handshake** (May 20)
   - This unblocks everything else
2. **AF_XDP descriptor reuse** (May 28-Jun 1)
   - Biggest throughput gain
3. **Multi-queue forwarder** (Jun 2-3)
   - Enables linear scaling

### Priority 2 (Performance)
4. **eBPF steering** (Jun 3)
   - Hardware-level load balancing
5. **Frame pooling** (Jun 5)
   - Reduce allocations
6. **GC optimization** (Jun 7)
   - Reduce pause times

### Priority 3 (Validation)
7. **Comprehensive benchmarking** (Jun 8-10)
   - Profile real hardware
   - Document results

---

## Implementation Dependencies

```
Phase 1: Handshake & Session (May 14-27)
  ├─ 1.1 ML-KEM Integration
  ├─ 1.2 State Machine
  └─ 1.3 Unit Tests
         ↓
Phase 2: AF_XDP Optimization (May 28-Jun 10)
  ├─ 2.1 Descriptor Reuse ◄── CRITICAL BLOCKER
  ├─ 2.2 Hardware Validation Script
  ├─ 2.3 Multi-Queue (depends on 2.1)
  ├─ 2.4 eBPF Steering (depends on 2.3)
  ├─ 2.5 Frame Pooling (depends on 2.1)
  ├─ 2.6 GC Optimization
  └─ 2.7 Benchmarking ◄── FINAL VALIDATION
         ↓
Phase 3: Hardening & Verification (Jun 11+)
  ├─ 3.1 Sustained load test
  ├─ 3.2 Lean 4 proofs
  └─ 3.3 Security audit
```

---

## Hardware Requirements

For validation to work:
- Linux kernel 5.10+ (6.x preferred)
- XDP-capable NIC (ixgbe, i40e, mlx5, etc.)
- Root or equivalent privileges
- Dedicated test host (no noisy neighbors)
- Optional: secondary host for latency testing

---

## Next Steps

1. **Mark Phase 1 complete** or finish blocking items
2. **Implement 2.1 (Descriptor Reuse)** - biggest ROI
3. **Run initial benchmarks** on target hardware
4. **Iterate** based on profiling results

---

## Questions to Answer

- [ ] Go version: Is Go 1.24 available? (needed for crypto/mlkem)
- [ ] Hardware: What NIC drivers available on test host?
- [ ] Target: 3-5 Gbps or 10 Gbps baseline?
- [ ] Latency: What RTT for WAN profiles?

