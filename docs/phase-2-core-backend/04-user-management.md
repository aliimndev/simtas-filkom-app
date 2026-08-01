# Job 04 — User Management (Backend)

**Phase:** 2 — Core Backend
**Referensi PRD:** Section 6.1 (FR-USER-001 s/d FR-USER-005)
**Prerequisites:** Job 03 (Auth & RBAC) ✅
**Estimasi:** 3 hari

---

## Objective

Implementasi seluruh fitur manajemen pengguna oleh Admin Fakultas: CRUD user, import bulk via CSV/Excel, reset password oleh admin, aktivasi/deaktivasi akun, dan manajemen tahun akademik. Setelah job ini selesai, Admin dapat mengelola semua akun pengguna sistem secara penuh.

---

## Checklist

### Dependencies Tambahan
- [x] Install library untuk parsing Excel/CSV:
  ```bash
  go get github.com/xuri/excelize/v2   # Excel
  go get github.com/gocarina/gocsv     # CSV
  ```

### User Repository & Use Case
- [x] Buat `internal/domain/repository/user_repository.go` — interface:
  ```go
  type UserRepository interface {
    FindAll(ctx context.Context, filter UserFilter) ([]*entity.User, int64, error)
    FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
    FindByEmail(ctx context.Context, email string) (*entity.User, error)
    Create(ctx context.Context, user *entity.User) error
    Update(ctx context.Context, user *entity.User) error
    SoftDelete(ctx context.Context, id uuid.UUID) error
    BulkCreate(ctx context.Context, users []*entity.User) error
    SetActiveStatus(ctx context.Context, id uuid.UUID, isActive bool) error
    ResetPassword(ctx context.Context, id uuid.UUID, passwordHash string) error
    InvalidateUserSessions(ctx context.Context, id uuid.UUID) error
  }
  ```
- [x] `UserFilter` struct: `Role`, `IsActive`, `StudyProgram`, `Search` (nama/email/NIM), `Page`, `PerPage`
- [x] Buat `internal/usecase/user_usecase.go` dengan logic:
  - Auto-generate password saat create: 12 karakter random (huruf besar + kecil + angka)
  - Send welcome email setelah create sukses
  - `InvalidateUserSessions` dipanggil saat deactivate akun

### User Handler — Endpoint Admin

**GET `/api/v1/admin/users`** _(Admin only)_
- [x] Query params: `role`, `is_active`, `study_program`, `search`, `page` (default 1), `per_page` (default 20)
- [ ] Response:
  ```json
  {
    "success": true,
    "data": [ { user objects } ],
    "meta": { "page": 1, "per_page": 20, "total": 85, "total_pages": 5 }
  }
  ```

**GET `/api/v1/admin/users/:id`** _(Admin only)_
- [x] Return detail user termasuk role name
- [x] Return `404` jika tidak ditemukan atau sudah soft-deleted

**POST `/api/v1/admin/users`** _(Admin only)_
- [x] Request body:
  ```json
  {
    "email": "mahasiswa@example.com",
    "full_name": "Nama Lengkap",
    "nim_nidn": "12345678",
    "role": "mahasiswa",
    "study_program": "Teknik Informatika"
  }
  ```
- [x] Validasi: email valid & unique, role valid, full_name tidak kosong
- [x] Auto-generate password, hash dengan bcrypt
- [x] Send welcome email dengan password sementara
- [x] Audit log: `USER_CREATED`
- [x] Response: `201 Created` dengan user object (tanpa password hash)

**PUT `/api/v1/admin/users/:id`** _(Admin only)_
- [x] Field yang bisa diupdate: `full_name`, `nim_nidn`, `study_program`, `profile_photo_url`
- [x] Email dan role tidak bisa diubah via endpoint ini (butuh endpoint khusus jika diperlukan)
- [x] Audit log: `USER_UPDATED` dengan old_value dan new_value

**DELETE `/api/v1/admin/users/:id`** _(Admin only)_
- [x] Soft delete (set `deleted_at`)
- [x] Validasi: tidak bisa hapus diri sendiri
- [x] Audit log: `USER_DELETED`

**PATCH `/api/v1/admin/users/:id/activate`** _(Admin only)_
- [x] Set `is_active = true`
- [x] Audit log: `USER_ACTIVATED`

