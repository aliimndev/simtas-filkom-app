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
- [ ] Buat `internal/domain/repository/thesis_repository.go` — interface:
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
- [ ] `ThesisFilter` struct: `Status`, `AcademicYearID`, `StudyProgram`, `FieldOfStudy`, `SupervisorID`, `Search` (judul/nama/NIM), `Page`, `PerPage`
- [ ] Buat `internal/usecase/thesis_usecase.go`

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

- [ ] Buat `pkg/statemachine/thesis_state.go` — validasi transisi:
  ```go
  var ValidTransitions = map[string][]string{
    "submitted":     {"approved", "rejected", "cancelled"},
    "approved":      {"in_progress", "cancelled"},
    "in_progress":   {"seminar_ready", "cancelled"},
    "seminar_ready": {"seminar_done"},
    "seminar_done":  {"defense_ready"},
    "defense_ready": {"defence_done"},
    "defense_done":  {"graduated"},
  }

  func CanTransition(from, to string) bool
  ```

### Handler — Endpoint Mahasiswa

**POST `/api/v1/theses`** _(Mahasiswa only)_
- [ ] Request body:
  ```json
  {
    "title": "Judul Skripsi Mahasiswa",
    "abstract": "Abstrak minimal 100 kata...",
    "field_of_study": "Kecerdasan Buatan",
    "thesis_type": "skripsi"
  }
  ```
- [ ] Validasi:
  - Judul minimal 10 kata, maksimal 500 karakter
  - Abstrak minimal 100 kata
  - `thesis_type`: `skripsi` atau `tugas_akhir`
  - Mahasiswa tidak boleh punya thesis aktif (status bukan `cancelled` atau `graduated`)
- [ ] Set `academic_year_id` dari tahun akademik yang sedang aktif
- [ ] Set status awal: `submitted`
- [ ] Audit log: `THESIS_SUBMITTED`
- [ ] Email notification ke Kaprodi (semua akun dengan role `kaprodi`)
- [ ] Response: `201 Created`

### Handler — Endpoint Kaprodi

**GET `/api/v1/theses`** _(Admin + Kaprodi: semua; Dosen Pembimbing: miliknya; Mahasiswa: miliknya)_
- [ ] Query params: `status`, `academic_year_id`, `study_program`, `field_of_study`, `supervisor_id`, `search`, `page`, `per_page`
- [ ] Scope data berdasarkan role (implementasi di use case layer):
  - Admin/Kaprodi → semua thesis
  - Dosen Pembimbing → thesis dengan `thesis_supervisors.supervisor_id = me`
  - Dosen Penguji → thesis yang dia ditugaskan sebagai penguji
  - Mahasiswa → thesis miliknya sendiri

**GET `/api/v1/theses/:id`** _(semua role, scope sama)_
- [ ] Return detail thesis lengkap:
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
- [ ] Request body:
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
- [ ] Validasi: `decision` harus `approved` atau `rejected`
- [ ] Validasi state machine: status harus `submitted`
- [ ] Update status thesis sesuai decision
- [ ] Set `approved_at` jika decision = approved
- [ ] Audit log: `THESIS_APPROVED` atau `THESIS_REJECTED`
- [ ] Email notification ke mahasiswa

**PUT `/api/v1/theses/:id/assign-supervisor`** _(Kaprodi only)_
- [ ] Request body:
  ```json
  {
    "supervisor_ids": ["uuid-dosen-1", "uuid-dosen-2"]
  }
  ```
- [ ] Validasi:
  - Status thesis harus `approved`
  - Minimal 1, maksimal 2 supervisor
  - Setiap ID harus user dengan role `dosen_pembimbing` dan `is_active = true`
- [ ] Insert ke `thesis_supervisors`
- [ ] Update status thesis → `in_progress`
- [ ] Audit log: `SUPERVISOR_ASSIGNED`
- [ ] Email notification ke mahasiswa dan setiap dosen pembimbing yang ditunjuk

**GET `/api/v1/lecturers`** _(Kaprodi + Admin)_
- [ ] Return list dosen pembimbing aktif dengan info beban bimbingan:
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
- [ ] Urutkan dari beban paling rendah (memudahkan load balancing)

**PATCH `/api/v1/theses/:id/cancel`** _(Admin + Kaprodi)_
- [ ] Set status → `cancelled`
- [ ] Request body: `{ "reason": "..." }`
- [ ] Audit log: `THESIS_CANCELLED`

### Email Notification (gunakan email service dari Job 11 — stub dulu)
- [ ] Saat status `submitted` → email ke semua Kaprodi
- [ ] Saat status `approved` → email ke mahasiswa
- [ ] Saat status `rejected` → email ke mahasiswa
- [ ] Saat pembimbing ditunjuk → email ke mahasiswa + dosen pembimbing

> **Catatan:** Untuk job ini, implementasikan fungsi email notification sebagai stub (log ke console). Implementasi penuh email service ada di Job 11. Pastikan interface email service sudah didefinisikan agar mudah di-swap nanti.

### Interface Email Service
- [ ] Buat `internal/domain/service/email_service.go`:
  ```go
  type EmailService interface {
    SendThesisSubmitted(to []string, thesis *entity.Thesis) error
    SendThesisApproved(to string, thesis *entity.Thesis) error
    SendThesisRejected(to string, thesis *entity.Thesis, notes string) error
    SendSupervisorAssigned(studentEmail string, supervisorEmails []string, thesis *entity.Thesis) error
    // ... method lain ditambahkan di job berikutnya
  }
  ```
- [ ] Buat `pkg/email/stub_email_service.go` — implementasi yang hanya log ke console (untuk development)

---

## Done Criteria

- [ ] `POST /api/v1/theses` oleh mahasiswa → thesis dibuat, status `submitted`
- [ ] `POST /api/v1/theses` oleh mahasiswa yang sudah punya thesis aktif → `400 Bad Request`
- [ ] `PUT /api/v1/theses/:id/review` dengan decision `approved` → status jadi `approved`
- [ ] `PUT /api/v1/theses/:id/review` oleh mahasiswa → `403 Forbidden`
- [ ] `PUT /api/v1/theses/:id/review` saat status bukan `submitted` → `422 Unprocessable Entity`
- [ ] `PUT /api/v1/theses/:id/assign-supervisor` → status jadi `in_progress`, supervisors terdaftar
- [ ] `PUT /api/v1/theses/:id/assign-supervisor` dengan ID dosen yang rolenya bukan `dosen_pembimbing` → `400`
- [ ] `GET /api/v1/theses` oleh mahasiswa → hanya return thesis miliknya
- [ ] `GET /api/v1/theses` oleh Kaprodi → return semua thesis
- [ ] `GET /api/v1/lecturers` → return dosen diurutkan dari beban terendah
- [ ] Email notification (stub) tercatat di console untuk setiap event
- [ ] Semua action tercatat di `audit_logs`
