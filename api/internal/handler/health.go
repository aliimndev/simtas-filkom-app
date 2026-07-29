package handler

import (
	"time"

	"github.com/aliimndev/simtas-filkom-app/api/pkg/response"
	"github.com/gin-gonic/gin"
)

func RegisterHealthRoutes(rg *gin.RouterGroup) {
	rg.GET("/health", HealthCheck)
}

func HealthCheck(c *gin.Context) {
	response.Success(c, "Service is healthy", gin.H{
		"status":    "ok",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}