# Job 17 — Frontend Halaman Mahasiswa

**Phase:** 4 — Frontend
**Referensi PRD:** Section 6.2–6.7 (alur thesis mahasiswa end-to-end)
**Prerequisites:** Job 16 (Frontend Admin Pages) ✅
**Estimasi:** 4 hari

---

## Objective

Implementasi semua halaman yang digunakan mahasiswa: pengajuan judul, halaman thesis saya, bimbingan & log konsultasi, manajemen dokumen, seminar proposal, sidang skripsi, dan arsip. Ini adalah halaman dengan traffic tertinggi di sistem.

---

## Checklist

### Halaman Thesis Saya (`/student/thesis`)

**File:** `frontend/src/app/(dashboard)/student/thesis/page.tsx`

- [x] Jika belum punya thesis aktif → tampilkan empty state dengan tombol "Ajukan Judul Skripsi"
- [x] Jika sudah punya thesis → tampilkan detail thesis:
  - Judul, abstrak, bidang, jenis TA
  - Status dengan `StatusBadge`
  - Progress tracker visual (stepper horizontal):
    ```
    [Pengajuan] → [Bimbingan] → [Seminar] → [Sidang] → [Lulus]
    ```
    Setiap step: icon + label + status (done/current/upcoming)
  - Info dosen pembimbing (nama, email, foto)
  - Tombol aksi sesuai status saat ini
  - Timeline aktivitas terbaru (5 entri terakhir: log bimbingan, dokumen, jadwal)

#### Modal Pengajuan Judul

**File:** `frontend/src/components/features/thesis/ThesisSubmitModal.tsx`

- [x] Form React Hook Form + Zod:
  ```ts
  const schema = z.object({
    title: z.string()
      .min(10, "Judul minimal 10 kata")
      .max(500, "Judul maksimal 500 karakter"),
    abstract: z.string().min(100, "Abstrak minimal 100 kata"),
    field_of_study: z.string().min(1, "Pilih bidang keahlian"),
    thesis_type: z.enum(["skripsi", "tugas_akhir"]),
  })
  ```
- [x] Counter kata real-time untuk judul dan abstrak
- [x] Dropdown bidang keahlian (list dari konstanta, bisa ditambah via env)
- [x] Konfirmasi sebelum submit: "Pastikan judul dan abstrak sudah benar sebelum diajukan."

### Halaman Bimbingan (`/student/consultations`)

**File:** `frontend/src/app/(dashboard)/student/consultations/page.tsx`

- [x] Summary card di atas: Total Bimbingan | Disetujui | Menunggu | Bimbingan Terakhir
- [x] Timeline log konsultasi (diurutkan dari terbaru):
  - Setiap entry: tanggal, topik, catatan, tindak lanjut, status badge
  - Jika ada attachment → link download
  - Status `pending` → bisa diedit atau dihapus
- [x] Filter: Status, Rentang Tanggal
- [x] Tombol "Catat Bimbingan Baru" → modal form
- [x] Pagination

#### Modal Form Log Bimbingan

**File:** `frontend/src/components/features/consultations/ConsultationFormModal.tsx`

- [x] Form:
  ```ts
  const schema = z.object({
    consultation_date: z.string().min(1, "Pilih tanggal"),
    topics_discussed: z.string().min(5, "Isi topik yang dibahas"),
    notes: z.string().optional(),
    follow_up: z.string().optional(),
  })
  ```
- [x] Date picker tidak bisa pilih tanggal masa depan
- [x] Mode create dan edit (prefill data)

### Halaman Dokumen (`/student/documents`)

**File:** `frontend/src/app/(dashboard)/student/documents/page.tsx`

- [x] Grid dokumen dikelompokkan per tipe:
  ```
  📄 Proposal Skripsi        [Disetujui v2]  [Lihat] [Riwayat]
  📄 Draft Bab 1             [Revisi v1]     [Upload Revisi]
  📄 Draft Bab 2             [Belum Upload]  [Upload]
  ...
  📄 Dokumen Seminar         [Menunggu Review v1]
  📄 Dokumen Sidang          [Terkunci 🔒]   ← belum bisa upload
  📄 Skripsi Final           [Terkunci 🔒]
  ```
- [x] Status visual per dokumen: warna dan icon sesuai status
- [x] Dokumen "Terkunci": tampil dengan overlay dan tooltip "Tersedia setelah lulus seminar"
- [x] Tombol aksi: Upload, Upload Revisi, Lihat (download), Riwayat Versi

#### Modal Upload Dokumen

**File:** `frontend/src/components/features/documents/DocumentUploadModal.tsx`

