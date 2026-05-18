Self-hosted `xdp-runner` setup

This document describes the minimal steps to register and run a self-hosted GitHub Actions runner for AF_XDP integration testing.

1. Register the runner

- Follow GitHub docs to create and register a self-hosted runner for the repository. During registration, assign the label `xdp-runner` so the workflow `.github/workflows/asavie-integration.yml` can target it.

2. Prepare the host

- Copy the repository contents onto the host (clone the repo) and place the provisioning dashboard JSON in `/var/lib/grafana/dashboards/afxdp` if you want Grafana auto-provisioned. The included helper script can do this for you:

```bash
sudo cp -r /path/to/repo /opt/smip-mwp
cd /opt/smip-mwp
sudo ./scripts/setup_xdp_runner.sh
```

3. Register the service (optional)

- Install the example systemd unit to run the setup script at boot (edit paths as needed):

```bash
sudo cp scripts/xdp-runner.service /etc/systemd/system/xdp-runner.service
sudo systemctl daemon-reload
sudo systemctl enable --now xdp-runner.service
```

4. Runner privileges

- AF_XDP requires network privileges. Easiest options:
  - Run the GitHub runner as root (not recommended for multi-tenant hosts). OR
  - Run the runner service with `CAP_NET_RAW` and `CAP_NET_ADMIN` capabilities. Example with systemd `AmbientCapabilities=CAP_NET_RAW CAP_NET_ADMIN`.

5. Trigger the workflow

- Use the Actions UI to run the `asavie-integration` workflow and provide inputs (interface name, run_bench). After the run completes, download artifacts under `benchmarks/**`.

6. Collect artifacts and analyze

- Download uploaded artifacts from the workflow run and use `go tool pprof` to analyze CPU/memory profiles saved under `benchmarks/`.
