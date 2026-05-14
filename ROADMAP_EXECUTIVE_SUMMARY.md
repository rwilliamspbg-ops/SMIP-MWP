# SMIP-MWP Project Roadmap: Executive Summary

**Project**: Sovereign Mohawk Internet Protocol - Forge Track  
**Objective**: Production-grade, post-quantum secure transport stack  
**Timeline**: 12 weeks (May 14 - Aug 5, 2026)  
**Status**: Roadmap finalized; Week 1 ready to start  

---

## Strategic Goals

| Goal | Target | By When | Business Impact |
|------|--------|---------|-----------------|
| **Encryption Throughput** | 10 Gbps sustained | Aug 5 | Competitive with modern VPNs/SDN |
| **Latency** | <1ms overhead vs plain UDP | Aug 5 | Suitable for low-latency services (trading, gaming) |
| **PQC Handshake** | <50ms full hybrid key exchange | Aug 5 | Post-quantum resistant from day 1 |
| **Routing Stability** | Zero routing flaps; loop-free proven | Aug 5 | Formal verification (Lean 4) |
| **Security** | Zero critical audit findings | Aug 5 | Production-deployable |

---

## High-Level Phases

```
PHASE 1: MVP HANDSHAKE & FORWARDING
├─ Week 1-2 (May 14-27)
├─ Deliverable: Full PQC hybrid handshake + UDP overlay forwarding
├─ Success: Handshake <50ms, forwarding 1k+ pps, <5ms latency
└─ Resource: 1-2 engineers (crypto + transport)

PHASE 2: AF_XDP OPTIMIZATION & MULTI-QUEUE
├─ Week 3-4 (May 28-Jun 10)
├─ Deliverable: Kernel-bypass forwarding, zero-copy descriptor reuse, multi-queue scaling
├─ Success: 3-5 Gbps forwarding, <500ns/pkt, linear queue scaling
└─ Resource: 1-2 engineers (systems + performance)

PHASE 3: HARDENING & FORMAL VERIFICATION
├─ Month 2-3 (Jun 11-Aug 5)
├─ Deliverable: Dynamic scaling, comprehensive benchmarking, security audit, Lean 4 proofs
├─ Success: 10 Gbps sustained, all metrics validated, zero audit findings
└─ Resource: 2-3 engineers (verification + security + performance)

PHASE 4: DEPLOYMENT (Optional, beyond roadmap)
├─ Month 4+ (Aug+)
├─ Deliverable: Testbed setup, pilot deployment, full rollout
└─ Resource: Operations + DevOps team
```

---

## Current State → Production Ready

### Current (Today, May 14)

| Component | Status | Notes |
|-----------|--------|-------|
| Wire Format | ✅ Complete | 96-byte header, tested serialization |
| Crypto KEX | ⚠️ Stubbed | ML-KEM uses random bytes; x25519 working |
| AEAD Session | ✅ Core | AES-GCM + ChaCha20; in-place ops ready |
| AF_XDP Framework | ✅ Skeleton | Socket lifecycle, batch processing structure |
| Routing Engine | ✅ Basic | Exact-match lookup; predictive stub |
| **Actual Forwarding** | ❌ NONE | No main entry point; no live forwarding |

### Production Ready (Aug 5)

| Component | Status | Notes |
|-----------|--------|-------|
| Wire Format | ✅ Production | Tested, optimized header |
| Crypto KEX | ✅ PQC | Real ML-KEM (Go 1.24+) + ML-DSA signatures |
| AEAD Session | ✅ Hardened | In-place ops, replay protection, timeouts |
| AF_XDP Fast Path | ✅ 10 Gbps | Zero-copy, multi-queue, eBPF steering |
| Routing Engine | ✅ Intelligent | Predictive routing + federated learning |
| **Actual Forwarding** | ✅ VERIFIED | 10 Gbps sustained, <1ms latency, zero flaps |

---

