Running traffic generators for hardware validation

MoonGen
- Build MoonGen on the traffic generator host per MoonGen docs (LuaJIT required).
- Example run (sender):

```bash
# on MoonGen host, build first
make
sudo ./build/MoonGen benchmarks/moongen/l3_load_latency.lua --dev0 0 --dev1 1 --rate 10000 --duration 600
```

TRex
- Use TRex for sustained high-rate tests and complex profiles. See TRex quick notes in `benchmarks/trex/README_TREX.md`.

Best practices
- Pin generator and DUT CPUs, disable turbo, set IRQ affinity.
- Ensure identical firmware and appropriate drivers on NICs.
- Collect pcap+stats and the DUT's `benchmarks/*.prof` artifacts for correlation.
