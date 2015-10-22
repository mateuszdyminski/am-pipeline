package com.test;

import backtype.storm.Config;
import backtype.storm.StormSubmitter;
import backtype.storm.topology.TopologyBuilder;
import com.esotericsoftware.kryo.serializers.FieldSerializer;
import com.test.model.Location;
import com.test.model.User;
import storm.kafka.BrokerHosts;
import storm.kafka.KafkaSpout;
import storm.kafka.SpoutConfig;
import storm.kafka.ZkHosts;

/**
 * User mdyminski
 */
public class Pipeline {

    public static void main(String[] args) throws Exception {
        TopologyBuilder builder = new TopologyBuilder();

        BrokerHosts hosts = new ZkHosts("zk:2181");
        SpoutConfig spoutConfig = new SpoutConfig(hosts, "time-series", "/am-pipeline", "am-pipeline");
        spoutConfig.scheme = new UsersSpoutScheme();
        KafkaSpout kafkaSpout = new KafkaSpout(spoutConfig);

        builder.setSpout("spout", kafkaSpout, 1);

        builder.setBolt("insert", new Inserter(), 2).shuffleGrouping("spout");
        builder.setBolt("index", new Indexer(), 2).shuffleGrouping("insert");

        Config conf = new Config();
        conf.setDebug(true);
        conf.setFallBackOnJavaSerialization(false);
        conf.registerSerialization(User.class, FieldSerializer.class);
        conf.registerSerialization(Location.class, FieldSerializer.class);

        conf.setNumWorkers(2);
        StormSubmitter.submitTopologyWithProgressBar(args[0], conf, builder.createTopology());
    }
}
