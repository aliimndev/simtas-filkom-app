# Job 02 — Database Schema & Migrasi

**Phase:** 1 — Foundation
**Referensi PRD:** Section 12 (Database Overview), Section 6 (semua Functional Requirements — entity definition)
**Prerequisites:** Job 01 (Project Setup) ✅
**Estimasi:** 3–4 hari

---

## Objective

Merancang dan mengimplementasikan seluruh skema database PostgreSQL menggunakan file migrasi yang terversi. Setelah job ini selesai, semua tabel tersedia di database dengan relasi, constraint, index yang benar, serta seed data awal tersedia untuk development.

---

## Checklist

### Setup Migration Tool
- [ ] Install `golang-migrate`:
  ```bash
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```
- [ ] Buat folder `backend/migrations/`
- [ ] Konvensi penamaan: `000001_create_roles_table.up.sql` / `000001_create_roles_table.down.sql`
- [ ] Buat helper `pkg/database/migrate.go` untuk jalankan migrasi saat startup (via flag `--migrate`)

### Migrasi — Grup 1: Master Data

#### `000001_create_roles_table`
- [ ] Tabel **`roles`**:
  ```sql
  id          SERIAL PRIMARY KEY,
  name        VARCHAR(50) UNIQUE NOT NULL,
              -- nilai: admin_fakultas, kaprodi, mahasiswa,
              --        dosen_pembimbing, dosen_penguji
  created_at  TIMESTAMPTZ DEFAULT NOW()
  ```

#### `000002_create_academic_years_table`
- [ ] Tabel **`academic_years`**:
  ```sql
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        VARCHAR(20) NOT NULL,   -- contoh: "2026/2027"
  semester    VARCHAR(10) NOT NULL,   -- ganjil / genap
  start_date  DATE NOT NULL,
  end_date    DATE NOT NULL,
  is_active   BOOLEAN DEFAULT FALSE,
  created_at  TIMESTAMPTZ DEFAULT NOW(),
  updated_at  TIMESTAMPTZ DEFAULT NOW()
  ```
  - Constraint: hanya 1 baris `is_active = true` pada satu waktu (partial unique index)

### Migrasi — Grup 2: Users

#### `000003_create_users_table`
- [ ] Tabel **`users`**:
  ```sql
  id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email                VARCHAR(255) UNIQUE NOT NULL,
  password_hash        VARCHAR(255) NOT NULL,
  full_name            VARCHAR(255) NOT NULL,
  nim_nidn             VARCHAR(50),
  role_id              INTEGER NOT NULL REFERENCES roles(id),
  study_program        VARCHAR(100),
  profile_photo_url    TEXT,
  is_active            BOOLEAN DEFAULT TRUE,
  must_change_password BOOLEAN DEFAULT TRUE,
  login_attempt_count  INTEGER DEFAULT 0,
  locked_until         TIMESTAMPTZ,
  last_login_at        TIMESTAMPTZ,
  created_at           TIMESTAMPTZ DEFAULT NOW(),
  updated_at           TIMESTAMPTZ DEFAULT NOW(),
  deleted_at           TIMESTAMPTZ  -- soft delete
  ```

#### `000004_create_auth_tables`
- [ ] Tabel **`password_reset_tokens`**:
  ```sql
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token       VARCHAR(255) UNIQUE NOT NULL,
  expires_at  TIMESTAMPTZ NOT NULL,
  used_at     TIMESTAMPTZ,
  created_at  TIMESTAMPTZ DEFAULT NOW()
  ```
- [ ] Tabel **`token_blacklist`**:
  ```sql
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  token_jti   VARCHAR(255) UNIQUE NOT NULL,
  expires_at  TIMESTAMPTZ NOT NULL,
  created_at  TIMESTAMPTZ DEFAULT NOW()
  ```
  - Buat job/cron untuk hapus expired entries secara berkala

### Migrasi — Grup 3: Thesis Core

