CREATE TABLE IF NOT EXISTS theses (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id       UUID        NOT NULL REFERENCES users(id),
    academic_year_id UUID        NOT NULL REFERENCES academic_years(id),
    title            VARCHAR(500) NOT NULL,
    abstract         TEXT,
    field_of_study   VARCHAR(100),
    thesis_type      VARCHAR(20) NOT NULL CHECK (thesis_type IN ('skripsi', 'tugas_akhir')),
    status           VARCHAR(30) NOT NULL DEFAULT 'submitted'
        CHECK (status IN (
            'submitted', 'approved', 'rejected', 'in_progress',
            'seminar_ready', 'seminar_done', 'defense_ready',
            'defense_done', 'graduated', 'cancelled'
        )),
    kaprodi_notes    TEXT,
    submitted_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    approved_at      TIMESTAMPTZ,
    graduated_at     TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS thesis_supervisors (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    thesis_id     UUID        NOT NULL REFERENCES theses(id) ON DELETE CASCADE,
    supervisor_id UUID        NOT NULL REFERENCES users(id),
    assigned_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_by   UUID        NOT NULL REFERENCES users(id),
    UNIQUE (thesis_id, supervisor_id)
);
