# SIMTAS FILKOM v1.0 — Project Overview

**Sistem Manajemen Tugas Akhir dan Skripsi**
**Fakultas Ilmu Komputer Universitas Djuanda**

---

## Ringkasan

SIMTAS FILKOM adalah sistem manajemen tugas akhir/skripsi berbasis web untuk
Fakultas Ilmu Komputer Universitas Djuanda. Aplikasi menggunakan Hono + Bun di
API dan SvelteKit di frontend, dengan 5 role user, state machine untuk alur
thesis, audit log, dan notifikasi email.

Dokumen berikut merepresentasikan cara kerja dan pekerjaan yang tersisa menuju release:

- [ENGINEERING-WORKFLOW.md](./ENGINEERING-WORKFLOW.md) — workflow engineering A–Z untuk membangun, menguji, merilis, dan mengoperasikan SIMTAS
- [AGENT-EXECUTION-LOOP.md](./AGENT-EXECUTION-LOOP.md) — protocol agent untuk frontier ticket, self-correction, verification, dan safety stop
- [ROADMAP.md](./ROADMAP.md) — rencana kerja menuju go-live (berlaku saat ini)
- [VERCEL.md](./VERCEL.md) — konfigurasi deployment frontend dan pemisahan API
- [superpowers/plans/2026-07-28-ts-rewrite-cutover-governance.md](./superpowers/plans/2026-07-28-ts-rewrite-cutover-governance.md) — governance cutover dan parity gate
- [superpowers/plans/2026-07-28-ts-rewrite-phase1-walking-skeleton.md](./superpowers/plans/2026-07-28-ts-rewrite-phase1-walking-skeleton.md) — rencana walking skeleton Phase 1

---

## Tech Stack

| Layer | Teknologi |
|-------|-----------|
| Frontend | SvelteKit, Svelte, TypeScript, Tailwind CSS |
| Backend | Hono, Bun, TypeScript, JWT |
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

- [ENGINEERING-WORKFLOW.md](./ENGINEERING-WORKFLOW.md) — workflow engineering A–Z
- [ROADMAP.md](./ROADMAP.md) — roadmap menuju go-live
- [VERCEL.md](./VERCEL.md) — deployment frontend dan pemisahan API
