package main

import (
	"flag"

	"github.com/mateuszdyminski/am-pipeline/web-api/pkg/analyzer"
	"github.com/mateuszdyminski/am-pipeline/web-api/pkg/config"
	"github.com/mateuszdyminski/am-pipeline/web-api/pkg/server"
	"github.com/mateuszdyminski/am-pipeline/web-api/pkg/signals"

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
	analyzer, err := analyzer.NewAnalyzer(cfg)
	if err != nil {
		log.Fatal("can't create analyzer", err)
	}

	server.ListenAndServe(ctx, cfg, analyzer)
}
