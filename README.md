# AM pipeline
Simple data pipeline showing how to create full text search over some dataset. 

### Description of master branch
We are using Kafka, Cassandra and Elasticsearch.

There are 4 components of AM pipeline:
* Feeder - reads CSV file with users as input and puts them into Kafka
* Receiver - reads users from Kafka and put them in Cassandra 3
* Indexer - reads users from Kafka and put them in Elasticsearch
* Analyzer - UI which provides users visualization
    * locations of users on the map
    * search for any particular field or search over all fields
    * pagination of results

### Description of rethink branch
We are using Kafka, RethinkDB and Elasticsearch.

There are 4 components of AM pipeline:
* Feeder - same like in master branch
* Receiver - reads users from Kafka insert them in RethinkDB
* Indexer - listens on changes from RethinkDB and inserts data into Elasticsearch
* Analyzer - same like in master branch

### Description of storm branch
We are using Kafka, Storm, Cassandra and Elasticsearch.

There are 4 components of AM pipeline:
* Feeder - same like in master branch
* Storm - reads users from Kafka then persists and indexes them within Storm topology
* Analyzer - same like in master branch

## Requirements 

* docker 
* docker-compose
* go >= 1.3
* cqlsh tool compatible with Cassandra 3(for master branch)
* bower - to install JS requirements
* maven (for branch storm)
* java jdk >= 6 (for branch storm)
* storm (for branch storm)

### Requirements
To install all go and bower requirements:
```
./start.sh deps
```

## Infrastructure

### To start all services

```
./start.sh infra start
```

### To stop all services
```
./start.sh infra stop
```

### To remove state of all services
```
./start.sh infra rm
```

## Pipeline

### To run Feeder:
```
./start.sh feeder
```
It will start process which reads and parses data from CSV file and pumps it into Kafka.


### To run Receiver<Optional>:
```
./start.sh receiver
```
It will start process which reads data from Kafka inserts it into Cassandra. It's optional step - analyzer is using Elasticsearch as backend.

### To run Indexer:
```
./start.sh indexer
```
It will start process which reads data from Kafka and pumps it into Elasticsearch.

### To run Analyzer:
```
./start.sh analyzer
```
It will start local webserver on port 9000. Go to http://localhost:9000/ and start searching for AM users. 

### <Internal use only> Dump local data from Postgres to CSV

```
COPY (SELECT pnum, longitude, latitude, email, weight, height, nickname, country, city, caption, gender, dob FROM aminno_member ORDER BY pnum DESC limit 100000 ) TO '/tmp/am-users.csv' WITH CSV HEADER DELIMITER '|' quote '''';```


