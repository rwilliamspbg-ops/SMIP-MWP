# SMIP-MWP

Sovereign Mohawk Internet Protocol (SMIP), also referred to as Mohawk Wire Protocol (MWP), is a next-generation sovereign transport and routing stack derived from the original Sovereign Mohawk Protocol.

This repository now tracks the production-oriented SMIP/MWP re-architecture as an executable program of work across data, control, security, and management planes.

## Core Outcome

Deliver a production-ready Sovereign Internet Protocol stack that is:
- Competitive with or faster than modern QUIC/TCP for latency/throughput
- Natively post-quantum secure
- Formally verified for routing stability and policy invariants
- AI-driven for predictive, sovereignty-aware routing
- Deployable as either:
  - an overlay over existing IP infrastructure, or
  - a clean-slate sovereign backbone

## Program Structure

### Core Working Group

SMIP delivery is organized around a single cross-functional core team:
- Protocol architects
- Networking and packet-processing engineers
- Formal verification (Lean 4) specialists
- Performance engineers and benchmarking owners

### Codebase Split

The project model is split into two coordinated tracks:
- **Mohawk Intelligence**: original federated intelligence and distributed learning capabilities
- **Mohawk Forge**: high-performance transport, routing, and wire-protocol implementation

## Formal Success Metrics

SMIP is considered production-ready only when all baseline targets are met:
- **Forwarding throughput:** at least **10 Gbps per node** sustained encrypted forwarding baseline (bidirectional aggregate) on commodity hardware (**8+ physical cores, 32 GB RAM, 25 GbE NIC with kernel-bypass support**), with higher line-rate utilization treated as a stretch target
- **Latency overhead:** less than **1 ms added latency** versus plain UDP in LAN/WAN test profiles
- **Handshake performance:** full **PQC hybrid handshake under 50 ms** in controlled benchmark profiles (LAN and low-RTT WAN emulation), including key exchange + signature verification + session key confirmation, validated against baseline CPU crypto acceleration capabilities (AVX2 minimum); reference budget split: key exchange ≤ 20 ms, signature verification ≤ 15 ms, session confirmation ≤ 15 ms. 0-RTT/1-RTT resumed handshakes are tracked separately with stricter targets
- **Routing stability:** **zero route flaps** across formally verified topologies

## Reference Architecture

### 1) Data Plane (Fast Path)

Implemented in Go (or Rust where memory-safety benefits are required), with lightweight node agents optimized for kernel-bypass I/O:
- DPDK
- AF_XDP
- io_uring

Responsibilities:
- Plain encrypted packet forwarding via PQC-hybrid tunnels
- Minimal per-packet overhead
- Support for datagram and reliable stream transports

### 2) Control + Security Slow Path (Asynchronous)

Heavy operations are intentionally moved off the packet fast path:
- Epoch-based consensus
- zk-SNARK identity/route attestations
- Ledger updates and heavy verification

This separation preserves wire-speed forwarding while maintaining strong governance and trust guarantees.

### 3) Wire Format

Lightweight header derived from existing Mohawk message structure:
- Source sovereign cryptographic ID
- Destination sovereign cryptographic ID
- Flow label
- Sequence number
- PQC ephemeral key material or session ID
- Optional integrity/attestation tags

### 4) Transport Modes

- **Datagram mode:** unreliable, high-throughput forwarding
- **Reliable mode:** Mohawk-TCP-like reliability with QUIC-inspired multiplexing and congestion control

### 5) Distributed Network Intelligence

Federated learning agents are repurposed for routing intelligence:
- Continuous local and peer condition modeling (congestion, jitter, trust, bandwidth)
- Lightweight federated model updates
- Predictive next-hop selection and proactive rerouting around low-sovereignty or congested paths

### 6) Sovereign Route Attestation

Traditional routing tables are replaced/augmented with formally provable route advertisements:
- Lean 4 proofs for loop-freedom
- Policy and sovereignty compliance guarantees
- Migration compatibility with existing BGP domains through tunnel interop
- Conformance tests must validate runtime implementation behavior against verified models for convergence, failover, and loop-freedom under topology changes; any observed loop, sovereignty/policy violation, or convergence divergence beyond defined timeout bounds is a conformance failure

### 7) Cryptography and Identity

- Hybrid key exchange: **x25519 + ML-KEM-768**, combined via HKDF over concatenated shared secrets (`x25519_shared_secret || mlkem768_shared_secret`) for interoperable key schedule derivation
- Signature scheme baseline: **ML-DSA-65** (with policy-driven upgrade path to stronger parameter sets where required)
- Perfect forward secrecy for sovereign tunnel sessions
- zk-SNARKs used periodically for identity and aggregate route attestation, not per-packet

## Stack Layers

- **Data Plane:** Go/Rust packet-forwarding agents with PQC-encrypted transport
- **Control Plane:** federated intelligence + formally verified route advertisement
- **Security Plane:** hybrid PQC handshakes + zk-SNARK attestations + sovereign identity
- **Management Plane:** Mohawk governance and consensus for policy lifecycle

## Deployment Model

SMIP supports phased deployment:
1. Overlay mode on existing IP networks
2. Hybrid sovereign tunnels across legacy domains
3. Full sovereign clean-slate backbone segments

## Repository Scope (Current)

This repository currently serves as the canonical SMIP/MWP architecture baseline and implementation contract.

Implementation source foundation:
- https://github.com/rwilliamspbg-ops/Sovereign-Mohawk-Proto

Planned execution proceeds from architecture/specification to benchmarked implementation milestones for Mohawk Forge and Mohawk Intelligence.
