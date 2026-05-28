#!/usr/bin/env bash

# Colors for output matching upstream layout
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

set -e

print_message() {
    echo -e "${2}${1}${NC}"
}

# ASCII Art Banner
cat << "EOF"
 ____  _         _ _   _ _____ _____    ____      
| __ )(_)_ __ __| | \ | | ____|_   _|  / ___| ___ 
|  _ \| | '__/ _` |  \| |  _|   | |   | |  _ / _ \
| |_) | | | | (_| | |\  | |___  | |   | |_| | (_) |
|____/|_|_|  \__,_|_| \_|_____| |_|    \____|\___/ 
EOF

print_message "\n🔧 Updating base container packages..." "$YELLOW"
apt-get update && apt-get upgrade -y

print_message "\n🔧 Installing base system packages and audio decoders..." "$YELLOW"
apt-get install -y alsa-utils curl bc jq apache2-utils ffmpeg sox libasound2 wget tar

# Core workspace paths matching native system definitions
CONFIG_DIR="/etc/birdnet-go"
DATA_DIR="/var/lib/birdnet-go"
CONFIG_FILE="$CONFIG_DIR/config.yaml"

mkdir -p "$CONFIG_DIR"
mkdir -p "$DATA_DIR/clips"

# 📥 DOWNLOAD BASE CONFIGURATION FROM YOUR PROJECT
print_message "\n📥 Downloading base configuration file template..." "$YELLOW"
curl -s --fail https://raw.githubusercontent.com/bcardi0427/birdnet-go-aianalyzer/main/internal/conf/config.yaml > "$CONFIG_FILE"

# 🎥 CONFIGURATION STEP: WEB PORT
WEB_PORT=8080
sed -i -E '/webserver:/,/port:/ s/(port:\s*)[0-9]+/\1'$WEB_PORT'/' "$CONFIG_FILE"

# 🔊 CONFIGURATION STEP: AUDIO FORMAT EXPORT
print_message "\n🔊 Audio Export Configuration" "$GREEN"
print_message "Select audio format for captured sounds:" "$NC"
print_message "1) WAV  2) FLAC  3) AAC (Default)  4) MP3  5) Opus" "$YELLOW"
read -r -p "Select format (1-5) [3]: " format_choice
case ${format_choice:-3} in
    1) format="wav";; 2) format="flac";; 3) format="aac";; 4) format="mp3";; 5) format="opus";;
esac
sed -i "s/type: wav/type: $format/" "$CONFIG_FILE"

# 🌍 CONFIGURATION STEP: LOCATION (API AUTOMATION)
print_message "\n🌍 Attempting automatic location detection via IP..." "$YELLOW"
if ip_location=$(curl -s "https://ipapi.co/json/" 2>/dev/null) && [ -n "$ip_location" ]; then
    lat=$(echo "$ip_location" | jq -r '.latitude')
    lon=$(echo "$ip_location" | jq -r '.longitude')
    city=$(echo "$ip_location" | jq -r '.city')
    tz=$(echo "$ip_location" | jq -r '.timezone')
    print_message "📍 Auto-detected Location: $city (Lat: $lat, Lon: $lon, TZ: $tz)" "$GREEN"
else
    lat="0.0" && lon="0.0" && tz="UTC"
    print_message "⚠️ Defaulting location to 0.0, 0.0 (UTC)" "$YELLOW"
fi
sed -i "s/latitude: 00.000/latitude: $lat/" "$CONFIG_FILE"
sed -i "s/longitude: 00.000/longitude: $lon/" "$CONFIG_FILE"

# 🔒 CONFIGURATION STEP: BASIC AUTH
print_message "\n🔒 Security Configuration" "$GREEN"
read -r -p "Enable password protection for the interface? (y/n): " enable_auth
if [[ $enable_auth == "y" ]]; then
    read -s -r -p "Enter password: " password
    printf '\n'
    # Enable basic auth and set the plain text password (birdnet-go will automatically encrypt it on startup)
    sed -i '/basicauth:/,/password:/ s/enabled: false/enabled: true/' "$CONFIG_FILE"
    sed -i "s|password: \"\"|password: \"$password\"|" "$CONFIG_FILE"
fi

# ⚙️ INFERENCE OPTIMIZATION SEEDING
sed -i 's/usexnnpack: false/usexnnpack: true/' "$CONFIG_FILE"

# 📦 FETCH NATIVE COMPILED BINARY (Direct executable, no Docker overhead)
print_message "\n📦 Deploying native compiled BirdNET-Go executable binary..." "$YELLOW"
wget -q https://github.com/bcardi0427/birdnet-go-aianalyzer/releases/latest/download/birdnet-go-linux-amd64
chmod +x birdnet-go-linux-amd64
mv birdnet-go-linux-amd64 /opt/birdnet-go-executable

# 🚀 SYSTEMD SERVICE GENERATION
print_message "\n🚀 Generating native systemd daemon layout..." "$YELLOW"
cat << EOF > /etc/systemd/system/birdnet-go.service
[Unit]
Description=BirdNET-Go AI Analyzer Native Service
After=network.target sound.target

[Service]
Type=simple
User=root
WorkingDirectory=/var/lib/birdnet-go
Environment=TZ="${tz}"
ExecStart=/opt/birdnet-go-executable -config /etc/birdnet-go/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

print_message "\n🚀 Igniting system daemon channels..." "$YELLOW"
systemctl daemon-reload
systemctl enable --now birdnet-go.service

print_message "\n🎉 Installation complete! Native BirdNET-Go is running smoothly on Debian 13." "$GREEN"