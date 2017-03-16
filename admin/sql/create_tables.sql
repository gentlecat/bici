BEGIN;

CREATE TABLE athlete (
  id           BIGINT PRIMARY KEY,
  data         JSONB,
  access_token TEXT NOT NULL UNIQUE
);

CREATE TABLE summit (
  id     INT PRIMARY KEY,
  name   TEXT NOT NULL,
  points INT  NOT NULL CHECK (points > 0)
);

CREATE TABLE segment (
  id   INT PRIMARY KEY,
  data JSONB
);

CREATE TABLE summit_segments (
  summit_id  INT REFERENCES summit (id) ON DELETE CASCADE,
  segment_id BIGINT REFERENCES segment (id) ON DELETE CASCADE,

  PRIMARY KEY (summit_id, segment_id)
);

CREATE TABLE activity (
  id         BIGINT PRIMARY KEY,
  athlete_id BIGINT NOT NULL REFERENCES athlete (id) ON DELETE CASCADE,
  data       JSONB
);

CREATE TABLE activity_efforts (
  activity_id BIGINT NOT NULL REFERENCES activity (id) ON DELETE CASCADE,
  segment_id  BIGINT NOT NULL REFERENCES segment (id) ON DELETE CASCADE
);

COMMIT;
