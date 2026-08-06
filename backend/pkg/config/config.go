package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppPort string
	AppEnv  string

	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBMaxOpenConns   int
	DBMaxIdleConns   int
	DBConnMaxLifetime time.Duration

	JWTSecret        string
	JWTExpiry        time.Duration
	JWTRefreshExpiry time.Duration

	// Storage: either supabase or s3 (MinIO, R2, etc.)
	SupabaseURL string
	// SupabaseKey (anon) is reserved for future Supabase client init; the
	// storage service (Job 21) authenticates with SupabaseServiceRoleKey.
	SupabaseKey             string
	SupabaseServiceRoleKey  string
	SupabaseDocumentsBucket string
	SupabaseArchivesBucket  string
	StorageProvider         string

	// S3-compatible storage (MinIO, R2, etc.)
	S3Endpoint        string
	S3Region          string
	S3AccessKey       string
	S3SecretKey       string
	S3DocumentsBucket string
	S3ArchivesBucket  string

	ResendAPIKey  string
	EmailFrom     string
	EmailFromName string
	EmailDevMode  bool
	FrontendURL   string

	CORSAllowedOrigins string
}

// defaultSecrets lists JWT secrets that must NOT be used in production.
var defaultSecrets = map[string]bool{
	"your-super-secret-key": true,
	"secret":                true,
	"changeme":              true,
}

// Validate checks critical configuration values and panics on invalid
// combinations that would leave the server in an insecure state.
// Call this immediately after Load() in main().
func (c *Config) Validate() {
	if c.AppEnv == "production" {
		if defaultSecrets[c.JWTSecret] {
			panic("FATAL: JWT_SECRET must be set to a strong, unique value in production. " +
				"The default value is insecure and would allow token forgery.")
		}
		if len(c.JWTSecret) < 32 {
			panic("FATAL: JWT_SECRET must be at least 32 characters in production.")
		}
		if c.DBPassword == "postgres" {
			panic("FATAL: DB_PASSWORD must be changed from the default 'postgres' in production.")
		}
		if c.CORSAllowedOrigins == "http://localhost:3000" {
			panic("FATAL: CORS_ALLOWED_ORIGINS must be set to your production domain, not localhost.")
		}
	}
}

func Load() *Config {
	return &Config{
		AppPort: getEnv("APP_PORT", "8080"),
		AppEnv:  getEnv("APP_ENV", "development"),

		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "postgres"),
		DBPassword:       getEnv("DB_PASSWORD", "postgres"),
		DBName:           getEnv("DB_NAME", "simtas_filkom"),
		DBMaxOpenConns:   getInt("DB_MAX_OPEN_CONNS", 100),
		DBMaxIdleConns:   getInt("DB_MAX_IDLE_CONNS", 10),
		DBConnMaxLifetime: getDuration("DB_CONN_MAX_LIFETIME", 1*time.Hour),

		JWTSecret:        getEnv("JWT_SECRET", "your-super-secret-key"),
		JWTExpiry:        getDuration("JWT_EXPIRY", 24*time.Hour),
		JWTRefreshExpiry: getDuration("JWT_REFRESH_EXPIRY", 168*time.Hour),

		SupabaseURL:            getEnv("SUPABASE_URL", ""),
		SupabaseKey:            getEnv("SUPABASE_KEY", ""),
		SupabaseServiceRoleKey: getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
		// SUPABASE_BUCKET is kept as a legacy fallback so existing .env files
		// keep working after Job 21 introduced per-purpose bucket names.
		SupabaseDocumentsBucket: getEnv("SUPABASE_DOCUMENTS_BUCKET", getEnv("SUPABASE_BUCKET", "simtas-documents")),
		SupabaseArchivesBucket:  getEnv("SUPABASE_ARCHIVES_BUCKET", "simtas-archives"),
		StorageProvider:         getEnv("STORAGE_PROVIDER", "local"),

		// S3-compatible storage (MinIO, R2, etc.)
		S3Endpoint:        getEnv("S3_ENDPOINT", ""),
		S3Region:          getEnv("S3_REGION", "us-east-1"),
		S3AccessKey:       getEnv("S3_ACCESS_KEY", ""),
		S3SecretKey:       getEnv("S3_SECRET_KEY", ""),
		S3DocumentsBucket: getEnv("S3_DOCUMENTS_BUCKET", "simtas-documents"),
		S3ArchivesBucket:  getEnv("S3_ARCHIVES_BUCKET", "simtas-archives"),

		ResendAPIKey:  getEnv("RESEND_API_KEY", ""),
		EmailFrom:     getEnv("EMAIL_FROM", "noreply@filkom.unida.ac.id"),
		EmailFromName: getEnv("EMAIL_FROM_NAME", "SIMTAS FILKOM"),
		EmailDevMode:  getBool("EMAIL_DEV_MODE", false),
		FrontendURL:   getEnv("FRONTEND_URL", "http://localhost:3000"),

		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err == nil {
			return d
		}
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			return b
		}
	}
	return fallback
}
