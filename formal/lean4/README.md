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

How to contribute proofs

- Add Lean files under `Lean/Smip/`.
- Write small, incremental lemmas and run `lake build` frequently.
- Use `#check` and `#eval` in files for quick experiments in the VS Code Lean REPL.

<!-- CI retrigger: minor edit to trigger workflow -->
