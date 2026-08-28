-- Seed: 5 roles for SIMTAS FILKOM
-- Safe to re-run (INSERT ... ON CONFLICT DO NOTHING)

INSERT INTO roles (name) VALUES
    ('admin_fakultas'),
    ('kaprodi'),
    ('mahasiswa'),
    ('dosen_pembimbing'),
    ('dosen_penguji')
ON CONFLICT (name) DO NOTHING;
