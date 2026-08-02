# Job 22 — API Documentation (Swagger/OpenAPI)

**Phase:** 5 — Integration & Testing
**Referensi PRD:** Section 7 (Non-Functional Requirements — RESTful API Documentation)
**Prerequisites:** Job 21 (Storage Integration) ✅
**Estimasi:** 2 hari

---

## Objective

Generate dokumentasi API lengkap menggunakan Swagger/OpenAPI dari kode Go, pastikan semua endpoint terdokumentasi dengan request/response schema, contoh, dan keterangan autentikasi. Dokumentasi ini digunakan oleh developer frontend dan sebagai referensi integrasi.

---

## Checklist

### Setup Swaggo

- [x] Install swaggo:
  ```bash
  go install github.com/swaggo/swag/cmd/swag@v1.16.4
  go get github.com/swaggo/gin-swagger
  go get github.com/swaggo/files
  go get github.com/swaggo/swag
  ```
- [x] Tambah `swag init` ke Makefile (`make docs`):
  ```makefile
  docs:
      swag init -g cmd/server/main.go -o docs
  ```
- [x] Setup route Swagger UI di Gin (hanya di non-production):
  ```go
  if cfg.AppEnv != "production" {
      engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
  }
  ```
- [x] Akses di: `http://localhost:8080/swagger/index.html`

### Anotasi General (`cmd/server/main.go`)

- [x] Tambah anotasi utama:
  ```go
  // @title           SIMTAS FILKOM API
  // @version         1.0
  // @description     API untuk Sistem Manajemen Tugas Akhir dan Skripsi
  //                  Fakultas Ilmu Komputer Universitas Djuanda
  // @termsOfService  https://filkom.unida.ac.id

  // @contact.name   Admin SIMTAS FILKOM
  // @contact.email  admin@filkom.unida.ac.id

  // @license.name  MIT
  // @license.url   https://opensource.org/licenses/MIT

  // @host      localhost:8080
  // @BasePath  /api/v1

  // @securityDefinitions.apikey BearerAuth
  // @in header
  // @name Authorization
  // @description Masukkan token dengan format: Bearer {token}
  ```

### Anotasi Endpoint per Handler

Dokumentasikan SEMUA endpoint dari job 03–13. Contoh format:

#### Auth Handler
```go
// Login godoc
// @Summary      Login pengguna
// @Description  Autentikasi dengan email dan password, return JWT token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Kredensial login"
// @Success      200  {object}  response.ApiResponse{data=dto.LoginResponse}
// @Failure      401  {object}  response.ApiResponse "Email atau password salah"
// @Failure      403  {object}  response.ApiResponse "Akun terkunci"
// @Failure      429  {object}  response.ApiResponse "Terlalu banyak percobaan"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) { ... }
```

- [x] Dokumentasikan semua endpoint Auth (6 endpoint)
- [x] Dokumentasikan semua endpoint User Management (11 endpoint)
- [x] Dokumentasikan semua endpoint Thesis (7 endpoint)
- [x] Dokumentasikan semua endpoint Consultation (7 endpoint)
- [x] Dokumentasikan semua endpoint Document (6 endpoint)
- [x] Dokumentasikan semua endpoint Seminar (7 endpoint)
- [x] Dokumentasikan semua endpoint Defense (9 endpoint)
- [x] Dokumentasikan semua endpoint Archive (6 endpoint)
- [x] Dokumentasikan semua endpoint Dashboard (6 endpoint)
- [x] Dokumentasikan semua endpoint Academic Year (4 endpoint)
- [x] Dokumentasikan endpoint Audit Log (2 endpoint)
- [x] Dokumentasikan endpoint Health + Internal test-email (2 endpoint)

### Dokumentasi DTO Structs

- [x] Request/response DTO direferensikan langsung dari anotasi handler (swag menghasilkan schema otomatis dari struct Go)
- [x] Upload endpoint menggunakan `formData` + `@Accept multipart/form-data` (documents, archive, user import)
- [x] Named request types `revisionNotesRequest`, `documentReviewRequest`, `testEmailRequest` didefinisikan di package handler agar swag dapat me-resolve schema
- [x] Enum fields didokumentasikan lewat deskripsi `@Param` (contoh: daftar tipe dokumen di `POST /theses/{thesis_id}/documents`)

### Grouping Endpoint dengan Tags

- [x] Grup tags yang rapi di Swagger UI:
  - `Authentication` — login, logout, refresh, me, forgot/reset password
  - `User Management` — CRUD user, import, aktivasi
  - `Academic Years` — manajemen tahun akademik
  - `Thesis Submission` — pengajuan dan review judul
  - `Supervision` — log konsultasi bimbingan
  - `Documents` — upload dan review dokumen
  - `Seminar Proposal` — alur seminar
  - `Thesis Defense` — alur sidang
  - `Archives` — arsip digital
  - `Dashboard` — endpoint dashboard per role
  - `Audit Log` — audit trail
  - `Health` + `Internal` — endpoint publik/dev

### Security Pada Endpoint Protected

- [x] Tandai semua endpoint yang butuh auth:
  ```go
  // @Security BearerAuth
  ```
- [x] Swagger UI menampilkan tombol "Authorize" untuk input token (via `@securityDefinitions.apikey BearerAuth`)
- [x] Test endpoint langsung dari Swagger UI berfungsi
- [x] Endpoint publik (`/health`, `/auth/login`, `/auth/forgot-password`, `/auth/reset-password`) TIDAK memakai `@Security BearerAuth`

### Export OpenAPI Spec

- [x] Hasil `swag init` menghasilkan `docs/swagger.json`, `docs/swagger.yaml`, dan `docs/docs.go`
- [x] Commit file `docs/` (package Go yang di-import `cmd/server/main.go`) ke repository
- [x] Tambah `swag init` ke CI pipeline (verifikasi docs up to date, swag di-pin ke v1.16.4):
  ```yaml
  - name: Verify swagger docs are up to date
    run: |
      export PATH="$PATH:$(go env GOPATH)/bin"
      swag init -g cmd/server/main.go -o docs
      git diff --exit-code docs/ || (echo "Swagger docs are out of date. Run 'make docs' and commit the changes." && exit 1)
  ```

### Postman Collection (Bonus)

- [ ] Export `swagger.json` → import ke Postman sebagai Postman Collection
- [ ] Simpan collection di `docs/postman/SIMTAS_FILKOM.postman_collection.json`
- [ ] Buat Postman Environment: `SIMTAS_Local` dengan variabel `base_url`, `access_token`

---

## Done Criteria

- [x] `http://localhost:8080/swagger/index.html` menampilkan dokumentasi lengkap (non-production)
- [x] Semua endpoint (63 endpoint: 62 + `/documents/{id}/review`) terdokumentasi dengan deskripsi, request, dan response schema
- [x] Request/response schema otomatis dari struct DTO (swag)
- [x] Tombol "Authorize" di Swagger UI → input token → test endpoint protected berfungsi
- [x] `swag init` berjalan tanpa error
- [x] CI pipeline gagal jika `swag init` menghasilkan diff dari file yang di-commit
- [x] `docs/swagger.json` + `docs/swagger.yaml` + `docs/docs.go` ter-commit di repository
- [x] Endpoint dikelompokkan dalam tags yang rapi di Swagger UI
- [x] Swagger UI tidak muncul di production (`APP_ENV=production`)
- [x] Catatan: upload multipart (documents/archive/import) didokumentasikan dengan `formData`
