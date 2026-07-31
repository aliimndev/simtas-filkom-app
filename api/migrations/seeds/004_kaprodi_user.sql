-- Seed: default Kaprodi account (for development/testing)
-- Password: Kaprodi@2027!  (bcrypt cost 12, Go $2a$ format)
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
    '$2a$12$OsM3GlIn7P.cRHl/hqE09eYFlntL6MVTEPEKJKgW03ZxHUIZ6i/NW',
    'Kepala Program Studi FILKOM',
    NULL,
    2,   -- kaprodi
    'Teknik Informatika',
    TRUE,
    TRUE
)
ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash;
