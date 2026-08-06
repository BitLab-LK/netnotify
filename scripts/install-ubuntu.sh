#!/usr/bin/env bash
set -Eeuo pipefail

REPO="BitLab-LK/netnotify"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/netnotify"
STATE_DIR="/var/lib/netnotify"
LOG_DIR="/var/log/netnotify"
SERVICE_FILE="/etc/systemd/system/netnotify.service"
ENV_FILE="${CONFIG_DIR}/netnotify.env"
CONFIG_FILE="${CONFIG_DIR}/config.yaml"
TEMPLATE_DIR="${CONFIG_DIR}/templates"
NETDATA_NOTIFY_DIR="/etc/netdata/custom-plugins.d"
NETDATA_NOTIFY_FILE="${NETDATA_NOTIFY_DIR}/netnotify-health-alarm-notify.sh"

require_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    echo "This installer must run as root. Re-run with sudo." >&2
    exit 1
  fi
}

is_ubuntu() {
  [[ -r /etc/os-release ]] && . /etc/os-release && [[ "${ID:-}" == "ubuntu" || "${ID_LIKE:-}" == *"ubuntu"* || "${ID_LIKE:-}" == *"debian"* ]]
}

prompt() {
  local name="$1" label="$2" default="${3:-}" secret="${4:-false}" value
  if [[ ! -t 0 && ! -c /dev/tty ]]; then
    if [[ -n "${default}" ]]; then
      printf -v "${name}" '%s' "${default}"
      return 0
    else
      echo "Error: '${label}' is required but terminal is non-interactive and no default value was supplied." >&2
      exit 1
    fi
  fi

  while true; do
    if [[ "${secret}" == "true" ]]; then
      read -r -s -p "${label}${default:+ [${default}]}: " value < /dev/tty
      echo >&2
    else
      read -r -p "${label}${default:+ [${default}]}: " value < /dev/tty
    fi
    value="${value:-${default}}"
    if [[ -n "${value}" ]]; then
      printf -v "${name}" '%s' "${value}"
      return 0
    fi
    echo "A value is required." >&2
  done
}

prompt_choice() {
  local name="$1" label="$2" default="$3" value
  if [[ ! -t 0 && ! -c /dev/tty ]]; then
    printf -v "${name}" '%s' "${default}"
    return 0
  fi

  while true; do
    read -r -p "${label} [user/group, default: ${default}]: " value < /dev/tty
    value="${value:-${default}}"
    case "${value}" in
      user|group) printf -v "${name}" '%s' "${value}"; return 0 ;;
      *) echo "Enter 'user' or 'group'." >&2 ;;
    esac
  done
}

install_packages() {
  export DEBIAN_FRONTEND=noninteractive
  apt-get update -y
  apt-get install -y ca-certificates curl tar gzip
}

arch_name() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    *) echo "Unsupported CPU architecture: $(uname -m)" >&2; exit 1 ;;
  esac
}

download_binary() {
  local arch version asset url tmp
  arch="$(arch_name)"
  tmp="$(mktemp -d)"
  trap "rm -rf '${tmp}'" EXIT

  if [[ -f "./netnotify" ]]; then
    echo "Installing local netnotify binary..."
    install -m 0755 "./netnotify" "${INSTALL_DIR}/netnotify"
    return 0
  elif [[ -f "./build/netnotify" ]]; then
    echo "Installing built netnotify binary..."
    install -m 0755 "./build/netnotify" "${INSTALL_DIR}/netnotify"
    return 0
  fi

  version="${NETNOTIFY_VERSION:-latest}"
  if [[ "${version}" == "latest" ]]; then
    url="https://github.com/${REPO}/releases/latest/download/netnotify_linux_${arch}.tar.gz"
  else
    url="https://github.com/${REPO}/releases/download/${version}/netnotify_linux_${arch}.tar.gz"
  fi
  asset="${tmp}/netnotify.tar.gz"
  echo "Downloading netnotify from ${url}"
  if curl -fL --retry 3 --connect-timeout 10 -o "${asset}" "${url}"; then
    tar -xzf "${asset}" -C "${tmp}"
    install -m 0755 "$(find "${tmp}" -type f -name netnotify | head -n 1)" "${INSTALL_DIR}/netnotify"
    return 0
  fi

  if command -v go >/dev/null 2>&1 && [[ -f "./cmd/netnotify/main.go" ]]; then
    echo "Release asset download failed. Building netnotify from source using go..."
    go build -o "${INSTALL_DIR}/netnotify" ./cmd/netnotify
    return 0
  fi

  echo "Could not download release asset from ${url}." >&2
  echo "No published release was found for ${REPO}. Please publish a GitHub release or build the binary locally." >&2
  exit 1
}

create_user_and_dirs() {
  if ! id netnotify >/dev/null 2>&1; then
    useradd --system --home "${STATE_DIR}" --shell /usr/sbin/nologin netnotify
  fi
  install -d -m 0755 "${CONFIG_DIR}" "${TEMPLATE_DIR}"
  install -d -m 0750 -o netnotify -g netnotify "${STATE_DIR}" "${LOG_DIR}"
}

