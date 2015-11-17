package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	standardLog "log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/mateuszdyminski/am-pipeline/models"
	"gopkg.in/olivere/elastic.v2"
)

var configPath string

// Config holds configuration of feeder.
type Config struct {
	Host     string
	Statics  string
	Elastics []string
}

func init() {
	flag.Usage = func() {
		flag.PrintDefaults()
	}

	flag.StringVar(&configPath, "config", "config/conf.toml", "config path")

	log.SetOutput(os.Stderr)
}

func main() {
	// load config
	flag.Parse()

	bytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Can't open config file!")
	}

	var conf Config
	if err := toml.Unmarshal(bytes, &conf); err != nil {
		log.Fatalf("Can't decode config file!")
	}

	log.Printf("Config: %+v", conf)

	launchServer(&conf)
}

func launchServer(conf *Config) {
	client, err := elastic.NewClient(elastic.SetURL(conf.Elastics...), elastic.SetTraceLog(standardLog.New(os.Stdout, "", standardLog.LstdFlags)))
	if err != nil {
		log.Fatalf("Can't create elastic client. Err: %v", err)
	}

	api := &RestApi{client}

	r := mux.NewRouter()
	// Handle routes
	var statics StaticRoutes
	r.HandleFunc("/restapi/users", api.users).Methods("GET")
	r.HandleFunc("/restapi/autocomplete", api.nickAutocomplete).Methods("GET")
	r.HandleFunc("/restapi/aggregations", api.aggregations).Methods("GET")

	r.Handle("/{path:.*}", http.FileServer(append(statics, http.Dir(conf.Statics)))).Name("static")

	http.Handle("/", loggingHandler{r})

	// Listen on hostname:port
	log.Printf("Listening on %s...", conf.Host)
	if err = http.ListenAndServe(conf.Host, nil); err != nil {
		log.Fatalf("Error: %s", err)
	}
}

type StaticRoutes []http.FileSystem

func (sr StaticRoutes) Open(name string) (f http.File, err error) {
	for _, s := range sr {
		if f, err = s.Open(name); err == nil {
			f = disabledDirListing{f}
			return
		}
	}
	return
}

type disabledDirListing struct {
	http.File
}

func (f disabledDirListing) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

type loggingHandler struct {
	http.Handler
}

func (h loggingHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// path := req.URL.Path
	t := time.Now()
	h.Handler.ServeHTTP(w, req)

	elapsed := time.Since(t)
	log.Printf("%s [%s] \"%s %s %s\" \"%s\" \"%s\" \"Took: %s\"", req.RemoteAddr,
		t.Format("02/Jan/2006:15:04:05 -0700"), req.Method, req.RequestURI, req.Proto, req.Referer(), req.UserAgent(), elapsed)
}

type RestApi struct {
	e *elastic.Client
}

func (r *RestApi) users(w http.ResponseWriter, req *http.Request) {
	field := req.URL.Query().Get("field")
	query := req.URL.Query().Get("query")
	distance := req.URL.Query().Get("distance")
	latStr := req.URL.Query().Get("lat")
	lonStr := req.URL.Query().Get("lon")
	size := req.URL.Query().Get("l")
	skip := req.URL.Query().Get("s")
	wildcard := req.URL.Query().Get("w")

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
		fmt.Fprintf(w, "Query and Distance can't be empty!")
		return
	}

	var elasticQuery elastic.Query
	var long, lat float64
	if distance != "" {
		long, err = strconv.ParseFloat(lonStr, 64)
		if err != nil {
			fmt.Fprintf(w, "Longitude can't be empty!")
			return
		}

		lat, err = strconv.ParseFloat(latStr, 64)
		if err != nil {
			fmt.Fprintf(w, "Latitude can't be empty!")
			return
		}

		log.Infof("Start searching for users in distance: %s from (lat, lng): (%f, %f)", distance, lat, long)

		// Search with a geo query
		geoQuery := elastic.NewGeoDistanceFilter("location").Distance(distance + "km").Lat(lat).Lon(long)
		elasticQuery = elastic.NewFilteredQuery(elastic.NewMatchAllQuery()).Filter(geoQuery)
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

	searchResult, err := r.e.Search().
		Index("users").
		Type("user").
		Query(elasticQuery).
		From(skipInt).Size(sizeInt).
		Do()
	if err != nil {
		fmt.Fprintf(w, "Can't search for users. Err: %v", err)
		return
	}

	response := UsersResponse{}

	// Here's how you iterate through results with full control over each step.
	if searchResult.Hits != nil {
		log.Printf("Found a total of %d users", searchResult.Hits.TotalHits)

		response.Total = searchResult.Hits.TotalHits

		// Iterate through results
		for _, hit := range searchResult.Hits.Hits {
			var u models.User
			err := json.Unmarshal(*hit.Source, &u)
			if err != nil {
				fmt.Fprintf(w, "Can't deserialize user. Err: %v", err)
				return
			}

			u.Score = hit.Score

			response.Users = append(response.Users, u)
			log.Infof("Found user %v", u.Email)
		}
	} else {
		fmt.Fprintf(w, "Found no users")
		return
	}

	json, err := json.Marshal(response)
	if err != nil {
		fmt.Fprintf(w, "Can't serialize users. Err: %v", err)
		return
	}

	w.Write(json)
}

