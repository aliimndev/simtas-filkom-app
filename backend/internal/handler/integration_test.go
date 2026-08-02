//go:build integration

package handler_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/domain/entity"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
)

const (
	adminEmail      = "admin-it@filkom.unida.ac.id"
	studentEmail    = "student-it@filkom.unida.ac.id"
	adminPassword   = "Password@123"
	studentPassword = "StudentPass1"
)

// newIntegrationApp seeds the test DB (admin + student + active academic year)
// and returns the fully-wired router.
func newIntegrationApp(t *testing.T) http.Handler {
	t.Helper()
	db := testutil.SetupTestDB(t)
	router := testutil.SetupTestRouter(t, db)

	testutil.SeedUser(t, db, adminEmail, adminPassword, "Admin IT", "admin_fakultas")
	testutil.SeedUser(t, db, studentEmail, studentPassword, "Student IT", "mahasiswa")
	testutil.SeedActiveAcademicYear(t, db)
	return router
}

func loginToken(t *testing.T, router http.Handler, email, pass string) string {
	t.Helper()
	w := testutil.DoJSON(router, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    email,
		"password": pass,
	}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("login %s failed: %d %s", email, w.Code, w.Body.String())
	}
	var body struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	testutil.DecodeBody(t, w, &body)
	if body.Data.AccessToken == "" {
		t.Fatal("login response missing access_token")
	}
	return body.Data.AccessToken
}

// TestLoginSuccess — Job 23: login sukses mengembalikan token.
func TestLoginSuccess(t *testing.T) {
	router := newIntegrationApp(t)
	token := loginToken(t, router, adminEmail, adminPassword)
	if token == "" {
		t.Fatal("expected a token")
	}
}

// TestLoginWrongPassword — Job 23: password salah → 401.
func TestLoginWrongPassword(t *testing.T) {
	router := newIntegrationApp(t)
	w := testutil.DoJSON(router, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    adminEmail,
		"password": "WrongPass1",
	}, "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestAuthMe — Job 23: GET /auth/me dengan token valid dan tanpa token.
func TestAuthMe(t *testing.T) {
	router := newIntegrationApp(t)
	token := loginToken(t, router, adminEmail, adminPassword)

	w := testutil.DoJSON(router, http.MethodGet, "/api/v1/auth/me", nil, token)
	assert.Equal(t, http.StatusOK, w.Code)

	wNoToken := testutil.DoJSON(router, http.MethodGet, "/api/v1/auth/me", nil, "")
	assert.Equal(t, http.StatusUnauthorized, wNoToken.Code)
}

// TestThesisSubmissionAndRBAC — Job 23: mahasiswa submit judul; mahasiswa
// tidak bisa akses endpoint admin (RBAC → 403).
func TestThesisSubmissionAndRBAC(t *testing.T) {
	router := newIntegrationApp(t)
	studentToken := loginToken(t, router, studentEmail, studentPassword)
	adminToken := loginToken(t, router, adminEmail, adminPassword)

	// Mahasiswa akses endpoint admin → 403.
	w := testutil.DoJSON(router, http.MethodGet, "/api/v1/admin/users", nil, studentToken)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// Admin akses endpoint admin → 200.
	wAdmin := testutil.DoJSON(router, http.MethodGet, "/api/v1/admin/users", nil, adminToken)
	assert.Equal(t, http.StatusOK, wAdmin.Code)

	// Mahasiswa submit judul → 201.
	wSubmit := testutil.DoJSON(router, http.MethodPost, "/api/v1/theses", thesisPayload(), studentToken)
	assert.Equal(t, http.StatusCreated, wSubmit.Code, "submit thesis body: %s", wSubmit.Body.String())
}

// TestDocumentUploadValidation — Job 23: upload non-PDF → 400.
func TestDocumentUploadValidation(t *testing.T) {
	router := newIntegrationApp(t)
	studentToken := loginToken(t, router, studentEmail, studentPassword)

	// Buat thesis dulu supaya ada konteks upload.
	wSubmit := testutil.DoJSON(router, http.MethodPost, "/api/v1/theses", thesisPayload(), studentToken)
	if wSubmit.Code != http.StatusCreated {
		t.Fatalf("setup thesis failed: %d %s", wSubmit.Code, wSubmit.Body.String())
	}
	var thesisBody struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	testutil.DecodeBody(t, wSubmit, &thesisBody)
	thesisID := thesisBody.Data.ID

	// File non-PDF → 400.
	w := testutil.DoMultipart(router, http.MethodPost, fmt.Sprintf("/api/v1/theses/%s/documents", thesisID),
		map[string]string{"document_type": entity.DocTypeProposal},
		"file", "bab1.txt", []byte("plain text, not a pdf"), studentToken)
	assert.Equal(t, http.StatusBadRequest, w.Code, "non-pdf upload body: %s", w.Body.String())
}

// TestArchiveSearch — Job 23: GET /archives dengan query kata kunci.
func TestArchiveSearch(t *testing.T) {
	router := newIntegrationApp(t)
	token := loginToken(t, router, adminEmail, adminPassword)

	w := testutil.DoJSON(router, http.MethodGet, "/api/v1/archives?q=machine+learning", nil, token)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestLoginLocked — Job 23: 5x password salah → akun terkunci (403).
// Per usecase: percobaan ke-5 yang gagal langsung mengunci akun (newCount >= 5),
// jadi 4x 401 dulu, lalu percobaan ke-5 → 403, dan login dengan password benar
// setelah terkunci juga tetap 403.
func TestLoginLocked(t *testing.T) {
	router := newIntegrationApp(t)

	for i := 0; i < 4; i++ {
		w := testutil.DoJSON(router, http.MethodPost, "/api/v1/auth/login", map[string]string{
			"email":    studentEmail,
			"password": "WrongPass1",
		}, "")
		assert.Equal(t, http.StatusUnauthorized, w.Code,
			"attempt %d: want 401, got %d %s", i+1, w.Code, w.Body.String())
	}

	// Percobaan ke-5 (gagal) → akun terkunci (403).
	w := testutil.DoJSON(router, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    studentEmail,
		"password": "WrongPass1",
	}, "")
	assert.Equal(t, http.StatusForbidden, w.Code, "locked login body: %s", w.Body.String())

	// Password benar pun tetap 403 karena akun sudah terkunci.
	wOk := testutil.DoJSON(router, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    studentEmail,
		"password": studentPassword,
	}, "")
	assert.Equal(t, http.StatusForbidden, wOk.Code, "locked login with correct password: %d %s", wOk.Code, wOk.Body.String())
}

func repeatWords(word string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += word
	}
	return out
}

// thesisPayload returns a valid CreateThesisRequest body (abstract ≥ 100 kata).
func thesisPayload() map[string]string {
	// Title must be >= 10 kata (usecase: ErrTitleTooShort bila < 10).
	return map[string]string{
		"title":          "Sistem Informasi Manajemen Tugas Akhir Berbasis Web Untuk Fakultas Ilmu Komputer",
		"abstract":       repeatWords("abstrak ", 110),
		"field_of_study": "Rekayasa Perangkat Lunak",
		"thesis_type":    "skripsi",
	}
}
