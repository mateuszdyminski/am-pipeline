package com.test;

import backtype.storm.task.TopologyContext;
import backtype.storm.topology.BasicOutputCollector;
import backtype.storm.topology.OutputFieldsDeclarer;
import backtype.storm.topology.base.BaseBasicBolt;
import backtype.storm.tuple.Fields;
import backtype.storm.tuple.Tuple;
import backtype.storm.tuple.Values;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.test.model.User;
import org.elasticsearch.action.admin.indices.create.CreateIndexRequestBuilder;
import org.elasticsearch.action.index.IndexRequestBuilder;
import org.elasticsearch.action.index.IndexResponse;
import org.elasticsearch.client.Client;
import org.elasticsearch.client.transport.TransportClient;
import org.elasticsearch.common.transport.InetSocketTransportAddress;
import org.elasticsearch.index.mapper.MapperParsingException;

import java.io.IOException;
import java.util.Map;

/**
 * User mdyminski
 */
public class Indexer extends BaseBasicBolt {

    private Client client;
    private ObjectMapper mapper;

    public Indexer() {
    }

    @Override
    public void execute(Tuple tuple, BasicOutputCollector collector) {
        User user = (User) tuple.getValue(0);

        try {
            IndexRequestBuilder indexRequestBuilder = client.prepareIndex("users", "user", "" + user.id);
            byte[] json = mapper.writeValueAsBytes(user);
            indexRequestBuilder.setSource(json).execute().actionGet();

            collector.emit(new Values(user));
        } catch (IOException e) {
            throw new RuntimeException("can't index user: " + user.toString(), e);
        } catch (MapperParsingException e) {
            throw new RuntimeException("can't index user: " + user.toString(), e);
        }
    }

    @Override
    public void declareOutputFields(OutputFieldsDeclarer declarer) {
        declarer.declare(new Fields("user"));
    }

    @Override
    public void prepare(Map stormConf, TopologyContext context) {
        super.prepare(stormConf, context);

        mapper = new ObjectMapper();
        client = new TransportClient().addTransportAddress(new InetSocketTransportAddress("es", 9300));

        try {
            CreateIndexRequestBuilder createIndexRequestBuilder = client.
                    admin().
                    indices().
                    prepareCreate("users").
                    setSource(User.ES_MAPPINGS);
            createIndexRequestBuilder.execute().actionGet();
        } catch (Exception e) {
            // do nothing : ) index could already exists
        }
    }
}