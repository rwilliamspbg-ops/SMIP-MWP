# Performance Summary

This file provides an executive summary of measured performance artifacts and guidance for interpreting benchmark outputs produced by the project's benchmark runner and CI.

Where artifacts live

- Raw and processed benchmark outputs are in `benchmarks/` with timestamps and pprof profiles when enabled. Use those artifacts for reproducible analysis.

Interpreting results

- Throughput: look for tx/rx rates printed by the runner or extracted from pprof traces.
- Latency: per-worker histograms and Prometheus summaries are captured by the benchmark harness where enabled.
- Pprof: use `go tool pprof` to load CPU and memory profiles from `benchmarks/*-cpu.prof` and `*-mem.prof`.

Example pprof usage

```bash
go tool pprof -http=:8080 benchmarks/bench-codespaces-*-cpu.prof
```

Reporting

- For CI-driven benchmark runs, reference the workflow run that generated the artifacts and include the run ID/timestamp when reporting performance.
