# Job 28 — User Acceptance Testing (UAT)

**Phase:** 7 — Launch
**Referensi PRD:** Section 22 (Acceptance Criteria), Section 6 (Functional Requirements)
**Prerequisites:** Job 27 (Monitoring & Logging) ✅
**Estimasi:** 3–5 hari (melibatkan pengguna nyata)

---

## Objective

Validasi sistem secara end-to-end oleh pengguna nyata (atau role-play tim) sebelum go-live resmi. Setiap skenario UAT harus lulus sepenuhnya. Temuan bug dicatat dan diperbaiki sebelum melanjutkan ke Job 30.

---

## Checklist

### Persiapan UAT

- [ ] Siapkan environment UAT (bisa gunakan production atau staging Vercel preview)
- [ ] Siapkan akun test untuk setiap role (lihat Job 29 — Seed Data)
- [ ] Bagikan URL + kredensial ke tester (Admin Fakultas, Kaprodi, 2 Mahasiswa, 2 Dosen Pembimbing, 2 Dosen Penguji)
- [ ] Siapkan form pelaporan bug: Google Form atau GitHub Issues dengan template
- [ ] Tentukan batas waktu: temuan bug harus dilaporkan dalam X hari

---

## Skenario UAT

### BLOK A — Autentikasi

| No | Skenario | Langkah | Expected Result | Status |
|----|----------|---------|-----------------|--------|
| A1 | Login valid | Masukkan email + password benar | Masuk ke dashboard sesuai role | ⬜ |
| A2 | Login gagal | Masukkan password salah | Pesan error, form tidak reset | ⬜ |
| A3 | Akun terkunci | Login salah 5x berturut-turut | Pesan akun terkunci, tidak bisa login | ⬜ |
| A4 | Lupa password | Klik "Lupa Password", masukkan email | Email reset terkirim ke inbox | ⬜ |
| A5 | Reset password | Klik link di email, set password baru | Password berhasil diubah, bisa login | ⬜ |
| A6 | Login pertama | Login dengan akun baru (must_change_password) | Diarahkan ke halaman ganti password | ⬜ |
| A7 | Logout | Klik logout | Kembali ke halaman login, token invalid | ⬜ |
| A8 | Akses paksa | Akses `/admin/users` sebagai mahasiswa | Redirect atau 403 | ⬜ |

### BLOK B — Admin Fakultas

| No | Skenario | Langkah | Expected Result | Status |
|----|----------|---------|-----------------|--------|
| B1 | Buat user mahasiswa | Admin → Manajemen User → Tambah | User dibuat, welcome email terkirim | ⬜ |
| B2 | Import 10 mahasiswa | Upload file Excel template yang sudah diisi | 10 user dibuat, laporan sukses/gagal tampil | ⬜ |
| B3 | Import dengan baris error | Upload Excel dengan 1 baris email duplikat | 9 berhasil, 1 error dilaporkan per baris | ⬜ |
| B4 | Nonaktifkan user | Admin → klik Nonaktifkan pada user aktif | User tidak bisa login, badge merah | ⬜ |
| B5 | Reset password user | Admin → Reset Password pada user | Email password baru terkirim ke user | ⬜ |
| B6 | Tambah tahun akademik | Admin → Tahun Akademik → Tambah | Tahun akademik baru tersedia | ⬜ |
| B7 | Aktifkan tahun akademik | Klik Aktifkan pada tahun akademik baru | Yang lama nonaktif, yang baru aktif | ⬜ |
| B8 | Jadwalkan seminar | Admin → Jadwal → Jadwalkan Seminar (ada pengajuan pending) | Jadwal tersimpan, email terkirim ke semua pihak | ⬜ |
| B9 | Lihat audit log | Admin → Audit Log | Semua aksi dari skenario sebelumnya tercatat | ⬜ |

### BLOK C — Mahasiswa (Alur Penuh End-to-End)

| No | Skenario | Langkah | Expected Result | Status |
|----|----------|---------|-----------------|--------|
| C1 | Ajukan judul | Mahasiswa → Skripsi Saya → Ajukan Judul | Judul terkirim, status "Menunggu Review" | ⬜ |
| C2 | Judul ditolak | Kaprodi tolak judul → cek akun mahasiswa | Status "Ditolak", catatan Kaprodi tampil | ⬜ |
| C3 | Ajukan ulang | Mahasiswa submit judul baru setelah ditolak | Bisa submit ulang | ⬜ |
| C4 | Judul disetujui | Kaprodi approve + tetapkan pembimbing | Status "Bimbingan", info pembimbing tampil | ⬜ |
| C5 | Catat bimbingan | Mahasiswa → Bimbingan → Catat Bimbingan | Log bimbingan tersimpan, status "Pending" | ⬜ |
| C6 | Upload proposal | Mahasiswa → Dokumen → Upload Proposal (PDF) | Status "Menunggu Review" | ⬜ |
| C7 | Upload bukan PDF | Coba upload file .docx | Pesan error format file | ⬜ |
| C8 | Upload >10MB | Coba upload PDF 15MB | Pesan error ukuran file | ⬜ |
| C9 | Ajukan seminar sebelum waktunya | Klik ajukan seminar sebelum dokumen disetujui | Pesan error gate, tombol disabled | ⬜ |
| C10 | Ajukan seminar | Setelah dokumen seminar disetujui → ajukan seminar | Pengajuan terkirim, status "Menunggu Jadwal" | ⬜ |
| C11 | Lihat jadwal seminar | Setelah dijadwalkan → cek halaman seminar | Info tanggal, waktu, ruangan, penguji tampil | ⬜ |
| C12 | Lihat hasil seminar | Setelah penguji input nilai → cek hasil | Nilai akhir dan breakdown tampil | ⬜ |
| C13 | Alur sidang | Ulangi C9–C12 untuk sidang skripsi | Alur sidang berfungsi sama | ⬜ |
| C14 | Arsip skripsi | Setelah yudisium → upload skripsi final | Arsip tersimpan, bisa dicari dan didownload | ⬜ |