- [x] Komponen `FileUpload` dengan:
  - Drag & drop atau klik untuk pilih file
  - Hanya terima `.pdf`
  - Batas ukuran 10 MB dengan pesan error yang jelas
  - Preview nama file dan ukuran setelah dipilih
- [x] Field catatan (opsional): "Catatan untuk dosen reviewer"
- [x] Progress bar upload (saat menggunakan real storage di Job 21)
- [x] Sukses → toast + badge status dokumen berubah ke "Menunggu Review"

#### Modal Riwayat Versi

**File:** `frontend/src/components/features/documents/DocumentHistoryModal.tsx`

- [x] Tabel riwayat: Versi, Tanggal Upload, Status, Catatan Reviewer, Aksi (Download)
- [x] Versi aktif di-highlight

### Halaman Seminar Proposal (`/student/seminar`)

**File:** `frontend/src/app/(dashboard)/student/seminar/page.tsx`

- [x] Jika belum ada seminar → check gate:
  - Jika dokumen seminar belum approved → tampilkan info "Upload dan dapatkan persetujuan dokumen seminar terlebih dahulu" + link ke halaman dokumen
  - Jika dokumen sudah approved → tampilkan tombol "Ajukan Seminar Proposal"
- [x] Jika seminar `pending` → card status "Menunggu penjadwalan dari Admin/Kaprodi"
- [x] Jika seminar `scheduled` → card jadwal:
  - Tanggal, waktu, ruangan
  - Daftar penguji
  - Countdown timer ke hari H
- [x] Jika seminar `passed`/`failed` → card hasil:
  - Nilai akhir (besar, prominent)
  - Grade category (A/B+/B/C/Tidak Lulus)
  - Breakdown nilai per komponen (anonim — tanpa nama penguji)
  - Status: Lulus / Tidak Lulus
  - Jika lulus → info langkah selanjutnya (persiapkan dokumen sidang)
- [x] Konfirmasi sebelum pengajuan seminar

### Halaman Sidang Skripsi (`/student/defense`)

**File:** `frontend/src/app/(dashboard)/student/defense/page.tsx`

- [x] Struktur sama seperti halaman seminar, dengan perbedaan:
  - Gate: seminar harus `passed` + dokumen sidang harus `approved`
  - Jika gate belum terpenuhi → tampilkan checklist prasyarat dengan status tiap item
  - Jika `passed` → tampilkan informasi yudisium dan link upload skripsi final

### Halaman Arsip (`/archives`) — Shared semua role

**File:** `frontend/src/app/(dashboard)/archives/page.tsx`

- [x] Search bar full-text yang prominent di atas halaman
- [x] Filter collapsible: Tahun Lulus, Bidang Keahlian, Program Studi
- [x] Grid card arsip (bukan tabel) — lebih visual:
  ```
  ┌─────────────────────────────────┐
  │ 📘 Judul Skripsi (truncated)    │
  │ 👤 Nama Mahasiswa · NIM         │
  │ 👨‍🏫 Pembimbing: Dr. Ahmad        │
  │ 🏷 Machine Learning · 2027      │
  │            [Lihat Detail]        │
  └─────────────────────────────────┘
  ```
- [x] Pagination
- [x] Halaman detail arsip (`/archives/:id`):
  - Info lengkap: judul, penulis, pembimbing, penguji, abstrak, kata kunci
  - Tombol "Download Skripsi" (role-based: mahasiswa hanya miliknya sendiri)

---

## Done Criteria

- [x] `/student/thesis` → empty state jika belum ada thesis, form pengajuan berfungsi
- [x] Pengajuan judul → status berubah ke "Menunggu Review", toast sukses
- [x] Counter kata real-time berfungsi pada form judul dan abstrak
- [x] `/student/consultations` → create log bimbingan, muncul di timeline
- [x] `/student/documents` → grid dokumen dengan status lock/unlock sesuai tahap
- [x] Upload PDF valid → status "Menunggu Review"
- [x] Upload bukan PDF → error "Hanya file PDF yang diizinkan"
- [x] Upload >10MB → error ukuran file
- [x] Riwayat versi dokumen tampil di modal
- [x] `/student/seminar` → gate logic visual berfungsi (locked/unlocked)
- [x] Setelah seminar dijadwalkan → info jadwal tampil dengan countdown
- [x] Hasil seminar → nilai dan breakdown tampil
- [x] `/archives` → search full-text berfungsi, hasil diurutkan by relevance
- [x] Download arsip oleh mahasiswa lain → error 403, tampil toast "Akses ditolak"
- [x] Semua halaman responsive
