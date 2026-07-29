# Job 23 — Testing Strategy

**Phase:** 5 — Integration & Testing
**Referensi PRD:** Section 22 (Acceptance Criteria), Section 7 (Non-Functional Requirements)
**Prerequisites:** Job 22 (API Documentation) ✅
**Estimasi:** 4 hari

---

## Objective

Implementasi test suite yang memadai: unit test untuk business logic kritis di backend, integration test untuk API endpoint, dan smoke test frontend. Target: coverage cukup untuk memvalidasi semua acceptance criteria sebelum go-live, bukan 100% coverage.

---

## Checklist

### Backend — Unit Tests

Install testing dependencies:
```bash
go get github.com/stretchr/testify
go get github.com/stretchr/mock
go get github.com/DATA-DOG/go-sqlmock
```

#### Prioritas Test (berdasarkan business criticality):

**1. Grading Calculator** (`pkg/grading/calculator_test.go`)
- [ ] Test `CalculateFinalScore` dengan 2 penguji, nilai berbeda, bobot standar
- [ ] Test nilai akhir tepat di batas: 59.9 (failed), 60.0 (passed), 74.9 (passed with revision), 75.0 (passed)
- [ ] Test dengan 1 penguji (edge case minimum)
- [ ] Test kalkulasi tidak melebihi 100 dan tidak kurang dari 0

**2. Thesis State Machine** (`pkg/statemachine/thesis_state_test.go`)
- [ ] Test setiap transisi valid (semua pasangan from→to yang diizinkan)
- [ ] Test setiap transisi invalid (contoh: `submitted` → `graduated` langsung)
- [ ] Test semua status memiliki at least 1 valid transition

**3. Auth Use Case** (`internal/usecase/auth_usecase_test.go`)
- [ ] Test login sukses → return token
- [ ] Test login gagal (password salah) → increment attempt count
- [ ] Test login setelah 5 kali gagal → account locked
- [ ] Test token blacklist: logout → token ditambah ke blacklist
- [ ] Test reset password: token valid → password berubah
- [ ] Test reset password: token expired → error

**4. Gate Checker** (`internal/usecase/gate_checker_test.go`)
- [ ] Test `CanSubmitSeminar`: dokumen `seminar_doc` belum approved → false
- [ ] Test `CanSubmitSeminar`: dokumen `seminar_doc` approved → true
- [ ] Test `CanSubmitDefense`: seminar belum `passed` → false
- [ ] Test `CanSubmitDefense`: seminar `passed` + dokumen `defense_doc` approved → true

**5. Document Versioning** (`internal/usecase/document_usecase_test.go`)
- [ ] Test upload dokumen baru → version = 1
- [ ] Test upload lagi dengan type yang sama → version = 2
- [ ] Test file size validation > 10MB → error
- [ ] Test file type bukan PDF → error

#### Mock Setup
- [ ] Buat mock repository menggunakan `testify/mock` untuk semua repository interface
- [ ] Struktur: `internal/usecase/mocks/mock_user_repository.go`, dst.
- [ ] Atau gunakan `mockery` untuk auto-generate:
  ```bash
  go install github.com/vektra/mockery/v2@latest
  mockery --all --dir internal/domain/repository --output internal/usecase/mocks
  ```

### Backend — Integration Tests

**File:** `internal/handler/*_handler_test.go`

Setup test environment:
```go
// internal/testutil/setup.go
func SetupTestDB(t *testing.T) *gorm.DB {
    // Gunakan SQLite in-memory atau PostgreSQL test database
    // Jalankan migrasi
    // Return DB connection
}

func SetupTestRouter(t *testing.T, db *gorm.DB) *gin.Engine {
    // Setup router dengan semua middleware dan handler
    // Inject test DB
}
```

- [ ] Gunakan PostgreSQL test database terpisah: `simtas_filkom_test`
- [ ] Setiap test membersihkan data setelah selesai (defer cleanup)

**Endpoint tests yang wajib dibuat:**

- [ ] `POST /auth/login` — sukses, gagal, locked
- [ ] `GET /auth/me` — dengan token valid, tanpa token, token expired
- [ ] `POST /theses` — mahasiswa submit judul, duplikat thesis, role salah
- [ ] `PUT /theses/:id/review` — Kaprodi approve, Kaprodi reject, mahasiswa coba review
- [ ] `PUT /theses/:id/assign-supervisor` — valid, dosen role salah, thesis status salah
- [ ] `POST /theses/:id/documents` — upload PDF valid, upload non-PDF, file terlalu besar
- [ ] `PATCH /documents/:id/review` — approve, revision, reviewer bukan pembimbing
- [ ] `POST /seminars/:id/scores` — penguji valid, penguji tidak ditugaskan, sudah pernah submit
- [ ] `PUT /defenses/:id/graduation` — prasyarat terpenuhi, prasyarat belum terpenuhi
- [ ] `GET /archives?q=machine+learning` — full-text search berfungsi

