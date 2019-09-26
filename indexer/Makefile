GOLANG_VERSION := 1.13
ALPINE_VERSION := 3.10

NAME ?= $(shell echo $${PWD\#\#*/})
VERSION ?= $(shell git describe --always)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d %H:%M:%S')
LAST_COMMIT_USER ?= $(shell git log -1 --format='%cn <%ce>')
LAST_COMMIT_HASH ?= $(shell git log -1 --format=%H)
LAST_COMMIT_TIME ?= $(shell git log -1 --format=%cd --date=format:'%Y-%m-%d %H:%M:%S')

GIT_REPO := github.com/mateuszdyminski/am-pipeline/indexer
DOCKER_REPO := mateuszdyminski

.DEFAULT_GOAL := all
.PHONY: all lint test build docker-build docker-push release

all: lint test build

lint:
	golangci-lint run -v

test:
	go test -v ./...

build:
	CGO_ENABLED=0 \
	go build \
	-v \
	-ldflags "-s -w -X '$(GIT_REPO)/pkg/version.AppName=$(NAME)' -X '$(GIT_REPO)/pkg/version.AppVersion=$(VERSION)' -X '$(GIT_REPO)/pkg/version.BuildTime=$(BUILD_TIME)' -X '$(GIT_REPO)/pkg/version.LastCommitUser=$(LAST_COMMIT_USER)' -X '$(GIT_REPO)/pkg/version.LastCommitHash=$(LAST_COMMIT_HASH)' -X '$(GIT_REPO)/pkg/version.LastCommitTime=$(LAST_COMMIT_TIME)'" \
	-o $(NAME)-$(VERSION) .

docker-build:
	docker build \
	--build-arg GOLANG_VERSION="$(GOLANG_VERSION)" \
	--build-arg ALPINE_VERSION="$(ALPINE_VERSION)" \
	--build-arg NAME="$(NAME)" \
	--build-arg VERSION="$(VERSION)" \
	--build-arg BUILD_TIME="$(BUILD_TIME)" \
	--build-arg LAST_COMMIT_USER="$(LAST_COMMIT_USER)" \
	--build-arg LAST_COMMIT_HASH="$(LAST_COMMIT_HASH)" \
	--build-arg LAST_COMMIT_TIME="$(LAST_COMMIT_TIME)" \
	--label="build.version=$(VERSION)" \
	--tag="$(DOCKER_REPO)/$(NAME):latest" \
	--tag="$(DOCKER_REPO)/$(NAME):$(VERSION)" \
	.

docker-push:
	docker push "$(DOCKER_REPO)/$(NAME):latest"
	docker push "$(DOCKER_REPO)/$(NAME):$(VERSION)"

release: docker-build docker-push
