DROP TRIGGER IF EXISTS trig_update_archive_search_vector ON thesis_archives;
DROP FUNCTION IF EXISTS update_archive_search_vector();
DROP INDEX IF EXISTS idx_archives_search;
DROP TABLE IF EXISTS thesis_archives;
