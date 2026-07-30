package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/aliimndev/simtas-filkom-app/api/internal/handler"
	"github.com/aliimndev/simtas-filkom-app/api/pkg/config"
	"github.com/aliimndev/simtas-filkom-app/api/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// ── flags ────────────────────────────────────────────────────────────
	doMigrate := flag.Bool("migrate", false, "run database migrations then exit")
	doSeed := flag.Bool("seed", false, "run seed files then exit (requires --migrate first)")
	flag.Parse()

	// ── environment ──────────────────────────────────────────────────────
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system environment variables")
	}

	cfg := config.Load()

	// ── database ─────────────────────────────────────────────────────────
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close(db)

	// Resolve paths relative to the project root (works from any cwd).
	_, thisFile, _, _ := runtime.Caller(0)
	// thisFile = …/api/cmd/server/main.go  →  project root = …/api
	apiRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	migrationsPath := filepath.Join(apiRoot, "migrations")
	seedsPath := filepath.Join(apiRoot, "migrations", "seeds")

	// ── --migrate flag ───────────────────────────────────────────────────
	if *doMigrate {
		log.Println("running migrations…")
		if err := database.RunMigrations(db, migrationsPath); err != nil {
			log.Fatalf("migrations failed: %v", err)
		}
		log.Println("migrations complete")
		os.Exit(0)
	}

	// ── --seed flag ──────────────────────────────────────────────────────
	if *doSeed {
		log.Println("running seeds…")
		if err := database.RunSeeds(db, seedsPath); err != nil {
			log.Fatalf("seeds failed: %v", err)
		}
		log.Println("seeds complete")
		os.Exit(0)
	}

	// ── auto-migrate on startup (development convenience) ────────────────
	if cfg.AppEnv == "development" {
		log.Println("auto-running migrations (development mode)…")
		if err := database.RunMigrations(db, migrationsPath); err != nil {
			log.Fatalf("auto-migration failed: %v", err)
		}
	}

	// ── router ───────────────────────────────────────────────────────────
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", cfg.CORSAllowedOrigins)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Routes
	v1 := r.Group("/api/v1")
	{
		handler.RegisterHealthRoutes(v1)
	}

	// ── start ────────────────────────────────────────────────────────────
	port := cfg.AppPort
	if port == "" {
		port = "8080"
	}
	log.Printf("server starting on port %s (env=%s)", port, cfg.AppEnv)
	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("server error: %v", err)
		os.Exit(1)
	}
}
