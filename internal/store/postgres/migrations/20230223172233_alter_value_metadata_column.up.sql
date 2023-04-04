UPDATE metadata SET value = CONCAT('"',value,'"') where value NOT LIKE '[%' AND value NOT LIKE '{%' AND value NOT like '"%"' AND value NOT LIKE '%[^0-9.]%';

ALTER TABLE metadata ALTER COLUMN value TYPE jsonb USING (value::jsonb);
