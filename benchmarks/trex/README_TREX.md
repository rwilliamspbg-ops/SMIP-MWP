TRex quick notes

This directory contains guidance for running a stateless TRex scenario for raw throughput.

Typical TRex run (on traffic generator host):

1. Start TRex server (daemon) on traffic generator host:

```bash
sudo ./t-rex-64 -i
```

2. Use a basic stateless profile and run at target rate:

```bash
# Example stateless run at 10000 Mbps between ports 0 and 1
sudo ./trex-console -f my_profile.yaml -m 10000
```

See TRex docs for profile and topology configuration. For our hardware matrix, prefer stateless for max throughput and use stateful for flow-mix validation.
