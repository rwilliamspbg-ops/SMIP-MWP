# SMIP-MWP — Lean 4 Formalization (Phase 3)

This folder contains a starter Lake/Lean 4 project for formal proofs related to Phase 3 (routing correctness, loop-freedom, queue scaling properties).

Prerequisites

- Install `elan` (Lean toolchain installer) — see https://leanprover.github.io/installation/.
- Install `lake` (project manager): `elan toolchain install leanprover/lean4:stable` (usually included) and `lake` is available via the Lean toolchain.
- Recommended editor: VS Code with the `lean4` extension.

Quickstart

```bash
# from repo root
cd formal/lean4
# Ensure lean toolchain is installed (elan)
lake --version
lake build

# Open VS Code in the folder for interactive development
code -r .
```

Project layout

- `lakefile.lean` — Lake project file
- `Lean/Smip/Phase3.lean` — proof skeletons and TODO theorems
- `Lean/Smip/RoutingSpec.lean` — Lean spec for the router policy table and exact-match lookup behavior
- `Lean/Smip/RouterModel.lean` — Lean spec for routing entries, exact lookup, update, and fallback existence
- `Lean/Smip/CryptoSpec.lean` — Lean spec for HKDF cache bounds and hybrid handshake length constants
- `Lean/Smip/AFXDPSpec.lean` — Lean spec for session sharding and forwarder start/stop behavior
- `Lean/Smip/WireSpec.lean` — Lean spec for wire header layout and round-trip invariants
- `LEAN_FORMALIZATION_CHECKLIST.md` — tracked checklist for routing, crypto, wire, and AF_XDP formalization work

How to contribute proofs

- Add Lean files under `Lean/Smip/`.
- Write small, incremental lemmas and run `lake build` frequently.
- Use `#check` and `#eval` in files for quick experiments in the VS Code Lean REPL.

<!-- CI retrigger: minor edit to trigger workflow -->
<!-- CI retrigger: second touch to pick up workflow changes -->
<!-- CI retrigger: third touch to pick up latest workflow update -->
