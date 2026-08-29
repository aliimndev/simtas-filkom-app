CREATE TABLE IF NOT EXISTS surat_tugas (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    defense_id        UUID NOT NULL REFERENCES thesis_defenses(id) ON DELETE CASCADE,
    letter_number     VARCHAR(100) NOT NULL,
    issue_date        DATE NOT NULL,
    file_name         VARCHAR(255) NOT NULL,
    file_url          TEXT NOT NULL,
    issuer_id         UUID NOT NULL REFERENCES users(id),
    status            VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'issued', 'cancelled')),
    cancellation_reason TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_surat_tugas_active_defense
    ON surat_tugas (defense_id)
    WHERE status IN ('draft', 'issued');

CREATE UNIQUE INDEX IF NOT EXISTS idx_surat_tugas_letter_number
    ON surat_tugas (letter_number);
