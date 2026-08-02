package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// newTestDB opens a sqlmock-backed *gorm.DB. With Config.Conn provided the
// dialector reuses our connection pool, and DisableAutomaticPing stops gorm
// from pinging during Open — so Open succeeds without a real database and the
// handler's explicit PingContext is what exercises the health contract (Job 27).
func newTestDB(t *testing.T, expectPing bool) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	if expectPing {
		mock.ExpectPing()
	}

	db, err := gorm.Open(
		postgres.New(postgres.Config{Conn: sqlDB}),
		&gorm.Config{DisableAutomaticPing: true},
	)
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}
	return db, mock
}

func newHealthTestRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterHealthRoutes(r.Group("/api/v1"), db)
	return r
}

// TestHealthCheck_DatabaseDown — Job 27: database ping gagal → HTTP 503 agar
// UptimeRobot (dan alerting lain) mendeteksi outage.
func TestHealthCheck_DatabaseDown(t *testing.T) {
	// No ExpectPing → PingContext fails → 503.
	db, _ := newTestDB(t, false)
	r := newHealthTestRouter(db)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when database is down, got %d", w.Code)
	}
	if body := w.Body.String(); body == "" {
		t.Fatal("expected a JSON response body")
	}
}

// TestHealthCheck_DatabaseUp — database ping sukses → HTTP 200 dengan status ok.
func TestHealthCheck_DatabaseUp(t *testing.T) {
	db, mock := newTestDB(t, true)
	r := newHealthTestRouter(db)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 when database is up, got %d", w.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("ping expectation was not consumed: %v", err)
	}
}

// TestHealthCheck_NilDatabase — defensive path: nil db juga 503.
func TestHealthCheck_NilDatabase(t *testing.T) {
	r := newHealthTestRouter(nil)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/health", nil))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 for nil database, got %d", w.Code)
	}
}
