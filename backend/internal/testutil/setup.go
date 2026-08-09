//go:build integration

// Package testutil provides helpers for integration tests (build tag
// "integration") that exercise the full HTTP stack against a real PostgreSQL
// database. Run with: go test ./internal/handler/... -tags integration
package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/handler"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/config"
	"github.com/aliimndev/simtas-filkom-app/backend/pkg/database"
)

// testMigrationsPath resolves backend/migrations relative to this package
// (backend/internal/testutil → ../../migrations) so tests run from any cwd.
func testMigrationsPath() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return filepath.Join(dir, "..", "..", "migrations")
}

// SetupTestDB connects to the test database (DB_NAME defaults to
// simtas_filkom_test), runs migrations and seeds roles. If PostgreSQL is
// unreachable the test is skipped so the unit suite stays green without one.
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Point the connection at the dedicated test database unless the caller
	// explicitly set DB_NAME (e.g. a CI service named simtas_test).
	if os.Getenv("DB_NAME") == "" {
		_ = os.Setenv("DB_NAME", "simtas_filkom_test")
	}
	// config.Load() fails fast on a missing APP_ENV; tests run in development.
	_ = os.Setenv("APP_ENV", "development")

	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		// On CI the Postgres service is a hard requirement — a skip would let the
		// integration gate pass vacuously. Locally, skip so the suite still works
		// without a running database.
		if os.Getenv("CI") != "" {
			t.Fatalf("integration test requires PostgreSQL (DB_NAME=%s): %v", cfg.DBName, err)
		}
		t.Skipf("integration test requires PostgreSQL (DB_NAME=%s): %v", cfg.DBName, err)
	}
	t.Cleanup(func() { database.Close(db) })

	if err := database.RunMigrations(db, testMigrationsPath()); err != nil {
		t.Fatalf("migrations failed: %v", err)
	}
	seedRoles(t, db)

	// Fresh state per test: truncate every table between tests.
	t.Cleanup(func() { truncateAll(t, db) })
	return db
}

// SetupTestRouter builds the production router with the test DB so the
// integration tests exercise the real middleware + handlers.
func SetupTestRouter(t *testing.T, db *gorm.DB) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	// config.Load() fails fast on a missing APP_ENV; tests run in development.
	_ = os.Setenv("APP_ENV", "development")
	cfg := config.Load()
	cfg.JWTSecret = "integration-test-secret"
	engine := gin.New()
	r := handler.NewRouter(engine, db, cfg)
	r.Setup()
	// Shut down the audit-service worker goroutine before truncateAll runs.
	// Cleanups execute in LIFO order: SetupTestDB registers truncateAll and
	// database.Close BEFORE this function runs, so r.Shutdown() — registered
	// here — executes first, guaranteeing no orphaned worker goroutine INSERTs
	// into audit_logs while or after TRUNCATE locks the tables.
	t.Cleanup(func() { r.Shutdown() })
	return engine
}

// ── seeding helpers ───────────────────────────────────────────────────────

func seedRoles(t *testing.T, db *gorm.DB) {
	t.Helper()
	roles := []entity.Role{
		{ID: 1, Name: "admin_fakultas"},
		{ID: 2, Name: "kaprodi"},
		{ID: 3, Name: "mahasiswa"},
		{ID: 4, Name: "dosen_pembimbing"},
		{ID: 5, Name: "dosen_penguji"},
	}
	for _, r := range roles {
		if err := db.Where("id = ?", r.ID).FirstOrCreate(&r).Error; err != nil {
			t.Fatalf("seed role %s: %v", r.Name, err)
		}
	}
}

// SeedUser creates an active user with the given role and returns its entity.
// The plaintext password is returned so tests can log in with it.
func SeedUser(t *testing.T, db *gorm.DB, email, password, fullName, role string) entity.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	rid := roleID(role)
	if rid == 0 {
		t.Fatalf("seed user %s: unknown role %q", email, role)
	}
	// Role is built with a matching ID so GORM sees a non-blank association
	// primary key and skips the belongs-to save (a Role with only Name set would
	// trigger an upsert that conflicts with the seeded row and reset RoleID to 0,
	// causing a users_role_id_fkey violation).
	u := entity.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		FullName:     fullName,
		Role:         entity.Role{ID: rid, Name: role},
		RoleID:       rid,
		IsActive:     true,
	}
	if err := db.Create(&u).Error; err != nil {
		t.Fatalf("create user %s: %v", email, err)
	}
	return u
}

