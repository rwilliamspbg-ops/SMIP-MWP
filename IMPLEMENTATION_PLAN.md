# SMIP-MWP Concrete Implementation Plan

## Quick Reference: Current State vs. End State

### Current State (Today)
- ✅ Wire header serialization (96 bytes)
- ✅ Hybrid KEX skeleton (x25519 + ML-KEM stub with random bytes)
- ✅ AEAD session (AES-GCM + ChaCha20 fallback) with in-place ops
- ✅ AF_XDP forwarder structure (socket lifecycle, batch processing)
- ✅ Routing policy engine (exact-match lookup)
- ⚠️ **NO working handshake flow** (state machine not exercised end-to-end)
- ⚠️ **NO actual forwarding** (no main.go entry point)
- ⚠️ **NO eBPF programs** (planned but not implemented)
- ⚠️ **NO benchmarking** (targets known but unvalidated)

### End State (Production Ready)
- ✅ Full PQC hybrid handshake (<50ms, all states exercised)
- ✅ 10 Gbps encrypted forwarding on AF_XDP (zero-copy descriptor reuse)
- ✅ Multi-queue load balancing with dynamic scaling
- ✅ eBPF steering for sovereign flow classification
- ✅ Formal verification (Lean 4) of routing invariants
- ✅ Comprehensive benchmarking suite
- ✅ Production-grade security (no side-channels, DoS mitigation)

---

## Phase 1: MVP - Core Handshake & Single-Queue Forwarding

### Task 1.1.1: Integrate Real ML-KEM (crypto/mlkem Go 1.24+)

**File**: `crypto/kex.go`

**Current Code** (lines ~50-70):
```go
// ML-KEM-768 stub: generate random 1184-byte "public key" placeholder.
// TODO: replace with crypto/mlkem KEM once Go 1.24+ stdlib API stabilises.
h.mlkemPub = make([]byte, 1184)
h.mlkemPriv = make([]byte, 2400)
if _, err := io.ReadFull(rng, h.mlkemPub); err != nil {
    return nil, fmt.Errorf("kex: mlkem pub gen: %w", err)
}
```

**Change To** (when Go 1.24 available):
```go
import "crypto/mlkem"

// Generate ML-KEM-768 keypair
mlkemKey, err := mlkem.GenerateKey768(rng)
if err != nil {
    return nil, fmt.Errorf("kex: mlkem key gen: %w", err)
}
h.mlkemPub = mlkemKey.PublicKey().Bytes()
h.mlkemPriv = mlkemKey.Bytes()  // For later decapsulation
```

**Acceptance Criteria**:
- ✅ CI passes on Go 1.24+ (skip test on <1.24)
- ✅ ML-KEM key size verified (public: 1184B, private: 2400B)
- ✅ Decapsulation produces consistent shared secret

---

### Task 1.1.2: Complete Handshake State Machine

**File**: `internal/crypto/hybrid.go` (enhance existing skeleton)

**Current Code** (lines ~1-50):
```go
type HybridSession struct {
    State      SessionState
    Aead       cipher.AEAD
    // ... incomplete
}

// ProcessHandshakeMessage - state transitions stubbed
func (s *HybridSession) ProcessHandshakeMessage(...) error {
    switch s.State {
    // ... partial implementation
    }
}
```

**Tasks**:
1. Complete state transition logic for all 4 states
2. Add timeout tracking (start handshake → 30s timeout → abort)
3. Add replay protection (sequence window of 64 bits)
4. Add retry backoff (exponential, max 3 retries)

**Specific Changes**:
- [ ] Add `HandshakeTimeout` field + timeout goroutine
- [ ] Add `SeqWindow` uint64 for replay detection
- [ ] Add `RetryCount` int + exponential backoff logic
- [ ] Update `ProcessHandshakeMessage()` to enforce state ordering
- [ ] Add abort/reset if timeout or max retries exceeded

