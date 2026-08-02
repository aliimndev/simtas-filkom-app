# Job 11 — Email Notification System

**Phase:** 3 — Supporting Features
**Referensi PRD:** Section 6.10 (FR-NOTIF-001)
**Prerequisites:** Job 10 (Archive Module) ✅
**Estimasi:** 2 hari

---

## Objective

Implementasi sistem notifikasi email penuh menggunakan Resend sebagai provider, menggantikan semua stub email yang telah dipasang di job-job sebelumnya. Mencakup 7 event trigger utama v1.0 dengan template HTML profesional, pengiriman async, dan logging setiap email. Setelah job ini selesai, seluruh notifikasi email sistem berjalan nyata.

---

## Checklist

### Dependencies
- [x] Install Resend Go SDK:
  ```bash
  go get github.com/resend/resend-go/v2
  ```

### Email Service — Implementasi Resend

Ganti `backend/pkg/email/stub_email_service.go` dengan implementasi nyata (stub dihapus, digantikan `ResendEmailService`):

- [x] Buat `backend/pkg/email/resend_email_service.go`:
  ```go
  type ResendEmailService struct {
    client    *resend.Client
    fromEmail string
    fromName  string
    db        *gorm.DB  // untuk logging
  }

  func NewResendEmailService(apiKey, fromEmail, fromName string, db *gorm.DB) *ResendEmailService
  ```
- [x] Implementasikan semua method dari `EmailService` interface yang sudah didefinisikan di Job 05
- [x] Setiap method: render template → kirim via Resend → log ke `email_logs`
- [x] Pengiriman async menggunakan goroutine:
  ```go
  func (s *ResendEmailService) sendAsync(to []string, subject, html string, eventType string) {
    go func() {
      // kirim email
      // log ke email_logs (status sent/failed)
    }()
  }
  ```

### 7 Event Templates (v1.0)

Buat folder `backend/pkg/email/templates/` dengan file HTML per event:

#### Template 1: `thesis_submitted.html`
- [x] **Trigger:** Mahasiswa submit pengajuan judul
- [x] **Penerima:** Kaprodi (semua akun role kaprodi)
- [x] **Subjek:** `[SIMTAS] Pengajuan Judul Baru — {nama_mahasiswa}`
- [x] **Konten:**
  - Nama dan NIM mahasiswa
  - Judul yang diajukan
  - Abstrak singkat
  - Tanggal pengajuan
  - Tombol/link "Review Pengajuan" → ke halaman review Kaprodi

#### Template 2: `thesis_reviewed.html`
- [x] **Trigger:** Kaprodi approve/reject judul
- [x] **Penerima:** Mahasiswa
- [x] **Subjek (approved):** `[SIMTAS] Judul Skripsi Anda Disetujui`
- [x] **Subjek (rejected):** `[SIMTAS] Judul Skripsi Anda Perlu Revisi`
- [x] **Konten:**
  - Status keputusan (dengan warna: hijau untuk approved, merah untuk rejected)
  - Judul yang diajukan
  - Catatan/feedback dari Kaprodi
  - Tombol "Lihat Detail" → ke halaman thesis mahasiswa

#### Template 3: `supervisor_assigned.html`
- [x] **Trigger:** Kaprodi tetapkan dosen pembimbing
- [x] **Penerima:** Mahasiswa + setiap Dosen Pembimbing yang ditunjuk
- [x] **Subjek:** `[SIMTAS] Dosen Pembimbing Telah Ditetapkan`
- [x] **Konten (ke mahasiswa):**
  - Nama dosen pembimbing yang ditunjuk (jika 2 dosen, list keduanya)
  - Informasi kontak dosen (email, NIDN)
  - Petunjuk langkah selanjutnya (mulai bimbingan)
- [x] **Konten (ke dosen):**
  - Nama dan NIM mahasiswa yang dibimbing
  - Judul skripsi mahasiswa
  - Informasi kontak mahasiswa

#### Template 4: `seminar_scheduled.html`
- [x] **Trigger:** Admin/Kaprodi jadwalkan seminar proposal
- [x] **Penerima:** Mahasiswa + Dosen Pembimbing + setiap Dosen Penguji
- [x] **Subjek:** `[SIMTAS] Jadwal Seminar Proposal — {nama_mahasiswa}`
- [x] **Konten:**
  - Tanggal, waktu, ruangan seminar
  - Judul skripsi
  - Daftar penguji (untuk mahasiswa & pembimbing)
  - Daftar mahasiswa dan pembimbing (untuk penguji)
  - Catatan persiapan
- [x] **Reschedule:** subjek berubah menjadi `[SIMTAS] Perubahan Jadwal Seminar Proposal`

