ALTER TABLE billing_features
    ADD COLUMN behavior text NOT NULL DEFAULT 'basic';
