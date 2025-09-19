ALTER TABLE sessions 
    DROP COLUMN IF EXISTS deleted_at,
    DROP COLUMN IF EXISTS updated_at;