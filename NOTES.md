# NOTES — Preferensi & Catatan Kerja

## Preferensi pengguna
- Bahasa: **Indonesia** (menjawab/menulis materi dalam Bahasa Indonesia,
  kode/istilah teknis tetap bahasa Inggris).
- Konteks selalu dikaitkan ke **project SIMTAS** yang nyata, bukan materi abstrak.
- Pertimbangan utama: **deployment produksi untuk kampus** (kendali data, biaya,
  kapasitas). Bukan sekadar "cara termurah".

## Konteks teknis project (diingat utk pelajaran berikutnya)
- Backend: Go, gin, PostgreSQL, GORM.
- Interface `StorageService` (`backend/internal/domain/service/storage_service.go`):
  `Upload`, `GeneratePresignedURL`, `Delete`.
- Pemilihan impementasi di `router.go`: `STORAGE_PROVIDER` → `local` (stub) atau
  `supabase`.
- Stub lokal menulis ke `./tmp/uploads` dan serve via prefix URL.
- Path convention: `documents/{thesis_id}/{doc_type}/v{n}_{filename}.pdf`,
  `archives/{year}/{thesis_id}/{filename}.pdf`, dst.

## Rencana sesi lanjutan (ZPD)
- 0001: membandingkan opsi storage + konsep S3 API (selesai).
- 0002: menulis implementasi `StorageService` MinIO/R2 (skill praktik).
- 0003: presigned URL & keamanan download (mendalam).
- 0004: biaya/egress & keputusan final untuk kampus.
