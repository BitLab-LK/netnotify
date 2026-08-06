#!/usr/bin/env sh
set -eu
install -d -m 0755 /etc/netnotify /var/lib/netnotify /var/log/netnotify
id netnotify >/dev/null 2>&1 || useradd --system --home /var/lib/netnotify --shell /usr/sbin/nologin netnotify
install -m 0755 netnotify /usr/local/bin/netnotify
install -m 0644 configs/config.example.yaml /etc/netnotify/config.yaml
install -m 0644 build/systemd/netnotify.service /etc/systemd/system/netnotify.service
systemctl daemon-reload
systemctl enable netnotify
