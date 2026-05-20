#!/usr/bin/env bash
# perf-quick.sh - Quick performance testing script for SMIP-MWP
# Use this for fast iteration and baseline measurements

set -euo pipefail

# Configuration (override with environment variables)
COUNT="${COUNT:-1}"              # Number of benchmark iterations
BENCH_TARGET="${BENCH_TARGET:-.}" # Benchmark target (default: all packages)
OUTPUT_DIR="${OUTPUT_DIR:-./benchmarks/quick}"
PROFILE="${PROFILE:-false}"      # Enable CPU/memory profiling

mkdir -p "${OUTPUT_DIR}"

TS=$(date -u +"%Y%m%dT%H%M%SZ")
OUT_FILE="${OUTPUT_DIR}/quick-${TS}.txt"

echo "==========================================" > "${OUT_FILE}"
echo "SMIP-MWP Quick Performance Test" >> "${OUT_FILE}"
echo "==========================================" >> "${OUT_FILE}"
echo "Timestamp: ${TS}" >> "${OUT_FILE}"
echo "Benchmark Count: ${COUNT}" >> "${OUT_FILE}"
echo "Target: ${BENCH_TARGET}" >> "${OUT_FILE}"
echo "Profile Mode: ${PROFILE}" >> "${OUT_FILE}"
echo "" >> "${OUT_FILE}"

# System info
echo "=== Environment ===" >> "${OUT_FILE}"
go env | sed 's/^/  /' >> "${OUT_FILE}"
echo "" >> "${OUT_FILE}"

echo "=== System Info ===" >> "${OUT_FILE}"
uname -a >> "${OUT_FILE}" 2>&1 || true
lscpu 2>/dev/null | head -15 >> "${OUT_FILE}" || echo "N/A" >> "${OUT_FILE}"
free -h >> "${OUT_FILE}" 2>&1 || true
echo "" >> "${OUT_FILE}"

# Run benchmarks
echo "=== Running Benchmarks ===" >> "${OUT_FILE}"

# Build command based on profile mode
if [ "${PROFILE}" = "true" ]; then
    PROF_ARGS="-cpuprofile=${OUTPUT_DIR}/quick-${TS}-cpu.prof -memprofile=${OUTPUT_DIR}/quick-${TS}-mem.prof"
else
    PROF_ARGS=""
fi

CMD=(go test -bench=. -benchmem -run="^$" -count="${COUNT}")
if [ -n "${PROF_ARGS}" ]; then
    CMD+=("${PROF_ARGS}")
fi

# Add benchmark target
CMD+=("${BENCH_TARGET}")

# Execute
echo "Command: go test ${CMD[*]}" >> "${OUT_FILE}"
"${CMD[@]}" 2>&1 | tee -a "${OUT_FILE}" || true

echo "" >> "${OUT_FILE}"
echo "==========================================" >> "${OUT_FILE}"
echo "Quick Performance Test Complete" >> "${OUT_FILE}"
echo "==========================================" >> "${OUT_FILE}"

# Check for benchmark results
if grep -q "Benchmark" "${OUT_FILE}"; then
    echo "" >> "${OUT_FILE}"
    echo "=== Summary ===" >> "${OUT_FILE}"
    
    # Extract key metrics
    BENCH_OUTPUT=$(grep -E "^Benchmark[A-Za-z].*ns/op|ns/op\s+[0-9.]+M" "${OUT_FILE}" | head -30)
    echo "${BENCH_OUTPUT}" >> "${OUT_FILE}"
else
    echo "No benchmark results found in output." >> "${OUT_FILE}"
fi

# Generate summary
SUMMARY="${OUTPUT_DIR}/quick-${TS}-summary.txt"
cat > "${SUMMARY}" << EOF
Quick Performance Test Summary
===============================
Timestamp: ${TS}
Benchmark Count: ${COUNT}
Target: ${BENCH_TARGET}
Profile: ${PROFILE}

Output File: ${OUT_FILE}
CPU Profile: ${OUTPUT_DIR}/quick-${TS}-cpu.prof${PROFILE:+ (generated)}
Memory Profile: ${OUTPUT_DIR}/quick-${TS}-mem.prof${PROFILE:+ (generated)}

Key Steps:
1. Review output in: ${OUT_FILE}
2. View top functions: go tool pprof -top ${OUTPUT_DIR}/quick-${TS}-cpu.prof
3. Compare with previous runs for regression detection

Regression Alert: >5% increase in ns/op suggests optimization needed
EOF

echo ""
cat "${SUMMARY}"
echo ""
echo "Quick performance test complete!"
