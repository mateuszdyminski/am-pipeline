package indexer

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/mateuszdyminski/am-pipeline/indexer/pkg/config"
	"github.com/mateuszdyminski/am-pipeline/models"
	elastic "github.com/olivere/elastic/v7"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Indexer allows to Index data taken from Kafka in ElasticSearch
type Indexer struct {
	cfg           *config.Config
	kafkaConsumer sarama.ConsumerGroup
	esClient      *elastic.Client
	indexed       *prometheus.CounterVec
	indexedErr    *prometheus.CounterVec
	received      *prometheus.CounterVec
	receivedErr   *prometheus.CounterVec
}

// NewIndexer creates new Indexer.
func NewIndexer(cfg *config.Config) (*Indexer, error) {
	// kafka consumer group initialization
	config := sarama.NewConfig()
	config.Version = sarama.V2_3_0_0
	config.Consumer.Return.Errors = true
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	if cfg.ReadFromOldest {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	// init consumer
	brokers := cfg.Brokers
	group := "consumer-group"

	kafkaConsumer, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		return nil, fmt.Errorf("error while init consumer group. err: %s", err)
	}

	// elasticsearch client initialization
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}

	// connect to the cluster
	client, err := elastic.NewClient(
		elastic.SetURL(cfg.Elastics...),
		elastic.SetBasicAuth(cfg.ElasticUser, cfg.ElasticPassword),
		elastic.SetHttpClient(httpClient),
		elastic.SetSniff(false),
		elastic.SetScheme("https"),
	)
	if err != nil {
		return nil, fmt.Errorf("can't create elastic client. err: %v", err)
	}

	received := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "am",
			Subsystem: "indexer",
			Name:      "received_total",
			Help:      "The total number of received users to index.",
		},
		[]string{"topic"},
	)

	receivedErr := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "am",
			Subsystem: "indexer",
			Name:      "received_total_err",
			Help:      "The total number of errors during receiving users.",
		},
		[]string{"topic"},
	)

	indexed := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "am",
			Subsystem: "indexer",
			Name:      "indexed_total",
			Help:      "The total number of indexed users.",
		},
		[]string{"index"},
	)

	indexedErr := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "am",
			Subsystem: "indexer",
			Name:      "indexed_total_err",
			Help:      "The total number of errors during indexing users.",
		},
		[]string{"index"},
	)

	prometheus.Register(received)
	prometheus.Register(receivedErr)
	prometheus.Register(indexed)
	prometheus.Register(indexedErr)

	indexer := &Indexer{
		cfg:           cfg,
		kafkaConsumer: kafkaConsumer,
		esClient:      client,
		indexed:       indexed,
		indexedErr:    indexedErr,
		received:      received,
		receivedErr:   receivedErr,
	}

	return indexer, nil
}

// Index starts reading data from Kafka and indexing it in ELastic.
func (p *Indexer) Index() error {
	p.indexUsers(p.streamUsers())

	return nil
}

// BulkSize size of the bulk.
const BulkSize = 1

func (p *Indexer) indexUsers(users chan models.User) {
	exists, err := p.esClient.IndexExists("users").Do(context.Background())
	if err != nil {
		log.Fatalf("Can't check if index exists. Err: %v", err)
	}

	if !exists {
		log.Info("Creating index 'users'")
		// Create an index if not exists
		_, err = p.esClient.
			CreateIndex("users").
			BodyString(models.ElasticMappingString).
			Do(context.Background())
		if err != nil {
			log.Fatalf("Can't create index. Err: %v", err)
		}
	}

	var enqued int
	bulkRequest := p.esClient.Bulk()
	for user := range users {
		if enqued > 0 && enqued%BulkSize == 0 {
			if _, err := bulkRequest.Do(context.Background()); err != nil {
				p.indexedErr.WithLabelValues("users").Inc()
				log.Error("can't execute bulk. Err: %v", err)
				continue
			}

			p.indexed.WithLabelValues("users").Add(BulkSize)
			log.Infof("Bulk with %v users indexed! Total indexed users: %v", BulkSize, enqued)

			bulkRequest = p.esClient.Bulk()
		}

		bulkRequest.Add(
			elastic.NewBulkIndexRequest().
				Index("users").
				Type("_doc").
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

func (p *Indexer) streamUsers() chan models.User {
	out := make(chan models.User, 1024)
	topics := []string{p.cfg.Topic}
	ctx, cancel := context.WithCancel(context.Background())

	/**
	 * Setup a new Sarama consumer group
	 */
	consumer := Consumer{
		out:         out,
		ready:       make(chan bool),
		received:    p.received,
		receivedErr: p.receivedErr,
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if err := p.kafkaConsumer.Consume(ctx, topics, &consumer); err != nil {
				log.Panicf("Error from consumer: %v", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	log.Println("Sarama consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-ctx.Done():
			log.Println("terminating: context cancelled")
		case <-sigterm:
			log.Println("terminating: via signal")
		}
		cancel()
		wg.Wait()
		if err := p.kafkaConsumer.Close(); err != nil {
			log.Panicf("Error closing client: %v", err)
		}
	}()

	return out
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	counter     int
	out         chan models.User
	ready       chan bool
	received    *prometheus.CounterVec
	receivedErr *prometheus.CounterVec
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log.Infof("received message: %s", string(msg.Value))

		var user models.User
		if err := json.Unmarshal(msg.Value, &user); err != nil {
			consumer.receivedErr.WithLabelValues(msg.Topic).Inc()
			session.MarkMessage(msg, fmt.Sprintf("can't unmarshal data from queue. err: %s", err.Error()))
			log.Error("can't unmarshal data from queue", err)
			continue
		}

		if *user.Dob == "0000-00-00" {
			user.Dob = nil
		}

		consumer.out <- user

		session.MarkMessage(msg, "")

		consumer.counter++
		consumer.received.WithLabelValues(msg.Topic).Inc()

		if consumer.counter%1000 == 0 {
			log.Infof("received %n messages from Kafka", consumer.counter)
		}
	}

	return nil
}
