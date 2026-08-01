package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/pkg/jwt"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/response"
)

// TokenBlacklistChecker abstracts the blacklist lookup to avoid importing
// the repository package directly.
type TokenBlacklistChecker interface {
	IsTokenBlacklisted(ctx context.Context, jti string) (bool, error)
}

// SessionVersionChecker provides the current token_version of a user so the
// middleware can reject tokens issued before an admin invalidated sessions.
type SessionVersionChecker interface {
	GetUserTokenVersion(ctx context.Context, userID string) (int, error)
}

type AuthMiddleware struct {
	jwtManager *jwt.JWTManager
	sessionSvc SessionVersionChecker
	blacklist  TokenBlacklistChecker
}

func NewAuthMiddleware(jwtManager *jwt.JWTManager, sessionSvc SessionVersionChecker, blacklist TokenBlacklistChecker) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager, sessionSvc: sessionSvc, blacklist: blacklist}
}

// Authenticate validates Bearer token and injects claims into Gin context
func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "Token tidak ditemukan", nil)
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, http.StatusUnauthorized, "Format token tidak valid. Gunakan: Bearer <token>", nil)
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err != nil {
			if err == jwt.ErrExpiredToken {
				response.Error(c, http.StatusUnauthorized, "Token sudah kadaluarsa", nil)
			} else {
				response.Error(c, http.StatusUnauthorized, "Token tidak valid", nil)
			}
			c.Abort()
			return
		}

		// Check blacklist
		if m.blacklist != nil {
			blacklisted, err := m.blacklist.IsTokenBlacklisted(c.Request.Context(), claims.JTI)
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server", nil)
				c.Abort()
				return
			}
			if blacklisted {
				response.Error(c, http.StatusUnauthorized, "Token sudah tidak valid", nil)
				c.Abort()
				return
			}
		}

		// Check session version: tokens issued before session invalidation are rejected,
		// and tokens of deleted/nonexistent users are rejected too.
		if m.sessionSvc != nil {
			currentVersion, err := m.sessionSvc.GetUserTokenVersion(c.Request.Context(), claims.UserID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					response.Error(c, http.StatusUnauthorized, "Sesi Anda telah berakhir, silakan login kembali", nil)
				} else {
					response.Error(c, http.StatusInternalServerError, "Terjadi kesalahan server", nil)
				}
				c.Abort()
				return
			}
			if currentVersion != claims.TokenVersion {
				response.Error(c, http.StatusUnauthorized, "Sesi Anda telah berakhir, silakan login kembali", nil)
				c.Abort()
				return
			}
		}

		// Inject claims into context
		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Set("userEmail", claims.Email)
		c.Set("tokenJTI", claims.JTI)
		c.Set("accessToken", tokenString)

		c.Next()
	}
}
