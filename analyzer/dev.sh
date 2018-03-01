#!/bin/bash

build() {
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/am-web -a -tags netgo .
}

buildDocker() {
	docker build -t mateuszdyminski/am-web:latest .
}

pushDocker() {
	docker push mateuszdyminski/am-web
}

CMD="$1"
SUBCMD="$2"
shift
case "$CMD" in
	go)
		build
	;;
	docker)
		buildDocker
	;;
    push)
		pushDocker
	;;
	*)
		echo 'Choose one of following args: {go, docker, push}'
	;;
esac
