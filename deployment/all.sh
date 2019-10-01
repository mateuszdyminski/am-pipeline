#!/usr/bin/env bash

usage() {
	cat <<EOF
Usage: $(basename $0) <command>

Wrappers around pipeline workload:
    deploy                 Deploys pipeline workload on K8s cluster.
    purge                  Deletes workload from K8s cluster.
EOF
	exit 1
}

deploy() {
    BASE_DN=$1

    echo "Installing Pipeline, BaseDN: $BASE_DN"

    echo 'Installing Apache Kafka'

    if kubectl get Kafka -n kafka; then
        echo 'Apache Kafka already Installled'
    else
        helm repo add strimzi http://strimzi.io/charts/
        helm install strimzi/strimzi-kafka-operator --name kafka --namespace kafka --version 0.13
        kubectl apply -f kafka/kafka.yaml
    fi

    echo 'Installing Elasticsearch'

    if kubectl get ElasticSearch -n elastic; then
        echo 'Elasticsearch already Installled'
    else
        kubectl apply -f es/all-in-one.yaml
        kubectl apply -f es/elastic.yaml

        echo 'Installing Elastic Ingress'

        sed "s/BASE_DN/$BASE_DN/g" es/ingress.yaml > es/ingress_final.yaml
        kubectl apply -f es/ingress_final.yaml
        rm es/ingress_final.yaml
    fi
    
    echo 'Wait until Elastic is up and running'
    READY=""
    while [ -z $READY ]; do
        curl -s -k --noproxy '*' --user "elastic:password" https://elastic.$BASE_DN/_cat/health > statusout
        if grep -q 'elastic-cluster green' "statusout"; then
            READY='true'
            rm statusout
        fi
        
        [ -z "$READY" ] && sleep 5
        echo "waiting another 5s for Elastic..."
    done

    echo 'Installing Pipeline'

    kubectl apply -f am/namespace.yaml
    kubectl apply -f am/feeder-api.yaml
    kubectl apply -f am/indexer.yaml
    kubectl apply -f am/web-api.yaml
    
    sed "s/BASE_DN/$BASE_DN/g" am/web-ui.yaml > am/web-ui-final.yaml
    kubectl apply -f am/web-ui-final.yaml
    rm am/web-ui-final.yaml

    echo 'Installing Pipeline Ingresses'

    sed "s/BASE_DN/$BASE_DN/g" am/ingresses.yaml > am/ingresses_final.yaml
    kubectl apply -f am/ingresses_final.yaml
    rm am/ingresses_final.yaml

    echo 'All Done!'
}

purge() {
    echo 'Deleting Pipeline'
    kubectl delete -f am

    echo 'Deleting Kafka'
    helm del --purge kafka
    kubectl delete namespace kafka

    echo 'Deleting ElasticSearch'
    kubectl delete -f es
}

CMD="$1"
ARGS="${@:2}"

shift
case "$CMD" in
    deploy)
        deploy $ARGS
    ;;
    purge)
        purge
    ;;
	*)
		usage
	;;
esac
 