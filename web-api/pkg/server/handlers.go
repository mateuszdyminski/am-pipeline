package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/mateuszdyminski/am-pipeline/web-api/pkg/version"
)

func (s *Server) users(w http.ResponseWriter, r *http.Request) {
	field := r.URL.Query().Get("field")
	query := r.URL.Query().Get("query")
	distance := r.URL.Query().Get("distance")
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	size := r.URL.Query().Get("l")
	skip := r.URL.Query().Get("s")
	wildcard := r.URL.Query().Get("w")

	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		sizeInt = 100
	}

	skipInt, err := strconv.Atoi(skip)
	if err != nil {
		skipInt = 0
	}

	if field == "" {
		field = "_all"
	}

	if query == "" && distance == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("query and distance can't be empty!"))
		return
	}

	users, err := s.analyzer.Search(field, query, distance, latStr, lonStr, sizeInt, skipInt, wildcard)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("can't find users! err: %s", err.Error())))
		return
	}

	json, err := json.Marshal(users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("can't serialize users! err: %s", err.Error())))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func (s *Server) nickAutocomplete(w http.ResponseWriter, r *http.Request) {
	nick := r.URL.Query().Get("nick")
	if nick == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("nick can't be empty!"))
		return
	}

	nicks, err := s.analyzer.Nicks(nick)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("can't find nicks! err: %s", err.Error())))
		return
	}

	json, err := json.Marshal(nicks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("can't serialize nicks! err: %s", err.Error())))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

func (s *Server) aggregations(w http.ResponseWriter, r *http.Request) {
	field := r.URL.Query().Get("field")
	if field == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("field can't be empty!"))
		return
	}

	buckets, err := s.analyzer.Aggregations(field)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("can't find buckets! err: %s", err.Error())))
		return
	}

	json, err := json.Marshal(buckets)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("can't serialize buckets! err: %s", err.Error())))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(json)
}

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
