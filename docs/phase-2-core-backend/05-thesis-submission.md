# Job 05 — Thesis Submission (Pengajuan & Persetujuan Judul)

**Phase:** 2 — Core Backend
**Referensi PRD:** Section 6.2 (FR-THESIS-001 s/d FR-THESIS-004)
**Prerequisites:** Job 04 (User Management) ✅
**Estimasi:** 3 hari

---

## Objective

Implementasi alur pengajuan judul skripsi oleh mahasiswa dan proses review/approve/reject oleh Kaprodi, termasuk penunjukan dosen pembimbing setelah judul disetujui. Ini adalah titik awal alur akademik utama SIMTAS FILKOM.

---

## Checklist

### Thesis Repository & Use Case
- [x] Buat `backend/internal/domain/repository/thesis_repository.go` — interface:
  ```go
  type ThesisRepository interface {
    Create(ctx context.Context, thesis *entity.Thesis) error
    FindByID(ctx context.Context, id uuid.UUID) (*entity.Thesis, error)
    FindAll(ctx context.Context, filter ThesisFilter) ([]*entity.Thesis, int64, error)
    FindByStudentID(ctx context.Context, studentID uuid.UUID) (*entity.Thesis, error)
    UpdateStatus(ctx context.Context, id uuid.UUID, status string, notes string) error
    Update(ctx context.Context, thesis *entity.Thesis) error
    AssignSupervisor(ctx context.Context, thesisID, supervisorID, assignedBy uuid.UUID) error
    GetSupervisors(ctx context.Context, thesisID uuid.UUID) ([]*entity.User, error)
    FindActiveByStudentID(ctx context.Context, studentID uuid.UUID) (*entity.Thesis, error)
  }
  ```
- [x] `ThesisFilter` struct: `Status`, `AcademicYearID`, `StudyProgram`, `FieldOfStudy`, `SupervisorID`, `Search` (judul/nama/NIM), `Page`, `PerPage` (ditambah `StudentID` & `ExaminerID` untuk scope per role)
- [x] Buat `backend/internal/usecase/thesis_usecase.go`

### State Machine Thesis Status
Dokumentasikan dan implementasikan transisi status yang valid:

```
submitted
  ├─► approved    (oleh Kaprodi)
  └─► rejected    (oleh Kaprodi) → mahasiswa bisa submit ulang (status baru: submitted)

approved
  └─► in_progress (setelah pembimbing ditunjuk)

in_progress
  └─► seminar_ready (setelah dokumen seminar approved)

seminar_ready
  └─► seminar_done (setelah seminar selesai)

seminar_done
  └─► defense_ready (setelah dokumen sidang approved)

defense_ready
  └─► defense_done (setelah sidang selesai)

defense_done
  └─► graduated (setelah yudisium oleh Kaprodi)

* cancelled bisa dari status apapun (oleh Admin/Kaprodi)
```

- [x] Buat `backend/pkg/statemachine/thesis_state.go` — validasi transisi:
  ```go
  var ValidTransitions = map[string][]string{
    "submitted":     {"approved", "rejected", "cancelled"},
    "approved":      {"in_progress", "cancelled"},
    "in_progress":   {"seminar_ready", "cancelled"},
    "seminar_ready": {"seminar_done"},
    "seminar_done":  {"defense_ready"},
    "defense_ready": {"defense_done"},   // diperbaiki dari typo "defence_done"
    "defense_done":  {"graduated"},
  }

  func CanTransition(from, to string) bool
  ```
  > Catatan: typo `defence_done` pada dokumen diperbaiki menjadi `defense_done` agar konsisten dengan enum di database. `cancelled` bisa dari status apapun (di-handle di `CanTransition`), dan status terminal (`rejected`, `graduated`, `cancelled`) terdaftar sebagai key agar `ValidStatus` mengenalinya.

### Handler — Endpoint Mahasiswa

**POST `/api/v1/theses`** _(Mahasiswa only)_
- [x] Request body:
  ```json
  {
    "title": "Judul Skripsi Mahasiswa",
    "abstract": "Abstrak minimal 100 kata...",
    "field_of_study": "Kecerdasan Buatan",
    "thesis_type": "skripsi"
  }
  ```
- [x] Validasi:
  - Judul minimal 10 kata, maksimal 500 karakter
  - Abstrak minimal 100 kata
  - `thesis_type`: `skripsi` atau `tugas_akhir`
  - Mahasiswa tidak boleh punya thesis aktif (status bukan `cancelled` atau `graduated`)
- [x] Set `academic_year_id` dari tahun akademik yang sedang aktif (via `FindActive` baru di `AcademicYearRepository`)
- [x] Set status awal: `submitted`
- [x] Audit log: `THESIS_SUBMITTED`
- [x] Email notification ke Kaprodi (semua akun dengan role `kaprodi`, via `FindByRole` baru di `UserRepository`)
- [x] Response: `201 Created`

### Handler — Endpoint Kaprodi

**GET `/api/v1/theses`** _(Admin + Kaprodi: semua; Dosen Pembimbing: miliknya; Mahasiswa: miliknya)_
- [x] Query params: `status`, `academic_year_id`, `study_program`, `field_of_study`, `supervisor_id`, `search`, `page`, `per_page`
- [x] Scope data berdasarkan role (implementasi di use case layer):
  - Admin/Kaprodi → semua thesis
  - Dosen Pembimbing → thesis dengan `thesis_supervisors.supervisor_id = me`
  - Dosen Penguji → thesis yang dia ditugaskan sebagai penguji (via `defense_examiners` / `seminar_examiners`)
  - Mahasiswa → thesis miliknya sendiri

