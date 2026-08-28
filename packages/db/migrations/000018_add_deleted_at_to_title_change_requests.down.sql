DROP INDEX IF EXISTS idx_title_change_requests_deleted_at;
ALTER TABLE title_change_requests DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE title_change_requests RENAME COLUMN requested_by_id TO requested_by;
ALTER TABLE title_change_requests RENAME COLUMN reviewed_by_id TO reviewed_by;
ALTER TABLE title_change_requests RENAME COLUMN cancelled_by_id TO cancelled_by;
