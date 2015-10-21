package com.test;

import backtype.storm.Config;
import backtype.storm.StormSubmitter;
import backtype.storm.topology.TopologyBuilder;
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

        builder.setBolt("insert", new Inserter(), 3).shuffleGrouping("spout");
        builder.setBolt("index", new Indexer(), 3).shuffleGrouping("insert");

        Config conf = new Config();
        conf.setDebug(true);

        System.out.println("Submitting topology");
        conf.setNumWorkers(4);
        StormSubmitter.submitTopologyWithProgressBar(args[0], conf, builder.createTopology());
        System.out.println("Topology submitted");
    }
}
