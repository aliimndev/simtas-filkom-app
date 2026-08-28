-- Seed: initial active academic year 2026/2027 Ganjil
-- Safe to re-run (INSERT ... ON CONFLICT DO NOTHING)

INSERT INTO academic_years (id, name, semester, start_date, end_date, is_active)
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    '2026/2027',
    'ganjil',
    '2026-09-01',
    '2027-01-31',
    TRUE
)
ON CONFLICT (id) DO NOTHING;
