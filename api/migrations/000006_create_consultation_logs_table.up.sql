CREATE TABLE IF NOT EXISTS consultation_logs (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    thesis_id         UUID        NOT NULL REFERENCES theses(id) ON DELETE CASCADE,
    created_by        UUID        NOT NULL REFERENCES users(id),
    consultation_date DATE        NOT NULL,
    topics_discussed  TEXT        NOT NULL,
    notes             TEXT,
    follow_up         TEXT,
    attachment_url    TEXT,
    status            VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'approved')),
    approved_by       UUID        REFERENCES users(id),
    approved_at       TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
