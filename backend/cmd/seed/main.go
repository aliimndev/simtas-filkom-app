package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/config"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/database"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/utils"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

// cmd/seed — Production seed (Job 29).
//
// Menyiapkan akun production dengan password kuat acak dan must_change_password
// = true, sehingga tidak ada password default permanen. Idempotent: akun yang
// sudah ada tidak akan ditimpa.
//
// Usage:
//
//	# DEVELOPMENT / UAT: gunakan --dev untuk password yang bisa diingat
//	go run ./cmd/seed
//
//	# PRODUCTION: password acak 16 karakter, dicetak sekali ke stdout
//	EMAIL_ADMIN=admin@filkom.unida.ac.id \
//	EMAIL_KAPRODI=kaprodi@filkom.unida.ac.id \
//	go run ./cmd/seed --env=production
//
// Catatan keamanan: di production, pastikan output password tidak tertinggal
// di shell history atau log. Sebaiknya redirect ke file sementara lalu hapus.

const (
	roleAdmin   = 1 // admin_fakultas
	roleKaprodi = 2 // kaprodi
)

func main() {
	dev := flag.Bool("dev", false, "development mode: gunakan password sederhana")
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
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}
	defer sqlDB.Close()

	adminEmail := getEnv("EMAIL_ADMIN", "admin@filkom.unida.ac.id")
	kaprodiEmail := getEnv("EMAIL_KAPRODI", "kaprodi@filkom.unida.ac.id")

	// ── Create accounts ───────────────────────────────────────────────────
	seedUser(db, "Administrator FILKOM", adminEmail, roleAdmin, *dev)
	seedUser(db, "Kepala Program Studi FILKOM", kaprodiEmail, roleKaprodi, *dev)

	log.Println("seed selesai. Akun yang baru dibuat dicetak di atas dengan password sementara (must_change_password=true).")
}

func seedUser(db *gorm.DB, fullName, email string, roleID int, dev bool) {
	var count int64
	if err := db.Model(&entity.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		log.Fatalf("failed to check user %s: %v", email, err)
	}
	if count > 0 {
		log.Printf("akun sudah ada, dilewati: %s", email)
		return
	}

	password := utils.GenerateRandomPassword(16)
	if dev {
		password = "SIMTAS@dev123"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to hash password for %s: %v", email, err)
	}

	user := entity.User{
		Email:              email,
		PasswordHash:       string(hash),
		FullName:           fullName,
		RoleID:             roleID,
		IsActive:           true,
		MustChangePassword: true,
	}
	if err := db.Create(&user).Error; err != nil {
		log.Fatalf("failed to create user %s: %v", email, err)
	}

	// Cetak password sementara SEKALI ke stdout — jangan di-log ke file.
	fmt.Printf("\n[AKUN BARU] %s\n  email:    %s\n  password: %s\n  (wajib ganti password saat login pertama)\n\n", fullName, email, password)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
