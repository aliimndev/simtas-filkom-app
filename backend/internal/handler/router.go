package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/middleware"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/repository"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/usecase"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/audit"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/config"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/email"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/jwt"
)

// Router wires all dependencies and registers routes on the Gin engine
type Router struct {
	engine              *gin.Engine
	cfg                 *config.Config
	authHandler         *AuthHandler
	userHandler         *UserHandler
	academicYearHandler *AcademicYearHandler
	thesisHandler       *ThesisHandler
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
	auditRepository := repository.NewAuditRepository(db)

	auditService := audit.NewAuditService(auditRepository)
	emailService := email.NewStubEmailService(db)

	authUseCase := usecase.NewAuthUseCase(authRepository, jwtManager)
	userUseCase := usecase.NewUserUseCase(userRepository, emailService, auditService)
	academicYearUseCase := usecase.NewAcademicYearUseCase(academicYearRepository)
	thesisUseCase := usecase.NewThesisUseCase(thesisRepository, userRepository, academicYearRepository, emailService, auditService)

	authHandler := NewAuthHandler(authUseCase)
	userHandler := NewUserHandler(userUseCase)
	academicYearHandler := NewAcademicYearHandler(academicYearUseCase)
	thesisHandler := NewThesisHandler(thesisUseCase)

	authMiddleware := middleware.NewAuthMiddleware(jwtManager, authRepository, authRepository)

	return &Router{
		engine:              engine,
		cfg:                 cfg,
		authHandler:         authHandler,
		userHandler:         userHandler,
		academicYearHandler: academicYearHandler,
		thesisHandler:       thesisHandler,
		authMid:             authMiddleware,
	}
}

func (r *Router) Setup() {
	// ── Global middleware ─────────────────────────────────────────────────
	r.engine.Use(middleware.SecurityHeadersMiddleware())
	r.engine.Use(middleware.CORSMiddleware(r.cfg.CORSAllowedOrigins))

	// ── API v1 ────────────────────────────────────────────────────────────
	v1 := r.engine.Group("/api/v1")

	// Health check (public, unauthenticated)
	RegisterHealthRoutes(v1)

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
	}

	// ── Theses ────────────────────────────────────────────────────────────
	// POST /theses is mahasiswa-only; GET is available to all authenticated
	// users (scoped by role inside the use case).
	theses := v1.Group("/theses", r.authMid.Authenticate())
	{
		theses.GET("", r.thesisHandler.ListTheses)
		theses.GET("/:id", r.thesisHandler.GetThesis)
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
		thesesKaprodi.PUT("/:id/review", r.thesisHandler.ReviewThesis)
		thesesKaprodi.PUT("/:id/assign-supervisor", r.thesisHandler.AssignSupervisor)
	}

	thesesManage := v1.Group("/theses",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleAdminFakultas, middleware.RoleKaprodi),
	)
	{
		thesesManage.PATCH("/:id/cancel", r.thesisHandler.CancelThesis)
	}

	// ── Lecturers (load balancing hints) ─────────────────────────────────
	lecturers := v1.Group("/lecturers",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleKaprodi, middleware.RoleAdminFakultas),
	)
	{
		lecturers.GET("", r.thesisHandler.ListLecturers)
	}
}
