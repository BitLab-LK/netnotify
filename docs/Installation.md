# Installation

## One-line Ubuntu Installation

Run the interactive installer on an Ubuntu/Debian system:

```bash
curl -fsSL https://raw.githubusercontent.com/BitLab-LK/netnotify/main/scripts/install-ubuntu.sh | sudo bash
```

The installer prompts for:
- **Healthchecks.io Ping URL** (e.g. `https://hc-ping.com/your-uuid-here`).
- **Ping Interval** (default `1m`).
- **Local listen address** (default `127.0.0.1:8080`).

It installs the latest Linux binary, creates the locked-down system user `netnotify`, sets permissions (`0600` for environment secrets), and starts the systemd service.

## Non-interactive / Automated Install

To run headlessly without interactive prompts, pass environment variables:

```bash
sudo env \
  NETNOTIFY_HEARTBEAT_URL="https://hc-ping.com/your-uuid-here" \
  NETNOTIFY_HEARTBEAT_INTERVAL="1m" \
  bash -c "$(curl -fsSL https://raw.githubusercontent.com/BitLab-LK/netnotify/main/scripts/install-ubuntu.sh)"
```

## Verification

```bash
systemctl status netnotify
curl -fsS http://127.0.0.1:8080/health
netnotify ping --config /etc/netnotify/config.yaml
```
