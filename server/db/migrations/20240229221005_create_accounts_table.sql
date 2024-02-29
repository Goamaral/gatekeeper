-- migrate:up
CREATE TABLE accounts (
  uuid CHAR(36) PRIMARY KEY,
  company_uuid CHAR(36),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  wallet_address CHAR(42) NOT NULL UNIQUE,
  metadata JSON
);

-- migrate:down
DROP TABLE accounts;