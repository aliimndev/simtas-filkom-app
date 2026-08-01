# Job 20 — Frontend Dashboard Per Role

**Phase:** 4 — Frontend
**Referensi PRD:** Section 6.9 (FR-DASHBOARD-001 s/d FR-DASHBOARD-004)
**Prerequisites:** Job 19 (Frontend Kaprodi) ✅
**Estimasi:** 3 hari

---

## Objective

Implementasi halaman dashboard utama (`/dashboard`) yang menampilkan konten berbeda untuk setiap role. Admin dan Kaprodi mendapat dashboard analitik komprehensif; Mahasiswa mendapat progress tracker personal; Dosen Pembimbing mendapat ringkasan bimbingan; Dosen Penguji mendapat jadwal dan pending scoring. Setelah job ini selesai, seluruh frontend Phase 4 selesai.

---

## Checklist

### Route Dashboard Utama

**File:** `frontend/src/app/(dashboard)/dashboard/page.tsx`

- [ ] Render komponen dashboard sesuai role dari auth store:
  ```tsx
  export default function DashboardPage() {
    const { user } = useAuth()

    switch (user?.role) {
      case "admin_fakultas": return <AdminDashboard />
      case "kaprodi":        return <KaprodiDashboard />
      case "mahasiswa":      return <StudentDashboard />
      case "dosen_pembimbing": return <SupervisorDashboard />
      case "dosen_penguji":  return <ExaminerDashboard />
      default: return <LoadingSpinner />
    }
  }
  ```

---

### Dashboard Admin & Kaprodi

**File:** `frontend/src/components/features/dashboard/AdminDashboard.tsx`
**File:** `frontend/src/components/features/dashboard/KaprodiDashboard.tsx`

> Kedua komponen hampir identik — buat `AdminKaprodiDashboardBase` sebagai shared component, perbedaan hanya di tombol aksi yang tersedia.

#### Filter Bar (Sticky di atas)
- [ ] Dropdown: Tahun Akademik (default: aktif), Semester, Program Studi
- [ ] Filter ini mempengaruhi semua widget di bawahnya via React Query key

#### Row 1 — Summary Cards (4 kartu)
- [ ] Implementasikan komponen `StatCard`:
  ```tsx
  interface StatCardProps {
    title: string
    value: number | string
    subtitle?: string
    icon: LucideIcon
    trend?: { value: number; label: string }  // opsional perubahan dari bulan lalu
    color: "blue" | "green" | "orange" | "purple"
    href?: string  // jika clickable → redirect ke halaman terkait
  }
  ```
- [ ] 4 kartu: **Total Mahasiswa Aktif**, **Lulus Semester Ini**, **Seminar Bulan Ini**, **Sidang Bulan Ini**
- [ ] Setiap kartu clickable → link ke halaman terkait

#### Row 2 — Progress Funnel + Grafik Tren
- [ ] **Funnel Tahapan** (kiri, 60% lebar):
  - Bar horizontal untuk setiap status thesis
  - Label: Menunggu Review (8), Bimbingan (35), Siap Seminar (12), ...
  - Setiap bar clickable → filter `/kaprodi/theses` dengan status tersebut
  - Gunakan komponen chart sederhana dengan CSS/SVG (tanpa library berat)
- [ ] **Grafik Tren Bulanan** (kanan, 40% lebar):
  - Line chart jumlah lulusan per bulan (6 bulan terakhir)
  - Gunakan `Recharts` (install: `npm install recharts`)

#### Row 3 — Analitik Dosen + Operasional
- [ ] **Beban Dosen** (kiri, 50%):
  - Tabel: Nama Dosen, Bimbingan Aktif, Seminar, Sidang
  - Sort: beban tertinggi di atas
  - Warna baris: hijau (<3), kuning (3-5), merah (>5)
  - Tampilkan max 8 dosen, link "Lihat Semua"
- [ ] **Pending Actions** (kanan, 50%):
  - List item dengan count dan link:
    ```
    📋  5 Pengajuan judul menunggu review   →  /kaprodi/thesis-reviews
    📄 12 Dokumen menunggu review          →  /supervisor/documents
    🎓  3 Seminar menunggu penjadwalan     →  /kaprodi/seminars
    ⚖️   2 Sidang menunggu penjadwalan     →  /kaprodi/defenses
    ```
  - Jika count = 0 → item tersebut tidak ditampilkan atau di-grey out
  - Badge merah jika count > 0

#### Row 4 — Jadwal Mendatang
- [ ] List seminar dan sidang dalam 14 hari ke depan
- [ ] Group by tanggal
- [ ] Setiap item: icon tipe (seminar/sidang), nama mahasiswa, judul, waktu, ruangan
- [ ] Tampilkan max 5 item, link "Lihat Semua Jadwal"