## Key Milestones & Dates

| Date | Milestone | Success Criteria | Owner |
|------|-----------|-----------------|-------|
| **May 20** | Week 1 Complete | Handshake <50ms, all tests passing | Crypto Lead |
| **May 27** | Phase 1 Complete | UDP forwarding 1k+ pps, <5ms latency | Transport Lead |
| **Jun 3** | Phase 2.1 Complete | AF_XDP 3-5 Gbps, multi-queue linear scaling | Networking Lead |
| **Jun 10** | Phase 2 Complete | Frame pooling, GC <200μs, >8 Gbps | Perf Lead |
| **Jul 8** | Phase 3 Complete | 10 Gbps + latency validated, Lean proofs | All |
| **Aug 5** | **PRODUCTION READY** | All metrics met, audit clean | PM |

---

## Success Metrics (Production Acceptance)

### Hard Targets

| Metric | Target | Validation |
|--------|--------|-----------|
| **Throughput** | ≥10 Gbps (sustained, bidirectional) | 24-hour burn test on production NIC |
| **Latency** | <1ms overhead (p99) vs plain UDP | 100k packet sample, LAN + WAN profiles |
| **Handshake** | <50ms full PQC (avg + p99) | 100 trials, both LAN and RTT-simulated WAN |
| **Routing** | Zero flaps; convergence <15s | Formal proof + conformance test suite |
| **GC Pause** | <100μs (p99) | GC trace analysis on sustained load |
| **Security** | Zero critical findings | Third-party audit report |

### Soft Targets (Performance Bonuses)

- <500ns per-packet latency in fast path
- >95% frame pool hit rate
- Linear multi-queue scaling (2x queues = 2x throughput)
- <10μs eBPF steering overhead

---

## Resource Requirements

### Team Composition

| Role | FTE | Responsibilities | Timeline |
|------|-----|------------------|----------|
| **Crypto Lead** | 1.0 | ML-KEM integration, session state machine, PQC hardening | Weeks 1-4, then review |
| **Networking Lead** | 1.0 | UDP transport, AF_XDP, multi-queue, eBPF | Weeks 2-4, then optimization |
| **Perf Engineer** | 0.5 | Benchmarking, GC tuning, profiling | Ongoing (Weeks 3-12) |
| **Formal Methods** | 0.5 | Lean 4 routing proofs, conformance tests | Weeks 5-8 |
| **Security** | 0.5 | Audit prep, threat modeling, rate limiting | Weeks 8-10 |
| **DevOps** | 0.5 | CI/CD, testbed setup, deployment | Weeks 10-12 (if go-live) |

**Total**: ~4 FTE, flexible with demand curve

### Hardware & Tools

- **Test Hardware**: 12.5+ Gbps NIC (Intel 82599ES, Mellanox ConnectX, etc.) on two nodes
- **Kernel**: Linux 5.8+ (AF_XDP support) + LLVM/Clang (eBPF compilation)
- **Go**: 1.24+ (for crypto/mlkem; 1.23 with interim circl)
- **Profiling**: pprof, perf, flamegraph, GC tracer
- **CI/CD**: GitHub Actions + codecov

---

## Risk Assessment & Mitigation

| Risk | Likelihood | Impact | Mitigation | Status |
|------|-----------|--------|-----------|--------|
| Go 1.24 `crypto/mlkem` delayed | MEDIUM | HIGH | Use cloudflare/circl as interim | ✅ Active |
| AF_XDP unavailable on test HW | MEDIUM | HIGH | Fallback to UDP; validate in cloud VMs | ✅ Active |
| eBPF compilation issues | MEDIUM | MEDIUM | Pre-compile .o; CI validation | ✅ Planned |
| Latency regression w/ scaling | MEDIUM | MEDIUM | Profile + lock-free data structures | ✅ Planned |
| Lean proof divergence from Go | LOW | MEDIUM | Automated conformance tests; CI gates | ✅ Planned |
| **Schedule slip** | MEDIUM | HIGH | Weekly syncs; escalate blockers by day 1 | ✅ Active |

