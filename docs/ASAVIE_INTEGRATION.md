asavie/xdp integration

This document describes steps to implement kernel-backed UMEM and XDPSocket
integration using github.com/asavie/xdp. The repository already contains
build-tagged scaffolds under `internal/datapath/afxdp/`:

- forwarder_xdp_umem_asavie.go (withafxdp && asavie)
- forwarder_xdp_socket_asavie.go (withafxdp && asavie)

Implementation checklist

1. Add real UMEM allocation
   - Use `xdp.NewUmem()` (or equivalent) to allocate a UMEM region with the
     requested frame size and number of frames.
   - Pin memory and populate a frame-to-index mapping that matches the
     `UMEM` wrapper methods used by the forwarder (AllocateIndices, ReturnIndices,
     GetFrameByIndex).

2. Implement socket creation
   - Create a raw AF_XDP socket bound to `iface` and `queue` using the
     UMEM region. Ensure RX/TX/Fill/Completion rings are configured.
   - Provide efficient methods that return descriptor slices referencing the
     UMEM frame indices (so the forwarder can zero-copy operate on frames).

3. Map descriptor lifecycle to UMEM wrapper
   - When allocating descriptors, return frame indices that correspond to
     allocated frames in the UMEM wrapper.
   - On completion, return frame indices to UMEM free list.

4. Build & test
   - Build with tags: `go build -tags="withafxdp asavie" ./cmd/mohawk-node`
   - Run on a supported host with required kernel libs (`libbpf-dev`, `clang`,
     `llvm`, `libelf-dev`) and XDP-capable NIC driver.

5. Metrics & performance
   - Validate zero-copy path correctness by enabling `withafxdp` and
     measuring b/w and p99 latency using the existing bench harness.
   - Tune `BatchSize`, `FrameSize`, and UMEM fill thresholds.
   - Tune `BatchSize`, `FrameSize`, and `FillThreshold` (new forwarder config).

  Tuning guidance

  - `BatchSize`: number of packets processed per tick/iteration; larger values increase throughput but add latency.
  - `FillThreshold`: target number of descriptor slots to refill into UMEM; setting this to match `BatchSize` is a good starting point.
   - `FillThreshold`: target number of descriptor slots to refill into UMEM; setting this to match `BatchSize` is a good starting point.
   - `FillAdaptive`: enable dynamic fill adjustment based on observed completion rate.
   - `FillAdaptFactor`: multiplicative factor applied to observed completions to compute target refill (default 1.5).
   - `FillEMAAlpha`: EMA alpha for smoothing observed completion rate (0..1). Default is 0.25.
  - `FrameSize`: set to match your MTU plus headroom (2048 is a reasonable default).

  Start with:

  ```bash
  BatchSize=64 FillThreshold=64 FrameSize=2048
  ```

  Then vary `BatchSize` and `FillThreshold` together to find the best tradeoff on your hardware.

  Running the integration test

  To run the guarded integration test (requires kernel support and privileges), use:

  ```bash
  # Example (only run on a host with AF_XDP-capable NIC and required libs):
  RUN_XDP_INTEGRATION=1 XDP_IFACE=eth0 go test ./internal/datapath/afxdp -tags="withafxdp asavie" -run TestAsavieIntegrationSanity -v
  ```

  Notes:
  - `RUN_XDP_INTEGRATION` must be set to `1` to enable the live socket portion of the test.
  - `XDP_IFACE` should be the network interface name (e.g., `eth0`) that supports AF_XDP.
  - Running the test will attempt to create an AF_XDP socket and requires appropriate privileges and kernel support.

Notes

- The scaffolds currently return an explanatory error when called; they are
  intentionally build-tag guarded so the repository builds without the
  `asavie` tag.
- AF_XDP requires privileged operations and kernel support; CI runners may
  not provide the required privileges. Prefer dedicated hardware or a VM
  with host networking for full integration testing.
