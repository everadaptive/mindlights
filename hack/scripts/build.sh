#!/bin/bash
cd /go/src/github.com/everadaptive/mindlights

## linux
apt update
apt install -y libusb-1.0-0-dev

mkdir -p build/linux/
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o build/linux/mindlights .

## macOS
curl -L -H "Authorization: Bearer QQ==" -o /tmp/libusb.tar.gz https://ghcr.io/v2/homebrew/core/libusb/blobs/sha256:e202da5a53b0955b4310805b09e9f4af3b73eed57de5ae0d44063e84dca5eafd 
tar -zxf /tmp/libusb.tar.gz -C /tmp

mkdir -p build/macOS/
OSXCROSS_NO_INCLUDE_PATH_WARNINGS=1 MACOSX_DEPLOYMENT_TARGET=10.10 CGO_LDFLAGS="-L/tmp/libusb/1.0.25/lib" CC=o64-clang CXX=o64-clang++ GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o build/macOS/mindlights .
