# Job 10 — Archive Module (Arsip Digital Skripsi)

**Phase:** 3 — Supporting Features
**Referensi PRD:** Section 6.7 (FR-ARCHIVE-001 s/d FR-ARCHIVE-003)
**Prerequisites:** Job 09 (Defense Module) ✅
**Estimasi:** 2 hari

---

## Objective

Implementasi modul arsip digital: upload skripsi final ke storage setelah yudisium, pencarian full-text arsip menggunakan PostgreSQL `tsvector`, dan kontrol akses download berdasarkan role. Setelah job ini selesai, setiap skripsi yang telah lulus dapat diarsipkan dan ditemukan melalui pencarian.

---

## Checklist

### Archive Repository & Use Case
- [x] Buat `backend/internal/domain/repository/archive_repository.go` — interface:
  ```go
  type ArchiveRepository interface {
    Create(ctx context.Context, archive *entity.ThesisArchive) error
    FindByThesisID(ctx context.Context, thesisID uuid.UUID) (*entity.ThesisArchive, error)
    FindByID(ctx context.Context, id uuid.UUID) (*entity.ThesisArchive, error)
    Search(ctx context.Context, filter ArchiveFilter) ([]*entity.ThesisArchive, int64, error)
    Update(ctx context.Context, archive *entity.ThesisArchive) error
  }
  ```
- [x] `ArchiveFilter`:
  ```go
  type ArchiveFilter struct {
    Query          string    // full-text search
    GraduationYear int
    FieldOfStudy   string
    StudyProgram   string
    SupervisorID   *uuid.UUID
    Page           int
    PerPage        int
  }
  ```
- [x] Buat `backend/internal/usecase/archive_usecase.go`

### Handler — Archive Endpoints

**POST `/api/v1/theses/:thesis_id/archive`** _(Mahasiswa pemilik + Admin)_
- [x] Content-Type: `multipart/form-data`
- [x] Form fields:
  - `file` — file PDF skripsi final (required)
  - `abstract_id` — abstrak Bahasa Indonesia (required)
  - `abstract_en` — abstrak Bahasa Inggris (optional)
  - `keywords` — kata kunci dipisah koma: `"machine learning,deep learning,computer vision"` (required)
  - `graduation_year` — tahun lulus (required)
- [x] Validasi:
  - Thesis harus berstatus `graduated`
  - Thesis belum punya arsip (unique per thesis)
  - File: PDF, max 25 MB (lebih besar dari dokumen biasa)
  - `abstract_id` minimal 50 kata
  - `graduation_year` harus tahun yang valid (tidak lebih dari tahun sekarang)
  - `keywords` minimal 3 kata kunci
- [x] Upload file ke storage — path: `archives/{graduation_year}/{thesis_id}/{filename}`
  - Gunakan stub storage untuk dev (implementasi penuh di Job 21)
- [x] Simpan ke `thesis_archives` — trigger PostgreSQL akan auto-update `search_vector`
- [x] Audit log: `ARCHIVE_CREATED`
- [x] Email notification ke mahasiswa: "Arsip skripsi Anda telah tersedia" (stub)
- [x] Response: `201 Created`

**GET `/api/v1/archives`** _(Semua authenticated user)_
- [x] Query params: `q` (full-text search), `year`, `field_of_study`, `study_program`, `supervisor_id`, `page` (default 1), `per_page` (default 20)
- [x] Jika `q` tidak kosong: gunakan PostgreSQL full-text search
  ```sql
  WHERE search_vector @@ plainto_tsquery('simple', $1)
  ORDER BY ts_rank(search_vector, plainto_tsquery('simple', $1)) DESC
  ```
- [x] Jika `q` kosong: return semua arsip dengan filter lain, order by `archived_at DESC`
- [x] Response:
  ```json
  {
    "success": true,
    "data": [
      {
        "id": "uuid",
        "thesis_id": "uuid",
        "title": "Judul Skripsi",
        "student": { "full_name": "Nama", "nim": "12345678" },
        "supervisors": [ { "full_name": "Dr. Ahmad" } ],
        "keywords": ["machine learning", "deep learning"],
        "graduation_year": 2027,
        "field_of_study": "Kecerdasan Buatan",
        "study_program": "Teknik Informatika",
        "archived_at": "2027-06-15T10:00:00Z"
      }
    ],
    "meta": { "page": 1, "per_page": 20, "total": 45 }
  }
  ```

