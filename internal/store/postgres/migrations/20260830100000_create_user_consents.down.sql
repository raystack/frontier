DROP TRIGGER IF EXISTS trg_user_consents_prevent_delete ON user_consents;
DROP TRIGGER IF EXISTS trg_user_consents_prevent_update ON user_consents;

DROP FUNCTION IF EXISTS prevent_user_consent_deletes();
DROP FUNCTION IF EXISTS prevent_user_consent_updates();

DROP INDEX IF EXISTS uq_user_consents_signup;

-- a BEFORE DELETE trigger fires per row and does not block DROP TABLE.
DROP TABLE IF EXISTS user_consents;