**GET `/api/v1/theses/:id`** _(semua role, scope sama)_
- [x] Return detail thesis lengkap (DTO `ThesisDetail`):
  ```json
  {
    "id": "uuid",
    "title": "...",
    "abstract": "...",
    "field_of_study": "...",
    "thesis_type": "skripsi",
    "status": "in_progress",
    "kaprodi_notes": "...",
    "student": { "id": "...", "full_name": "...", "nim": "..." },
    "supervisors": [ { "id": "...", "full_name": "..." } ],
    "academic_year": { "name": "2026/2027", "semester": "ganjil" },
    "submitted_at": "...",
    "approved_at": "..."
  }
  ```

**PUT `/api/v1/theses/:id/review`** _(Kaprodi only)_
- [x] Request body:
  ```json
  {
    "decision": "approved",
    "notes": "Judul disetujui, silakan mulai proses bimbingan"
  }
  ```
  atau
  ```json
  {
    "decision": "rejected",
    "notes": "Judul terlalu luas, harap dipersempit ke topik spesifik"
  }
  ```
- [x] Validasi: `decision` harus `approved` atau `rejected`
- [x] Validasi state machine: status harus `submitted` (selain itu → `422 Unprocessable Entity`)
- [x] Update status thesis sesuai decision
- [x] Set `approved_at` jika decision = approved
- [x] Audit log: `THESIS_APPROVED` atau `THESIS_REJECTED`
- [x] Email notification ke mahasiswa

**PUT `/api/v1/theses/:id/assign-supervisor`** _(Kaprodi only)_
- [x] Request body:
  ```json
  {
    "supervisor_ids": ["uuid-dosen-1", "uuid-dosen-2"]
  }
  ```
- [x] Validasi:
  - Status thesis harus `approved`
  - Minimal 1, maksimal 2 supervisor (duplikat ID dideduplikasi)
  - Setiap ID harus user dengan role `dosen_pembimbing` dan `is_active = true` (via `FindByRole`)
- [x] Insert ke `thesis_supervisors`
- [x] Update status thesis → `in_progress`
- [x] Audit log: `SUPERVISOR_ASSIGNED`
- [x] Email notification ke mahasiswa dan setiap dosen pembimbing yang ditunjuk

**GET `/api/v1/lecturers`** _(Kaprodi + Admin)_
- [x] Return list dosen pembimbing aktif dengan info beban bimbingan:
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "full_name": "Dr. Ahmad",
        "nidn": "123456",
        "active_supervision_count": 3
      }
    ]
  }
  ```
- [x] Urutkan dari beban paling rendah (memudahkan load balancing)

**PATCH `/api/v1/theses/:id/cancel`** _(Admin + Kaprodi)_
- [x] Set status → `cancelled`
- [x] Request body: `{ "reason": "..." }` (opsional)
- [x] Audit log: `THESIS_CANCELLED`

### Email Notification (gunakan email service dari Job 11 — stub dulu)
- [x] Saat status `submitted` → email ke semua Kaprodi
- [x] Saat status `approved` → email ke mahasiswa
- [x] Saat status `rejected` → email ke mahasiswa
- [x] Saat pembimbing ditunjuk → email ke mahasiswa + dosen pembimbing

> **Catatan:** Untuk job ini, implementasikan fungsi email notification sebagai stub (log ke console). Implementasi penuh email service ada di Job 11. Pastikan interface email service sudah didefinisikan agar mudah di-swap nanti.

### Interface Email Service
- [x] Interface email service — **diperluas di `backend/pkg/email/email_service.go`** (bukan `backend/internal/domain/service/` seperti di dokumen; lokasi ini sudah dipakai Job 04 sehingga direuse agar tidak ada dua interface):
  ```go
  type EmailService interface {
    SendThesisSubmitted(ctx context.Context, to []string, thesis *entity.Thesis) error
    SendThesisApproved(ctx context.Context, to string, thesis *entity.Thesis) error
    SendThesisRejected(ctx context.Context, to string, thesis *entity.Thesis, notes string) error
    SendSupervisorAssigned(ctx context.Context, studentEmail string, supervisorEmails []string, thesis *entity.Thesis) error
    // ... method lain ditambahkan di job berikutnya
  }
  ```
- [x] Buat `backend/pkg/email/stub_email_service.go` — implementasi yang hanya log ke console (untuk development) — 4 method thesis ditambahkan ke stub yang ada

---

## Done Criteria

- [x] `POST /api/v1/theses` oleh mahasiswa → thesis dibuat, status `submitted`
- [x] `POST /api/v1/theses` oleh mahasiswa yang sudah punya thesis aktif → `400 Bad Request`
- [x] `PUT /api/v1/theses/:id/review` dengan decision `approved` → status jadi `approved`
- [x] `PUT /api/v1/theses/:id/review` oleh mahasiswa → `403 Forbidden` (via `RequireRole`)
- [x] `PUT /api/v1/theses/:id/review` saat status bukan `submitted` → `422 Unprocessable Entity`
- [x] `PUT /api/v1/theses/:id/assign-supervisor` → status jadi `in_progress`, supervisors terdaftar
- [x] `PUT /api/v1/theses/:id/assign-supervisor` dengan ID dosen yang rolenya bukan `dosen_pembimbing` → `400`
- [x] `GET /api/v1/theses` oleh mahasiswa → hanya return thesis miliknya
- [x] `GET /api/v1/theses` oleh Kaprodi → return semua thesis
- [x] `GET /api/v1/lecturers` → return dosen diurutkan dari beban terendah
- [x] Email notification (stub) tercatat di console untuk setiap event
- [x] Semua action tercatat di `audit_logs`
