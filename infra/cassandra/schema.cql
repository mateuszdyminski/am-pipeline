CREATE KEYSPACE IF NOT EXISTS am WITH replication = {'class':'SimpleStrategy', 'replication_factor':1};

CREATE TABLE IF NOT EXISTS am.users (
  id bigint,
  email text,
  dob text,
  weight int,
  height int,
  nickname text,
  country int,
  city text,
  caption text,
  longitude text,
  latitude text,
  gender int,
  PRIMARY KEY (id));