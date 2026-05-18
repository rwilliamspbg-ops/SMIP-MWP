#!/usr/bin/env bash
set -euo pipefail

# Helper script to run Asavie/XDP integration and optional benchmarks.
# Environment variables:
# - XDP_IFACE (required on runner)
# - RUN_BENCH (optional): if set to 1, run ./scripts/bench.sh after integration

IFACE=${XDP_IFACE:-}
if [ -z "$IFACE" ]; then
  echo "XDP_IFACE not set. Set XDP_IFACE to the interface name (e.g., eth0)."
  exit 2
fi

echo "Running Asavie/XDP integration test on interface $IFACE"

# Build and run the integration test
RUN_XDP_INTEGRATION=1 XDP_IFACE="$IFACE" go test ./internal/datapath/afxdp -tags="withafxdp asavie" -run TestAsavieIntegrationSanity -v

# Ensure benchmarks directory exists for artifact upload
mkdir -p benchmarks

if [ "${RUN_BENCH:-0}" = "1" ]; then
  echo "Running benchmark harness"
  chmod +x ./scripts/bench.sh
  # Run benchmarks with pprof enabled so artifacts are produced for analysis
  ./scripts/bench.sh --pprof

  # Attempt to capture Prometheus metrics snapshot if endpoint is reachable
  METRICS_ADDR=${METRICS_ADDR:-":9090"}
  TS=$(date -u +"%Y%m%dT%H%M%SZ")
  if command -v curl >/dev/null 2>&1; then
    echo "Attempting to scrape metrics from ${METRICS_ADDR}"
    if curl -fsS "http://${METRICS_ADDR}/metrics" -o "benchmarks/metrics-${TS}.txt"; then
      echo "Metrics snapshot saved to benchmarks/metrics-${TS}.txt"
    else
      echo "Metrics endpoint not available at ${METRICS_ADDR} or scrape failed"
    fi
  fi
fi

echo "Integration script completed"
