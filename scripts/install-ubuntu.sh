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
  install -d -m 0755 "${CONFIG_DIR}"
  install -d -m 0750 -o netnotify -g netnotify "${STATE_DIR}" "${LOG_DIR}"
}

write_config() {
  cat > "${CONFIG_FILE}" <<EOF_CONFIG
server:
  address: "${LISTEN_ADDRESS}"
  metrics: true
log:
  level: info
  format: json
  file: "${LOG_DIR}/netnotify.log"
  max_size_mb: 50
  max_backups: 5
  max_age_days: 30
heartbeat:
  enabled: true
  url: "${HEARTBEAT_URL}"
  interval: "${HEARTBEAT_INTERVAL}"
  send_stop_on_exit: true
EOF_CONFIG
  chmod 0640 "${CONFIG_FILE}"
  chown root:netnotify "${CONFIG_FILE}"

  cat > "${ENV_FILE}" <<EOF_ENV
NETNOTIFY_HEARTBEAT_URL=${HEARTBEAT_URL}
NETNOTIFY_HEARTBEAT_INTERVAL=${HEARTBEAT_INTERVAL}
EOF_ENV
  chmod 0600 "${ENV_FILE}"
  chown root:root "${ENV_FILE}"
}

write_systemd() {
  cat > "${SERVICE_FILE}" <<EOF_SERVICE
[Unit]
Description=netnotify Healthchecks.io Heartbeat Daemon
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
  systemctl restart netnotify || systemctl enable --now netnotify
}

main() {
  require_root
  if ! is_ubuntu; then
    echo "Warning: this installer is designed for Ubuntu/Debian systems." >&2
  fi

  if [[ -r "${ENV_FILE}" ]]; then
    set -a
    source "${ENV_FILE}" 2>/dev/null || true
    set +a
  fi
  if [[ -r "${CONFIG_FILE}" ]]; then
    local cfg_addr cfg_hb_url cfg_hb_interval
    cfg_addr="$(grep -E '^\s*address:' "${CONFIG_FILE}" | head -n 1 | awk '{print $2}' | tr -d '"')"
    cfg_hb_url="$(grep -E '^\s*url:' "${CONFIG_FILE}" | head -n 1 | awk '{print $2}' | tr -d '"')"
    cfg_hb_interval="$(grep -E '^\s*interval:' "${CONFIG_FILE}" | head -n 1 | awk '{print $2}' | tr -d '"')"
    
    NETNOTIFY_LISTEN_ADDRESS="${NETNOTIFY_LISTEN_ADDRESS:-${cfg_addr:-127.0.0.1:8080}}"
    NETNOTIFY_HEARTBEAT_URL="${NETNOTIFY_HEARTBEAT_URL:-${cfg_hb_url:-}}"
    NETNOTIFY_HEARTBEAT_INTERVAL="${NETNOTIFY_HEARTBEAT_INTERVAL:-${cfg_hb_interval:-1m}}"
  fi

  echo "netnotify Healthchecks.io Installer"
  prompt HEARTBEAT_URL "Healthchecks.io Ping URL (e.g. https://hc-ping.com/your-uuid)" "${NETNOTIFY_HEARTBEAT_URL:-}"
  prompt HEARTBEAT_INTERVAL "Ping interval" "${NETNOTIFY_HEARTBEAT_INTERVAL:-1m}"
  prompt LISTEN_ADDRESS "netnotify local listen address" "${NETNOTIFY_LISTEN_ADDRESS:-127.0.0.1:8080}"

  install_packages
  create_user_and_dirs
  download_binary
  write_config
  write_systemd

  echo
  echo "netnotify Healthchecks.io Heartbeat Daemon is updated and running!"
  echo "Config: ${CONFIG_FILE}"
  echo "Secrets: ${ENV_FILE}"
  echo "Health check: curl -fsS http://${LISTEN_ADDRESS}/health"
}

main "$@"
