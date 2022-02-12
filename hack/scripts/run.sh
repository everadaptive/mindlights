#!/bin/bash

function cleanup() {
rfcomm release /dev/rfcomm0
}

rfcomm connect /dev/rfcomm0 98:D3:31:70:71:0A 1 &
sleep 5
./mindlights --config ./hack/dmx-8.yaml

trap cleanup EXIT
