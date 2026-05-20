# SMIP-MWP Performance Testing on Windows

This guide covers running performance and stress tests on Windows using Docker.

## Prerequisites for Windows

### 1. Install WSL2 (Windows Subsystem for Linux)
```powershell
# Enable WSL2 feature in PowerShell as Administrator
wsl --install -d ubuntu-22.04
```

### 2. Setup Docker Desktop for Windows
```powershell
# Download from https://docs.docker.com/desktop/windows/
# Configure to use WSL2 backend (recommended)
docker --version  # Verify installation
```

### 3. Go Development Environment
```bash
# Download from https://go.dev/dl/ or use winget
winget install Go
# OR add Go path manually:
$env:PATH += ";C:\Program Files\Go\bin"
```

## Quick Start - Native Windows Testing

If you have Go installed natively on Windows:

```bash
# Navigate to repo
cd C:\Users\rwill\SMIP-MWP

# Run quick performance test
go run scripts/perf-quick.sh

# Run full benchmarks (native)
go test ./... -bench . -benchmem -run ^$ -count=3
```

## WSL2 Approach (Recommended for Full Testing)

### 1. Build Test Image in WSL2
```bash
wsl bash -c "cd /mnt/c/Users/rwill/SMIP-MWP && docker build \
    -f Dockerfile.test \
    -t smip-mwp:perf ."
```

### 2. Run Performance Tests from Windows
```powershell
$env:WSL_DISTRO_NAME="ubuntu-22.04"
$env:CGO_ENABLED=1

# Mount and run in WSL2
wsl bash -c "/mnt/c/Users/rwill/SMIP-MWP/scripts/perf-quick.sh 3 internal/crypto"
```

### 3. Run Stress Tests
```bash
wsl bash -c "cd /mnt/c/Users/rwill/SMIP-MWP && \
    LOAD_LEVEL=high DURATION=300 ./scripts/stress-test.sh"
```

## Docker Commands (WSL2 Backend)

### Build Performance Test Image
```powershell
docker build -f Dockerfile.test -t smip-mwp:perf .
```

### Run Full Performance Suite
```powershell
docker run --rm `
    --privileged `
    -v "${PWD}\benchmarks:/app/benchmarks" `
    -e NUM_PROCESSES=4 `
    -e WORKERS=4 `
    smip-mwp:perf \
    go test ./... -v -count=1
```

### Run with Profiling
```powershell
docker run --rm `
    --privileged `
    -v "${PWD}\benchmarks:/app/benchmarks" `
    -e PROFILE=true `
    smip-mwp:perf \
    ./scripts/perf-quick.sh 3 . --profile
```

### Run Stress Tests
```powershell
docker run --rm `
    --privileged `
    -v "${PWD}\benchmarks:/app/benchmarks" `
    -e LOAD_LEVEL=high `
    -e DURATION=300 `
    smip-mwp:stress \
    ./scripts/stress-test.sh
```

## Results Analysis

### View Benchmark Output
```powershell
# Latest benchmark results
Get-Content benchmarks/bench-localhost-**.txt | Select-Object -Last 50

# Extract only benchmark numbers
Select-String -Path "benchmarks\*-latest.txt" -Pattern "^Benchmark[A-Za-z]" 
```

### CPU Profile Analysis (in WSL2)
```bash
wsl bash -c "cd /mnt/c/Users/rwill/SMIP-MWP && \
    go tool pprof -http=:8081 benchmarks/bench-localhost-*-cpu.prof"
```

### Generate Flame Graph
```bash
wsl bash -c "cd /mnt/c/Users/rwill/SMIP-MWP && \
    go tool pprof \
        -nodecount=10 \
        -text \
        -raw \
        benchmarks/bench-localhost-*-cpu.prof"
```

## Performance Comparison Across Runs

### Create Comparison Report
```bash
wsl bash -c "cd /mnt/c/Users/rwill/SMIP-MWP && cat > benchmarks/comparison.txt << 'EOF'
# SMIP-MWP Performance Comparison Report
Generated: $(date)
=========================================="

Latest Run ($(ls -t benchmarks/ | head -1)):
$(cat benchmarks/bench-localhost-*20*.txt | grep -A3 "^Benchmark")

Previous Best (search for lowest ns/op):
$(for f in benchmarks/bench-*.txt; do \
    echo "=== $f ==="; \
    grep "Benchmark" "$f" | grep -v "^#" | head -1; \
done | head -5)
EOF
cat benchmarks/comparison.txt"
```

## Windows-Specific Optimizations

### 1. Set Priorities for Docker Containers
```powershell
docker run --rm `
    --cpus=8 `
    --memory=4g `
    --memory-swap=0 `
    -v "${PWD}\benchmarks:/app/benchmarks" \
    smip-mwp:perf \
    go test ./internal/crypto -bench=. -count=3
```

### 2. Disable Windows Defender for WSL2
```powershell
# In PowerShell as Administrator
Set-MpPreference -DisableRealtimeMonitoring $true
```

### 3. Use Hyper-V Enhanced Mode (Optional)
```powershell
docker-machine env wsl --hyper-v
```

## Troubleshooting on Windows

### Issue: "Permission denied" when running tests
**Solution**: Run Docker Desktop as Administrator or use WSL2 backend

### Issue: CGO_ENABLED=1 failing  
**Solution**: Install C++ Build Tools for Windows
```powershell
winget install Microsoft.VisualStudio.2022.BuildTools
```

### Issue: Slow benchmarks
**Solution**: 
- Enable Hyper-Threading
- Use dedicated CPU cores (`--cpus` flag)
- Reduce memory pressure if testing

### Issue: Can't access AF_XDP tests
**Solution**: Requires Linux kernel - use WSL2 with Ubuntu 22.04

## Performance Baselines

For Windows hardware reference:

| Hardware | Expected ns/op Range | Notes |
|----------|---------------------|-------|
| i7-13700K | 50-200ns | Single core crypto |
| Ryzen 9 7950X | 40-180ns | AMD platform baseline |
| Threadripper | <50ns | Multi-core optimized |

## Continuous Performance Monitoring

### Setup Watch Script
```powershell
# Add to startup or schedule with Task Scheduler
@"
$latest = (Get-ChildItem benchmarks/*.txt | Sort LastWriteTime -Descending)[0].Name
Write-Host "Latest benchmark: $latest"
Get-Content benchmarks/$latest | Select-String "Benchmark" | Select-Object -First 10
"@ | Set-Content -Path "$env:USERPROFILE\bin\smip-perf-watch.ps1"
```

## Next Steps

For full AF_XDP testing on Windows, you MUST use WSL2. See `PERFORMANCE_TESTING_GUIDE.md` for detailed Linux-based testing instructions.
