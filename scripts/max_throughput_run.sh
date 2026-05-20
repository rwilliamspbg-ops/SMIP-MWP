#!/usr/bin/env bash
set -euo pipefail

# scripts/max_throughput_run.sh
# Helper to prepare a test host, run AF_XDP preflight, apply tuning, and
# launch repo benchmarks + a selectable traffic generator (iperf3 recommended
# for quick checks, MoonGen/TRex for line-rate saturation).

usage() {
  cat <<EOF
Usage: $0 --iface IFACE --role sender|receiver [options]

Options:
  --iface IFACE         Network interface to test (required)
  --role ROLE           "sender" or "receiver" (required)
  --generator NAME      iperf3|moongen|trex (default: iperf3)
  --queues N            Set NIC combined queues (default: 8)
  --hugepages N         Reserve N 2MB hugepages (default: 1024)
  --duration SEC        Traffic duration in seconds (default: 60)
  --benchtime SEC       bench.go benchtime (default: 30)
  --auto-pin            Automatically pin NIC IRQs to consecutive CPUs (requires sudo)
  --cpu-start N         Starting CPU index for IRQ pinning (default: 2)
  --launch-cmd CMD      Command to launch `mohawk-node` (will be pinned)
  --pin-range A-B       Explicit CPU range to pin worker (overrides IRQ-derived)
  --logfile PATH        Log file for launched process (default: benchmarks/mohawk-node.log)
  -h, --help            Show this help
EOF
}

IFACE=""
ROLE=""
GEN="iperf3"
QUEUES=8
HUGEPAGES=1024
DURATION=60
BENCHTIME=30
AUTO_PIN=0
CPU_START=2
LAUNCH_CMD=""
PIN_RANGE=""
LOGFILE="benchmarks/mohawk-node.log"
irq_nums=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --iface) IFACE="$2"; shift 2 ;; 
    --role) ROLE="$2"; shift 2 ;; 
    --generator) GEN="$2"; shift 2 ;; 
    --queues) QUEUES="$2"; shift 2 ;; 
    --hugepages) HUGEPAGES="$2"; shift 2 ;; 
    --duration) DURATION="$2"; shift 2 ;; 
    --benchtime) BENCHTIME="$2"; shift 2 ;; 
      --auto-pin) AUTO_PIN=1; shift ;; 
      --cpu-start) CPU_START="$2"; shift 2 ;; 
      --launch-cmd) LAUNCH_CMD="$2"; shift 2 ;; 
      --pin-range) PIN_RANGE="$2"; shift 2 ;; 
      --logfile) LOGFILE="$2"; shift 2 ;; 
    -h|--help) usage; exit 0 ;; 
    *) echo "Unknown arg: $1" >&2; usage; exit 2 ;; 
  esac
done

if [[ -z "$IFACE" || -z "$ROLE" ]]; then
  usage; exit 2
fi

echo "max_throughput_run: iface=$IFACE role=$ROLE gen=$GEN queues=$QUEUES hugepages=$HUGEPAGES duration=$DURATION"

# 1) Preflight
./scripts/test_xdp.sh --iface "$IFACE" --output "benchmarks/xdp-preflight-$(hostname)-$(date -u +%Y%m%dT%H%M%SZ).txt" || true

# 2) System tuning (best-effort, non-destructive)
echo "Applying sysctl tuning (tmp) — requires sudo for some ops"
sudo sysctl -w net.core.rmem_max=268435456 || true
sudo sysctl -w net.core.wmem_max=268435456 || true
sudo sysctl -w net.core.netdev_max_backlog=250000 || true
sudo sysctl -w net.core.somaxconn=65536 || true

# Reserve hugepages
echo "$HUGEPAGES" | sudo tee /proc/sys/vm/nr_hugepages || true

# Set NIC queue count to match CPU/worker layout
if command -v ethtool >/dev/null 2>&1; then
  echo "Setting NIC combined queues: $QUEUES"
  sudo ethtool -L "$IFACE" combined $QUEUES || echo "ethtool set-L failed, please run manually"
fi

# Optional: stop irqbalance to allow static IRQ pinning (operator choice)
if command -v systemctl >/dev/null 2>&1; then
  echo "Stopping irqbalance (will be restarted by operator if desired)"
  sudo systemctl stop irqbalance || true
fi

