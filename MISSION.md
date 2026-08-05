# MISSION — Memilih & Mengerti Penyimpanan Objek untuk SIMTAS FILKOM

**Konteks project:** SIMTAS FILKOM (Sistem Manajemen Tugas Akhir & Skripsi)
adalah monorepo dengan backend **Go + PostgreSQL** dan frontend **Next.js**.
Dokumen skripsi (PDF) dan arsip diunggah dan disimpan ke *object storage*
(sekarang direncanakan Supabase Storage — lihat `docs/phase-5-integration/21-storage-integration.md`).

## Kenapa belajar ini (misi pengguna)

Kami sedang menuju **deployment produksi di lingkungan kampus** dan sedang
mempertimbangkan kembali Supabase sebagai penyimpanan file. Tujuan pembelajaran:

1. **Memahami apa itu object storage** dan mengapa hampir semua penyedia memakai
   satu standar API yang sama (**S3 API**).
2. **Membandingkan opsi penyimpanan** secara objektif (data control, biaya,
   kapasitas, kemudahan) sehingga bisa memutuskan sendiri, bukan ikut-ikutan.
3. **Memahami cara menukar penyimpanan** di codebase SIMTAS dengan benar —
   memanfaatkan interface `StorageService` yang sudah ada — tanpa menulis ulang
   seluruh backend.
4. **Memilih solusi yang tepat untuk kampus**: prioritas pada kendali data
   (data tiap mahasiswa), biaya jangka panjang, kapasitas PDF, dan kemudahan
   operasional/admin.

## Ukuran keberhasilan

- Mampu menjelaskan kenapa S3-compatibility membuat pilihan penyimpanan menjadi
  keputusan yang "murah" (tidak mengunci diri pada satu vendor).
- Mampu memilih & membenarkan satu rekomendasi untuk produksi kampus.
- (Praktik lanjutan) Mampu menulis implementasi `StorageService` baru (mis.
  MinIO / R2) hanya dengan mengubah satu konstruktor — bukan logic upload.
