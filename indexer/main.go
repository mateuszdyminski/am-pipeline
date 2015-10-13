package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/BurntSushi/toml"
	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	"github.com/mateuszdyminski/am-pipeline/models"
	"gopkg.in/olivere/elastic.v2"
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
	client, err := elastic.NewClient(elastic.SetURL(conf.Elastics...))
	if err != nil {
		log.Fatalf("Can't create elastic client. Err: %v", err)
	}

	exists, err := client.IndexExists("users").Do()
	if err != nil {
		log.Fatalf("Can't check if index exists. Err: %v", err)
	}

	if !exists {
		// Create an index if not exists
		_, err = client.
			CreateIndex("users").
			BodyString(models.ElasticMappingString).
			Do()
		if err != nil {
			log.Fatalf("Can't create index. Err: %v", err)
		}
	}

	var enqued int
	bulkRequest := client.Bulk()
	for user := range users {
		if enqued > 0 && enqued%BulkSize == 0 {
			if _, err := bulkRequest.Do(); err != nil {
				log.Fatalf("Can't execute bulk. Err: %v", err)
			}

			log.Printf("Bulk with %v users indexed!", BulkSize)

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
		if _, err := bulkRequest.Do(); err != nil {
			log.Fatalf("Can't execute bulk. Err: %v", err)
		}
	}
}

func streamUsers(conf *Config) chan models.User {
	consumer, err := sarama.NewConsumer(conf.Brokers, sarama.NewConfig())
	if err != nil {
		log.Fatalf("Can't create consumer! Err: %v", err)
	}

	// Trap SIGINT to trigger a graceful shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	var received, errors int

	partitionConsumer, err := consumer.ConsumePartition(conf.Topic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Fatalf("Can't create partition consumer! Err: %v", err)
	}

	out := make(chan models.User, 1024)

	go func() {
		for {
			select {
			case msg := <-partitionConsumer.Messages():
				received++

				var user models.User
				if err := json.Unmarshal(msg.Value, &user); err != nil {
					log.Fatalf("Can't unmarshal data from queue! Err: %v", err)
				}

				if *user.Dob == "0000-00-00" {
					user.Dob = nil
				}

				out <- user
			case err := <-partitionConsumer.Errors():
				errors++
				log.Printf("Error reading from topic! Err: %v", err)
			case <-signals:
				partitionConsumer.AsyncClose()
				close(out)
				log.Printf("Successfully consumed: %d; errors: %d", received, errors)
				return
			}
		}
	}()

	return out
}
