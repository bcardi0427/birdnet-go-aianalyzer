#!/usr/bin/env bash

# 1. Define your container resource footprint
export APP="BirdNET-Go-AIAnalyzer"
export var_tags="audio;monitoring"
export var_cpu="4"
export var_ram="2048"
export var_disk="12"
export var_os="debian"
export var_version="13"

# 2. Tell the builder exactly where your installation logic lives
export URL_INSTALL="https://raw.githubusercontent.com/bcardi0427/birdnet-go-aianalyzer/main/scripts/lxc-install.sh"

# 3. Explicitly pull and load the required community functions in order
# This manually structures the build environment instead of relying on their broken relative paths
source <(curl -fsSL https://raw.githubusercontent.com/community-scripts/ProxmoxVE/main/misc/api.func)
source <(curl -fsSL https://raw.githubusercontent.com/community-scripts/ProxmoxVE/main/misc/core.func)

# 4. Trigger the standard script system variables
build_config
variables

# 5. Execute the automated container construction loop
extract_func
lxc_container