package config

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// Config holds configuration of feeder.
type Config struct {
	// Kafka config
	Brokers []string
	Topic   string

	// HTTP Server config
	HTTPPort           int
	EnableCORS         bool

	// Elastisearch config
	Elastics        []string
	ElasticUser     string
	ElasticPassword string
}

// LoadConfig loads config from env vars.
func LoadConfig(configPath string) (*Config, error) {
	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var conf Config
	if err := toml.Unmarshal(bytes, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
