DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'payment_mode_type') THEN
        CREATE TYPE payment_mode_type AS ENUM ('prepaid', 'postpaid');
    END IF;
END $$;

ALTER TABLE billing_customers
    ADD COLUMN payment_mode payment_mode_type GENERATED ALWAYS AS (
        CASE WHEN credit_min IS NOT NULL AND credit_min < 0 THEN 'postpaid'::payment_mode_type ELSE 'prepaid'::payment_mode_type END
    ) STORED;
