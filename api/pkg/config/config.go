package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppPort string
	AppEnv  string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	JWTSecret        string
	JWTExpiry        time.Duration
	JWTRefreshExpiry time.Duration

	SupabaseURL    string
	SupabaseKey    string
	SupabaseBucket string

	ResendAPIKey  string
	EmailFrom     string
	EmailFromName string

	CORSAllowedOrigins string
}

func Load() *Config {
	return &Config{
		AppPort: getEnv("APP_PORT", "8080"),
		AppEnv:  getEnv("APP_ENV", "development"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "simtas_filkom"),

		JWTSecret:        getEnv("JWT_SECRET", "your-super-secret-key"),
		JWTExpiry:        getDuration("JWT_EXPIRY", 24*time.Hour),
		JWTRefreshExpiry: getDuration("JWT_REFRESH_EXPIRY", 168*time.Hour),

		SupabaseURL:    getEnv("SUPABASE_URL", ""),
		SupabaseKey:    getEnv("SUPABASE_KEY", ""),
		SupabaseBucket: getEnv("SUPABASE_BUCKET", "simtas-documents"),

		ResendAPIKey:  getEnv("RESEND_API_KEY", ""),
		EmailFrom:     getEnv("EMAIL_FROM", "noreply@filkom.unida.ac.id"),
		EmailFromName: getEnv("EMAIL_FROM_NAME", "SIMTAS FILKOM"),

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