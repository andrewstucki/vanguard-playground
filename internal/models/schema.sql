CREATE TABLE IF NOT EXISTS messages (
  id   TEXT PRIMARY KEY,
  text TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS sent_messages (
  id TEXT PRIMARY KEY,
  message_id TEXT NOT NULL,
  text TEXT NOT NULL,
  result TEXT NOT NULL
);