**Acceptance Criteria**:
- ✅ State machine enforces: UNINITIALIZED → AWAITING_PEER_PUBKEY → READY_FOR_AUTH → ESTABLISHED (no out-of-order)
- ✅ Timeout aborts handshake after 30s
- ✅ Replay window rejects duplicate packets within 64-packet window
- ✅ 3 retries then abort + log

---

### Task 1.1.3: Unit Tests for Full Handshake Flow

**New File**: `crypto/kex_test.go`

**Structure**:
```go
func TestHybridHandshakeFullFlow(t *testing.T) {
    // 1. Alice: generate keypair
    alice, _ := NewHybridKEX(rand.Reader)
    alicePub := alice.PublicKey()
    
    // 2. Bob: generate keypair + handshake
    bob, _ := NewHybridKEX(rand.Reader)
    bobPub := bob.PublicKey()
    
    // 3. Alice + Bob derive shared secret
    aliceSecret, _ := alice.Handshake(bobPub)
    bobSecret, _ := bob.Handshake(alicePub)
    
    // 4. Secrets match (critical!)
    if !bytes.Equal(aliceSecret, bobSecret) {
        t.Fatalf("shared secrets don't match")
    }
}

func TestStateMachineOrdering(t *testing.T) {
    // Verify state machine rejects invalid transitions
}

func TestReplayProtection(t *testing.T) {
    // Verify duplicate packets rejected
}

func TestHandshakeTimeout(t *testing.T) {
    // Verify timeout aborts after 30s
}
```

**Acceptance Criteria**:
- ✅ All tests pass
- ✅ Coverage >90% for crypto package
- ✅ Handshake time measured: target <50ms on LAN

---

### Task 1.2.1: Create cmd/mohawk-node Entry Point

**New File**: `cmd/mohawk-node/main.go`

**Structure**:
```go
package main

import (
    "context"
    "flag"
    "log"
    
    "smip-mwp-forge/internal/datapath"
    "smip-mwp-forge/internal/routing"
)

func main() {
    mode := flag.String("mode", "overlay", "overlay|xdp")
    listen := flag.String("listen", ":9000", "Listen address")
    flag.Parse()
    
    router := routing.NewRouter()
    
    switch *mode {
    case "overlay":
        startUDPOverlay(*listen, router)
    case "xdp":
        startAFXDP(*listen, router)
    }
}

func startUDPOverlay(addr string, router *routing.Router) {
    // UDP listener → packet processing
}

func startAFXDP(addr string, router *routing.Router) {
    // AF_XDP forwarder
}
```

**Acceptance Criteria**:
- ✅ Compiles and runs (./cmd/mohawk-node --help)
- ✅ Accepts --mode and --listen flags
- ✅ Graceful startup/shutdown

---

### Task 1.2.2: UDP Overlay Transport

**New File**: `internal/transport/udp.go`

**Structure**:
```go
type UDPTransport struct {
    conn     *net.UDPConn
    sessions map[[16]byte]*Session
    router   *routing.Router
    logger   *zap.Logger
}

func (t *UDPTransport) Listen(addr string) error {
    // Bind UDP socket
}

func (t *UDPTransport) ProcessPacket(buf []byte, addr *net.UDPAddr) {
    // 1. Parse header
    // 2. Lookup session
    // 3. Decrypt or initiate handshake
    // 4. Route + forward
}
```

**Tasks**:
- [ ] UDP listener on port 9000
- [ ] Packet parse → session lookup
- [ ] Decrypt path (fast path)
- [ ] Handshake initiation (slow path)
- [ ] Forward to next hop (unicast UDP to peer)

**Acceptance Criteria**:
- ✅ Can send/receive SMIP packets between two instances
- ✅ Latency <5ms per packet (software baseline)
- ✅ No drops under 1000 pps

---

### Task 1.3.1: Harden Routing Policy Engine

**File**: `internal/routing/router.go` (enhance existing)

**Current Limitations**:
- Only exact-match lookup
- No priority ranking
- Slow predictive lookup