write_templates() {
  for severity in critical warning clear test; do
    cat > "${TEMPLATE_DIR}/${severity}.tmpl" <<'TEMPLATE'
[{{ .Severity }}] {{ .Title }}

{{ .Summary }}
{{ if .Text }}{{ .Text }}{{ end }}
Source: {{ .Source }}
Status: {{ .Status }}
Started: {{ .StartsAt.Format "2006-01-02 15:04:05 UTC" }}
{{ range $k, $v := .Labels }}{{ $k }}={{ $v }}
{{ end }}
TEMPLATE
  done
  chmod 0644 "${TEMPLATE_DIR}"/*.tmpl
}

write_config() {
  local group_bool="false"
  [[ "${RECEIVER_TYPE}" == "group" ]] && group_bool="true"
  cat > "${CONFIG_FILE}" <<EOF_CONFIG
server:
  address: "${LISTEN_ADDRESS}"
  metrics: true
  basic_auth:
    enabled: false
    username: ""
    password: ""
log:
  level: info
  format: json
  file: "${LOG_DIR}/netnotify.log"
  max_size_mb: 50
  max_backups: 5
  max_age_days: 30
queue:
  workers: 4
  size: 1000
  retries: 3
  retry_interval: 5s
  dedup_ttl: 5m
  rate_per_second: 10
providers:
  default: gowa
  gowa:
    enabled: true
    base_url: "${GOWA_URL}"
    username: "\${NETNOTIFY_PROVIDERS_GOWA_USERNAME}"
    password: "\${NETNOTIFY_PROVIDERS_GOWA_PASSWORD}"
    device_id: "${GOWA_DEVICE_ID}"
    recipient: "${RECEIVER_ID}"
    group: ${group_bool}
    timeout: 10s
    tls_skip_verify: false
sources:
  netdata:
    provider: gowa
    default_recipient: "${RECEIVER_ID}"
templates:
  directory: "${TEMPLATE_DIR}"
EOF_CONFIG
  chmod 0640 "${CONFIG_FILE}"
  chown root:netnotify "${CONFIG_FILE}"

  cat > "${ENV_FILE}" <<EOF_ENV
NETNOTIFY_PROVIDERS_GOWA_USERNAME=${GOWA_USERNAME}
NETNOTIFY_PROVIDERS_GOWA_PASSWORD=${GOWA_PASSWORD}
EOF_ENV
  chmod 0600 "${ENV_FILE}"
  chown root:root "${ENV_FILE}"
}

write_systemd() {
  cat > "${SERVICE_FILE}" <<EOF_SERVICE
[Unit]
Description=netnotify notification gateway
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=netnotify
Group=netnotify
EnvironmentFile=-${ENV_FILE}
ExecStart=${INSTALL_DIR}/netnotify --config ${CONFIG_FILE}
Restart=on-failure
RestartSec=5s
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${STATE_DIR} ${LOG_DIR}

[Install]
WantedBy=multi-user.target
EOF_SERVICE
  systemctl daemon-reload
  systemctl enable --now netnotify
}

write_netdata_helper() {
  install -d -m 0755 "${NETDATA_NOTIFY_DIR}"
  cat > "${NETDATA_NOTIFY_FILE}" <<'EOF_NETDATA'
#!/usr/bin/env sh
set -eu
: "${NETNOTIFY_URL:=http://127.0.0.1:8080/v1/sources/netdata}"
curl -fsS -X POST "${NETNOTIFY_URL}" \
  --data-urlencode "alarm=${alarm:-${ALARM:-}}" \
  --data-urlencode "status=${status:-${STATUS:-}}" \
  --data-urlencode "hostname=${hostname:-${HOSTNAME:-}}" \
  --data-urlencode "chart=${chart:-${CHART:-}}" \
  --data-urlencode "family=${family:-${FAMILY:-}}" \
  --data-urlencode "summary=${info:-${INFO:-}}"
EOF_NETDATA
  chmod 0755 "${NETDATA_NOTIFY_FILE}"
}

main() {
  require_root
  if ! is_ubuntu; then
    echo "Warning: this installer is designed for Ubuntu/Debian systems." >&2
  fi

  echo "netnotify Ubuntu installer"
  prompt GOWA_URL "GoWA base URL, including https:// and port if needed" "${NETNOTIFY_GOWA_URL:-}"
  prompt GOWA_USERNAME "GoWA Basic Auth username" "${NETNOTIFY_GOWA_USERNAME:-}"
  prompt GOWA_PASSWORD "GoWA Basic Auth password" "${NETNOTIFY_GOWA_PASSWORD:-}" true
  prompt GOWA_DEVICE_ID "GoWA device ID (press Enter for default device)" "${NETNOTIFY_GOWA_DEVICE_ID:-default}"
  [[ "${GOWA_DEVICE_ID}" == "default" ]] && GOWA_DEVICE_ID=""
  prompt_choice RECEIVER_TYPE "WhatsApp receiver type" "${NETNOTIFY_RECEIVER_TYPE:-group}"
  prompt RECEIVER_ID "WhatsApp receiver ID (group JID or phone/JID)" "${NETNOTIFY_RECEIVER_ID:-}"
  prompt LISTEN_ADDRESS "netnotify listen address" "${NETNOTIFY_LISTEN_ADDRESS:-127.0.0.1:8080}"

  install_packages
  create_user_and_dirs
  download_binary
  write_templates
  write_config
  write_systemd
  write_netdata_helper

  echo
  echo "netnotify is installed and running."
  echo "Config: ${CONFIG_FILE}"
  echo "Secrets: ${ENV_FILE}"
  echo "Netdata helper: ${NETDATA_NOTIFY_FILE}"
  echo "Health check: curl -fsS http://${LISTEN_ADDRESS}/health"
}

main "$@"
