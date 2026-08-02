# Job 12 — Dashboard & Statistics (Backend)

**Phase:** 3 — Supporting Features
**Referensi PRD:** Section 6.9 (FR-DASHBOARD-001 s/d FR-DASHBOARD-004)
**Prerequisites:** Job 11 (Email Notification) ✅
**Estimasi:** 3 hari

---

## Objective

Implementasi semua endpoint API dashboard untuk setiap role: dashboard komprehensif Admin/Kaprodi (ringkasan akademik + analitik dosen + operasional), dashboard progress mahasiswa, dashboard dosen pembimbing, dan dashboard dosen penguji. Semua query dioptimasi menggunakan aggregasi di database level.

---

## Checklist

### Dashboard Repository
- [x] Buat `backend/internal/domain/repository/dashboard_repository.go` — interface:
  ```go
  type DashboardRepository interface {
    // Admin & Kaprodi
    GetAcademicSummary(ctx context.Context, filter DashboardFilter) (*AcademicSummary, error)
    GetThesisByStatus(ctx context.Context, filter DashboardFilter) ([]StatusCount, error)
    GetGraduationTrend(ctx context.Context, filter DashboardFilter) ([]MonthlyCount, error)
    GetLecturerWorkload(ctx context.Context, filter DashboardFilter) ([]LecturerWorkload, error)
    GetPendingActions(ctx context.Context) (*PendingActions, error)
    GetUpcomingSchedules(ctx context.Context, days int) (*UpcomingSchedules, error)
    GetActivityStats(ctx context.Context) (*ActivityStats, error)

    // Mahasiswa
    GetStudentProgress(ctx context.Context, studentID uuid.UUID) (*StudentProgress, error)

    // Dosen Pembimbing
    GetSupervisorDashboard(ctx context.Context, supervisorID uuid.UUID) (*SupervisorDashboard, error)

    // Dosen Penguji
    GetExaminerDashboard(ctx context.Context, examinerID uuid.UUID) (*ExaminerDashboard, error)
  }
  ```

- [x] `DashboardFilter`:
  ```go
  type DashboardFilter struct {
    AcademicYearID *uuid.UUID
    Semester       string  // ganjil / genap
    StudyProgram   string
  }
  ```

### Struct Hasil Query
- [x] Definisikan semua response struct di `backend/internal/handler/dto/dashboard_dto.go`:
  ```go
  type AcademicSummary struct {
    TotalActive        int     `json:"total_active"`
    TotalGraduated     int     `json:"total_graduated"`
    AvgCompletionMonths float64 `json:"avg_completion_months"`
  }

  type StatusCount struct {
    Status string `json:"status"`
    Count  int    `json:"count"`
  }

  type MonthlyCount struct {
    Month string `json:"month"`  // "2026-10"
    Count int    `json:"count"`
  }

  type LecturerWorkload struct {
    LecturerID     uuid.UUID `json:"lecturer_id"`
    FullName        string    `json:"full_name"`
    NIDN            string    `json:"nidn"`
    SupervisionCount int      `json:"supervision_count"`
    SeminarCount    int       `json:"seminar_count"`
    DefenseCount    int       `json:"defense_count"`
  }

  type PendingActions struct {
    PendingTitleReviews    int `json:"pending_title_reviews"`
    PendingDocumentReviews int `json:"pending_document_reviews"`
    PendingSeminars        int `json:"pending_seminars"`
    PendingDefenses        int `json:"pending_defenses"`
  }

  type StudentProgress struct {
    ThesisID        uuid.UUID     `json:"thesis_id"`
    Title           string        `json:"title"`
    Status          string        `json:"status"`
    CurrentStage    string        `json:"current_stage"`
    Supervisors     []UserSummary `json:"supervisors"`
    Documents       []DocStatus   `json:"documents"`
    ConsultationCount int         `json:"consultation_count"`
    LastConsultation  *time.Time  `json:"last_consultation"`
    UpcomingSeminar   *ScheduleInfo `json:"upcoming_seminar"`
    UpcomingDefense   *ScheduleInfo `json:"upcoming_defense"`
  }
  ```

### Handler — Dashboard Endpoints

**GET `/api/v1/dashboard/summary`** _(Admin + Kaprodi only)_
- [x] Query params: `academic_year_id`, `semester`, `study_program`
- [x] Response:
  ```json
  {
    "success": true,
    "data": {
      "academic_summary": {
        "total_active": 85,
        "total_graduated": 42,
        "avg_completion_months": 14.5
      },
      "by_status": [
        { "status": "submitted",     "label": "Menunggu Review",  "count": 8 },
        { "status": "in_progress",   "label": "Bimbingan",        "count": 35 },
        { "status": "seminar_ready", "label": "Siap Seminar",     "count": 12 },
        { "status": "seminar_done",  "label": "Pasca Seminar",    "count": 10 },
        { "status": "defense_ready", "label": "Siap Sidang",      "count": 8 },
        { "status": "defense_done",  "label": "Pasca Sidang",     "count": 5 },
        { "status": "graduated",     "label": "Lulus",            "count": 42 }
      ],
      "graduation_trend": [
        { "month": "2026-08", "count": 5 },
        { "month": "2026-09", "count": 8 }
      ]
    }
  }
  ```

