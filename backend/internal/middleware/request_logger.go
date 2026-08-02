package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/pkg/logger"
)

const requestIDKey = "requestID"

// RequestID returns the request ID attached by RequestLogger, if any.
func RequestID(c *gin.Context) string {
	if v, ok := c.Get(requestIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// RequestLogger logs every HTTP request with method, path, status, latency,
// client IP, and a request ID for tracing. The request body is never logged
// because it may contain sensitive data. The log line runs in a defer so a
// panicking request (recovered by the outer ErrorHandler) is still logged.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqID := uuid.NewString()
		c.Set(requestIDKey, reqID)
		c.Header("X-Request-ID", reqID)

		defer func() {
			logger.Get().Info("http_request",
				slog.String("request_id", reqID),
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
				slog.Int("status", c.Writer.Status()),
				slog.Duration("latency", time.Since(start)),
				slog.String("ip", c.ClientIP()),
				slog.String("user_agent", c.Request.UserAgent()),
			)
		}()

		c.Next()
	}
}
