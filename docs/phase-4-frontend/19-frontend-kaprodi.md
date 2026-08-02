# Job 19 — Frontend Halaman Kaprodi

**Phase:** 4 — Frontend
**Referensi PRD:** Section 5 (Target Users — Kaprodi), Section 6.2, 6.5, 6.6, 6.7
**Prerequisites:** Job 18 (Frontend Dosen) ✅
**Estimasi:** 3 hari

---

## Objective

Implementasi semua halaman untuk Kaprodi: review pengajuan judul, penunjukan pembimbing, penjadwalan seminar dan sidang, penetapan yudisium, dan monitoring seluruh mahasiswa. Kaprodi memiliki authority terbesar dalam alur akademik.

---

## Checklist

### Halaman Review Pengajuan Judul (`/kaprodi/thesis-reviews`)

**File:** `frontend/src/app/(dashboard)/kaprodi/thesis-reviews/page.tsx`

- [x] Tab: **Menunggu Review** | **Disetujui** | **Ditolak** | **Semua**
- [x] Badge count di tab "Menunggu Review"
- [x] Tabel per tab:
  - Kolom: Mahasiswa (nama+NIM), Judul, Bidang, Jenis TA, Tanggal Pengajuan, Aksi
  - Judul bisa di-hover untuk lihat abstrak (tooltip atau expand)
- [x] Tombol aksi di tab "Menunggu Review": "Review" → buka modal
- [x] Filter: Tahun Akademik, Program Studi, Bidang Keahlian

#### Modal Review Pengajuan

**File:** `frontend/src/components/features/thesis/ThesisReviewModal.tsx`

- [x] Tampilkan info lengkap pengajuan:
  - Nama & NIM mahasiswa
  - Judul lengkap
  - Abstrak (scrollable jika panjang)
  - Bidang keahlian, jenis TA
  - Tanggal pengajuan
- [x] Form keputusan:
  ```ts
  const reviewSchema = z.object({
    decision: z.enum(["approved", "rejected"]),
    notes: z.string().optional(),
  }).refine(
    (d) => d.decision === "approved" || (d.notes && d.notes.length > 10),
    { message: "Catatan wajib diisi saat menolak pengajuan", path: ["notes"] }
  )
  ```
- [x] Dua tombol aksi: "✅ Setujui" dan "❌ Tolak"
- [x] Setelah approve → langsung buka modal penunjukan pembimbing (chain modal)

#### Modal Penunjukan Dosen Pembimbing

**File:** `frontend/src/components/features/thesis/AssignSupervisorModal.tsx`

- [x] Tampilkan judul thesis yang baru disetujui
- [x] Pilih 1 atau 2 dosen pembimbing dari daftar:
  - List dosen dengan info: nama, NIDN, beban bimbingan saat ini
  - Sort dari beban terendah (rekomendasi load balancing)
  - Badge: "Beban Ringan" (<3), "Beban Sedang" (3-5), "Beban Berat" (>5)
  - Multi-select dengan checkbox, max 2
- [x] Validasi: minimal 1 dosen dipilih
- [x] Sukses → toast "Pembimbing berhasil ditetapkan, email notifikasi terkirim"

### Halaman Semua Thesis (`/kaprodi/theses`)

**File:** `frontend/src/app/(dashboard)/kaprodi/theses/page.tsx`

- [x] Tabel semua thesis dengan filter lengkap:
  - Search (judul, nama, NIM)
  - Filter Status (dropdown multi-select)
  - Filter Tahun Akademik
  - Filter Program Studi
  - Filter Dosen Pembimbing
- [x] Kolom: No, Mahasiswa, Judul, Pembimbing, Status, Tanggal Pengajuan, Progres, Aksi
- [x] Kolom "Progres": progress bar visual 0–100%
- [x] Klik baris → halaman detail thesis (lihat semua info, log, dokumen, jadwal)
- [x] Export button → `GET /reports/theses/export?format=excel` (jika v1.1 sudah tersedia)

### Halaman Detail Thesis (Kaprodi View) (`/kaprodi/theses/:id`)

