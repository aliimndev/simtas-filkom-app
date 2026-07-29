# Job 15 — Frontend Auth Pages

**Phase:** 4 — Frontend
**Referensi PRD:** Section 6.1 (FR-AUTH-001 s/d FR-AUTH-004, FR-USER-005)
**Prerequisites:** Job 14 (Frontend Setup) ✅
**Estimasi:** 2 hari

---

## Objective

Implementasi semua halaman autentikasi: login, forgot password, reset password, dan halaman profil/ganti password pengguna. Setelah job ini selesai, seluruh alur autentikasi frontend berfungsi penuh dan terhubung ke backend API.

---

## Checklist

### Halaman Login (`/login`)

**File:** `src/app/(auth)/login/page.tsx`

- [ ] Form dengan React Hook Form + Zod validation:
  ```ts
  const loginSchema = z.object({
    email: z.string().email("Format email tidak valid"),
    password: z.string().min(1, "Password tidak boleh kosong"),
  })
  ```
- [ ] Layout:
  - Card centered di tengah layar dengan background subtle
  - Logo FILKOM Unida di atas form
  - Judul: "SIMTAS FILKOM" + subtitle "Sistem Manajemen Tugas Akhir dan Skripsi"
  - Field: Email, Password (dengan toggle show/hide)
  - Tombol "Masuk" dengan loading state saat submit
  - Link "Lupa Password?" di bawah tombol
- [ ] Submit handler menggunakan TanStack Query `useMutation`:
  ```ts
  const loginMutation = useMutation({
    mutationFn: (data: LoginRequest) => authApi.login(data),
    onSuccess: (data) => {
      // Simpan token ke store
      // Redirect berdasarkan role:
      //   admin_fakultas → /dashboard
      //   kaprodi → /dashboard
      //   mahasiswa → /dashboard
      //   dosen_pembimbing → /dashboard
      //   dosen_penguji → /dashboard
      // Jika must_change_password === true → redirect /change-password
    },
    onError: (error) => {
      toast.error(getErrorMessage(error)) // "Email atau password salah"
    },
  })
  ```
- [ ] Handling error spesifik:
  - 401 → "Email atau password salah"
  - 403 (locked) → "Akun Anda terkunci. Silakan hubungi Admin Fakultas."
  - 429 → "Terlalu banyak percobaan. Tunggu beberapa menit."
- [ ] Jika sudah login → redirect otomatis ke `/dashboard`
- [ ] Responsive: tampil baik di mobile, tablet, desktop

### Halaman Lupa Password (`/forgot-password`)

**File:** `src/app/(auth)/forgot-password/page.tsx`

- [ ] Form:
  ```ts
  const forgotSchema = z.object({
    email: z.string().email("Format email tidak valid"),
  })
  ```
- [ ] Setelah submit sukses → tampilkan success state (bukan form lagi):
  ```
  ✅ Tautan reset password telah dikirim ke email Anda.
     Silakan cek inbox atau folder spam.
  ```
- [ ] Tombol "Kembali ke Login"
- [ ] Note: response selalu 200 dari backend (prevent email enumeration), tampilkan pesan yang sama apapun hasilnya

### Halaman Reset Password (`/reset-password?token=xxx`)

**File:** `src/app/(auth)/reset-password/page.tsx`

- [ ] Baca `token` dari query parameter saat halaman mount
- [ ] Jika tidak ada token → redirect ke `/forgot-password`
- [ ] Form:
  ```ts
  const resetSchema = z.object({
    new_password: z.string()
      .min(8, "Minimal 8 karakter")
      .regex(/[A-Z]/, "Harus ada huruf kapital")
      .regex(/[0-9]/, "Harus ada angka"),
    confirm_password: z.string(),
  }).refine((d) => d.new_password === d.confirm_password, {
    message: "Konfirmasi password tidak cocok",
    path: ["confirm_password"],
  })
  ```
- [ ] Password strength indicator visual (weak/medium/strong)
- [ ] Setelah sukses → toast "Password berhasil diubah" + redirect ke `/login` setelah 2 detik
- [ ] Jika token invalid/expired → tampilkan error state dengan link ke `/forgot-password`

### Halaman Ganti Password Pertama Kali (`/change-password`)

**File:** `src/app/(dashboard)/change-password/page.tsx`

- [ ] Muncul otomatis jika `must_change_password === true` setelah login
- [ ] Redirect paksa: jika `must_change_password === true`, akses halaman lain selalu diarahkan ke sini dulu
- [ ] Layout: halaman full (bukan modal), penjelasan singkat "Ini adalah login pertama Anda. Harap ganti password default."
- [ ] Form sama seperti reset password
- [ ] Setelah sukses → update `must_change_password = false` di auth store → redirect ke `/dashboard`

### Halaman Profil (`/profile`)

**File:** `src/app/(dashboard)/profile/page.tsx`

- [ ] Tampilkan info user:
  - Avatar / inisial nama
  - Nama lengkap, email, NIM/NIDN, program studi, role
  - Tombol "Edit Profil" (hanya untuk field tertentu: nama, foto profil)
- [ ] Section Ubah Password:
  ```ts
  const changePasswordSchema = z.object({
    current_password: z.string().min(1),
    new_password: z.string().min(8).regex(/[A-Z]/).regex(/[0-9]/),
    confirm_password: z.string(),
  }).refine(...)
  ```
- [ ] Submit ubah password → call `PUT /api/v1/users/me/password`
  - Sukses → toast "Password berhasil diubah", clear form
  - Error (wrong current password) → "Password lama tidak sesuai"

### TanStack Query — Auth Queries

Buat `src/lib/api/hooks/useAuth.ts`:
- [ ] `useCurrentUser()` — query `GET /auth/me`, enabled saat ada token
- [ ] `useLoginMutation()` — mutation login
- [ ] `useLogoutMutation()` — mutation logout (invalidate semua query cache + clear auth store)
- [ ] `useForgotPasswordMutation()`
- [ ] `useResetPasswordMutation()`

### Helper Error Message

- [ ] Buat `src/lib/utils/error.ts`:
  ```ts
  export function getErrorMessage(error: unknown): string {
    if (axios.isAxiosError(error)) {
      return error.response?.data?.message ?? "Terjadi kesalahan. Silakan coba lagi."
    }
    return "Terjadi kesalahan yang tidak diketahui."
  }
  ```

---

## Done Criteria

- [ ] Login dengan credential valid → redirect ke `/dashboard` sesuai role
- [ ] Login gagal → toast error dengan pesan yang sesuai
- [ ] Login saat akun terkunci → pesan spesifik tentang akun terkunci
- [ ] `/forgot-password` → form submit → success state tampil
- [ ] `/reset-password?token=valid` → ganti password → redirect ke login
- [ ] `/reset-password?token=expired` → tampil error state dengan link ke forgot-password
- [ ] Login pertama (`must_change_password=true`) → redirect paksa ke `/change-password`
- [ ] Setelah ganti password pertama → bisa akses halaman lain
- [ ] `/profile` → info user tampil, ubah password berfungsi
- [ ] Semua form menampilkan error validasi inline (bukan alert)
- [ ] Tombol submit menampilkan spinner saat loading
- [ ] Responsive di semua ukuran layar
