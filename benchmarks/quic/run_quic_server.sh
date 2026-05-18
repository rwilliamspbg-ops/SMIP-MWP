#!/usr/bin/env bash
set -euo pipefail

# Quick helper to run quic-go echo server from upstream examples
# Usage: ./run_quic_server.sh :4433
ADDR=${1:-:4433}
TMPDIR=${TMPDIR:-/tmp/quic-go}

if [ ! -d "$TMPDIR" ]; then
  git clone https://github.com/quic-go/quic-go.git "$TMPDIR"
fi

pushd "$TMPDIR/examples/echo" >/dev/null
# Generate certs if missing
if [ ! -f "../../testdata/cert.pem" ]; then
  echo "Generating self-signed certs..."
  openssl req -x509 -newkey rsa:2048 -keyout ../../testdata/key.pem -out ../../testdata/cert.pem -days 365 -nodes -subj "/CN=localhost"
fi

go run server/main.go --addr "$ADDR" --cert ../../testdata/cert.pem --key ../../testdata/key.pem
popd >/dev/null
