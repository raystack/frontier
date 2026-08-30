-- One consent record per consent, listing the documents it covers.
-- See docs/rfcs/0002-explicit-consent-at-signup.md, "Storage".
--
-- There is deliberately no foreign key to users(id). UserRepository.Delete does a hard
-- DELETE, so ON DELETE CASCADE would drop these records along with the account and
-- ON DELETE RESTRICT would block account deletion outright. The records have to outlive
-- the user, which is also why user_email is denormalized onto the row.
CREATE TABLE user_consents (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    user_id       UUID        NOT NULL, -- no FK to users(id): see above.
    user_email    TEXT        NOT NULL, -- denormalized so the record survives the user row.
    documents     JSONB       NOT NULL, -- [{id, title, version, url}, ...], copied from config at write time.
    source        TEXT        NOT NULL DEFAULT 'signup',
    auth_strategy TEXT,                 -- the flow's own strategy_name: oidc, mailotp or passkey.
    ip_address    TEXT,                 -- TEXT and nullable, not INET: it comes from a request header.
    consented_at  TIMESTAMPTZ NOT NULL, -- when the user accepted, not when the row was written.
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT documents_not_empty CHECK (
        jsonb_typeof(documents) = 'array' AND jsonb_array_length(documents) > 0
    )
);

-- At most one signup consent per user. Nothing repairs a record, so a second signup write
-- is a bug; this makes it fail rather than leave two rows disagreeing.
CREATE UNIQUE INDEX uq_user_consents_signup
    ON user_consents(user_id) WHERE source = 'signup';

-- Immutability, following 20250904105226_add_audit_records_immutability.up.sql. That
-- migration guards UPDATE only; DELETE is guarded here as well, because a deleted record
-- leaves a user who looks like they never consented. DROP TABLE is unaffected by a row
-- level trigger, so the down migration still works.
CREATE OR REPLACE FUNCTION prevent_user_consent_updates()
  RETURNS TRIGGER AS $$
BEGIN
      RAISE EXCEPTION 'user_consents cannot be updated to maintain consent integrity'
          USING ERRCODE = '45000',  -- User-defined error (Postgres convention: user-defined error codes are in the 45000-45999 range)
              DETAIL = 'Consent records are immutable once created';
END;
  $$ LANGUAGE plpgsql;

CREATE TRIGGER trg_user_consents_prevent_update
    BEFORE UPDATE ON user_consents
    FOR EACH ROW EXECUTE FUNCTION prevent_user_consent_updates();

COMMENT ON TRIGGER trg_user_consents_prevent_update ON user_consents IS
    'Enforces immutability of consent records by preventing any UPDATE operation.';

CREATE OR REPLACE FUNCTION prevent_user_consent_deletes()
  RETURNS TRIGGER AS $$
BEGIN
      RAISE EXCEPTION 'user_consents cannot be deleted to maintain consent integrity'
          USING ERRCODE = '45000',  -- User-defined error (Postgres convention: user-defined error codes are in the 45000-45999 range)
              DETAIL = 'Consent records must outlive the user they describe';
END;
  $$ LANGUAGE plpgsql;

CREATE TRIGGER trg_user_consents_prevent_delete
    BEFORE DELETE ON user_consents
    FOR EACH ROW EXECUTE FUNCTION prevent_user_consent_deletes();

COMMENT ON TRIGGER trg_user_consents_prevent_delete ON user_consents IS
    'Enforces immutability of consent records by preventing any DELETE operation.';
