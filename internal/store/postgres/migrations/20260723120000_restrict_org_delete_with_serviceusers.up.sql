-- Block deleting an organization that still has service users.

-- A plain foreign key to organizations(id) cannot be used here. The bootstrap
-- superuser lives under the virtual platform org (00000000-0000-0000-0000-000000000000),
-- which has no row in organizations by design. A BEFORE DELETE trigger only
-- fires when a real org is deleted, so the virtual platform org and its
-- bootstrap service account are never in scope.
--
-- Existing orphaned rows (if any) are left untouched; they are cleaned up out
-- of band, not by this migration.
CREATE OR REPLACE FUNCTION serviceusers_block_org_delete()
    RETURNS trigger AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM serviceusers WHERE org_id = OLD.id) THEN
        RAISE EXCEPTION 'cannot delete organization %: service users still reference it', OLD.id
            USING ERRCODE = 'foreign_key_violation';
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_serviceusers_block_org_delete ON organizations;
CREATE TRIGGER trg_serviceusers_block_org_delete
    BEFORE DELETE ON organizations
    FOR EACH ROW
    EXECUTE FUNCTION serviceusers_block_org_delete();
