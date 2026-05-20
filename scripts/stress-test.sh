#!/usr/bin/env bash
# stress-test.sh - High-load stress testing for SMIP-MWP
# Runs extended benchmarks with multiple iterations to simulate production load

set -euo pipefail

# Configuration (can be overridden via environment)
DURATION="${DURATION:-300}"      # Test duration in seconds
CONCURRENT="${CONCURRENT:-16}"   # Concurrent connections/operations
LOAD_LEVEL="${LOAD_LEVEL:-high}" # low, medium, high
OUTPUT_DIR="${OUTPUT_DIR:-./benchmarks/stress}"

mkdir -p "${OUTPUT_DIR}"

TS=$(date -u +"%Y%m%dT%H%M%SZ")
OUT_FILE="${OUTPUT_DIR}/stress-${TS}.txt"

echo "=== SMIP-MWP Stress Test ===" > "${OUT_FILE}"
echo "Timestamp: ${TS}" >> "${OUT_FILE}"
echo "Duration: ${DURATION}s" >> "${OUT_FILE}"
echo "Concurrent: ${CONCURRENT}" >> "${OUT_FILE}"
echo "Load Level: ${LOAD_LEVEL}" >> "${OUT_FILE}"
echo "" >> "${OUT_FILE}"

# System info
echo "=== System Info ===" >> "${OUT_FILE}"
uname -a >> "${OUT_FILE}" 2>&1 || true
lscpu | head -20 >> "${OUT_FILE}" 2>&1 || true
free -h >> "${OUT_FILE}" 2>&1 || true
echo "" >> "${OUT_FILE}"

# Go version
go env >> "${OUT_FILE}" 2>&1 || true
echo "" >> "${OUT_FILE}"

# Determine test command based on load level
case "${LOAD_LEVEL}" in
    low)
        NUM_RUNS=5
        MEMORY_PRESSURE=false
        ;;
    medium)
        NUM_RUNS=8
        MEMORY_PRESSURE=true
        ;;
    high)
        NUM_RUNS=12
        MEMORY_PRESSURE=true
        CONCURRENT=32
        ;;
    *)
        echo "Unknown load level: ${LOAD_LEVEL}, using default" >> "${OUT_FILE}"
        NUM_RUNS=8
        MEMORY_PRESSURE=true
        CONCURRENT=16
        ;;
esac

echo "=== Configured Load Parameters ===" >> "${OUT_FILE}"
echo "Runs: ${NUM_RUNS}" >> "${OUT_FILE}"
echo "Concurrent: ${CONCURRENT}" >> "${OUT_FILE}"
echo "Memory Pressure: ${MEMORY_PRESSURE}" >> "${OUT_FILE}"
echo "" >> "${OUT_FILE}"

# Build crypto benchmarks for stress testing
echo "=== Building Crypto Stress Tests ===" >> "${OUT_FILE}"
cd ./internal/crypto || exit 1
go test -c -o /app/crypto-stress.test . 2>&1 | tee -a "${OUT_FILE}" || true
cd /app > /dev/null

if [ -f "/app/crypto-stress.test" ]; then
    echo "=== Running Crypto Stress Tests (${NUM_RUNS} iterations) ===" >> "${OUT_FILE}"
    /app/crypto-stress.test \
        -bench=. \
        -benchmem \
        -run=^$ \
        -count=${NUM_RUNS} \
        -cpuprofile="${OUTPUT_DIR}/stress-${TS}-cpu.prof" \
        -memprofile="${OUTPUT_DIR}/stress-${TS}-mem.prof" \
    2>&1 | tee -a "${OUT_FILE}" || true
    echo "" >> "${OUT_FILE}"
fi

# Network-related stress tests (if they exist)
echo "=== Running Network Stress Tests ===" >> "${OUT_FILE}"
go test ./internal/network -bench=. -benchmem -run=^$ -count=${NUM_RUNS} 2>&1 | tee -a "${OUT_FILE}" || true
echo "" >> "${OUT_FILE}"

# Full crypto package stress test
echo "=== Running Full Crypto Stress Test ===" >> "${OUT_FILE}"
go test ./internal/crypto -bench=. -benchmem -run=^$ -count=${NUM_RUNS} 2>&1 | tee -a "${OUT_FILE}" || true
echo "" >> "${OUT_FILE}"

# Profile analysis
if [ -f "${OUTPUT_DIR}/stress-${TS}-cpu.prof" ]; then
    echo "=== Top CPU Functions ===" >> "${OUT_FILE}"
    go tool pprof -top -raw \
        "${OUTPUT_DIR}/stress-${TS}-cpu.prof" 2>&1 | tee -a "${OUT_FILE}" || true
    echo "" >> "${OUT_FILE}"
fi

if [ -f "${OUTPUT_DIR}/stress-${TS}-mem.prof" ]; then
    echo "=== Top Memory Functions ===" >> "${OUT_FILE}"
    go tool pprof -top -alloc_space \
        "${OUTPUT_DIR}/stress-${TS}-mem.prof" 2>&1 | tee -a "${OUT_FILE}" || true
    echo "" >> "${OUT_FILE}"
fi

echo "=== Stress Test Complete ===" >> "${OUT_FILE}"
echo "Full results: ${OUT_FILE}" >> "${OUT_FILE}"

# Generate summary
SUMMARY="${OUTPUT_DIR}/stress-${TS}-summary.txt"
cat > "${SUMMARY}" << EOF
SMIP-MWP Stress Test Summary
============================
Timestamp: ${TS}
Duration: ${DURATION}s
Load Level: ${LOAD_LEVEL}
Configured Runs: ${NUM_RUNS}
Concurrent Operations: ${CONCURRENT}

CPU Profile: ${OUTPUT_DIR}/stress-${TS}-cpu.prof (if generated)
Memory Profile: ${OUTPUT_DIR}/stress-${TS}-mem.prof (if generated)

Key Metrics to Monitor:
- ns/op: Lower is better for throughput
- B/op: Lower allocations is better
- allocs/op: Zero allocations preferred for hot paths
- P99 latency: Should be consistent

Regard threshold: Watch for >5% regression in ns/op across runs
EOF

cat "${OUT_FILE}"
echo ""
echo "Stress test complete. Summary available at: ${SUMMARY}"
