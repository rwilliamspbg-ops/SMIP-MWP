# Lean Formalization Checklist

This checklist tracks the Phase 3 and Phase 4 formalization work for SMIP-MWP.
It is scoped to the current Lean model in [Lean/Smip/Phase3.lean](Lean/Smip/Phase3.lean) and the Go modules it is intended to mirror.

## Current Status

- [x] Lean project scaffold exists and builds as a standalone Lake project.
- [x] Current Lean file parses cleanly with no reported syntax errors.
- [x] Current model inventory completed for routing, crypto, wire, AF_XDP, and CLI Go surfaces.
- [x] Remove the remaining proof hole in the Fin-based fmap lemma.
- [x] Add a Lean routing-policy spec for exact lookup and default policy seeding.
- [x] Add a Lean wire-header spec for layout and marshal/parse round-trips.
- [x] Add a Lean route-table spec for RouteEntry, exact lookup, and fallback lookup shape.
- [x] Add a Lean crypto constants spec for HKDF cache and handshake lengths.
- [x] Add a Lean AF_XDP lifecycle spec for shard bounds and start/stop behavior.
- [ ] Replace the routing skeleton with a model that matches the Go routing table and policy lookup behavior.
- [ ] Add explicit formal specs for the wire header layout and marshal/parse invariants.
- [ ] Add formal specs for crypto handshake, HKDF cache bounds, and lifecycle cleanup.
- [ ] Add formal specs for AF_XDP session sharding, run/stop behavior, and queue assignment.
- [x] Add CI enforcement so Lean builds fail on any `sorry`, `admit`, or `axiom`.
- [ ] Add conformance tests that compare Go behavior against Lean-specified cases.
	- Next proof: prove `lookupOrPredict_hash` equals the Go control-flow model for table representations (parity theorem) (requires modeling SHA256 or equivalence witness).
	- Progress: discovered a concrete counterexample where `predictiveIndex_sha256 ≠ predictiveIndex` (example: `src=1,dst=2,flow=3,len=5`).
	- Implication: cannot prove general equality without aligning `combHash` to the SHA256 model or restricting the input domain. Options: (1) set `combHash := sha256_model_be32` for spec parity, (2) use conformance vectors, (3) port verified SHA256 for full fidelity.
	- Action taken: aligned spec by setting `combHash := sha256_model_be32` so `predictiveIndex` reflects the modeled SHA256 extraction.
	- Result: the parity obligation reduces to checking the SHA model faithfully reflects Go's SHA256 first-4-bytes for test vectors (conformance), or porting verified SHA256 for full proof.
	- Generated a sample set of real SHA256 vectors: `formal/lean4/SmipSha256Vectors_sample.csv`.

## Module Coverage

### Routing

- [x] Add [Lean/Smip/RoutingSpec.lean](Lean/Smip/RoutingSpec.lean) for default policy and exact-match lookup.
 - [x] Add [Lean/Smip/RoutingSpec.lean](Lean/Smip/RoutingSpec.lean) for default policy and exact-match lookup.
 - [x] Add `lookupPolicyOrDefault` and priority lemmas to model Router default-fallback semantics.
- [x] Add [Lean/Smip/RouterModel.lean](Lean/Smip/RouterModel.lean) for route entries and lookup/update behavior.
- [ ] Formalize [internal/routing/router.go](../../internal/routing/router.go) RouteEntry fields.
- [ ] Formalize [internal/routing/router.go](../../internal/routing/router.go) exact lookup behavior.
- [ ] Formalize [internal/routing/router.go](../../internal/routing/router.go) predictive fallback selection.
 - [ ] Formalize [internal/routing/router.go](../../internal/routing/router.go) predictive fallback selection.
	- Next: model `PredictiveNextHop`'s SHA256-based index selection and prove `LookupOrPredict` parity.
	- Progress: added `predictiveIndex`, `predictiveNextHopByIndex`, and `lookupOrPredict_hash` in `Lean/Smip/RouterModel.lean`.
	- Progress: proved parity lemma for the case `predictiveIndex = 0` (`lookupOrPredict_index_zero_eq_first`).
	- Next proof: prove `lookupOrPredict_hash` equals the Go control-flow model for table representations (parity theorem) (requires modeling SHA256 or equivalence witness).
	- Progress: added a constructive SHA256 model `sha256_model_be32` and `predictiveIndex_sha256` in `Lean/Smip/RouterModel.lean`.
	- Progress: added a constructive SHA256 model `sha256_model_be32` and `predictiveIndex_sha256` in `Lean/Smip/RouterModel.lean`.
	- Progress: added `lookupOrPredict_sha256` and a parity lemma `lookupOrPredict_sha256_parity` reducing general parity to an index-equivalence obligation.
	- Progress: proved index-equivalence `predictiveIndex_eq_sha256` (definitions now align via `combHash := sha256_model_be32`).
	- Result: `lookupOrPredict_sha256_parity` now holds by reflexivity when combined with `predictiveIndex_eq_sha256`.
