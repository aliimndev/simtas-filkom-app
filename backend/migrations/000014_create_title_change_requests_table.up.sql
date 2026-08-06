CREATE TABLE IF NOT EXISTS title_change_requests (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    thesis_id       UUID        NOT NULL REFERENCES theses(id) ON DELETE CASCADE,
    requested_by    UUID        NOT NULL REFERENCES users(id),
    previous_title  TEXT        NOT NULL,
    requested_title TEXT        NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'PENDING'
        CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED', 'CANCELLED')),
    reviewed_by     UUID        REFERENCES users(id),
    reviewed_at     TIMESTAMPTZ,
    review_notes    TEXT,
    cancelled_by    UUID        REFERENCES users(id),
    cancelled_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- One pending request per thesis
CREATE UNIQUE INDEX IF NOT EXISTS idx_title_change_requests_pending_thesis
    ON title_change_requests (thesis_id)
    WHERE status = 'PENDING';

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_title_change_requests_thesis_id
    ON title_change_requests (thesis_id);

CREATE INDEX IF NOT EXISTS idx_title_change_requests_requested_by
    ON title_change_requests (requested_by);

CREATE INDEX IF NOT EXISTS idx_title_change_requests_status
    ON title_change_requests (status);
