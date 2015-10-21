package com.test;

import backtype.storm.spout.MultiScheme;
import backtype.storm.tuple.Fields;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.test.model.User;
import org.apache.storm.guava.collect.ImmutableList;

import java.io.IOException;
import java.util.List;

import static backtype.storm.utils.Utils.tuple;

/**
 * User mdyminski
 */
public class UsersSpoutScheme implements MultiScheme {

    private static final long serialVersionUID = -3267275196936420794L;

    public static final ObjectMapper mapper = new ObjectMapper();

    @Override
    public Iterable<List<Object>> deserialize(byte[] ser) {
        try {
            User user = mapper.readValue(ser, User.class);

            return ImmutableList.of(tuple(user));
        } catch (IOException e) {
            throw new RuntimeException("can't deserialize user!", e);
        }
    }

    @Override
    public Fields getOutputFields() {
        return new Fields("user");
    }
}
