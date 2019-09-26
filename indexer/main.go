package main

import (
	"flag"

	"github.com/mateuszdyminski/am-pipeline/indexer/pkg/config"
	"github.com/mateuszdyminski/am-pipeline/indexer/pkg/indexer"
	"github.com/mateuszdyminski/am-pipeline/indexer/pkg/server"
	"github.com/mateuszdyminski/am-pipeline/indexer/pkg/signals"

	log "github.com/sirupsen/logrus"
)

var configPath string

func init() {
	flag.Usage = func() {
		flag.PrintDefaults()
	}

	flag.StringVar(&configPath, "config", "config/conf.toml", "config path")
}

func main() {
	// load config
	flag.Parse()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal("can't load config file", err)
	}

	ctx := signals.SetupSignalContext()
	indexer, err := indexer.NewIndexer(cfg)
	if err != nil {
		log.Fatal("can't create indexer", err)
	}
	go indexer.Index()

	server.ListenAndServe(cfg, ctx)
}
