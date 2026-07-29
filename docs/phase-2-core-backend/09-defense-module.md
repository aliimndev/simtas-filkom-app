# Job 09 — Defense Module (Sidang Skripsi & Yudisium)

**Phase:** 2 — Core Backend
**Referensi PRD:** Section 6.6 (FR-DEFENSE-001 s/d FR-DEFENSE-005), Section 6.7 (FR-YUDISIUM-001)
**Prerequisites:** Job 08 (Seminar Module) ✅
**Estimasi:** 4 hari

---

## Objective

Implementasi alur lengkap sidang skripsi: pengajuan (dengan gate lulus seminar + dokumen sidang approved), penjadwalan, penilaian berbobot oleh penguji, hasil otomatis, revisi akhir, dan penetapan yudisium oleh Kaprodi. Setelah job ini selesai, seluruh alur akademik inti backend selesai dari ujung ke ujung.

---

## Checklist

### Defense Repository & Use Case
- [ ] Buat `internal/domain/repository/defense_repository.go` — interface:
  ```go
  type DefenseRepository interface {
    Create(ctx context.Context, defense *entity.ThesisDefense) error
    FindByID(ctx context.Context, id uuid.UUID) (*entity.ThesisDefense, error)
    FindByThesisID(ctx context.Context, thesisID uuid.UUID) (*entity.ThesisDefense, error)
    FindAll(ctx context.Context, filter DefenseFilter) ([]*entity.ThesisDefense, int64, error)
    UpdateSchedule(ctx context.Context, id uuid.UUID, scheduledAt time.Time, room string) error
    UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
    UpdateFinalScore(ctx context.Context, id uuid.UUID, score float64) error
    SetRevisionNotes(ctx context.Context, id uuid.UUID, notes string) error
    AssignExaminer(ctx context.Context, defenseID, examinerID, assignedBy uuid.UUID) error
    GetExaminers(ctx context.Context, defenseID uuid.UUID) ([]*entity.User, error)
    AddScore(ctx context.Context, score *entity.DefenseScore) error
    GetAllScores(ctx context.Context, defenseID uuid.UUID) ([]*entity.DefenseScore, error)
    HasExaminerScored(ctx context.Context, defenseID, examinerID uuid.UUID) (bool, error)
    CheckScheduleConflict(ctx context.Context, room string, scheduledAt time.Time, excludeID *uuid.UUID) (bool, error)
  }
  ```
- [ ] `DefenseFilter`: `Status`, `ThesisID`, `ExaminerID`, `DateFrom`, `DateTo`, `Page`, `PerPage`
- [ ] Buat `internal/usecase/defense_usecase.go`

### Komponen Penilaian Sidang (Bobot Tetap v1.0)
- [ ] Tambahkan ke `internal/domain/entity/grading.go`:
  ```go
  // Bobot sidang sama dengan seminar di v1.0
  var DefenseGradingComponents = []GradingComponent{
    {Name: "Presentasi",         Weight: 30.0},
    {Name: "Penguasaan Materi",  Weight: 30.0},
    {Name: "Kualitas Naskah",    Weight: 25.0},
    {Name: "Kemampuan Menjawab", Weight: 15.0},
  }
  ```
- [ ] Reuse `pkg/grading/calculator.go` dari Job 08

### Gate Checker — Sidang
- [ ] Tambahkan ke `internal/usecase/gate_checker.go`:
  ```go
  func (uc *DefenseUseCase) CanSubmitDefense(ctx context.Context, thesisID uuid.UUID) (bool, string, error)
  // Return: (canSubmit, reason jika tidak bisa, error)
  // Cek 1: thesis.status harus "seminar_done"
  // Cek 2: dokumen "defense_doc" harus berstatus "approved"
  ```

### Handler — Defense Endpoints

**POST `/api/v1/theses/:thesis_id/defenses`** _(Mahasiswa pemilik only)_
- [ ] Validasi gate:
  - `CanSubmitDefense()` harus return `true`
  - Tidak boleh ada defense aktif (status bukan `failed`) untuk thesis ini
- [ ] Buat record defense dengan status `pending`
- [ ] Update status thesis → `defense_ready`
- [ ] Audit log: `DEFENSE_SUBMITTED`
- [ ] Email notification ke Kaprodi dan Admin (stub)
- [ ] Response: `201 Created`

**GET `/api/v1/defenses`** _(Admin + Kaprodi: semua; Dosen Penguji: yang dia ditugaskan; Mahasiswa + Dosen Pembimbing: terkait thesis mereka)_
- [ ] Query params: `status`, `date_from`, `date_to`, `examiner_id`, `page`, `per_page`
- [ ] Return list defense dengan info lengkap

**GET `/api/v1/defenses/:id`** _(akses sama)_
- [ ] Return detail defense:
  ```json
  {
    "id": "uuid",
    "thesis": { "id": "...", "title": "...", "student": { ... } },
    "status": "scheduled",
    "scheduled_at": "2026-12-15T09:00:00Z",
    "room": "Ruang Sidang B",
    "examiners": [ { "id": "...", "full_name": "Dr. Siti" } ],
    "scores": [],
    "final_score": null,
    "revision_notes": null
  }
  ```

**PUT `/api/v1/defenses/:id/schedule`** _(Admin + Kaprodi only)_
- [ ] Request body:
  ```json
  {
    "scheduled_at": "2026-12-15T09:00:00Z",
    "room": "Ruang Sidang B",
    "examiner_ids": ["uuid-penguji-1", "uuid-penguji-2"]
  }
  ```
