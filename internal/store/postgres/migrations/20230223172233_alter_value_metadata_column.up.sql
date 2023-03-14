UPDATE metadata SET value = CONCAT('"',value,'"') where value NOT LIKE '[%' AND value NOT LIKE '{%';

ALTER TABLE metadata ALTER COLUMN value TYPE jsonb USING (value::jsonb);