**GET `/api/v1/dashboard/lecturer-analytics`** _(Admin + Kaprodi only)_
- [x] Query params: sama
- [x] Response:
  ```json
  {
    "success": true,
    "data": {
      "lecturers": [
        {
          "lecturer_id": "uuid",
          "full_name": "Dr. Ahmad",
          "nidn": "123456",
          "supervision_count": 5,
          "seminar_count": 8,
          "defense_count": 6
        }
      ],
      "workload_distribution": {
        "max_supervision": 8,
        "min_supervision": 1,
        "avg_supervision": 4.2,
        "highest_load": { "full_name": "Dr. X", "supervision_count": 8 },
        "lowest_load":  { "full_name": "Dr. Y", "supervision_count": 1 }
      }
    }
  }
  ```

**GET `/api/v1/dashboard/operational`** _(Admin + Kaprodi only)_
- [x] Response:
  ```json
  {
    "success": true,
    "data": {
      "pending_actions": {
        "pending_title_reviews": 5,
        "pending_document_reviews": 12,
        "pending_seminars": 3,
        "pending_defenses": 2
      },
      "upcoming_schedules": {
        "seminars": [
          {
            "id": "uuid",
            "student_name": "Budi",
            "thesis_title": "...",
            "scheduled_at": "2026-11-10T09:00:00Z",
            "room": "Ruang A"
          }
        ],
        "defenses": [ { ... } ]
      },
      "activity_stats": {
        "logins_today": 15,
        "documents_uploaded_this_week": 8,
        "consultations_this_week": 23
      }
    }
  }
  ```

**GET `/api/v1/dashboard/student`** _(Mahasiswa — hanya data miliknya)_
- [x] Return `StudentProgress` lengkap:
  ```json
  {
    "success": true,
    "data": {
      "thesis_id": "uuid",
      "title": "Judul Skripsi",
      "status": "in_progress",
      "current_stage": "Proses Bimbingan",
      "progress_percentage": 35,
      "supervisors": [ { "full_name": "Dr. Ahmad", "email": "..." } ],
      "documents": [
        { "type": "proposal", "status": "approved", "version": 1 },
        { "type": "seminar_doc", "status": "pending_review", "version": 1 }
      ],
      "consultation_count": 8,
      "last_consultation": "2026-10-15",
      "upcoming_seminar": null,
      "upcoming_defense": null,
      "pending_actions": [
        "Upload revisi dokumen seminar_doc"
      ]
    }
  }
  ```
- [x] `progress_percentage` dihitung berdasarkan stage:
  ```go
  var StageProgress = map[string]int{
    "submitted":     10,
    "approved":      15,
    "in_progress":   30,
    "seminar_ready": 45,
    "seminar_done":  60,
    "defense_ready": 75,
    "defense_done":  90,
    "graduated":     100,
  }
  ```
- [x] `pending_actions` berisi list string aksi yang harus dilakukan mahasiswa saat ini

**GET `/api/v1/dashboard/supervisor`** _(Dosen Pembimbing)_
- [x] Return ringkasan semua mahasiswa bimbingan:
  ```json
  {
    "success": true,
    "data": {
      "total_students": 5,
      "students": [
        {
          "thesis_id": "uuid",
          "student": { "full_name": "Budi", "nim": "123" },
          "title": "...",
          "status": "in_progress",
          "current_stage": "Proses Bimbingan",
          "pending_document_reviews": 2,
          "last_consultation": "2026-10-10",
          "consultation_count": 5,
          "days_since_last_consultation": 12
        }
      ],
      "pending_document_reviews": 3,
      "upcoming_schedules": [ { seminar/defense } ]
    }
  }
  ```
- [x] Urutkan mahasiswa: yang paling lama tidak bimbingan di atas (perlu perhatian lebih)

**GET `/api/v1/dashboard/examiner`** _(Dosen Penguji)_
- [x] Return:
  ```json
  {
    "success": true,
    "data": {
      "upcoming_assignments": [
        {
          "type": "seminar",
          "thesis_title": "...",
          "student_name": "...",
          "scheduled_at": "2026-11-10T09:00:00Z",
          "room": "Ruang A",
          "has_scored": false
        }
      ],
      "pending_scores": [ { seminar/defense yang belum diisi nilainya } ],
      "scoring_history": [ { seminar/defense yang sudah dinilai } ]
    }
  }
  ```

### Query Optimasi
- [x] Semua query aggregasi menggunakan SQL langsung (bukan loop di Go):
  ```go
  // Contoh: count thesis per status
  var results []StatusCount
  db.Raw(`
    SELECT status, COUNT(*) as count
    FROM theses
    WHERE deleted_at IS NULL
    AND academic_year_id = ?
    GROUP BY status
  `, academicYearID).Scan(&results)
  ```
- [x] Tambahkan index di database jika query lambat (EXPLAIN ANALYZE)
- [x] Target response time: < 500ms untuk semua dashboard endpoint

---

## Done Criteria

- [x] `GET /dashboard/summary` → return data akurat sesuai state database
- [x] `GET /dashboard/summary?academic_year_id=xxx` → data terfilter per tahun akademik
- [x] `GET /dashboard/lecturer-analytics` → workload per dosen akurat
- [x] `GET /dashboard/operational` → pending actions count sesuai data aktual
- [x] `GET /dashboard/student` → progress mahasiswa + pending actions tepat
- [x] `GET /dashboard/supervisor` → daftar mahasiswa bimbingan + status masing-masing
- [x] `GET /dashboard/examiner` → jadwal dan pending scores
- [x] Semua endpoint dashboard hanya bisa diakses role yang sesuai
- [x] Response time < 500ms untuk dataset 100 user (ukur dengan `time curl`)
- [x] `progress_percentage` terhitung dengan benar untuk setiap status
