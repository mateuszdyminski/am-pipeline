package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"

	"github.com/mateuszdyminski/am-pipeline/feeder-api/pkg/version"
	"github.com/mateuszdyminski/am-pipeline/models"
)

func (s *Server) pumpUsers(w http.ResponseWriter, r *http.Request) {
	ip := userIP(r)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		s.receivedErr.WithLabelValues(ip).Inc()
		return
	}

	var users []models.User
	if err := json.Unmarshal(body, &users); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		s.receivedErr.WithLabelValues(ip).Inc()
		return
	}

	for _, user := range users {
		data, err := json.Marshal(user)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			s.receivedErr.WithLabelValues(ip).Inc()
			return
		}

		s.received.WithLabelValues(ip).Inc()

		if err := s.p.Pump(fmt.Sprintf("%d", user.Pnum), data); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
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
