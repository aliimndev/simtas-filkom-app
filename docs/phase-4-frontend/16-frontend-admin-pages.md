# Job 16 — Frontend Admin Pages

**Phase:** 4 — Frontend
**Referensi PRD:** Section 6.1 (FR-USER-001 s/d FR-USER-004), Section 5 (Target Users — Admin Fakultas)
**Prerequisites:** Job 15 (Frontend Auth Pages) ✅
**Estimasi:** 3 hari

---

## Objective

Implementasi semua halaman Admin Fakultas: manajemen pengguna (CRUD, import, reset password, aktivasi), manajemen tahun akademik, dan halaman audit log. Setelah job ini selesai, Admin dapat mengelola seluruh data pengguna dan konfigurasi akademik melalui antarmuka yang lengkap.

---

## Checklist

### Halaman Manajemen User (`/admin/users`)

**File:** `frontend/src/app/(dashboard)/admin/users/page.tsx`

#### Tabel Pengguna
- [ ] Kolom: No, Nama, Email, NIM/NIDN, Role, Prodi, Status, Aksi
- [ ] Filter bar di atas tabel:
  - Search input (nama, email, NIM) dengan debounce 400ms
  - Dropdown filter Role
  - Dropdown filter Status (Aktif / Nonaktif)
  - Tombol "Reset Filter"
- [ ] Badge status: Aktif (hijau) / Nonaktif (merah)
- [ ] Badge role dengan warna berbeda per role
- [ ] Pagination di bawah tabel
- [ ] Tombol aksi per baris: Edit, Reset Password, Aktifkan/Nonaktifkan, Hapus
- [ ] Konfirmasi modal sebelum hapus dan nonaktifkan
- [ ] Tombol "Tambah User" dan "Import CSV/Excel" di header halaman
- [ ] Loading skeleton saat data di-fetch

#### TanStack Query Hooks
- [ ] `useUsers(filter)` — query list users dengan filter
- [ ] `useCreateUser()` — mutation create
- [ ] `useUpdateUser()` — mutation update
- [ ] `useDeleteUser()` — mutation soft delete
- [ ] `useToggleUserActive()` — mutation activate/deactivate
- [ ] `useResetUserPassword()` — mutation reset password
- [ ] Setelah setiap mutation sukses → invalidate query `users.list`

### Modal Tambah / Edit User

**File:** `frontend/src/components/features/users/UserFormModal.tsx`

- [ ] Modal dengan form React Hook Form + Zod:
  ```ts
  const userSchema = z.object({
    email: z.string().email(),
    full_name: z.string().min(3, "Minimal 3 karakter"),
    nim_nidn: z.string().optional(),
    role: z.enum(["admin_fakultas","kaprodi","mahasiswa","dosen_pembimbing","dosen_penguji"]),
    study_program: z.string().optional(),
  })
  ```
- [ ] Mode Create: semua field kosong, tombol "Simpan"
- [ ] Mode Edit: form di-prefill dengan data user, email tidak bisa diubah (disabled)
- [ ] Sukses create → toast "User berhasil dibuat. Password sementara telah dikirim ke email."
- [ ] Sukses edit → toast "Data user berhasil diperbarui."

### Modal Import User

**File:** `frontend/src/components/features/users/ImportUserModal.tsx`

- [ ] Step 1 — Upload:
  - Komponen FileUpload untuk file `.csv` atau `.xlsx`
  - Link download template: tombol "Download Template Excel"
  - Tombol "Proses Import"
- [ ] Step 2 — Hasil Import (setelah response):
  - Summary: "47 user berhasil dibuat, 3 baris gagal"
  - Jika ada error → tabel error: kolom Baris, Email, Alasan Gagal
  - Tombol "Tutup"
- [ ] Loading state saat upload + proses

### Halaman Detail User (`/admin/users/:id`)

**File:** `frontend/src/app/(dashboard)/admin/users/[id]/page.tsx`

- [ ] Info lengkap user: semua field + tanggal dibuat, last login
- [ ] Riwayat audit log untuk user ini (ambil dari `GET /admin/audit-logs?user_id=xxx`)
- [ ] Tombol aksi: Edit, Reset Password, Aktivasi/Deaktivasi
- [ ] Breadcrumb: Admin > Manajemen User > Nama User

