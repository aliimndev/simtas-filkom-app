CREATE TABLE IF NOT EXISTS thesis_defenses (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    thesis_id      UUID        NOT NULL REFERENCES theses(id) ON DELETE CASCADE,
    status         VARCHAR(30) NOT NULL DEFAULT 'pending'
        CHECK (status IN (
            'pending', 'scheduled', 'completed',
            'passed', 'failed', 'revision_required'
        )),
    scheduled_at   TIMESTAMPTZ,
    room           VARCHAR(100),
    revision_notes TEXT,
    final_score    DECIMAL(5,2),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS defense_examiners (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    defense_id  UUID        NOT NULL REFERENCES thesis_defenses(id) ON DELETE CASCADE,
    examiner_id UUID        NOT NULL REFERENCES users(id),
    assigned_by UUID        NOT NULL REFERENCES users(id),
    UNIQUE (defense_id, examiner_id)
);

CREATE TABLE IF NOT EXISTS defense_scores (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    defense_id       UUID         NOT NULL REFERENCES thesis_defenses(id) ON DELETE CASCADE,
    examiner_id      UUID         NOT NULL REFERENCES users(id),
    component_name   VARCHAR(100) NOT NULL,
    component_weight DECIMAL(5,2) NOT NULL CHECK (component_weight > 0 AND component_weight <= 100),
    score            DECIMAL(5,2) NOT NULL CHECK (score >= 0 AND score <= 100),
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (defense_id, examiner_id, component_name)
);
