################################################
# building the app

FROM golang AS build-env

WORKDIR /go/src/github.com/phoenix-mstu/go-modifying-mqtt-proxy
ADD cmd cmd
ADD internal internal

RUN go get -d ./...
RUN CGO_ENABLED=0 \
    go install github.com/phoenix-mstu/go-modifying-mqtt-proxy/cmd/proxy

################################################
# making main image

FROM scratch
COPY --from=build-env /go/bin/proxy /
ADD sample_configs /sample_configs

EXPOSE 1883
ENTRYPOINT ["/proxy"]