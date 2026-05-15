# SMIP-MWP Project Timeline & Progress Tracker

## Status Delta (May 15, 2026)

Recently completed delivery items:
- Benchmark harness script implemented: `scripts/bench.sh`
- Optional pprof collection added to benchmark script (`--pprof`)
- Benchmark policy/documentation added: `benchmarks/README.md`
- Benchmark CI workflow added and scheduled: `.github/workflows/benchmarks.yml`
- Benchmark workflow expanded to Ubuntu + macOS matrix with per-OS artifacts

Adjusted tracker interpretation:
- Phase 2 remains active for hardware-backed AF_XDP validation
- "Comprehensive benchmarks" setup work is complete; remaining work is target-hardware validation and threshold governance

## Visual Timeline

```
MAY 2026                            JUNE 2026                          JULY 2026
├─────────────────────────────┼────────────────────────────┼────────────────────────┤

PHASE 1: MVP HANDSHAKE & FORWARDING (Weeks 1-2)
├─ Week 1 (May 14-20)
│  ├─ ML-KEM Integration
│  ├─ Handshake State Machine
│  └─ Crypto Unit Tests
│
├─ Week 2 (May 21-27)
│  ├─ UDP Overlay Transport
│  ├─ Single-Queue Forwarding
│  ├─ Routing Hardening
│  └─ Target: <5ms latency, >1k pps

PHASE 2: AF_XDP OPTIMIZATION & MULTI-QUEUE (Weeks 3-4)
├─ Week 3 (May 28-Jun 3)
│  ├─ AF_XDP Descriptor Reuse
│  ├─ Multi-Queue Forwarder
│  ├─ eBPF Steering Program
│  └─ Target: 3-5 Gbps forwarding
│
├─ Week 4 (Jun 4-10)
│  ├─ Frame Pooling
│  ├─ GC Optimization
│  ├─ Initial Benchmarking
│  └─ Validation: Approach 10 Gbps

PHASE 3: HARDENING & VERIFICATION (Month 2-3)
├─ Month 2 (Jun 11-Jul 8)
│  ├─ Lean 4 Proofs
│  ├─ Dynamic Queue Scaling
│  ├─ Comprehensive Benchmarks
│  └─ Security Audit
│
├─ Month 3 (Jul 9-Aug 5)
│  ├─ Federated Intelligence Integration
│  ├─ Formal Verification Alignment
│  ├─ Production Hardening
│  └─ Target: All metrics validated

PHASE 4: DEPLOYMENT (Month 4+)
├─ Month 4+ (Aug+)
│  ├─ Testbed Setup
│  ├─ Pilot Deployment
│  └─ Full Rollout
```

---

## Task Completion Tracker

### Phase 1: MVP (Weeks 1-2)

#### Week 1: Core Handshake (May 14-20)

| Task | Owner | Status | ETA | Blocker |
|------|-------|--------|-----|---------|
| 1.1.1 ML-KEM integration | - | 🔴 NOT STARTED | May 17 | None |
| 1.1.2 Handshake state machine | - | 🔴 NOT STARTED | May 19 | 1.1.1 |
| 1.1.3 Crypto unit tests | - | 🔴 NOT STARTED | May 20 | 1.1.2 |
| **Phase 1 Milestone** | - | 🔴 0% | May 20 | - |

#### Week 2: UDP Forwarding (May 21-27)

| Task | Owner | Status | ETA | Blocker |
|------|-------|--------|-----|---------|
| 1.2.1 cmd/mohawk-node entry point | - | 🔴 NOT STARTED | May 22 | None |
| 1.2.2 UDP overlay transport | - | 🔴 NOT STARTED | May 24 | 1.2.1 |
| 1.3.1 Routing engine hardening | - | 🔴 NOT STARTED | May 26 | None |
| **Validation**: <5ms latency, 1k+ pps | - | 🔴 NOT VALIDATED | May 27 | 1.2.2 |
| **Phase 1 Complete** | - | 🔴 0% | May 27 | - |

