CREATE TABLE
    IF NOT EXISTS organizations_kyc (
        org_id uuid NOT NULL UNIQUE REFERENCES organizations (id),
        status BOOLEAN DEFAULT FALSE,
        link text,
        created_at timestamptz NOT NULL DEFAULT NOW (),
        updated_at timestamptz NOT NULL DEFAULT NOW (),
        deleted_at timestamptz,
        CONSTRAINT link_non_empty_when_status_true CHECK (
            (status = FALSE)
                OR (
                status = TRUE
                    AND link IS NOT NULL
                    AND link <> ''
                )
            )
);

-- Creating the function to update the `updated_at` field
CREATE OR REPLACE FUNCTION trigger_set_org_kyc_timestamp() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_org_kyc_timestamp
    BEFORE UPDATE ON organizations_kyc
    FOR EACH ROW
EXECUTE PROCEDURE trigger_set_org_kyc_timestamp();
