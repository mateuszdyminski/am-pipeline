package main

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/mateuszdyminski/am-pipeline/models"

	"github.com/BurntSushi/toml"
	"github.com/Shopify/sarama"
	_ "github.com/go-sql-driver/mysql"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	log "github.com/sirupsen/logrus"
)

var configPath string

// Config holds configuration of feeder.
type Config struct {
	Brokers            []string
	Topic              string
	DbString           string
	CsvPath            string
	SourceDataType     string
	PushgatewayAddress string
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
		log.Fatal("can't open config file", err)
	}

	var conf Config
	if err := toml.Unmarshal(bytes, &conf); err != nil {
		log.Fatal("can't decode config file", err)
	}

	feeder, err := NewFeeder(conf)
	if err := toml.Unmarshal(bytes, &conf); err != nil {
		log.Fatal("can't create feeder!", err)
	}

	feeder.Start()
}

type Feeder struct {
	read           *prometheus.CounterVec
	readErr        *prometheus.CounterVec
	sent           *prometheus.CounterVec
	sentErr        *prometheus.CounterVec
	completionTime prometheus.Gauge
	duration       prometheus.Gauge
	producer       sarama.SyncProducer
	cfg            Config
}

func NewFeeder(cfg Config) (*Feeder, error) {
	completionTime := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "am",
		Subsystem: "feeder",
		Name:      "feed_last_timestamp_seconds",
		Help:      "The timestamp of the last successful feed read.",
	})

	duration := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "am",
		Subsystem: "feeder",
		Name:      "feed_duration_seconds",
		Help:      "The duration of the last users feed.",
	})

	read := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "am",
			Subsystem: "feeder",
			Name:      "read_total",
			Help:      "The total number of read users before send them to Kafka.",
		},
		[]string{"source"},
	)

	readErr := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "am",
			Subsystem: "feeder",
			Name:      "read_total_err",
			Help:      "The total number of read errors.",
		},
		[]string{"source"},
	)

	sent := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "am",
			Subsystem: "feeder",
			Name:      "sent_total",
			Help:      "The total number of sent users to Kafka.",
		},
		[]string{"topic"},
	)

	sentErr := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "am",
			Subsystem: "feeder",
			Name:      "sent_total_err",
			Help:      "The total number of sent errors.",
		},
		[]string{"topic"},
	)

	config := sarama.NewConfig()
	config.Version = sarama.V2_3_0_0
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, err
	}

	feeder := &Feeder{
		sent:           sent,
		sentErr:        sentErr,
		read:           read,
		readErr:        readErr,
		completionTime: completionTime,
		duration:       duration,
		producer:       producer,
		cfg:            cfg,
	}

	return feeder, nil
}

func (f *Feeder) Start() {
	// when the read phase is over we need to send the metrics to pushgateway
	pusher := push.New(f.cfg.PushgatewayAddress, "users_feed").
		Collector(f.read).
		Collector(f.readErr).
		Collector(f.sent).
		Collector(f.sentErr).
		Collector(f.duration).
		Collector(f.completionTime)

	start := time.Now()

	// pump data into Kafka
	f.pumpData(f.streamUsers())

	f.duration.Set(time.Since(start).Seconds())
	f.completionTime.SetToCurrentTime()

	if err := pusher.Push(); err != nil {
		log.Error("could not push metrics to Pushgateway:", err)
	}

	log.Info("Metrics pushed to Pushgateway")
}

func (f *Feeder) streamUsers() chan models.User {
	if f.cfg.SourceDataType == "db" {
		return f.streamDbUsers()
	}

	if f.cfg.SourceDataType == "csv" {
		return f.streamCsvUsers()
	}

	log.Fatalf("Can't find proper data source type")
	return nil
}

