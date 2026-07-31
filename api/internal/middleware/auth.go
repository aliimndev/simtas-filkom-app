package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/aliimndev/simtas-filkom-app/api/pkg/jwt"
	"github.com/aliimndev/simtas-filkom-app/api/pkg/response"
)

// TokenBlacklistChecker abstracts the blacklist lookup to avoid importing
// the repository package directly.
type TokenBlacklistChecker interface {
	IsTokenBlacklisted(ctx context.Context, jti string) (bool, error)
}

type AuthMiddleware struct {
	jwtManager  *jwt.JWTManager
	blacklistFn TokenBlacklistChecker
}

func NewAuthMiddleware(jwtManager *jwt.JWTManager, blacklistFn TokenBlacklistChecker) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager, blacklistFn: blacklistFn}
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
		blacklisted, err := m.blacklistFn.IsTokenBlacklisted(c.Request.Context(), claims.JTI)
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

		// Inject claims into context
		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Set("userEmail", claims.Email)
		c.Set("tokenJTI", claims.JTI)
		c.Set("accessToken", tokenString)

		c.Next()
	}
}
