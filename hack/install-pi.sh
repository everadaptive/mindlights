#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
INSTALL_DIR="/home/pi/Desktop/mindlights"

echo "Installing dependencies..."
sudo apt-get install -y libusb-1.0-0 libftdi1-2

echo "Installing mindlights to $INSTALL_DIR..."

mkdir -p "$INSTALL_DIR"

cp "$SCRIPT_DIR/stream" "$INSTALL_DIR/stream"
cp "$SCRIPT_DIR/headset-02-attention.yaml" "$INSTALL_DIR/headset-02-attention.yaml"
cp "$SCRIPT_DIR/run-pi.sh" "$INSTALL_DIR/run-pi.sh"
cp -r "$SCRIPT_DIR/app" "$INSTALL_DIR/app"

chmod +x "$INSTALL_DIR/stream" "$INSTALL_DIR/run-pi.sh"

cat > /home/pi/Desktop/mindlights.desktop <<EOF
[Desktop Entry]
Name=Mindlights
Comment=EEG Stream Server
Exec=bash -c '$INSTALL_DIR/run-pi.sh; read -p "Press enter to close"'
Icon=utilities-terminal
Terminal=true
Type=Application
EOF

chmod +x /home/pi/Desktop/mindlights.desktop

echo "Done. Double-click 'Mindlights' on the desktop to start."
