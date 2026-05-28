#!/usr/bin/env bash
set -euo pipefail

# BirdNET-Go AI Analyzer fork installer wrapper.
# This runs the upstream-style installer from this fork branch so users get fork updates.

REPO_OWNER="${REPO_OWNER:-bcardi0427}"
REPO_NAME="${REPO_NAME:-birdnet-go-aianalyzer}"
REPO_BRANCH="${REPO_BRANCH:-aianalyzer/main}"

INSTALL_URL="https://raw.githubusercontent.com/${REPO_OWNER}/${REPO_NAME}/${REPO_BRANCH}/install.sh"
TMP_INSTALL_SCRIPT="/tmp/birdnet-go-install.sh"

echo "Downloading installer from:"
echo "  ${INSTALL_URL}"

curl -fsSL "${INSTALL_URL}" -o "${TMP_INSTALL_SCRIPT}"
bash "${TMP_INSTALL_SCRIPT}" "$@"

# Add local hostname resolution helper for birdnet-go.local
if [ "$(id -u)" -eq 0 ]; then
    if ! grep -q "birdnet-go.local" /etc/hosts; then
        echo "Adding birdnet-go.local to /etc/hosts..."
        echo "127.0.0.1 birdnet-go.local" >> /etc/hosts
        echo "Successfully added local hostname resolution!"
    fi
else
    echo ""
    echo "Tip: To access the dashboard via http://birdnet-go.local:8080,"
    echo "     re-run this installer with sudo, or manually add this line to your /etc/hosts file:"
    echo "     127.0.0.1 birdnet-go.local"
fi

