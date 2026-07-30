-- Seed: default Admin Fakultas account
-- Password: Admin@2027!  (bcrypt cost 10)
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
    '$2b$10$0Vf5zdq0xhTIYT0a3UwP7urQ7QTWan5X.PshXMtk7xGr1UjFxJPKW',
    'Administrator FILKOM',
    NULL,
    1,   -- admin_fakultas
    NULL,
    TRUE,
    TRUE
)
ON CONFLICT (email) DO NOTHING;
