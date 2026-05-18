TRex Stateless Profile Template

Purpose
- Template to create a TRex stateless YAML profile matching DUT NIC topology.

Usage
- Copy this file to `benchmarks/trex/my_profile.yaml` and fill interface indexes and desired traffic flows.

Example structure

```yaml
# Replace interface mapping and payloads as needed
profile_name: afxdp-line-rate-test
interfaces:
  - 0  # generator NIC port 0
  - 1  # generator NIC port 1
streams:
  - name: ipv4-l3-load
    packet:
      - Ethernet:
          dst: 00:11:22:33:44:55
          src: aa:bb:cc:dd:ee:ff
      - IPv4:
          src: 192.0.2.1
          dst: 198.51.100.1
      - UDP:
          src_port: 1234
          dst_port: 443
    mode:
      type: continuous
      rate: 10000Mbps
    size: 1500
    flow-variables:
      src_ip: {type: inc, start: 192.0.2.1, step: 1}
      dst_ip: {type: fixed, value: 198.51.100.1}

# Add more streams for bi-directional tests and latency measurements
```

Hints
- Match `interfaces` ordering to how TRex maps ports to physical NIC ports on the generator host.
- Use smaller `size` and `rate` values for iterative testing, then ramp to target line rate.
- For latency tests, include timestamping payloads or use MoonGen latency scripts instead.
