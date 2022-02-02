# MindLights v3.0

## Usage

## Building
```
CGO_ENABLED=1 go build mindlights .
```

## Release Build
```
docker run --rm -v $(pwd):/go/src/github.com/everadaptive/mindlights --entrypoint=/go/src/github.com/everadaptive/mindlights/hack/scripts/build.sh docker.io/dockercore/golang-cross:latest
```