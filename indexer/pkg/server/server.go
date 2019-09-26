package server

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/mateuszdyminski/am-pipeline/indexer/pkg/config"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

var (
	healthy int32 = 1
	ready   int32 = 1
)

type Server struct {
	mux *mux.Router
}

func NewServer(cfg *config.Config, options ...func(*Server)) *Server {
	s := &Server{mux: mux.NewRouter()}

	for _, f := range options {
		f(s)
	}

	// general handlers
	s.mux.HandleFunc("/health", s.health)
	s.mux.HandleFunc("/ready", s.ready)
	s.mux.HandleFunc("/version", s.version)

	// metrics
	s.mux.Handle("/metrics", promhttp.Handler())

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", runtime.Version())

	s.mux.ServeHTTP(w, r)
}

func ListenAndServe(cfg *config.Config, cancelCtx context.Context) {
	inst := NewInstrument()
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      inst.Wrap(NewServer(cfg)),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 1 * time.Minute,
		IdleTimeout:  15 * time.Second,
	}

	// run server in background
	go func() {
		log.Info().Msgf("HTTP Server started at port: %d", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server crashed")
		}
	}()

	// wait for SIGTERM or SIGINT
	<-cancelCtx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()

	// all calls to /healthz and /readyz will fail from now on
	atomic.StoreInt32(&healthy, 0)
	atomic.StoreInt32(&ready, 0)

	time.Sleep(time.Duration(int64(3) * int64(time.Second)))

	log.Info().Msgf("Shutting down HTTP server with timeout: %v", time.Duration(3)*time.Second)

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("HTTP server graceful shutdown failed")
	} else {
		log.Info().Msg("HTTP server stopped")
	}
}
