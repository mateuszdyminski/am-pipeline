package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/BurntSushi/toml"
	"github.com/Shopify/sarama"
	"github.com/mateuszdyminski/am-pipeline/models"
)

var configPath string

// Config holds configuration of feeder.
type Config struct {
	Brokers []string
	Topic   string
	CsvPath string
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

	// pump data into Kafka
	pumpData(&conf, streamUsers(&conf))
}

func streamUsers(conf *Config) chan models.User {
	f, err := os.Open(conf.CsvPath)
	if err != nil {
		log.Fatal("can't file with users:", err)
	}
	defer f.Close()

	log.Infof("Start reading CSV file!")

	r := csv.NewReader(bufio.NewReader(f))
	r.Comma = '|'
	r.LazyQuotes = true
	records, err := r.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Read %d lines!", len(records))

	out := make(chan models.User, 1024)
	go func() {
		for i, line := range records {
			if len(line) != 12 {
				log.Fatal(fmt.Errorf("Wrong number of parsed fields: %d. Index %d", len(line), i))
			}
			u := models.User{}

			var err error
			u.Pnum, err = strconv.ParseInt(line[0], 10, 64)
			if err != nil {
				log.Fatalf("Can't deserialize pnum. Val: %s", line[0])
			}

			long, err := strconv.ParseFloat(line[1], 64)
			if err != nil {
				log.Fatalf("Can't deserialize longitude. Val: %s. Line: %d", line[1], i)
			}

			lat, err := strconv.ParseFloat(line[2], 64)
			if err != nil {
				log.Fatalf("Can't deserialize latitude. Val: %s. Line: %d", line[2], i)
			}

			u.Location = &models.Location{long, lat}
			if long == 0 || lat == 0 {
				log.Warningf("At least one value of location could be wrong. Vals long, %d, lat: %d", long, lat)
			}

			u.Email = line[3]
			weight, err := strconv.Atoi(line[4])
			if err != nil {
				log.Fatalf("Can't deserialize weight. Val: %s", line[4])
			}
			u.Weight = &weight

			height, err := strconv.Atoi(line[5])
			if err != nil {
				log.Fatalf("Can't deserialize height. Val: %s", line[5])
			}
			u.Height = &height

			u.Nickname = &line[6]
			u.Country, err = strconv.Atoi(line[7])
			if err != nil {
				log.Fatalf("Can't deserialize country: Val: %s", line[7])
			}
			u.City = &line[8]
			u.Caption = &line[9]
			gender, err := strconv.Atoi(line[10])
			if err != nil {
				log.Fatalf("Can't deserialize gender. Val: %s", line[10])
			}
			u.Gender = &gender
			u.Dob = &line[11]

			out <- u
		}

		log.Infof("All users sent. Closing channel")
		close(out)
	}()

	return out
}

func pumpData(conf *Config, users chan models.User) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer, err := sarama.NewAsyncProducer(conf.Brokers, config)
	if err != nil {
		log.Fatalf("Can't create producer! Err: %v", err)
	}

	// Trap SIGINT to trigger a graceful shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	var (
		wg                          sync.WaitGroup
		enqueued, successes, errors int
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _ = range producer.Successes() {
			successes++
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for err := range producer.Errors() {
			log.Println(err)
			errors++
		}
	}()

ProducerLoop:
	for user := range users {
		b, _ := json.Marshal(user)
		message := &sarama.ProducerMessage{Topic: conf.Topic, Value: sarama.ByteEncoder(b)}
		select {
		case producer.Input() <- message:
			enqueued++

		case <-signals:
			producer.AsyncClose() // Trigger a shutdown of the producer.
			break ProducerLoop
		}
	}

	producer.AsyncClose()

	wg.Wait()

	log.Printf("Successfully produced: %d; errors: %d", successes, errors)
}
