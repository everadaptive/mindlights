FROM docker.io/dockercore/golang-cross:latest

ADD . /go/src/github.com/everadaptive/mindlights

WORKDIR /go/src/github.com/everadaptive/mindlights

ENTRYPOINT [ "./hack/scripts/build.sh" ]