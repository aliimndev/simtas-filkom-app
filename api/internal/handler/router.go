package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/api/internal/middleware"
	"github.com/aliimndev/simtas-filkom-app/api/internal/repository"
	"github.com/aliimndev/simtas-filkom-app/api/internal/usecase"
	"github.com/aliimndev/simtas-filkom-app/api/pkg/config"
	"github.com/aliimndev/simtas-filkom-app/api/pkg/jwt"
)

// Router wires all dependencies and registers routes on the Gin engine
type Router struct {
	engine      *gin.Engine
	cfg         *config.Config
	authHandler *AuthHandler
	authMid     *middleware.AuthMiddleware
}

func NewRouter(engine *gin.Engine, db *gorm.DB, cfg *config.Config) *Router {
	jwtManager := jwt.NewJWTManager(
		cfg.JWTSecret,
		cfg.JWTExpiry,
		cfg.JWTRefreshExpiry,
	)

	authRepository := repository.NewAuthRepository(db)
	authUseCase := usecase.NewAuthUseCase(authRepository, jwtManager)
	authHandler := NewAuthHandler(authUseCase)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager, authRepository)

	return &Router{
		engine:      engine,
		cfg:         cfg,
		authHandler: authHandler,
		authMid:     authMiddleware,
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

	// ── Admin-only routes ─────────────────────────────────────────────────
	// Endpoints added in later jobs (user management, etc.)
	_ = v1.Group("/admin",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleAdminFakultas),
	)

	// ── Admin + Kaprodi routes ────────────────────────────────────────────
	// Dashboard, thesis approval, scheduling — added in later jobs
	_ = v1.Group("/management",
		r.authMid.Authenticate(),
		middleware.RequireRole(middleware.RoleAdminFakultas, middleware.RoleKaprodi),
	)
}
