-- One consent record per consent act, listing the documents it covers.
-- See docs/rfcs/0002-explicit-consent-at-signup.md, "Storage".
CREATE TABLE user_consents (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v7(),
    -- no FK to users(id): Delete is a hard DELETE, so CASCADE would drop these
    -- records with the account and RESTRICT would block deleting it at all.
    user_id       UUID        NOT NULL,
    user_email    TEXT        NOT NULL, -- denormalized so the record outlives the user row.
    documents     JSONB       NOT NULL, -- [{id, title, version, url}, ...], copied from config at write time.
    source        TEXT        NOT NULL DEFAULT 'signup',
    auth_strategy TEXT,
    ip_address    TEXT,                 -- TEXT and nullable, not INET: it comes from a request header.
    consented_at  TIMESTAMPTZ NOT NULL, -- when the user accepted, not when the row was written.
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT documents_not_empty CHECK (
        jsonb_typeof(documents) = 'array' AND jsonb_array_length(documents) > 0
    )
);

-- At most one signup consent per user: nothing repairs a record, so a second
-- write is a bug and should fail rather than leave two rows disagreeing.
CREATE UNIQUE INDEX uq_user_consents_signup
    ON user_consents(user_id) WHERE source = 'signup';

-- Following 20250904105226_add_audit_records_immutability.up.sql, which guards
-- UPDATE. DELETE is guarded too: a deleted record leaves a user who looks like
-- they never consented.
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
