BEGIN;

DROP INDEX IF EXISTS idx_defenses_active_thesis;

DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM thesis_defenses WHERE status = 'cancelled') THEN
    RAISE EXCEPTION 'Cannot rollback Sidang cancellation while cancelled rows exist';
  END IF;
END $$;

ALTER TABLE thesis_defenses
    DROP CONSTRAINT IF EXISTS thesis_defenses_status_check;

ALTER TABLE thesis_defenses
    ADD CONSTRAINT thesis_defenses_status_check
    CHECK (status IN ('pending', 'scheduled', 'completed', 'passed', 'failed', 'revision_required'));

COMMIT;
