# Job 18 — Frontend Halaman Dosen (Pembimbing & Penguji)

**Phase:** 4 — Frontend
**Referensi PRD:** Section 5 (Target Users — Dosen Pembimbing & Dosen Penguji), Section 6.3–6.6
**Prerequisites:** Job 17 (Frontend Mahasiswa) ✅
**Estimasi:** 3 hari

---

## Objective

Implementasi semua halaman untuk Dosen Pembimbing (mahasiswa bimbingan, review dokumen, log konsultasi) dan Dosen Penguji (jadwal pengujian, input nilai). Kedua role ini memiliki halaman yang berbeda namun berbagi beberapa komponen.

---

## Checklist

### DOSEN PEMBIMBING

### Halaman Mahasiswa Bimbingan (`/supervisor/students`)

**File:** `frontend/src/app/(dashboard)/supervisor/students/page.tsx`

- [x] Kartu ringkasan di atas: Total Mahasiswa | Dokumen Pending Review | Jadwal Minggu Ini
- [x] Tabel mahasiswa bimbingan:
  - Kolom: Nama & NIM, Judul (truncated), Status, Bimbingan Terakhir, Dokumen Pending, Aksi
  - Badge status thesis
  - Highlight baris: mahasiswa yang >14 hari tidak bimbingan → warna warning
  - Kolom "Bimbingan Terakhir": tampilkan "X hari lalu" (relative time)
- [x] Filter: Status thesis
- [x] Klik baris → halaman detail mahasiswa bimbingan

### Halaman Detail Mahasiswa Bimbingan (`/supervisor/students/:thesis_id`)

**File:** `frontend/src/app/(dashboard)/supervisor/students/[thesisId]/page.tsx`

- [x] Breadcrumb: Dashboard > Mahasiswa Bimbingan > Nama Mahasiswa
- [x] Header: info mahasiswa (nama, NIM, judul, status)
- [x] Tab navigation:
  - **Log Bimbingan** — timeline konsultasi + tombol "Catat Bimbingan"
  - **Dokumen** — daftar dokumen yang perlu/sudah di-review
  - **Progress** — progress tracker tahapan + info jadwal seminar/sidang

#### Tab Log Bimbingan
- [x] Timeline sama seperti yang mahasiswa lihat
- [x] Log dengan status `pending` → tombol "Setujui" di samping entry
- [x] Konfirmasi sebelum approve: "Setujui log bimbingan ini?"
- [x] Tombol "Catat Bimbingan" → modal form (sama dengan milik mahasiswa)

#### Tab Dokumen
- [x] List dokumen per tipe, hanya tampilkan yang relevan untuk pembimbing
- [x] Dokumen `pending_review` → card dengan highlight + tombol "Review Sekarang"
- [x] Klik "Review" → modal review dokumen

#### Modal Review Dokumen

**File:** `frontend/src/components/features/documents/DocumentReviewModal.tsx`

- [x] Tampilkan info dokumen: nama file, tipe, versi, tanggal upload, ukuran
- [x] Tombol "Preview / Download" → buka presigned URL di tab baru
- [x] Form keputusan:
  ```ts
  const reviewSchema = z.object({
    decision: z.enum(["approved", "revision_required"]),
    notes: z.string().optional(),
  }).refine(
    (d) => d.decision === "approved" || (d.notes && d.notes.length > 10),
    { message: "Catatan wajib diisi saat meminta revisi", path: ["notes"] }
  )
  ```
- [x] Dua tombol jelas: "✅ Setujui" (hijau) dan "🔄 Minta Revisi" (orange)
- [x] Jika "Minta Revisi" dipilih → field notes menjadi required dengan animasi
- [x] Sukses → badge status dokumen berubah, toast notifikasi

### Halaman Dokumen Pending Review (`/supervisor/documents`)

**File:** `frontend/src/app/(dashboard)/supervisor/documents/page.tsx`

- [x] Agregasi semua dokumen `pending_review` dari semua mahasiswa bimbingan
- [x] Tabel: Mahasiswa, Tipe Dokumen, Versi, Tanggal Upload, Aksi
- [x] Urutkan dari yang paling lama pending (terlama dulu)
- [x] Tombol "Review" → modal DocumentReviewModal
- [x] Filter: Mahasiswa (search), Tipe Dokumen
- [x] Badge count di sidebar menu jika ada pending

### Halaman Jadwal Dosen Pembimbing (`/supervisor/schedules`)

**File:** `frontend/src/app/(dashboard)/supervisor/schedules/page.tsx`

- [x] Daftar seminar dan sidang mahasiswa bimbingan yang terjadwal
- [x] Tab: Mendatang | Sudah Lewat
- [x] Info per jadwal: mahasiswa, judul, tipe (seminar/sidang), tanggal, waktu, ruangan, penguji
- [x] Tampilan kalender sederhana (list by tanggal)

