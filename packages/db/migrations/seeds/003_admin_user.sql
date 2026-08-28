-- Seed: default Admin Fakultas account
-- Password: Admin@2027!  (bcrypt cost 12, Go $2a$ format)
-- must_change_password = true  →  admin MUST change on first login
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
    'b0000000-0000-0000-0000-000000000001',
    'admin@filkom.unida.ac.id',
    '$2a$12$MZwlGk5OvHhPQbALJjmHu.hxkgv8TEqDL71ln5ONPO7vY5xILnCsC',
    'Administrator FILKOM',
    NULL,
    1,   -- admin_fakultas
    NULL,
    TRUE,
    TRUE
)
ON CONFLICT (email) DO UPDATE SET
    password_hash = EXCLUDED.password_hash;
