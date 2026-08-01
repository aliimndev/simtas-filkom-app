# Job 03 — Authentication & RBAC

**Phase:** 1 — Foundation
**Referensi PRD:** Section 6.1 (FR-AUTH-001 s/d FR-AUTH-004, FR-USER-003, FR-USER-005), Section 14 (Authentication & Authorization)
**Prerequisites:** Job 02 (Database Schema) ✅
**Estimasi:** 3–4 hari

---

## Objective

Implementasi sistem autentikasi JWT lengkap dengan access token + refresh token, token blacklist untuk logout, self-service reset password via email, account lock setelah login gagal berulang, dan middleware RBAC untuk semua 5 role. Setelah job ini selesai, seluruh mekanisme keamanan inti tersedia dan siap digunakan oleh semua modul berikutnya.

---

## Checklist

### Auth Service & Repository
- [x] Buat `backend/internal/domain/repository/auth_repository.go` — interface:
  ```go
  type AuthRepository interface {
    FindUserByEmail(ctx context.Context, email string) (*entity.User, error)
    UpdateLoginAttempt(ctx context.Context, userID uuid.UUID, count int, lockedUntil *time.Time) error
    UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
    BlacklistToken(ctx context.Context, jti string, expiresAt time.Time) error
    IsTokenBlacklisted(ctx context.Context, jti string) (bool, error)
    CreatePasswordResetToken(ctx context.Context, token *entity.PasswordResetToken) error
    FindPasswordResetToken(ctx context.Context, token string) (*entity.PasswordResetToken, error)
    MarkPasswordResetTokenUsed(ctx context.Context, tokenID uuid.UUID) error
    UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
  }
  ```
- [x] Buat implementasi `backend/internal/repository/auth_repository_impl.go` dengan GORM

### Auth Use Case
- [x] Buat `backend/internal/usecase/auth_usecase.go`:
  - `Login(email, password string)` → validate credentials, check account lock, generate tokens
  - `Logout(tokenJTI string, expiresAt time.Time)` → blacklist token
  - `RefreshToken(refreshToken string)` → validate, rotate, return new access token
  - `GetMe(userID uuid.UUID)` → return user data
  - `ForgotPassword(email string)` → generate reset token, send email
  - `ResetPassword(token, newPassword string)` → validate token, update password

### Auth Handler (HTTP)
- [x] Buat `backend/internal/handler/auth_handler.go` dengan Gin:

  **POST `/api/v1/auth/login`**
  - Request: `{ "email": "...", "password": "..." }`
  - Response sukses (200):
    ```json
    {
      "success": true,
      "data": {
        "access_token": "eyJ...",
        "refresh_token": "eyJ...",
        "expires_in": 86400,
        "user": {
          "id": "uuid",
          "email": "...",
          "full_name": "...",
          "role": "mahasiswa",
          "must_change_password": false
        }
      }
    }
    ```
  - Response gagal (401): `{ "success": false, "message": "Email atau password salah" }`
  - Response locked (403): `{ "success": false, "message": "Akun terkunci. Coba lagi dalam X menit" }`

  **POST `/api/v1/auth/logout`** _(requires auth)_
  - Ekstrak JTI dari token, blacklist
  - Response: `{ "success": true, "message": "Berhasil logout" }`

  **POST `/api/v1/auth/refresh`**
  - Request: `{ "refresh_token": "eyJ..." }`
  - Response: `{ "success": true, "data": { "access_token": "...", "expires_in": 86400 } }`

  **GET `/api/v1/auth/me`** _(requires auth)_
  - Response: `{ "success": true, "data": { user object } }`

  **POST `/api/v1/auth/forgot-password`**
  - Request: `{ "email": "..." }`
  - Response: `{ "success": true, "message": "Jika email terdaftar, tautan reset telah dikirim" }`
  - Selalu return 200 untuk prevent email enumeration

  **POST `/api/v1/auth/reset-password`**
  - Request: `{ "token": "...", "new_password": "...", "confirm_password": "..." }`
  - Validasi: token exist, belum expired, belum used, password match, min 8 char
  - Response: `{ "success": true, "message": "Password berhasil diubah" }`

### JWT Helper (`backend/pkg/jwt/jwt.go`)
- [x] `GenerateAccessToken(userID, role, email string) (string, string, error)` — return (token, jti, error)
- [x] `GenerateRefreshToken(userID string) (string, error)`
- [x] `ValidateToken(tokenString string) (*Claims, error)`
- [x] JWT Claims struct:
  ```go
  type Claims struct {
    UserID string `json:"user_id"`
    Role   string `json:"role"`
    Email  string `json:"email"`
    JTI    string `json:"jti"`  // unique token ID untuk blacklist
    jwt.RegisteredClaims
  }
  ```

### Middleware

#### `backend/internal/middleware/auth.go` — AuthMiddleware
- [x] Ekstrak token dari header `Authorization: Bearer <token>`
- [x] Validasi signature dan expiry
- [x] Cek token tidak ada di `token_blacklist`
- [x] Inject ke Gin context: `c.Set("userID", claims.UserID)`, `c.Set("userRole", claims.Role)`, `c.Set("userEmail", claims.Email)`
- [x] Return `401 Unauthorized` jika token invalid/expired/blacklisted