#### Template 5: `seminar_result.html`
- [x] **Trigger:** Semua penguji submit nilai seminar (auto-finalize)
- [x] **Penerima:** Mahasiswa + Dosen Pembimbing
- [x] **Subjek (passed):** `[SIMTAS] Selamat! Anda Lulus Seminar Proposal`
- [x] **Subjek (failed):** `[SIMTAS] Hasil Seminar Proposal`
- [x] **Konten:**
  - Nilai akhir seminar
  - Status (Lulus / Tidak Lulus)
  - Breakdown nilai per komponen (tanpa nama penguji — anonim)
  - Langkah selanjutnya (jika lulus: persiapkan dokumen sidang)

#### Template 6: `defense_scheduled.html`
- [x] **Trigger:** Admin/Kaprodi jadwalkan sidang
- [x] **Penerima:** Mahasiswa + Dosen Pembimbing + setiap Dosen Penguji
- [x] **Subjek:** `[SIMTAS] Jadwal Sidang Skripsi — {nama_mahasiswa}`
- [x] **Konten:** Sama seperti seminar_scheduled, dengan tambahan informasi bahwa ini adalah sidang skripsi final
- [x] **Reschedule:** subjek berubah menjadi `[SIMTAS] Perubahan Jadwal Sidang Skripsi`

#### Template 7: `graduation.html`
- [x] **Trigger:** Kaprodi tetapkan status graduated (yudisium)
- [x] **Penerima:** Mahasiswa + Dosen Pembimbing
- [x] **Subjek:** `[SIMTAS] Selamat! Skripsi Anda Dinyatakan Lulus`
- [x] **Konten:**
  - Ucapan selamat
  - Ringkasan: judul skripsi, nilai sidang, tanggal yudisium
  - Informasi langkah selanjutnya: upload skripsi final untuk arsip
  - Logo FILKOM Unida

### Template HTML Base Layout
- [x] Buat `backend/pkg/email/templates/base.html` — layout yang digunakan semua template:
  ```html
  <!-- Header: Logo FILKOM Unida + nama sistem -->
  <!-- Body: konten dinamis -->
  <!-- Footer: Kontak Admin Fakultas | Fakultas Ilmu Komputer Universitas Djuanda | disclaimer -->
  ```
- [x] Gunakan inline CSS (kompatibel dengan semua email client)
- [x] Responsif untuk mobile (max-width 600px)
- [x] Warna brand: sesuaikan dengan identitas FILKOM Unida
- [x] Semua link menggunakan URL dari env `FRONTEND_URL`

### Template Engine
- [x] Gunakan Go `html/template` package:
  ```go
  func renderTemplate(templateName string, data interface{}) (string, error) {
    tmpl, err := template.ParseFiles(
      "pkg/email/templates/base.html",
      "pkg/email/templates/"+templateName,
    )
    // ...
  }
  ```

### Email Logging
- [x] Setelah setiap pengiriman (berhasil atau gagal), insert ke `email_logs`:
  ```go
  emailLog := &entity.EmailLog{
    RecipientEmail: to,
    EventType:      eventType,
    Subject:        subject,
    Status:         "sent",  // atau "failed"
    Provider:       "resend",
    ErrorMessage:   "",  // isi jika failed
  }
  ```
- [x] Log error ke application logger jika pengiriman gagal (jangan panic)

### Swap Stub → Real Implementation
- [x] Update `cmd/server/main.go`: inisiasi `ResendEmailService` dan inject ke semua use case yang membutuhkan
- [x] Pastikan `EmailService` interface dipakai (bukan struct langsung) — mudah swap ke provider lain
- [x] Tambah env variable: `FRONTEND_URL=https://simtas.filkom.unida.ac.id`

### Testing Email (Development)
- [x] Tambahkan env `EMAIL_DEV_MODE=true` — jika true, log email ke console TANPA mengirim ke Resend
- [x] Endpoint internal (hanya di `development` env): `POST /api/v1/internal/test-email` — kirim email test ke address tertentu

---

## Done Criteria

- [ ] `ResendEmailService` berhasil mengirim email nyata ke inbox (test dengan akun Resend sandbox)
- [x] Semua template ter-render tanpa error
- [ ] Template tampil dengan benar di Gmail dan Outlook (test manual)
- [x] Email dikirim async — HTTP response tidak menunggu email terkirim
- [x] Setiap pengiriman tercatat di tabel `email_logs`
- [x] Email gagal → status `failed` di `email_logs`, tidak crash aplikasi
- [x] `EMAIL_DEV_MODE=true` → email tidak terkirim, hanya log ke console
- [x] Semua stub di job 05–10 berhasil diganti dengan implementasi nyata
- [ ] Trigger event 1–7 berhasil mengirim email ke penerima yang tepat dalam end-to-end test
