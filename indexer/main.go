package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"

	"gopkg.in/olivere/elastic.v5"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/mateuszdyminski/am-pipeline/models"
)

var configPath string

// Config holds configuration of feeder.
type Config struct {
	Brokers  []string
	Topic    string
	Elastics []string
}

func init() {
	flag.Usage = func() {
		flag.PrintDefaults()
	}

	flag.StringVar(&configPath, "config", "config/conf.toml", "config path")
}

func main() {
	// load config
	flag.Parse()

	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Can't open config file!")
	}

	var conf Config
	if err := toml.Unmarshal(bytes, &conf); err != nil {
		log.Fatalf("Can't decode config file!")
	}

	indexUsers(&conf, streamUsers(&conf))
}

// BulkSize size of the bulk.
const BulkSize = 100

func indexUsers(conf *Config, users chan models.User) {
	// connect to the cluster
	client, err := elastic.NewClient(elastic.SetURL(conf.Elastics...), elastic.SetSniff(false))
	if err != nil {
		log.Fatalf("Can't create elastic client. Err: %v", err)
	}

	exists, err := client.IndexExists("users").Do(context.Background())
	if err != nil {
		log.Fatalf("Can't check if index exists. Err: %v", err)
	}

	if !exists {
		log.Info("Creating index 'users'")
		// Create an index if not exists
		_, err = client.
			CreateIndex("users").
			BodyString(models.ElasticMappingString).
			Do(context.Background())
		if err != nil {
			log.Fatalf("Can't create index. Err: %v", err)
		}
	}

	var enqued int
	bulkRequest := client.Bulk()
	for user := range users {
		if enqued > 0 && enqued%BulkSize == 0 {
			if _, err := bulkRequest.Do(context.Background()); err != nil {
				log.Fatalf("Can't execute bulk. Err: %v", err)
			}

			log.Infof("Bulk with %v users indexed! Total indexed users: %v", BulkSize, enqued)

			bulkRequest = client.Bulk()
		}

		bulkRequest.Add(
			elastic.NewBulkIndexRequest().
				Index("users").
				Type("user").
				Id(fmt.Sprintf("%d", user.Pnum)).
				Doc(user))

		enqued++
	}

	if bulkRequest.NumberOfActions() > 0 {
		if _, err := bulkRequest.Do(context.Background()); err != nil {
			log.Fatalf("Can't execute bulk. Err: %v", err)
		}
	}
}

func streamUsers(conf *Config) chan models.User {
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true

	// init consumer
	brokers := conf.Brokers
	topics := []string{conf.Topic}
	consumer, err := cluster.NewConsumer(brokers, "my-consumer-group", topics, config)
	if err != nil {
		log.Panicf("Can't create Kafka client! Err: %+v", err.Error())
	}

	var received, errors int

	// Trap SIGINT to trigger a graceful shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	out := make(chan models.User, 1024)

	// consume errors
	go func() {
		for err := range consumer.Errors() {
			log.Errorf("Error reading from topic! Err: %+v", err.Error())
		}
	}()

	// consume notifications
	go func() {
		for ntf := range consumer.Notifications() {
			log.Infof("Rebalanced: %+v\n", ntf)
		}
	}()

	go func() {
		for {
			select {
			case msg, ok := <-consumer.Messages():
				if ok {
					received++

					var user models.User
					if err := json.Unmarshal(msg.Value, &user); err != nil {
						log.Errorf("Can't unmarshal data from queue! Err: %v", err)
						continue
					}

					if *user.Dob == "0000-00-00" {
						user.Dob = nil
					}

					out <- user
					consumer.MarkOffset(msg, "")
				}
			case <-signals:
				log.Infof("Start consumer closing")
				consumer.Close()
				log.Infof("Consumer closed!")
				close(out)
				log.Infof("Successfully consumed: %d; errors: %d", received, errors)
				return
			}
		}
	}()

	return out
}
