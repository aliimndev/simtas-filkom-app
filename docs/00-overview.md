# SIMTAS FILKOM v1.0 — Project Overview & Timeline

**Sistem Manajemen Tugas Akhir dan Skripsi**
**Fakultas Ilmu Komputer Universitas Djuanda**
**Target Release: 2027**

---

## Cara Menggunakan Dokumen Ini

Setiap file di folder `docs/` merepresentasikan satu **job/milestone** yang harus diselesaikan secara **sequential** (berurutan). Satu job harus selesai dan lulus kriteria **Done** sebelum melanjutkan ke job berikutnya.

**Format setiap dokumen job:**
- **Objective** — tujuan yang harus dicapai
- **Referensi PRD** — section PRD yang relevan
- **Prerequisites** — job sebelumnya yang harus selesai
- **Checklist** — daftar tugas teknis yang harus dikerjakan
- **Done Criteria** — definisi "selesai" yang terukur

---

## Ringkasan Phase & Timeline

| Phase | Fokus | Jobs | Estimasi Durasi |
|-------|-------|------|-----------------|
| **Phase 1** | Foundation | 01–03 | 2 minggu |
| **Phase 2** | Core Backend | 04–09 | 6 minggu |
| **Phase 3** | Supporting Features | 10–13 | 3 minggu |
| **Phase 4** | Frontend | 14–20 | 6 minggu |
| **Phase 5** | Integration & Testing | 21–23 | 2 minggu |
| **Phase 6** | Deployment | 24–27 | 1 minggu |
| **Phase 7** | Launch | 28–30 | 1 minggu |
| **Total** | | 30 jobs | **~21 minggu** |

---

## Alur Sequential (Dependency Map)

```
[01] Project Setup
  └─► [02] Database Schema
        └─► [03] Auth & RBAC
              └─► [04] User Management
                    └─► [05] Thesis Submission
                          └─► [06] Supervision
                                └─► [07] Document Management
                                      └─► [08] Seminar Module
                                            └─► [09] Defense Module
                                                  └─► [10] Archive Module
                                                        └─► [11] Email Notification
                                                              └─► [12] Dashboard
                                                                    └─► [13] Audit Log
                                                                          └─► [14] Frontend Setup
                                                                                └─► [15] Frontend Auth Pages
                                                                                      └─► [16] Frontend Admin Pages
                                                                                            └─► [17] Frontend Mahasiswa
                                                                                                  └─► [18] Frontend Dosen
                                                                                                        └─► [19] Frontend Kaprodi
                                                                                                              └─► [20] Frontend Dashboard
                                                                                                                    └─► [21] Storage Integration
                                                                                                                          └─► [22] API Documentation
                                                                                                                                └─► [23] Testing Strategy
                                                                                                                                      └─► [24] Deployment Backend
                                                                                                                                            └─► [25] Deployment Frontend
                                                                                                                                                  └─► [26] CI/CD Pipeline
                                                                                                                                                        └─► [27] Monitoring & Logging
                                                                                                                                                              └─► [28] UAT Checklist
                                                                                                                                                                    └─► [29] Seed Data
                                                                                                                                                                          └─► [30] Go-Live Checklist
```

---

## Tech Stack Referensi

| Layer | Teknologi |
|-------|-----------|
| Frontend | Next.js 16, React 19, TypeScript, Tailwind CSS, TanStack Query, React Hook Form, Zod |
| Backend | Go 1.24+, Gin Framework, GORM, JWT |
| Database | PostgreSQL 16 |
| Storage | Supabase Storage |
| Email | Resend |
| Deploy FE | Vercel |
| Deploy BE | VPS Ubuntu + Nginx + Docker + GitHub Actions |

---

## Status Tracking

Gunakan tabel ini untuk melacak progress harian:

| Job | Nama | Status | Tanggal Selesai |
|-----|------|--------|-----------------|
| 01 | Project Setup | ✅ Selesai | — |
| 02 | Database Schema | ✅ Selesai | — |
| 03 | Auth & RBAC | ✅ Selesai | — |
| 04 | User Management | ✅ Selesai | — |
| 05 | Thesis Submission | ✅ Selesai | — |
| 06 | Supervision | ⬜ Belum | — |
| 07 | Document Management | ⬜ Belum | — |
| 08 | Seminar Module | ⬜ Belum | — |
| 09 | Defense Module | ⬜ Belum | — |
| 10 | Archive Module | ⬜ Belum | — |
| 11 | Email Notification | ⬜ Belum | — |
| 12 | Dashboard | ⬜ Belum | — |
| 13 | Audit Log | ⬜ Belum | — |
| 14 | Frontend Setup | ⬜ Belum | — |
| 15 | Frontend Auth Pages | ⬜ Belum | — |
| 16 | Frontend Admin Pages | ⬜ Belum | — |
| 17 | Frontend Mahasiswa | ⬜ Belum | — |
| 18 | Frontend Dosen | ⬜ Belum | — |
| 19 | Frontend Kaprodi | ⬜ Belum | — |
| 20 | Frontend Dashboard | ⬜ Belum | — |
| 21 | Storage Integration | ⬜ Belum | — |
| 22 | API Documentation | ⬜ Belum | — |
| 23 | Testing Strategy | ⬜ Belum | — |
| 24 | Deployment Backend | ⬜ Belum | — |
| 25 | Deployment Frontend | ⬜ Belum | — |
| 26 | CI/CD Pipeline | ⬜ Belum | — |
| 27 | Monitoring & Logging | ⬜ Belum | — |
| 28 | UAT Checklist | ⬜ Belum | — |
| 29 | Seed Data | ⬜ Belum | — |
| 30 | Go-Live Checklist | ⬜ Belum | — |

**Legend:** ⬜ Belum &nbsp;|&nbsp; 🔄 In Progress &nbsp;|&nbsp; ✅ Selesai

---

## Konvensi Commit

```
feat(job-XX): deskripsi singkat
fix(job-XX): deskripsi bug fix
docs(job-XX): update dokumentasi
test(job-XX): tambah/update test
chore: update dependency / config
```

---

## Referensi

- [PRD.md](../PRD.md) — Product Requirements Document lengkap
- [LICENSE](../LICENSE) — Lisensi proyek
