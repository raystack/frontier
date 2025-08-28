ALTER TABLE auditlogs
    ADD COLUMN IF NOT EXISTS idempotency_key uuid, -- dont make it unique constraint
    ADD COLUMN IF NOT EXISTS resource jsonb DEFAULT '{}' NOT NULL,
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone,
    ADD COLUMN IF NOT EXISTS occurred_at timestamp with time zone DEFAULT now() NOT NULL;

-- Backfill occurred_at from created_at for existing records
UPDATE auditlogs
SET occurred_at = created_at
WHERE occurred_at IS NULL;