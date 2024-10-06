CREATE TABLE IF NOT EXISTS "schema_migrations" (version varchar(128) primary key);
CREATE TABLE challenges (
  id INTEGER PRIMARY KEY,
  wallet_address CHAR(42) NOT NULL,
  token CHAR(16) NOT NULL UNIQUE,
  expired_at TIMESTAMP NOT NULL
);
CREATE TABLE companies (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  api_key CHAR(48) NOT NULL
);
CREATE TABLE accounts (
  company_id INTEGER NOT NULL,
  wallet_address CHAR(42) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  metadata JSON,
  FOREIGN KEY (company_id) REFERENCES companies(id),
  PRIMARY KEY (company_id, wallet_address)
);
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20231122185055'),
  ('20240221213521'),
  ('20240229221005');