#### `000005_create_theses_table`
- [ ] Tabel **`theses`**:
  ```sql
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  student_id       UUID NOT NULL REFERENCES users(id),
  academic_year_id UUID NOT NULL REFERENCES academic_years(id),
  title            VARCHAR(500) NOT NULL,
  abstract         TEXT,
  field_of_study   VARCHAR(100),
  thesis_type      VARCHAR(20) NOT NULL,  -- skripsi / tugas_akhir
  status           VARCHAR(30) NOT NULL DEFAULT 'submitted',
                   -- submitted, approved, rejected, in_progress,
                   -- seminar_ready, seminar_done, defense_ready,
                   -- defense_done, graduated, cancelled
  kaprodi_notes    TEXT,
  submitted_at     TIMESTAMPTZ DEFAULT NOW(),
  approved_at      TIMESTAMPTZ,
  graduated_at     TIMESTAMPTZ,
  created_at       TIMESTAMPTZ DEFAULT NOW(),
  updated_at       TIMESTAMPTZ DEFAULT NOW(),
  deleted_at       TIMESTAMPTZ
  ```
  - Constraint: `CHECK (status IN ('submitted','approved','rejected','in_progress','seminar_ready','seminar_done','defense_ready','defense_done','graduated','cancelled'))`

#### `000006_create_thesis_supervisors_table`
- [ ] Tabel **`thesis_supervisors`**:
  ```sql
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  thesis_id     UUID NOT NULL REFERENCES theses(id) ON DELETE CASCADE,
  supervisor_id UUID NOT NULL REFERENCES users(id),
  assigned_at   TIMESTAMPTZ DEFAULT NOW(),
  assigned_by   UUID NOT NULL REFERENCES users(id),
  UNIQUE(thesis_id, supervisor_id)
  ```

### Migrasi — Grup 4: Consultation

#### `000007_create_consultation_logs_table`
- [ ] Tabel **`consultation_logs`**:
  ```sql
  id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  thesis_id           UUID NOT NULL REFERENCES theses(id) ON DELETE CASCADE,
  created_by          UUID NOT NULL REFERENCES users(id),
  consultation_date   DATE NOT NULL,
  topics_discussed    TEXT NOT NULL,
  notes               TEXT,
  follow_up           TEXT,
  attachment_url      TEXT,
  status              VARCHAR(20) DEFAULT 'pending',  -- pending / approved
  approved_by         UUID REFERENCES users(id),
  approved_at         TIMESTAMPTZ,
  created_at          TIMESTAMPTZ DEFAULT NOW(),
  updated_at          TIMESTAMPTZ DEFAULT NOW()
  ```

### Migrasi — Grup 5: Documents

#### `000008_create_documents_table`
- [ ] Tabel **`documents`**:
  ```sql
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  thesis_id       UUID NOT NULL REFERENCES theses(id) ON DELETE CASCADE,
  uploaded_by     UUID NOT NULL REFERENCES users(id),
  document_type   VARCHAR(50) NOT NULL,
                  -- proposal, draft_chapter, seminar_doc,
                  -- defense_doc, final_thesis, revision_sheet,
                  -- endorsement_letter
  chapter_number  INTEGER,  -- hanya untuk draft_chapter (1-5)
  version         INTEGER NOT NULL DEFAULT 1,
  file_name       VARCHAR(255) NOT NULL,
  file_url        TEXT NOT NULL,
  file_size       BIGINT,  -- bytes
  status          VARCHAR(30) DEFAULT 'pending_review',
                  -- pending_review, approved, revision_required
  reviewer_id     UUID REFERENCES users(id),
  reviewer_notes  TEXT,
  reviewed_at     TIMESTAMPTZ,
  created_at      TIMESTAMPTZ DEFAULT NOW(),
  updated_at      TIMESTAMPTZ DEFAULT NOW()
  ```

### Migrasi — Grup 6: Seminar

