#!/usr/bin/env bash
set -euo pipefail

# Minimal setup script for an xdp-runner host. Run as root or with sudo.

echo "Installing prerequisites..."
apt-get update
apt-get install -y build-essential git clang llvm libelf-dev libbpf-dev pkg-config iproute2 curl

echo "Ensure Go is installed (1.26+). If not, install manually from golang.org or your distro package manager."

echo "Creating Grafana dashboard provisioning directory..."
mkdir -p /var/lib/grafana/dashboards/afxdp

echo "Copying dashboard JSON (if present in repo)..."
if [ -f "$(pwd)/docs/grafana/provisioning/afxdp-dashboard.json" ]; then
  cp "$(pwd)/docs/grafana/provisioning/afxdp-dashboard.json" /var/lib/grafana/dashboards/afxdp/
  echo "Dashboard copied."
else
  echo "No dashboard JSON found in repo path docs/grafana/provisioning/afxdp-dashboard.json"
fi

echo "Runner setup complete. Register a GitHub self-hosted runner and label it 'xdp-runner'."

echo "Note: To allow AF_XDP socket creation, run the runner as root or grant CAP_NET_RAW/CAP_NET_ADMIN to the runner process."

exit 0
