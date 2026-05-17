Ansible playbook to prepare bare-metal hosts for throughput testing.

Usage:
  # edit inventory with sender/receiver hosts
  ansible-playbook -i inventory.ini playbook.yml -e iface=eth0 -e launch_cmd='./cmd/mohawk-node --iface=eth0 --metrics-addr=:9090'

Notes:
 - The playbook assumes Ubuntu-compatible hosts. Adjust package list for other OSes.
 - The playbook runs `scripts/max_throughput_run.sh` on each host; ensure the repo is present on the target host path matching the playbook location.
