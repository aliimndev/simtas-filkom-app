package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"gorm.io/gorm"
)

// RunMigrations runs all pending UP migrations.
// Uses a separate sql.DB connection so closing the migrator does NOT affect
// the GORM connection pool.
func RunMigrations(db *gorm.DB, migrationsPath string) error {
	dsn, err := getDSN(db)
	if err != nil {
		return err
	}

	// Dedicated connection for migrations only
	migDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open migration db: %w", err)
	}
	defer migDB.Close()

	m, err := newMigrator(migDB, migrationsPath)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}

	version, dirty, _ := m.Version()
	log.Printf("database migrations OK — version %d (dirty=%v)", version, dirty)
	return nil
}

// RollbackAll rolls back every applied migration.
func RollbackAll(db *gorm.DB, migrationsPath string) error {
	dsn, err := getDSN(db)
	if err != nil {
		return err
	}

	migDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("open migration db: %w", err)
	}
	defer migDB.Close()

	m, err := newMigrator(migDB, migrationsPath)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate down: %w", err)
	}

	log.Println("database rolled back to version 0")
	return nil
}

// RunSeeds executes all SQL files in seedsPath in lexicographic order.
func RunSeeds(db *gorm.DB, seedsPath string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB: %w", err)
	}

	pattern := filepath.Join(seedsPath, "*.sql")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("glob seed files: %w", err)
	}
	if len(files) == 0 {
		log.Printf("no seed files found in %s", seedsPath)
		return nil
	}

	for _, f := range files {
		if err := runSQLFile(sqlDB, f); err != nil {
			return fmt.Errorf("seed file %s: %w", filepath.Base(f), err)
		}
		log.Printf("seed applied: %s", filepath.Base(f))
	}
	return nil
}

// ─── helpers ────────────────────────────────────────────────────────────────

// getDSN builds a postgres DSN from environment variables
func getDSN(_ *gorm.DB) (string, error) {
	return buildDSNFromEnv(), nil
}

// buildDSNFromEnv reads DB env vars to build a postgres DSN
func buildDSNFromEnv() string {
	host := getEnvOrDefault("DB_HOST", "localhost")
	port := getEnvOrDefault("DB_PORT", "5432")
	user := getEnvOrDefault("DB_USER", "postgres")
	password := getEnvOrDefault("DB_PASSWORD", "postgres")
	dbname := getEnvOrDefault("DB_NAME", "simtas_filkom")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Jakarta",
		host, port, user, password, dbname)
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func newMigrator(sqlDB *sql.DB, migrationsPath string) (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("create postgres driver: %w", err)
	}

	sourceURL := "file://" + filepath.ToSlash(migrationsPath)
	m, err := migrate.NewWithDatabaseInstance(sourceURL, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("create migrator: %w", err)
	}
	return m, nil
}

func runSQLFile(db *sql.DB, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	if _, err := tx.Exec(string(content)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("exec sql: %w", err)
	}

	return tx.Commit()
}
