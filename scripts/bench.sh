#!/usr/bin/env bash
set -euo pipefail

# scripts/bench.sh
# Standardized benchmark runner for SMIP-MWP
# Usage:
#   ./scripts/bench.sh                 # run default full repo benchmarks
#   ./scripts/bench.sh -- ./pkg -bench "^BenchmarkFoo$"  # pass extra args to go test

OUT_DIR="benchmarks"
mkdir -p "${OUT_DIR}"

TS=$(date -u +"%Y%m%dT%H%M%SZ")
HOST=$(hostname -s 2>/dev/null || echo "localhost")
OUT_FILE="${OUT_DIR}/bench-${HOST}-${TS}.txt"

echo "SMIP-MWP benchmark run: ${TS}" > "${OUT_FILE}"
echo "Command line: $*" >> "${OUT_FILE}"
echo >> "${OUT_FILE}"

echo "=== go env ===" >> "${OUT_FILE}"
go env >> "${OUT_FILE}" 2>&1 || true
echo >> "${OUT_FILE}"

echo "=== system info ===" >> "${OUT_FILE}"
uname -a >> "${OUT_FILE}" 2>&1 || true
lscpu >> "${OUT_FILE}" 2>&1 || true
echo >> "${OUT_FILE}"

echo "=== benchmarks ===" >> "${OUT_FILE}"

# Default benchmark command
DEFAULT_CMD=(go test ./... -bench . -benchmem -run ^$ -count=1)

if [ "$#" -gt 0 ]; then
  # If the caller provides arguments, use them as the go test args
  echo "Running custom command: $*" | tee -a "${OUT_FILE}"
  (set -x; "$@") 2>&1 | tee -a "${OUT_FILE}"
else
  echo "Running default: ${DEFAULT_CMD[*]}" | tee -a "${OUT_FILE}"
  (set -x; "${DEFAULT_CMD[@]}") 2>&1 | tee -a "${OUT_FILE}"
fi

echo >> "${OUT_FILE}"
echo "Benchmark run complete. Output: ${OUT_FILE}"

exit 0
