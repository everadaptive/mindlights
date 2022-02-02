#!/bin/bash
cd /go/src/github.com/everadaptive/mindlights

rm -rf build

## linux
apt update
apt install -y libusb-1.0-0-dev

mkdir -p build/linux_amd64/
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o build/linux_amd64/mindlights .

mkdir -p build/linux_arm7/
dpkg --add-architecture armhf
apt update
apt install -y crossbuild-essential-armhf libusb-1.0-0-dev:armhf
GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc  go build -o build/linux_arm7/mindlights .

## macOS
curl -L -H "Authorization: Bearer QQ==" -o /tmp/libusb.tar.gz https://ghcr.io/v2/homebrew/core/libusb/blobs/sha256:e202da5a53b0955b4310805b09e9f4af3b73eed57de5ae0d44063e84dca5eafd 
tar -zxf /tmp/libusb.tar.gz -C /tmp

mkdir -p build/macOS/
OSXCROSS_NO_INCLUDE_PATH_WARNINGS=1 MACOSX_DEPLOYMENT_TARGET=10.10 CGO_LDFLAGS="-L/tmp/libusb/1.0.25/lib" CC=o64-clang CXX=o64-clang++ GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o build/macOS/mindlights .
