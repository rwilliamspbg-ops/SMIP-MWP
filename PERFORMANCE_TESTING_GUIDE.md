# SMIP-MWP Performance & Stress Testing Guide

This guide provides comprehensive instructions for testing and stress-testing the SMIP-MWP repository using Docker locally.

## Quick Start

### Basic Tests (Native Go)
```bash
# Run all tests
go test ./... -v

# Run benchmarks
./scripts/bench.sh

# Run with profiling
./scripts/bench.sh --pprof

# Stress testing
DURATION=300 LOAD_LEVEL=high ./scripts/stress-test.sh
```

### Docker-Based Testing (Recommended for Cross-Platform)
```bash
# Build performance test image
docker build -f Dockerfile.test -t smip-mwp:perf .

# Run comprehensive test suite
docker run --rm --privileged \
    -v $(pwd)/benchmarks:/app/benchmarks \
    -e NUM_PROCESSES=4 \
    -e WORKERS=4 \
    smip-mwp:perf \
    go test ./... -bench . -benchmem -run ^$ -count=3

# Stress test with high load
docker run --rm --privileged \
    -v $(pwd)/benchmarks:/app/benchmarks \
    -e LOAD_LEVEL=high \
    -e DURATION=300 \
    smip-mwp:perf \
    ./scripts/stress-test.sh

# Cleanup previous results
docker run --rm --privileged \
    -v $(pwd)/benchmarks:/app/benchmarks \
    smip-mwp:perf \
    bash -c "find /app/benchmarks -type f -name '*.txt' -mtime +1 -delete"
```

## Docker Performance Testing Setup

### Prerequisites

For full performance testing with AF_XDP support, you need a Linux environment. Windows users can use WSL2 or WSLg.

```bash
# Check Docker availability
docker --version

# Build test image (Linux required for AF_XDP)
docker build -f Dockerfile.test -t smip-mwp:perf .
```

### Running Performance Tests with Docker Compose

```bash
# Create performance test environment
docker compose -f docker-performance-test.yaml up -d perf-runner

# View logs
docker logs -f smip-mwp-perf-test

# Run AF_XDP tests (Linux host only)
docker compose -f docker-performance-test.yaml run afxdp-tester go test -tags=withafxdp ./internal/... -v -bench .

# Run stress tests
docker compose -f docker-performance-test.yaml up -d stress-tester
```

## Performance Analysis

### Viewing Benchmark Results
```bash
# Latest benchmark results
cat benchmarks/bench-${HOST}-$(ls -t benchmarks/ | head -1).txt

# Compare multiple runs
echo "Run 1:"
grep -A5 "Benchmark" benchmarks/bench-*.txt | head -20

echo ""
echo "Run 2:"
grep -A5 "Benchmark" benchmarks/bench-*.txt | tail -20
```

### Analyzing CPU Profiles
```bash
# Launch pprof web UI for latest CPU profile
go tool pprof -http=:8081 benchmarks/bench-${HOST}-*-cpu.prof

# Show top 10 functions by time spent
go tool pprof -top -cum benchmarks/bench-${HOST}-*-cpu.prof

# Generate flamegraph
go tool pprof graphics=flamegraph benchmarks/bench-${HOST}-*-cpu.prof
```

### Analyzing Memory Profiles
```bash
# Show memory allocations
go tool pprof -alloc_space benchmarks/bench-${HOST}-*-mem.prof

# Show objects allocation count
go tool pprof -alloc_objects benchmarks/bench-${HOST}-*-mem.prof
```

## Stress Test Configuration

### Load Levels

| Level | Iterations | Concurrent | Use Case |
|-------|-----------|------------|----------|
| low   | 5         | 8          | Baseline testing |
| medium| 8         | 16         | Normal load simulation |
| high  | 12        | 32         | Production stress test |

### Running Different Stress Tests
```bash
# Low load (baseline)
LOAD_LEVEL=low ./scripts/stress-test.sh

# Medium load (normal operation)
LOAD_LEVEL=medium DURATION=600 ./scripts/stress-test.sh

# High load (production stress)
LOAD_LEVEL=high DURATION=1200 ./scripts/stress-test.sh

# Custom concurrent connections
CONCURRENT=64 LOAD_LEVEL=high ./scripts/stress-test.sh
```

## Key Performance Metrics

### Understanding Benchmark Output
```
BenchmarkName-8     15234578        78.9ns/op    allocs/op=0, B/op=0

BenchmarkName-8     10000000        123.4ns/op   allocs/op=2, B/op=1024
```

### What to Monitor
- **ns/op**: Nanoseconds per operation - Lower is better
- **B/op**: Bytes allocated per operation - Lower is better  
- **allocs/op**: Number of allocations - Zero preferred for hot paths
- **P99 latency**: Should remain consistent under load
- **GC frequency**: Check gctrace output if enabled

### Regression Thresholds
Watch for sustained regression >5% in:
- Throughput (ops/sec)
- Latency (ns/op)
- Memory allocations (B/op)

## Continuous Integration Integration

### Recommended CI Workflow

```yaml
# Add to .github/workflows/performance.yml
name: Performance Testing
on:
  schedule:
    - cron: '0 2 * * 1'  # Weekly on Monday at 2AM UTC
  workflow_dispatch:

jobs:
  perf-test:
    runs-on: ubuntu-22.04
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      - name: Build and Test
        run: |
          go test ./... -count=1 -timeout 60m
          ./scripts/bench.sh --pprof
      
      - name: Stress Test
        env:
          LOAD_LEVEL: high
          DURATION: 300
        run: |
          ./scripts/stress-test.sh
      
      - name: Upload Results
        uses: actions/upload-artifact@v3
        with:
          name: performance-results
          path: benchmarks/
```

## Debugging Performance Issues

### Enable GC Tracing
```bash
export GODEBUG=gctrace=1,gcflags=all=-m=2,-m=2,-m=2
go test ./... -bench .
```

### Memory Leaks Detection
```bash
go tool pprof -memprofile=mem.prof your_binary 2>&1 | \
    grep -A5 "InuseSpace"
```

### Thread Contention Analysis
```bash
go tool trace -output=trace.out ./your_binary &
# Run workload then
go tool trace trace.out
go tool trace web ui
```

## Troubleshooting

### No Results in Benchmarks
```bash
# Check Go version
go version

# Ensure tests are being run
go test ./internal/crypto -list=.

# List available benchmarks
go test ./internal/crypto -benchname=
```

### Permission Issues (Docker on Windows)
```bash
# Run with elevated privileges if needed
sudo docker build ...  # In WSL2 only

# Or use --privileged flag
docker run --privileged your-image
```

### AF_XDP Tests Only Work on Linux
```bash
# For AF_XDP testing, ensure:
- Linux host kernel 5.10+
- Required packages: libbpf-dev, clang, llvm
- Root privileges for BPF programs
```

## Best Practices

1. **Baseline First**: Always establish a baseline before optimizations
2. **Environment Consistency**: Use same hardware/software for comparisons
3. **Profiling**: Enable profiles for deep analysis
4. **Artifact Retention**: Keep results for 14+ days for regression tracking
5. **Documentation**: Log any configuration changes affecting performance

## Next Steps

For advanced topics, see:
- `docs/PERFORMANCE.md` - Performance architecture
- `IMPLEMENTATION_PLAN.md` - Implementation details  
- `benchmarks/README.md` - Benchmark policies

---

**Last Updated**: $(date +%Y-%m-%d)
