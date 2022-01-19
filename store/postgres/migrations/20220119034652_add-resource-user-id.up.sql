ALTER TABLE resources
    ADD COLUMN user_id uuid REFERENCES users (id);
