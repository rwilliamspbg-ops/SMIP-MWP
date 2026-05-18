#!/usr/bin/env bash
set -euo pipefail

# Quick helper to run quic-go echo client from upstream examples
# Usage: ./run_quic_client.sh <server-ip:4433> [requests] [concurrency]
ADDR=${1:?server address}
REQS=${2:-10000}
CONC=${3:-100}
TMPDIR=${TMPDIR:-/tmp/quic-go}

if [ ! -d "$TMPDIR" ]; then
  git clone https://github.com/quic-go/quic-go.git "$TMPDIR"
fi

pushd "$TMPDIR/examples/echo" >/dev/null
go run client/main.go --addr "$ADDR" --requests "$REQS" --concurrency "$CONC"
popd >/dev/null
