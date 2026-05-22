# Contributing to SMIP-MWP

SMIP-MWP is a formally verified, performance-sensitive protocol repository. Contributions are welcome when they improve correctness, performance, observability, or documentation without weakening the verification or benchmark story.

## Before you open a pull request

1. Work from a clean branch based on `main`.
2. Read the relevant docs first: [README.md](README.md), [docs/USAGE.md](docs/USAGE.md), and [docs/PERFORMANCE.md](docs/PERFORMANCE.md).
3. If your change touches AF_XDP or generated artifacts, run the relevant validation script before asking for review.
4. Keep PRs focused. One behavioral change per PR is much easier to review and verify than a mixed change set.

## Required checks

At minimum, run the checks that apply to your change:

```bash
go test ./... -v
go vet ./...
```

For AF_XDP-related changes:

```bash
./scripts/test_xdp.sh --run-go-test
go test ./... -tags=withafxdp -run '^$'
```

For performance-sensitive changes:

```bash
./scripts/bench.sh --pprof
```

If your change affects generated files, make sure the generated outputs and source inputs still match. The repository includes a dedicated workflow for that check.

## Pull request expectations

1. Explain the problem being solved and the user-visible impact.
2. Call out any protocol, performance, or verification consequences explicitly.
3. Include the validation commands you ran and summarize the results.
4. If the change is performance-related, include before/after numbers from the same benchmark path.
5. If the change touches formalization or generated code, link the relevant files and describe how you kept the artifacts in sync.

## Review checklist

Reviewers will usually look for these things:

- Correctness under normal and edge-case traffic
- No regression in benchmark artifacts or hot-path allocations
- Clear separation between protocol logic, datapath logic, and verification artifacts
- No drift between generated content and the source of truth

## Commit and PR formatting

- Use short, descriptive commit messages.
- Prefer a PR title that states the behavioral change.
- Add screenshots, benchmark snippets, or profile links when they help reviewers verify the outcome.

## When to ask for help

If you are unsure whether a change belongs in the fast path, the control plane, or the formal model, open a draft PR early and describe the uncertainty. That usually gets you a faster and more precise review than waiting for a large final patch.