**File:** `frontend/src/app/(dashboard)/kaprodi/theses/[id]/page.tsx`

- [x] Halaman komprehensif dengan tab:
  - **Overview**: info dasar, pembimbing, progress
  - **Dokumen**: semua dokumen dan statusnya
  - **Bimbingan**: riwayat log konsultasi
  - **Seminar**: info seminar + nilai
  - **Sidang**: info sidang + nilai
- [x] Tombol aksi sesuai status thesis saat ini:
  - Status `submitted` → "Review Pengajuan"
  - Status `approved` → "Tetapkan Pembimbing"
  - Status `defense_done` → "Tetapkan Yudisium"
  - Semua status → "Batalkan Thesis" (dengan konfirmasi)

### Halaman Manajemen Seminar (`/kaprodi/seminars`)

**File:** `frontend/src/app/(dashboard)/kaprodi/seminars/page.tsx`

- [x] Tab: **Menunggu Jadwal** | **Terjadwal** | **Selesai**
- [x] Badge count di "Menunggu Jadwal"
- [x] Tabel per tab dengan info relevan
- [x] Tombol "Jadwalkan" → modal penjadwalan (sama dengan milik Admin, `ScheduleModal`)
- [x] Tombol "Ubah Jadwal" → modal reschedule
- [x] Filter: Tahun Akademik, Program Studi, Tanggal

### Halaman Manajemen Sidang (`/kaprodi/defenses`)

**File:** `frontend/src/app/(dashboard)/kaprodi/defenses/page.tsx`

- [x] Struktur sama persis dengan halaman seminar
- [x] Tombol tambahan di baris yang sudah selesai dan lulus: "Tetapkan Yudisium"

#### Modal Penetapan Yudisium

**File:** `frontend/src/components/features/defense/GraduationModal.tsx`

- [x] Tampilkan info:
  - Nama mahasiswa, NIM, judul
  - Nilai sidang akhir dan grade
  - Checklist prasyarat (semua harus ✅):
    - ✅ Seminar Proposal: Lulus
    - ✅ Sidang Skripsi: Lulus
    - ✅ Dokumen Final: Disetujui
- [x] Field catatan (opsional): "Pesan selamat / catatan yudisium"
- [x] Tombol "Tetapkan Lulus" — warna hijau prominent
- [x] Konfirmasi tegas: "Tindakan ini tidak dapat dibatalkan. Mahasiswa akan dinyatakan LULUS secara resmi."
- [x] Sukses → toast "Yudisium berhasil. Email selamat terkirim ke mahasiswa."

### Halaman Monitoring (`/kaprodi/monitoring`)

**File:** `frontend/src/app/(dashboard)/kaprodi/monitoring/page.tsx`

- [x] Tampilan tabel ringkas semua mahasiswa aktif dengan:
  - Status tahap saat ini
  - Hari sejak update terakhir (flag mahasiswa yang "mandek")
  - Quick filter: mahasiswa yang >30 hari tidak ada aktivitas
- [x] Berguna untuk identifikasi mahasiswa yang perlu perhatian khusus

---

## Done Criteria

- [x] `/kaprodi/thesis-reviews` → tab "Menunggu Review" tampil count badge
- [x] Modal review → setujui thesis → otomatis buka modal penunjukan pembimbing
- [x] Modal penunjukan → list dosen diurutkan dari beban terendah
- [x] Setelah penetapan pembimbing → thesis status berubah ke "Bimbingan"
- [x] Tolak thesis → field catatan wajib diisi
- [x] `/kaprodi/seminars` → jadwalkan seminar via modal berfungsi
- [x] Modal reschedule → perubahan jadwal berhasil, tab pindah ke "Terjadwal"
- [x] Modal yudisium → checklist prasyarat tampil sebelum tombol aktif
- [x] Tetapkan yudisium → status berubah ke "Lulus", konfirmasi muncul sebelumnya
- [x] `/kaprodi/theses` → filter lengkap berfungsi, progress bar akurat
- [x] Semua halaman responsive
- [x] Loading dan error state ditangani dengan baik
