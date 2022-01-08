CREATE TABLE IF NOT EXISTS relations
(
    id                   uuid PRIMARY KEY     DEFAULT uuid_generate_v4(),
    subject_namespace_id VARCHAR REFERENCES namespaces (id),
    subject_id           VARCHAR,
    object_namespace_id  VARCHAR REFERENCES namespaces (id),
    object_id            VARCHAR,
    role_id              VARCHAR REFERENCES roles (id),
    created_at           timestamptz NOT NULL DEFAULT NOW(),
    updated_at           timestamptz NOT NULL DEFAULT NOW(),
    deleted_at           timestamptz,
    UNIQUE (subject_namespace_id, subject_id, role_id, object_namespace_id, object_id)
);

CREATE TABLE IF NOT EXISTS resources
(
    id           varchar PRIMARY KEY,
    name         varchar,
    project_id   uuid REFERENCES projects (id),
    group_id     uuid REFERENCES groups (id),
    org_id       uuid REFERENCES organizations (id),
    namespace_id VARCHAR REFERENCES namespaces (id),
    created_at   timestamptz NOT NULL DEFAULT NOW(),
    updated_at   timestamptz NOT NULL DEFAULT NOW(),
    deleted_at   timestamptz
);