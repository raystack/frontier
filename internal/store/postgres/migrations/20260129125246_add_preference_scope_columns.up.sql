-- Add scope columns for org-scoped user preferences
-- scope_type references namespaces table (like resource_type)
-- scope_id is text (like resource_id) to support different ID formats
-- Using zero values instead of NULL for global preferences:
-- - scope_type: 'app/platform' (global/unscoped)
-- - scope_id: '00000000-0000-0000-0000-000000000000' (zero UUID)
ALTER TABLE preferences ADD COLUMN scope_type text NOT NULL DEFAULT 'app/platform' REFERENCES namespaces (name) ON DELETE CASCADE;
ALTER TABLE preferences ADD COLUMN scope_id text NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';

-- Update unique constraint to include scope columns
ALTER TABLE preferences DROP CONSTRAINT resource_type_name_unique;

ALTER TABLE preferences ADD CONSTRAINT uq_preferences_resource_scope
    UNIQUE (resource_type, resource_id, scope_type, scope_id, name);
