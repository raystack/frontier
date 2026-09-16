ALTER TABLE domains DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE billing_checkouts DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE billing_transactions DROP COLUMN IF EXISTS deleted_at;
