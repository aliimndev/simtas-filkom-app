-- Seed: default Kaprodi account (for development/testing)
-- Password: Kaprodi@2027!  (bcrypt cost 10)
-- must_change_password = true  →  kaprodi MUST change on first login
-- Safe to re-run (INSERT ... ON CONFLICT DO NOTHING)

INSERT INTO users (
    id,
    email,
    password_hash,
    full_name,
    nim_nidn,
    role_id,
    study_program,
    is_active,
    must_change_password
)
VALUES (
    'b0000000-0000-0000-0000-000000000002',
    'kaprodi@filkom.unida.ac.id',
    '$2b$10$8NtAMfP3qq8wkN2Gdi0phOp4XcMgYgjk.Z2l45pWVG6/n55AHoRLq',
    'Kepala Program Studi FILKOM',
    NULL,
    2,   -- kaprodi
    'Teknik Informatika',
    TRUE,
    TRUE
)
ON CONFLICT (email) DO NOTHING;