- [ ] Formalize [internal/routing/router.go](../../internal/routing/router.go) policy priority/update invariants.
- [ ] Formalize [internal/routing/router_enhanced.go](../../internal/routing/router_enhanced.go) additive wrapper behavior.

### Wire Format

- [x] Add [Lean/Smip/WireSpec.lean](Lean/Smip/WireSpec.lean) for header size, offsets, and round-trip structure.
- [ ] Formalize [internal/wire/header.go](../../internal/wire/header.go) header size and field offsets.
- [ ] Formalize [internal/wire/header.go](../../internal/wire/header.go) Marshal/Parse round-trip invariants.
- [ ] Formalize [internal/wire/header.go](../../internal/wire/header.go) zero-copy view/setter invariants.

### Crypto

- [x] Add [Lean/Smip/CryptoSpec.lean](Lean/Smip/CryptoSpec.lean) for HKDF cache and handshake constants.
- [ ] Formalize [internal/crypto/hybrid.go](../../internal/crypto/hybrid.go) HKDF cache bounds.
- [ ] Formalize [internal/crypto/hybrid.go](../../internal/crypto/hybrid.go) cache key derivation and lookup determinism.
- [ ] Formalize [kex.go](../../kex.go) session-secret derivation and length checks.
- [ ] Formalize [internal/crypto/handshake.go](../../internal/crypto/handshake.go) cleanup/state lifecycle behavior.

### AF_XDP

- [x] Add [Lean/Smip/AFXDPSpec.lean](Lean/Smip/AFXDPSpec.lean) for session sharding and forwarder lifecycle.
- [ ] Formalize [internal/datapath/afxdp/forwarder.go](../../internal/datapath/afxdp/forwarder.go) run/stop lifecycle.
- [ ] Formalize [internal/datapath/afxdp/forwarder.go](../../internal/datapath/afxdp/forwarder.go) session shard soundness.
- [ ] Formalize [internal/datapath/afxdp/forwarder.go](../../internal/datapath/afxdp/forwarder.go) queue and worker assignment.
- [ ] Formalize [internal/datapath/afxdp/worker_pool.go](../../internal/datapath/afxdp/worker_pool.go) buffer ownership and reuse invariants.
- [ ] Formalize [internal/datapath/afxdp/metrics.go](../../internal/datapath/afxdp/metrics.go) metrics label stability.

### Verification Gates

- [ ] Add Lean CI that runs `lake build` for [formal/lean4](.).
- [ ] Add a repository check that rejects new `sorry`, `admit`, and `axiom` usage.
- [ ] Add Go-to-Lean conformance tests for the formalized surfaces.
- [ ] Add a release gate that requires the Lean checklist to be fully checked before merge.

## Completion Notes

- The current Lean file is a routing-focused skeleton, not yet a full mirror of the Go implementation.
- The first concrete proof debt is the `sorry` in the Fin-indexed fmap lemma.
- The rest of the plan should be implemented incrementally, starting with routing and wire formats before expanding to crypto and AF_XDP internals.

## Execution Order

1. Finish routing parity in Lean, starting from [internal/routing/router.go](../../internal/routing/router.go).
2. Add wire-format invariants for [internal/wire/header.go](../../internal/wire/header.go).
3. Add crypto/session formal specs for [internal/crypto/hybrid.go](../../internal/crypto/hybrid.go) and [kex.go](../../kex.go).
4. Add AF_XDP lifecycle and sharding specs for [internal/datapath/afxdp/forwarder.go](../../internal/datapath/afxdp/forwarder.go).
5. Expand conformance tests from the formalized Lean properties into Go test cases.
6. Keep Lean CI locked to no `sorry`, `admit`, or `axiom` usage.