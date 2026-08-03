package handler

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/service"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/middleware"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/usecase"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/config"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/email"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/jwt"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/storage"
)

// Router wires all dependencies and registers routes on the Gin engine
type Router struct {
	engine              *gin.Engine
	cfg                 *config.Config
	db                  *gorm.DB
	authHandler         *AuthHandler
	userHandler         *UserHandler
	academicYearHandler *AcademicYearHandler
	thesisHandler       *ThesisHandler
	consultationHandler *ConsultationHandler
	documentHandler     *DocumentHandler
	seminarHandler      *SeminarHandler
	defenseHandler      *DefenseHandler
	archiveHandler      *ArchiveHandler
	dashboardHandler    *DashboardHandler
	auditHandler        *AuditHandler
	internalHandler     *InternalHandler
	authMid             *middleware.AuthMiddleware
}

func NewRouter(engine *gin.Engine, db *gorm.DB, cfg *config.Config) *Router {
	jwtManager := jwt.NewJWTManager(
		cfg.JWTSecret,
		cfg.JWTExpiry,
		cfg.JWTRefreshExpiry,
	)

	authRepository := repository.NewAuthRepository(db)
	userRepository := repository.NewUserRepository(db)
	academicYearRepository := repository.NewAcademicYearRepository(db)
	thesisRepository := repository.NewThesisRepository(db)
	consultationRepository := repository.NewConsultationRepository(db)
	documentRepository := repository.NewDocumentRepository(db)
	seminarRepository := repository.NewSeminarRepository(db)
	defenseRepository := repository.NewDefenseRepository(db)
	archiveRepository := repository.NewArchiveRepository(db)
	dashboardRepository := repository.NewDashboardRepository(db)
	auditRepository := repository.NewAuditRepository(db)

	auditService := audit.NewAuditService(auditRepository)
	// Email: real Resend implementation in production; console-logging dev mode
	// when EMAIL_DEV_MODE=true or when no API key is configured.
	emailService := email.NewResendEmailService(cfg.ResendAPIKey, cfg.EmailFrom, cfg.EmailFromName, cfg.FrontendURL, db, cfg.EmailDevMode || cfg.ResendAPIKey == "")

	// Storage (Job 21 + production): real Supabase when STORAGE_PROVIDER=supabase
	// and a project URL + service-role key are configured; real S3-compatible
	// when STORAGE_PROVIDER=s3 (MinIO, R2, B2, etc.); otherwise the local stub.
	var storageService service.StorageService = storage.NewStubStorageService("", "")
	switch cfg.StorageProvider {
	case "supabase":
		if cfg.SupabaseURL != "" && cfg.SupabaseServiceRoleKey != "" {
			storageService = storage.NewSupabaseStorageService(cfg.SupabaseURL, cfg.SupabaseServiceRoleKey, cfg.SupabaseDocumentsBucket, cfg.SupabaseArchivesBucket)
		} else {
			// Loudly fall back: silently using the local stub in production would
			// mask a missing-credentials misconfiguration.
			slog.Warn("STORAGE_PROVIDER=supabase but SUPABASE_URL / SUPABASE_SERVICE_ROLE_KEY are empty; falling back to local storage")
		}
	case "s3":
		if cfg.S3Endpoint != "" && cfg.S3AccessKey != "" && cfg.S3SecretKey != "" {
			svc, err := storage.NewS3StorageService(
				cfg.S3Endpoint,
				cfg.S3Region,
				cfg.S3AccessKey,
				cfg.S3SecretKey,
				cfg.S3DocumentsBucket,
				cfg.S3ArchivesBucket,
				true, // MinIO/R2 typically need path-style
			)
			if err != nil {
				slog.Warn("failed to init S3 storage, falling back to local", "error", err)
			} else {
				storageService = svc
			}
		} else {
			slog.Warn("STORAGE_PROVIDER=s3 but S3_ENDPOINT / S3_ACCESS_KEY / S3_SECRET_KEY are empty; falling back to local storage")
		}
	}

	authUseCase := usecase.NewAuthUseCase(authRepository, jwtManager, auditService)
	userUseCase := usecase.NewUserUseCase(userRepository, emailService, auditService)
	academicYearUseCase := usecase.NewAcademicYearUseCase(academicYearRepository)
	thesisUseCase := usecase.NewThesisUseCase(thesisRepository, userRepository, academicYearRepository, emailService, auditService)
	consultationUseCase := usecase.NewConsultationUseCase(consultationRepository, thesisRepository, emailService, auditService)
	documentUseCase := usecase.NewDocumentUseCase(documentRepository, thesisRepository, storageService, emailService, auditService)
	seminarUseCase := usecase.NewSeminarUseCase(seminarRepository, thesisRepository, userRepository, documentUseCase, emailService, auditService)
	defenseUseCase := usecase.NewDefenseUseCase(defenseRepository, seminarRepository, thesisRepository, userRepository, documentUseCase, emailService, auditService)
	archiveUseCase := usecase.NewArchiveUseCase(archiveRepository, thesisRepository, storageService, emailService, auditService)
	dashboardUseCase := usecase.NewDashboardUseCase(dashboardRepository)
	auditUseCase := usecase.NewAuditUseCase(auditRepository)

	authHandler := NewAuthHandler(authUseCase)
	userHandler := NewUserHandler(userUseCase)
	academicYearHandler := NewAcademicYearHandler(academicYearUseCase)
	thesisHandler := NewThesisHandler(thesisUseCase)
	consultationHandler := NewConsultationHandler(consultationUseCase)
	documentHandler := NewDocumentHandler(documentUseCase)
	seminarHandler := NewSeminarHandler(seminarUseCase)
	defenseHandler := NewDefenseHandler(defenseUseCase)
	archiveHandler := NewArchiveHandler(archiveUseCase)
	dashboardHandler := NewDashboardHandler(dashboardUseCase)
	auditHandler := NewAuditHandler(auditUseCase)
	internalHandler := NewInternalHandler(emailService)

	authMiddleware := middleware.NewAuthMiddleware(jwtManager, authRepository, authRepository)

	return &Router{
		engine:              engine,
		cfg:                 cfg,
		db:                  db,
		authHandler:         authHandler,
		userHandler:         userHandler,
		academicYearHandler: academicYearHandler,
		thesisHandler:       thesisHandler,
		consultationHandler: consultationHandler,
		documentHandler:     documentHandler,
		seminarHandler:      seminarHandler,
		defenseHandler:      defenseHandler,
		archiveHandler:      archiveHandler,
		dashboardHandler:    dashboardHandler,
		auditHandler:        auditHandler,
		internalHandler:     internalHandler,
		authMid:             authMiddleware,
	}
}

