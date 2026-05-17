#!/usr/bin/env bash
set -euo pipefail

# scripts/ci_local.sh
# Run the same key checks as CI locally: download modules, run tests and vet.

echo "Downloading modules..."
go mod download

echo "Running tests..."
go test ./... -v

echo "Running go vet..."
go vet ./...

echo "CI local checks passed"
