#!/usr/bin/env bash
set -euo pipefail

# scripts/run_experiments.sh
# Runs three experiment configurations on a host and collects pprof + flamegraphs.

if [ -z "${XDP_IFACE-}" ]; then
  echo "Please set XDP_IFACE environment variable (e.g. XDP_IFACE=eth0)" >&2
  exit 2
fi

OUT_DIR="benchmarks"
mkdir -p "${OUT_DIR}"

configs=("baseline:32" "throughput:128" "lowlat:16")

TS=$(date -u +"%Y%m%dT%H%M%SZ")

for c in "${configs[@]}"; do
  name=${c%%:*}
  batch=${c##*:}
  echo "=== Running experiment ${name} (BatchSize=${batch}) ==="

  export BatchSize=${batch}
  export FillThreshold=${batch}
  export FrameSize=2048
  export RUN_XDP_INTEGRATION=1
  export RUN_BENCH=1

  # run the integration helper which runs the integration test + bench
  chmod +x ./scripts/run_asavie_integration.sh
  ./scripts/run_asavie_integration.sh

  # tag the produced files (bench script uses timestamped filenames)
  # move any new profiles into experiment-specific directory
  EXP_DIR="${OUT_DIR}/${TS}-${name}"
  mkdir -p "${EXP_DIR}"
  mv ${OUT_DIR}/*${TS}* "${EXP_DIR}/" 2>/dev/null || true

  # process profiles into flamegraphs if available
  if command -v dot >/dev/null 2>&1; then
    echo "Generating flamegraphs for ${name}"
    ./scripts/process_benchmarks.sh || true
    if [ -d "${OUT_DIR}/flamegraphs" ]; then
      mv ${OUT_DIR}/flamegraphs "${EXP_DIR}/flamegraphs" || true
    fi
  else
    echo "graphviz not found; skipping flamegraph generation"
  fi

  echo "=== Experiment ${name} complete; artifacts in ${EXP_DIR} ==="
done

echo "All experiments complete. Check ${OUT_DIR} for artifacts." 
