QUIC Comparison Run Instructions

Objective
- Provide quick commands to run a `quic-go` echo server and a simple client for apples-to-apples comparison with AF_XDP datapath.

Quick server (uses upstream quic-go examples)

```bash
# Clone quic-go examples and run the echo server
git clone https://github.com/quic-go/quic-go.git /tmp/quic-go
cd /tmp/quic-go/examples/echo
# Generate certs if needed (script in repo) or use provided certs
go run server/main.go --addr :4433 --cert ../../testdata/cert.pem --key ../../testdata/key.pem
```

Quick client

```bash
# On client host
cd /tmp/quic-go/examples/echo
# Run client to connect and send N requests
go run client/main.go --addr <DUT_IP>:4433 --requests 10000 --concurrency 100
```

Notes
- Use the same CPU pinning and NIC configuration on the DUT as used for AF_XDP runs.
- For better load generation, consider writing a custom client in `quic-go` that opens many streams concurrently and measures p99/p999 latencies.
- If `go run` is slow or network-restricted, vendor the minimal example into `benchmarks/quic/` and run locally.
