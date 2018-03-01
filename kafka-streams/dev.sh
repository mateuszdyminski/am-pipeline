#!/bin/bash

build() {
	mvn clean install
}

buildDocker() {
	docker build -t mateuszdyminski/streams:latest .
}

pushDocker() {
	docker push mateuszdyminski/streams
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
