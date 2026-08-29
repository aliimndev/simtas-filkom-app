-- Prevent concurrent submissions from creating more than one active Seminar per Thesis.
-- Failed attempts remain historical and may be recreated when the lifecycle allows it.
CREATE UNIQUE INDEX IF NOT EXISTS idx_seminars_active_thesis
    ON seminars (thesis_id)
    WHERE status IN ('pending', 'scheduled', 'passed');
