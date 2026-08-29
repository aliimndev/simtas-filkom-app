ALTER TABLE thesis_defenses
    DROP CONSTRAINT IF EXISTS thesis_defenses_status_check;

ALTER TABLE thesis_defenses
    ADD CONSTRAINT thesis_defenses_status_check
    CHECK (status IN ('pending', 'scheduled', 'completed', 'passed', 'failed', 'revision_required', 'cancelled'));

CREATE UNIQUE INDEX IF NOT EXISTS idx_defenses_active_thesis
    ON thesis_defenses (thesis_id)
    WHERE status IN ('pending', 'scheduled');
