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

- [ ] Install swaggo:
  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  go get github.com/swaggo/gin-swagger
  go get github.com/swaggo/files
  go get github.com/swaggo/swag
  ```
- [ ] Tambah `swag init` ke Makefile:
  ```makefile
  docs:
      swag init -g cmd/server/main.go -o docs/swagger
  ```
- [ ] Setup route Swagger UI di Gin (hanya di non-production):
  ```go
  if cfg.Env != "production" {
      r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
  }
  ```
- [ ] Akses di: `http://localhost:8080/swagger/index.html`

### Anotasi General (`cmd/server/main.go`)

- [ ] Tambah anotasi utama:
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

- [ ] Dokumentasikan semua endpoint Auth (6 endpoint)
- [ ] Dokumentasikan semua endpoint User Management (8 endpoint)
- [ ] Dokumentasikan semua endpoint Thesis (7 endpoint)
- [ ] Dokumentasikan semua endpoint Consultation (6 endpoint)
- [ ] Dokumentasikan semua endpoint Document (5 endpoint)
- [ ] Dokumentasikan semua endpoint Seminar (7 endpoint)
- [ ] Dokumentasikan semua endpoint Defense (7 endpoint)
- [ ] Dokumentasikan semua endpoint Archive (5 endpoint)
- [ ] Dokumentasikan semua endpoint Dashboard (6 endpoint)
- [ ] Dokumentasikan semua endpoint Academic Year (4 endpoint)
- [ ] Dokumentasikan endpoint Audit Log (2 endpoint)

### Dokumentasi DTO Structs

- [ ] Tambah tag `swaggertype` dan `example` ke semua DTO:
  ```go
  type LoginRequest struct {
      Email    string `json:"email"    example:"mahasiswa@filkom.unida.ac.id" validate:"required,email"`
      Password string `json:"password" example:"Password@123"                validate:"required"`
  }

  type LoginResponse struct {
      AccessToken  string   `json:"access_token"`
      RefreshToken string   `json:"refresh_token"`
      ExpiresIn    int      `json:"expires_in"    example:"86400"`
      User         UserInfo `json:"user"`
  }
  ```
- [ ] Semua DTO yang digunakan sebagai request/response memiliki field `example`
- [ ] Enum fields menggunakan `enums` tag:
  ```go
  ThesisType string `json:"thesis_type" enums:"skripsi,tugas_akhir" example:"skripsi"`
  ```

### Grouping Endpoint dengan Tags

- [ ] Grup tags yang rapi di Swagger UI:
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
  - `Audit Logs` — audit trail

### Security Pada Endpoint Protected

- [ ] Tandai semua endpoint yang butuh auth:
  ```go
  // @Security BearerAuth
  ```
- [ ] Swagger UI menampilkan tombol "Authorize" untuk input token
- [ ] Test endpoint langsung dari Swagger UI berfungsi

### Export OpenAPI Spec

- [ ] Hasil `swag init` menghasilkan `docs/swagger/swagger.json` dan `swagger.yaml`
- [ ] Commit file `docs/swagger/` ke repository
- [ ] Tambah `swag init` ke CI pipeline (verifikasi docs up to date):
  ```yaml
  - name: Check Swagger docs are up to date
    run: |
      swag init -g cmd/server/main.go -o /tmp/swagger-check
      diff docs/swagger/swagger.json /tmp/swagger-check/swagger.json
  ```

### Postman Collection (Bonus)

- [ ] Export `swagger.json` → import ke Postman sebagai Postman Collection
- [ ] Simpan collection di `docs/postman/SIMTAS_FILKOM.postman_collection.json`
- [ ] Buat Postman Environment: `SIMTAS_Local` dengan variabel `base_url`, `access_token`

---

## Done Criteria

- [ ] `http://localhost:8080/swagger/index.html` menampilkan dokumentasi lengkap
- [ ] Semua endpoint (≥60 endpoint) terdokumentasi dengan deskripsi, request, dan response schema
- [ ] Semua DTO memiliki field `example` yang realistis
- [ ] Tombol "Authorize" di Swagger UI → input token → test endpoint protected berfungsi
- [ ] `swag init` berjalan tanpa error
- [ ] CI pipeline gagal jika `swag init` menghasilkan diff dari file yang di-commit
- [ ] `docs/swagger/swagger.json` ter-commit di repository
- [ ] Endpoint dikelompokkan dalam tags yang rapi di Swagger UI
- [ ] Swagger UI tidak muncul di production (`APP_ENV=production`)
