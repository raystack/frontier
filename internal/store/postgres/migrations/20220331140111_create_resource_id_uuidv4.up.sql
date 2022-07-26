ALTER TABLE resources
    ADD COLUMN id uuid DEFAULT uuid_generate_v4();

ALTER TABLE resources
    ADD PRIMARY KEY (id);
