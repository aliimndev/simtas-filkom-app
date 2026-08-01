package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/handler"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/config"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/database"
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

	// ── Start server ──────────────────────────────────────────────────────
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("server starting on %s (env=%s)", addr, cfg.AppEnv)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
