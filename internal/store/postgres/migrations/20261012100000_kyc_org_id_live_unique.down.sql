ALTER TABLE organizations_kyc ADD CONSTRAINT organizations_kyc_org_id_key UNIQUE (org_id);
DROP INDEX IF EXISTS uq_organizations_kyc_org_id_live;
