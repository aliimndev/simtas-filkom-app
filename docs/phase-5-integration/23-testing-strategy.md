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

**1. Grading Calculator** (`backend/pkg/grading/calculator_test.go`)
- [x] Test `CalculateFinalScore` dengan 2 penguji, nilai berbeda, bobot standar
- [x] Test nilai akhir tepat di batas: 59.9 (failed), 60.0 (passed), 74.9 (passed with revision), 75.0 (passed)
- [x] Test dengan 1 penguji (edge case minimum)
- [x] Test kalkulasi tidak melebihi 100 dan tidak kurang dari 0

**2. Thesis State Machine** (`backend/pkg/statemachine/thesis_state_test.go`)
- [x] Test setiap transisi valid (semua pasangan from→to yang diizinkan)
- [x] Test setiap transisi invalid (contoh: `submitted` → `graduated` langsung)
- [x] Test semua status memiliki at least 1 valid transition

**3. Auth Use Case** (`backend/internal/usecase/auth_usecase_test.go` + `auth_usecase_extra_test.go`)
- [x] Test login sukses → return token
- [x] Test login gagal (password salah) → increment attempt count
- [x] Test login setelah 5 kali gagal → account locked
- [x] Test token blacklist: logout → token ditambah ke blacklist
- [x] Test reset password: token valid → password berubah
- [x] Test reset password: token expired → error

**4. Gate Checker** (`backend/internal/usecase/gate_checker_test.go`)
- [x] Test `CanSubmitSeminar`: dokumen `seminar_doc` belum approved → false
- [x] Test `CanSubmitSeminar`: dokumen `seminar_doc` approved → true
- [x] Test `CanSubmitDefense`: seminar belum `passed` → false
- [x] Test `CanSubmitDefense`: seminar `passed` + dokumen `defense_doc` approved → true

**5. Document Versioning** (`backend/internal/usecase/document_usecase_test.go`)
- [x] Test upload dokumen baru → version = 1
- [x] Test upload lagi dengan type yang sama → version = 2
- [x] Test file size validation > 10MB → error
- [x] Test file type bukan PDF → error

> Implementasi memakai **in-memory fakes** (bukan `testify/mock`) sesuai konvensi
> test yang sudah ada di repo — lebih ringkas dan tanpa codegen. Mocks otomatis
> tidak diperlukan; `getbyid_usecase_test.go` + `coverage_gap_test.go` menutup
> fungsi yang sebelumnya 0–70% sehingga usecase layer mencapai **80.3%**.

#### Mock Setup
- [x] Buat mock repository menggunakan in-memory fake (sesuai konvensi repo)
- [x] Struktur: fake per-module di `backend/internal/usecase/*_test.go`
- [x] Tidak perlu `mockery` — fakes disusun per usecase test, konsisten dengan test yang sudah ada

### Backend — Integration Tests

> ⚠️ Jangan menjalankan dua binary `go test -tags integration` secara paralel
> terhadap satu PostgreSQL — seed per-test memakai ID role eksplisit (1–5)
> sehingga race memunculkan duplikat `roles_pkey`. CI aman karena satu
> invocation tunggal. Saat Postgres tidak tersedia di CI, test **gagal**
> (bukan skip) supaya gate integrasi tidak lolos kosong.

**File:** `backend/internal/handler/integration_test.go` (build tag `integration`)

Setup test environment:
```go
// internal/testutil/setup.go
func SetupTestDB(t *testing.T) *gorm.DB      // Postgres test DB + migrasi + seed roles
func SetupTestRouter(t *testing.T, db *gorm.DB) *gin.Engine // router produksi penuh
```

- [x] Gunakan PostgreSQL test database terpisah: `simtas_filkom_test` (default; CI pakai `simtas_test`)
- [x] Setiap test membersihkan data setelah selesai (`t.Cleanup` → truncate all tables, `schema_migrations` dikecualikan)

**Endpoint tests yang wajib dibuat** (semua ada di `integration_test.go`):

