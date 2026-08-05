# Daftar Pertanyaan untuk Kaprodi FILKOM — Validasi Alur SIMTAS

> **Tujuan:** Memastikan SIMTAS FILKOM (Sistem Manajemen Tugas Skripsi, target rilis 2026, deployment produksi) sesuai dengan alur resmi fakultas.
> **Konteks:** Belum punya dokumen pedoman → minta dokumen dulu (Langkah 0).
> **Sumber asumsi:** codebase SIMTAS (PRD, migrasi DB, entity, use case).

**Legenda prioritas**
- **P0** — menyimpang = rusak inti state machine / blokir deployment.
- **P1** — menyimpang = ubah modul lokal (patch/penyesuaian).
- **P2** — edge-case / detail.

---

## Langkah 0 — Minta dokumen (P0, jadikan pembuka pertemuan)

1. Minta **Panduan Akademik / Pedoman Penyusunan Skripsi & SOP Bimbingan TA** FILKOM versi terbaru.
2. Minta **SK/SOP penilaian** (komponen & bobot) dan **SK mekanisme seminar/sidang**.
3. Tanya: dokumen mana yang berlaku sekarang, dan apakah ada revisi yang direncanakan.

> Catatan: situs FILKOM (`filkom.unida.ac.id`) punya menu *Academic Guide* dan *Repository* — minta file resminya. Banyak jawaban P0/P1 biasanya sudah ada di pedoman, sehingga pertemuan jadi konfirmasi cepat, bukan menggali dari nol.

---

## P0 — Struktur inti & deployment-readiness

### [W1] Seminar proposal — D1.1
- **Asumsi SIMTAS:** ada seminar proposal sebagai gate wajib sebelum sidang (`seminar_ready → seminar_done → defense_ready`).
- **Tanya:** "Apakah FILKOM mewajibkan **seminar proposal** sebelum sidang skripsi?"
- **Jika TIDAK:** "Tahap pra-sidang resmi apa? (seminar hasil / kompre / sidang langsung?)"
- **Dampak:** modul seminar (Job 08) dihapus/diganti; state machine berubah.

### [W3] Komponen & bobot penilaian — D5.1
- **Asumsi:** 4 komponen bobot tetap — Presentasi 30, Penguasaan Materi 30, Kualitas Naskah 25, Kemampuan Menjawab 15 — sama untuk seminar & sidang.
- **Tanya:** "Apakah komponen & bobot ini resmi? Apakah bobot seminar = sidang?"
- **Jika berbeda:** "Berapa komponen & bobot resminya?"
- **Dampak:** ubah `backend/internal/domain/entity/grading.go`; pertimbangkan dinamiskan (v2→v1) agar tak hardcode tiap perubahan.
- **[D5.3]** Tanya ambang lulus: "<60 gagal, 60–74 revisi, ≥75 lulus — benar?"
- **[D5.4]** Tanya cara hitung final: "rata-rata semua penguji?"
- **[D5.5]** Tanya skala huruf (A–E) & rentang resminya.

### [W2] Yudisium — D1.3
- **Asumsi:** yudisium ditetapkan manual oleh Kaprodi.
- **Tanya:** "Apakah penetapan lulus manual oleh Kaprodi, atau otomatis saat semua penguji selesai & revisi disetujui?"
- **Dampak bila otomatis:** transisi `graduated` jadi otomatis; gate manual Kaprodi jadi konfirmasi saja.

### [W6] Integrasi data pusat — D10.1
- **Asumsi:** v1 tanpa integrasi SIAKAD.
- **Tanya:** "Untuk produksi, apakah wajib sinkron data mahasiswa/dosen dari SIAKAD/feeder PDDIKTI? Format sync (push/pull, real-time/batch)?"
- **Dampak bila wajib:** deployment tertahan; integrasi v2→v1; butuh adapter data.
- **[D10.2]** Tanya format NIM & NIDN resmi (pola/berapa digit).

### [W11] Nama program studi — D10.3 ⭐
- **Fakta (dari situs FILKOM):** prodi resmi = **Ilmu Komputer (Computer Science, S.Kom)** — seed SIMTAS menulis `'Teknik Informatika'`, kemungkinan tidak akurat.
- **Tanya:** "Prodi resmi FILKOM = Ilmu Komputer? Apakah akan ada prodi baru?"
- **Tanya:** "Daftar bidang keahlian resmi (untuk `field_of_study`) atau bebas?"
- **Dampak:** koreksi seed `'Teknik Informatika'` → `'Ilmu Komputer'`; mungkin field prodi jadi enum/1 opsi.

---

## P1 — Modul lokal (menyimpang = patch/penyesuaian)

### [W4] Dosen pembimbing — D3
- **Asumsi:** 1–2 pembimbing, setara, ditunjuk Kaprodi.
- **[D3.1]** Tanya jumlah resmi pembimbing.
- **[D3.2]** Tanya: "Ada peran khusus (Pembimbing Utama/UTU) dengan bobot beda?" → dampak: ubah validasi + tambah flag `is_primary` di `thesis_supervisors`.
- **[D3.3]** Tanya siapa berwenang mengganti pembimbing.
- **[D3.4]** Tanya beban maksimal pembimbingan per dosen.

