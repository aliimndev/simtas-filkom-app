# SIMTAS FILKOM v1.0 — Project Overview

**Sistem Manajemen Tugas Akhir dan Skripsi**
**Fakultas Ilmu Komputer Universitas Djuanda**

---

## Ringkasan

SIMTAS FILKOM adalah sistem manajemen tugas akhir/skripsi berbasis web untuk
Fakultas Ilmu Komputer Universitas Djuanda. Aplikasi sudah dibangun dengan
Clean Architecture (Go + Gin + PostgreSQL) di backend dan Next.js di frontend,
dengan 5 role user, state machine untuk alur thesis, audit log, dan notifikasi email.

Dokumen berikut merepresentasikan pekerjaan yang tersisa menuju release:

- [ROADMAP.md](./ROADMAP.md) — rencana kerja menuju go-live (berlaku saat ini)
- [PRODUCTION-READINESS-REVIEW.md](./PRODUCTION-READINESS-REVIEW.md) — audit kelayakan production
- [runbook.md](./runbook.md) — prosedur operasional server production
- [pertanyaan-kaprodi-flow.md](./pertanyaan-kaprodi-flow.md) — validasi alur dengan Kaprodi
- [PRD.md](../PRD.md) — product requirements document

---

## Tech Stack

| Layer | Teknologi |
|-------|-----------|
| Frontend | Next.js 16, React 19, TypeScript, Tailwind CSS, TanStack Query, React Hook Form, Zod |
| Backend | Go 1.24+, Gin Framework, GORM, JWT |
| Database | PostgreSQL 16 |
| Storage | S3-compatible object storage (Supabase Storage / swappable via `StorageService`) |
| Email | Resend |
| Deploy FE | Vercel |
| Deploy BE | VPS Ubuntu + Nginx + Docker + GitHub Actions |

---

## Konvensi Commit

```
feat: deskripsi singkat
fix: deskripsi bug fix
docs: update dokumentasi
test: tambah/update test
chore: update dependency / config
```

---

## Referensi

- [PRD.md](../PRD.md) — Product Requirements Document lengkap
