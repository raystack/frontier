ALTER TABLE billing_customers
    ADD COLUMN payment_mode TEXT GENERATED ALWAYS AS (
        CASE WHEN credit_min IS NOT NULL AND credit_min < 0 THEN 'postpaid' ELSE 'prepaid' END
    ) STORED;
