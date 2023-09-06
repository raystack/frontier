CREATE TABLE IF NOT EXISTS preferences (
   id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
   name text NOT NULL,
   value text,
   resource_type text REFERENCES namespaces (name) ON DELETE CASCADE,
   resource_id text NOT NULL,
   created_at timestamptz NOT NULL DEFAULT NOW(),
   updated_at timestamptz NOT NULL DEFAULT NOW(),
   CONSTRAINT resource_type_name_unique UNIQUE (resource_type, resource_id, name)
);
CREATE INDEX IF NOT EXISTS preferences_resource_type_resource_id_idx ON preferences(resource_type, resource_id);