func (r *RestApi) nickAutocomplete(w http.ResponseWriter, req *http.Request) {
	nick := req.URL.Query().Get("nick")
	if nick == "" {
		fmt.Fprintf(w, "Query can't be empty!")
		return
	}

	log.Infof("Start searching for autocomplete with nick: %s", nick)

	// Search with a term query
	matchQuery := elastic.NewMatchQuery("nickname.autocomplete", nick)
	searchResult, err := r.e.Search().
		Index("users").
		Type("user").
		Query(&matchQuery).
		From(0).Size(20).
		Do()
	if err != nil {
		fmt.Fprintf(w, "Can't search for autocomplete. Err: %v", err)
		return
	}

	nicks := make([]string, 0)

	// Here's how you iterate through results with full control over each step.
	if searchResult.Hits != nil {
		log.Infof("Found a total of %d nicks", searchResult.Hits.TotalHits)

		// Iterate through results
		for _, hit := range searchResult.Hits.Hits {
			var u models.User
			err := json.Unmarshal(*hit.Source, &u)
			if err != nil {
				fmt.Fprintf(w, "Can't deserialize user. Err: %v", err)
				return
			}

			nicks = append(nicks, *u.Nickname)
			log.Infof("Found user with nick %v", *u.Nickname)
		}
	} else {
		fmt.Fprintf(w, "Found no users")
		return
	}

	json, err := json.Marshal(nicks)
	if err != nil {
		fmt.Fprintf(w, "Can't serialize users. Err: %v", err)
		return
	}

	w.Write(json)
}

func (r *RestApi) aggregations(w http.ResponseWriter, req *http.Request) {
	field := req.URL.Query().Get("field")
	if field == "" {
		fmt.Fprintf(w, "Field can't be empty!")
		return
	}

	log.Printf("Start searching for aggregations of field: %s", field)

	termsAgg := elastic.NewTermsAggregation().Field(field)

	searchResult, err := r.e.Search().
		Index("users").
		Type("user").
		Aggregation("top_field", termsAgg).
		Do()
	if err != nil {
		fmt.Fprintf(w, "Can't search for autocomplete. Err: %v", err)
		return
	}

	buckets := make([]Bucket, 0)
	if searchResult.Aggregations != nil {
		terms, exists := searchResult.Aggregations.Terms("top_field")
		if !exists {
			fmt.Fprintf(w, "Found no buckets")
			return
		}
		if terms != nil && terms.Buckets != nil && len(terms.Buckets) > 0 {
			for _, b := range terms.Buckets {
				buckets = append(buckets, Bucket{b.Key, b.DocCount})
			}
		}
	} else {
		fmt.Fprintf(w, "Found no buckets")
		return
	}

	json, err := json.Marshal(buckets)
	if err != nil {
		fmt.Fprintf(w, "Can't serialize buckets. Err: %v", err)
		return
	}

	w.Write(json)
}

// UsersResponse holds information about users and total number of hits.
type UsersResponse struct {
	Users []models.User `json:"users,omitempty"`
	Total int64         `json:"total,omitempty"`
}

// Bucket holds info about the key and the document count.
type Bucket struct {
	Key      interface{} `json:"key,omitempty"`
	DocCount int64       `json:"count,omitempty"`
}
