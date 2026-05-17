#!/usr/bin/env bash
set -euo pipefail

# scripts/release.sh
# Build release artifacts for mohawk-node

OUT_DIR="dist"
GOOS=${GOOS:-linux}
GOARCH=${GOARCH:-amd64}
LDFLAGS=${LDFLAGS:-"-s -w"}

usage(){
  cat <<EOF
Usage: $0 [--out-dir dir]
  Builds a release tarball containing the mohawk-node binary and deployment files.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --out-dir) OUT_DIR="$2"; shift 2;;
    -h|--help) usage; exit 0;;
    *) echo "Unknown arg: $1"; usage; exit 2;;
  esac
done

mkdir -p "$OUT_DIR"
echo "Building mohawk-node ($GOOS/$GOARCH) -> $OUT_DIR"
CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="$LDFLAGS" -o "$OUT_DIR/mohawk-node" ./cmd/mohawk-node

cp deploy/mohawk-node.service "$OUT_DIR/"
cp -r scripts "$OUT_DIR/"

tar -C "$OUT_DIR" -czf "$OUT_DIR/mohawk-node-${GOOS}-${GOARCH}.tar.gz" .
echo "Release artifact: $OUT_DIR/mohawk-node-${GOOS}-${GOARCH}.tar.gz"
