# Job 08 — Seminar Proposal Module

**Phase:** 2 — Core Backend
**Referensi PRD:** Section 6.5 (FR-SEMINAR-001 s/d FR-SEMINAR-006)
**Prerequisites:** Job 07 (Document Management) ✅
**Estimasi:** 4 hari

---

## Objective

Implementasi alur lengkap seminar proposal: pengajuan oleh mahasiswa (dengan gate dari dokumen), penjadwalan oleh Admin/Kaprodi dengan penunjukan penguji, input nilai berbobot per penguji, kalkulasi nilai akhir otomatis, dan pencatatan revisi pasca seminar.

---

## Checklist

### Seminar Repository & Use Case
- [ ] Buat `internal/domain/repository/seminar_repository.go` — interface:
  ```go
  type SeminarRepository interface {
    Create(ctx context.Context, seminar *entity.Seminar) error
    FindByID(ctx context.Context, id uuid.UUID) (*entity.Seminar, error)
    FindByThesisID(ctx context.Context, thesisID uuid.UUID) (*entity.Seminar, error)
    FindAll(ctx context.Context, filter SeminarFilter) ([]*entity.Seminar, int64, error)
    UpdateSchedule(ctx context.Context, id uuid.UUID, scheduledAt time.Time, room string) error
    UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
    UpdateFinalScore(ctx context.Context, id uuid.UUID, score float64) error
    AssignExaminer(ctx context.Context, seminarID, examinerID, assignedBy uuid.UUID) error
    GetExaminers(ctx context.Context, seminarID uuid.UUID) ([]*entity.User, error)
    AddScore(ctx context.Context, score *entity.SeminarScore) error
    GetScoresByExaminer(ctx context.Context, seminarID, examinerID uuid.UUID) ([]*entity.SeminarScore, error)
    GetAllScores(ctx context.Context, seminarID uuid.UUID) ([]*entity.SeminarScore, error)
    HasExaminerScored(ctx context.Context, seminarID, examinerID uuid.UUID) (bool, error)
    CheckScheduleConflict(ctx context.Context, room string, scheduledAt time.Time, excludeID *uuid.UUID) (bool, error)
  }
  ```
- [ ] `SeminarFilter`: `Status`, `ThesisID`, `ExaminerID`, `DateFrom`, `DateTo`, `Page`, `PerPage`
- [ ] Buat `internal/usecase/seminar_usecase.go`

### Komponen Penilaian (Bobot Tetap v1.0)
- [ ] Definisikan konstanta bobot di `internal/domain/entity/grading.go`:
  ```go
  var SeminarGradingComponents = []GradingComponent{
    {Name: "Presentasi",         Weight: 30.0},
    {Name: "Penguasaan Materi",  Weight: 30.0},
    {Name: "Kualitas Naskah",    Weight: 25.0},
    {Name: "Kemampuan Menjawab", Weight: 15.0},
  }

  type GradingComponent struct {
    Name   string
    Weight float64
  }
  ```
- [ ] Buat `pkg/grading/calculator.go`:
  ```go
  // Hitung nilai akhir dari semua score penguji
  // Formula: rata-rata nilai berbobot dari semua penguji
  // Nilai_Penguji_i = Σ (score_komponen_j × bobot_j / 100)
  // Nilai_Akhir = Σ Nilai_Penguji_i / jumlah_penguji
  func CalculateFinalScore(scores []SeminarScore, examiners int) float64
  func CalculateExaminerScore(components []ComponentScore) float64
  ```

### Handler — Seminar Endpoints

**POST `/api/v1/theses/:thesis_id/seminars`** _(Mahasiswa pemilik only)_
- [ ] Validasi gate:
  - `CanSubmitSeminar()` harus return `true` (dokumen `seminar_doc` sudah `approved`)
  - Thesis harus berstatus `in_progress`
  - Tidak boleh ada seminar aktif (status bukan `failed`) untuk thesis ini
- [ ] Buat record seminar dengan status `pending`
- [ ] Update status thesis → `seminar_ready`
- [ ] Audit log: `SEMINAR_SUBMITTED`
- [ ] Email notification ke Kaprodi dan Admin (stub)
- [ ] Response: `201 Created`

**GET `/api/v1/seminars`** _(Admin + Kaprodi: semua; Dosen Penguji: yang dia ditugaskan; Mahasiswa + Dosen Pembimbing: terkait thesis mereka)_
- [ ] Query params: `status`, `date_from`, `date_to`, `examiner_id`, `page`, `per_page`
- [ ] Return list seminar dengan info thesis, mahasiswa, jadwal, penguji

**GET `/api/v1/seminars/:id`** _(akses sama)_
- [ ] Return detail seminar:
  ```json
  {
    "id": "uuid",
    "thesis": { "id": "...", "title": "...", "student": { ... } },
    "status": "scheduled",
    "scheduled_at": "2026-11-10T09:00:00Z",
    "room": "Ruang Sidang A",
    "examiners": [ { "id": "...", "full_name": "Dr. Budi" } ],
    "scores": [],
    "final_score": null,
    "notes": null
  }
  ```

