#!/usr/bin/env sh
set -eu
systemctl disable --now netnotify 2>/dev/null || true
rm -f /etc/systemd/system/netnotify.service /usr/local/bin/netnotify
systemctl daemon-reload 2>/dev/null || true