**New Capabilities**:
- [ ] Longest-prefix-match (LPM) for CIDR-style policies
- [ ] Priority chain: (0=highest) → fallback to lower priority
- [ ] Concurrent updates with RWMutex (fine-grained locking)
- [ ] Metrics: lookup latency, hit rate per policy
- [ ] Predictive stub: accept external route updates

**Code Changes**:
```go
type Router struct {
    sync.RWMutex
    policies     map[uint64]RoutePolicy
    metrics      RouterMetrics  // NEW
    predictiveFn PredictiveFunc // NEW
}

type RouterMetrics struct {
    LookupLatencyNs atomic.Int64
    HitCount        atomic.Int64
    MissCount       atomic.Int64
}

// LookupPolicy now:
// 1. Try exact-match
// 2. Try LPM (longer prefix first)
// 3. Try predictive (if configured)
// 4. Fallback to default
```

**Acceptance Criteria**:
- ✅ Lookup latency <100ns (measured with pprof)
- ✅ Support 1M+ policies (tested with synthetic load)
- ✅ No allocations in hot path

---

## Phase 2: AF_XDP Optimization & Multi-Queue

### Task 2.1.1: Complete AF_XDP Descriptor Reuse

**File**: `internal/datapath/afxdp/forwarder.go` (enhance existing)

**Current Status**: Framework exists; needs plumbing.

**Missing Pieces**:
- [ ] `reuseDescriptorForForward()` implementation (in-place modification)
- [ ] `EncryptInPlace()` in crypto package (called from here)
- [ ] Session lookup in fast path (cache hit optimization)
- [ ] Benchmark: zero-copy validation

**Key Function**:
```go
func (f *Forwarder) reuseDescriptorForForward(
    frame []byte, desc *xdp.Desc, sess *Session, hdr wire.Header,
) bool {
    // 1. In-place header rewrite (DstID → NextHopID, etc.)
    // 2. In-place re-encrypt with new seq number
    // 3. Reuse descriptor for TX (same memory region)
    // 4. Update desc.Len and submit
}
```

**Acceptance Criteria**:
- ✅ 3-5 Gbps forwarding on test hardware
- ✅ Memory stays in UMEM (perf counter validation)
- ✅ <500ns per-packet latency (including crypto)

---

### Task 2.1.2: Test AF_XDP on Hardware

**New File**: `scripts/test_xdp.sh`

**Script Tasks**:
```bash
#!/bin/bash
# 1. Check kernel version (≥5.8 for AF_XDP)
# 2. Check NIC driver support (i40e, ixgbe, mlx5, etc.)
# 3. Load XDP program
# 4. Send test traffic
# 5. Measure throughput + latency
```

**Acceptance Criteria**:
- ✅ Script runs successfully on supported hardware
- ✅ Gracefully skips on unsupported platforms
- ✅ Reports throughput + latency numbers

---

### Task 2.2.1: Multi-Queue Forwarder

**New File**: `internal/datapath/afxdp/multi.go` (from outline)

**Key Structure**:
```go
type MultiQueueForwarder struct {
    forwarders []*Forwarder
    config     Config
    wg         sync.WaitGroup
}

func (m *MultiQueueForwarder) Run(ctx context.Context) {
    for i, fwd := range m.forwarders {
        m.wg.Add(1)
        go func(id int, f *Forwarder) {
            defer m.wg.Done()
            runtime.LockOSThread()  // Pin to core
            f.Run(ctx)
        }(i, fwd)
    }
}
```

**Tasks**:
- [ ] Auto-detect NIC queue count
- [ ] Create one Forwarder per queue
- [ ] Pin each to dedicated CPU core (if possible)
- [ ] Aggregate stats from all queues

**Acceptance Criteria**:
- ✅ Linear throughput scaling (2x queues ≈ 2x throughput)
- ✅ Per-queue latency consistent
- ✅ CPU utilization evenly distributed

---

### Task 2.3.1: eBPF XDP Steering Program

**New File**: `bpf/xdp_steer.c`

