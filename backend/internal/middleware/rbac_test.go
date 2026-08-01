package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/aliimndev/simtas-filkom-app/backend/pkg/jwt"
)

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		role         string
		allowed      []string
		expectStatus int
	}{
		{"allowed role passes", RoleAdminFakultas, []string{RoleAdminFakultas}, http.StatusOK},
		{"multiple allowed roles", RoleKaprodi, []string{RoleAdminFakultas, RoleKaprodi}, http.StatusOK},
		{"disallowed role forbidden", RoleMahasiswa, []string{RoleAdminFakultas}, http.StatusForbidden},
		{"missing role forbidden", "", []string{RoleAdminFakultas}, http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/", func(c *gin.Context) {
				if tt.role != "" {
					c.Set("userRole", tt.role)
				}
			}, RequireRole(tt.allowed...), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectStatus)
			}
		})
	}
}

type fakeBlacklist struct {
	blacklisted   map[string]bool
	tokenVersions map[string]int
}

func (f *fakeBlacklist) IsTokenBlacklisted(_ context.Context, jti string) (bool, error) {
	return f.blacklisted[jti], nil
}

func (f *fakeBlacklist) GetUserTokenVersion(_ context.Context, userID string) (int, error) {
	return f.tokenVersions[userID], nil
}

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := jwt.NewJWTManager("test-secret", time.Hour, time.Hour)
	userID := uuid.New()

	makeToken := func() string {
		token, _, err := manager.GenerateAccessToken(userID, RoleMahasiswa, "student@test.com", 0)
		if err != nil {
			t.Fatal(err)
		}
		return token
	}

	tests := []struct {
		name         string
		authHeader   string
		blacklist    *fakeBlacklist
		expectStatus int
	}{
		{"valid token", "Bearer " + makeToken(), &fakeBlacklist{}, http.StatusOK},
		{"missing header", "", &fakeBlacklist{}, http.StatusUnauthorized},
		{"bad scheme", "Token abc", &fakeBlacklist{}, http.StatusUnauthorized},
		{"garbage token", "Bearer not.a.jwt", &fakeBlacklist{}, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAuthMiddleware(manager, tt.blacklist, tt.blacklist)
			r := gin.New()
			r.GET("/", m.Authenticate(), func(c *gin.Context) {
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			r.ServeHTTP(w, req)

			if w.Code != tt.expectStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.expectStatus)
			}
		})
	}
}

// TestBlacklistedTokenRejected proves a token whose JTI is blacklisted is rejected.
func TestBlacklistedTokenRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := jwt.NewJWTManager("test-secret", time.Hour, time.Hour)
	token, jti, err := manager.GenerateAccessToken(uuid.New(), RoleMahasiswa, "s@t.com", 0)
	if err != nil {
		t.Fatal(err)
	}

	m := NewAuthMiddleware(manager, &fakeBlacklist{blacklisted: map[string]bool{jti: true}}, &fakeBlacklist{blacklisted: map[string]bool{jti: true}})
	r := gin.New()
	r.GET("/", m.Authenticate(), func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
