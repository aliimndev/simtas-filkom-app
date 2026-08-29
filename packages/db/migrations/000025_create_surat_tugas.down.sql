BEGIN;

DROP INDEX IF EXISTS idx_surat_tugas_letter_number;
DROP INDEX IF EXISTS idx_surat_tugas_active_defense;
DROP TABLE IF EXISTS surat_tugas;

COMMIT;