**Program Logic**:
```c
SEC("xdp")
int xdp_smip_steer(struct xdp_md *ctx) {
    // 1. Parse Ethernet + SMIP header
    // 2. Extract SrcID, DstID, FlowLabel
    // 3. Compute flow key hash
    // 4. Lookup in flow_to_queue map
    // 5. Redirect to target queue via XSKMAP
}
```

**Related Files**:
- [ ] `internal/datapath/afxdp/ebpf.go` (loader + map management)
- [ ] Build rule in Makefile (compile .bpf.c → .bpf.o)

**Acceptance Criteria**:
- ✅ Program compiles (clang available)
- ✅ Attaches to XDP hook
- ✅ Flow steering adds <100ns latency

---

### Task 2.4.1: Frame Pooling

**New File**: `internal/datapath/afxdp/pool.go` (from outline)

**Key Code**:
```go
type FramePool struct {
    pool sync.Pool
    size int
}

func (p *FramePool) Get() []byte {
    b := p.pool.Get().(*[]byte)
    *b = (*b)[:0]  // Reset length
    return *b
}

func (p *FramePool) Put(b []byte) {
    if cap(b) != p.size {
        return  // Wrong size, discard
    }
    p.pool.Put(&b)
}
```

**Integration Points**:
- [ ] Use in `prepareForward()` (get frame from pool)
- [ ] Return to pool after TX submit

**Acceptance Criteria**:
- ✅ GC pause time <100μs (measured with GC trace)
- ✅ Pool efficiency >95% (hit/miss ratio)
- ✅ Allocation rate <0.1% of throughput

---

## Phase 3: Formal Verification & Production Hardening

### Task 3.1.1: Lean 4 Routing Model

**Location**: `Sovereign-Mohawk-Proto` repository (external)

**Key Theorems**:
- Loop-freedom under topology changes
- Policy compliance (no unauthorized deviations)
- Convergence time bounds (<5s single-failure, <15s multi-failure)

**Integration**:
- [ ] Generate Go test cases from Lean proofs
- [ ] Add CI gate: conformance tests must pass

**Acceptance Criteria**:
- ✅ Loop-freedom proven
- ✅ Convergence bounds met
- ✅ Zero divergence between Lean model + Go implementation

---

### Task 3.2.1: Dynamic Queue Scaling

**New File**: `internal/datapath/afxdp/scaler.go` (from outline)

**Key Logic**:
```go
type Scaler struct {
    multi *MultiQueueForwarder
    config ScalingConfig
}

func (s *Scaler) monitorLoop(ctx context.Context) {
    for range ticker.C {
        stats := s.multi.GetStats()
        ppsPerQueue := stats.RxPackets / stats.Queues
        
        if ppsPerQueue > threshold {
            s.scaleUp()    // Add queue
        } else if ppsPerQueue < lowThreshold {
            s.scaleDown()  // Remove queue
        }
    }
}
```

**Acceptance Criteria**:
- ✅ Auto-scale within 5s
- ✅ No packet loss during scale
- ✅ Prevents oscillation (cooldown enforced)

---

### Task 3.3.1: Benchmark Harness

**New File**: `scripts/bench.sh`

**Benchmarks**:
1. **Throughput**: 10 Gbps sustained (bidirectional)
2. **Latency**: <1ms overhead (p99)
3. **Handshake**: <50ms (100 samples)
4. **Memory**: GC pause <100μs, allocation <0.1%

**Script Outline**:
```bash
#!/bin/bash
# 1. Start two SMIP nodes
# 2. Send traffic (iperf-style, but SMIP packets)
# 3. Measure + log results
# 4. Compare against targets
# 5. Generate report
```

**Acceptance Criteria**:
- ✅ All targets met
- ✅ Results reproducible (5-run avg + stddev)
- ✅ Report generated (CSV + graphs)

---

### Task 3.4.1: Security Audit Checklist

**Items**:
- [ ] Crypto code review (no key material in logs)
- [ ] Sequence number overflow checks
- [ ] Replay attack mitigation verified
- [ ] DoS protections (rate limiting, anti-amplification)
- [ ] Side-channel analysis (timing attacks)

