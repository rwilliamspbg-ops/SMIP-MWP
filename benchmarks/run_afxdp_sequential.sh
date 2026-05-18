#!/usr/bin/env bash
set -euo pipefail
BASE=/workspaces/SMIP-MWP
PIDFILE=${BASE}/benchmarks/afxdp-w4-b4.pid
if [ ! -f "$PIDFILE" ]; then
  echo "PID file not found: $PIDFILE" >&2
  exit 1
fi
PID=$(cat "$PIDFILE")
# wait for the first run to finish
while kill -0 "$PID" 2>/dev/null; do
  sleep 5
done
# run batch=64 next
cd $BASE/internal/datapath/afxdp
BENCH_WORKERS=4 CRYPTO_BATCH_SIZE=64 go test -bench BenchmarkRunXDPLoop_MultiWorker_WithCrypto -run '^$' -benchtime=300s -benchmem -cpuprofile=${BASE}/benchmarks/afxdp-mw-w4-b64-300s-cpu.prof -memprofile=${BASE}/benchmarks/afxdp-mw-w4-b64-300s-mem.prof ./... > ${BASE}/benchmarks/afxdp-mw-w4-b64-300s.txt 2>&1
