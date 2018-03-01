package com.am.streams;

import com.am.streams.model.User;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.apache.kafka.common.serialization.Serde;
import org.apache.kafka.common.serialization.Serdes;
import org.apache.kafka.streams.KafkaStreams;
import org.apache.kafka.streams.StreamsBuilder;
import org.apache.kafka.streams.StreamsConfig;
import org.apache.kafka.streams.Topology;
import org.apache.kafka.streams.kstream.KStream;
import org.apache.kafka.streams.kstream.TimeWindows;
import org.apache.kafka.streams.kstream.Windowed;

import java.io.IOException;
import java.util.Optional;
import java.util.Properties;

public class Start {

    public static void main(final String[] args) throws Exception {
        final String bootstrapServers = args.length > 0 ? args[0] : "localhost:9092";
        final Properties streamsConfiguration = new Properties();
        streamsConfiguration.put(StreamsConfig.APPLICATION_ID_CONFIG, "am-example");
        streamsConfiguration.put(StreamsConfig.CLIENT_ID_CONFIG, "am-client");
        streamsConfiguration.put(StreamsConfig.BOOTSTRAP_SERVERS_CONFIG, bootstrapServers);
        streamsConfiguration.put(StreamsConfig.DEFAULT_KEY_SERDE_CLASS_CONFIG, Serdes.String().getClass().getName());
        streamsConfiguration.put(StreamsConfig.DEFAULT_VALUE_SERDE_CLASS_CONFIG, Serdes.ByteArray().getClass().getName());
        streamsConfiguration.put(StreamsConfig.COMMIT_INTERVAL_MS_CONFIG, 10 * 1000);
        streamsConfiguration.put(StreamsConfig.CACHE_MAX_BYTES_BUFFERING_CONFIG, 0);

        final StreamsBuilder builder = new StreamsBuilder();

        final Serde<Integer> integerSerde = Serdes.Integer();

        final Serde<Windowed<Integer>> windowedIntegerSerde = new WindowedSerde<>(integerSerde);

        final Serde<User> userSerde;

        final Topology topology = builder.build();

        final KStream<String, byte[]> users = builder.stream("time-series");

        final ObjectMapper om = new ObjectMapper();
        users
                .mapValues(bytes -> {
                    try {
                        return Optional.<User>of(om.readValue(bytes, User.class));
                    } catch (IOException e) {
                        System.out.printf("can't deserialize user. err: " + e.getMessage());
                        return Optional.<User>empty();
                    }
                })
                .filter((k,v) -> v.isPresent())
                .mapValues(v -> v.get())
                .groupBy((k, v) -> v.getCountry())
                .windowedBy(TimeWindows.of(10 * 1000L))
                .count()
                .toStream()
                .foreach((k, v) -> System.out.println("County: " + k.key() + " occurs in last 10s: " + v + " times!"));


        final KafkaStreams streams = new KafkaStreams(topology, streamsConfiguration);
        streams.cleanUp();
        streams.start();

        Runtime.getRuntime().addShutdownHook(new Thread(streams::close));
    }
}
