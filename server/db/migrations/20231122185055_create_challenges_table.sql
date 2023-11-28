-- migrate:up
CREATE TABLE challenges (
  id INTEGER PRIMARY KEY,
  wallet_address CHAR(42) NOT NULL,
  token CHAR(16) NOT NULL UNIQUE,
  expired_at DATETIME NOT NULL
);

-- migrate:down
DROP TABLE challenges;