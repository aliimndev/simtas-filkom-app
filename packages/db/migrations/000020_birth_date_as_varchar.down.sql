ALTER TABLE users
  ALTER COLUMN birth_date TYPE DATE USING birth_date::DATE;
