-- migrate:up
CREATE TABLE companies (
  uuid CHAR(36) PRIMARY KEY,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  api_key CHAR(48) NOT NULL,
  admin_account_uuid CHAR(36)
);

-- migrate:down
DROP TABLE companies;