# SMIP-MWP

[![CI](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/ci.yml/badge.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/rwilliamspbg-ops/SMIP-MWP)](https://goreportcard.com/report/github.com/rwilliamspbg-ops/SMIP-MWP)
[![GitHub go.mod Go version](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

High-performance, post-quantum-ready sovereign forwarding stack (SMIP / MWP).

This repository is a focused implementation of the data- and control-plane
concepts required to deliver a sovereign transport with hybrid PQC tunnels
and a kernel-bypass forwarding data plane.

Quick Links
- Code: https://github.com/rwilliamspbg-ops/SMIP-MWP
- CI: https://github.com/rwilliamspbg-ops/SMIP-MWP/actions

Quick Start (developer)

1. Build & run the unit tests (default build, no kernel deps):

    go test ./... -v

2. Run the benchmarks locally:

    # run all package benchmarks (crypto and datapath)
    go test ./... -bench . -benchmem -run ^$

Benchmarks included
- internal/crypto/benchmark_test.go — measures zero-copy AEAD EncryptInPlace/DecryptInPlace
- internal/datapath/afxdp/benchmark_loop_test.go — measures the receive->forward loop using in-repo test doubles

You can also use the standardized runner script to capture environment and
timestamped output:

  ./scripts/bench.sh

AF_XDP / Kernel-bypass

The AF_XDP datapath is build-tagged behind `withafxdp`. To build or test with
AF_XDP support you must enable the tag and provide kernel/libbpf support:

    go test ./... -v -tags=withafxdp

See IMPLEMENTATION_PLAN.md for a short hardware/kernel checklist (hugepages,
libbpf, supported NIC drivers).

Metrics & Observability

The example node exposes Prometheus metrics when started with `--metrics-addr`.
Run the node and visit `/metrics` to scrape runtime counters (RX/TX/DROPPED,
crypto errors, per-worker histograms).

Contributing

Please open issues or PRs against the main repository. For design rationale
and long-form architecture notes see the top-level notes and
IMPLEMENTATION_PLAN.md.

License

This project is MIT-licensed. See LICENSE for details.
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
- **Forwarding throughput:** at least **10 Gbps per node** sustained encrypted forwarding baseline (bidirectional aggregate) on commodity hardware (**8+ physical cores, 32 GB RAM, 25 Gbps NIC with kernel-bypass support**). This is a minimum admission threshold for production rollout; operational target utilization is **>= 80% of available line rate** after hardening.
- **Latency overhead:** less than **1 ms added latency** versus plain UDP in LAN/WAN test profiles
- **Handshake performance:** full **PQC hybrid handshake under 50 ms** in controlled benchmark profiles (LAN and low-RTT WAN emulation), including key exchange + signature verification + session key confirmation, validated against baseline CPU crypto acceleration capabilities (AVX2 minimum).
  - Reference budget split (validation baseline): key exchange ≤ 18 ms, signature verification ≤ 12 ms, session confirmation ≤ 10 ms, measured as sequential upper-bound accounting; remaining budget covers coordination + low-RTT transport overhead
  - 0-RTT/1-RTT resumed handshakes are tracked separately with stricter targets
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
- Conformance tests must validate runtime implementation behavior against verified models for convergence, failover, and loop-freedom under topology changes; any observed loop, sovereignty/policy violation, or convergence divergence beyond timeout bounds (**<= 5 s for single-link failure, <= 15 s for multi-link failure scenarios**) is a conformance failure

### 7) Cryptography and Identity

- Hybrid key exchange: **x25519 + ML-KEM-768**, combined through a two-stage HKDF combiner with domain separation: `prk_classical = HKDF-Extract(salt, x25519_shared_secret)`, `prk_pqc = HKDF-Extract(salt, mlkem768_shared_secret)`, `hybrid_prk = HKDF-Extract(salt, prk_classical || prk_pqc)`, final session keys from `HKDF-Expand(hybrid_prk, context_info, L)`
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

## Status & CI Badges

- CI: [![CI](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions/workflows/ci.yml/badge.svg)](https://github.com/rwilliamspbg-ops/SMIP-MWP/actions)
- Go tests: `go test ./...` (see CI)

## Runtime Capabilities

This repository includes:

- A tested, unit-safe AF_XDP forwarder stub (default build) and a build-tagged
  AF_XDP integration (`withafxdp`) using the `github.com/asavie/xdp` bindings.
- A deterministic hybrid PQC session (`x25519 + ML-KEM stub`) with HKDF key
  derivation and AEAD encryption (AES-GCM preferred, ChaCha20-Poly1305 fallback).
- A zero-allocation `HeaderView` and copy-structured `Header` helpers for
  parsing and marshaling packet headers used in the fast path.
- A testable receive->steer->transmit packet loop with a Prometheus metrics
  integration exposing runtime counters for RX/TX/DROPPED/handshakes/crypto-errors.

## Metrics & Observability

The node exposes Prometheus metrics on an HTTP endpoint when started with the
`--metrics-addr` flag (default `:9090`). Metrics include:

- `smip_mwp_afxdp_rx_packets_total`
- `smip_mwp_afxdp_tx_packets_total`
- `smip_mwp_afxdp_dropped_packets_total`
- `smip_mwp_crypto_handshakes_total`
- `smip_mwp_crypto_errors_total`

Enable metrics in the example node: `go run ./cmd/mohawk-node --metrics-addr=:9090`.

## Operating Systems & Kernel Roadmap

AF_XDP and high-performance dataplane features require specific kernel and
userspace support. Roadmap and recommendations:

- Development / Test: Ubuntu LTS (22.04 / 24.04) or Fedora latest — install
  `libbpf-dev`, `clang`, `llvm`, `libelf-dev`, and configure hugepages.
- Production: recent Linux kernels (5.10+ recommended; 6.x preferred) with
  XDP/BPF enhancements. NIC drivers supporting XDP native mode produce the
  best throughput.
- Container/Orchestration: use privileged containers with `SYS_ADMIN` or
  hostNetwork + CAP_NET_RAW for attaching XDP programs; consider privileged
  sidecars for eBPF management.

## Developer Quick Start

1. Build & run tests (default, no AF_XDP):

```bash
go test ./... -v
```

2. To build with AF_XDP support (requires kernel & deps):

```bash
go mod tidy
go test ./... -v -tags=withafxdp
# or run the node
go run -tags=withafxdp ./cmd/mohawk-node --iface=eth0 --metrics-addr=:9090
```

If you plan to run with AF_XDP, ensure you have the privileges and kernel
support described above.


Planned execution proceeds from architecture/specification to benchmarked implementation milestones for Mohawk Forge and Mohawk Intelligence.