### Halaman Manajemen Tahun Akademik (`/admin/academic-years`)

**File:** `frontend/src/app/(dashboard)/admin/academic-years/page.tsx`

- [ ] Tabel: Nama, Semester, Tanggal Mulai, Tanggal Selesai, Status (Aktif/Nonaktif), Aksi
- [ ] Tombol "Tambah Tahun Akademik"
- [ ] Hanya 1 yang boleh aktif — badge "Aktif" di row yang aktif
- [ ] Aksi per row: Edit, Aktifkan (jika belum aktif)
- [ ] Konfirmasi saat mengaktifkan: "Mengaktifkan tahun akademik ini akan menonaktifkan yang sedang aktif. Lanjutkan?"

#### Modal Tambah/Edit Tahun Akademik
- [ ] Form: Nama (contoh: "2026/2027"), Semester (dropdown: Ganjil/Genap), Tanggal Mulai, Tanggal Selesai
- [ ] Validasi Zod: tanggal selesai harus setelah tanggal mulai

### Halaman Jadwal (Admin) (`/admin/schedules`)

**File:** `frontend/src/app/(dashboard)/admin/schedules/page.tsx`

- [ ] Tab: "Seminar Proposal" | "Sidang Skripsi"
- [ ] Tabel jadwal mendatang dan yang sudah lewat
- [ ] Kolom: Mahasiswa, Judul (truncated), Tanggal & Waktu, Ruangan, Penguji, Status, Aksi
- [ ] Tombol "Jadwalkan Seminar" / "Jadwalkan Sidang" (buka modal)
- [ ] Filter: Status, Tanggal

#### Modal Penjadwalan Seminar/Sidang

**File:** `frontend/src/components/features/schedules/ScheduleModal.tsx`

- [ ] Step 1: Pilih mahasiswa dari dropdown (thesis dengan status sesuai)
- [ ] Step 2: Isi jadwal:
  - Date picker untuk tanggal
  - Time picker untuk jam
  - Input ruangan
  - Multi-select penguji (dari list dosen_penguji aktif)
- [ ] Validasi: minimal 2 penguji, tanggal minimal 3 hari (seminar) atau 7 hari (sidang)
- [ ] Tampilkan warning jika ada konflik jadwal (dari response 409 backend)

### Halaman Audit Log (`/admin/audit-logs`)

**File:** `frontend/src/app/(dashboard)/admin/audit-logs/page.tsx`

- [ ] Tabel: Waktu, User, Action, Entitas, IP Address
- [ ] Filter: Action (dropdown), Tanggal Dari–Sampai, User (search)
- [ ] Klik row → modal detail dengan `old_value` dan `new_value` dalam format JSON yang readable
- [ ] Pagination dengan per_page: 50
- [ ] Badge warna per kategori action (auth, user, thesis, document, seminar, defense)

### Shared Components untuk Admin

- [ ] **`ConfirmModal`** — modal konfirmasi reusable dengan props: title, message, confirmLabel, variant (danger/warning)
- [ ] **`FilterBar`** — wrapper layout untuk filter controls di atas tabel
- [ ] **`TableActions`** — dropdown aksi per baris tabel (Edit, Delete, dll)
- [ ] **`EmptyState`** — tampilan tabel kosong dengan icon dan pesan

---

## Done Criteria

- [ ] `/admin/users` → tabel user tampil dengan pagination dan filter berfungsi
- [ ] Search user dengan debounce → hasil terfilter tanpa reload halaman
- [ ] Buat user baru via modal → user muncul di tabel, toast sukses
- [ ] Import Excel 50 baris → modal hasil tampil dengan summary dan error list
- [ ] Nonaktifkan user → badge status berubah, konfirmasi modal muncul sebelumnya
- [ ] Reset password → toast sukses, email terkirim ke user
- [ ] `/admin/academic-years` → tabel tahun akademik, aktifkan tahun baru → yang lama berubah nonaktif
- [ ] `/admin/schedules` → tabel jadwal per tab, modal penjadwalan berfungsi
- [ ] `/admin/audit-logs` → tabel audit dengan filter tanggal dan action
- [ ] Semua halaman responsive di mobile dan tablet
- [ ] Loading skeleton muncul saat fetch data
- [ ] Error state (misal API down) → tampil pesan error yang informatif, bukan blank page
