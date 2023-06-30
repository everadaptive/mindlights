# MindLights v3.0

MindLights is a toolit for using [NeuroSky EEGs](https://neurosky.com/biosensors/eeg-sensor/) to control lights using the DMX protocol

## Usage

## Building
```
CGO_ENABLED=1 go build mindlights .
```

## Release Build
```
docker run --rm -v $(pwd):/go/src/github.com/everadaptive/mindlights --entrypoint=/go/src/github.com/everadaptive/mindlights/hack/scripts/build.sh docker.io/dockercore/golang-cross:latest
```