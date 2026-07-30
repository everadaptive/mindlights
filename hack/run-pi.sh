#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
chmod +x "$SCRIPT_DIR/stream" "$SCRIPT_DIR/run-pi.sh" 2>/dev/null || true
CONFIG="${1:-$SCRIPT_DIR/headset-02-attention.yaml}"
APP_DIR="$SCRIPT_DIR/app"
BINARY="$SCRIPT_DIR/stream"

if [ ! -f "$BINARY" ]; then
  echo "Error: binary not found at $BINARY" >&2
  exit 1
fi

if [ ! -f "$CONFIG" ]; then
  echo "Error: config not found at $CONFIG" >&2
  exit 1
fi

if [ ! -d "$APP_DIR" ]; then
  echo "Warning: web UI not found at $APP_DIR — running without it"
  exec "$BINARY" stream --config "$CONFIG" --host 0.0.0.0
fi

echo "Starting mindlights stream server..."
echo "  Config:  $CONFIG"
echo "  Web UI:  http://$(hostname -I | awk '{print $1}'):30080/app"
exec "$BINARY" stream --config "$CONFIG" --host 0.0.0.0 --spa-path "$APP_DIR"
