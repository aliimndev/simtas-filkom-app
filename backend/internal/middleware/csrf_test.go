package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestCSRFMiddlewareExemptPaths — exempt state-changing routes pass without a
// token; non-exempt ones are still rejected (403).
func TestCSRFMiddlewareExemptPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	eng := gin.New()
	eng.Use(CSRFMiddleware(false, "/api/v1/auth/login"))
	eng.POST("/api/v1/auth/login", func(c *gin.Context) { c.Status(http.StatusOK) })
	eng.POST("/api/v1/other", func(c *gin.Context) { c.Status(http.StatusOK) })

	do := func(path string) int {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, nil)
		eng.ServeHTTP(w, req)
		return w.Code
	}

	if code := do("/api/v1/auth/login"); code != http.StatusOK {
		t.Errorf("exempt login POST: got %d, want 200", code)
	}
	if code := do("/api/v1/other"); code != http.StatusForbidden {
		t.Errorf("non-exempt POST: got %d, want 403", code)
	}
}
