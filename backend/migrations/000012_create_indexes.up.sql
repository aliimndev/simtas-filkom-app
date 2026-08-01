-- users
CREATE INDEX IF NOT EXISTS idx_users_role_id    ON users (role_id);
CREATE INDEX IF NOT EXISTS idx_users_is_active  ON users (is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);

-- theses
CREATE INDEX IF NOT EXISTS idx_theses_student_id       ON theses (student_id);
CREATE INDEX IF NOT EXISTS idx_theses_status           ON theses (status);
CREATE INDEX IF NOT EXISTS idx_theses_academic_year_id ON theses (academic_year_id);
CREATE INDEX IF NOT EXISTS idx_theses_deleted_at       ON theses (deleted_at);

-- documents
CREATE INDEX IF NOT EXISTS idx_documents_thesis_id    ON documents (thesis_id);
CREATE INDEX IF NOT EXISTS idx_documents_type_status  ON documents (document_type, status);

-- consultation_logs
CREATE INDEX IF NOT EXISTS idx_consultation_logs_thesis_id ON consultation_logs (thesis_id);
CREATE INDEX IF NOT EXISTS idx_consultation_logs_status    ON consultation_logs (status);

-- seminars
CREATE INDEX IF NOT EXISTS idx_seminars_thesis_id    ON seminars (thesis_id);
CREATE INDEX IF NOT EXISTS idx_seminars_status       ON seminars (status);
CREATE INDEX IF NOT EXISTS idx_seminars_scheduled_at ON seminars (scheduled_at);

-- thesis_defenses
CREATE INDEX IF NOT EXISTS idx_thesis_defenses_thesis_id    ON thesis_defenses (thesis_id);
CREATE INDEX IF NOT EXISTS idx_thesis_defenses_status       ON thesis_defenses (status);
CREATE INDEX IF NOT EXISTS idx_thesis_defenses_scheduled_at ON thesis_defenses (scheduled_at);

-- audit_logs
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id    ON audit_logs (user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_entity     ON audit_logs (entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs (created_at DESC);

-- token_blacklist
CREATE INDEX IF NOT EXISTS idx_token_blacklist_jti     ON token_blacklist (token_jti);
CREATE INDEX IF NOT EXISTS idx_token_blacklist_expires ON token_blacklist (expires_at);

-- password_reset_tokens
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens (user_id);
