# AM pipeline
Simple data pipeline showing how to create full text search over some dataset.

## Description
We are using Kafka and Elasticsearch.

There are 4 components of AM pipeline:
* Feeder - reads CSV file with users as input and puts them into Kafka
* Indexer - reads users from Kafka and put them in Elasticsearch
* Kafka-Streams - reads users from Kafka and calculates statistics on top of them
* Analyzer - UI which provides users visualization
    * locations of users on the map
    * search for any particular field or search over all fields
    * pagination of results

## Kafka Presentation

Please install [[https://godoc.org/golang.org/x/tools/present][https://godoc.org/golang.org/x/tools/present]]

Then run:
```
cd presentation && present
```

## Requirements

* minikube
* installed kubectl
* go >= 1.8
* yarn - to install JS requirements
* maven (for kafka streams)
* java jdk >= 8

## Infrastructure

### To start all services

Read README.md file in `infra/kube` directory

