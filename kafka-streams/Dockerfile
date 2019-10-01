FROM openjdk:8u151-jre-slim

RUN mkdir -p /usr/share/streams

WORKDIR /usr/share/streams

ADD target/am-streams-1.0-SNAPSHOT-jar-with-dependencies.jar /usr/share/streams/am-streams-1.0-SNAPSHOT-jar-with-dependencies.jar

ENTRYPOINT java -jar /usr/share/streams/am-streams-1.0-SNAPSHOT-jar-with-dependencies.jar broker.kafka:9092
