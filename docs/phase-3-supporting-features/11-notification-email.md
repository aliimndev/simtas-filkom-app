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
- [ ] Install Resend Go SDK:
  ```bash
  go get github.com/resend/resend-go/v2
  ```

### Email Service — Implementasi Resend

Ganti `backend/pkg/email/stub_email_service.go` dengan implementasi nyata:

- [ ] Buat `backend/pkg/email/resend_email_service.go`:
  ```go
  type ResendEmailService struct {
    client    *resend.Client
    fromEmail string
    fromName  string
    db        *gorm.DB  // untuk logging
  }

  func NewResendEmailService(apiKey, fromEmail, fromName string, db *gorm.DB) *ResendEmailService
  ```
- [ ] Implementasikan semua method dari `EmailService` interface yang sudah didefinisikan di Job 05
- [ ] Setiap method: render template → kirim via Resend → log ke `email_logs`
- [ ] Pengiriman async menggunakan goroutine:
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
- **Trigger:** Mahasiswa submit pengajuan judul
- **Penerima:** Kaprodi (semua akun role kaprodi)
- **Subjek:** `[SIMTAS] Pengajuan Judul Baru — {nama_mahasiswa}`
- **Konten:**
  - Nama dan NIM mahasiswa
  - Judul yang diajukan
  - Abstrak singkat
  - Tanggal pengajuan
  - Tombol/link "Review Pengajuan" → ke halaman review Kaprodi

#### Template 2: `thesis_reviewed.html`
- **Trigger:** Kaprodi approve/reject judul
- **Penerima:** Mahasiswa
- **Subjek (approved):** `[SIMTAS] Judul Skripsi Anda Disetujui`
- **Subjek (rejected):** `[SIMTAS] Judul Skripsi Anda Perlu Revisi`
- **Konten:**
  - Status keputusan (dengan warna: hijau untuk approved, merah untuk rejected)
  - Judul yang diajukan
  - Catatan/feedback dari Kaprodi
  - Tombol "Lihat Detail" → ke halaman thesis mahasiswa

#### Template 3: `supervisor_assigned.html`
- **Trigger:** Kaprodi tetapkan dosen pembimbing
- **Penerima:** Mahasiswa + setiap Dosen Pembimbing yang ditunjuk
- **Subjek:** `[SIMTAS] Dosen Pembimbing Telah Ditetapkan`
- **Konten (ke mahasiswa):**
  - Nama dosen pembimbing yang ditunjuk (jika 2 dosen, list keduanya)
  - Informasi kontak dosen (email, NIDN)
  - Petunjuk langkah selanjutnya (mulai bimbingan)
- **Konten (ke dosen):**
  - Nama dan NIM mahasiswa yang dibimbing
  - Judul skripsi mahasiswa
  - Informasi kontak mahasiswa

#### Template 4: `seminar_scheduled.html`
- **Trigger:** Admin/Kaprodi jadwalkan seminar proposal
- **Penerima:** Mahasiswa + Dosen Pembimbing + setiap Dosen Penguji
- **Subjek:** `[SIMTAS] Jadwal Seminar Proposal — {nama_mahasiswa}`
- **Konten:**
  - Tanggal, waktu, ruangan seminar
  - Judul skripsi
  - Daftar penguji (untuk mahasiswa & pembimbing)
  - Daftar mahasiswa dan pembimbing (untuk penguji)
  - Catatan persiapan
- **Reschedule:** subjek berubah menjadi `[SIMTAS] Perubahan Jadwal Seminar Proposal`

#### Template 5: `seminar_result.html`
- **Trigger:** Semua penguji submit nilai seminar (auto-finalize)
- **Penerima:** Mahasiswa + Dosen Pembimbing
- **Subjek (passed):** `[SIMTAS] Selamat! Anda Lulus Seminar Proposal`
- **Subjek (failed):** `[SIMTAS] Hasil Seminar Proposal`
- **Konten:**
  - Nilai akhir seminar
  - Status (Lulus / Tidak Lulus)
  - Breakdown nilai per komponen (tanpa nama penguji — anonim)
  - Langkah selanjutnya (jika lulus: persiapkan dokumen sidang)

#### Template 6: `defense_scheduled.html`
- **Trigger:** Admin/Kaprodi jadwalkan sidang
- **Penerima:** Mahasiswa + Dosen Pembimbing + setiap Dosen Penguji
- **Subjek:** `[SIMTAS] Jadwal Sidang Skripsi — {nama_mahasiswa}`
- **Konten:** Sama seperti seminar_scheduled, dengan tambahan informasi bahwa ini adalah sidang skripsi final
- **Reschedule:** subjek berubah menjadi `[SIMTAS] Perubahan Jadwal Sidang Skripsi`

#### Template 7: `graduation.html`
- **Trigger:** Kaprodi tetapkan status graduated (yudisium)
- **Penerima:** Mahasiswa + Dosen Pembimbing
- **Subjek:** `[SIMTAS] Selamat! Skripsi Anda Dinyatakan Lulus`
- **Konten:**
  - Ucapan selamat
  - Ringkasan: judul skripsi, nilai sidang, tanggal yudisium
  - Informasi langkah selanjutnya: upload skripsi final untuk arsip
  - Logo FILKOM Unida

### Template HTML Base Layout
- [ ] Buat `backend/pkg/email/templates/base.html` — layout yang digunakan semua template:
  ```html
  <!-- Header: Logo FILKOM Unida + nama sistem -->
  <!-- Body: konten dinamis -->
  <!-- Footer: Kontak Admin Fakultas | Fakultas Ilmu Komputer Universitas Djuanda | disclaimer -->
  ```
- [ ] Gunakan inline CSS (kompatibel dengan semua email client)
- [ ] Responsif untuk mobile (max-width 600px)
- [ ] Warna brand: sesuaikan dengan identitas FILKOM Unida
- [ ] Semua link menggunakan URL dari env `FRONTEND_URL`

### Template Engine
- [ ] Gunakan Go `html/template` package:
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
- [ ] Setelah setiap pengiriman (berhasil atau gagal), insert ke `email_logs`:
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
- [ ] Log error ke application logger jika pengiriman gagal (jangan panic)

### Swap Stub → Real Implementation
- [ ] Update `cmd/server/main.go`: inisiasi `ResendEmailService` dan inject ke semua use case yang membutuhkan
- [ ] Pastikan `EmailService` interface dipakai (bukan struct langsung) — mudah swap ke provider lain
- [ ] Tambah env variable: `FRONTEND_URL=https://simtas.filkom.unida.ac.id`

### Testing Email (Development)
- [ ] Tambahkan env `EMAIL_DEV_MODE=true` — jika true, log email ke console TANPA mengirim ke Resend
- [ ] Endpoint internal (hanya di `development` env): `POST /api/v1/internal/test-email` — kirim email test ke address tertentu

---

## Done Criteria

- [ ] `ResendEmailService` berhasil mengirim email nyata ke inbox (test dengan akun Resend sandbox)
- [ ] Semua 7 template ter-render tanpa error
- [ ] Template tampil dengan benar di Gmail dan Outlook (test manual)
- [ ] Email dikirim async — HTTP response tidak menunggu email terkirim
- [ ] Setiap pengiriman tercatat di tabel `email_logs`
- [ ] Email gagal → status `failed` di `email_logs`, tidak crash aplikasi
- [ ] `EMAIL_DEV_MODE=true` → email tidak terkirim, hanya log ke console
- [ ] Semua stub di job 05–10 berhasil diganti dengan implementasi nyata
- [ ] Trigger event 1–7 berhasil mengirim email ke penerima yang tepat dalam end-to-end test