# 2a) Optional automatic IRQ pinning
if [[ "$AUTO_PIN" -eq 1 ]]; then
  echo "Auto-pinning IRQs for interface $IFACE starting at CPU $CPU_START"
  irq_lines=$(grep -i "$IFACE" /proc/interrupts || true)
  if [[ -z "$irq_lines" ]]; then
    echo "No IRQs found for interface $IFACE in /proc/interrupts; skipping auto-pin"
  else
    mapfile -t irq_nums < <(echo "$irq_lines" | awk -F: '{print $1}' | tr -d ' ') || mapfile -t irq_nums < <(echo)
    idx=0
    for irq in "${irq_nums[@]}"; do
      target_cpu=$((CPU_START + idx))
      # compute hex mask for the target CPU using python3
      if command -v python3 >/dev/null 2>&1; then
        mask=$(python3 -c "import sys; cpu=int(sys.argv[1]); print(format(1<<cpu,'x'))" "$target_cpu")
      else
        # fallback: simple mask for CPUs < 64 using printf and shifting in shell
        mask=$(printf '%x' $((1 << target_cpu))) || mask="1"
      fi
      echo "Pinning IRQ $irq -> cpu $target_cpu (mask=0x$mask)"
      sudo sh -c "printf '%s' $mask > /proc/irq/$irq/smp_affinity" || echo "Failed to write smp_affinity for IRQ $irq"
      idx=$((idx+1))
    done
    echo "Auto-pin complete. Recommended CPU list for workers:"
    cpu_list_start=$CPU_START
    cpu_list_end=$((CPU_START + idx - 1))
    echo "$cpu_list_start-$cpu_list_end"
  fi
fi

# Determine final CPU pin range
if [[ -n "$PIN_RANGE" ]]; then
  cpu_list="$PIN_RANGE"
elif [[ "$AUTO_PIN" -eq 1 && ${#irq_nums[@]} -gt 0 ]]; then
  cpu_list_start=$CPU_START
  cpu_list_end=$((CPU_START + ${#irq_nums[@]} - 1))
  cpu_list="${cpu_list_start}-${cpu_list_end}"
else
  cpu_list=""
fi

# 3) Build repo with AF_XDP (compile-only check)
if go test ./... -tags=withafxdp -run '^$' >/dev/null 2>&1; then
  echo "Go withafxdp compile check: OK"
else
  echo "Go withafxdp compile check failed — fix build on host before proceeding" >&2
fi

# 4) Run repo microbench (captures pprof in benchmarks/)
echo "Running repo afxdp benchmarks (benchtime=${BENCHTIME}s)"
./scripts/bench.sh --pprof -- go test ./internal/datapath/afxdp -bench . -benchmem -run ^$ -count=1 -benchtime=${BENCHTIME}s || true

# 5) Traffic generation
case "$GEN" in
  iperf3)
    if [[ "$ROLE" = "receiver" ]]; then
      echo "Starting iperf3 server (receiver). Run iperf3 client on sender host."
      echo "Command (receiver): iperf3 -s"
      echo "Command (sender):  iperf3 -c <receiver-ip> -P 1 -t $DURATION -b 40G"
    else
      echo "Run the following on the sender to attempt high-rate test (iperf3 may not saturate 25/50Gbps):"
      echo "iperf3 -c <receiver-ip> -P 1 -t $DURATION -b 40G"
    fi
    ;;
  moongen)
    echo "MoonGen recommended for line-rate generation. Example (run on sender):"
    echo "sudo ./build/MoonGen examples/pg_advanced.lua $IFACE $DURATION --pkt-size 128 --rate 100"
    echo "See MoonGen docs for build/install steps."
    ;;
  trex)
    echo "TRex is another option for hardware line-rate. See TRex docs."
    ;;
  *)
    echo "Unknown generator: $GEN"; exit 2
    ;;
esac

# 6) Guidance: IRQ pinning and worker affinity
cat <<'GUIDE'
Post-run actions and tips:
 - Inspect /proc/interrupts and bind each NIC queue IRQ to a dedicated core via /proc/irq/<IRQ>/smp_affinity.
 - Pin your Forwarder worker goroutines using taskset or by setting OS thread affinity inside the process.
 - Use MoonGen/TRex for sustained 25/50Gbps traffic; iperf3 is useful for quick functional checks.
 - Collect pprof CPU profiles produced by ./scripts/bench.sh and analyze hotspots with `go tool pprof`.
GUIDE

if [[ "$AUTO_PIN" -eq 1 && ${#irq_nums[@]} -gt 0 ]]; then
  echo
  echo "Worker affinity example (pin future process to worker CPUs):"
  cpu_list_start=$CPU_START
  cpu_list_end=$((CPU_START + ${#irq_nums[@]} - 1))
  echo "taskset -c ${cpu_list_start}-${cpu_list_end} <cmd>"
  echo "Or to pin an existing PID:" 
  echo "  sudo taskset -pc ${cpu_list_start}-${cpu_list_end} <pid>"
fi

# If requested, launch the provided command and pin it to the selected CPUs
if [[ -n "$LAUNCH_CMD" ]]; then
  if [[ -z "$cpu_list" ]]; then
    echo "No CPU list available to pin; launching without pinning"
    sh -c "$LAUNCH_CMD" >> "$LOGFILE" 2>&1 &
    echo $! > benchmarks/mohawk-node.pid
    echo "Launched process pid=$(cat benchmarks/mohawk-node.pid) logfile=$LOGFILE"
  else
    echo "Launching and pinning command to CPUs: $cpu_list"
    # start process under taskset and capture PID
    sh -c "taskset -c $cpu_list $LAUNCH_CMD >> $LOGFILE 2>&1 & echo \$!" > benchmarks/mohawk-node.pid
    pid=$(cat benchmarks/mohawk-node.pid)
    echo "Launched pid=$pid logfile=$LOGFILE"
  fi
fi

exit 0
