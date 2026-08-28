ALTER TABLE email_logs DROP CONSTRAINT IF EXISTS email_logs_status_check;
ALTER TABLE email_logs ADD CONSTRAINT email_logs_status_check
    CHECK (status IN ('sent', 'failed'));

ALTER TABLE email_logs DROP COLUMN IF EXISTS attempts;
ALTER TABLE email_logs DROP COLUMN IF EXISTS body;
