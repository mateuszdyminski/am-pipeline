package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/Shopify/sarama"
	log "github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"
	"github.com/mateuszdyminski/am-pipeline/models"
	"github.com/wvanbergen/kafka/consumergroup"
)

var configPath string

// Config holds configuration of feeder.
type Config struct {
	Zookeepers   []string
	Topic     string
	CassNodes []string
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

	storeUsers(&conf, streamUsers(&conf))

}

// BatchSize size of batch used in cassandra ingestion.
const BatchSize = 100

func storeUsers(conf *Config, users chan models.User) {
	// connect to the cluster
	cluster := gocql.NewCluster(conf.CassNodes...)
	cluster.Keyspace = "am"
	cluster.Consistency = gocql.Quorum
	cluster.ProtoVersion = 3
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("Can't create cassandra session. Err: %v", err)
	}
	defer session.Close()

	var enqued int
	batch := session.NewBatch(gocql.UnloggedBatch)
	stmt := `INSERT INTO users(id, email, dob, weight, height, nickname, country, city, caption, longitude, latitude, gender) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	for user := range users {
		if enqued > 0 && enqued%BatchSize == 0 {
			err := session.ExecuteBatch(batch)
			if err != nil {
				log.Fatalf("Can't execute batch. Err: %v", err)
			}

			log.Printf("Batch with %v users saved! Total users inserted: %v", BatchSize, enqued)

			batch = gocql.NewBatch(gocql.UnloggedBatch)
		}

		// insert a user
		batch.Query(stmt,
			user.Pnum,
			user.Email,
			user.Dob,
			user.Weight,
			user.Height,
			user.Nickname,
			user.Country,
			user.City,
			user.Caption,
			user.Location.Longitude,
			user.Location.Latitude,
			user.Gender)

		enqued++
	}

	if batch.Size() > 0 {
		if err := session.ExecuteBatch(batch); err != nil {
			log.Fatalf("Can't execute batch. Err: %v", err)
		}
	}

}

func streamUsers(conf *Config) chan models.User {
	config := consumergroup.NewConfig()
	config.Offsets.Initial = sarama.OffsetOldest
	config.Offsets.CommitInterval = 100 * time.Millisecond

	consumer, err := consumergroup.JoinConsumerGroup(
		"inserter",
		[]string{conf.Topic},
		conf.Zookeepers,
		config)
	if err != nil {
		log.Fatalf("Can't create consumer. Err: %v", err)
	}

	var received, errors int

	// Trap SIGINT to trigger a graceful shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	out := make(chan models.User, 1024)
	go func() {
		for {
			select {
			case msg := <-consumer.Messages():
				received++

				var user models.User
				if err := json.Unmarshal(msg.Value, &user); err != nil {
					log.Fatalf("Can't unmarshal data from queue! Err: %v", err)
				}

				if *user.Dob == "0000-00-00" {
					user.Dob = nil
				}

				out <- user
				consumer.CommitUpto(msg)
			case err := <-consumer.Errors():
				errors++
				log.Printf("Error reading from topic! Err: %v", err)
			case <-signals:
				log.Printf("Start consumer closing")
				consumer.Close()
				log.Printf("Consumer closed!")
				close(out)
				log.Printf("Successfully consumed: %d; errors: %d", received, errors)
				return
			}
		}
	}()

	return out
}
