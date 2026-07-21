#!/bin/bash
cd /go/src/github.com/everadaptive/mindlights

rm -rf build

## linux amd64
apt update
apt install -y libusb-1.0-0-dev libftdi1-dev

mkdir -p build/linux_amd64/
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o build/linux_amd64/mindlights ./cmd/scan

## linux arm7
mkdir -p build/linux_arm7/
dpkg --add-architecture armhf
apt update
apt install -y crossbuild-essential-armhf libusb-1.0-0-dev:armhf libftdi1-dev:armhf
GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc go build -o build/linux_arm7/mindlights ./cmd/scan

## macOS — USB/FTDI display backends are excluded on darwin via build tags,
## so this Linux container can cross-compile without an osxcross toolchain.
##
## WARNING: these CGO_ENABLED=0 binaries can crash at startup on recent
## macOS with a dyld `clock_gettime` error, because Go's darwin runtime
## expects to reach syscalls through libSystem. They are fine as a quick
## smoke-test, but ship the CGO_ENABLED=1 binaries built on the native
## macOS GitHub Actions runners (see .github/workflows/build.yml) instead.
mkdir -p build/macOS_amd64/
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o build/macOS_amd64/mindlights ./cmd/scan

mkdir -p build/macOS_arm64/
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o build/macOS_arm64/mindlights ./cmd/scan
