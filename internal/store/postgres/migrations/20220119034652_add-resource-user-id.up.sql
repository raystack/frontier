ALTER TABLE resources
    ADD COLUMN user_id uuid REFERENCES users (id),
    ADD CONSTRAINT either_user_or_group
        CHECK (user_id IS NOT NULL OR group_id IS NOT NULL);
