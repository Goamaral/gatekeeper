-- migrate:up
CREATE TABLE challenges (
  id INTEGER PRIMARY KEY,
  wallet_address TEXT NOT NULL,
  token TEXT NOT NULL
);

-- migrate:down
DROP TABLE challenges;