func roleID(role string) int {
	switch role {
	case "admin_fakultas":
		return 1
	case "kaprodi":
		return 2
	case "mahasiswa":
		return 3
	case "dosen_pembimbing":
		return 4
	case "dosen_penguji":
		return 5
	default:
		return 0
	}
}

// SeedActiveAcademicYear inserts the shared active academic year.
func SeedActiveAcademicYear(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	year := entity.AcademicYear{
		ID:        uuid.New(),
		Name:      "2026/2027",
		Semester:  "ganjil",
		StartDate: mustParseDate(t, "2026-09-01"),
		EndDate:   mustParseDate(t, "2027-01-31"),
		IsActive:  true,
	}
	if err := db.Create(&year).Error; err != nil {
		t.Fatalf("seed academic year: %v", err)
	}
	return year.ID
}

func mustParseDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("parse date %q: %v", s, err)
	}
	return d
}

// excludedTables are never truncated: schema_migrations is golang-migrate's
// version table (wiping it would force re-running migrations against tables
// that already exist on the next test).
var excludedTables = map[string]bool{"schema_migrations": true}

func truncateAll(t *testing.T, db *gorm.DB) {
	t.Helper()
	var tables []string
	if err := db.Raw(`SELECT tablename FROM pg_tables WHERE schemaname = 'public'`).Scan(&tables).Error; err != nil {
		t.Fatalf("list tables: %v", err)
	}
	truncatable := make([]string, 0, len(tables))
	for _, t := range tables {
		if !excludedTables[t] {
			truncatable = append(truncatable, t)
		}
	}
	if len(truncatable) == 0 {
		return
	}
	if err := db.Exec("TRUNCATE " + joinTables(truncatable) + " RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

func joinTables(tables []string) string {
	out := ""
	for i, t := range tables {
		if i > 0 {
			out += ", "
		}
		out += t
	}
	return out
}

// ── HTTP helpers ──────────────────────────────────────────────────────────

// csrfCookieName mirrors middleware.CSRFMiddleware's cookie/header names so the
// test client can drive the Double Submit Cookie pattern the way the frontend
// does (GET seeds the cookie, state-changing requests echo it as a header).
const (
	csrfTestCookieName = "XSRF-TOKEN"
	csrfTestHeaderName = "X-XSRF-TOKEN"
)

// seedCSRFCookie performs a GET on a public route so CSRFMiddleware sets the
// XSRF-TOKEN cookie, and returns it. state-changing requests then echo it.
func seedCSRFCookie(router http.Handler) *http.Cookie {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	for _, ck := range w.Result().Cookies() {
		if ck.Name == csrfTestCookieName {
			return ck
		}
	}
	return &http.Cookie{Name: csrfTestCookieName, Value: "test-csrf-fallback"}
}

// applyCSRF attaches the CSRF cookie + header to a state-changing request,
// mirroring the frontend (which reads XSRF-TOKEN from document.cookie and
// echoes it as X-XSRF-TOKEN). Safe methods are exempt from CSRF checks.
func applyCSRF(router http.Handler, req *http.Request) {
	method := req.Method
	if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
		return
	}
	ck := seedCSRFCookie(router)
	req.AddCookie(ck)
	req.Header.Set(csrfTestHeaderName, ck.Value)
}

// Client is a cookie-aware HTTP client for integration tests. It mirrors the
// frontend: it drives the Double Submit Cookie CSRF pattern (GET seeds the
// XSRF-TOKEN cookie; state-changing requests echo it as a header) and persists
// cookies across requests so the HttpOnly refresh-token cookie survives login
// and rotation. Use it for flows that depend on cookie state.
type Client struct {
	router  http.Handler
	cookies map[string]*http.Cookie
}

// NewClient wraps a router in a cookie-persistent test client.
func NewClient(router http.Handler) *Client {
	return &Client{router: router, cookies: map[string]*http.Cookie{}}
}

// Router exposes the underlying handler so callers can mix the stateless
// helpers (DoJSON/DoMultipart) with the cookie-aware client when needed.
func (c *Client) Router() http.Handler {
	return c.router
}

// seedCSRF ensures the client holds an XSRF-TOKEN cookie (fetched via a GET,
// exactly as a first page load would).
func (c *Client) seedCSRF() {
	if _, ok := c.cookies[csrfTestCookieName]; ok {
		return
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	c.router.ServeHTTP(w, req)
	for _, ck := range w.Result().Cookies() {
		c.cookies[ck.Name] = ck
	}
	if _, ok := c.cookies[csrfTestCookieName]; !ok {
		c.cookies[csrfTestCookieName] = &http.Cookie{Name: csrfTestCookieName, Value: "test-csrf-fallback"}
	}
}

// attachCookies copies every stored cookie onto the outgoing request.
func (c *Client) attachCookies(req *http.Request) {
	for _, ck := range c.cookies {
		req.AddCookie(ck)
	}
}

// captureCookies stores any Set-Cookie from the response (refresh rotation,
// logout clearing, CSRF refresh).
func (c *Client) captureCookies(w *httptest.ResponseRecorder) {
	for _, ck := range w.Result().Cookies() {
		// A cleared cookie (MaxAge < 0) removes the stored value so the client
		// truly stops sending it, matching a browser that expires the cookie.
		if ck.MaxAge < 0 {
			delete(c.cookies, ck.Name)
			continue
		}
		c.cookies[ck.Name] = ck
	}
}

// Do performs a JSON request carrying the client's cookies and CSRF header and
// returns the recorder. State-changing methods get the CSRF token echoed.
func (c *Client) Do(method, path string, body any, token string) *httptest.ResponseRecorder {
	c.seedCSRF()
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	c.attachCookies(req)
	if method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions {
		if ck := c.cookies[csrfTestCookieName]; ck != nil {
			req.Header.Set(csrfTestHeaderName, ck.Value)
		}
	}
	w := httptest.NewRecorder()
	c.router.ServeHTTP(w, req)
	c.captureCookies(w)
	return w
}

// CookieValue returns the current value of a stored cookie (or "" if absent).
func (c *Client) CookieValue(name string) string {
	if ck := c.cookies[name]; ck != nil {
		return ck.Value
	}
	return ""
}

// SetCookie seeds a cookie into the client's jar (used to replay a captured
// token, e.g. a stale refresh token, as a separate browser session would).
func (c *Client) SetCookie(name, value string) {
	c.cookies[name] = &http.Cookie{Name: name, Value: value, Path: "/"}
}

// DoJSON performs a JSON request against the router and returns the recorder.
func DoJSON(router http.Handler, method, path string, body any, token string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	applyCSRF(router, req)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// DoMultipart performs a multipart/form-data request (file upload) and
// returns the recorder.
func DoMultipart(router http.Handler, method, path string, fields map[string]string, fileField, fileName string, fileContent []byte, token string) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			panic(err)
		}
	}
	if fileField != "" {
		fw, err := mw.CreateFormFile(fileField, fileName)
		if err != nil {
			panic(err)
		}
		if _, err := fw.Write(fileContent); err != nil {
			panic(err)
		}
	}
	if err := mw.Close(); err != nil {
		panic(err)
	}

	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	applyCSRF(router, req)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// TestPDF is a minimal-but-valid PDF (the validator only sniffs the %PDF-
// magic), for uploads that must pass the PDF check.
func TestPDF() []byte {
	return []byte("%PDF-1.4\n% test document\n1 0 obj\n<<>>\nendobj\n%%EOF\n")
}

// SubmitThesis posts a thesis submission as multipart/form-data with a valid
// draft proposal PDF attached (POST /theses now requires the proposal file).
func SubmitThesis(t *testing.T, router http.Handler, fields map[string]string, token string) *httptest.ResponseRecorder {
	t.Helper()
	return DoMultipart(router, http.MethodPost, "/api/v1/theses", fields, "file", "draft-proposal.pdf", TestPDF(), token)
}

// DecodeBody unmarshals the recorder body into out.
func DecodeBody(t *testing.T, w *httptest.ResponseRecorder, out any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), out); err != nil {
		t.Fatalf("decode body %q: %v", w.Body.String(), err)
	}
}
