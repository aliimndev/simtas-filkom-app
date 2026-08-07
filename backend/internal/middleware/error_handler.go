package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"

	"github.com/aliimndev/simtas-filkom-app/backend/pkg/logger"
)

// ErrorHandler recovers panics raised inside handlers, logs the stack trace,
// reports the panic to Sentry when configured, and returns a standardized
// 500 response so the server never crashes.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Get().Error("panic recovered",
					slog.Any("error", r),
					slog.String("request_id", RequestID(c)),
					slog.String("path", c.Request.URL.Path),
					slog.String("stack", string(debug.Stack())),
				)
				sentry.CurrentHub().RecoverWithContext(c.Request.Context(), r)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Terjadi kesalahan pada server",
				})
			}
		}()
		c.Next()
	}
}
