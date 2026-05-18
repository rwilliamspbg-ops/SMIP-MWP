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
