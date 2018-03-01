FROM golang:1.10-alpine

RUN mkdir -p /usr/share/web

WORKDIR /usr/share/web

COPY config/kube.toml ./config/kube.toml

COPY statics ./statics

ADD build/am-web /usr/share/web/am-web

EXPOSE 8080

ENTRYPOINT /usr/share/web/am-web -config=/usr/share/web/config/kube.toml
