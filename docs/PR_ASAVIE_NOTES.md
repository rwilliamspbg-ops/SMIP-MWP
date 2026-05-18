ASAVIE / AF_XDP Integration — PR notes

Summary
- Adds build-tagged AF_XDP UMEM and socket wrappers (mock + asavie adapter).
- Adds guarded integration test and a manual GitHub Actions workflow (`.github/workflows/asavie-integration.yml`) that targets a self-hosted runner labeled `xdp-runner`.
- Implements adaptive UMEM refill (EMA-driven) and exposes Prometheus metrics for fill EMA/target and processing latency.

Provisioning & Runner
- Grafana dashboard for the metrics: `docs/grafana/afxdp-dashboard.json`.
- Grafana provisioning example: `docs/grafana/provisioning/dashboards.yaml` and `docs/grafana/provisioning/afxdp-dashboard.json` (copy target for `/var/lib/grafana/dashboards/afxdp`).
- Runner helper script: `scripts/setup_xdp_runner.sh` — installs prerequisites and copies dashboard to `/var/lib/grafana/dashboards/afxdp` when run as root.
- Systemd unit example: `scripts/xdp-runner.service` to run the setup script at boot.
- Runner README: `docs/runner/README.md` with registration and run instructions.

How to run integration (short)
1. Provision a host with AF_XDP-capable NIC and required kernel libs.
2. Register a self-hosted runner for this repo and label it `xdp-runner`.
3. On the host, clone this repo and run `sudo ./scripts/setup_xdp_runner.sh`.
4. In Actions UI, trigger `Asavie XDP Integration` workflow and provide `iface` (e.g., `eth0`) and `run_bench=true` if you want benchmarks.

Notes
- The integration runs privileged operations; consider using a dedicated test host or VM with host networking.
- Benchmark pprof artifacts are uploaded as workflow artifacts under `benchmarks/**`. Processing into flamegraphs is left to the operator — I can add an automated post-processing step if desired.
