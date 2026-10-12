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
