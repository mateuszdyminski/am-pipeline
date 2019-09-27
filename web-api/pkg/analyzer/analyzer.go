package analyzer

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mateuszdyminski/am-pipeline/models"
	"github.com/mateuszdyminski/am-pipeline/web-api/pkg/config"

	elastic "github.com/olivere/elastic/v7"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Analyzer allows to Analyzer data taken from ElasticSearch
type Analyzer struct {
	cfg          *config.Config
	esClient     *elastic.Client
	esRequest    *prometheus.CounterVec
	esRequestErr *prometheus.CounterVec
}

// NewAnalyzer creates new Analyzer.
func NewAnalyzer(cfg *config.Config) (*Analyzer, error) {
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

	esRequest := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "am",
			Subsystem: "analyzer",
			Name:      "elastic_request_success_total",
			Help:      "The total number of success request to Elastic.",
		},
		[]string{"index"},
	)

	esRequestErr := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "am",
			Subsystem: "analyzer",
			Name:      "elastic_request_error_total",
			Help:      "The total number of errors requests from Elastic.",
		},
		[]string{"index"},
	)

	prometheus.Register(esRequest)
	prometheus.Register(esRequestErr)

	analyzer := &Analyzer{
		cfg:          cfg,
		esClient:     client,
		esRequest:    esRequest,
		esRequestErr: esRequestErr,
	}

	return analyzer, nil
}

// Index starts reading data from Kafka and indexing it in ELastic.
func (a *Analyzer) Search(
	field string,
	query string,
	distance string,
	latStr string,
	lonStr string,
	sizeInt int,
	skipInt int,
	wildcard string,
) (*UsersResponse, error) {
	var elasticQuery elastic.Query
	var long, lat float64
	if distance != "" {
		var err error
		long, err = strconv.ParseFloat(lonStr, 64)
		if err != nil {
			return nil, fmt.Errorf("longitude can't be empty!")
		}

		lat, err = strconv.ParseFloat(latStr, 64)
		if err != nil {
			return nil, fmt.Errorf("latitude can't be empty!")
		}

		log.Infof("Start searching for users in distance: %s from (lat, lng): (%f, %f)", distance, lat, long)

		// Search with a geo query
		geoQuery := elastic.NewGeoDistanceQuery("location").Distance(distance + "km").Lat(lat).Lon(long)

		elasticQuery = elastic.NewBoolQuery().Must(elastic.NewMatchAllQuery()).Filter(geoQuery)
	} else {
		log.Infof("Start searching for users: %s", query)

		if wildcard == "true" {
			// Search for
			elasticQuery = elastic.NewSimpleQueryStringQuery(query).Field(field)
		} else {
			// Search with a term query
			elasticQuery = elastic.NewMatchQuery(field, query)
		}
	}

	searchResult, err := a.esClient.Search().
		Index("users").
		Query(elasticQuery).
		From(skipInt).Size(sizeInt).
		Do(context.Background())
	if err != nil {
		a.esRequestErr.WithLabelValues("users").Inc()
		return nil, fmt.Errorf("can't search for users. err: %w", err)
	}
	a.esRequest.WithLabelValues("users").Inc()

	response := UsersResponse{}

	// Here's how you iterate through results with full control over each step.
	if searchResult.Hits != nil {
		log.Printf("Found a total of %d users", searchResult.TotalHits())

		response.Total = searchResult.TotalHits()

		// Iterate through results
		for _, hit := range searchResult.Hits.Hits {
			var u models.User
			err := json.Unmarshal(hit.Source, &u)
			if err != nil {
				return nil, fmt.Errorf("can't deserialize user. err: %w", err)
			}

			u.Score = hit.Score

			response.Users = append(response.Users, u)
			log.Infof("Found user %v", u.Email)
		}
	} else {
		return nil, errors.New("no users found")
	}

	return &response, nil
}

// Index starts reading data from Kafka and indexing it in ELastic.
func (a *Analyzer) Nicks(nick string) ([]string, error) {
	log.Infof("Start searching for autocomplete with nick: %s", nick)

	// Search with a term query
	matchQuery := elastic.NewMatchQuery("nickname.autocomplete", nick)
	searchResult, err := a.esClient.Search().
		Index("users").
		Query(matchQuery).
		From(0).Size(20).
		Do(context.Background())
	if err != nil {
		a.esRequestErr.WithLabelValues("users").Inc()
		return nil, fmt.Errorf("can't search for autocomplete. err: %w", err)
	}

	a.esRequest.WithLabelValues("users").Inc()

	nicks := make([]string, 0)

	// Here's how you iterate through results with full control over each step.
	if searchResult.Hits != nil {
		log.Infof("Found a total of %d nicks", searchResult.TotalHits())

		// Iterate through results
		for _, hit := range searchResult.Hits.Hits {
			var u models.User
			err := json.Unmarshal(hit.Source, &u)
			if err != nil {
				return nil, fmt.Errorf("can't deserialize user. err: %w", err)
			}

			nicks = append(nicks, *u.Nickname)
			log.Infof("Found user with nick %v", *u.Nickname)
		}
	} else {
		return nil, errors.New("no nicks found")
	}

	return nicks, nil
}

func (a *Analyzer) Aggregations(field string) ([]Bucket, error) {
	termsAgg := elastic.NewTermsAggregation().Field(field)

	searchResult, err := a.esClient.Search().
		Index("users").
		Aggregation("top_field", termsAgg).
		Do(context.Background())
	if err != nil {
		a.esRequestErr.WithLabelValues("users").Inc()
		return nil, fmt.Errorf("can't search for autocomplete. err: %w", err)
	}
	a.esRequest.WithLabelValues("users").Inc()

	buckets := make([]Bucket, 0)
	if searchResult.Aggregations != nil {
		terms, exists := searchResult.Aggregations.Terms("top_field")
		if !exists {
			return nil, errors.New("no buckets found")
		}
		if terms != nil && terms.Buckets != nil && len(terms.Buckets) > 0 {
			for _, b := range terms.Buckets {
				buckets = append(buckets, Bucket{b.Key, b.DocCount})
			}
		}
	} else {
		return nil, errors.New("no buckets found")
	}

	return buckets, nil
}

// UsersResponse holds information about users and total number of hits.
type UsersResponse struct {
	Users []models.User `json:"users,omitempty"`
	Total int64         `json:"total"`
}

// Bucket holds info about the key and the document count.
type Bucket struct {
	Key      interface{} `json:"key,omitempty"`
	DocCount int64       `json:"count"`
}
