package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	r "github.com/dancannon/gorethink"
	"github.com/mateuszdyminski/am-pipeline/models"
	"gopkg.in/olivere/elastic.v2"
)

var configPath string

// Config holds configuration of feeder.
type Config struct {
	Rethinkdb string
	Topic     string
	Elastics  []string
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

			log.Printf("Bulk with %v users indexed! Total indexed users: %v", BulkSize, enqued)

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
	session, err := r.Connect(r.ConnectOpts{Address: conf.Rethinkdb})
	if err != nil {
		log.Fatalln(err.Error())
	}

	var received int

	// Trap SIGINT to trigger a graceful shutdown.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	rows, err := r.DB("am").Table("users").Changes().Run(session)
	if err != nil {
		log.Fatalln(err.Error())
	}

	out := make(chan models.User, 1024)
	go func() {
		for {
			select {
			case <-signals:
				log.Printf("Start rethink driver closing")
				rows.Close()
				session.Close()
				log.Printf("Rethink driver closed!")
				return
			}
		}
	}()

	go func() {
		var value r.ChangeResponse
		for rows.Next(&value) {
			log.Infof("Get user changes: %+v", value)

			usr, err := models.ParseUser(value.NewValue)
			if err != nil {
				log.Fatalf("Can't decode rethinkDB changes. Err: %v", err)
			}

			out <- usr
			received++
		}

		close(out)
		log.Printf("Successfully consumed: %d users from rethinkdb", received)
	}()

	return out
}
