# Job 29 — Seed Data Production

**Phase:** 7 — Launch
**Referensi PRD:** Section 5 (Target Users & Roles)
**Prerequisites:** Job 28 (UAT Checklist selesai dan sign-off) ✅
**Estimasi:** 1 hari

---

## Objective

Menyiapkan data awal production yang diperlukan sebelum sistem digunakan: role, tahun akademik aktif, akun Admin dan Kaprodi, serta akun dosen pembimbing dan penguji. Setelah job ini selesai, sistem production siap digunakan oleh pengguna nyata.

---

## Checklist

### Strategi Seed Data

- [ ] Pisahkan seed data antara **development** (data dummy banyak) dan **production** (data minimal dan nyata)
- [ ] Buat script seed production: `backend/migrations/seeds/production/`
- [ ] Script production tidak boleh berisi data dummy atau password default yang lemah
- [ ] Jalankan seed hanya sekali (idempotent — aman dijalankan ulang tanpa duplikat data)

### 1. Roles (sudah ada dari Job 02, verifikasi saja)

- [ ] Verifikasi 5 role sudah ada di tabel `roles`:
  ```sql
  SELECT * FROM roles ORDER BY id;
  -- 1: admin_fakultas
  -- 2: kaprodi
  -- 3: mahasiswa
  -- 4: dosen_pembimbing
  -- 5: dosen_penguji
  ```

### 2. Tahun Akademik Production

- [ ] Insert tahun akademik aktif untuk awal penggunaan:
  ```sql
  INSERT INTO academic_years (id, name, semester, start_date, end_date, is_active)
  VALUES (
    gen_random_uuid(),
    '2026/2027',
    'ganjil',
    '2026-09-01',
    '2027-01-31',
    true
  )
  ON CONFLICT DO NOTHING;
  ```
- [ ] Sesuaikan dengan kalender akademik FILKOM Unida yang berlaku

### 3. Akun Admin Fakultas

- [ ] Buat minimal 2 akun Admin Fakultas (backup jika 1 lupa password):
  ```
  Admin 1: admin@filkom.unida.ac.id
  Admin 2: admin2@filkom.unida.ac.id (atau email teknisi IT kampus)
  ```
- [ ] Password: generate strong password (16 karakter, random)
- [ ] Set `must_change_password = true` agar admin ganti password saat pertama login
- [ ] Catat password sementara dan sampaikan ke admin yang bersangkutan secara aman (jangan via chat publik)
- [ ] Script:
  ```go
  // Jalankan via: go run cmd/seed/main.go --env=production
  adminPassword := generateStrongPassword(16)
  hash, _ := bcrypt.GenerateFromPassword([]byte(adminPassword), 12)
  db.Create(&User{
      Email:               "admin@filkom.unida.ac.id",
      PasswordHash:        string(hash),
      FullName:            "Administrator FILKOM",
      RoleID:              1, // admin_fakultas
      IsActive:            true,
      MustChangePassword:  true,
  })
  fmt.Printf("Password admin: %s\n", adminPassword) // print sekali, jangan log ke file
  ```

### 4. Akun Kaprodi

- [ ] Buat akun Kaprodi sesuai data nyata:
  ```
  Email: [email resmi Kaprodi FILKOM]
  Nama:  [Nama lengkap Kaprodi]
  NIDN:  [NIDN Kaprodi]
  ```
- [ ] Sama seperti admin: `must_change_password = true`, password kuat

### 5. Akun Dosen Pembimbing & Penguji

Dua opsi:

**Opsi A — Import via Excel (Direkomendasikan)**
- [ ] Admin Fakultas login → Manajemen User → Import CSV/Excel
- [ ] Siapkan file Excel dengan data lengkap semua dosen:
  - Dosen Pembimbing: nama, NIDN, email, prodi
  - Dosen Penguji: nama, NIDN, email, prodi
  - (Satu dosen bisa punya role dosen_pembimbing saja, atau keduanya — tergantung kebijakan)
- [ ] Sistem auto-generate password dan kirim welcome email

**Opsi B — Input Manual satu per satu**
- [ ] Gunakan jika jumlah dosen sedikit (<10) atau import tidak memungkinkan

### 6. Akun Mahasiswa

- [ ] Mahasiswa **tidak** di-seed secara massal di awal production
- [ ] Mahasiswa didaftarkan oleh Admin saat pertama kali akan menggunakan sistem
- [ ] Opsi pendaftaran awal:
  - Admin import dari data SIAKAD (manual export ke Excel)
  - Atau Admin input satu per satu saat mahasiswa datang ke fakultas

### 7. Verifikasi Data Awal

Setelah semua seed selesai, verifikasi:
- [ ] `SELECT COUNT(*) FROM roles;` → 5 rows
- [ ] `SELECT * FROM academic_years WHERE is_active = true;` → 1 row
- [ ] `SELECT email, role_id FROM users WHERE role_id IN (1,2);` → minimal 3 rows (2 admin + 1 kaprodi)
- [ ] Login dengan akun admin → berhasil + diarahkan ganti password
- [ ] Login dengan akun kaprodi → berhasil + diarahkan ganti password

### 8. Data untuk UAT (Jika UAT di Environment Terpisah)

Jika UAT menggunakan environment staging (bukan production), siapkan data UAT tambahan:

- [ ] 2 akun mahasiswa test: `mhs.test1@filkom.unida.ac.id`, `mhs.test2@filkom.unida.ac.id`
- [ ] 2 akun dosen pembimbing test
- [ ] 2 akun dosen penguji test
- [ ] 1 thesis yang sudah dalam status `in_progress` (untuk test alur lebih cepat)
- [ ] 1 thesis yang sudah dalam status `seminar_done` (untuk test alur sidang langsung)
- [ ] Tandai akun test dengan prefix `test.` agar mudah dibersihkan setelah UAT

### 9. Hapus Data UAT Setelah UAT Selesai

- [ ] Jika UAT dilakukan di production environment (tidak direkomendasikan tapi mungkin):
  ```sql
  -- Hapus semua data yang dibuat saat UAT
  -- Gunakan hati-hati — pastikan tidak ada data nyata yang terhapus
  DELETE FROM theses WHERE student_id IN (
    SELECT id FROM users WHERE email LIKE 'test.%'
  );
  DELETE FROM users WHERE email LIKE 'test.%';
  ```

---

## Done Criteria

- [ ] Semua roles tersedia di database production
- [ ] Tahun akademik aktif sudah dikonfigurasi sesuai kalender FILKOM
- [ ] Minimal 2 akun Admin aktif
- [ ] Minimal 1 akun Kaprodi aktif
- [ ] Semua akun dosen pembimbing & penguji sudah dibuat (via import atau manual)
- [ ] Semua akun `must_change_password = true` — tidak ada default password yang permanen
- [ ] Admin berhasil login dan menyelesaikan ganti password pertama
- [ ] Kaprodi berhasil login dan menyelesaikan ganti password pertama
- [ ] Tidak ada data dummy atau test yang tertinggal di production
