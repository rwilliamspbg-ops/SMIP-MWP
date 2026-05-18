#!/usr/bin/env bash
set -eu
BASE="/workspaces/SMIP-MWP"
FILES=(
  "$BASE/benchmarks/crypto-decrypt-120s-cpu.prof"
  "$BASE/benchmarks/afxdp-mw-w4-b4-300s-cpu.prof"
  "$BASE/benchmarks/afxdp-mw-w4-b64-300s-cpu.prof"
)
LOG="$BASE/benchmarks/pprof/watcher.log"
mkdir -p "$(dirname "$LOG")"
echo "watcher start: $(date)" > "$LOG"

while true; do
  allok=true
  for f in "${FILES[@]}"; do
    if [ ! -s "$f" ]; then
      allok=false
      break
    fi
  done
  if $allok; then
    break
  fi
  echo "$(date): waiting for profiles..." >> "$LOG"
  sleep 15
done

echo "$(date): profiles present, building test binaries" >> "$LOG"
cd "$BASE"
# Build test binaries for pprof; ignore errors but log
if go test -c -o internal/datapath/afxdp/afxdp.test ./internal/datapath/afxdp >> "$LOG" 2>&1; then
  echo "built afxdp.test" >> "$LOG"
else
  echo "failed to build afxdp.test" >> "$LOG"
fi
if go test -c -o internal/crypto/crypto.test ./internal/crypto >> "$LOG" 2>&1; then
  echo "built crypto.test" >> "$LOG"
else
  echo "failed to build crypto.test" >> "$LOG"
fi

mkdir -p "$BASE/benchmarks/pprof"

if [ -f internal/datapath/afxdp/afxdp.test ]; then
  echo "generating afxdp b4 pprof top" >> "$LOG"
  go tool pprof -top internal/datapath/afxdp/afxdp.test "$BASE/benchmarks/afxdp-mw-w4-b4-300s-cpu.prof" > "$BASE/benchmarks/pprof/afxdp-mw-w4-b4-300s-cpu-top.txt" 2>>"$LOG" || true
  echo "generating afxdp b64 pprof top" >> "$LOG"
  go tool pprof -top internal/datapath/afxdp/afxdp.test "$BASE/benchmarks/afxdp-mw-w4-b64-300s-cpu.prof" > "$BASE/benchmarks/pprof/afxdp-mw-w4-b64-300s-cpu-top.txt" 2>>"$LOG" || true
fi

if [ -f internal/crypto/crypto.test ]; then
  echo "generating crypto decrypt pprof top" >> "$LOG"
  go tool pprof -top internal/crypto/crypto.test "$BASE/benchmarks/crypto-decrypt-120s-cpu.prof" > "$BASE/benchmarks/pprof/crypto-decrypt-120s-cpu-top.txt" 2>>"$LOG" || true
fi

echo "$(date): pprof tops written" >> "$LOG"
