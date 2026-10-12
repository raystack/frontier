CREATE UNIQUE INDEX IF NOT EXISTS uq_organizations_kyc_org_id_live ON organizations_kyc (org_id) WHERE deleted_at IS NULL;
ALTER TABLE organizations_kyc DROP CONSTRAINT IF EXISTS organizations_kyc_org_id_key;
