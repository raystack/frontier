ALTER TABLE resources
    DROP COLUMN IF EXISTS user_id,
    DROP CONSTRAINT IF EXISTS either_user_or_group;