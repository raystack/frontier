-- | Column name      | type     | Example                  |
-- |------------------|----------|--------------------------|
-- | id <PK>          | string   | org_manager              |
-- | name             | string   | Org Manager              |
-- | types            | []string | ["user" + "group"]       |
-- | namespace        | string   | organization             |
-- | metadata         | json     | {"resource": "firehose"} |

CREATE TABLE IF NOT EXISTS roles
(
    id                      varchar         PRIMARY KEY,
    name                    varchar         UNIQUE NOT NULL,
    types                   varchar[]       NOT NULL ,
    namespace               varchar         NOT NULL,
    metadata                jsonb,
    created_at              timestamptz     NOT NULL            DEFAULT NOW(),
    updated_at              timestamptz     NOT NULL            DEFAULT NOW(),
    deleted_at              timestamptz
);
