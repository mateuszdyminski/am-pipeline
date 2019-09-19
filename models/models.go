package models

// User holds info about AM user.
type User struct {
	Pnum     int64     `json:"id,omitempty" gorethink:"id"`
	Email    *string   `json:"email,omitempty" gorethink:"email"`
	Dob      *string   `json:"dob,omitempty" gorethink:"dob"`
	Weight   *int      `json:"weight,omitempty" gorethink:"weight"`
	Height   *int      `json:"height,omitempty" gorethink:"height"`
	Nickname *string   `json:"nickname,omitempty" gorethink:"nickname"`
	Country  int       `json:"country,omitempty" gorethink:"country"`
	City     *string   `json:"city,omitempty" gorethink:"city"`
	Caption  *string   `json:"caption,omitempty" gorethink:"caption"`
	Location *Location `json:"location,omitempty" gorethink:"location"`
	Gender   *int      `json:"gender,omitempty" gorethink:"gender"`
	Score    *float64  `json:"score,omitempty"`
}

func ParseUser(value interface{}) (User, error) {
	values := value.(map[string]interface{})
	usr := User{}

	id := values["id"].(float64)
	usr.Pnum = int64(id)

	if values["email"] != nil {
		email := values["email"].(string)
		usr.Email = &email
	}

	if values["dob"] != nil {
		dob := values["dob"].(string)
		usr.Dob = &dob
	}

	weight := int(values["weight"].(float64))
	usr.Weight = &weight

	height := int(values["height"].(float64))
	usr.Height = &height

	nick := values["nickname"].(string)
	usr.Nickname = &nick

	usr.Country = int(values["country"].(float64))

	city := values["city"].(string)
	usr.City = &city

	caption := values["caption"].(string)
	usr.Caption = &caption

	usr.Location = &Location{}

	locValues := values["location"].(map[string]interface{})

	usr.Location.Latitude = locValues["lat"].(float64)
	usr.Location.Longitude = locValues["lon"].(float64)

	gender := int(values["gender"].(float64))
	usr.Gender = &gender

	return usr, nil
}

// Location holds inforamtion
type Location struct {
	Longitude float64 `json:"lon,omitempty" gorethink:"lon"`
	Latitude  float64 `json:"lat,omitempty" gorethink:"lat"`
}

// ElasticMappingString describes index of user in Elasticsearch.
const ElasticMappingString = `
        {
            "settings" : {
                "index.mapping.single_type": true,
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
                "_doc" : {
                    "properties" : {
                        "id" : { "type" : "text", "index" : "not_analyzed" },
                        "email" : { "type" : "text", "index" : "not_analyzed" },
                        "dob" : { "type" : "date" },
                        "weight" : { "type" : "integer" },
                        "height" : { "type" : "integer" },
                        "nickname" : {
                            "type" : "text",
                            "analyzer": "nickname",
                            "fields" : {
                                "autocomplete" : {
                                    "type" : "text",
                                    "analyzer" : "nickname_autocomplete",
                                    "search_analyzer" : "nickname"
                                }
                            }
                        },
                        "country" : { "type" : "integer" },
                        "city" : { "type" : "text" },
                        "caption" : { "type" : "text", "index" : "analyzed" },
                        "location" : { "type" : "geo_point" },
                        "gender" : { "type" : "integer" }
                    }
                }
            }
        }`