```go
// Contoh integration test
func TestThesisSubmission(t *testing.T) {
    db := testutil.SetupTestDB(t)
    router := testutil.SetupTestRouter(t, db)
    
    // Seed: buat user mahasiswa + academic year aktif
    student := testutil.SeedStudent(db)
    token := testutil.GetToken(router, student.Email, "password")
    
    t.Run("submit thesis success", func(t *testing.T) {
        body := map[string]interface{}{
            "title":          "Judul Skripsi Test Minimal Sepuluh Kata",
            "abstract":       strings.Repeat("kata ", 100),
            "field_of_study": "Kecerdasan Buatan",
            "thesis_type":    "skripsi",
        }
        w := testutil.DoRequest(router, "POST", "/api/v1/theses", body, token)
        assert.Equal(t, 201, w.Code)
    })
    
    t.Run("cannot submit if already has active thesis", func(t *testing.T) {
        // Submit lagi dengan mahasiswa yang sama
        w := testutil.DoRequest(router, "POST", "/api/v1/theses", body, token)
        assert.Equal(t, 400, w.Code)
    })
}
```

### Frontend — Tests

Install:
```bash
npm install -D @testing-library/react @testing-library/jest-dom @testing-library/user-event
npm install -D jest jest-environment-jsdom ts-jest
npm install -D msw  # Mock Service Worker untuk mock API
```

**Unit test komponen:**
- [ ] `StatusBadge` — test semua status menampilkan label dan warna yang benar
- [ ] `ProgressStepper` — test setiap status thesis menampilkan step yang benar sebagai active/done
- [ ] Kalkulasi `progress_percentage` — test setiap status menghasilkan persentase yang sesuai
- [ ] `getErrorMessage` helper — test berbagai jenis error input

**Smoke tests (minimal):**
- [ ] Halaman login render tanpa crash
- [ ] Form login — submit dengan data kosong menampilkan error validasi
- [ ] Halaman dashboard render sesuai role (mock auth store)

### Makefile Targets

- [ ] Tambah ke `backend/Makefile`:
  ```makefile
  test:
      go test ./... -v -count=1

  test-unit:
      go test ./pkg/... ./internal/usecase/... -v

  test-integration:
      go test ./internal/handler/... -v -tags integration

  test-coverage:
      go test ./... -coverprofile=coverage.out
      go tool cover -html=coverage.out -o coverage.html
  ```

### CI — Test Pipeline

- [ ] Update `.github/workflows/ci.yml` — tambah step test:
  ```yaml
  - name: Run backend tests
    run: make test
    working-directory: backend
    env:
      DB_URL: postgres://postgres:postgres@localhost:5432/simtas_test

  - name: Run frontend tests
    run: npm test -- --watchAll=false
    working-directory: frontend
  ```
- [ ] Tambah PostgreSQL service ke CI job:
  ```yaml
  services:
    postgres:
      image: postgres:16
      env:
        POSTGRES_DB: simtas_test
        POSTGRES_PASSWORD: postgres
      options: >-
        --health-cmd pg_isready
        --health-interval 10s
  ```

### Coverage Target

Bukan 100% — fokus pada kualitas test, bukan kuantitas:
- Backend use case layer: **≥ 80% coverage**
- Backend handler layer: **≥ 60% coverage** (endpoint kritis)
- Frontend komponen kritis: **≥ 50% coverage**

---

## Done Criteria

- [ ] `make test-unit` → semua unit test pass
- [ ] `make test-integration` → semua integration test pass
- [ ] Grading calculator: edge case 59.9 → failed, 60.0 → passed terverifikasi
- [ ] State machine: transisi invalid → test fail dengan error yang benar
- [ ] Gate checker: `CanSubmitSeminar` return false jika dokumen belum approved
- [ ] Integration test login: 5x gagal → akun terkunci (403)
- [ ] Integration test upload dokumen: non-PDF → 400, >10MB → 400
- [ ] Integration test RBAC: mahasiswa akses endpoint Kaprodi → 403
- [ ] CI pipeline menjalankan semua test otomatis dan gagal jika ada test yang merah
- [ ] `go tool cover` report coverage ≥ 80% untuk package `usecase`
- [ ] Frontend: `npm test` → semua test pass
- [ ] **MILESTONE Phase 5:** Sistem tervalidasi secara otomatis, siap untuk deployment