**Acceptance Criteria**:
- ✅ Audit report completed
- ✅ Zero critical findings
- ✅ High findings remediated

---

## File Structure After All Tasks Complete

```
smip-mwp-forge/
├── cmd/
│   └── mohawk-node/
│       └── main.go                    [Entry point]
├── internal/
│   ├── crypto/
│   │   ├── hybrid.go               [Session + AEAD + state machine]
│   │   ├── hybrid_test.go
│   │   ├── kex_test.go             [NEW: Full handshake tests]
│   │   └── security.go             [NEW: Replay/overflow checks]
│   ├── transport/
│   │   └── udp.go                  [NEW: UDP overlay]
│   ├── datapath/
│   │   └── afxdp/
│   │       ├── forwarder.go        [Enhanced: complete flow]
│   │       ├── multi.go            [NEW: Multi-queue]
│   │       ├── pool.go             [NEW: Frame pooling]
│   │       ├── scaler.go           [NEW: Dynamic scaling]
│   │       └── ebpf.go             [NEW: eBPF loader]
│   └── routing/
│       ├── router.go               [Enhanced: LPM + priority + metrics]
│       └── router_test.go
├── bpf/
│   ├── xdp_steer.c                [NEW: eBPF steering program]
│   └── Makefile                    [NEW: Build rules]
├── scripts/
│   ├── bench.sh                    [NEW: Benchmark harness]
│   └── test_xdp.sh                [NEW: Hardware validation]
├── go.mod                          [Updated with new deps]
├── Makefile                        [NEW: Build targets]
└── IMPLEMENTATION_PLAN.md          [This file]
```

---

## Dependency Versions

Add to `go.mod`:
```
require (
    cilium/ebpf v0.16.0+
    slavc/xdp v0.3.0+
    vishvananda/netlink v1.3.0+
    uber-go/zap v1.27.0+
    cloudflare/circl v1.5.0+
    golang.org/x/crypto v0.31.0+
    golang.org/x/net v0.33.0+
    golang.org/x/sys v0.28.0+
)
```

---

## Weekly Execution Checklist

### Week 1: MVP Handshake
- [ ] ML-KEM integration (or stub confirmation)
- [ ] Handshake state machine complete
- [ ] kex_test.go: full flow passing
- [ ] cmd/mohawk-node scaffolding
- **Goal**: Handshake working end-to-end

### Week 2: UDP Overlay
- [ ] UDP transport layer
- [ ] Single-queue forwarding validation
- [ ] Routing engine hardened
- [ ] Initial latency/throughput measurements (<5ms, <1000 pps)
- **Goal**: Forwarding working on software path

### Week 3: AF_XDP Optimization
- [ ] AF_XDP descriptor reuse complete
- [ ] Multi-queue setup
- [ ] eBPF steering
- [ ] Validation: 3-5 Gbps forwarding
- **Goal**: Hardware-accelerated forwarding

### Week 4: Optimization & Benchmarking
- [ ] Frame pooling integrated
- [ ] GC optimization
- [ ] Benchmark harness complete
- [ ] Initial results vs. targets
- **Goal**: Get close to 10 Gbps target

---

## Success Criteria Summary

| Phase | Deliverable | Success Metric |
|-------|-------------|----------------|
| **1** | Handshake flow | <50ms, all states exercised |
| **1** | UDP forwarding | <5ms latency, >1000 pps throughput |
| **2** | AF_XDP + multi-queue | 3-5 Gbps, <500ns/pkt latency |
| **2** | eBPF steering | Flow policy updates <1s |
| **3** | Benchmarks | 10 Gbps target validated |
| **3** | Security | Zero critical audit findings |

---

## How to Use This Plan

1. **Print This**: Reference during standups
2. **Track Tasks**: Check off as completed
3. **Update Metrics**: Record actual vs. target
4. **Flag Blockers**: If task delayed >1 day, escalate
5. **Weekly Sync**: Review completion %, adjust next week

**Start with Task 1.1.1 (ML-KEM integration)** — it's on the critical path.
