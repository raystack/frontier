ALTER TABLE users ADD COLUMN slug varchar UNIQUE NOT NULL;

-- TODO() initialise the slug values for pre-existing data
-- UPDATE users SET slug = CONCAT(LOWER(first_name), '-', LOWER(last_name), '-', EXTRACT(EPOCH FROM created_at)::int); (Discuss)