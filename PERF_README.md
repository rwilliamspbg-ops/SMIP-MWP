# SMIP-MWP Performance Testing Quick Start

Complete setup for performance and stress testing your local Docker setup.

## 🚀 Quick Commands

### Option 1: Native Go Tests (Windows/WSL2)
```bash
cd C:\Users\rwill\SMIP-MWP

# Quick benchmark run (native Go)
go test ./internal/crypto -bench . -benchmem -run ^$ -count=3

# Full suite with profiling
./scripts/bench.sh --pprof

# Stress test high load
LOAD_LEVEL=high DURATION=300 ./scripts/stress-test.sh
```

### Option 2: Docker Test Environment (Cross-platform)
```powershell
# Build performance image
docker build -f Dockerfile.test -t smip-mwp:perf .

# Run performance tests
docker run --rm --privileged `
    -v "${PWD}\benchmarks:/app/benchmarks" `
    smip-mwp:perf \
    ./scripts/perf-quick.sh 3 internal/crypto

# Run stress tests
docker run --rm --privileged `
    -v "${PWD}\benchmarks:/app/benchmarks" `
    -e LOAD_LEVEL=high `
    smip-mwp:stress \
    ./scripts/stress-test.sh
```

### Option 3: WSL2 with Docker (Recommended for AF_XDP)
```bash
# In WSL2
cd /mnt/c/Users/rwill/SMIP-MWP

# Build image
docker build -f Dockerfile.test -t smip-mwp:perf .

# Run full suite
./scripts/perf-quick.sh 3 ./...

# Or use docker compose
docker compose -f docker-performance-test.yaml up perf-runner
```

## 📊 Performance Files Created

| File | Purpose |
|------|---------|
| `Dockerfile.test` | Performance test image builder |
| `Dockerfile.stress` | Stress testing image |
| `Dockerfile.afxdp` | AF_XDP fast-path tests (Linux only) |
| `docker-performance-test.yaml` | Docker Compose config |
| `scripts/perf-quick.sh` | Quick benchmark runner |
| `scripts/stress-test.sh` | High-load stress testing |
| `PERFORMANCE_TESTING_GUIDE.md` | Complete testing guide |
| `WINDOWS_PERFORMANCE_TESTING.md` | Windows-specific instructions |

## 🎯 Key Metrics to Monitor

```
BenchmarkName-8     15234578        78.9ns/op    allocs/op=0, B/op=0
BenchmarkName-8     10000000        123.4ns/op   allocs/op=2, B/op=1024
```

**Watch for:**
- **ns/op**: Lower is better (throughput)
- **B/op**: Lower is better (memory allocations)
- **allocs/op**: Zero preferred for hot paths
- **Regression threshold**: >5% increase in ns/op

## 🔍 Analyzing Results

### View latest benchmark
```bash
cat benchmarks/bench-localhost-*.txt | grep "Benchmark"
```

### CPU Profile Analysis
```bash
go tool pprof -top benchmarks/bench-localhost-*-cpu.prof
```

### Memory Profile Analysis
```bash
go tool pprof -alloc_space benchmarks/bench-localhost-*-mem.prof
```

## 📈 Stress Test Levels

| Level | Iterations | Concurrent | Use Case |
|-------|-----------|------------|----------|
| low   | 5         | 8          | Baseline testing |
| medium| 8         | 16         | Normal load |
| high  | 12        | 32         | Production stress |

Run with custom settings:
```bash
LOAD_LEVEL=high DURATION=600 CONCURRENT=64 ./scripts/stress-test.sh
```

## 🛠️ Docker Commands Reference

```bash
# Build test image
docker build -f Dockerfile.test -t smip-mwp:perf .

# Quick benchmark (native Go)
go run scripts/perf-quick.sh 3 internal/crypto

# Run with profiling enabled
./scripts/bench.sh --pprof

# Stress testing container
docker compose -f docker-performance-test.yaml up stress-tester

# Cleanup old benchmarks
docker run --rm smip-mwp:perf \
    bash -c "find /app/benchmarks -type f -name '*.txt' -mtime +1 -delete"
```

## 📚 Full Documentation

For detailed guides, see:
- `PERFORMANCE_TESTING_GUIDE.md` - Complete testing workflow
- `WINDOWS_PERFORMANCE_TESTING.md` - Windows-specific setup
- `docs/PERFORMANCE.md` - Performance architecture and metrics

## ⚡ Next Steps

1. **Establish baseline**: Run quick tests to get initial performance numbers
2. **Enable profiling**: Use `--pprof` flag for CPU/memory analysis
3. **Run stress tests**: Use high load to simulate production conditions
4. **Analyze profiles**: Use pprof tools to find optimization opportunities
5. **Track regression**: Monitor for >5% sustained increases in latency

---

**Tip**: For AF_XDP testing, use WSL2 with Ubuntu 22.04 running Docker on Linux kernel.
