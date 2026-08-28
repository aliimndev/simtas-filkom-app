CREATE TABLE IF NOT EXISTS academic_years (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(20) NOT NULL,
    semester   VARCHAR(10) NOT NULL CHECK (semester IN ('ganjil', 'genap')),
    start_date DATE        NOT NULL,
    end_date   DATE        NOT NULL,
    is_active  BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_academic_year_dates CHECK (end_date > start_date)
);

-- Only one active academic year at a time
CREATE UNIQUE INDEX idx_academic_years_single_active
    ON academic_years (is_active)
    WHERE is_active = TRUE;
