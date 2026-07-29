

---

```markdown
# Product Requirements Document (PRD)
# SIMTAS FILKOM v1.0

**Sistem Manajemen Tugas Akhir dan Skripsi**  
**Fakultas Ilmu Komputer Universitas Djuanda**

**Version:** 1.0  
**Target Release:** 2027  
**Last Updated:** 2026-07-29
**Status:** Reviewed

---

## 1. Project Overview

**SIMTAS FILKOM** adalah sistem informasi berbasis web yang dirancang untuk mengelola seluruh proses tugas akhir dan skripsi mahasiswa di Fakultas Ilmu Komputer Universitas Djuanda secara digital, terstruktur, dan terintegrasi.

Sistem ini menggantikan proses manual yang selama ini dilakukan, yang rentan terhadap:
- Kehilangan dokumen fisik
- Keterlambatan informasi dan notifikasi
- Kesulitan pelacakan progress mahasiswa
- Inkonsistensi prosedur dan penilaian
- Beban administratif yang tinggi
- Kesulitan pelaporan dan audit

SIMTAS FILKOM menyediakan platform terpusat yang mengintegrasikan seluruh stakeholder (mahasiswa, dosen pembimbing, dosen penguji, Kaprodi, dan Admin Fakultas) dalam satu ekosistem digital yang aman, cepat, dan mudah digunakan.

---

## 2. Goals & Objectives

### Primary Goals
1. **Digitalisasi Penuh** — Menghilangkan proses manual dan kertas dalam pengelolaan tugas akhir
2. **Transparansi** — Memberikan visibilitas real-time tentang status dan progress setiap mahasiswa
3. **Efisiensi** — Mengurangi waktu administratif dan mempercepat alur persetujuan
4. **Akuntabilitas** — Mencatat jejak audit lengkap untuk setiap tindakan dalam sistem
5. **Standarisasi** — Memastikan konsistensi prosedur dan penilaian di seluruh fakultas

### Success Metrics
- **Adoption Rate:** 100% mahasiswa angkatan 2027 menggunakan sistem
- **Time Reduction:** Rata-rata waktu approval pengajuan berkurang 50%
- **Data Completeness:** 100% thesis memiliki rekam jejak lengkap dari awal hingga arsip
- **User Satisfaction:** Target NPS (Net Promoter Score) ≥ 70
- **System Uptime:** 99.5% availability

---

## 3. Scope

### In Scope (v1.0)

**Core Workflows:**
- Pengajuan judul/topik skripsi oleh mahasiswa
- Review dan persetujuan judul oleh Kaprodi
- Penunjukan dosen pembimbing oleh Kaprodi
- Manajemen bimbingan dan log konsultasi
- Upload dan manajemen dokumen dengan approval workflow
- Persetujuan seminar proposal
- Penjadwalan seminar proposal
- Penilaian seminar proposal dengan sistem bobot tetap
- Revisi pasca seminar proposal
- Pengajuan sidang skripsi
- Penjadwalan sidang skripsi
- Penilaian dosen penguji dengan sistem bobot tetap
- Revisi akhir pasca sidang
- Penetapan status kelulusan/yudisium
- Arsip digital skripsi final

**Nice to Have (Deferred):**
- Export laporan ke PDF dan Excel
- Dashboard analitik dan statistik lanjutan
- Konfigurasi bobot penilaian dinamis

**Supporting Features (Core):**
- Manajemen pengguna (CRUD, import CSV/Excel, aktivasi, reset password)
- Role-Based Access Control (5 roles)
- Dashboard progress untuk Mahasiswa dan Dosen
- Email notification (7 trigger events utama)
- Audit log seluruh aktivitas kritis
- Pencarian, filter, dan pagination
- Manajemen tahun akademik

**Technical Infrastructure:**
- Frontend: Next.js 16 + React 19 + TypeScript
- Backend: Go 1.24+ (Gin Framework, Clean Architecture)
- Database: PostgreSQL
- Storage: Supabase Storage
- Authentication: JWT
- API Documentation: Swagger/OpenAPI
- API Versioning: /api/v1/...

---

## 4. Out of Scope (v1.0)

Fitur-fitur berikut **tidak termasuk** dalam v1.0 dan akan dipertimbangkan untuk versi berikutnya:

### v1.1 (Nice to Have)
- **Export PDF dan Excel** — laporan thesis, beban kerja, progress akademik
- **Dashboard Analitik dan Statistik Lanjutan** — grafik, chart distribusi beban, trend
- **Konfigurasi Bobot Penilaian Dinamis** — panel admin untuk mengubah komponen dan bobot

### v2.0 (Future)
- **Document Versioning** — rollback, compare version, restore version
- **Integrasi dengan SIAKAD** — sync data mahasiswa dan dosen otomatis
- **WhatsApp Notification** — notifikasi via WhatsApp selain email
- **Self-registration** — mahasiswa/dosen registrasi sendiri
- **Mobile native app** (iOS/Android) — v1.0 hanya web responsive
- **Video conference integration** — fitur bimbingan online terintegrasi
- **Plagiarism checker** — deteksi plagiarisme terintegrasi
- **E-signature** — tanda tangan digital untuk lembar pengesahan
- **API untuk external systems** — integrasi dengan sistem eksternal
- **Real-time chat/messaging** — komunikasi real-time antar pengguna
- **Advanced analytics & ML** — prediksi waktu kelulusan, rekomendasi pembimbing
- **Multi-language support** — v1.0 hanya Bahasa Indonesia
- **Public thesis repository** — akses publik ke arsip skripsi

---

## 5. Target Users & Roles

### User Roles & Responsibilities

| Role | Jumlah Estimasi | Responsibilities |
|------|----------------|------------------|
| **Admin Fakultas** | 2-3 | Manajemen pengguna (CRUD, import, reset password, self-service reset), manajemen tahun akademik, penjadwalan seminar/sidang, monitoring seluruh aktivitas, generate laporan |
| **Kaprodi** | 1-2 | Review dan approval pengajuan judul, penunjukan dosen pembimbing, approval pengajuan seminar/sidang, penetapan status yudisium, monitoring progress seluruh mahasiswa, generate laporan |
| **Mahasiswa** | ~80-90 | Pengajuan judul, upload dokumen, log konsultasi, pengajuan seminar/sidang, upload revisi, tracking progress pribadi |
| **Dosen Pembimbing** | ~15-20 | Review dan approval dokumen mahasiswa bimbingan, log konsultasi, memberikan feedback dan catatan revisi, monitoring progress mahasiswa bimbingan |
| **Dosen Penguji** | ~15-20 | Input penilaian pada seminar proposal dan sidang skripsi sesuai bobot tetap yang berlaku |

**Total Active Users:** ~100 akun aktif (bukan concurrent users). Dengan arsitektur Next.js + Go + PostgreSQL, kapasitas tersebut realistis dan masih memiliki ruang untuk berkembang.

---

## 6. Functional Requirements

### 6.1 Authentication & User Management

#### FR-AUTH-001: Login
- **Deskripsi:** Pengguna login menggunakan email dan password
- **Input:** Email, password
- **Output:** JWT access token dan refresh token
- **Business Rules:**
  - Password minimal 8 karakter
  - Account lock setelah 5 kali login gagal
  - Session timeout 24 jam (access token), 7 hari (refresh token)
  - Audit log setiap login attempt

#### FR-AUTH-002: Logout
- **Deskripsi:** Pengguna logout dan invalidate token
- **Business Rules:**
  - Blacklist token yang di-logout
  - Redirect ke halaman login

#### FR-AUTH-003: Refresh Token
- **Deskripsi:** Generate access token baru tanpa login ulang
- **Business Rules:**
  - Refresh token valid 7 hari
  - Setelah refresh, rotate refresh token

#### FR-AUTH-004: Role-Based Access Control (RBAC)
- **Deskripsi:** Setiap endpoint protected berdasarkan role
- **Business Rules:**
  - Setiap user hanya memiliki 1 role aktif
  - Role tidak dapat diubah oleh user sendiri
  - Akses endpoint sesuai matrix RBAC (detail di section 14)

#### FR-USER-001: CRUD User (Admin Only)
- **Deskripsi:** Admin kelola akun pengguna
- **Operasi:** Create, Read, Update, Delete (soft delete)
- **Atribut User:** email, nama lengkap, NIM/NIDN, role, program studi, status (aktif/nonaktif), profile_photo_url
- **Business Rules:**
  - Email harus unique
  - Password default: auto-generate dan dikirim via email
  - Audit log setiap perubahan

#### FR-USER-002: Import User via CSV/Excel
- **Deskripsi:** Admin import bulk user dari file
- **Format:** Template CSV/Excel standar dengan kolom: email, nama, NIM/NIDN, role, prodi
- **Business Rules:**
  - Validasi format per baris
  - Skip baris dengan error, lanjutkan yang valid
  - Generate error report untuk baris gagal
  - Send welcome email otomatis ke user yang berhasil diimport

#### FR-USER-003: Reset Password (Admin)
- **Deskripsi:** Admin reset password user
- **Business Rules:**
  - Generate password baru otomatis
  - Send email dengan password baru
  - User diminta ganti password saat first login

#### FR-USER-005: Self-Service Reset Password
- **Deskripsi:** Pengguna meminta reset password sendiri melalui email
- **Business Rules:**
  - Pengguna memasukkan email untuk meminta tautan reset
  - Sistem menghasilkan token sekali pakai dengan masa berlaku terbatas
  - Tautan reset dikirim via email
  - Pengguna membuat password baru melalui halaman reset
  - Token tidak bisa digunakan setelah expired atau sudah digunakan

#### FR-USER-004: Activate/Deactivate Account
- **Deskripsi:** Admin aktifkan atau nonaktifkan akun
- **Business Rules:**
  - User nonaktif tidak dapat login
  - Session aktif ter-invalidate saat account deactivated

---

### 6.2 Manajemen Pengajuan Judul

#### FR-THESIS-001: Pengajuan Judul (Mahasiswa)
- **Deskripsi:** Mahasiswa mengajukan judul skripsi
- **Input:** Judul, abstrak, bidang keahlian, jenis TA (skripsi/tugas akhir), dokumen proposal awal (opsional)
- **Business Rules:**
  - Mahasiswa hanya bisa memiliki 1 thesis aktif (status bukan `cancelled` atau `graduated`)
  - Judul minimal 10 kata, maksimal 200 karakter
  - Abstrak minimal 100 kata
  - Status awal: `submitted`
  - Email notification ke Kaprodi

#### FR-THESIS-002: Review Pengajuan Judul (Kaprodi)
- **Deskripsi:** Kaprodi review dan approve/reject judul
- **Input:** Keputusan (approve/reject), catatan/feedback
- **Business Rules:**
  - Jika approve: status → `approved`, lanjut ke penunjukan pembimbing
  - Jika reject: status → `rejected`, mahasiswa dapat revisi dan submit ulang
  - Email notification ke mahasiswa

#### FR-THESIS-003: Penunjukan Dosen Pembimbing (Kaprodi)
- **Deskripsi:** Kaprodi menetapkan dosen pembimbing setelah judul disetujui
- **Input:** Pilih dosen pembimbing (1 atau 2 dosen)
- **Business Rules:**
  - Dosen yang dipilih harus memiliki role `dosen_pembimbing`
  - Sistem tampilkan daftar dosen dengan beban bimbingan saat ini (load balancing hint)
  - Status thesis → `in_progress`
  - Email notification ke mahasiswa dan dosen pembimbing

#### FR-THESIS-004: View Thesis List & Detail
- **Deskripsi:** Semua role dapat melihat daftar thesis sesuai akses
- **Akses:**
  - Admin/Kaprodi: semua thesis
  - Dosen Pembimbing: thesis mahasiswa bimbingannya
  - Dosen Penguji: thesis yang dia uji
  - Mahasiswa: thesis miliknya sendiri
- **Filter:** Status, tahun akademik, program studi, dosen pembimbing, bidang keahlian
- **Pagination:** 20 item per halaman

---

### 6.3 Manajemen Bimbingan

#### FR-CONSULT-001: Create Log Konsultasi
- **Deskripsi:** Mahasiswa atau Dosen Pembimbing mencatat log bimbingan
- **Input:** Tanggal konsultasi, topik yang dibahas, catatan hasil konsultasi, tindak lanjut, attachment (opsional)
- **Business Rules:**
  - Tanggal tidak boleh di masa depan
  - Mahasiswa hanya bisa buat log untuk thesis-nya sendiri
  - Dosen hanya bisa buat log untuk mahasiswa bimbingannya
  - Status awal: `pending`
  - Email notification ke dosen (jika dibuat mahasiswa) atau mahasiswa (jika dibuat dosen)

#### FR-CONSULT-002: Approve Log Konsultasi (Dosen)
- **Deskripsi:** Dosen pembimbing approve log konsultasi
- **Business Rules:**
  - Status → `approved`
  - Counter jumlah bimbingan bertambah
  - Email notification ke mahasiswa

#### FR-CONSULT-003: View History Konsultasi
- **Deskripsi:** Mahasiswa dan dosen melihat riwayat konsultasi lengkap
- **Display:** Timeline chronological, filter by date range
- **Metrics:** Total konsultasi, rata-rata interval konsultasi, konsultasi terakhir

---

### 6.4 Manajemen Dokumen

#### FR-DOC-001: Upload Dokumen (Mahasiswa)
- **Deskripsi:** Mahasiswa upload dokumen skripsi bertahap
- **Tipe Dokumen:**
  - Proposal Skripsi
  - Draft Bab 1, 2, 3, 4, 5 (per bab)
  - Dokumen Seminar Proposal
  - Dokumen Sidang Skripsi
  - Skripsi Final
  - Lembar Revisi
  - Lembar Pengesahan
- **Input:** File (PDF), tipe dokumen, versi, catatan
- **Business Rules:**
  - Format file: PDF wajib untuk dokumen utama
  - Ukuran maksimal: 10 MB (configurable)
  - Status awal: `pending_review`
  - Versioning: setiap upload baru menjadi versi baru, versi lama tetap tersimpan
  - Email notification ke dosen pembimbing

#### FR-DOC-002: Review Dokumen (Dosen Pembimbing)
- **Deskripsi:** Dosen pembimbing review dan approve/request revision
- **Input:** Keputusan (approve/revision), catatan feedback
- **Business Rules:**
  - Jika approve: status → `approved`
  - Jika revision: status → `revision_required`, mahasiswa upload versi baru
  - Email notification ke mahasiswa
  - Gate Logic: mahasiswa tidak bisa lanjut ke tahap berikutnya jika dokumen prerequisite belum `approved`

#### FR-DOC-003: Download Dokumen
- **Deskripsi:** User download dokumen
- **Business Rules:**
  - Mahasiswa: hanya dokumennya sendiri
  - Dosen Pembimbing: dokumen mahasiswa bimbingannya
  - Dosen Penguji: dokumen mahasiswa yang dia uji
  - Admin/Kaprodi: semua dokumen
  - Generate presigned URL dari storage (expired 15 menit)

#### FR-DOC-004: Document Revision (v1.0)
- **Deskripsi:** Sistem menyimpan file revisi lama sebagai arsip dan mencatat metadata riwayat upload
- **Display:** Riwayat upload dengan metadata sederhana: tanggal upload, nama file, pengunggah, status
- **Business Rules:**
  - File lama tetap tersimpan sebagai arsip internal
  - File terbaru menjadi dokumen aktif
  - Tidak ada fitur rollback, compare version, atau restore version (akan ada di v2.0)
  - Download versi spesifik tersedia

---

### 6.5 Manajemen Seminar Proposal

#### FR-SEMINAR-001: Pengajuan Seminar Proposal (Mahasiswa)
- **Deskripsi:** Mahasiswa mengajukan seminar proposal
- **Business Rules:**
  - Gate: hanya bisa ajukan jika dokumen seminar proposal sudah `approved`
  - Status awal: `pending`
  - Email notification ke Kaprodi dan Admin

#### FR-SEMINAR-002: Penjadwalan Seminar (Admin/Kaprodi)
- **Deskripsi:** Admin atau Kaprodi menjadwalkan seminar
- **Input:** Tanggal, waktu, ruangan, dosen penguji (minimal 2)
- **Business Rules:**
  - Tidak boleh bentrok jadwal (ruangan, dosen, mahasiswa)
  - Tanggal minimal 3 hari dari sekarang
  - Status → `scheduled`
  - Email notification ke mahasiswa, pembimbing, dan penguji

#### FR-SEMINAR-003: Update Jadwal Seminar
- **Deskripsi:** Admin/Kaprodi ubah jadwal seminar
- **Business Rules:**
  - Email notification reschedule ke semua pihak terkait

#### FR-SEMINAR-004: Input Nilai Seminar (Dosen Penguji)
- **Deskripsi:** Dosen penguji input nilai per komponen
- **Input:** Nilai per komponen penilaian (bobot tetap sesuai ketentuan fakultas)
- **Business Rules:**
  - Setiap komponen: skala 0-100
  - Penguji hanya bisa input nilai untuk seminar yang dia ditugaskan
  - Penguji tidak bisa input nilai 2 kali untuk seminar yang sama
  - Sistem auto-calculate nilai akhir = rata-rata berbobot dari semua penguji
  - Formula: `Nilai_Akhir = Σ (Nilai_Komponen_i × Bobot_i) untuk semua penguji, lalu rata-rata`
  - Bobot tetap (v1.0): Presentasi 30%, Penguasaan Materi 30%, Dokumen 25%, Tanya Jawab 15%

#### FR-SEMINAR-005: Hasil Seminar
- **Deskripsi:** Sistem generate hasil seminar otomatis
- **Business Rules:**
  - Jika semua penguji sudah input nilai: status → `completed`
  - Jika nilai akhir < 60: status → `failed`, mahasiswa harus mengulang seminar
  - Jika nilai akhir ≥ 60: status → `passed`, mahasiswa bisa lanjut ke tahap sidang
  - Email notification hasil ke mahasiswa dan pembimbing

#### FR-SEMINAR-006: Revisi Pasca Seminar
- **Deskripsi:** Catat kebutuhan revisi pasca seminar
- **Input:** Catatan revisi dari penguji
- **Business Rules:**
  - Mahasiswa upload dokumen revisi, dosen pembimbing approve

---

### 6.6 Manajemen Sidang Skripsi

#### FR-DEFENSE-001: Pengajuan Sidang (Mahasiswa)
- **Deskripsi:** Mahasiswa mengajukan sidang skripsi
- **Business Rules:**
  - Gate: hanya bisa ajukan jika seminar proposal `passed` dan dokumen sidang `approved`
  - Status awal: `pending`
  - Email notification ke Kaprodi dan Admin

#### FR-DEFENSE-002: Penjadwalan Sidang (Admin/Kaprodi)
- **Deskripsi:** Admin atau Kaprodi menjadwalkan sidang
- **Input:** Tanggal, waktu, ruangan, dosen penguji (minimal 2, bisa berbeda dari seminar)
- **Business Rules:**
  - Tidak boleh bentrok jadwal
  - Tanggal minimal 7 hari dari sekarang
  - Status → `scheduled`
  - Email notification ke mahasiswa, pembimbing, dan penguji

#### FR-DEFENSE-003: Input Nilai Sidang (Dosen Penguji)
- **Deskripsi:** Dosen penguji input nilai per komponen
- **Business Rules:** Sama seperti FR-SEMINAR-004, dengan bobot tetap sesuai ketentuan fakultas yang berlaku

#### FR-DEFENSE-004: Hasil Sidang
- **Deskripsi:** Sistem generate hasil sidang otomatis
- **Business Rules:**
  - Jika semua penguji sudah input nilai: status → `completed`
  - Jika nilai akhir < 60: status → `failed`, perlu sidang ulang
  - Jika nilai akhir ≥ 60 dan < 75: lulus dengan revisi
  - Jika nilai akhir ≥ 75: lulus tanpa revisi atau dengan revisi minor
  - Email notification hasil ke mahasiswa dan pembimbing

#### FR-DEFENSE-005: Revisi Akhir
- **Deskripsi:** Catat dan kelola revisi akhir pasca sidang
- **Input:** Catatan revisi dari penguji
- **Business Rules:**
  - Mahasiswa upload dokumen revisi
  - Dosen pembimbing approve revisi
  - Setelah approve: mahasiswa siap upload skripsi final untuk arsip

---

### 6.7 Yudisium & Arsip Digital

#### FR-YUDISIUM-001: Penetapan Yudisium (Kaprodi)
- **Deskripsi:** Kaprodi tetapkan status lulus setelah semua revisi selesai
- **Business Rules:**
  - Gate: sidang `completed` dan revisi akhir `approved`
  - Status thesis → `graduated`
  - Email notification "Selamat, Anda telah lulus" ke mahasiswa
  - Audit log penetapan yudisium

#### FR-ARCHIVE-001: Upload Skripsi Final ke Arsip
- **Deskripsi:** Mahasiswa atau Admin upload skripsi final ke arsip digital
- **Input:** File PDF skripsi final, metadata (abstrak Bahasa Indonesia, abstrak Bahasa Inggris, kata kunci, tahun lulus)
- **Business Rules:**
  - Gate: hanya thesis dengan status `graduated`
  - File tersimpan di storage terpisah (public-read untuk civitas akademika)
  - Metadata ter-index untuk pencarian
  - Email notification "Arsip tersedia" ke mahasiswa

#### FR-ARCHIVE-002: Search & Browse Arsip
- **Deskripsi:** Pencarian arsip skripsi
- **Search by:** Judul, penulis, NIM, kata kunci, abstrak, tahun, bidang keahlian, dosen pembimbing
- **Business Rules:**
  - Full-text search pada judul dan abstrak (PostgreSQL `tsvector`)
  - Pagination: 20 item per halaman
  - Filter: tahun, bidang, prodi

#### FR-ARCHIVE-003: Download Skripsi dari Arsip
- **Deskripsi:** User download skripsi final dari arsip
- **Business Rules:**
  - Mahasiswa: hanya skripsinya sendiri
  - Dosen/Kaprodi/Admin: semua skripsi
  - Generate presigned URL (expired 30 menit)
  - Log setiap download

---

### 6.8 Konfigurasi Penilaian Dinamis (Future — v2.0)

#### FR-GRADING-001: Manage Grading Components (Admin/Kaprodi)
- **Deskripsi:** Admin atau Kaprodi kelola komponen penilaian dan bobotnya (akan hadir di v2.0)
- **Atribut:** Nama komponen, bobot (%), tipe (seminar/sidang), tahun akademik aktif, status (aktif/nonaktif)
- **Business Rules:**
  - Total bobot semua komponen aktif untuk 1 tipe harus = 100%
  - Validasi realtime saat perubahan bobot
  - Perubahan konfigurasi tidak retroaktif (hanya berlaku untuk penilaian baru)
  - Default components:
    - Presentasi: 30%
    - Penguasaan Materi: 30%
    - Kualitas Naskah: 25%
    - Kemampuan Menjawab: 15%
  - Audit log setiap perubahan konfigurasi

#### FR-GRADING-002: View Active Grading Config
- **Deskripsi:** Snapshot konfigurasi yang digunakan pada setiap penilaian (akan hadir di v2.0)
- **Business Rules:**
  - Penilaian yang sudah dibuat tidak berubah meski konfigurasi diubah
  - Display konfigurasi yang digunakan pada setiap penilaian (snapshot)

---

### 6.9 Dashboard & Reporting

#### FR-DASHBOARD-001: Dashboard Admin/Kaprodi
- **Widgets:**
  - **Ringkasan Akademik:**
    - Total mahasiswa tugas akhir aktif
    - Breakdown per tahapan (pengajuan, bimbingan, seminar, sidang, lulus)
    - Jumlah lulusan (per tahun akademik, semester)
    - Rata-rata durasi penyelesaian skripsi (dalam bulan)
    - Grafik perkembangan: line chart jumlah mahasiswa per tahap per bulan
  - **Analitik Dosen:**
    - Jumlah mahasiswa bimbingan per dosen (tabel + bar chart)
    - Jumlah seminar dan sidang per dosen
    - Distribusi beban dosen (tertinggi/terendah)
  - **Operasional:**
    - Jadwal seminar/sidang minggu ini dan minggu depan
    - Pengajuan yang menunggu approval (count + link ke detail)
    - Dokumen yang menunggu review (count + link)
    - Statistik aktivitas sistem (jumlah login hari ini, jumlah dokumen diupload minggu ini)
- **Filter:** Tahun akademik, semester, program studi
- **Business Rules:**
  - Data real-time (refresh otomatis setiap 5 menit)
  - Semua angka clickable untuk drill-down ke detail

#### FR-DASHBOARD-002: Dashboard Mahasiswa
- **Widgets:**
  - Progress tracker visual: tahap saat ini + status setiap tahap
  - Thesis info: judul, pembimbing, status
  - Dokumen status: daftar dokumen dengan status review
  - Jadwal: seminar/sidang yang dijadwalkan
  - Konsultasi terakhir: tanggal + catatan
  - Notifikasi: pending actions (upload dokumen, revisi, dll)

#### FR-DASHBOARD-003: Dashboard Dosen Pembimbing
- **Widgets:**
  - Daftar mahasiswa bimbingan dengan status masing-masing
  - Dokumen yang menunggu review
  - Jadwal seminar/sidang mahasiswa bimbingan
  - Statistik: total mahasiswa, rata-rata progress, mahasiswa dengan progress tertinggal

#### FR-DASHBOARD-004: Dashboard Dosen Penguji
- **Widgets:**
  - Jadwal seminar/sidang yang ditugaskan
  - Penilaian yang belum diinput
  - History penilaian

---

### 6.10 Notification System

#### FR-NOTIF-001: Email Notification
- **7 Event Triggers (v1.0):**
  1. Pengajuan judul berhasil dikirim (ke: Mahasiswa, Kaprodi)
  2. Judul disetujui oleh Kaprodi (ke: Mahasiswa)
  3. Judul ditolak oleh Kaprodi (ke: Mahasiswa)
  4. Dosen pembimbing telah ditetapkan (ke: Mahasiswa, Dosen Pembimbing)
  5. Jadwal seminar proposal diterbitkan (ke: Mahasiswa, Pembimbing, Penguji)
  6. Jadwal sidang diterbitkan/diubah (ke: Mahasiswa, Pembimbing, Penguji)
  7. Hasil penilaian sidang dan kelulusan (ke: Mahasiswa, Pembimbing)

- **Template Design:**
  - Header: Logo FILKOM Unida + Nama aplikasi
  - Body: Informasi lengkap sesuai event + link ke halaman terkait
  - Footer: Kontak Admin Fakultas + disclaimer
  - Desain profesional, responsive, konsisten

- **Business Rules:**
  - Email dikirim dari backend Go menggunakan provider Resend
  - API Key disimpan pada environment variable backend (backend adalah satu-satunya pihak yang mengakses API Key)
  - Email dikirim async (tidak block HTTP response)
  - Retry mechanism: belum diterapkan pada v1.0 (akan ada di v2.0)
  - Log semua email di tabel `email_logs` (status: sent/failed, provider, error_message)
  - Queue dan background worker: belum diterapkan pada v1.0

#### FR-NOTIF-002: In-App Notification (Optional for v1.0)
- Notifikasi dalam aplikasi dengan badge count
- Mark as read functionality
- Link ke halaman terkait

---

### 6.11 Export & Reporting (Nice to Have — v1.1)

#### FR-EXPORT-001: Export Daftar Thesis (Admin/Kaprodi)
- **Format:** PDF, Excel
- **Content:** Tabel daftar thesis dengan kolom: NIM, Nama, Judul, Pembimbing, Status, Tanggal Pengajuan, Tanggal Lulus
- **Filter:** Tahun akademik, semester, prodi, status

#### FR-EXPORT-002: Export Beban Kerja Dosen
- **Format:** PDF, Excel
- **Content:** Tabel dosen dengan kolom: Nama, NIDN, Jumlah Bimbingan, Jumlah Seminar, Jumlah Sidang

#### FR-EXPORT-003: Export Progress Akademik
- **Format:** PDF, Excel
- **Content:** Breakdown mahasiswa per tahap, statistik waktu penyelesaian, trend per semester

#### FR-EXPORT-004: Export Riwayat Thesis (Mahasiswa/Admin)
- **Format:** PDF
- **Content:** Timeline lengkap 1 mahasiswa dari pengajuan hingga lulus

### 6.12 Grading Configuration (Future — v2.0)

#### FR-GRADING-001: Manage Grading Components (Admin/Kaprodi)
- **Deskripsi:** Kelola komponen penilaian dan bobotnya (akan hadir di v2.0)

#### FR-GRADING-002: View Active Grading Config
- **Deskripsi:** Snapshot konfigurasi yang digunakan pada setiap penilaian (akan hadir di v2.0)

---

### 6.13 Audit Log

#### FR-AUDIT-001: Comprehensive Audit Trail
- **Logged Actions:**
  - User login/logout
  - User CRUD operations
  - Thesis status changes
  - Document upload/review
  - Approval/rejection actions
  - Grading configuration changes
  - Scheduling changes
  - Grade submission
- **Atribut:** Timestamp, user_id, action, entity_type, entity_id, old_value, new_value, IP address, user agent
- **Business Rules:**
  - Audit log tidak bisa diedit/dihapus (append-only)
  - Retention: permanent (atau se