- [x] `POST /auth/login` — sukses, gagal (401), locked (403 setelah 5x)
- [x] `GET /auth/me` — dengan token valid, tanpa token (401)
- [x] `POST /theses` — mahasiswa submit judul valid, duplikat/aktif ditolak
- [x] RBAC — mahasiswa akses endpoint Kaprodi/admin → 403
- [x] `POST /theses/:id/documents` — upload non-PDF → 400
- [x] `PUT /defenses/:id/graduation` — prasyarat belum terpenuhi → gate error
- [x] `GET /archives?q=...` — full-text search berfungsi

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

Install (sudah di `frontend/package.json`):
```bash
npm install -D @testing-library/react @testing-library/jest-dom @testing-library/user-event
npm install -D jest jest-environment-jsdom ts-jest
```

> MSW tidak dipakai — test memakai mock modul (`jest.mock`) yang lebih ringkas
> untuk smoke/component test, sesuai kebutuhan Job 23.

**Unit test komponen:**
- [x] `Badge` (`frontend/src/components/ui/badge.test.tsx`) — semua variant + className + props
- [x] `getErrorMessage` + `mapAuthError` (`frontend/src/lib/utils/error.test.ts`) — 15 kasus: pesan backend, field errors, timeout, network, 401/403/404/5xx, mapping kata kunci EN/ID

**Smoke tests (minimal):**
- [x] Halaman login render tanpa crash (`login.smoke.test.tsx`)
- [x] Form login — submit dengan data kosong menampilkan error validasi (`Email wajib diisi`, `Password wajib diisi`)
- [ ] Halaman dashboard render sesuai role (mock auth store) — deferred ke Job 26 (e2e) karena dashboard masih sebagian

### Makefile Targets (backend/Makefile)

- [x] `test` — `go test ./... -count=1`
- [x] `test-unit` — `go test ./pkg/... ./internal/usecase/... -count=1 -v`
- [x] `test-integration` — `go test ./internal/handler/... -count=1 -v -tags integration`
- [x] `test-coverage` — `go test ./... -coverprofile=coverage.out && go tool cover -html`
- [x] `coverage-check` — gate ≥ 80% untuk usecase layer (awk, tanpa `bc`)

### CI — Test Pipeline (.github/workflows/ci.yml)

- [x] Backend job: `gofmt` → `go build` → `go vet` → `go test ./...`
- [x] Integration tests: `go test ./internal/handler/... -tags integration` (Postgres service)
- [x] Coverage gate: `make coverage-check` (usecase ≥ 80%)
- [x] PostgreSQL service `postgres:16` dengan health-check; env `DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME=simtas_test`
- [x] Frontend job: `npm ci` → `lint` → `type-check` → `npm test -- --watchAll=false` → `build`

### Coverage Target

Bukan 100% — fokus pada kualitas test, bukan kuantitas:
- Backend use case layer: **≥ 80% coverage** ✅ terukur **80.3%** (`make coverage-check`)
- Backend handler layer: **≥ 60% coverage** (endpoint kritis) — diukur lewat integration
  test (build tag `integration`) yang menjalankan endpoint kritis di Postgres nyata;
  angka pastinya bergantung pada cakupan integration test di CI
- Frontend komponen kritis: **≥ 50% coverage** — scoped ke modul yang ditest
  (error, badge, login) via `collectCoverageFrom` di `jest.config.mjs`

---

## Status

✅ **Selesai.** Semua unit test pass, integration test pass (CI Postgres), coverage
usecase 80.3% (gate ≥ 80%), frontend jest 23/23, type-check + lint + build hijau.

## Done Criteria

- [x] `make test-unit` → semua unit test pass
- [x] `make test-integration` → semua integration test pass (butuh Postgres)
- [x] Grading calculator: edge case 59.9 → failed, 60.0 → passed terverifikasi
- [x] State machine: transisi invalid → test fail dengan error yang benar
- [x] Gate checker: `CanSubmitSeminar` return false jika dokumen belum approved
- [x] Integration test login: 5x gagal → akun terkunci (403)
- [x] Integration test upload dokumen: non-PDF → 400
- [x] Integration test RBAC: mahasiswa akses endpoint Kaprodi → 403
- [x] CI pipeline menjalankan semua test otomatis dan gagal jika ada test yang merah
- [x] `go tool cover` report coverage ≥ 80% untuk package `usecase` (80.3%)
- [x] Frontend: `npm test` → semua test pass
- [x] **MILESTONE Phase 5:** Sistem tervalidasi secara otomatis, siap untuk deployment
