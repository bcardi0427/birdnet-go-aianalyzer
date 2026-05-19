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
