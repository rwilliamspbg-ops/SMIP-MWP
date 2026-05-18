# Phase 2 → Phase 3: Hardware Deployment Quick Start

**Objective**: Deploy AF_XDP forwarder on real hardware and validate 10 Gbps target  
**Estimated Duration**: 2-3 days  
**Success Criteria**: 10 Gbps sustained throughput with <500 ns p99 latency  

---

## Pre-Deployment Checklist

### Hardware Requirements
- [ ] AF_XDP capable NIC (Intel ixgbe/i40e or Mellanox mlx5)
- [ ] Linux kernel 5.8+ (check: `uname -r`)
- [ ] BPF support enabled in kernel (check: `/sys/kernel/debug/tracing/`)
- [ ] 4+ CPU cores
- [ ] Traffic generation capability (iperf3, pktgen, or testpmd)

### Kernel Configuration
```bash
# Check kernel support
grep -i xdp /boot/config-$(uname -r)    # Should have CONFIG_NET_CLS_BPF=y
cat /sys/kernel/config/CONFIG_BPF      # Should be enabled

# Verify AF_XDP support
ip link show                             # Should support XDP

# Load XDP program prerequisites
sudo apt-get install -y llvm clang libelf-dev
sudo apt-get install -y bpftool
```

### NIC Driver Update (if needed)
```bash
# Intel ixgbe example
sudo apt-get install -y linux-modules-extra-$(uname -r)
sudo ethtool -i eth0     # Verify driver version

# Mellanox mlx5 example
# Usually pre-configured in modern distributions
```

---

## Deployment Steps

### Step 1: Prepare Forwarder Binary
```bash
cd /workspaces/SMIP-MWP

# Build with AF_XDP support
go build -tags=withafxdp -o ./bin/mohawk-node ./cmd/mohawk-node

# Verify binary
file ./bin/mohawk-node
./bin/mohawk-node --help
```

### Step 2: Configure AF_XDP Interface
```bash
# Identify target interface
ip link show

# Disable hardware checksum offload (for testing)
sudo ethtool -K eth0 tx-checksum-ip-generic off
sudo ethtool -K eth0 rx-checksum off

# Disable IRQ coalescing for low latency
sudo ethtool -C eth0 rx-usecs 0
sudo ethtool -C eth0 tx-usecs 0

# Verify settings
ethtool -C eth0
```

### Step 3: Set Up Traffic Generation
```bash
# Terminal 1: Start forwarder
sudo ./bin/mohawk-node \
    --iface eth0 \
    --dry-run=false \
    --frames 4096 \
    --frame-size 2048 \
    --batch-size 64 \
    --workers 4 \
    --metrics-addr :9090

# Terminal 2: Start prometheus (optional, for metrics)
# See MONITORING.md for setup

# Terminal 3: Generate traffic (iperf3 example)
iperf3 -c <forwarder-ip> -i 1 -t 60 -b 10G
```

### Step 4: Monitor Performance
```bash
# Watch forwarder metrics (in another terminal)
watch -n 1 'curl -s http://localhost:9090/metrics | grep afxdp'

# Monitor system performance
watch -n 1 'top -b -n 1 | head -15'
watch -n 1 'iostat -x 1 1 | head -10'
```

---

## Optimization Implementation (If Needed)

### If throughput < 9 Gbps
1. Implement Tier 1 optimizations (see OPTIMIZATION_ROADMAP.md)
   - Sharded session map (100-200 ns gain)
   - Batch pre-allocation (50-100 ns gain)
   - Session caching (50-100 ns gain)

2. Re-benchmark and re-run hardware test

3. If still < 9 Gbps: Implement Tier 2 (adaptive batching)

### If latency > 1 µs
1. Profile with pprof:
   ```bash
   go test -bench=. -cpuprofile=cpu.prof ./internal/datapath/afxdp
   go tool pprof cpu.prof
   ```

2. Identify hotspots and optimize

---

## Performance Testing Protocol

### Phase 1: Baseline (30 minutes)
```
1. Start forwarder at 1 Gbps traffic
2. Let settle for 5 minutes
3. Collect:
   - RX/TX packet count
   - Throughput
   - CPU utilization
   - Tail latencies (if available)
4. Record in: benchmarks/HARDWARE_BASELINE.txt
```

