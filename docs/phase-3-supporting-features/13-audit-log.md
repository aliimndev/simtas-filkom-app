# Job 13 — Audit Log & System Logging

**Phase:** 3 — Supporting Features
**Referensi PRD:** Section 6.13 (FR-AUDIT-001), Section 7 (Non-Functional Requirements — Logging & Error Handling)
**Prerequisites:** Job 12 (Dashboard) ✅
**Estimasi:** 2 hari

---

## Objective

Implementasi sistem audit log terpusat yang sudah dipanggil sebagai stub di semua job sebelumnya, endpoint untuk query audit log, structured logging untuk aplikasi, dan global error handler yang konsisten. Setelah job ini selesai, seluruh aktivitas kritis sistem tercatat dan dapat di-query oleh Admin.

---

## Checklist

### Audit Log Service (Implementasi Penuh)

Semua job sebelumnya sudah memanggil audit log secara stub. Job ini mengimplementasikannya secara nyata.

- [ ] Buat `backend/internal/domain/repository/audit_repository.go`:
  ```go
  type AuditRepository interface {
    Create(ctx context.Context, log *entity.AuditLog) error
    FindAll(ctx context.Context, filter AuditFilter) ([]*entity.AuditLog, int64, error)
    FindByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]*entity.AuditLog, error)
  }
  ```
- [ ] `AuditFilter`:
  ```go
  type AuditFilter struct {
    UserID     *uuid.UUID
    Action     string
    EntityType string
    EntityID   *uuid.UUID
    DateFrom   *time.Time
    DateTo     *time.Time
    Page       int
    PerPage    int
  }
  ```
- [ ] Buat `backend/pkg/audit/audit_service.go` — service yang bisa diinjeksi ke semua use case:
  ```go
  type AuditService struct {
    repo AuditRepository
  }

  func (s *AuditService) Log(ctx context.Context, params AuditParams) {
    // Non-blocking: jalankan di goroutine agar tidak delay response
    go func() {
      log := &entity.AuditLog{
        UserID:     params.UserID,
        Action:     params.Action,
        EntityType: params.EntityType,
        EntityID:   params.EntityID,
        OldValue:   params.OldValue,  // jsonb
        NewValue:   params.NewValue,  // jsonb
        IPAddress:  params.IPAddress,
        UserAgent:  params.UserAgent,
      }
      if err := s.repo.Create(context.Background(), log); err != nil {
        // log error ke application logger, jangan panic
        appLogger.Error("failed to write audit log", "error", err)
      }
    }()
  }
  ```
- [ ] `AuditParams` struct dengan field: `UserID`, `Action`, `EntityType`, `EntityID`, `OldValue` (interface{}), `NewValue` (interface{}), `IPAddress`, `UserAgent`
- [ ] Helper untuk extract IP dan UserAgent dari Gin context

### Konstanta Action
- [ ] Buat `backend/pkg/audit/actions.go` — semua konstanta action:
  ```go
  const (
    // Auth
    ActionUserLogin          = "USER_LOGIN"
    ActionUserLoginFailed    = "USER_LOGIN_FAILED"
    ActionUserLogout         = "USER_LOGOUT"
    ActionPasswordReset      = "USER_PASSWORD_RESET"

    // User Management
    ActionUserCreated        = "USER_CREATED"
    ActionUserUpdated        = "USER_UPDATED"
    ActionUserDeleted        = "USER_DELETED"
    ActionUserActivated      = "USER_ACTIVATED"
    ActionUserDeactivated    = "USER_DEACTIVATED"
    ActionUserBulkImported   = "USER_BULK_IMPORTED"

    // Thesis
    ActionThesisSubmitted    = "THESIS_SUBMITTED"
    ActionThesisApproved     = "THESIS_APPROVED"
    ActionThesisRejected     = "THESIS_REJECTED"
    ActionThesisCancelled    = "THESIS_CANCELLED"
    ActionSupervisorAssigned = "SUPERVISOR_ASSIGNED"
    ActionThesisGraduated    = "THESIS_GRADUATED"

    // Consultation
    ActionConsultationCreated  = "CONSULTATION_CREATED"
    ActionConsultationApproved = "CONSULTATION_APPROVED"
    ActionConsultationUpdated  = "CONSULTATION_UPDATED"

    // Document
    ActionDocumentUploaded    = "DOCUMENT_UPLOADED"
    ActionDocumentApproved    = "DOCUMENT_APPROVED"
    ActionDocumentRevision    = "DOCUMENT_REVISION_REQUESTED"
    ActionDocumentDownloaded  = "DOCUMENT_DOWNLOADED"

    // Seminar
    ActionSeminarSubmitted    = "SEMINAR_SUBMITTED"
    ActionSeminarScheduled    = "SEMINAR_SCHEDULED"
    ActionSeminarRescheduled  = "SEMINAR_RESCHEDULED"
    ActionSeminarScoreSubmit  = "SEMINAR_SCORE_SUBMITTED"
    ActionSeminarFinalized    = "SEMINAR_FINALIZED"

    // Defense
    ActionDefenseSubmitted    = "DEFENSE_SUBMITTED"
    ActionDefenseScheduled    = "DEFENSE_SCHEDULED"
    ActionDefenseRescheduled  = "DEFENSE_RESCHEDULED"
    ActionDefenseScoreSubmit  = "DEFENSE_SCORE_SUBMITTED"
    ActionDefenseFinalized    = "DEFENSE_FINALIZED"

    // Archive
    ActionArchiveCreated      = "ARCHIVE_CREATED"
    ActionArchiveDownloaded   = "ARCHIVE_DOWNLOADED"
  )
  ```

