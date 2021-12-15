CREATE TABLE IF NOT EXISTS policies(
                                       id           uuid PRIMARY KEY     DEFAULT uuid_generate_v4(),
                                       role_id      VARCHAR REFERENCES roles (id),
                                       namespace_id VARCHAR REFERENCES namespaces (id),
                                       action_id    VARCHAR REFERENCES actions (id),
                                       created_at   timestamptz NOT NULL DEFAULT NOW(),
                                       updated_at   timestamptz NOT NULL DEFAULT NOW(),
                                       deleted_at   timestamptz,
                                       UNIQUE (role_id, namespace_id, action_id)
);