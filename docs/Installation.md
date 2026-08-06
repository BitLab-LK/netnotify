# Installation

## One-line Ubuntu installation

Use the public GitHub repository installer:

```bash
curl -fsSL https://raw.githubusercontent.com/BitLab-LK/netnotify/main/scripts/install-ubuntu.sh | sudo bash
```

The installer is interactive and asks for:

- GoWA hosted base URL, for example `https://gowa.example.com`.
- GoWA Basic Auth username.
- GoWA Basic Auth password, entered without terminal echo.
- Optional GoWA device ID for multi-device deployments.
- Message receiver type: `user` or `group`.
- WhatsApp receiver ID, such as a group JID ending in `@g.us` or an individual phone/JID.
- netnotify listen address, defaulting to `127.0.0.1:8080`.

It installs the latest release asset for `linux/amd64` or `linux/arm64`, creates a locked-down `netnotify` system user, writes `/etc/netnotify/config.yaml`, stores secrets in `/etc/netnotify/netnotify.env` with mode `0600`, installs external templates, enables the `netnotify.service` systemd unit, and writes a Netdata-compatible helper script to `/etc/netdata/custom-plugins.d/netnotify-health-alarm-notify.sh`.

## Verify

```bash
systemctl status netnotify
curl -fsS http://127.0.0.1:8080/health
netnotify validate --config /etc/netnotify/config.yaml
```

## Non-interactive defaults

The installer accepts environment defaults while still prompting for missing values:

```bash
sudo env \
  NETNOTIFY_GOWA_URL=https://gowa.example.com \
  NETNOTIFY_GOWA_USERNAME=admin \
  NETNOTIFY_RECEIVER_TYPE=group \
  NETNOTIFY_RECEIVER_ID=120363000000000000@g.us \
  bash -c "$(curl -fsSL https://raw.githubusercontent.com/BitLab-LK/netnotify/main/scripts/install-ubuntu.sh)"
```
This document describes production operation of netnotify.

## Overview

netnotify is a Go service organized around source parsers, domain notifications, a queueing application core and provider notifiers. Configuration is loaded from YAML and can be overridden by environment variables and CLI flags.

## Operational guidance

- Run as an unprivileged user.
- Store provider credentials in environment variables or a root-readable environment file.
- Keep TLS verification enabled unless using a controlled lab with private certificates.
- Monitor /health and /metrics.
- Validate configuration before restart with `netnotify validate --config /etc/netnotify/config.yaml`.
