#!/usr/bin/env bash
set -euo pipefail

# scripts/process_benchmarks.sh
# Best-effort post-processing for profiles produced under benchmarks/
# Usage: run from repo root after downloading workflow artifacts

OUT_DIR="benchmarks"
FLAME_DIR="${OUT_DIR}/flamegraphs"
mkdir -p "${FLAME_DIR}"

echo "Building test binaries for all packages..."
BIN_DIR="/tmp/smip-mwp-bench-bins-$$"
mkdir -p "${BIN_DIR}"
for pkg in $(go list ./...); do
  safe=$(echo "${pkg}" | tr '/.' '__')
  out="${BIN_DIR}/${safe}.test"
  echo "Compiling ${pkg} -> ${out}"
  if go test -c "${pkg}" -o "${out}" >/dev/null 2>&1; then
    echo " -> compiled"
  else
    echo " -> compile failed for ${pkg}, skipping"
    rm -f "${out}"
  fi
done

echo "Processing profiles in ${OUT_DIR}..."
shopt -s nullglob
for prof in ${OUT_DIR}/*cpu.prof; do
  base=$(basename "${prof}")
  echo "Processing ${prof}"
  found=0
  for bin in ${BIN_DIR}/*.test; do
    # try pprof with this binary
    outsvg="${FLAME_DIR}/${base}--$(basename "${bin}").svg"
    echo "  trying binary ${bin} -> ${outsvg}"
    if go tool pprof -svg "${bin}" "${prof}" > "${outsvg}" 2>/dev/null; then
      echo "  generated ${outsvg}"
      found=1
      break
    else
      rm -f "${outsvg}" || true
    fi
  done
  if [ ${found} -eq 0 ]; then
    echo "  no matching binary found for ${prof}; consider running this script on a host that built the tests with the same toolchain/flags."
  fi
done

echo "Done. Flamegraphs (if any) are under ${FLAME_DIR}"
