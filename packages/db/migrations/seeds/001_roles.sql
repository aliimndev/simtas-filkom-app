-- Seed: 5 roles for SIMTAS FILKOM
-- Safe to re-run (INSERT ... ON CONFLICT DO NOTHING)

INSERT INTO roles (id, name) VALUES
    (1, 'admin_fakultas'),
    (2, 'kaprodi'),
    (3, 'mahasiswa'),
    (4, 'dosen_pembimbing'),
    (5, 'dosen_penguji')
ON CONFLICT (name) DO NOTHING;

-- Keep the SERIAL sequence ahead of the explicit seed IDs so later inserts get
-- a fresh role ID instead of retrying the seeded value 1.
SELECT setval(
    pg_get_serial_sequence('roles', 'id'),
    COALESCE((SELECT MAX(id) FROM roles), 1),
    true
);
