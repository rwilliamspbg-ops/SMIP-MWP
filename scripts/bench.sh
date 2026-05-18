#!/usr/bin/env bash
set -euo pipefail

# scripts/bench.sh
# Standardized benchmark runner for SMIP-MWP
# Usage:
#   ./scripts/bench.sh                 # run default full repo benchmarks
#   ./scripts/bench.sh -- ./pkg -bench "^BenchmarkFoo$"  # pass extra args to go test

OUT_DIR="benchmarks"
mkdir -p "${OUT_DIR}"

TS=$(date -u +"%Y%m%dT%H%M%S%N")
HOST=$(hostname -s 2>/dev/null || echo "localhost")
# include PID to avoid filename collisions across very fast successive runs
OUT_FILE="${OUT_DIR}/bench-${HOST}-${TS}-$${$}.txt"

PROFILE=0
if [ "${1-}" = "--pprof" ]; then
  PROFILE=1
  shift
fi

# Allow a leading `--` separator so callers can use:
#   ./scripts/bench.sh -- go test ./pkg -bench ...
# This keeps backward compatibility with callers that pass `--`.
if [ "${1-}" = "--" ]; then
  shift
fi

echo "SMIP-MWP benchmark run: ${TS}" > "${OUT_FILE}"
echo "Profile enabled: ${PROFILE}" >> "${OUT_FILE}"
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

# Helper to add profiling args when requested
add_profiling_args() {
  local -n arr=$1
  # only add if not already present
  local hasCPU=0 hasMEM=0
  for a in "${arr[@]}"; do
    case "$a" in
      -cpuprofile=*) hasCPU=1 ;;
      -memprofile=*) hasMEM=1 ;;
    esac
  done
  if [ "$PROFILE" -eq 1 ]; then
    if [ $hasCPU -eq 0 ]; then
      arr+=("-cpuprofile=${OUT_DIR}/bench-${HOST}-${TS}-cpu.prof")
    fi
    if [ $hasMEM -eq 0 ]; then
      arr+=("-memprofile=${OUT_DIR}/bench-${HOST}-${TS}-mem.prof")
    fi
  fi
}

if [ "$#" -gt 0 ]; then
  # If the caller provides arguments, treat them as the go test command.
  CMD=("$@")
  # If profiling requested and the command is 'go test', append profiling args
  if [ "$PROFILE" -eq 1 ] && [ "${CMD[0]}" = "go" ]; then
    # append profiling to the 'go test' argument list
    # find the index of 'test'
    for i in "${!CMD[@]}"; do
      if [ "${CMD[$i]}" = "test" ]; then
        # split into prefix and args
        prefix=("${CMD[@]:0:$((i+1))}")
        args=("${CMD[@]:$((i+1))}")
        add_profiling_args args
        CMD=("${prefix[@]}" "${args[@]}")
        break
      fi
    done
  fi
  echo "Running custom command: ${CMD[*]}" | tee -a "${OUT_FILE}"
  (set -x; "${CMD[@]}") 2>&1 | tee -a "${OUT_FILE}"
else
  CMD=("${DEFAULT_CMD[@]}")
  add_profiling_args CMD
  echo "Running default: ${CMD[*]}" | tee -a "${OUT_FILE}"
  (set -x; "${CMD[@]}") 2>&1 | tee -a "${OUT_FILE}"
fi

echo >> "${OUT_FILE}"
echo "Benchmark run complete. Output: ${OUT_FILE}" | tee -a "${OUT_FILE}"

exit 0
