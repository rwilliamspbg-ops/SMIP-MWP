Benchmarks and CI policy
=========================

This folder documents the project's benchmark runner, how to interpret results,
and recommended CI policy for collecting repeatable benchmark artifacts.

Runner
------
Use the provided script to run benchmarks and capture environment details and
profiles.

    ./scripts/bench.sh            # run full repo benchmarks (default)
    ./scripts/bench.sh --pprof    # run default and collect cpu/mem profiles
    ./scripts/bench.sh -- go test ./internal/crypto -bench . -benchmem -run ^$ -count=1 --pprof

The script writes timestamped output into the `benchmarks/` directory. When
`--pprof` is used the script automatically appends `-cpuprofile` and
`-memprofile` flags to `go test` invocations and places the profiles in the
benchmarks directory.

Interpreting results
--------------------
- `ns/op` and `B/op` show per-iteration time and allocation impact. Look for
  regressions across runs rather than absolute values.
- `allocs/op` indicates allocation count; zero/allocation-free hot paths are
  ideal on the AF_XDP fast path.
- CPU profiles (pprof) can be inspected locally with:

    go tool pprof -http=:8080 benchmarks/bench-<host>-<ts>-cpu.prof

  This launches an interactive UI for flame graphs and top consumers.

CI policy (recommended)
-----------------------
1. Keep benchmark CI short: run a focused, small set of microbenchmarks on a
   schedule (e.g., weekly) rather than on every push. Use a dedicated runner
   with stable CPU characteristics if possible.
2. Upload artifacts (text output + pprof files) and retain them for a fixed
   window (e.g., 14 days) to allow regression triage.
3. Record the runner's hardware metadata (go env, uname, lscpu). Always
   compare runs from identical runner types.
4. Use statistical smoothing / thresholding for alerts. Because benchmarks
   have noise, consider percent-change thresholds (e.g., 5% sustained
   regression across three consecutive runs) before opening automated alerts.

Benchmark hygiene
-----------------
- Use `-count=1` for true raw run values; repeated runs may be useful to
  compute medians but can also mask transient noise.
- When adding new benchmarks, keep them small and focused: one hot-path per
  benchmark. Avoid expensive system-wide operations in unit-level benchmarks.

If you want, I can add a scheduled GitHub Actions job that also posts a short
summary to a Slack channel or PR comment when a regression is detected.
