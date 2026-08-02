package handler

import (
	"net/http"
	"time"

	"github.com/aliimndev/simtas-filkom-app/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// serverStartAt records when the process started so /health can report uptime
// (Job 27 — Monitoring & Logging).
var serverStartAt = time.Now()

func RegisterHealthRoutes(rg *gin.RouterGroup, db *gorm.DB) {
	rg.GET("/health", func(c *gin.Context) { HealthCheck(c, db) })
}

// HealthCheck godoc
// @Summary      Health check
// @Description  Mengecek status layanan dan koneksi database. Menghasilkan 503 saat database tidak dapat dijangkau.
// @Tags         Health
// @Produce      json
// @Success      200  {object}  response.APIResponse "Layanan sehat"
// @Failure      503  {object}  response.APIResponse "Database tidak dapat dijangkau"
// @Router       /health [get]
func HealthCheck(c *gin.Context, db *gorm.DB) {
	status := "ok"
	httpStatus := http.StatusOK
	databaseStatus := "ok"

	if db == nil {
		databaseStatus = "unavailable"
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	} else if sqlDB, err := db.DB(); err != nil {
		databaseStatus = "error"
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	} else if err := sqlDB.PingContext(c.Request.Context()); err != nil {
		databaseStatus = "error"
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	response.Success(c, httpStatus, gin.H{
		"status":         status,
		"version":        "1.0.0",
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"uptime_seconds": int64(time.Since(serverStartAt).Seconds()),
		"database":       databaseStatus,
	})
}