**Phase 1 Success Criteria**:
- ✅ Handshake completes in <50ms (LAN)
- ✅ UDP forwarding validated at 1k+ pps
- ✅ <5ms added latency vs plain UDP
- ✅ Zero packet loss under sustained load

---

### Phase 2: AF_XDP (Weeks 3-4)

#### Week 3: AF_XDP + Multi-Queue (May 28-Jun 3)

| Task | Owner | Status | ETA | Blocker |
|------|-------|--------|-----|---------|
| 2.1.1 AF_XDP descriptor reuse | - | 🔴 NOT STARTED | Jun 1 | Phase 1 ✅ |
| 2.1.2 Hardware validation script | - | 🔴 NOT STARTED | Jun 2 | None |
| 2.2.1 Multi-queue forwarder | - | 🔴 NOT STARTED | Jun 2 | 2.1.1 |
| 2.3.1 eBPF steering program | - | 🔴 NOT STARTED | Jun 3 | 2.2.1 |
| **Validation**: 3-5 Gbps forwarding | - | 🔴 NOT VALIDATED | Jun 3 | 2.3.1 |
| **Phase 2.1 Complete** | - | 🔴 0% | Jun 3 | - |

#### Week 4: Optimization (Jun 4-10)

| Task | Owner | Status | ETA | Blocker |
|------|-------|--------|-----|---------|
| 2.4.1 Frame pooling | - | 🔴 NOT STARTED | Jun 5 | 2.3.1 |
| 2.4.2 GC optimization | - | 🔴 NOT STARTED | Jun 7 | 2.4.1 |
| 3.3.1 Benchmark harness (proto) | - | 🔴 NOT STARTED | Jun 8 | 2.4.2 |
| **Initial Results**: Toward 10 Gbps | - | 🔴 NOT VALIDATED | Jun 10 | 3.3.1 |
| **Phase 2 Complete** | - | 🔴 0% | Jun 10 | - |

**Phase 2 Success Criteria**:
- ✅ 3-5 Gbps sustained forwarding
- ✅ <500ns per-packet latency
- ✅ Linear scaling: 2x queues ≈ 2x throughput
- ✅ eBPF steering adds <100ns overhead

---

### Phase 3: Hardening (Month 2-3)

#### Mid-Month 2 (Jun 11-25)

| Task | Owner | Status | ETA | Blocker |
|------|-------|--------|-----|---------|
| 3.1.1 Lean 4 routing model | - | 🔴 NOT STARTED | Jun 18 | Phase 2 ✅ |
| 3.2.1 Dynamic queue scaling | - | 🔴 NOT STARTED | Jun 22 | 3.1.1 |
| **Validation**: Auto-scale working | - | 🔴 NOT VALIDATED | Jun 25 | 3.2.1 |

#### Late Month 2 (Jun 26-Jul 8)

| Task | Owner | Status | ETA | Blocker |
|------|-------|--------|-----|---------|
| 3.3.1 Comprehensive benchmarks | - | 🔴 NOT STARTED | Jul 1 | Phase 2 ✅ |
| 3.4.1 Security audit | - | 🔴 NOT STARTED | Jul 6 | None |
| **Validation**: All targets met | - | 🔴 NOT VALIDATED | Jul 8 | 3.3.1 |
| **Phase 3a Complete** | - | 🔴 0% | Jul 8 | - |

#### Month 3 (Jul 9-Aug 5)

| Task | Owner | Status | ETA | Blocker |
|------|-------|--------|-----|---------|
| Federated intelligence integration | - | 🔴 NOT STARTED | Jul 20 | 3.3.1 |
| Formal verification alignment | - | 🔴 NOT STARTED | Jul 27 | 3.1.1 |
| Production hardening | - | 🔴 NOT STARTED | Aug 3 | All above |
| **Phase 3 Complete** | - | 🔴 0% | Aug 5 | - |