- [ ] Validasi:
  - Status defense harus `pending` atau `scheduled` (reschedule)
  - `scheduled_at` minimal 7 hari dari sekarang (lebih ketat dari seminar)
  - Minimal 2 penguji, role `dosen_penguji`, `is_active = true`
  - Cek konflik jadwal ruangan dan penguji (sama seperti seminar)
- [ ] Update status → `scheduled`
- [ ] Audit log: `DEFENSE_SCHEDULED` atau `DEFENSE_RESCHEDULED`
- [ ] Email notification ke mahasiswa, pembimbing, penguji (stub)

**POST `/api/v1/defenses/:id/scores`** _(Dosen Penguji yang ditugaskan only)_
- [ ] Struktur sama persis dengan seminar scores
- [ ] Validasi sama: penguji ditugaskan, belum submit, status `scheduled`, semua komponen diisi, nilai 0–100
- [ ] Trigger `TryFinalizeDefense()` setelah save
- [ ] Audit log: `DEFENSE_SCORE_SUBMITTED`

**GET `/api/v1/defenses/:id/result`** _(semua pihak terkait)_
- [ ] Return hasil lengkap dengan breakdown nilai
- [ ] Tambahkan field `grade_category` berdasarkan `final_score`:
  ```go
  func GetGradeCategory(score float64) string {
    switch {
    case score >= 85: return "A"
    case score >= 75: return "B+"
    case score >= 70: return "B"
    case score >= 60: return "C"
    default:          return "Tidak Lulus"
    }
  }
  ```

### Auto-Calculate Final Score (Defense)
- [ ] Implementasikan `TryFinalizeDefense()` di use case:
  ```go
  func (uc *DefenseUseCase) TryFinalizeDefense(ctx context.Context, defenseID uuid.UUID) error {
    // Cek apakah semua penguji sudah submit
    // Hitung final score
    // Tentukan status:
    //   < 60  → "failed"
    //   60-74 → "passed" (dengan revisi — dicatat sebagai revision_required)
    //   >= 75 → "passed"
    // Update defense: final_score, status
    // Update thesis status → "defense_done"
    // Send email notification hasil sidang (stub)
    // Audit log: DEFENSE_FINALIZED
  }
  ```

**PUT `/api/v1/defenses/:id/revision`** _(Admin + Kaprodi only)_
- [ ] Request body: `{ "revision_notes": "Perbaiki kesimpulan dan saran di BAB 5..." }`
- [ ] Hanya bisa dilakukan setelah defense selesai dinilai
- [ ] Set `revision_notes`
- [ ] Set status defense → `revision_required` jika belum
- [ ] Audit log: `DEFENSE_REVISION_NOTED`
- [ ] Email notification ke mahasiswa (stub)

### Yudisium

**PUT `/api/v1/theses/:thesis_id/graduation`** _(Kaprodi only)_
- [ ] Validasi gate:
  - Thesis harus berstatus `defense_done`
  - Defense harus berstatus `passed` atau `revision_required` (revisi sudah selesai)
  - Dokumen `final_thesis` harus berstatus `approved` (mahasiswa sudah upload skripsi final setelah revisi)
- [ ] Request body: `{ "notes": "Selamat, skripsi Anda telah memenuhi semua persyaratan." }`
- [ ] Set thesis status → `graduated`
- [ ] Set `graduated_at = NOW()`
- [ ] Audit log: `THESIS_GRADUATED`
- [ ] Email notification khusus "Selamat, skripsi Anda dinyatakan lulus" ke mahasiswa (stub)

### Endpoint Kombinasi — Jadwal Mendatang (untuk Dashboard)

**GET `/api/v1/schedules/upcoming`** _(Admin + Kaprodi)_
- [ ] Return seminar + sidang yang terjadwal dalam 14 hari ke depan
- [ ] Digunakan oleh dashboard operasional
- [ ] Response:
  ```json
  {
    "seminars": [ { seminar objects dengan info mahasiswa } ],
    "defenses": [ { defense objects dengan info mahasiswa } ]
  }
  ```

---

## Done Criteria

- [ ] `POST .../defenses` sebelum seminar lulus → `422` dengan pesan jelas
- [ ] `POST .../defenses` sebelum dokumen defense di-approve → `422`
- [ ] `POST .../defenses` setelah semua gate terpenuhi → defense dibuat, status `pending`
- [ ] `PUT /defenses/:id/schedule` dengan jadwal <7 hari → `400`
- [ ] `POST /defenses/:id/scores` dengan semua penguji submit → `final_score` terhitung otomatis
- [ ] `final_score` < 60 → status defense `failed`, thesis kembali ke `defense_ready`
- [ ] `final_score` >= 60 → status defense `passed`, thesis ke `defense_done`
- [ ] `GET /defenses/:id/result` → breakdown nilai per penguji + `grade_category`
- [ ] `PUT .../graduation` sebelum defense `passed` → `422`
- [ ] `PUT .../graduation` setelah semua prasyarat terpenuhi → thesis status `graduated`, `graduated_at` di-set
- [ ] `GET /schedules/upcoming` → return seminar dan sidang dalam 14 hari ke depan
- [ ] Email notification (stub) terpicu untuk semua event
- [ ] Semua action tercatat di `audit_logs`
- [ ] **MILESTONE:** Seluruh alur akademik backend (Job 05–09) dapat dijalankan end-to-end: pengajuan → approval → bimbingan → dokumen → seminar → sidang → yudisium