### Phase 2: Ramp Test (1 hour)
```
Traffic ramps: 1G → 2G → 3G → 4G → 5G → 7G → 10G → 12G

For each level:
- 5 minutes warm-up
- 5 minutes measurement
- Record: throughput, latency, drops, CPU
- Stop if: throughput < expected OR latency > threshold OR packet loss > 0.1%
```

### Phase 3: Sustained Load (1+ hours)
```
Run at 10 Gbps for:
- 1 hour: Verify stability
- 2 hours: Verify thermal stability
- Check for: packet loss, memory leaks, CPU drift
```

### Phase 4: Stress Test (optional, 1+ hours)
```
1. Ramp to 15 Gbps (above target)
2. Hold for 30 minutes
3. Verify graceful degradation (no crashes)
4. Return to 10 Gbps
5. Verify recovery
```

---

## Expected Results

### Minimum Success (7.5+ Gbps)
```
Throughput:     7.5+ Gbps
Latency (p50):  < 2 µs
Latency (p99):  < 10 µs
CPU usage:      60-80% per core
Packet loss:    0%
```

### Target Success (10+ Gbps)
```
Throughput:     10+ Gbps
Latency (p50):  < 1.5 µs
Latency (p99):  < 5 µs
CPU usage:      75-95% per core
Packet loss:    0%
```

### Maximum Success (10 Gbps with optimization)
```
Throughput:     10+ Gbps
Latency (p50):  < 1 µs
Latency (p99):  < 3 µs
CPU usage:      60-75% per core
Packet loss:    0%
```

---

## Troubleshooting

### Issue: "AF_XDP not supported" or "EINVAL"
**Solution**:
- Verify kernel version >= 5.8
- Check NIC driver compatibility
- Update NIC driver if needed
- Try different interface if available

### Issue: Low throughput (< 5 Gbps)
**Solutions**:
- Check if polling or interrupt driven
- Verify batch size is adequate (64-256)
- Check CPU frequency (turbo boost enabled?)
- Profile with pprof to identify bottleneck

### Issue: High latency variance
**Solutions**:
- Disable power saving: `cpupower frequency-set -d 3.5G`
- Disable turbo boost if inconsistent: `echo 0 > /sys/devices/system/cpu/intel_pstate/no_turbo`
- Increase batch interval tolerance
- Profile tail latencies

### Issue: Packet loss
**Solutions**:
- Increase UMEM frame count: `--frames 8192`
- Increase batch size: `--batch-size 128`
- Reduce traffic load temporarily
- Check for NIC/driver limitations

---

## Comparison Matrix

| Metric | Current (Stub) | Target (HW) | Success |
|--------|---|---|---|
| Build | ✅ Pass | ✅ Pass | ✅ |
| Tests | ✅ 50+ pass | ✅ 50+ pass | ✅ |
| Throughput | 7.06 Gbps* | 10 Gbps | ? |
| Latency | 1.7 µs* | <500 ns | ? |
| CPU/core | N/A | 75-95% | ? |

*Theoretical based on benchmarks, actual hardware may vary

---

## Real-World Deployment Considerations

### Production Checklist
- [ ] Error handling for NIC errors
- [ ] Monitoring/alerting setup
- [ ] Graceful degradation under overload
- [ ] Regular performance validation
- [ ] Security review (no user input to kernel)
- [ ] Documentation for operations team

### Performance Tuning (Post-Validation)
- [ ] Adjust batch size based on workload
- [ ] Tune UMEM frame allocation
- [ ] Configure CPU affinity strategy
- [ ] Set interrupt coalescing if needed
- [ ] Monitor and adjust HKDF cache TTL

---

## Next Milestones

| Milestone | Timeline | Criteria |
|-----------|----------|----------|
| Hardware ready | Day 1 | NIC + kernel + tools configured |
| Baseline established | Day 1-2 | Measure current performance |
| 7.5 Gbps achieved | Day 2-3 | Hit minimum success target |
| 10 Gbps achieved | Day 3-5 | Hit primary target |
| Optimized | Day 5-7 | Hit maximum target |
| Production ready | Week 3+ | Full validation + hardening |

---

## Contact & Support

For issues or questions during deployment:
1. Check TROUBLESHOOTING section above
2. Review OPTIMIZATION_ROADMAP.md for systematic optimization
3. Check kernel/driver documentation for HW-specific issues
4. Reference Linux AF_XDP documentation: https://www.kernel.org/doc/html/latest/networking/af_xdp.html

---

**Ready to deploy on hardware! Expected outcome: 10+ Gbps sustained throughput.**

