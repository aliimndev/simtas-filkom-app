-- Add token_version for session invalidation (admin deactivate / password reset)
ALTER TABLE users ADD COLUMN IF NOT EXISTS token_version INTEGER NOT NULL DEFAULT 0;
