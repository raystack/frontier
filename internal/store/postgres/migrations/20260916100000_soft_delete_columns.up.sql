ALTER TABLE domains ADD COLUMN IF NOT EXISTS deleted_at timestamptz;
ALTER TABLE billing_checkouts ADD COLUMN IF NOT EXISTS deleted_at timestamptz;
ALTER TABLE billing_transactions ADD COLUMN IF NOT EXISTS deleted_at timestamptz;
