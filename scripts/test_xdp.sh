#!/usr/bin/env bash
set -euo pipefail

# AF_XDP host preflight and optional compile-only validation.
#
# Usage:
#   ./scripts/test_xdp.sh
#   ./scripts/test_xdp.sh --iface eth0
#   ./scripts/test_xdp.sh --iface eth0 --run-go-test
#   ./scripts/test_xdp.sh --output benchmarks/xdp-preflight.txt

OUT_DIR="benchmarks"
mkdir -p "$OUT_DIR"
TS="$(date -u +"%Y%m%dT%H%M%SZ")"
HOST="$(hostname -s 2>/dev/null || echo "localhost")"
OUT_FILE="$OUT_DIR/xdp-preflight-${HOST}-${TS}.txt"

IFACE=""
RUN_GO_TEST=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --iface)
      IFACE="${2:-}"
      shift 2
      ;;
    --run-go-test)
      RUN_GO_TEST=1
      shift
      ;;
    --output)
      OUT_FILE="${2:-}"
      shift 2
      ;;
    -h|--help)
      cat <<'EOF'
Usage: ./scripts/test_xdp.sh [options]

Options:
  --iface <name>      Network interface to validate (default: route-derived)
  --run-go-test       Run compile-only withafxdp check (go test -run '^$')
  --output <file>     Write report to custom file
  -h, --help          Show this help
EOF
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 2
      ;;
  esac
done

if [[ -z "$IFACE" ]]; then
  IFACE="$(ip route get 1.1.1.1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if($i=="dev") {print $(i+1); exit}}')"
fi

PASS=0
WARN=0
FAIL=0

say() {
  echo "$*" | tee -a "$OUT_FILE"
}

pass() {
  PASS=$((PASS+1))
  say "PASS: $*"
}

warn() {
  WARN=$((WARN+1))
  say "WARN: $*"
}

fail() {
  FAIL=$((FAIL+1))
  say "FAIL: $*"
}

version_ge() {
  # returns true if $1 >= $2 (semantic-ish, numeric dot-separated)
  local IFS=.
  local i
  local -a a=($1) b=($2)
  local len=${#a[@]}
  if [[ ${#b[@]} -gt $len ]]; then
    len=${#b[@]}
  fi
  for ((i=0; i<len; i++)); do
    local av=${a[i]:-0}
    local bv=${b[i]:-0}
    if ((10#$av > 10#$bv)); then return 0; fi
    if ((10#$av < 10#$bv)); then return 1; fi
  done
  return 0
}

mkdir -p "$(dirname "$OUT_FILE")"
: > "$OUT_FILE"

say "SMIP-MWP AF_XDP preflight"
say "timestamp: $TS"
say "host: $HOST"
say "iface: ${IFACE:-<unset>}"
say "run_go_test: $RUN_GO_TEST"
say

for cmd in uname ip awk grep sed ethtool go; do
  if command -v "$cmd" >/dev/null 2>&1; then
    pass "required command available: $cmd"
  else
    fail "required command missing: $cmd"
  fi
done

kernel_raw="$(uname -r | sed 's/-.*//')"
if version_ge "$kernel_raw" "5.10"; then
  pass "kernel version $kernel_raw meets baseline (>=5.10)"
else
  fail "kernel version $kernel_raw below baseline (>=5.10)"
fi

if [[ -n "$IFACE" ]] && ip link show "$IFACE" >/dev/null 2>&1; then
  pass "interface exists: $IFACE"
else
  fail "interface not found; pass --iface <name>"
fi

if [[ -n "$IFACE" ]] && ip link show "$IFACE" >/dev/null 2>&1; then
  driver="$(ethtool -i "$IFACE" 2>/dev/null | awk '/^driver:/{print $2}')"
  if [[ -n "$driver" ]]; then
    pass "driver detected: $driver"
    case "$driver" in
      ixgbe|i40e|ice|mlx5_core|ena)
        pass "driver is in known XDP-friendly list"
        ;;
      *)
        warn "driver not in default known list; verify XDP support manually"
        ;;
    esac
  else
    warn "unable to detect NIC driver with ethtool"
  fi
fi

if [[ -f /usr/include/bpf/libbpf.h ]]; then
  pass "libbpf headers found"
else
  warn "libbpf headers not found (/usr/include/bpf/libbpf.h)"
fi

for cmd in clang llvm-config; do
  if command -v "$cmd" >/dev/null 2>&1; then
    pass "tool found: $cmd"
  else
    warn "tool missing: $cmd"
  fi
done

huge_total="$(awk '/HugePages_Total/ {print $2}' /proc/meminfo 2>/dev/null || echo 0)"
if [[ -n "$huge_total" ]] && [[ "$huge_total" =~ ^[0-9]+$ ]] && (( huge_total > 0 )); then
  pass "hugepages configured: HugePages_Total=$huge_total"
else
  warn "no hugepages configured (HugePages_Total=0)"
fi

if [[ "$(id -u)" -eq 0 ]]; then
  pass "running as root (XDP attach operations permitted)"
else
  warn "not running as root; attach/runtime operations may fail"
fi

if [[ "$RUN_GO_TEST" -eq 1 ]]; then
  say
  say "running compile-only check: go test ./... -tags=withafxdp -run '^$'"
  if go test ./... -tags=withafxdp -run '^$' >> "$OUT_FILE" 2>&1; then
    pass "withafxdp compile-only check passed"
  else
    fail "withafxdp compile-only check failed (see report)"
  fi
fi

say
say "summary: PASS=$PASS WARN=$WARN FAIL=$FAIL"
say "report: $OUT_FILE"

if (( FAIL > 0 )); then
  exit 1
fi
exit 0