**GET `/api/v1/archives/:id`** _(Semua authenticated user)_
- [x] Return detail lengkap arsip termasuk abstrak:
  ```json
  {
    "id": "uuid",
    "thesis_id": "uuid",
    "title": "...",
    "abstract_id": "...",
    "abstract_en": "...",
    "keywords": ["..."],
    "graduation_year": 2027,
    "student": { ... },
    "supervisors": [ { ... } ],
    "file_name": "skripsi_final.pdf",
    "archived_at": "..."
  }
  ```
- [x] **Tidak** menyertakan `file_url` langsung — download lewat endpoint terpisah

**GET `/api/v1/archives/:id/download`** _(Role-based access)_
- [x] Kontrol akses:
  - Mahasiswa: hanya bisa download arsip skripsinya sendiri
  - Dosen Pembimbing / Dosen Penguji / Kaprodi / Admin: bisa download semua arsip
- [x] Generate presigned URL dari storage (expired 30 menit)
- [x] Audit log: `ARCHIVE_DOWNLOADED` (catat user yang download + arsip yang didownload)
- [x] Response: `{ "success": true, "data": { "download_url": "...", "expires_in": 1800 } }`

**GET `/api/v1/theses/:thesis_id/archive`** _(Akses sama dengan detail arsip)_
- [x] Shortcut untuk get arsip berdasarkan thesis ID
- [x] Return sama dengan `GET /api/v1/archives/:id`

### Full-Text Search — Query Builder
- [x] Buat `backend/internal/repository/archive_repository_impl.go` dengan query:
  ```go
  func (r *archiveRepo) Search(ctx context.Context, filter ArchiveFilter) ([]*entity.ThesisArchive, int64, error) {
    query := r.db.WithContext(ctx).
      Joins("JOIN theses ON theses.id = thesis_archives.thesis_id").
      Joins("JOIN users ON users.id = theses.student_id").
      Preload("Thesis.Student").
      Preload("Thesis.Supervisors")

    if filter.Query != "" {
      query = query.Where(
        "thesis_archives.search_vector @@ plainto_tsquery('simple', ?)",
        filter.Query,
      ).Order("ts_rank(thesis_archives.search_vector, plainto_tsquery('simple', ?)) DESC", filter.Query)
    }
    if filter.GraduationYear > 0 {
      query = query.Where("thesis_archives.graduation_year = ?", filter.GraduationYear)
    }
    if filter.StudyProgram != "" {
      query = query.Where("users.study_program = ?", filter.StudyProgram)
    }
    // ... filter lainnya
  }
  ```

### Statistics Endpoint (digunakan Dashboard)

**GET `/api/v1/archives/stats`** _(Admin + Kaprodi)_
- [x] Return:
  ```json
  {
    "total_archives": 120,
    "by_year": [ { "year": 2027, "count": 45 }, { "year": 2026, "count": 38 } ],
    "by_field": [ { "field": "Kecerdasan Buatan", "count": 30 } ],
    "by_study_program": [ { "program": "Teknik Informatika", "count": 80 } ]
  }
  ```

---

## Done Criteria

- [x] `POST .../archive` pada thesis belum `graduated` → `422 Unprocessable Entity`
- [x] `POST .../archive` pada thesis `graduated` → arsip tersimpan, `search_vector` ter-update
- [x] `GET /api/v1/archives?q=machine+learning` → return arsip yang relevan, diurutkan by relevance
- [x] `GET /api/v1/archives?year=2027&study_program=Teknik+Informatika` → filter berfungsi
- [x] `GET /archives/:id/download` oleh mahasiswa pemilik → presigned URL berhasil
- [x] `GET /archives/:id/download` oleh mahasiswa lain → `403 Forbidden`
- [x] `GET /archives/:id/download` oleh dosen → berhasil (akses ke semua arsip)
- [x] Upload arsip kedua untuk thesis yang sama → `409 Conflict`
- [x] `GET /archives/stats` → statistik akurat
- [x] Audit log `ARCHIVE_DOWNLOADED` tercatat setiap kali download
