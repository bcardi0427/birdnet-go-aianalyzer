#!/usr/bin/env bash

# Hardcode the target execution logic to point to YOUR standalone repo
export APP="BirdNET-Go-AIAnalyzer"
var_tags="audio;monitoring"
var_cpu="4"                  # Updated to match your current resource footprint
var_ram="2048"
var_disk="12"                # Updated to match your current 12G root disk
var_os="debian"
var_version="13"             # Updated to match Debian 13 (Trixie)

# Pull in the community script system functions dynamically
coronation_functions=$(curl -fsSL https://raw.githubusercontent.com/community-scripts/ProxmoxVE/main/misc/build.func)
eval "$coronation_functions"

# Set up standard container configuration variables
build_config
variables

# Define the target installation code source (your raw install script)
export URL_INSTALL="https://raw.githubusercontent.com/bcardi0427/birdnet-go-aianalyzer/main/scripts/lxc-install.sh"

# Execute the core container deployment loop
extract_func
lxc_container
