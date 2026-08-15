package main

// SIMTAS FILKOM API (Job 22)
//
//	@title           SIMTAS FILKOM API
//	@version         1.1.0
//	@description     API untuk Sistem Manajemen Tugas Akhir dan Skripsi Fakultas Ilmu Komputer Universitas Djuanda.
//	@termsOfService  https://filkom.unida.ac.id
//
//	@contact.name   Admin SIMTAS FILKOM
//	@contact.email  admin@filkom.unida.ac.id
//
//	@license.name  MIT
//	@license.url   https://opensource.org/licenses/MIT
//
//	@host      localhost:8080
//	@BasePath  /api/v1
//
//	@securityDefinitions.apikey BearerAuth
//	@in header
//	@name Authorization
//	@description Masukkan token dengan format: Bearer {token}
import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/aliimndev/simtas-filkom-app/backend/docs" // generated OpenAPI docs
	"github.com/aliimndev/simtas-filkom-app/backend/internal/handler"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/config"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/database"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/logger"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/scheduler"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// version is injected at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	// ── CLI flags ─────────────────────────────────────────────────────────
	doMigrate := flag.Bool("migrate", false, "run database migrations then exit")
	doSeed := flag.Bool("seed", false, "run seed files then exit")
	flag.Parse()

	// ── Environment ───────────────────────────────────────────────────────
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system environment variables")
	}

	cfg := config.Load()
	cfg.Validate() // Panic on insecure production configuration

	// ── Structured logging (Job 13) ──────────────────────────────────────
	logger.Init(cfg.AppEnv)

	// ── Sentry error tracking (Job 30) — no-op when SENTRY_DSN is unset ─
	if cfg.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			Environment:      cfg.AppEnv,
			Release:          version,
			TracesSampleRate: 0.1,
		}); err != nil {
			log.Printf("warn: sentry init failed: %v", err)
		} else {
			defer sentry.Flush(2 * time.Second)
		}
	}

	// ── Database ──────────────────────────────────────────────────────────
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	// Note: db connection is kept open for the lifetime of the server process.
	// database.Close(db) is intentionally not deferred here.

	// Resolve paths relative to the api root (works regardless of cwd)
	_, thisFile, _, _ := runtime.Caller(0)
	apiRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	migrationsPath := filepath.Join(apiRoot, "migrations")
	seedsPath := filepath.Join(apiRoot, "migrations", "seeds")

	// ── --migrate flag ────────────────────────────────────────────────────
	if *doMigrate {
		log.Println("running migrations…")
		if err := database.RunMigrations(db, migrationsPath); err != nil {
			log.Fatalf("migrations failed: %v", err)
		}
		log.Println("migrations complete")
		os.Exit(0)
	}

	// ── --seed flag ───────────────────────────────────────────────────────
	if *doSeed {
		log.Println("running seeds…")
		if err := database.RunSeeds(db, seedsPath); err != nil {
			log.Fatalf("seeds failed: %v", err)
		}
		log.Println("seeds complete")
		os.Exit(0)
	}

	// ── Auto-migrate in development ───────────────────────────────────────
	if cfg.AppEnv == "development" {
		if err := database.RunMigrations(db, migrationsPath); err != nil {
			log.Fatalf("auto-migration failed: %v", err)
		}
	}

	// ── Gin engine ────────────────────────────────────────────────────────
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.Default()

	// Trusted-proxy configuration: by default no proxy is trusted, so a client
	// cannot spoof X-Forwarded-For to bypass the per-IP rate limits (the login
	// endpoint is limited to 10 req/min/IP). When deploying behind nginx or
	// another reverse proxy, set TRUSTED_PROXIES to its CIDRs so real client
	// IPs are resolved (and rate-limited) correctly.
	if len(cfg.TrustedProxies) > 0 {
		if err := engine.SetTrustedProxies(cfg.TrustedProxies); err != nil {
			log.Fatalf("invalid TRUSTED_PROXIES: %v", err)
		}
	} else {
		engine.SetTrustedProxies(nil)
	}

	// Cap multipart form memory to the configured request body limit. File
	// uploads are additionally bounded by the document layer's own 10 MB check.
	engine.MaxMultipartMemory = int64(cfg.MaxRequestBodyBytes)

	// Register all routes (auth + middleware)
	r := handler.NewRouter(engine, db, cfg)
	r.Setup()

	// ── Swagger UI (Job 22, non-production only) ─────────────────────────
	if cfg.AppEnv != "production" {
		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// ── Token blacklist cleanup (Job 13) ─────────────────────────────────
	stopCleanup := scheduler.StartTokenCleanup(db)
	defer stopCleanup()

	// ── Durable email retry (Job 24) ─────────────────────────────────────
	// Re-enqueues email_logs rows stuck as "queued" (crashed mid-delivery) or
	// "failed" so sends survive restarts and transient provider outages.
	stopEmailRetry := scheduler.StartEmailRetry(db, r.EmailService())
	defer stopEmailRetry()

	// ── Start server (production-ready with timeouts + graceful shutdown) ─
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	srv := &http.Server{
		Handler:           engine,
		Addr:              addr,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MiB
	}

	go func() {
		log.Printf("server starting on %s (env=%s)", addr, cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}

	// Drain queued audit-log entries before closing the DB connection.
	r.Shutdown()

	log.Println("server stopped")
}
