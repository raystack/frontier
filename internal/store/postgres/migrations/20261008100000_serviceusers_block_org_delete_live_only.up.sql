-- Service user rows are kept with deleted_at set, so only live ones block an org delete.
CREATE OR REPLACE FUNCTION serviceusers_block_org_delete()
    RETURNS trigger AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM serviceusers WHERE org_id = OLD.id AND deleted_at IS NULL) THEN
        RAISE EXCEPTION 'cannot delete organization %: service users still reference it', OLD.id
            USING ERRCODE = 'foreign_key_violation';
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