func (f *Feeder) streamDbUsers() chan models.User {
	db, err := sql.Open("mysql", f.cfg.DbString)
	if err != nil {
		log.Fatal("can't open database conn:", err)
	}

	log.Infof("Start reading users from DB!")

	out := make(chan models.User, 1024)
	go func() {
		rows, err := db.Query("select pnum, dob, weight, height, nickname, country, city, caption, longitude, latitude, gender from aminno_member LIMIT 100000")
		if err != nil {
			log.Fatal("can't run query", err)
		}
		defer rows.Close()

		noOfUsers := 0
		for rows.Next() {
			u := models.User{}
			u.Location = &models.Location{}
			if err := rows.Scan(&u.Pnum, &u.Dob, &u.Weight, &u.Height, &u.Nickname, &u.Country, &u.City, &u.Caption, &u.Location.Longitude, &u.Location.Latitude, &u.Gender); err != nil {
				log.Error("can't scan values. ", err)
				f.readErr.WithLabelValues("db").Inc()
				continue
			}

			out <- u
			f.read.WithLabelValues("db").Inc()

			if noOfUsers%100 == 0 {
				log.Printf("Total no of read users: %v", noOfUsers)
			}

			noOfUsers++
		}

		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}

		log.Infof("All users sent")
		close(out)
	}()

	return out
}

func (f *Feeder) streamCsvUsers() chan models.User {
	file, err := os.Open(f.cfg.CsvPath)
	if err != nil {
		log.Fatal("can't file with users:", err)
	}
	defer file.Close()

	log.Infof("Start reading CSV file!")

	r := csv.NewReader(bufio.NewReader(file))
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
				log.Errorf("wrong number of parsed fields: %d. Index %d", len(line), i)
				f.readErr.WithLabelValues("csv").Inc()
				continue
			}
			u := models.User{}
			log.Infof("read user record: %s", line)

			var err error
			u.Pnum, err = strconv.ParseInt(line[0], 10, 64)
			if err != nil {
				log.Errorf("can't deserialize pnum. Val: %s", line[0])
				f.readErr.WithLabelValues("csv").Inc()
				continue
			}

			long, err := strconv.ParseFloat(line[1], 64)
			if err != nil {
				log.Errorf("can't deserialize longitude. Val: %s. Line: %d", line[1], i)
				f.readErr.WithLabelValues("csv").Inc()
				continue
			}

			lat, err := strconv.ParseFloat(line[2], 64)
			if err != nil {
				log.Errorf("can't deserialize latitude. Val: %s. Line: %d", line[2], i)
				f.readErr.WithLabelValues("csv").Inc()
				continue
			}

			u.Location = &models.Location{long, lat}
			if long == 0 || lat == 0 {
				log.Warningf("at least one value of location could be wrong. Vals long, %d, lat: %d", long, lat)
			}

			u.Email = &line[3]
			weight, err := strconv.Atoi(line[4])
			if err != nil {
				log.Errorf("can't deserialize weight. Val: %s", line[4])
				f.readErr.WithLabelValues("csv").Inc()
				continue
			}
			u.Weight = &weight

			height, err := strconv.Atoi(line[5])
			if err != nil {
				log.Errorf("can't deserialize height. Val: %s", line[5])
				f.readErr.WithLabelValues("csv").Inc()
				continue
			}
			u.Height = &height

			u.Nickname = &line[6]
			u.Country, err = strconv.Atoi(line[7])
			if err != nil {
				log.Errorf("can't deserialize country: Val: %s", line[7])
				f.readErr.WithLabelValues("csv").Inc()
				continue
			}
			u.City = &line[8]
			u.Caption = &line[9]
			gender, err := strconv.Atoi(line[10])
			if err != nil {
				log.Errorf("can't deserialize gender. Val: %s", line[10])
				f.readErr.WithLabelValues("csv").Inc()
				continue
			}
			u.Gender = &gender
			u.Dob = &line[11]

			out <- u
			f.read.WithLabelValues("csv").Inc()
			continue
		}

		log.Infof("All users sent. Closing channel")
		close(out)
	}()

	return out
}

func (f *Feeder) pumpData(users chan models.User) {
	defer func() {
		if err := f.producer.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	var successes, errors int

	for user := range users {
		b, _ := json.Marshal(user)
		message := &sarama.ProducerMessage{
			Topic:     f.cfg.Topic,
			Value:     sarama.ByteEncoder(b),
			Timestamp: time.Now(),
		}

		partition, offset, err := f.producer.SendMessage(message)
		if err != nil {
			log.Error("can't send message", err)
			f.sentErr.WithLabelValues(f.cfg.Topic).Inc()
			errors++
		} else {
			log.Infof("user[%d] sent to partition %d at offset %d", user.Pnum, partition, offset)
			f.sent.WithLabelValues(f.cfg.Topic).Inc()
			successes++
		}
	}

	log.Printf("Successfully produced: %d; errors: %d", successes, errors)
}
