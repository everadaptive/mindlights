#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ARCH=$(uname -m)
case "$ARCH" in
  aarch64) TAG="arm64" ;;
  x86_64)  TAG="amd64" ;;
  *)        TAG="$ARCH" ;;
esac
IMAGE="mindlights-stream:$TAG"
CONFIG="${1:-$SCRIPT_DIR/headset-02-attention.yaml}"

if ! docker image inspect "$IMAGE" > /dev/null 2>&1; then
  echo "Image $IMAGE not found. Building..."
  docker buildx build --platform linux/$TAG -t "$IMAGE" -f "$SCRIPT_DIR/../Dockerfile.stream" --load "$SCRIPT_DIR/.."
fi

echo "Starting mindlights ($IMAGE)..."
echo "  Web UI: http://$(hostname -I | awk '{print $1}'):30080/app"

exec docker run --rm \
  --net=host \
  --privileged \
  -v "$CONFIG":/mindlights/config.yaml:ro \
  "$IMAGE"
