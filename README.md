# netnotify

Lightweight, Dedicated Healthchecks.io Heartbeat Daemon for Servers & VMs.

[![CI](https://github.com/BitLab-LK/netnotify/actions/workflows/ci.yml/badge.svg)](https://github.com/BitLab-LK/netnotify/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

`netnotify` is a dedicated single-purpose heartbeat daemon for servers and VMs. It periodically pings Healthchecks.io (or any uptime heartbeat service) to act as a **Dead Man's Switch**, alerting you immediately if your VM goes offline or crashes.

```mermaid
flowchart LR
  A[netnotify Daemon] -->|Startup Ping| B[Healthchecks.io]
  A -->|Every 1m Ping| B
  A -->|On Shutdown: GET /stop| B
```

## Features

- **Automated Heartbeats:** Sends startup pings and periodic pings (default `1m`) to Healthchecks.io.
- **Graceful Shutdown Signal:** Sends a `/stop` signal on `systemctl stop` to prevent false alerts during maintenance.
- **Local Health Endpoint:** Exposes `http://127.0.0.1:8080/health` and `/metrics` for local checks.
- **Hardened systemd Service:** Includes systemd service configuration with `0600` secret security.
- **1-Click Installer:** Interactive setup for Ubuntu / Debian.

## One-line Ubuntu Install

Run the installer on your Ubuntu VM:

```bash
curl -fsSL https://raw.githubusercontent.com/BitLab-LK/netnotify/main/scripts/install-ubuntu.sh | sudo bash
```

The installer prompts for your **Healthchecks.io Ping URL** and **Ping Interval**, writes `/etc/netnotify/config.yaml`, stores secrets in `/etc/netnotify/netnotify.env`, and starts the hardened systemd service.

## Quick Start

```bash
cp configs/config.example.yaml config.yaml
export NETNOTIFY_HEARTBEAT_URL="https://hc-ping.com/your-uuid-here"
go run ./cmd/netnotify --config config.yaml
```

Test a manual heartbeat ping:

```bash
go run ./cmd/netnotify ping --config config.yaml
```

Check local health endpoint:

```bash
curl -fsS http://127.0.0.1:8080/health
```

## License

MIT.
