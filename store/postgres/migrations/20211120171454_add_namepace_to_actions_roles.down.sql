ALTER TABLE roles
    DROP COLUMN IF EXISTS namespace_id;

ALTER TABLE actions
    DROP COLUMN IF EXISTS namespace_id;