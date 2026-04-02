-- Replace case-sensitive unique index with case-insensitive one
DROP INDEX IF EXISTS idx_user_pats_unique_title;
CREATE UNIQUE INDEX idx_user_pats_unique_title ON user_pats(user_id, org_id, LOWER(title)) WHERE deleted_at IS NULL;