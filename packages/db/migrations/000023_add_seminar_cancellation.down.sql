DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM seminars WHERE status = 'cancelled') THEN
    RAISE EXCEPTION 'Cannot rollback Seminar cancellation while cancelled rows exist';
  END IF;
END $$;

ALTER TABLE seminars
    DROP CONSTRAINT IF EXISTS seminars_status_check;

ALTER TABLE seminars
    ADD CONSTRAINT seminars_status_check
    CHECK (status IN ('pending', 'scheduled', 'completed', 'passed', 'failed'));

ALTER TABLE seminars
    DROP COLUMN IF EXISTS cancellation_reason;