**PATCH `/api/v1/admin/users/:id/deactivate`** _(Admin only)_
- [x] Set `is_active = false`
- [x] Invalidate semua session aktif (blacklist token via JTI lookup — atau set flag di user) — via `token_version` claim
- [x] Validasi: tidak bisa deactivate diri sendiri
- [x] Audit log: `USER_DEACTIVATED`

**POST `/api/v1/admin/users/:id/reset-password`** _(Admin only)_
- [x] Auto-generate password baru (12 karakter)
- [x] Hash dan update di database
- [x] Set `must_change_password = true`
- [x] Send email ke user dengan password baru
- [x] Audit log: `USER_PASSWORD_RESET`

### Import User

**GET `/api/v1/admin/users/import-template`** _(Admin only)_
- [x] Return file Excel template dengan kolom: `email`, `full_name`, `nim_nidn`, `role`, `study_program`
- [x] Sertakan sheet "Petunjuk" dengan keterangan nilai valid untuk kolom `role`
- [x] Content-Type: `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`

**POST `/api/v1/admin/users/import`** _(Admin only)_
- [x] Accept: `multipart/form-data` dengan field `file`
- [x] Support format: `.csv` dan `.xlsx`
- [x] Ukuran file maksimal: 5 MB
- [x] Proses parsing:
  1. Baca semua baris
  2. Validasi per baris: email valid, role valid, field wajib tidak kosong, email belum terdaftar
  3. Baris valid → bulk insert dalam 1 transaction
  4. Baris error → catat ke error list, skip
  5. Send welcome email untuk setiap user yang berhasil dibuat (async)
- [ ] Response:
  ```json
  {
    "success": true,
    "data": {
      "total_rows": 50,
      "success_count": 47,
      "error_count": 3,
      "errors": [
        { "row": 5, "email": "invalid-email", "reason": "Format email tidak valid" },
        { "row": 12, "email": "existing@example.com", "reason": "Email sudah terdaftar" },
        { "row": 23, "email": "", "reason": "Email tidak boleh kosong" }
      ]
    }
  }
  ```
- [x] Audit log: `USER_BULK_IMPORTED` dengan info jumlah success/error

### Manajemen Tahun Akademik

**GET `/api/v1/academic-years`** _(All authenticated)_
- [x] Return list semua tahun akademik, diurutkan terbaru dulu

**POST `/api/v1/academic-years`** _(Admin only)_
- [x] Request: `{ "name": "2026/2027", "semester": "ganjil", "start_date": "2026-09-01", "end_date": "2027-01-31" }`
- [x] Validasi: `semester` harus `ganjil` atau `genap`

**PUT `/api/v1/academic-years/:id`** _(Admin only)_
- [x] Update data tahun akademik (kecuali yang sudah aktif & ada thesis berjalan)

**PATCH `/api/v1/academic-years/:id/activate`** _(Admin only)_
- [x] Set `is_active = true` pada tahun akademik ini
- [x] Set `is_active = false` pada semua tahun akademik lain (hanya 1 aktif)

### Helper: Pagination
- [x] Buat `pkg/pagination/pagination.go`:
  ```go
  type Pagination struct {
    Page      int   `json:"page"`
    PerPage   int   `json:"per_page"`
    Total     int64 `json:"total"`
    TotalPages int  `json:"total_pages"`
  }

  func NewPagination(page, perPage int, total int64) Pagination
  func (p *Pagination) Offset() int
  ```

### Helper: Auto-generate Password
- [x] Buat `pkg/utils/password.go`:
  ```go
  func GenerateRandomPassword(length int) string
  // Karakter: huruf besar + kecil + angka, min 1 dari tiap kategori
  ```

---

## Done Criteria

- [x] `POST /api/v1/admin/users` → buat user, welcome email terkirim, response 201
- [x] `GET /api/v1/admin/users?search=ali&role=mahasiswa&page=1` → hasil terfilter + pagination benar
- [x] `PATCH /api/v1/admin/users/:id/deactivate` → user tidak bisa login (token_version invalidasi)
- [x] `POST /api/v1/admin/users/import` dengan file 50 baris (3 error) → 47 user dibuat, 3 error dilaporkan
- [x] `GET /api/v1/admin/users/import-template` → file Excel template berhasil didownload
- [x] `PATCH /api/v1/academic-years/:id/activate` → hanya 1 tahun akademik yang `is_active = true`
- [x] Semua operasi tercatat di `audit_logs`
- [x] Endpoint admin tidak bisa diakses oleh role non-admin → `403 Forbidden` (via `RequireRole` middleware)
