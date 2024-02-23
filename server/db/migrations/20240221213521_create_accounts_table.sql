-- migrate:up
CREATE TABLE accounts (
  uuid CHAR(36) PRIMARY KEY,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  api_key CHAR(48) NOT NULL,
  email VARCHAR(255) NOT NULL,
  wallet_address CHAR(42) NOT NULL UNIQUE
);

-- migrate:down
DROP TABLE accounts;