FROM golang:1.10-alpine

RUN mkdir -p /usr/share/indexer

WORKDIR /usr/share/indexer

COPY config/kube.toml ./config/kube.toml

ADD build/indexer /usr/share/indexer

ENTRYPOINT /usr/share/indexer/indexer -config=/usr/share/indexer/config/kube.toml
