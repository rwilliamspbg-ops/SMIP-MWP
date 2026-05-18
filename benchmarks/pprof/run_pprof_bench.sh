#!/usr/bin/env bash
set -euo pipefail

# Helper to run the AF_XDP benchmark with pprof CPU and memory profiles
# Usage: ./run_pprof_bench.sh [duration_seconds]
DUR=${1:-60}
OUTDIR=${OUTDIR:-$(pwd)/benchmarks/pprof}
mkdir -p "$OUTDIR"

# Run bench for internal/datapath/afxdp
go test -bench BenchmarkRunXDPLoop_WithCrypto -run '^$' -benchtime=${DUR}s -benchmem -cpuprofile=${OUTDIR}/afxdp-cpu.prof -memprofile=${OUTDIR}/afxdp-mem.prof ./internal/datapath/afxdp

# Example pprof invocation (local machine)
# go tool pprof -http=:8080 ./afxdp.test ${OUTDIR}/afxdp-cpu.prof

echo "Profiles written to ${OUTDIR}" 
