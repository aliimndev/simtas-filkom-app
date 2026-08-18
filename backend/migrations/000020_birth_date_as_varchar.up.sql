-- birth_date dibuat VARCHAR(10) agar round-trip "YYYY-MM-DD" tetap utuh
-- untuk <input type="date"> di frontend (hindari parsed sebagai TIMESTAMPTZ).
ALTER TABLE users
  ALTER COLUMN birth_date TYPE VARCHAR(10) USING TO_CHAR(birth_date, 'YYYY-MM-DD');
