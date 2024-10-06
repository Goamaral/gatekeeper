-- migrate:up
CREATE TABLE accounts (
  company_id INTEGER NOT NULL,
  wallet_address CHAR(42) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  metadata JSON,
  FOREIGN KEY (company_id) REFERENCES companies(id),
  PRIMARY KEY (company_id, wallet_address)
);

-- migrate:down
DROP TABLE accounts;