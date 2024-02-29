INSERT INTO "companies" (uuid, api_key, admin_account_uuid)
VALUES (
  "018df6cc-ab90-7592-ae2d-a5c3dd9a79f3",
  "018df6ccab907592ae2da5c3dd9a79f3AFF3MAUaKHt9DVuBBi4Jzw",
  "018df6cf-44a3-7c50-80fc-055b707fc6d4"
);

INSERT INTO "accounts" (uuid, company_uuid, wallet_address, metadata)
VALUES (
  "018df6cf-44a3-7c50-80fc-055b707fc6d4",
  "018df6cc-ab90-7592-ae2d-a5c3dd9a79f3",
  "0x25a3aaf7a4fF88A8aa53ff63CFE5e8C16ce93756",
  '{"email": "odor@gatekeeper.com"}'
);