-- Revert scope columns addition
ALTER TABLE preferences DROP CONSTRAINT IF EXISTS uq_preferences_resource_scope;

ALTER TABLE preferences ADD CONSTRAINT resource_type_name_unique
    UNIQUE (resource_type, resource_id, name);

ALTER TABLE preferences DROP COLUMN IF EXISTS scope_id;
ALTER TABLE preferences DROP COLUMN IF EXISTS scope_type;
