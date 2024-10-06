-- migrate:up
CREATE TABLE companies (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  api_key CHAR(48) NOT NULL
);

-- migrate:down
DROP TABLE companies;