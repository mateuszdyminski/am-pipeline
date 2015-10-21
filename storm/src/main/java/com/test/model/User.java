package com.test.model;

import com.fasterxml.jackson.annotation.JsonInclude;

import java.io.Serializable;

/**
 * User mdyminski
 */
@JsonInclude(JsonInclude.Include.NON_NULL)
public class User implements Serializable {

    public long id;
    public String email;
    public String dob;
    public int weight;
    public int height;
    public String nickname;
    public int country;
    public String city;
    public String caption;
    public Location location;
    public int gender;

    @Override
    public String toString() {
        return "User{" +
                "id=" + id +
                ", email='" + email + '\'' +
                ", dob='" + dob + '\'' +
                ", weight=" + weight +
                ", height=" + height +
                ", nickname='" + nickname + '\'' +
                ", country=" + country +
                ", city='" + city + '\'' +
                ", caption='" + caption + '\'' +
                ", location=" + location +
                ", gender=" + gender +
                '}';
    }

    public static final String ES_MAPPINGS = "{\n" +
            "            \"settings\" : {\n" +
            "                \"analysis\" : {\n" +
            "                    \"filter\" : {\n" +
            "                        \"autocomplete\" : {\n" +
            "                            \"type\" : \"edge_ngram\",\n" +
            "                            \"min_gram\" : 1,\n" +
            "                            \"max_gram\" : 20\n" +
            "                        }\n" +
            "                    },\n" +
            "                    \"analyzer\" : {\n" +
            "                        \"nickname\" : {\n" +
            "                            \"type\" : \"standard\",\n" +
            "                            \"stopwords\" : []\n" +
            "                        },\n" +
            "                        \"nickname_autocomplete\" : {\n" +
            "                            \"type\" : \"custom\",\n" +
            "                            \"tokenizer\" : \"standard\",\n" +
            "                            \"filter\" : [\"lowercase\", \"autocomplete\"]\n" +
            "                        }\n" +
            "                    }\n" +
            "                }\n" +
            "            },\n" +
            "            \"mappings\" : {\n" +
            "                \"user\" : {\n" +
            "                    \"properties\" : {\n" +
            "                        \"id\" : { \"type\" : \"string\", \"index\" : \"not_analyzed\" },\n" +
            "                        \"email\" : { \"type\" : \"string\", \"index\" : \"not_analyzed\" },\n" +
            "                        \"dob\" : { \"type\" : \"date\" },\n" +
            "                        \"weight\" : { \"type\" : \"integer\" },\n" +
            "                        \"height\" : { \"type\" : \"integer\" },\n" +
            "                        \"nickname\" : {\n" +
            "                            \"type\" : \"multi_field\",\n" +
            "                            \"fields\" : {\n" +
            "                                \"nickname\" : {\n" +
            "                                    \"type\" : \"string\",\n" +
            "                                    \"analyzer\" : \"nickname\"\n" +
            "                                },\n" +
            "                                \"autocomplete\" : {\n" +
            "                                    \"type\" : \"string\",\n" +
            "                                    \"index_analyzer\" : \"nickname_autocomplete\",\n" +
            "                                    \"search_analyzer\" : \"nickname\"\n" +
            "                                }\n" +
            "                            }\n" +
            "                        },\n" +
            "                        \"country\" : { \"type\" : \"integer\" },\n" +
            "                        \"city\" : { \"type\" : \"string\" },\n" +
            "                        \"caption\" : { \"type\" : \"string\" },\n" +
            "                        \"location\" : { \"type\" : \"geo_point\" },\n" +
            "                        \"gender\" : { \"type\" : \"integer\" }\n" +
            "                    }\n" +
            "                }\n" +
            "            }\n" +
            "        }";
}

