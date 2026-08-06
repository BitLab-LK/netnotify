#!/usr/bin/env sh
set -eu
: "${NETNOTIFY_URL:=http://127.0.0.1:8080/v1/sources/netdata}"
curl -fsS -X POST "$NETNOTIFY_URL" \
  --data-urlencode "alarm=${alarm:-${ALARM:-}}" \
  --data-urlencode "status=${status:-${STATUS:-}}" \
  --data-urlencode "hostname=${hostname:-${HOSTNAME:-}}" \
  --data-urlencode "chart=${chart:-${CHART:-}}" \
  --data-urlencode "family=${family:-${FAMILY:-}}" \
  --data-urlencode "summary=${info:-${INFO:-}}"
