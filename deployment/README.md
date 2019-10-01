# Kubernetes Installation of All component in am-pipeline

Document describes how to deploy whole AM pipeline on kubernetes.

## Prerequisits

* `helm`
* `kubectl`
* `hey` [https://github.com/rakyll/hey](https://github.com/rakyll/hey) - for perf tests

## External endpoints - after deployment

* WEB UI - [https://web.YOUR_DOMAIN_HERE](https://web.YOUR_DOMAIN_HERE)
* WEB API - [https://api.YOUR_DOMAIN_HERE](https://api.YOUR_DOMAIN_HERE)
* FEEDER API - [https://feeder.YOUR_DOMAIN_HERE](https://feeder.YOUR_DOMAIN_HERE)
* INDEXER API - [https://indexer.YOUR_DOMAIN_HERE](https://indexer.YOUR_DOMAIN_HERE)

## Deploy everything in single step

Install all services (Kafka, Elasticsearch, Pipeline):

```bash
./all.sh deploy <your.domain.here.com>
```

## Kafka

Add strimzi repo:

```bash
helm repo add strimzi http://strimzi.io/charts/
```

Install Kafka Operator:

```bash
helm install strimzi/strimzi-kafka-operator --name kafka --namespace kafka --version 0.13
```

Install Kafka with predefied topic `users`:

```bash
kubectl apply -f kafka/kafka.yaml
```

## Elasticsearch

Install Elastic Operator:

```bash
kubectl apply -f es/all-in-one.yaml
```

Install ElasticSearch:

```bash
kubectl apply -f es/elastic.yaml
```

## Pipeline

Steps how to install am-pipeline:

Create namespace:

```bash
kubectl apply -f am/namespace.yaml
```

Install feeder-api:

```bash
kubectl apply -f am/feeder-api.yaml
```

Install indexer:

```bash
kubectl apply -f am/indexer.yaml
```

Install web API:

```bash
kubectl apply -f am/web-api.yaml
```

Install web UI:

```bash
BASE_DN=your.domain.here.com
sed "s/BASE_DN/$BASE_DN/g" am/web-ui.yaml > am/web-ui-final.yaml
kubectl apply -f am/web-ui-final.yaml
rm am/web-ui-final.yaml
```

Install feeder(CronJob):

```bash
kubectl apply -f am/feeder.yaml
```

### Exposing with Ingresses

Install ingresses(we assume that ingress controller is already configured):

```bash
BASE_DN=your.domain.here.com
sed "s/BASE_DN/$BASE_DN/g" am/ingresses.yaml > am/ingresses_final.yaml
kubectl apply -f am/ingresses_final.yaml
rm am/ingresses_final.yaml

sed "s/BASE_DN/$BASE_DN/g" es/ingress.yaml > es/ingress_final.yaml
kubectl apply -f es/ingress_final.yaml
rm es/ingress_final.yaml
```

## Destroy everything

```bash
./all.sh purge
```
