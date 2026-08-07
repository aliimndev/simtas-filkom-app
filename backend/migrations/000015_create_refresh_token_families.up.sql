-- Refresh token rotation: one row per refresh-token family. The row stores
-- the currently-valid refresh token's JTI; on each refresh the JTI is rotated
-- atomically (CAS), so replaying a stale token is detected and the whole
-- family is revoked (token-theft mitigation).
CREATE TABLE IF NOT EXISTS refresh_token_families (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    family_id  UUID         NOT NULL,
    token_jti  VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ  NOT NULL,
    rotated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refresh_token_families_user  ON refresh_token_families (user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_token_families_family ON refresh_token_families (family_id);
CREATE INDEX IF NOT EXISTS idx_refresh_token_families_expires ON refresh_token_families (expires_at);
