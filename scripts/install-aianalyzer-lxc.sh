#!/usr/bin/env bash
set -euo pipefail

# Upgrade a Proxmox helper-script BirdNET-Go LXC to the AI Analyzer fork.
# This preserves /opt/birdnet/data and only replaces the native binary.

REPO_OWNER="${REPO_OWNER:-bcardi0427}"
REPO_NAME="${REPO_NAME:-birdnet-go-aianalyzer}"
RELEASE_TAG="${RELEASE_TAG:-latest}"

# Auto-detect architecture for the default asset name
ARCH="$(uname -m)"
case "${ARCH}" in
  x86_64)
    DEFAULT_ASSET="birdnet-go-linux-amd64"
    ;;
  aarch64|arm64)
    DEFAULT_ASSET="birdnet-go-linux-arm64"
    ;;
  *)
    # Fallback to amd64 if detection is inconclusive
    DEFAULT_ASSET="birdnet-go-linux-amd64"
    ;;
esac

ASSET_NAME="${ASSET_NAME:-${DEFAULT_ASSET}}"


SERVICE_NAME="${SERVICE_NAME:-birdnet}"
BINARY_PATH="${BINARY_PATH:-/usr/local/bin/birdnet-go}"
DATA_DIR="${DATA_DIR:-/opt/birdnet/data}"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

fail() {
  echo "ERROR: $*" >&2
  exit 1
}

require_root() {
  if [ "$(id -u)" -ne 0 ]; then
    fail "Run this script as root inside the BirdNET-Go LXC."
  fi
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || fail "Missing required command: $1"
}

resolve_download_url() {
  if [ "${RELEASE_TAG}" = "latest" ]; then
    echo "https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest/download/${ASSET_NAME}"
  else
    echo "https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${RELEASE_TAG}/${ASSET_NAME}"
  fi
}

require_root
require_command curl
require_command systemctl

[ -d "${DATA_DIR}" ] || fail "Expected data directory not found: ${DATA_DIR}"
[ -f "${BINARY_PATH}" ] || fail "Expected BirdNET-Go binary not found: ${BINARY_PATH}"
systemctl cat "${SERVICE_NAME}" >/dev/null 2>&1 || fail "Expected systemd service not found: ${SERVICE_NAME}"

DOWNLOAD_URL="$(resolve_download_url)"
NEW_BINARY="${TMP_DIR}/birdnet-go"
BACKUP_PATH="${BINARY_PATH}.backup.$(date +%Y%m%d-%H%M%S)"

echo "Downloading AI Analyzer binary:"
echo "  ${DOWNLOAD_URL}"

if ! curl -fL "${DOWNLOAD_URL}" -o "${NEW_BINARY}"; then
  cat >&2 <<EOF

Could not download the AI Analyzer binary.

Expected release asset:
  ${ASSET_NAME}

Expected source:
  ${DOWNLOAD_URL}

This usually means a GitHub release/build artifact has not been published yet.
EOF
  exit 1
fi

chmod 755 "${NEW_BINARY}"

echo "Stopping ${SERVICE_NAME}.service"
systemctl stop "${SERVICE_NAME}"

echo "Backing up current binary:"
echo "  ${BACKUP_PATH}"
cp "${BINARY_PATH}" "${BACKUP_PATH}"

echo "Installing AI Analyzer binary:"
echo "  ${BINARY_PATH}"
install -m 755 "${NEW_BINARY}" "${BINARY_PATH}"

echo "Starting ${SERVICE_NAME}.service"
systemctl start "${SERVICE_NAME}"

echo
systemctl status "${SERVICE_NAME}" --no-pager

echo
echo "AI Analyzer binary install completed."
echo "Previous binary backup:"
echo "  ${BACKUP_PATH}"