**Phase 3 Success Criteria**:
- ✅ 10 Gbps sustained forwarding (≥80% line rate)
- ✅ <1ms latency overhead (p99)
- ✅ Dynamic scaling works + no packet loss
- ✅ Lean 4 proofs verified
- ✅ Security audit: zero criticals
- ✅ GC pause <100μs, allocation <0.1%

---

## Performance Metrics Dashboard

### Current Baseline (Today)

| Metric | Current | Target | Delta | Status |
|--------|---------|--------|-------|--------|
| **Throughput** | 0 pps (no forwarding) | 10 Gbps | - | 🔴 N/A |
| **Latency** | N/A | <1ms overhead | - | 🔴 N/A |
| **Handshake** | N/A | <50ms | - | 🔴 N/A |
| **GC Pause** | TBD | <100μs | - | 🔴 TBD |
| **Allocation Rate** | TBD | <0.1% throughput | - | 🔴 TBD |

### Week 1 Check-In (Target: May 20)

| Metric | Expected | Actual | Status |
|--------|----------|--------|--------|
| Handshake latency | <50ms | - | 🔴 TODO |
| Crypto unit tests | 100% pass | - | 🔴 TODO |
| KEX determinism | Matching secrets | - | 🔴 TODO |

### Week 2 Check-In (Target: May 27)

| Metric | Expected | Actual | Status |
|--------|----------|--------|--------|
| UDP forwarding | 1k+ pps | - | 🔴 TODO |
| End-to-end latency | <5ms | - | 🔴 TODO |
| Packet loss | 0% | - | 🔴 TODO |

### Week 3 Check-In (Target: Jun 3)

| Metric | Expected | Actual | Status |
|--------|----------|--------|--------|
| AF_XDP throughput | 3-5 Gbps | - | 🔴 TODO |
| Per-packet latency | <500ns | - | 🔴 TODO |
| Multi-queue scaling | Linear (2x ≈ 2x) | - | 🔴 TODO |

### Week 4 Check-In (Target: Jun 10)

| Metric | Expected | Actual | Status |
|--------|----------|--------|--------|
| Approach 10 Gbps | >8 Gbps | - | 🔴 TODO |
| GC pause | <200μs | - | 🔴 TODO |
| Allocation rate | <1% | - | 🔴 TODO |

### Mid-Phase 3 Check-In (Target: Jul 8)

| Metric | Expected | Actual | Status |
|--------|----------|--------|--------|
| Sustained 10 Gbps | 10 Gbps ≥80% line rate | - | 🔴 TODO |
| Latency overhead | <1ms (p99) | - | 🔴 TODO |
| Dynamic scaling | Working, no loss | - | 🔴 TODO |
| Lean proofs | Loop-free proven | - | 🔴 TODO |

### Production Ready (Target: Aug 5)

| Metric | Required | Actual | Status |
|--------|----------|--------|--------|
| **Throughput** | 10 Gbps | - | 🔴 TODO |
| **Latency Overhead** | <1ms (p99) | - | 🔴 TODO |
| **Handshake** | <50ms | - | 🔴 TODO |
| **Routing Stability** | Zero flaps | - | 🔴 TODO |
| **Security** | Zero criticals | - | 🔴 TODO |
| **GC Pause** | <100μs | - | 🔴 TODO |
| **Scaling** | Linear | - | 🔴 TODO |

---

## Risk Status

| Risk | Current | Mitigation | Status |
|------|---------|-----------|--------|
| Go 1.24 `crypto/mlkem` delayed | 🟡 MEDIUM | Use circl interim | ✅ Active |
| AF_XDP unavailable | 🟡 MEDIUM | UDP overlay fallback | ✅ Active |
| eBPF compilation issues | 🟡 MEDIUM | Pre-compile + CI | ✅ Planned |
| Latency regression scaling | 🟡 MEDIUM | Profile + lock-free | ✅ Planned |
| Lean verification divergence | 🟡 MEDIUM | CI conformance tests | ✅ Planned |

---

## Decision Log

### Decision 1: Start with UDP Overlay (Not Direct AF_XDP)
- **Rationale**: Faster validation of handshake + routing logic
- **Risk**: Extra week on software path
- **Mitigation**: Phase 2 directly ports to AF_XDP (minimal changes)
- **Status**: ✅ APPROVED