**Contingency**: If any Phase blocker hits, pivot to alternate approach within 48 hours.

---

## Decision Points

### Decision 1: Start with UDP Overlay (Not Direct AF_XDP)
- **Rationale**: Faster handshake + routing validation; Phase 2 ports to AF_XDP seamlessly
- **Risk**: Extra week on software path
- **Approved**: ✅ YES (expected benefit > risk)

### Decision 2: ML-KEM Stubbed Until Go 1.24
- **Rationale**: crypto/mlkem not available yet; use circl interim for compatibility
- **Risk**: Not true PQC until Go 1.24
- **Approved**: ✅ YES (temporary; drop-in replacement ready)

### Decision 3: Formal Verification in Phase 3 (Not Phase 1)
- **Rationale**: First prove correctness; then prove optimality + invariants
- **Risk**: Formal proofs may reveal design flaws late
- **Approved**: ✅ YES (mitigate with aggressive testing in Phases 1-2)

### Decision 4: Skip Detailed Consensus Implementation (Phase 1-2)
- **Rationale**: Focus on transport core; consensus is separate slow path
- **Risk**: Integration complexity when combining later
- **Approved**: ✅ YES (routing policy engine as proxy; full consensus in integration track)

---

## Dependencies & External Links

### Internal Dependencies
- **Sovereign-Mohawk-Proto** repo: Lean 4 formalization, consensus specs
- **Mohawk Intelligence** track: Federated routing intelligence service

### External Dependencies
- **Go stdlib**: 1.24+ (crypto/mlkem)
- **Kernel**: Linux 5.8+ (AF_XDP), eBPF support
- **Hardware**: 10+ Gbps NIC with driver support (i40e, ixgbe, mlx5, etc.)
- **Third-party libs**: cilium/ebpf, slavc/xdp, cloudflare/circl, uber-go/zap

### Blockers (If Not Resolved)
- ❌ Go 1.24 delayed past June 1 → **escalate**: Switch to Rust AF_XDP engine
- ❌ No compatible NIC found → **escalate**: Use software path for Phase 1-2; defer HW to Phase 3
- ❌ eBPF XDP unavailable → **escalate**: Use hardware RSS only; skip eBPF optimization

---

## Deliverables Checklist

### Phase 1 (Week 2 EOD, May 27)
- [ ] `crypto/kex.go` + `internal/crypto/hybrid.go` complete + tested
- [ ] `cmd/mohawk-node/main.go` entry point
- [ ] `internal/transport/udp.go` UDP overlay forwarding
- [ ] `WEEK1_QUICK_START.md` validated
- [ ] Metrics: Handshake <50ms, Forwarding >1k pps, Latency <5ms

### Phase 2 (Week 4 EOD, Jun 10)
- [ ] AF_XDP kernel-bypass forwarder (descriptor reuse)
- [ ] Multi-queue load balancer
- [ ] eBPF XDP steering program (bpf/xdp_steer.c)
- [ ] Frame pooling + GC optimization
- [ ] Metrics: 3-5 Gbps, <500ns/pkt, linear scaling

### Phase 3 (Month 3 EOD, Aug 5)
- [ ] Lean 4 routing model + loop-freedom proof
- [ ] Dynamic queue scaling (Scaler.go)
- [ ] Comprehensive benchmark suite (scripts/bench.sh)
- [ ] Security audit report
- [ ] Metrics: 10 Gbps, <1ms overhead, zero flaps, zero criticals

### Phase 4 (If Go-Live, Month 4+)
- [ ] Testbed setup (2-3 node cluster)
- [ ] Pilot deployment (canary 5% traffic)
- [ ] Full rollout + 30-day SLA validation

---

## Documentation Artifacts

