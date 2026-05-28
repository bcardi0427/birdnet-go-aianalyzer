#!/usr/bin/env bash

# 1. Define your container resource footprint
APP="BirdNET-Go-AIAnalyzer"
var_tags="audio;monitoring"
var_cpu="4"
var_ram="2048"
var_disk="12"
var_os="debian"
var_version="13"

# 2. Point the framework to your custom install script path
# The build.func orchestrator looks for the file named "${APP,,}-install.sh" 
# inside an 'install' folder relative to its execution, OR you can override it
# by downloading your script as the target installation script right before building.
URL_INSTALL="https://raw.githubusercontent.com/bcardi0427/birdnet-go-aianalyzer/main/scripts/lxc-install.sh"

# 3. Pull down and load the master orchestrator script 
# (This automatically loads core.func, api.func, and sets up build_config/variables safely)
source <(curl -fsSL https://raw.githubusercontent.com/community-scripts/ProxmoxVE/main/misc/build.func)

# 4. Initialize the standard system variables and menus
header_info "$APP"
variables
color
catch_errors

# 5. Launch the provisioning wizard loop
start

# 6. Dynamically swap in your custom install logic into the framework's build directory
# This forces the builder to use your script instead of looking for it on the community repo
function build_container() {
  # We override the build_container function temporarily to fetch your specific URL
  # right before the container executes it.
  mkdir -p ./install
  curl -fsSL "$URL_INSTALL" -o "./install/${APP,,}-install.sh"
  chmod +x "./install/${APP,,}-install.sh"
  
  # Now call the original build logic embedded inside build.func
  # (Which handles container creation, networking, storage, and runs your script)
  _build_container
}

# 7. Execute the construction loop
build_container
description

msg_ok "Completed successfully!\n"