func (r *Router) Setup() {
	// ── Global middleware ─────────────────────────────────────────────────
	r.engine.Use(middleware.ErrorHandler())
	r.engine.Use(middleware.RequestLogger())
	r.engine.Use(middleware.SecurityHeadersMiddleware())
	r.engine.Use(middleware.CORSMiddleware(r.cfg.CORSAllowedOrigins))

	// ── API v1 ────────────────────────────────────────────────────────────
	v1 := r.engine.Group("/api/v1")

	// Health check (public, unauthenticated) — Job 27: reports DB status + 503
	RegisterHealthRoutes(v1, r.db)

	// ── Local storage fallback (Job 21) ───────────────────────────────────
	// In development with STORAGE_PROVIDER=local, serve uploaded files from
	// ./tmp/uploads so generated presigned-style URLs actually work.
	if r.cfg.StorageProvider == "local" {
		r.engine.Static("/tmp/uploads", "./tmp/uploads")
	}

	// Public auth routes
	authPublic := v1.Group("/auth")
	{
		// Rate-limited login: max 10 req/min per IP
		authPublic.POST("/login",
			middleware.RateLimitMiddleware(10, time.Minute),
			r.authHandler.Login,
		)
		authPublic.POST("/refresh", r.authHandler.RefreshToken)
		authPublic.POST("/forgot-password", r.authHandler.ForgotPassword)
		authPublic.POST("/reset-password", r.authHandler.ResetPassword)
	}

	// Protected auth routes (token required)
	authProtected := v1.Group("/auth")
	authProtected.Use(r.authMid.Authenticate())
	{
		authProtected.POST("/logout", r.authHandler.Logout)
		authProtected.GET("/me", r.authHandler.GetMe)
	}

	// Change own password (any authenticated user) — Job 04/15
	me := v1.Group("/users/me", r.authMid.Authenticate())
	{
		me.PUT("/password", r.userHandler.ChangeMyPassword)
	}

	// ── Academic years ────────────────────────────────────────────────────
	// GET /api/v1/academic-years is available to all authenticated users
	academicYearsPublic := v1.Group("/academic-years")
	academicYearsPublic.Use(r.authMid.Authenticate())
	{
		academicYearsPublic.GET("", r.academicYearHandler.List)
	}

	// POST/PUT/PATCH /api/v1/academic-years are admin-only
	academicYearsAdmin := v1.Group("/academic-years",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleAdminFakultas),
	)
	{
		academicYearsAdmin.POST("", r.academicYearHandler.Create)
		academicYearsAdmin.PUT("/:id", r.academicYearHandler.Update)
		academicYearsAdmin.PATCH("/:id/activate", r.academicYearHandler.Activate)
	}

	// ── Admin-only routes ─────────────────────────────────────────────────
	admin := v1.Group("/admin",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleAdminFakultas),
	)
	{
		// User management
		admin.GET("/users", r.userHandler.ListUsers)
		admin.GET("/users/:id", r.userHandler.GetUser)
		admin.POST("/users", r.userHandler.CreateUser)
		admin.PUT("/users/:id", r.userHandler.UpdateUser)
		admin.DELETE("/users/:id", r.userHandler.DeleteUser)
		admin.PATCH("/users/:id/activate", r.userHandler.ActivateUser)
		admin.PATCH("/users/:id/deactivate", r.userHandler.DeactivateUser)
		admin.POST("/users/:id/reset-password", r.userHandler.ResetPassword)
		admin.GET("/users/import-template", r.userHandler.ImportTemplate)
		admin.POST("/users/import", r.userHandler.ImportUsers)

		// Audit logs (Job 13) — admin only
		admin.GET("/audit-logs", r.auditHandler.List)
	}

	// Audit log entity history (Job 13) — admin + kaprodi
	auditEntity := v1.Group("/admin/audit-logs/entity",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleAdminFakultas, middleware.RoleKaprodi),
	)
	{
		auditEntity.GET("/:entity_type/:entity_id", r.auditHandler.ByEntity)
	}

	// ── Theses ────────────────────────────────────────────────────────────
	// POST /theses is mahasiswa-only; GET is available to all authenticated
	// users (scoped by role inside the use case).
	theses := v1.Group("/theses", r.authMid.Authenticate())
	{
		theses.GET("", r.thesisHandler.ListTheses)
		theses.GET("/:thesis_id", r.thesisHandler.GetThesis)
	}

	thesesStudent := v1.Group("/theses",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleMahasiswa),
	)
	{
		thesesStudent.POST("", r.thesisHandler.CreateThesis)
	}

	thesesKaprodi := v1.Group("/theses",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleKaprodi),
	)
	{
		thesesKaprodi.PUT("/:thesis_id/review", r.thesisHandler.ReviewThesis)
		thesesKaprodi.PUT("/:thesis_id/assign-supervisor", r.thesisHandler.AssignSupervisor)
	}

	thesesManage := v1.Group("/theses",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleAdminFakultas, middleware.RoleKaprodi),
	)
	{
		thesesManage.PATCH("/:thesis_id/cancel", r.thesisHandler.CancelThesis)
	}

	// ── Consultations (Job 06) ───────────────────────────────────────────
	// POST create: Mahasiswa pemilik + Dosen Pembimbing (checked in use case)
	// GET list/detail/summary: Mahasiswa pemilik + Dosen Pembimbing + Admin + Kaprodi
	//   (ownership checked inside the use case via ThesisAccess).
	consultations := v1.Group("/theses/:thesis_id/consultations", r.authMid.Authenticate())
	{
		consultations.POST("", r.consultationHandler.Create)
		consultations.GET("", r.consultationHandler.List)
		consultations.GET("/summary", r.consultationHandler.Summary)
		consultations.GET("/:id", r.consultationHandler.Get)
		consultations.PUT("/:id", r.consultationHandler.Update)
		consultations.PATCH("/:id/approve", r.consultationHandler.Approve)
		consultations.DELETE("/:id", r.consultationHandler.Delete)
	}

	// ── Documents (Job 07) ──────────────────────────────────────────────
	// Upload: Mahasiswa pemilik only (checked in use case)
	// List/detail/download/history: owner + supervisor + examiner + admin + kaprodi
	documents := v1.Group("/theses/:thesis_id/documents", r.authMid.Authenticate())
	{
		documents.POST("", r.documentHandler.Upload)
		documents.GET("", r.documentHandler.List)
		documents.GET("/history", r.documentHandler.History)
		documents.GET("/:id", r.documentHandler.Get)
		documents.GET("/:id/download", r.documentHandler.Download)
	}

	// PATCH /documents/:id/review — Dosen Pembimbing thesis terkait only
	documentReview := v1.Group("/documents", r.authMid.Authenticate())
	{
		documentReview.PATCH("/:id/review", r.documentHandler.Review)
	}

	// ── Seminars (Job 08) ───────────────────────────────────────────────
	// POST /theses/:thesis_id/seminars — Mahasiswa pemilik only (checked in use case)
	// GET /seminars — role-scoped (checked in use case)
	seminars := v1.Group("/seminars", r.authMid.Authenticate())
	{
		seminars.GET("", r.seminarHandler.List)
		seminars.GET("/:id", r.seminarHandler.Get)
		seminars.GET("/:id/result", r.seminarHandler.Result)
	}

	seminarsManage := v1.Group("/seminars",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleAdminFakultas, middleware.RoleKaprodi),
	)
	{
		seminarsManage.PUT("/:id/schedule", r.seminarHandler.Schedule)
		seminarsManage.PUT("/:id/revision", r.seminarHandler.SetRevisionNotes)
	}

	seminarScores := v1.Group("/seminars", r.authMid.Authenticate())
	{
		seminarScores.POST("/:id/scores", r.seminarHandler.SubmitScores)
	}

	seminarSubmit := v1.Group("/theses/:thesis_id/seminars", r.authMid.Authenticate())
	{
		seminarSubmit.POST("", r.seminarHandler.Submit)
	}

	// ── Defenses (Job 09) ───────────────────────────────────────────────
	// POST /theses/:thesis_id/defenses — Mahasiswa pemilik only (checked in use case)
	// GET /defenses — role-scoped (checked in use case)
	defenses := v1.Group("/defenses", r.authMid.Authenticate())
	{
		defenses.GET("", r.defenseHandler.List)
		defenses.GET("/:id", r.defenseHandler.Get)
		defenses.GET("/:id/result", r.defenseHandler.Result)
	}

	defensesManage := v1.Group("/defenses",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleAdminFakultas, middleware.RoleKaprodi),
	)
	{
		defensesManage.PUT("/:id/schedule", r.defenseHandler.Schedule)
		defensesManage.PUT("/:id/revision", r.defenseHandler.SetRevisionNotes)
	}

	defenseScores := v1.Group("/defenses", r.authMid.Authenticate())
	{
		defenseScores.POST("/:id/scores", r.defenseHandler.SubmitScores)
	}

	defenseSubmit := v1.Group("/theses/:thesis_id/defenses", r.authMid.Authenticate())
	{
		defenseSubmit.POST("", r.defenseHandler.Submit)
	}

	// Yudisium — Kaprodi only
	graduation := v1.Group("/theses/:thesis_id/graduation",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleKaprodi),
	)
	{
		graduation.PUT("", r.defenseHandler.Graduate)
	}

	// Upcoming schedules — Admin + Kaprodi (used by the operational dashboard)
	upcoming := v1.Group("/schedules/upcoming",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleAdminFakultas, middleware.RoleKaprodi),
	)
	{
		upcoming.GET("", r.defenseHandler.Upcoming)
	}

	// ── Archives (Job 10) ──────────────────────────────────────────────
	// POST /theses/:thesis_id/archive — Mahasiswa pemilik + Admin (checked in use case)
	// GET /archives — all authenticated users (full-text search + filters)
	archives := v1.Group("/archives", r.authMid.Authenticate())
	{
		archives.GET("", r.archiveHandler.Search)
		archives.GET("/:id", r.archiveHandler.Get)
		archives.GET("/:id/download", r.archiveHandler.Download)
	}

	// Archives stats — Admin + Kaprodi only (separate group so the route is guarded)
	archiveStats := v1.Group("/archives",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleAdminFakultas, middleware.RoleKaprodi),
	)
	{
		archiveStats.GET("/stats", r.archiveHandler.Stats)
	}

	archiveCreate := v1.Group("/theses/:thesis_id/archive", r.authMid.Authenticate())
	{
		archiveCreate.POST("", r.archiveHandler.Create)
		archiveCreate.GET("", r.archiveHandler.GetByThesis)
	}

	// ── Lecturers (load balancing hints) ─────────────────────────────────
	lecturers := v1.Group("/lecturers",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleKaprodi, middleware.RoleAdminFakultas),
	)
	{
		lecturers.GET("", r.thesisHandler.ListLecturers)
	}

	// ── Dashboard (Job 12) ──────────────────────────────────────────────
	// Admin + Kaprodi: summary, lecturer analytics, operational
	dashboardManage := v1.Group("/dashboard",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleAdminFakultas, middleware.RoleKaprodi),
	)
	{
		dashboardManage.GET("/summary", r.dashboardHandler.Summary)
		dashboardManage.GET("/lecturer-analytics", r.dashboardHandler.LecturerAnalytics)
		dashboardManage.GET("/operational", r.dashboardHandler.Operational)
	}

	// Mahasiswa: own progress
	dashboardStudent := v1.Group("/dashboard",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleMahasiswa),
	)
	{
		dashboardStudent.GET("/student", r.dashboardHandler.Student)
	}

	// Dosen Pembimbing: supervised students
	dashboardSupervisor := v1.Group("/dashboard",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleDosenPembimbing),
	)
	{
		dashboardSupervisor.GET("/supervisor", r.dashboardHandler.Supervisor)
	}

	// Dosen Penguji: assignments + scores
	dashboardExaminer := v1.Group("/dashboard",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleDosenPenguji),
	)
	{
		dashboardExaminer.GET("/examiner", r.dashboardHandler.Examiner)
	}

	// ── Internal endpoints (Job 11, development only) ───────────────────
	// POST /api/v1/internal/test-email — sends a diagnostic test email.
	if r.cfg.AppEnv == "development" {
		internal := v1.Group("/internal")
		{
			internal.POST("/test-email", r.internalHandler.TestEmail)
		}
	}
}
