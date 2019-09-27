CADDY_VERSION := 1.0.3

NAME ?= am-$(shell echo $${PWD\#\#*/})
VERSION ?= $(shell git describe --always)

GIT_REPO := github.com/mateuszdyminski/am-pipeline/indexer
DOCKER_REPO := mateuszdyminski

.DEFAULT_GOAL := all
.PHONY: all lint test build docker-build docker-push release

all: lint test build

lint:
	echo 'TBD'

test:
	echo 'TBD'

build:
	echo 'TBD'

docker-build:
	docker build \
	--build-arg CADDY_VERSION="$(CADDY_VERSION)" \
	--label="build.version=$(VERSION)" \
	--tag="$(DOCKER_REPO)/$(NAME):latest" \
	--tag="$(DOCKER_REPO)/$(NAME):$(VERSION)" \
	.

docker-push:
	docker push "$(DOCKER_REPO)/$(NAME):latest"
	docker push "$(DOCKER_REPO)/$(NAME):$(VERSION)"

release: docker-build docker-push
