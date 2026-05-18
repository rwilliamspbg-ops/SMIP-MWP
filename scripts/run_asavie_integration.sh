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
  ./scripts/bench.sh
fi

echo "Integration script completed"