---

### DOSEN PENGUJI

### Halaman Jadwal Pengujian (`/examiner/schedules`)

**File:** `frontend/src/app/(dashboard)/examiner/schedules/page.tsx`

- [x] Header cards: Jadwal Mendatang | Nilai Belum Diinput | Sudah Dinilai
- [x] Tab: Mendatang | Perlu Input Nilai | Sudah Selesai
- [x] Setiap jadwal: card dengan info lengkap
  - Tipe: badge "Seminar Proposal" atau "Sidang Skripsi"
  - Mahasiswa: nama, NIM, foto
  - Judul skripsi
  - Tanggal, waktu, ruangan
  - Status nilai: "Belum Dinilai" (orange) / "Sudah Dinilai" (hijau)
  - Tombol "Input Nilai" jika belum dinilai
  - Tombol "Lihat Nilai" jika sudah dinilai
- [x] Notifikasi badge di sidebar jika ada yang belum dinilai

### Halaman Input Nilai (`/examiner/scoring/:seminar_id` atau `/:defense_id`)

**File:** `frontend/src/app/(dashboard)/examiner/scoring/[type]/[id]/page.tsx`

- [x] Header: info seminar/sidang (mahasiswa, judul, tanggal)
- [x] Link download dokumen skripsi sebelum menilai
- [x] Form penilaian per komponen:
  ```tsx
  // Komponen penilaian ditampilkan dinamis dari API (bobot per komponen)
  {components.map((comp) => (
    <div key={comp.name}>
      <label>{comp.name} (Bobot: {comp.weight}%)</label>
      <input
        type="number"
        min={0}
        max={100}
        placeholder="0 – 100"
      />
      {/* Preview nilai terbobot: (score × weight / 100) */}
      <span>Kontribusi: {weighted.toFixed(1)} poin</span>
    </div>
  ))}
  ```
- [x] Preview nilai akhir real-time (dihitung di frontend sebelum submit):
  ```
  ┌─────────────────────────────────────────┐
  │  Preview Nilai Anda                      │
  │  Presentasi         85 × 30% = 25.5     │
  │  Penguasaan Materi  78 × 30% = 23.4     │
  │  Kualitas Naskah    80 × 25% = 20.0     │
  │  Kemampuan Menjawab 75 × 15% = 11.25   │
  │  ─────────────────────────────          │
  │  Total Nilai Anda:         80.15        │
  └─────────────────────────────────────────┘
  ```
- [x] Konfirmasi sebelum submit: "Nilai tidak dapat diubah setelah disimpan. Lanjutkan?"
- [x] Sukses → redirect ke halaman jadwal, toast "Nilai berhasil disimpan"
- [x] Jika sudah pernah submit → halaman ini menampilkan nilai yang sudah diinput (read-only)

### Halaman Riwayat Penilaian (`/examiner/scoring`)

**File:** `frontend/src/app/(dashboard)/examiner/scoring/page.tsx`

- [x] List semua seminar/sidang yang sudah dinilai oleh penguji ini
- [x] Tabel: Mahasiswa, Judul, Tipe, Tanggal, Nilai Saya, Nilai Akhir (jika semua sudah submit)
- [x] Klik row → detail breakdown nilai

---

## Done Criteria

### Dosen Pembimbing
- [x] `/supervisor/students` → list mahasiswa dengan highlight warning jika >14 hari tidak bimbingan
- [x] Klik mahasiswa → halaman detail dengan 3 tab berfungsi
- [x] Approve log bimbingan dari tab → status berubah + toast
- [x] Review dokumen (approve) → badge status dokumen berubah ke "Disetujui"
- [x] Review dokumen (revisi) → notes wajib diisi, status berubah ke "Perlu Revisi"
- [x] Preview dokumen (download) → presigned URL terbuka di tab baru
- [x] `/supervisor/documents` → agregasi semua dokumen pending, urutkan dari terlama
- [x] Badge count di sidebar bertambah/berkurang sesuai dokumen pending

### Dosen Penguji
- [x] `/examiner/schedules` → jadwal tampil dengan status "Belum Dinilai" / "Sudah Dinilai"
- [x] Form input nilai → preview nilai real-time terhitung otomatis
- [x] Submit nilai → konfirmasi muncul, setelah submit form menjadi read-only
- [x] Akses halaman input nilai yang sudah disubmit → tampil read-only
- [x] Badge count di sidebar untuk pending scoring

### Shared
- [x] Semua halaman responsive
- [x] Loading skeleton muncul saat fetch
- [x] Error state informatif jika API error
