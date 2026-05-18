Recent documentation updates

- Date: 2026-05-18
- Summary: Added measured baseline and recommended canonical configuration for hardware validation to `docs/PERFORMANCE.md` and updated `ROADMAP_EXECUTIVE_SUMMARY.md` with a "Recent Performance Update" and recommendations. Small README tweaks included.
- Relevant commit: 972cbf7 (pushed to `main`).

Purpose of this PR

- Provide a small, reviewable PR that documents the edits already pushed to `main`, so maintainers can review and optionally revert or adjust the pushed changes.

Suggested review points

- Verify the reported benchmark numbers and phrasing in `docs/PERFORMANCE.md`.
- Confirm the recommended canonical configuration for hardware tests: `CRYPTO_WORKERS=1`, `CRYPTO_BATCH_SIZE=4`, and the Ansible/MoonGen helper usage.

Next steps after merging

- Run hardware validation with the canonical configuration and collect new `benchmarks/*-cpu.prof` artifacts.

Release: merged `formal/lean4-init` (PR #10)

- Date: 2026-05-18
- Merge commit: 3ddae2bc7b2540ec46fe1651ef363c852224e401
- Summary: Squash-merged PR #10 which added the Lean 4 lake project scaffold and `Phase3.lean` formalization skeleton, Fin-indexed reachability proofs, FMap modeling and invariants, and documentation updates. `lake build` completes successfully after changes.

Notes

- The branch was rebased onto `main` and conflicts in `docs/PERFORMANCE.md` were resolved during the merge.
- If you maintain a release log, consider adding this merge under the "Formalization" section.
