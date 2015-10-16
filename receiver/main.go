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
	"github.com/mateuszdyminski/am-pipeline/models"
	"github.com/wvanbergen/kafka/consumergroup"
	r "github.com/dancannon/gorethink"
)

var configPath string

// Config holds configuration of feeder.
type Config struct {
	Zookeepers   []string
	Topic     string
	Rethinkdb string
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
	session, err := r.Connect(r.ConnectOpts{ Address: conf.Rethinkdb })
	if err != nil {
	    log.Fatalln(err.Error())
	}
	setupDb(session)

	var enqued int
	usersBatch := make([]models.User, 0)
	for user := range users {
		if enqued > 0 && enqued%BatchSize == 0 {
			resp, err := r.DB("am").Table("users").Insert(usersBatch).RunWrite(session, r.RunOpts{
				MinBatchRows: BatchSize,
				MaxBatchRows: BatchSize,
			})
			if err != nil {
				log.Fatalf("Can't insert users to rethinkdb. Err: %v", err)
			}

			usersBatch = make([]models.User, 0)
			log.Infof("%d users inserted. Total inserted users: %d", resp.Inserted, enqued)
		}

		usersBatch = append(usersBatch, user)
		enqued++
	}

	 if len(usersBatch) > 0 {
		 _, err := r.DB("am").Table("users").Insert(usersBatch).RunWrite(session, r.RunOpts{
			 MinBatchRows: BatchSize,
			 MaxBatchRows: BatchSize,
		 })
		 if err != nil {
			 log.Fatalf("Can't insert users to rethinkdb. Err: %v", err)
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

func setupDb(session *r.Session) {
	// Setup database + tables
	if err := r.DBCreate("am").Exec(session); err != nil {
		log.Warningf("Can't create DB in rethinkdb. Propably DB is aready created. Err: %v", err)
	} else {
		log.Infof("DB created")
	}

	if err := r.DB("am").TableCreate("users").Exec(session); err != nil {
		log.Warningf("Can't create table in rethinkdb. Propably table is aready created. Err: %v", err)
	} else {
		log.Infof("DB table created!")
	}
}