#### Auto-Refresh
- [ ] Dashboard di-refetch setiap 5 menit menggunakan TanStack Query `refetchInterval: 5 * 60 * 1000`

---

### Dashboard Mahasiswa

**File:** `frontend/src/components/features/dashboard/StudentDashboard.tsx`

- [ ] **Greeting Header**: "Selamat pagi, [Nama]! 👋"
- [ ] **Progress Stepper** (prominent, full width):
  ```
  ① Pengajuan  →  ② Bimbingan  →  ③ Seminar  →  ④ Sidang  →  ⑤ Lulus
      ✅               ✅ (aktif)        ⏳               🔒          🔒
  ```
  - Step yang sudah lewat: filled + checkmark
  - Step aktif: filled + animasi pulse
  - Step mendatang: outline/grey
- [ ] **Thesis Info Card**: judul, pembimbing, status badge
- [ ] **Quick Actions** (tombol besar, sesuai stage saat ini):
  - `in_progress` → "Upload Dokumen" + "Catat Bimbingan"
  - `seminar_ready` → "Lihat Jadwal Seminar"
  - `defense_ready` → "Lihat Jadwal Sidang"
  - dll.
- [ ] **Status Dokumen** (grid kartu kecil):
  - Setiap tipe dokumen: icon + nama + badge status
  - Dokumen yang perlu tindakan di-highlight
- [ ] **Bimbingan**: card ringkasan (total bimbingan, bimbingan terakhir, tindak lanjut terakhir)
- [ ] **Jadwal Mendatang**: card jika ada seminar/sidang terjadwal (countdown)
- [ ] **Jika belum punya thesis**: hero card besar dengan CTA "Mulai Ajukan Judul Skripsi Anda"

---

### Dashboard Dosen Pembimbing

**File:** `frontend/src/components/features/dashboard/SupervisorDashboard.tsx`

- [ ] **Summary cards**: Total Bimbingan | Dokumen Pending Review | Jadwal Minggu Ini
- [ ] **Mahasiswa yang Butuh Perhatian** (prioritas):
  - List mahasiswa yang >14 hari tidak ada aktivitas bimbingan
  - Link ke halaman detail masing-masing
- [ ] **Dokumen Pending Review**: top 3 dokumen terlama, link "Lihat Semua"
- [ ] **Jadwal Mendatang**: seminar/sidang mahasiswa bimbingan dalam 7 hari
- [ ] **Daftar Mahasiswa**: mini-tabel dengan status dan progress bar tiap mahasiswa

---

### Dashboard Dosen Penguji

**File:** `frontend/src/components/features/dashboard/ExaminerDashboard.tsx`

- [ ] **Summary cards**: Jadwal Mendatang | Nilai Belum Diinput | Total Sudah Dinilai
- [ ] **Perlu Tindakan** (jika ada pending):
  - Banner kuning: "Anda memiliki X penilaian yang belum diinput"
  - Tombol "Input Nilai Sekarang"
- [ ] **Jadwal Mendatang**: list seminar/sidang yang akan datang
- [ ] **Riwayat Penilaian**: 5 entri terakhir yang sudah dinilai

---

### Shared Dashboard Components

- [ ] **`StatCard`** — kartu statistik (digunakan semua dashboard)
- [ ] **`ProgressStepper`** — stepper horizontal dengan status per step
- [ ] **`UpcomingScheduleList`** — list jadwal mendatang (shared seminar + sidang)
- [ ] **`PendingActionList`** — list pending actions dengan count dan link
- [ ] **`MiniBarChart`** — bar chart sederhana menggunakan Recharts
- [ ] **`LineChart`** — line chart trend menggunakan Recharts

---

## Done Criteria

- [ ] Login sebagai Admin → `AdminDashboard` muncul dengan semua widget
- [ ] Login sebagai Kaprodi → `KaprodiDashboard` muncul
- [ ] Login sebagai Mahasiswa → `StudentDashboard` dengan progress stepper sesuai status thesis
- [ ] Mahasiswa belum punya thesis → CTA pengajuan judul muncul
- [ ] Login sebagai Dosen Pembimbing → `SupervisorDashboard` dengan mahasiswa yang butuh perhatian
- [ ] Login sebagai Dosen Penguji → `ExaminerDashboard` dengan pending badge jika ada
- [ ] Filter tahun akademik pada AdminDashboard → semua widget ter-refresh
- [ ] StatCard clickable → redirect ke halaman terkait
- [ ] Funnel bar clickable → redirect ke `/kaprodi/theses?status=xxx`
- [ ] Dashboard auto-refresh setiap 5 menit (tanpa reload halaman)
- [ ] Semua dashboard responsive, tampil baik di tablet
- [ ] **MILESTONE Phase 4:** Seluruh frontend SIMTAS FILKOM v1.0 selesai diimplementasi
