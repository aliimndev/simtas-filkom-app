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
	"gorm.io/gorm"
)

// RunMigrations runs all pending UP migrations from the migrations/ directory.
// migrationsPath should be the absolute (or relative-to-cwd) path to the
// folder that contains the numbered *.up.sql / *.down.sql files.
func RunMigrations(db *gorm.DB, migrationsPath string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB: %w", err)
	}

	m, err := newMigrator(sqlDB, migrationsPath)
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

// RollbackAll rolls back every applied migration (for testing / teardown).
func RollbackAll(db *gorm.DB, migrationsPath string) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get underlying sql.DB: %w", err)
	}

	m, err := newMigrator(sqlDB, migrationsPath)
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

// RunSeeds executes all SQL files inside seedsPath in lexicographic order.
// Each file is run inside its own transaction; errors abort that file only.
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

func newMigrator(sqlDB *sql.DB, migrationsPath string) (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("create postgres driver: %w", err)
	}

	// filepath.ToSlash ensures forward slashes on Windows too.
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
