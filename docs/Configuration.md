# Configuration

`netnotify` configuration is loaded from YAML and can be overridden by environment variables and CLI flags.

## Example YAML Configuration

```yaml
server:
  address: "127.0.0.1:8085"
  metrics: true

log:
  level: "info"
  format: "json"
  file: "/var/log/netnotify/netnotify.log"

heartbeat:
  enabled: true
  url: "https://hc-ping.com/your-uuid-here"
  interval: "1m"
  send_stop_on_exit: true
```

## Environment Variables

| Variable | Description | Default |
| :--- | :--- | :--- |
| `NETNOTIFY_HEARTBEAT_URL` | Healthchecks.io Ping URL | `""` |
| `NETNOTIFY_HEARTBEAT_INTERVAL` | Ping frequency (e.g. `1m`, `5m`) | `1m` |
| `NETNOTIFY_HEARTBEAT_SEND_STOP_ON_EXIT` | Send `/stop` ping on service shutdown | `true` |
| `NETNOTIFY_SERVER_ADDRESS` | Local listen address for `/health` | `127.0.0.1:8085` |