### [W5] Dosen penguji & rangkap — D4
- **Asumsi:** banyak penguji (many-to-many), assigned, bobot sama.
- **[D4.1]** Tanya jumlah penguji resmi.
- **[D4.2]** Tanya siapa menunjuk penguji (Kaprodi/Kajur).
- **[D4.3]** Tanya: "Bolehkah pembimbing jadi penguji untuk skripsi yang sama?" → jika dilarang: tambah validasi konflik kepentingan saat assign examiner.
- **[D4.4]** Tanya apakah ada penguji eksternal.
- **[D4.5]** Tanya bobot tiap penguji di final (sama?).

### [W7] Channel notifikasi — D9
- **Asumsi:** email (`@filkom.unida.ac.id`) via Resend.
- **[D9.1]** Tanya channel resmi: email, WhatsApp, atau keduanya? → bila WA utama: modul notifikasi email-only kurang; integrasi WA v2→v1.
- **[D9.2]** Tanya event wajib notifikasi (7 trigger cukup/kurang/lebih?).
- **[D9.3]** Tanya notifikasi ke pembimbing untuk semua event?

### [W8] Akses arsip — D8
- **Asumsi:** arsip public-read untuk civitas, presigned URL 30 menit.
- **[D8.3]** Tanya: "Arsip skripsi wajib public, atau ada embargo/batas akses?" → bila embargo: ubah model akses arsip → privat + approval.
- **[D8.1]** Tanya frekuensi yudisium (akhir semester? per periode?).
- **[D8.4]** Tanya kebutuhan lembar pengesahan/tanda tangan digital.

### [W9] Peran tambahan — D2
- **Asumsi:** 5 role (`admin_fakultas`, `kaprodi`, `mahasiswa`, `dosen_pembimbing`, `dosen_penguji`).
- **[D2.1]** Tanya: Admin Fakultas & Kaprodi benar terpisah dengan otoritas berbeda?
- **[D2.2]** Tanya siapa menunjuk pembimbing (Kaprodi/Dekan/Kajur).
- **[D2.4]** Tanya peran lain (Koordinator TA, Sekretaris Prodi) → dampak: tambah role + sesuaikan matrix RBAC.

---

## P2 — Edge-case & detail

### Tahapan & state machine (D1) — sisa
- **[D1.2]** Revisi pasca-seminar wajib sebelum sidang?
- **[D1.4]** Aturan pengajuan ulang judul jika ditolak (tunggu/syarat)?
- **[D1.5]** SLA antar tahap (max bulan seminar→sidang)?

### Dokumen wajib & approval (D6)
- **[D6.1]** Dokumen wajib tiap tahap (proposal, per-BAB, naskah lengkap)?
- **[D6.2]** Siapa approve tiap dokumen — pembimbing, kaprodi, atau keduanya?
- **[D6.3]** Format & batas ukuran (PDF only? max MB)?
- **[D6.4]** Butuh lembar pengesahan/tanda tangan?
- **[D6.5]** Revisi versi lama wajib tersimpan (versioning)?

### Penjadwalan seminar/sidang (D7)
- **[D7.1]** Siapa penjadwal resmi (Admin/Kaprodi/Kajur)?
- **[D7.2]** Ada kuota/jendela per periode?
- **[D7.3]** Seminar/sidang terbuka/tertutup?
- **[D7.4]** Format ruang/venue & apakah bisa online?
- **[D7.5]** Min H-? pemberitahuan jadwal?

### Aturan akademik khusus (D11)
- **[D11.1]** Boleh >1 TA aktif (ganti topik)?
- **[D11.2]** Aturan ganti judul mid-way?
- **[D11.3]** Masa kedaluwarsa skripsi (max semester)?
- **[D11.4]** Status cuti/undur diri vs TA aktif?
- **[D11.5]** Aturan judul min/maks kata?

### [W10] Batas revisi & kedaluwarsa — D5.6 / D11.3
- **Asumsi:** tanpa batas revisi/kedaluwarsa.
- **Tanya:** "Maks berapa kali revisi? Berapa semester maksimal TA?"
- **Dampak bila ada:** tambah counter revisi & cek expiry.

---

## Ringkasan dampak teknis bila jawaban menyimpang

| Pivot | Dampak ke codebase SIMTAS |
|---|---|
| W1 — seminar tak ada | hapus/ganti modul seminar (Job 08), ubah state machine |
| W2 — yudisium otomatis | transisi `graduated` otomatis (hapus gate manual) |
| W3 — bobot beda | ubah `grading.go`; mungkin dinamiskan (v2→v1) |
| W4 — pembimbing beda | ubah validasi min/max + flag `is_primary` |
| W5 — rangkap dilarang | validasi examiner ≠ supervisor thesis itu |
| W6 — SIAKAD wajib | integrasi v2→v1; deployment tertahan; butuh adapter |
| W7 — WhatsApp utama | notifikasi WA v2→v1 |
| W8 — arsip embargo | model akses arsip → privat + approval |
| W9 — peran tambahan | tambah role + sesuaikan RBAC matrix |
| W10 — batas revisi | counter revisi + cek expiry |
| W11 — nama prodi | koreksi seed `'Teknik Informatika'` → `'Ilmu Komputer'` |

---

## Tips pertemuan
- Buka dengan **Langkah 0** (minta dokumen) — bisa memperpendek banyak P0/P1.
- Kerjakan **P0** lebih dulu; jika salah satu menyimpang, blok relevan di P1/P2 bisa ikut berubah.
- Bawa tabel "asumsi SIMTAS saat ini" agar Kaprodi tahu apa yang sedang divalidasi (kolom "Asumsi" di tiap item).
- Untuk tiap jawaban, konfirmasi: "apakah aturan ini tertulis di dokumen, atau kebijakan ad-hoc?" — penting untuk audit-trail produksi.
