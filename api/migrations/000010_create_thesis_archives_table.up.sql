CREATE TABLE IF NOT EXISTS thesis_archives (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    thesis_id       UUID        NOT NULL UNIQUE REFERENCES theses(id),
    file_url        TEXT        NOT NULL,
    file_name       VARCHAR(255) NOT NULL,
    abstract_id     TEXT        NOT NULL,
    abstract_en     TEXT,
    keywords        TEXT[],
    graduation_year INTEGER     NOT NULL,
    archived_by     UUID        NOT NULL REFERENCES users(id),
    archived_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    search_vector   TSVECTOR,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- GIN index for full-text search
CREATE INDEX idx_archives_search ON thesis_archives USING GIN (search_vector);

-- Trigger function to auto-update search_vector
CREATE OR REPLACE FUNCTION update_archive_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('simple', coalesce(
            (SELECT title FROM theses WHERE id = NEW.thesis_id), ''
        )), 'A') ||
        setweight(to_tsvector('simple', coalesce(NEW.abstract_id, '')), 'B') ||
        setweight(to_tsvector('simple', coalesce(array_to_string(NEW.keywords, ' '), '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trig_update_archive_search_vector
    BEFORE INSERT OR UPDATE ON thesis_archives
    FOR EACH ROW EXECUTE FUNCTION update_archive_search_vector();
