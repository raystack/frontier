CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
DROP TABLE IF EXISTS namespaces CASCADE;
CREATE TABLE IF NOT EXISTS namespaces(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name text NOT NULL UNIQUE,
    metadata jsonb,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
    );
DROP TABLE IF EXISTS organizations CASCADE;
CREATE TABLE IF NOT EXISTS organizations (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name text UNIQUE NOT NULL,
    title text,
    metadata jsonb,
    state text DEFAULT 'enabled',
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
    );
DROP TABLE IF EXISTS projects CASCADE;
CREATE TABLE IF NOT EXISTS projects (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name text UNIQUE NOT NULL,
    title text,
    org_id uuid NOT NULL REFERENCES organizations(id),
    metadata jsonb,
    state text DEFAULT 'enabled',
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
    );
DROP TABLE IF EXISTS groups CASCADE;
CREATE TABLE IF NOT EXISTS groups (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name text NOT NULL,
    title text,
    org_id uuid NOT NULL REFERENCES organizations(id),
    metadata jsonb,
    state text DEFAULT 'enabled',
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz,
    UNIQUE(org_id, name)
    );
DROP TABLE IF EXISTS users CASCADE;
CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name text UNIQUE NOT NULL,
    title text,
    email text UNIQUE NOT NULL,
    metadata jsonb,
    state text DEFAULT 'enabled',
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
    );
DROP TABLE IF EXISTS serviceusers CASCADE;
CREATE TABLE IF NOT EXISTS serviceusers (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id uuid,
    title text,
    metadata jsonb,
    state text DEFAULT 'enabled',
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
    );
DROP TABLE IF EXISTS serviceuser_credentials CASCADE;
CREATE TABLE IF NOT EXISTS serviceuser_credentials (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    serviceuser_id uuid NOT NULL REFERENCES serviceusers(id),
    title text,
    secret_hash text,
    public_key jsonb,
    metadata jsonb,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
    );

DROP TABLE IF EXISTS actions CASCADE;
DROP TABLE IF EXISTS permissions CASCADE;
CREATE TABLE IF NOT EXISTS permissions(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name text NOT NULL,
    slug text NOT NULL UNIQUE,
    namespace_name text NOT NULL REFERENCES namespaces(name),
    metadata jsonb,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz,
    UNIQUE(namespace_name, name)
    );
DROP TABLE IF EXISTS roles CASCADE;
CREATE TABLE IF NOT EXISTS roles (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id uuid NOT NULL,
    name text NOT NULL,
    permissions jsonb,
    metadata jsonb,
    state text DEFAULT 'enabled',
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz,
    UNIQUE(org_id, name)
    );
DROP TABLE IF EXISTS relations CASCADE;
CREATE TABLE IF NOT EXISTS relations (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    subject_namespace_name text REFERENCES namespaces (name),
    subject_id text,
    subject_subrelation_name text,
    object_namespace_name text REFERENCES namespaces (name),
    object_id text,
    relation_name text,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz,
    UNIQUE (
               subject_namespace_name,
               subject_id,
               object_namespace_name,
               object_id,
               relation_name
           )
    );
DROP TABLE IF EXISTS resources CASCADE;
CREATE TABLE IF NOT EXISTS resources (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    urn text NOT NULL UNIQUE,
    name text,
    title text,
    principal_id uuid,
    principal_type text REFERENCES namespaces (name),
    project_id uuid REFERENCES projects (id),
    namespace_name text REFERENCES namespaces (name),
    metadata jsonb,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz
    );
DROP TABLE IF EXISTS policies CASCADE;
CREATE TABLE IF NOT EXISTS policies (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id uuid REFERENCES roles (id),
    resource_id uuid NOT NULL,
    resource_type text REFERENCES namespaces (name),
    principal_id uuid NOT NULL,
    principal_type text REFERENCES namespaces (name),
    metadata jsonb,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW(),
    deleted_at timestamptz,
    UNIQUE(role_id, resource_id, resource_type, principal_id, principal_type)
    );
DROP TABLE IF EXISTS flows CASCADE;
CREATE TABLE IF NOT EXISTS flows(
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    method text,
    email text,
    nonce text,
    metadata jsonb,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    expires_at timestamptz DEFAULT (NOW() + INTERVAL '7 days')
    );
DROP TABLE IF EXISTS sessions CASCADE;
CREATE TABLE IF NOT EXISTS sessions (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    authenticated_at timestamptz,
    metadata jsonb,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    expires_at timestamptz DEFAULT (NOW() + INTERVAL '7 days')
    );
DROP TABLE IF EXISTS metadata_keys CASCADE;
DROP TABLE IF EXISTS metaschema CASCADE;
CREATE TABLE IF NOT EXISTS metaschema (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    name text UNIQUE NOT NULL,
    schema text,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    updated_at timestamptz NOT NULL DEFAULT NOW()
    );
DROP TABLE IF EXISTS invitations CASCADE;
CREATE TABLE IF NOT EXISTS invitations (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id uuid NOT NULL REFERENCES organizations(id),
    user_id text NOT NULL,
    metadata jsonb,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    expires_at timestamptz DEFAULT (NOW() + INTERVAL '7 days')
);
DROP TABLE IF EXISTS auditlogs;
CREATE TABLE IF NOT EXISTS auditlogs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  org_id UUID,
  source TEXT NOT NULL,
  action TEXT NOT NULL,
  actor JSONB NOT NULL,
  target JSONB NOT NULL,
  metadata JSONB NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW()
);

-- create index
CREATE INDEX organizations_state_idx ON organizations(state);
CREATE INDEX projects_state_idx ON projects(state);
CREATE INDEX projects_org_id_idx ON projects(org_id);

CREATE INDEX groups_state_idx ON groups(state);
CREATE INDEX groups_org_id_idx ON groups(org_id);
CREATE INDEX users_state_idx ON users(state);
CREATE INDEX serviceusers_org_id_idx ON serviceusers(org_id);

CREATE INDEX roles_state_idx ON roles(state);
CREATE INDEX roles_org_id_idx ON roles(org_id);
CREATE INDEX permissions_slug_idx ON permissions(slug);

CREATE INDEX invitations_user_id_idx ON invitations(user_id);

CREATE INDEX resources_principal_id_idx ON resources(principal_id);
CREATE INDEX resources_principal_type_idx ON resources(principal_type);
CREATE INDEX resources_project_id_idx ON resources(project_id);

CREATE INDEX IF NOT EXISTS auditlogs_created_at_idx ON auditlogs (created_at);
CREATE INDEX IF NOT EXISTS auditlogs_source_idx ON auditlogs (source);
CREATE INDEX IF NOT EXISTS auditlogs_action_idx ON auditlogs (action);
CREATE INDEX IF NOT EXISTS auditlogs_org_id_idx ON auditlogs (org_id);
