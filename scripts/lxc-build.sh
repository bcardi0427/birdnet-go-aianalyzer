#!/usr/bin/env bash

# Hardcode the target execution logic to point to YOUR standalone repo
export APP="BirdNET-Go-AIAnalyzer"
export var_tags="audio;monitoring;ai;nature"
export var_cpu="4"                  # Updated to match your current resource footprint
export var_ram="2048"
export var_disk="12"                # Updated to match your current 12G root disk
export var_os="debian"
export var_version="13"             # Updated to match Debian 13 (Trixie)
export var_unprivileged="1"

# Intercept curl to redirect the installation script request to our custom script
curl() {
  if [[ "$*" == *"birdnet-go-aianalyzer-install.sh"* ]]; then
    command curl -fsSL "https://raw.githubusercontent.com/bcardi0427/birdnet-go-aianalyzer/main/scripts/lxc-install.sh"
  else
    command curl "$@"
  fi
}
export -f curl

# Pull in the community script system functions dynamically
source <(command curl -fsSL https://raw.githubusercontent.com/community-scripts/ProxmoxVE/main/misc/build.func)

# Initialize PVE Helper Script lifecycle
header_info "$APP"
variables
color
catch_errors

# Launch the provisioning wizard loop
start
build_container
description

msg_ok "Completed successfully!\n"
echo -e "${CREATING}${GN}${APP} setup has been successfully initialized!${CL}"