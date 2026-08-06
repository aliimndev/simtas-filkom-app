-- Email retry queue (durable): persist the rendered HTML body so a failed send
-- can be re-attempted after a process restart, track how many delivery cycles
-- have failed, and widen the status set to include 'queued' (enqueued, not yet
-- delivered).
ALTER TABLE email_logs ADD COLUMN IF NOT EXISTS body TEXT;
ALTER TABLE email_logs ADD COLUMN IF NOT EXISTS attempts INTEGER NOT NULL DEFAULT 0;

ALTER TABLE email_logs DROP CONSTRAINT IF EXISTS email_logs_status_check;
ALTER TABLE email_logs ADD CONSTRAINT email_logs_status_check
    CHECK (status IN ('queued', 'sent', 'failed'));
