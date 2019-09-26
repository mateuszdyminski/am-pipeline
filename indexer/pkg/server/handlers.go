package server

import (
	"encoding/json"
	"net/http"
	"sync/atomic"

	"github.com/mateuszdyminski/am-pipeline/indexer/pkg/version"
)

func (s *Server) version(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{
		"appName":       version.AppName,
		"version":       version.AppVersion,
		"buildTime":     version.BuildTime,
		"gitCommitHash": version.LastCommitHash,
		"gitCommitUser": version.LastCommitUser,
		"gitCommitTime": version.LastCommitTime,
	}

	d, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(d)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&healthy) == 1 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&ready) == 1 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
}
