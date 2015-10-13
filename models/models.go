package models

// User holds info about AM user.
type User struct {
	Pnum     int64     `json:"id,omitempty"`
	Email    string    `json:"email,omitempty"`
	Dob      *string   `json:"dob,omitempty"`
	Weight   *int      `json:"weight,omitempty"`
	Height   *int      `json:"height,omitempty"`
	Nickname *string   `json:"nickname,omitempty"`
	Country  int       `json:"country,omitempty"`
	City     *string   `json:"city,omitempty"`
	Caption  *string   `json:"caption,omitempty"`
	Location *Location `json:"location,omitempty"`
	Gender   *int      `json:"gender,omitempty"`
	Score    *float64  `json:"score,omitempty"`
}

// Location holds inforamtion
type Location struct {
	Longitude float64 `json:"lon,omitempty"`
	Latitude  float64 `json:"lat,omitempty"`
}

// ElasticMappingString describes index of user in ELasticsearch.
const ElasticMappingString = `
        {
            "settings" : {
                "analysis" : {
                    "filter" : {
                        "autocomplete" : {
                            "type" : "edge_ngram",
                            "min_gram" : 1,
                            "max_gram" : 20
                        }
                    },
                    "analyzer" : {
                        "nickname" : {
                            "type" : "standard",
                            "stopwords" : []
                        },
                        "nickname_autocomplete" : {
                            "type" : "custom",
                            "tokenizer" : "standard",
                            "filter" : ["lowercase", "autocomplete"]
                        }
                    }
                }
            },
            "mappings" : {
                "user" : {
                    "properties" : {
                        "id" : { "type" : "string", "index" : "not_analyzed" },
                        "email" : { "type" : "string", "index" : "not_analyzed" },
                        "dob" : { "type" : "date" },
                        "weight" : { "type" : "integer" },
                        "height" : { "type" : "integer" },
                        "nickname" : {
                            "type" : "multi_field",
                            "fields" : {
                                "nickname" : {
                                    "type" : "string",
                                    "analyzer" : "nickname"
                                },
                                "autocomplete" : {
                                    "type" : "string",
                                    "index_analyzer" : "nickname_autocomplete",
                                    "search_analyzer" : "nickname"
                                }
                            }
                        },
                        "country" : { "type" : "integer" },
                        "city" : { "type" : "string" },
                        "caption" : { "type" : "string" },
                        "location" : { "type" : "geo_point" },
                        "gender" : { "type" : "integer" }
                    }
                }
            }
        }`
