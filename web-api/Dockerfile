# Arguments
ARG GOLANG_VERSION
ARG ALPINE_VERSION

# Build part
FROM golang:${GOLANG_VERSION}-alpine${ALPINE_VERSION} AS builder

RUN apk --no-cache add make git; \
    adduser -D -h /tmp/build build
USER build
WORKDIR /tmp/build

COPY --chown=build Makefile Makefile
COPY --chown=build go.mod go.mod
COPY --chown=build go.sum go.sum
RUN go mod download

ARG VERSION
ARG NAME
ARG BUILD_TIME
ARG LAST_COMMIT_USER
ARG LAST_COMMIT_HASH
ARG LAST_COMMIT_TIME

COPY --chown=build pkg pkg
COPY --chown=build main.go main.go
RUN make build

# Exec part
FROM alpine:${ALPINE_VERSION}

ARG VERSION
ARG NAME

RUN mkdir -p /web-api
RUN chown guest /web-api
USER guest

# Copy from repo
RUN mkdir -p /web-api/config
COPY config/kube.toml /web-api/config/

# Copy from builder
COPY --from=builder /tmp/build/${NAME}-${VERSION} /usr/bin/${NAME}

# Exec
CMD ["am-web-api", "--config=/web-api/config/kube.toml"]
