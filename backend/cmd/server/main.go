package main

// SIMTAS FILKOM API (Job 22)
//
//	@title           SIMTAS FILKOM API
//	@version         1.0
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
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/aliimndev/simtas-filkom-app/backend/docs" // generated OpenAPI docs
	"github.com/aliimndev/simtas-filkom-app/backend/internal/handler"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/config"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/database"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/logger"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/scheduler"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

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

	// ── Structured logging (Job 13) ──────────────────────────────────────
	logger.Init(cfg.AppEnv)

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

	// ── Start server ──────────────────────────────────────────────────────
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("server starting on %s (env=%s)", addr, cfg.AppEnv)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
