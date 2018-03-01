FROM golang:1.10-alpine

RUN mkdir -p /usr/share/feeder

WORKDIR /usr/share/feeder

COPY config/kube.toml ./config/kube.toml

COPY data ./data

ADD build/feeder /usr/share/feeder/feeder

ENTRYPOINT /usr/share/feeder/feeder -config=/usr/share/feeder/config/kube.toml