### Decision 2: ML-KEM Stubbed Until Go 1.24
- **Rationale**: crypto/mlkem not available yet; uses cloudflare/circl as interim
- **Risk**: Crypto not PQC-resistant until Go 1.24
- **Mitigation**: Drop-in replacement once available
- **Status**: ✅ APPROVED (temporary)

### Decision 3: eBPF Steering as Phase 2 (Not Phase 1)
- **Rationale**: Phase 1 focuses on correctness; Phase 2 adds performance optimization
- **Risk**: Software RSS insufficient; may need eBPF earlier
- **Mitigation**: Monitor CPU cost of RSS; escalate if >10% overhead
- **Status**: ✅ APPROVED

---

## Blocker Escalation

### Critical Blockers (Project Risk)
- [ ] Go 1.24 crypto/mlkem delayed past June 1 → escalate to use alternative PQC library
- [ ] AF_XDP unavailable on all test hardware → evaluate Rust DPDK fork
- [ ] Lean 4 formalization blocked → defer to Month 3; continue with Go tests

### High Priority (Phase at Risk)
- [ ] Handshake takes >50ms in Week 1 → profile + optimize immediately
- [ ] UDP forwarding limited to <500 pps in Week 2 → review crypto overhead
- [ ] AF_XDP <2 Gbps in Week 3 → check hardware support + driver version

### Medium Priority (Optional Optimization)
- [ ] eBPF steering adds >200ns overhead → consider alternative steering method
- [ ] GC pause >200μs → profile heap allocations + tune pool sizes
- [ ] Lean proofs diverge from Go → extend conformance tests

---

## Weekly Sync Agenda Template

**Date**: [Week X Start Date]

**Completed Last Week**:
- [ ] Task A: Status
- [ ] Task B: Status

**Planned This Week**:
- [ ] Task C: Est. completion [Day]
- [ ] Task D: Est. completion [Day]

**Blockers**:
- [ ] Blocker 1: Impact | Mitigation
- [ ] Blocker 2: Impact | Mitigation

**Metrics** (current vs. target):
- Throughput: X pps → Target Y pps
- Latency: X ms → Target <Y ms

**Decisions Needed**:
- [ ] Decision A: Options + recommendation
- [ ] Decision B: Options + recommendation

**Next Week's Focus**:
- High priority: [Tasks]
- Medium priority: [Tasks]

---

## How to Update This Tracker

1. **Daily**: Update task status (🔴 NOT STARTED → 🟡 IN PROGRESS → 🟢 COMPLETE)
2. **Weekly**: Update metrics + blocker status
3. **Blockers**: Add/remove with impact assessment
4. **Risks**: Escalate if probability increases or impact assessed as higher
5. **Archive**: Save completed phases to separate doc for postmortem

---

## Sign-Off

| Role | Name | Signature | Date |
|------|------|-----------|------|
| Project Lead | - | - | - |
| Tech Lead (Crypto) | - | - | - |
| Tech Lead (Networking) | - | - | - |
| Product Owner | - | - | - |

---

## Appendix: Acronyms & Definitions

- **SMIP**: Sovereign Mohawk Internet Protocol
- **MWP**: Mohawk Wire Protocol
- **AF_XDP**: Address Family XDP (kernel-bypass I/O)
- **eBPF**: Extended Berkeley Packet Filter
- **PQC**: Post-Quantum Cryptography
- **KEX**: Key Exchange
- **AEAD**: Authenticated Encryption with Associated Data
- **GC**: Garbage Collection
- **STW**: Stop-The-World (GC pause)
- **RSS**: Receive Side Scaling (NIC-based load balancing)
- **UMEM**: User Memory pool (AF_XDP)
- **p99**: 99th percentile (latency tail)
- **pps**: Packets Per Second
- **LPM**: Longest Prefix Match
- **RWMutex**: Reader-Writer Mutual Exclusion Lock
