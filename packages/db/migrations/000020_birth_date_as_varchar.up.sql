-- birth_date dibuat VARCHAR(10) agar round-trip "YYYY-MM-DD" tetap utuh
-- untuk <input type="date"> di frontend (hindari parsed sebagai TIMESTAMPTZ).
-- Idempoten: hanya konversi bila kolom belum berupa VARCHAR (mis. masih
-- bertipe date/timestamp). Migrasi 000019 sudah membuatnya VARCHAR(10),
-- sehingga TO_CHAR di sini tidak valid dan dilewati.
DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_name = 'users'
      AND column_name = 'birth_date'
      AND data_type <> 'character varying'
  ) THEN
    ALTER TABLE users
      ALTER COLUMN birth_date TYPE VARCHAR(10) USING TO_CHAR(birth_date, 'YYYY-MM-DD');
  END IF;
END $$;
