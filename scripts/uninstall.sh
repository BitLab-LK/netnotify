#!/usr/bin/env bash
set -Eeuo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "This script must run as root. Re-run with sudo." >&2
  exit 1
fi

echo "Stopping and disabling netnotify systemd service..."
systemctl disable --now netnotify 2>/dev/null || true

echo "Removing netnotify files and directories..."
rm -f /etc/systemd/system/netnotify.service
rm -f /usr/local/bin/netnotify
rm -rf /etc/netnotify
rm -rf /var/log/netnotify
rm -rf /var/lib/netnotify
rm -f /etc/netdata/custom-plugins.d/netnotify-health-alarm-notify.sh

systemctl daemon-reload 2>/dev/null || true

if id netnotify >/dev/null 2>&1; then
  echo "Removing netnotify system user..."
  userdel netnotify 2>/dev/null || true
fi

echo "netnotify has been completely uninstalled from this VM."
