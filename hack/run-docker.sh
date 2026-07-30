#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
IMAGE="mindlights-stream:latest"
CONFIG="${1:-$SCRIPT_DIR/headset-02-attention.yaml}"

if ! docker image inspect "$IMAGE" > /dev/null 2>&1; then
  echo "Loading image..."
  docker load -i "$SCRIPT_DIR/mindlights-stream.tar"
fi

echo "Starting mindlights..."
echo "  Web UI: http://$(hostname -I | awk '{print $1}'):30080/app"

exec docker run --rm \
  --net=host \
  --privileged \
  -v "$CONFIG":/mindlights/config.yaml:ro \
  "$IMAGE"
