package pumper

import (
	"fmt"
	"time"

	"github.com/mateuszdyminski/am-pipeline/feeder-api/pkg/config"

	"github.com/Shopify/sarama"
	"github.com/prometheus/client_golang/prometheus"
)

// Pumper allows to pump data into Kafka
type Pumper struct {
	cfg      *config.Config
	producer sarama.SyncProducer
	sent     *prometheus.CounterVec
	sentErr  *prometheus.CounterVec
}

// NewPumper creates new Pumper.
func NewPumper(cfg *config.Config) (*Pumper, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_3_0_0
	config.Producer.Retry.Max = 10
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewRandomPartitioner

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("can't create kafka producer: %w", err)
	}

	sent := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "am",
			Subsystem: "feeder_api",
			Name:      "sent_total",
			Help:      "The total number of sent users to Kafka.",
		},
		[]string{"topic"},
	)

	sentErr := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "am",
			Subsystem: "feeder_api",
			Name:      "sent_total_err",
			Help:      "The total number of sent errors.",
		},
		[]string{"topic"},
	)

	prometheus.Register(sent)
	prometheus.Register(sentErr)

	pumper := &Pumper{
		cfg:      cfg,
		producer: producer,
		sent:     sent,
		sentErr:  sentErr,
	}

	return pumper, nil
}

// Pump sends payload to Apache Kafka.
func (p *Pumper) Pump(key string, data []byte) error {
	message := &sarama.ProducerMessage{
		Topic:     p.cfg.Topic,
		Value:     sarama.ByteEncoder(data),
		Timestamp: time.Now(),
	}

	_, _, err := p.producer.SendMessage(message)

	if err != nil {
		p.sentErr.WithLabelValues(p.cfg.Topic).Inc()
	} else {
		p.sent.WithLabelValues(p.cfg.Topic).Inc()
	}

	return err
}