#### `backend/internal/middleware/rbac.go` — RoleMiddleware
- [x] `RequireRole(roles ...string) gin.HandlerFunc`
- [x] Cek role dari context (di-set AuthMiddleware)
- [x] Return `403 Forbidden` jika role tidak diizinkan
- [x] Response format: `{ "success": false, "message": "Akses ditolak: role tidak diizinkan" }`

#### `backend/internal/middleware/rate_limit.go` — RateLimitMiddleware
- [x] Rate limit pada `POST /api/v1/auth/login`: max 10 req/menit per IP
- [x] Gunakan in-memory map dengan sliding window (cukup untuk 100 user v1.0)
- [x] Return `429 Too Many Requests` jika limit terlampaui

#### `backend/internal/middleware/security.go` — SecurityHeadersMiddleware
- [x] Tambah headers:
  ```
  X-Content-Type-Options: nosniff
  X-Frame-Options: DENY
  X-XSS-Protection: 1; mode=block
  Referrer-Policy: strict-origin-when-cross-origin
  ```
- [x] CORS: konfigurasi dari env `CORS_ALLOWED_ORIGINS`

### Route Registration
- [x] Buat `backend/internal/handler/router.go` — setup semua route dengan groups:
  ```go
  v1 := router.Group("/api/v1")

  // Public routes
  auth := v1.Group("/auth")
  auth.POST("/login", authHandler.Login)
  auth.POST("/refresh", authHandler.RefreshToken)
  auth.POST("/forgot-password", authHandler.ForgotPassword)
  auth.POST("/reset-password", authHandler.ResetPassword)

  // Protected routes
  protected := v1.Group("/", middleware.AuthMiddleware())
  protected.POST("/auth/logout", authHandler.Logout)
  protected.GET("/auth/me", authHandler.GetMe)

  // Admin only
  adminOnly := v1.Group("/admin",
    middleware.AuthMiddleware(),
    middleware.RequireRole("admin_fakultas"),
  )
  // (endpoint admin ditambahkan di job berikutnya)

  // Admin or Kaprodi
  adminKaprodi := v1.Group("/",
    middleware.AuthMiddleware(),
    middleware.RequireRole("admin_fakultas", "kaprodi"),
  )
  ```

### RBAC Matrix Lengkap

| Endpoint | Admin | Kaprodi | Mahasiswa | Dos. Pembimbing | Dos. Penguji |
|----------|:-----:|:-------:|:---------:|:---------------:|:------------:|
| POST /auth/login | ✅ | ✅ | ✅ | ✅ | ✅ |
| GET /auth/me | ✅ | ✅ | ✅ | ✅ | ✅ |
| POST /admin/users | ✅ | ❌ | ❌ | ❌ | ❌ |
| GET /admin/users | ✅ | ❌ | ❌ | ❌ | ❌ |
| GET /theses (all) | ✅ | ✅ | ❌ | ❌ | ❌ |
| POST /theses | ❌ | ❌ | ✅ | ❌ | ❌ |
| PUT /theses/:id/approve | ❌ | ✅ | ❌ | ❌ | ❌ |
| POST /documents | ❌ | ❌ | ✅ | ❌ | ❌ |
| PATCH /documents/:id/review | ❌ | ❌ | ❌ | ✅ | ❌ |
| POST /seminars/:id/scores | ❌ | ❌ | ❌ | ❌ | ✅ |
| PUT /seminars/:id/schedule | ✅ | ✅ | ❌ | ❌ | ❌ |
| PUT /defenses/:id/graduation | ❌ | ✅ | ❌ | ❌ | ❌ |
| GET /dashboard/summary | ✅ | ✅ | ❌ | ❌ | ❌ |

### Security — Password Policy
- [x] Validasi password baru: minimal 8 karakter, minimal 1 huruf kapital, 1 angka
- [x] Gunakan `bcrypt` dengan cost factor 12
- [x] Jangan log atau expose password hash di response/log

---

## Done Criteria

- [x] `POST /api/v1/auth/login` credential valid → return access token + refresh token
- [x] `POST /api/v1/auth/login` credential salah → `401 Unauthorized`
- [x] `POST /api/v1/auth/login` gagal 5x → akun terkunci, response `403` dengan pesan lock
- [x] `GET /api/v1/auth/me` dengan valid token → return user data
- [x] `GET /api/v1/auth/me` tanpa token → `401 Unauthorized`
- [x] `GET /api/v1/auth/me` dengan expired token → `401 Unauthorized`
- [x] `POST /api/v1/auth/logout` → token di-blacklist
- [x] Token yang sudah di-blacklist → `401 Unauthorized` jika digunakan lagi
- [x] Akses endpoint `/admin/users` dengan token role `mahasiswa` → `403 Forbidden`
- [x] Alur forgot password + reset password berhasil end-to-end
- [x] Rate limiting: lebih dari 10 request login/menit dari 1 IP → `429`
- [x] Security headers muncul di semua response
