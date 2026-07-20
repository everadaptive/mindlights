# MindLights v3.0

MindLights is a toolkit for using [NeuroSky EEGs](https://neurosky.com/biosensors/eeg-sensor/) to control lights via DMX or stream brainwave data over OSC.

<img alt="3 modified MindFlex headsets" src="assets/headsets.jpeg " height="200"/>
<img alt="Color gradient showing Attention levels over time" src="assets/spectrum.jpeg " height="200"/>

## Usage

```
mindlights --config <config-file>
```

### Visualizations

| Flag | Description |
|------|-------------|
| `attention-light` | Map attention level to a DMX light color |
| `meditation-light` | Map meditation level to a DMX light color |
| `moving-head` | Control a moving head fixture |
| `osc` | Stream all EEG data over OSC (UDP) |
| `csv` | Log EEG data to a CSV file |

### OSC Output

```
mindlights --config ./hack/headset-02-attention.yaml \
           --display=dummy \
           --visualization=osc \
           --osc-host=192.168.1.100 \
           --osc-port=9000
```

OSC addresses (all values are `int32`):

| Address | Description |
|---------|-------------|
| `/mindlights/<name>/signal` | Signal quality (0 = perfect) |
| `/mindlights/<name>/attention` | Attention level (0–100) |
| `/mindlights/<name>/meditation` | Meditation level (0–100) |
| `/mindlights/<name>/eeg/delta` | Delta band power |
| `/mindlights/<name>/eeg/theta` | Theta band power |
| `/mindlights/<name>/eeg/lowAlpha` | Low alpha band power |
| `/mindlights/<name>/eeg/highAlpha` | High alpha band power |
| `/mindlights/<name>/eeg/lowBeta` | Low beta band power |
| `/mindlights/<name>/eeg/highBeta` | High beta band power |
| `/mindlights/<name>/eeg/lowGamma` | Low gamma band power |
| `/mindlights/<name>/eeg/highGamma` | High gamma band power |

### Config File

See example configs in `hack/`. Key fields:

```yaml
display: dummy            # dummy | udmx | serialdmx | ftdidmx
visualization: osc        # attention-light | meditation-light | moving-head | osc | csv
osc-host: 127.0.0.1
osc-port: 9000

eeg_headsets:
  - name: headset-02
    bluetoothAddress: "98:D3:31:80:7B:67"   # MAC on Linux; /dev/tty.* path on macOS
    display:
      size: 8
      start: 1
```

## Building

### Linux (NixOS)

```bash
nix-shell --run "CGO_ENABLED=1 go build -o build/linux_amd64/mindlights ./cmd/scan"
```

### macOS

Pair the headset in System Settings → Bluetooth. It will appear as `/dev/tty.<name>-SerialPort`. Use that path as `bluetoothAddress` in your config.

```bash
CGO_ENABLED=0 go build -o mindlights ./cmd/scan
```

### Docker (cross-compile all platforms)

```bash
docker run --rm \
  -v $(pwd):/go/src/github.com/everadaptive/mindlights \
  --entrypoint=/go/src/github.com/everadaptive/mindlights/hack/scripts/build.sh \
  docker.io/dockercore/golang-cross:latest
```

Outputs: `build/linux_amd64/`, `build/linux_arm7/`, `build/macOS_amd64/`, `build/macOS_arm64/`

## Running on Linux

The app connects directly to the headset via raw Bluetooth RFCOMM sockets, which requires BlueZ to be stopped first:

```bash
sudo systemctl stop bluetooth
nix-shell --run "./build/linux_amd64/mindlights --config ./hack/headset-02-attention.yaml"
```
