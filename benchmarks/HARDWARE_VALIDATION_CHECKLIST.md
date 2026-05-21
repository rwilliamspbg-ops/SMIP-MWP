# Hardware Validation Checklist

This checklist contains commands and artifacts to produce when running hardware validation for AF_XDP.

## Preflight
- Reserve test window and nodes.
- Stop irqbalance on DUT:
  sudo systemctl stop irqbalance || sudo pkill irqbalance || true
- Reserve hugepages:
  echo 4096 | sudo tee /proc/sys/vm/nr_hugepages
- Install prerequisites:
  sudo apt-get update && sudo apt-get install -y ethtool tcpdump
- Verify NIC:
  ip -brief link
  sudo ethtool -i <IFACE>

## IRQ pinning
- List IRQs for NIC:
  grep <IFACE> -n /proc/interrupts || cat /proc/interrupts | grep <IFACE>
- Pin IRQ 40 to CPU2 (example):
  echo 2 | sudo tee /proc/irq/40/smp_affinity_list
- Verify mapping:
  cat /proc/irq/40/smp_affinity_list

## Start DUT receiver (example helper)
- Use helper script (recommended):
  sudo ./scripts/max_throughput_run.sh --iface <IFACE> --role receiver --generator moongen --queues 16 --hugepages 4096 --auto-pin --cpu-start 2
- Or explicit run:
  sudo taskset -c 2-17 ./mohawk-node --mode afxdp --iface <IFACE> --queues 16 --cpuprofile /tmp/bench_cpu.prof --memprofile /tmp/bench_mem.prof

## Start MoonGen (generator)
- Place `moongen/send-udp.lua` on generator and run:
  ./build/MoonGen moongen/send-udp.lua txPort=0 rate=10000000000

## Monitoring & artifacts
- Collect dmesg:
  sudo dmesg -T > /tmp/dmesg.txt
- Interrupt map:
  cat /proc/interrupts > /tmp/interrupts.txt
- ethtool stats:
  sudo ethtool -S <IFACE> > /tmp/ethtool_stats.txt
- pprof files from DUT (if used):
  /tmp/bench_cpu.prof /tmp/bench_mem.prof
- tcpdump (optional):
  sudo tcpdump -i <IFACE> -w /tmp/traffic.pcap

## Post-run
- Archive and upload artifacts:
  tar czf afxdp-hw-<DATE>.tgz /tmp/bench*.prof /tmp/dmesg.txt /tmp/interrupts.txt /tmp/ethtool_stats.txt /tmp/traffic.pcap
- Attach artifacts to PR #18 or provide download link.

## Acceptance criteria
- No crashes or kernel oops.
- Throughput meets target for tested packet sizes.
- Multi-worker throughput scales as expected.
- `go test -bench` shows `0 B/op` and `0 allocs/op` on core loops.

