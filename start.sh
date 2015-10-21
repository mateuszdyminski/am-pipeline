#!/bin/bash

deps() {
	echo "Installing dependencies..."

	go get github.com/Shopify/sarama	
	go get github.com/BurntSushi/toml
	go get github.com/Sirupsen/logrus
	go get github.com/gorilla/mux
	go get github.com/mateuszdyminski/am-pipeline/models
	go get github.com/gocql/gocql
	go get gopkg.in/olivere/elastic.v2
	go get github.com/wvanbergen/kafka/consumergroup
	go get github.com/dancannon/gorethink

	cd analyzer/statics && bower install

	echo "Installing done!"
}

startFeeder() {
	go run feeder/main.go --config=feeder/config/conf.toml
}

startIndexer() {
	go run indexer/main.go --config=indexer/config/conf.toml
}

startReceiver() {
	go run receiver/main.go --config=receiver/config/conf.toml
}

startAnalyzer() {
	go run analyzer/main.go --config=analyzer/config/conf.toml
}

CMD="$1"
SUBCMD="$2"
shift
case "$CMD" in
	deps)
		deps
	;;
	feeder)
		startFeeder
	;;
	indexer)
		startIndexer
	;;
	receiver)
		startReceiver
	;;
	analyzer)
		startAnalyzer
	;;
	infra)
		case "$SUBCMD" in
			start) docker-compose -f infra/all/docker-compose.yml up -d ;;
			stop) docker-compose -f infra/all/docker-compose.yml stop ;;
			kill) docker-compose -f infra/all/docker-compose.yml kill ;;
			rm) docker-compose -f infra/all/docker-compose.yml rm ;;
			*) echo 'Choose one of following args: {start, stop, kill, rm}'
		esac
	;;
	storm)
		case "$SUBCMD" in
			start) storm jar storm/target/pipeline-0.1.jar com.test.Pipeline am-pipeline -c nimbus.host=127.0.0.1 -c nimbus.thrift.port=49627 ;;
			kill) storm kill am-pipeline -c nimbus.host=127.0.0.1 -c nimbus.thrift.port=49627 ;;
			build) cd storm && mvn package ;;
			*) echo 'Choose one of following args: {start, kill, build}'
		esac
	;;
	*)
		echo 'Choose one of following args: {deps, feeder, indexer, receiver, analyzer, infra, storm}'
	;;
esac
