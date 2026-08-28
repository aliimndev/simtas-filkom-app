-- 000018: align title_change_requests schema with entity.TitleChangeRequest.
--
-- 000014 named the actor columns requested_by/reviewed_by/cancelled_by, but
-- GORM derives column names from the entity fields (RequestedByID,
-- ReviewedByID, CancelledByID) and the rest of the schema uses the *_id
-- convention (role_id, student_id, thesis_id...). Without renaming, INSERT and
-- Preload failed with: column "requested_by_id" does not exist. Rename them.
--
-- The entity also declares gorm.DeletedAt but 000014 omitted the column, so
-- every GORM query appended "deleted_at IS NULL" and failed with: column
-- "deleted_at" does not exist. Add it + index.
ALTER TABLE title_change_requests RENAME COLUMN requested_by TO requested_by_id;
ALTER TABLE title_change_requests RENAME COLUMN reviewed_by TO reviewed_by_id;
ALTER TABLE title_change_requests RENAME COLUMN cancelled_by TO cancelled_by_id;
ALTER TABLE title_change_requests ADD COLUMN deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_title_change_requests_deleted_at
    ON title_change_requests (deleted_at);
