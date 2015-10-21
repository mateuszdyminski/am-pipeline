package com.test;

import backtype.storm.task.TopologyContext;
import backtype.storm.topology.BasicOutputCollector;
import backtype.storm.topology.OutputFieldsDeclarer;
import backtype.storm.topology.base.BaseBasicBolt;
import backtype.storm.tuple.Fields;
import backtype.storm.tuple.Tuple;
import backtype.storm.tuple.Values;
import com.datastax.driver.core.BoundStatement;
import com.datastax.driver.core.Cluster;
import com.datastax.driver.core.PreparedStatement;
import com.datastax.driver.core.Session;
import com.datastax.driver.core.exceptions.InvalidQueryException;
import com.test.model.User;

import java.util.Map;

/**
 * User mdyminski
 */
public class Inserter extends BaseBasicBolt {

    private Session session;
    private PreparedStatement insertStatement;

    public Inserter() {
    }

    @Override
    public void execute(Tuple tuple, BasicOutputCollector basicOutputCollector) {
        User user = (User) tuple.getValue(0);

        BoundStatement statement = new BoundStatement(insertStatement);
        statement.setLong(0, user.id);
        statement.setString(1, user.email);
        if (user.dob != null && user.dob.equals("0000-00-00")) {
            user.dob = null;
        }
        statement.setString(2, user.dob);
        statement.setInt(3, user.weight);
        statement.setInt(4, user.height);
        statement.setString(5, user.nickname);
        statement.setInt(6, user.country);
        statement.setString(7, user.city);
        statement.setString(8, user.caption);
        statement.setDouble(9, user.location.lon);
        statement.setDouble(10, user.location.lat);
        statement.setInt(11, user.gender);

        session.execute(statement);

        basicOutputCollector.emit(new Values(user));
    }

    @Override
    public void declareOutputFields(OutputFieldsDeclarer declarer) {
        declarer.declare(new Fields("user"));
    }

    @Override
    public void prepare(Map stormConf, TopologyContext context) {
        super.prepare(stormConf, context);

        // initialize cassandra
        Cluster cluster = Cluster.builder().addContactPoints(new String[]{"cass"}).build();
        session = cluster.connect();
        // Create the schema if it does not exist.
        try {
            session.execute("USE am");
        } catch (InvalidQueryException e) {
            session.execute("CREATE KEYSPACE IF NOT EXISTS am WITH replication = {'class':'SimpleStrategy', 'replication_factor':1};");

            session.execute("CREATE TABLE IF NOT EXISTS am.users (\n" +
                    "  id bigint,\n" +
                    "  email text,\n" +
                    "  dob text,\n" +
                    "  weight int,\n" +
                    "  height int,\n" +
                    "  nickname text,\n" +
                    "  country int,\n" +
                    "  city text,\n" +
                    "  caption text,\n" +
                    "  longitude double,\n" +
                    "  latitude double,\n" +
                    "  gender int,\n" +
                    "  PRIMARY KEY (id));");
        }

        insertStatement = session.prepare("INSERT INTO am.users(id, email, dob, weight, height, nickname, country, city, caption, longitude, latitude, gender) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)");
    }
}
