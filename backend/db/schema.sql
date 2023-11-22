CREATE TABLE IF NOT EXISTS "schema_migrations" (version varchar(128) primary key);
CREATE TABLE challenges (
  id INTEGER PRIMARY KEY,
  wallet_address TEXT NOT NULL,
  token TEXT NOT NULL
);
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20231122185055');
