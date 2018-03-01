public class SimpleConsumer {
    public void consume(String zookeeper, String groupId, String topic) {
        Properties props = new Properties();
        props.put("zookeeper.connect", "localhost:2181");
        props.put("group.id", "group-1");
        props.put("auto.commit.interval.ms", "1000");
        ConsumerConnector consumer = Consumer.createJavaConsumerConnector( // HL
            new ConsumerConfig(props)); // HL
        
        Map<String, Integer> topicCount = new HashMap<>();
        topicCount.put(topic, 1);

        Map<String, List<KafkaStream<byte[], byte[]>>> consumerStreams = 
            consumer.createMessageStreams(topicCount);
        List<KafkaStream<byte[], byte[]>> streams = consumerStreams.get(topic); // HL
        for (final KafkaStream stream : streams) {
            ConsumerIterator<byte[], byte[]> it = stream.iterator();
            while (it.hasNext()) {
                System.out.println("Message from Single Topic: " + new String(it.next().message()));
            }
        }
        if (consumer != null) {
            consumer.shutdown();
        }
    }
}