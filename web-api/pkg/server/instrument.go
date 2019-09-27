package server

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type Instrument struct {
	Histogram *prometheus.HistogramVec
	Counter   *prometheus.CounterVec
}

func NewInstrument() *Instrument {
	// used for monitoring and alerting (RED method)
	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "am",
		Subsystem: "analyzer",
		Name:      "http_request_duration_seconds",
		Help:      "Seconds spent serving HTTP requests.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"method", "path", "status"})
	// used for horizontal pod auto-scaling (Kubernetes HPA v2)
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "am",
			Subsystem: "analyzer",
			Name:      "http_requests_total",
			Help:      "The total number of HTTP requests.",
		},
		[]string{"status"},
	)

	prometheus.MustRegister(histogram)
	prometheus.MustRegister(counter)

	return &Instrument{
		Histogram: histogram,
		Counter:   counter,
	}
}

func (i Instrument) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		interceptor := &interceptor{ResponseWriter: w, statusCode: http.StatusOK}
		path := urlToLabel(r.URL.Path)
		next.ServeHTTP(interceptor, r)
		var (
			status = strconv.Itoa(interceptor.statusCode)
			took   = time.Since(begin)
		)
		log.Infof("%s - %s %s %s %s %s Took: %s",
			getRemoteAddr(r),
			r.Method,
			r.RequestURI,
			r.Proto,
			status,
			r.UserAgent(),
			took.String())

		i.Histogram.WithLabelValues(r.Method, path, status).Observe(took.Seconds())
		i.Counter.WithLabelValues(status).Inc()
	})
}

var invalidChars = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// converts a URL path to a string compatible with Prometheus label value.
func urlToLabel(path string) string {
	result := invalidChars.ReplaceAllString(path, "_")
	result = strings.ToLower(strings.Trim(result, "_"))
	if result == "" {
		result = "root"
	}
	return result
}

type interceptor struct {
	http.ResponseWriter
	statusCode int
	recorded   bool
}

func (i *interceptor) WriteHeader(code int) {
	if !i.recorded {
		i.statusCode = code
		i.recorded = true
	}
	i.ResponseWriter.WriteHeader(code)
}

func (i *interceptor) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := i.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("interceptor: can't cast parent ResponseWriter to Hijacker")
	}
	return hj.Hijack()
}

func getRemoteAddr(r *http.Request) string {
	forwaredFor := r.Header.Get("X-Forwarded-For")
	if forwaredFor == "" {
		return r.RemoteAddr
	}

	return forwaredFor
}