#### `000009_create_seminars_table`
- [ ] Tabel **`seminars`**:
  ```sql
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  thesis_id    UUID NOT NULL REFERENCES theses(id) ON DELETE CASCADE,
  status       VARCHAR(20) DEFAULT 'pending',
               -- pending, scheduled, completed, passed, failed
  scheduled_at TIMESTAMPTZ,
  room         VARCHAR(100),
  notes        TEXT,
  final_score  DECIMAL(5,2),  -- dihitung otomatis
  created_at   TIMESTAMPTZ DEFAULT NOW(),
  updated_at   TIMESTAMPTZ DEFAULT NOW()
  ```
- [ ] Tabel **`seminar_examiners`**:
  ```sql
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  seminar_id   UUID NOT NULL REFERENCES seminars(id) ON DELETE CASCADE,
  examiner_id  UUID NOT NULL REFERENCES users(id),
  assigned_by  UUID NOT NULL REFERENCES users(id),
  UNIQUE(seminar_id, examiner_id)
  ```
- [ ] Tabel **`seminar_scores`**:
  ```sql
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  seminar_id       UUID NOT NULL REFERENCES seminars(id) ON DELETE CASCADE,
  examiner_id      UUID NOT NULL REFERENCES users(id),
  component_name   VARCHAR(100) NOT NULL,
  component_weight DECIMAL(5,2) NOT NULL,  -- persentase bobot
  score            DECIMAL(5,2) NOT NULL,  -- 0-100
  created_at       TIMESTAMPTZ DEFAULT NOW(),
  updated_at       TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(seminar_id, examiner_id, component_name)
  ```

### Migrasi — Grup 7: Thesis Defense

#### `000010_create_thesis_defenses_table`
- [ ] Tabel **`thesis_defenses`**:
  ```sql
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  thesis_id      UUID NOT NULL REFERENCES theses(id) ON DELETE CASCADE,
  status         VARCHAR(30) DEFAULT 'pending',
                 -- pending, scheduled, completed, passed,
                 -- failed, revision_required
  scheduled_at   TIMESTAMPTZ,
  room           VARCHAR(100),
  revision_notes TEXT,
  final_score    DECIMAL(5,2),
  created_at     TIMESTAMPTZ DEFAULT NOW(),
  updated_at     TIMESTAMPTZ DEFAULT NOW()
  ```
- [ ] Tabel **`defense_examiners`** (struktur sama dengan `seminar_examiners`)
- [ ] Tabel **`defense_scores`** (struktur sama dengan `seminar_scores`)

### Migrasi — Grup 8: Archive

#### `000011_create_thesis_archives_table`
- [ ] Tabel **`thesis_archives`**:
  ```sql
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  thesis_id       UUID NOT NULL UNIQUE REFERENCES theses(id),
  file_url        TEXT NOT NULL,
  file_name       VARCHAR(255) NOT NULL,
  abstract_id     TEXT NOT NULL,   -- abstrak Bahasa Indonesia
  abstract_en     TEXT,            -- abstrak Bahasa Inggris
  keywords        TEXT[],          -- array kata kunci
  graduation_year INTEGER NOT NULL,
  archived_by     UUID NOT NULL REFERENCES users(id),
  archived_at     TIMESTAMPTZ DEFAULT NOW(),
  search_vector   TSVECTOR,        -- untuk full-text search
  created_at      TIMESTAMPTZ DEFAULT NOW(),
  updated_at      TIMESTAMPTZ DEFAULT NOW()
  ```
- [ ] Buat trigger PostgreSQL auto-update `search_vector`:
  ```sql
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
  ```

### Migrasi — Grup 9: System

#### `000012_create_system_tables`
- [ ] Tabel **`audit_logs`**:
  ```sql
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID REFERENCES users(id) ON DELETE SET NULL,
  action       VARCHAR(100) NOT NULL,  -- contoh: USER_CREATED, THESIS_APPROVED
  entity_type  VARCHAR(50),            -- contoh: user, thesis, document
  entity_id    UUID,
  old_value    JSONB,
  new_value    JSONB,
  ip_address   INET,
  user_agent   TEXT,
  created_at   TIMESTAMPTZ DEFAULT NOW()
  ```