### Swap Stub → Real Audit Log
- [ ] Inject `AuditService` ke semua use case yang memanggil audit log (job 03–12)
- [ ] Pastikan setiap action yang didefinisikan benar-benar dipanggil di tempat yang sesuai
- [ ] IPAddress dan UserAgent diambil dari Gin context dan diteruskan ke use case via context

### Handler — Audit Log Endpoints (Admin Only)

**GET `/api/v1/admin/audit-logs`** _(Admin only)_
- [ ] Query params: `user_id`, `action`, `entity_type`, `entity_id`, `date_from`, `date_to`, `page`, `per_page`
- [ ] Response:
  ```json
  {
    "success": true,
    "data": [
      {
        "id": "uuid",
        "user": { "id": "...", "full_name": "Admin", "email": "..." },
        "action": "THESIS_APPROVED",
        "entity_type": "thesis",
        "entity_id": "uuid",
        "old_value": { "status": "submitted" },
        "new_value": { "status": "approved" },
        "ip_address": "192.168.1.1",
        "user_agent": "Mozilla/5.0...",
        "created_at": "2026-10-15T10:00:00Z"
      }
    ],
    "meta": { "page": 1, "per_page": 50, "total": 1250 }
  }
  ```
- [ ] Default per_page: 50 (lebih besar dari endpoint lain karena sifat monitoring)

**GET `/api/v1/admin/audit-logs/entity/:entity_type/:entity_id`** _(Admin + Kaprodi)_
- [ ] Return riwayat audit untuk 1 entitas spesifik (misal: semua perubahan pada thesis tertentu)
- [ ] Berguna untuk menelusuri riwayat lengkap sebuah skripsi

### Structured Application Logger

- [ ] Buat `backend/pkg/logger/logger.go` — wrapper untuk `log/slog` (Go standard library):
  ```go
  var Logger *slog.Logger

  func Init(env string) {
    if env == "production" {
      // JSON format untuk production (mudah di-parse oleh log aggregator)
      Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
      }))
    } else {
      // Text format untuk development
      Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelDebug,
      }))
    }
    slog.SetDefault(Logger)
  }
  ```
- [ ] Buat `backend/internal/middleware/request_logger.go` — log setiap HTTP request:
  ```go
  // Log format: METHOD PATH STATUS LATENCY IP
  // Contoh: GET /api/v1/theses 200 12ms 192.168.1.1
  ```
- [ ] Jangan log request body (bisa berisi data sensitif)
- [ ] Tambahkan request ID ke setiap log entry (untuk tracing)

### Global Error Handler

- [ ] Buat `backend/internal/middleware/error_handler.go` — Gin recovery middleware yang menangkap panic:
  ```go
  func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
      defer func() {
        if err := recover(); err != nil {
          // Log stack trace
          slog.Error("panic recovered", "error", err, "stack", debug.Stack())
          c.JSON(500, gin.H{
            "success": false,
            "message": "Terjadi kesalahan pada server",
          })
          c.Abort()
        }
      }()
      c.Next()
    }
  }
  ```
- [ ] Buat `backend/pkg/apperror/errors.go` — custom error types:
  ```go
  type AppError struct {
    Code    int
    Message string
    Err     error
  }

  var (
    ErrNotFound      = &AppError{Code: 404, Message: "Data tidak ditemukan"}
    ErrUnauthorized  = &AppError{Code: 401, Message: "Akses tidak diizinkan"}
    ErrForbidden     = &AppError{Code: 403, Message: "Akses ditolak"}
    ErrBadRequest    = &AppError{Code: 400, Message: "Request tidak valid"}
    ErrConflict      = &AppError{Code: 409, Message: "Data sudah ada"}
    ErrUnprocessable = &AppError{Code: 422, Message: "Proses tidak dapat dilakukan"}
    ErrInternal      = &AppError{Code: 500, Message: "Terjadi kesalahan pada server"}
  )
  ```
- [ ] Handler Gin menggunakan `AppError` untuk return response yang konsisten

### Token Blacklist Cleanup (Scheduled Task)
- [ ] Buat `backend/pkg/scheduler/token_cleanup.go` — goroutine yang berjalan setiap 1 jam:
  ```go
  func StartTokenCleanup(db *gorm.DB) {
    ticker := time.NewTicker(1 * time.Hour)
    go func() {
      for range ticker.C {
        db.Where("expires_at < ?", time.Now()).Delete(&entity.TokenBlacklist{})
        slog.Info("token blacklist cleanup done")
      }
    }()
  }
  ```
- [ ] Jalankan di `main.go` setelah startup

---

## Done Criteria

- [ ] Setiap action dari job 03–12 berhasil tercatat di tabel `audit_logs`
- [ ] `GET /admin/audit-logs` → return paginated list dengan filter
- [ ] `GET /admin/audit-logs?action=THESIS_APPROVED` → filter by action bekerja
- [ ] `GET /admin/audit-logs/entity/thesis/:id` → riwayat lengkap 1 thesis
- [ ] Audit log ditulis non-blocking (tidak memperlambat HTTP response)
- [ ] Request logger mencatat setiap HTTP request dengan method, path, status, latency
- [ ] Panic di handler ter-recover, return `500` dengan format standard (tidak crash server)
- [ ] `AppError` digunakan konsisten di semua handler (bukan plain `c.JSON(400, ...)`)
- [ ] Token blacklist cleanup berjalan dan menghapus expired entries
- [ ] **MILESTONE Phase 3:** Seluruh backend API lengkap dan siap untuk diintegrasikan dengan frontend