### BLOK D — Dosen Pembimbing

| No | Skenario | Langkah | Expected Result | Status |
|----|----------|---------|-----------------|--------|
| D1 | Lihat mahasiswa bimbingan | Dashboard → Mahasiswa Bimbingan | List mahasiswa dengan status masing-masing | ⬜ |
| D2 | Approve log bimbingan | Tab Bimbingan mahasiswa → Setujui log | Status berubah "Approved", counter bertambah | ⬜ |
| D3 | Review dokumen — setujui | Tab Dokumen → Review → Setujui | Status dokumen "Disetujui" | ⬜ |
| D4 | Review dokumen — revisi | Tab Dokumen → Review → Minta Revisi + catatan | Status "Perlu Revisi", mahasiswa bisa upload ulang | ⬜ |
| D5 | Download dokumen mahasiswa | Klik tombol Lihat/Download | File PDF terbuka di tab baru | ⬜ |
| D6 | Lihat jadwal seminar | Supervisor → Jadwal | Jadwal seminar mahasiswa bimbingan tampil | ⬜ |

### BLOK E — Dosen Penguji

| No | Skenario | Langkah | Expected Result | Status |
|----|----------|---------|-----------------|--------|
| E1 | Lihat jadwal pengujian | Dashboard → Jadwal | Seminar/sidang yang ditugaskan tampil | ⬜ |
| E2 | Input nilai seminar | Klik Input Nilai → isi semua komponen | Preview nilai real-time terhitung | ⬜ |
| E3 | Konfirmasi sebelum submit nilai | Klik Simpan Nilai | Modal konfirmasi muncul | ⬜ |
| E4 | Nilai tersimpan | Setelah konfirmasi → nilai tersimpan | Form menjadi read-only | ⬜ |
| E5 | Input nilai penguji kedua | Login penguji kedua → input nilai yang berbeda | Nilai akhir = rata-rata berbobot keduanya | ⬜ |
| E6 | Coba input nilai lagi | Penguji yang sudah submit coba input ulang | Halaman read-only, tidak bisa input ulang | ⬜ |

### BLOK F — Kaprodi

| No | Skenario | Langkah | Expected Result | Status |
|----|----------|---------|-----------------|--------|
| F1 | Review pengajuan judul | Kaprodi → Pengajuan Judul → Review | Modal dengan info lengkap muncul | ⬜ |
| F2 | Tolak dengan catatan | Pilih Tolak, isi catatan kosong → submit | Validasi error: catatan wajib diisi | ⬜ |
| F3 | Setujui + tetapkan pembimbing | Approve → modal penunjukan pembimbing | Pembimbing berhasil ditetapkan | ⬜ |
| F4 | Jadwalkan seminar | Kaprodi → Seminar → Jadwalkan | Modal penjadwalan berfungsi | ⬜ |
| F5 | Tetapkan yudisium | Setelah sidang lulus → Tetapkan Yudisium | Konfirmasi muncul, status → Lulus | ⬜ |
| F6 | Lihat dashboard statistik | Dashboard Kaprodi | Semua widget dan chart tampil dengan data akurat | ⬜ |
| F7 | Filter dashboard | Ganti filter tahun akademik | Semua widget ter-update | ⬜ |

### BLOK G — Non-Functional

| No | Skenario | Cara Test | Expected Result | Status |
|----|----------|-----------|-----------------|--------|
| G1 | Responsive mobile | Buka di Chrome DevTools → iPhone 12 | Layout tidak rusak | ⬜ |
| G2 | Responsive tablet | Buka di Chrome DevTools → iPad | Layout tidak rusak | ⬜ |
| G3 | Email notification | Trigger salah satu event | Email masuk ke inbox (cek spam juga) | ⬜ |
| G4 | Loading state | Buka halaman dengan koneksi lambat (throttle) | Skeleton loader tampil, bukan blank | ⬜ |
| G5 | Error handling | Matikan backend sementara → akses halaman | Pesan error informatif, bukan stack trace | ⬜ |
| G6 | Pencarian arsip | Cari dengan kata kunci dari abstrak | Hasil relevan muncul | ⬜ |

---

## Pelaporan Bug

Setiap bug yang ditemukan dicatat dengan format:

```
ID: UAT-XXX
Skenario: [nomor skenario]
Role: [role yang menemukan bug]
Deskripsi: [apa yang terjadi]
Expected: [seharusnya apa yang terjadi]
Actual: [yang terjadi sebenarnya]
Screenshot/Video: [lampirkan jika ada]
Severity: Critical / High / Medium / Low
```

**Kriteria go/no-go setelah UAT:**
- ✅ Semua skenario Blok A–F lulus → boleh go-live
- ✅ Tidak ada bug Critical atau High yang belum diperbaiki
- ⚠️ Bug Medium/Low boleh di-defer ke patch v1.0.1
- ❌ Ada bug Critical → wajib diperbaiki sebelum go-live

---

## Done Criteria

- [ ] Semua 50+ skenario UAT dijalankan
- [ ] Status semua skenario: ✅ Pass
- [ ] Tidak ada bug Critical atau High yang terbuka
- [ ] Bug Medium/Low dicatat di GitHub Issues
- [ ] Sign-off dari Kaprodi atau Admin Fakultas sebagai perwakilan pengguna
