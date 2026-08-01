package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"github.com/aliimndev/simtas-filkom-app/backend/pkg/response"
)

// Role name constants — keep in sync with seed data roles table
const (
	RoleAdminFakultas   = "admin_fakultas"
	RoleKaprodi         = "kaprodi"
	RoleMahasiswa       = "mahasiswa"
	RoleDosenPembimbing = "dosen_pembimbing"
	RoleDosenPenguji    = "dosen_penguji"
)

// RequireRole returns a middleware that only allows requests from users with
// one of the specified roles. Must be used after Authenticate().
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			response.Error(c, http.StatusForbidden, "Akses ditolak: role tidak ditemukan", nil)
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok || !slices.Contains(allowedRoles, role) {
			response.Error(c, http.StatusForbidden, "Akses ditolak: role tidak diizinkan", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