- [ ] Tabel **`email_logs`**:
  ```sql
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  recipient_email  VARCHAR(255) NOT NULL,
  event_type       VARCHAR(100) NOT NULL,
  subject          VARCHAR(500),
  status           VARCHAR(20) DEFAULT 'sent',  -- sent / failed
  provider         VARCHAR(50) DEFAULT 'resend',
  error_message    TEXT,
  created_at       TIMESTAMPTZ DEFAULT NOW()
  ```

### Index
- [ ] `users`: `CREATE INDEX idx_users_email ON users(email)` (sudah di-cover UNIQUE)
- [ ] `users`: `CREATE INDEX idx_users_role_id ON users(role_id)`
- [ ] `users`: `CREATE INDEX idx_users_is_active ON users(is_active) WHERE deleted_at IS NULL`
- [ ] `theses`: `CREATE INDEX idx_theses_student_id ON theses(student_id)`
- [ ] `theses`: `CREATE INDEX idx_theses_status ON theses(status)`
- [ ] `theses`: `CREATE INDEX idx_theses_academic_year ON theses(academic_year_id)`
- [ ] `documents`: `CREATE INDEX idx_documents_thesis_id ON documents(thesis_id)`
- [ ] `documents`: `CREATE INDEX idx_documents_type_status ON documents(document_type, status)`
- [ ] `audit_logs`: `CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id)`
- [ ] `audit_logs`: `CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id)`
- [ ] `audit_logs`: `CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC)`
- [ ] `thesis_archives`: `CREATE INDEX idx_archives_search ON thesis_archives USING GIN(search_vector)`
- [ ] `token_blacklist`: `CREATE INDEX idx_token_blacklist_jti ON token_blacklist(token_jti)`
- [ ] `token_blacklist`: `CREATE INDEX idx_token_blacklist_expires ON token_blacklist(expires_at)`

### GORM Models (Go Structs)
- [ ] Buat file Go struct di `backend/internal/domain/entity/` untuk setiap entitas
- [ ] Contoh konvensi:
  ```go
  type User struct {
    ID                  uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    Email               string         `gorm:"uniqueIndex;not null"`
    PasswordHash        string         `gorm:"not null"`
    FullName            string         `gorm:"not null"`
    NimNidn             string
    RoleID              int
    Role                Role           `gorm:"foreignKey:RoleID"`
    IsActive            bool           `gorm:"default:true"`
    MustChangePassword  bool           `gorm:"default:true"`
    CreatedAt           time.Time
    UpdatedAt           time.Time
    DeletedAt           gorm.DeletedAt `gorm:"index"`
  }
  ```
- [ ] Semua model menggunakan UUID sebagai primary key
- [ ] Soft delete via `gorm.DeletedAt` pada entitas utama

### Seed Data
- [ ] File `backend/migrations/seeds/001_roles.sql` — insert 5 role
- [ ] File `backend/migrations/seeds/002_academic_year.sql` — 1 tahun akademik aktif
- [ ] File `backend/migrations/seeds/003_admin_user.sql` — akun admin default:
  - Email: `admin@filkom.unida.ac.id`
  - Password: `Admin@2027!` (bcrypt hash)
  - `must_change_password: true`
- [ ] File `backend/migrations/seeds/004_kaprodi_user.sql` — 1 akun Kaprodi untuk testing

---

## Done Criteria

- [ ] `migrate -path ./migrations -database "$DB_URL" up` berhasil tanpa error
- [ ] `migrate -path ./migrations -database "$DB_URL" down` rollback semua tabel tanpa error
- [ ] Semua FK constraint berfungsi — insert data dengan FK invalid → error
- [ ] Partial unique index `is_active` pada `academic_years` berfungsi
- [ ] Full-text search trigger berfungsi pada `thesis_archives`
- [ ] GIN index `search_vector` terbuat
- [ ] Seed data tersedia: 5 roles, 1 academic year, akun admin, akun kaprodi
- [ ] Semua GORM model dapat di-compile dan auto-migrate tanpa error
