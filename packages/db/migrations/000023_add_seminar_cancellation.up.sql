ALTER TABLE seminars
    ADD COLUMN IF NOT EXISTS cancellation_reason TEXT;

ALTER TABLE seminars
    DROP CONSTRAINT IF EXISTS seminars_status_check;

ALTER TABLE seminars
    ADD CONSTRAINT seminars_status_check
    CHECK (status IN ('pending', 'scheduled', 'completed', 'passed', 'failed', 'cancelled'));