**Created This Week**:
1. **IMPLEMENTATION_PLAN.md** (detailed task breakdown)
2. **TIMELINE_AND_TRACKER.md** (visual + progress dashboard)
3. **WEEK1_QUICK_START.md** (actionable Week 1 tasks)
4. **This document**: Executive summary

**To Be Created**:
- Weekly status reports (every Monday)
- Benchmark analysis reports (Phases 2-3)
- Security audit report (Phase 3)
- Deployment playbooks (Phase 4)
- Post-mortem & lessons learned (post-go-live)

---

## Communication Plan

### Weekly Sync (Every Monday, 10 AM)
- **Agenda**: Completed tasks, blockers, metrics vs. targets
- **Duration**: 30 min
- **Attendees**: Project lead, tech leads (crypto, networking), perf engineer
- **Async option**: Status doc shared 24h prior for review

### Blocker Escalation
- **<1 day delay**: Team resolves; update TIMELINE_AND_TRACKER.md
- **1-3 day delay**: Escalate to tech lead; pivot plan
- **>3 day delay**: Escalate to project lead; stakeholder notification

### Stakeholder Updates
- **Bi-weekly**: Metrics + milestones (email + dashboard)
- **Monthly**: Full status report (deck + metrics)
- **Critical**: Immediate notification if Phase at risk

---

## Go-Live Decision Gate

**Production Ready When**:
1. ✅ All Phase 3 tasks complete
2. ✅ Metrics: 10 Gbps, <1ms, zero flaps, PQC <50ms
3. ✅ Audit: Zero critical, all high findings remediated
4. ✅ Formal proof: Loop-freedom + convergence validated
5. ✅ Testing: 24-hour burn test, 3 pilot deployments

**Sign-Off Required From**:
- [ ] Project Lead
- [ ] Tech Lead (Crypto)
- [ ] Tech Lead (Networking)
- [ ] Security Officer
- [ ] Formal Methods Lead

---

## Budget & Allocation

### Development (Phases 1-3)
- **Personnel**: 4 FTE × 12 weeks
- **Hardware**: 2× 12.5Gbps test nodes (~$5k)
- **Software**: Go 1.24+, open-source tools (free)
- **Total**: ~$120k (personnel) + $5k (hardware) = **$125k**

### Post-Go-Live (Phase 4)
- **Operational**: Monitoring + incident response + optimizations
- **Support**: Federated learning integration + policy updates
- **Total**: TBD (depends on operational load)

---

## Success = Execution

**Key to Success**:
1. **Discipline**: Weekly tracking; blockers escalated immediately
2. **Focus**: Phase 1 → Phase 2 → Phase 3 (no scope creep)
3. **Validation**: Metrics at each milestone (not just "looks good")
4. **Communication**: Transparent status + early warning on risks
5. **Flexibility**: Willingness to pivot if blocked (Plan B ready)

**This roadmap is achievable in 12 weeks with this team and focus.**

---

## Start: Week 1 (May 14-20)

**Immediately**:
1. Read [WEEK1_QUICK_START.md](WEEK1_QUICK_START.md)
2. Check Go version (`go version` → need 1.24 or use circl)
3. Start Task 1.1.1: ML-KEM integration
4. Daily updates to TIMELINE_AND_TRACKER.md

**Goal**: Handshake working end-to-end by May 20.

---

## Questions?

- **Technical**: See [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md) for task details
- **Progress**: See [TIMELINE_AND_TRACKER.md](TIMELINE_AND_TRACKER.md) for live dashboard
- **Week 1**: See [WEEK1_QUICK_START.md](WEEK1_QUICK_START.md) for daily tasks
- **Architecture**: See [README.md](README.md) for system design

---

**Roadmap Approved**: ✅ Ready to Execute  
**Start Date**: May 14, 2026  
**Target GA**: August 5, 2026  
**Budget**: $125k (development phase)  

**Let's ship this. 🚀**