**PUT `/api/v1/seminars/:id/schedule`** _(Admin + Kaprodi only)_
- [ ] Request body:
  ```json
  {
    "scheduled_at": "2026-11-10T09:00:00Z",
    "room": "Ruang Sidang A",
    "examiner_ids": ["uuid-penguji-1", "uuid-penguji-2"]
  }
  ```
- [ ] Validasi:
  - Status seminar harus `pending` atau `scheduled` (untuk reschedule)
  - `scheduled_at` minimal 3 hari dari sekarang
  - Minimal 2 penguji
  - Setiap penguji harus role `dosen_penguji` dan `is_active = true`
  - Cek konflik jadwal: ruangan tidak boleh bentrok (dalam ±2 jam di tanggal yang sama)
  - Cek konflik penguji: penguji tidak boleh jadwal seminar/sidang lain di waktu yang sama
- [ ] Replace examiner list (hapus yang lama, insert yang baru)
- [ ] Update status → `scheduled`
- [ ] Audit log: `SEMINAR_SCHEDULED` atau `SEMINAR_RESCHEDULED`
- [ ] Email notification ke mahasiswa, pembimbing, dan penguji (stub)

**POST `/api/v1/seminars/:id/scores`** _(Dosen Penguji yang ditugaskan only)_
- [ ] Request body:
  ```json
  {
    "scores": [
      { "component_name": "Presentasi",         "score": 85 },
      { "component_name": "Penguasaan Materi",  "score": 78 },
      { "component_name": "Kualitas Naskah",    "score": 80 },
      { "component_name": "Kemampuan Menjawab", "score": 75 }
    ]
  }
  ```
- [ ] Validasi:
  - Penguji harus terdaftar di `seminar_examiners` untuk seminar ini
  - Penguji belum pernah submit nilai untuk seminar ini (`HasExaminerScored` = false)
  - Status seminar harus `scheduled`
  - Semua komponen wajib diisi (sesuai `SeminarGradingComponents`)
  - Setiap nilai: 0–100
  - Nama komponen harus match dengan konstanta
- [ ] Simpan semua komponen score dengan bobot yang sesuai
- [ ] Cek apakah semua penguji sudah input nilai → jika ya, trigger kalkulasi nilai akhir
- [ ] Audit log: `SEMINAR_SCORE_SUBMITTED`

**GET `/api/v1/seminars/:id/result`** _(semua pihak terkait)_
- [ ] Cek apakah semua penguji sudah submit nilai
- [ ] Return hasil lengkap:
  ```json
  {
    "seminar_id": "uuid",
    "final_score": 80.25,
    "status": "passed",
    "is_complete": true,
    "examiner_scores": [
      {
        "examiner": { "full_name": "Dr. Budi" },
        "examiner_score": 82.5,
        "components": [
          { "name": "Presentasi", "weight": 30, "score": 85, "weighted": 25.5 }
        ]
      }
    ],
    "grading_components": [
      { "name": "Presentasi", "weight": 30 }
    ]
  }
  ```

### Auto-Calculate Final Score (trigger setelah semua penguji submit)
- [ ] Implementasikan di use case setelah score submission:
  ```go
  func (uc *SeminarUseCase) TryFinalizeSeminar(ctx context.Context, seminarID uuid.UUID) error {
    examiners := // get seminar examiners count
    scoredCount := // count distinct examiners who submitted scores
    if scoredCount < examiners { return nil } // belum semua submit

    allScores := // get all scores
    finalScore := grading.CalculateFinalScore(allScores, examiners)

    status := "passed"
    if finalScore < 60 { status = "failed" }

    // update seminar: final_score, status = passed/failed
    // update thesis status: seminar_done
    // send email notification hasil seminar
    // audit log: SEMINAR_FINALIZED
  }
  ```

**PUT `/api/v1/seminars/:id/revision`** _(Admin + Kaprodi only)_
- [ ] Request body: `{ "revision_notes": "Perbaiki rumusan masalah di BAB 1..." }`
- [ ] Hanya bisa dilakukan setelah seminar `passed`
- [ ] Simpan catatan revisi di field `notes`
- [ ] Audit log: `SEMINAR_REVISION_NOTED`

---

## Done Criteria

- [ ] `POST .../seminars` sebelum dokumen seminar di-approve → `422 Unprocessable Entity`
- [ ] `POST .../seminars` setelah dokumen approved → seminar dibuat, status `pending`
- [ ] `PUT /seminars/:id/schedule` → status jadi `scheduled`, penguji terdaftar
- [ ] `PUT /seminars/:id/schedule` dengan jadwal <3 hari dari sekarang → `400`
- [ ] `PUT /seminars/:id/schedule` dengan ruangan konflik → `409 Conflict`
- [ ] `POST /seminars/:id/scores` oleh penguji yang tidak ditugaskan → `403`
- [ ] `POST /seminars/:id/scores` oleh penguji yang sudah submit → `409`
- [ ] Setelah semua penguji submit nilai → `final_score` terhitung, status `passed` atau `failed`
- [ ] Kalkulasi: 2 penguji, nilai masing-masing 80 dan 90 dengan bobot standar → rata-rata = 85
- [ ] `GET /seminars/:id/result` → detail breakdown nilai per penguji per komponen
- [ ] Email notification (stub) terpicu untuk setiap event
- [ ] Semua action tercatat di `audit_logs`
