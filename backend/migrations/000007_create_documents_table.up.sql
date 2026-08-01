CREATE TABLE IF NOT EXISTS documents (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    thesis_id       UUID        NOT NULL REFERENCES theses(id) ON DELETE CASCADE,
    uploaded_by     UUID        NOT NULL REFERENCES users(id),
    document_type   VARCHAR(50) NOT NULL
        CHECK (document_type IN (
            'proposal', 'draft_chapter', 'seminar_doc',
            'defense_doc', 'final_thesis', 'revision_sheet',
            'endorsement_letter'
        )),
    chapter_number  INTEGER     CHECK (chapter_number BETWEEN 1 AND 5),
    version         INTEGER     NOT NULL DEFAULT 1,
    file_name       VARCHAR(255) NOT NULL,
    file_url        TEXT        NOT NULL,
    file_size       BIGINT,
    status          VARCHAR(30) NOT NULL DEFAULT 'pending_review'
        CHECK (status IN ('pending_review', 'approved', 'revision_required')),
    reviewer_id     UUID        REFERENCES users(id),
    reviewer_notes  TEXT,
    reviewed_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
