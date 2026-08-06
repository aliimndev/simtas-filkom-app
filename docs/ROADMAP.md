# 🗺️ ROADMAP RELEASE — SIMTAS FILKOM v1.0

**Tanggal:** 6 Agustus 2026
**Status:** 90% selesai — menuju go-live untuk ratusan mahasiswa & dosen
**Sumber detail audit:** [PRODUCTION-READINESS-REVIEW.md](./PRODUCTION-READINESS-REVIEW.md)

---

## Ringkasan

Seluruh modul inti aplikasi **sudah selesai**:

- ✅ Backend: Clean Architecture (Go + Gin), 12 domain usecase, state machine thesis, audit log, RBAC, middleware security lengkap
- ✅ Frontend: Next.js 16, semua halaman untuk 5 role, ErrorBoundary, RHF+Zod di sebagian form
- ✅ Database: 15 migrasi + GIN index full-text search + seed data
- ✅ CI/CD: GitHub Actions (lint, test, coverage ≥80%, swagger check) + CD ke VPS/Vercel
- ✅ Docs operasional: `deploy/README.md`, `docs/runbook.md`
- ✅ Security: CSRF, CSP/HSTS, JWT secret validation, global rate limit, sanitasi input, refresh token rotation, httpOnly cookie

Yang tersisa **bukan fitur besar**, melainkan **3 fitur belum ada + hardening operasional + testing + UAT/go-live**.

---

## 🎯 Fase 1 — Fitur Yang Belum Ada (5–7 hari)

Item yang belum ada di codebase dan perlu dibuat sebelum rilis:

| # | Fitur | Detail | Estimasi |
|---|-------|--------|----------|
| 1 | **In-App Notification** | Bell icon + badge count di dashboard. Backend: tabel `notifications`, API list/read/unread. Email sudah jalan, ini pelengkap. | 3–4 hari |
| 2 | **Email Retry Queue** | `email_logs` sudah ada tapi tanpa retry. Tambah buffered channel + worker pool + retry (3x) + dead letter untuk status gagal. | 2 hari |
| 3 | **Skeleton Loading States** | Skeleton component (`ui/skeleton.tsx`) sudah ada tapi baru dipakai di dashboard. Terapkan di semua halaman list (theses, archives, seminars, defenses, documents, schedules, admin users). | 1 hari |

> Nota: **Form validation RHF+Zod** sudah dipakai di auth, profile, thesis, supervision — tinggal rapikan sisanya (opsional, bukan blocker).

---

## 🔒 Fase 2 — Hardening Operasional (2–3 hari)

Pekerjaan di VPS/infrastruktur, bukan kode aplikasi:

| # | Item | Status | Detail |
|---|------|--------|--------|
| 1 | **Automated DB Backup** | Script ada (`deploy/scripts/backup.sh`), **cron belum aktif** | Pasang crontab daily 02.00 + upload off-site + test restore |
| 2 | **SSL/TLS (Let's Encrypt)** | Hanya dokumentasi | Setup certbot di Nginx untuk `api.simtas.filkom.unida.ac.id` & `simtas.filkom.unida.ac.id` |
| 3 | **N+1 Query Audit** | Belum diaudit | Review `Preload` di repository list endpoints; ganti yang N+1 dengan JOIN/batch |
| 4 | **Sentry / Error Tracking** | Belum ada | Install Sentry di backend + frontend; atur alert ke email admin |
| 5 | **Monitoring Dashboard** | `infrastructure/monitoring/` masih placeholder | Prometheus + Grafana (scrape `/metrics` backend) |

---

## 🧪 Fase 3 — Testing Lengkap (3–4 hari)

| # | Item | Detail |
|---|------|--------|
| 1 | **API Test Collection** | Postman/Insomnia collection atau k6 script untuk semua endpoint utama (auth, RBAC matrix, pagination, error shape) |
| 2 | **Load Test** | k6: 50/100/200 VU baseline, target p95 < 500ms–1s; endpoit kunci: theses list, dashboard summary, upload, archive search |
| 3 | **Security Test** | OWASP ZAP scan + manual: SQLi, XSS, CSRF, auth bypass, file upload malicious |
| 4 | **Full Lifecycle E2E** | Playwright: login → submit → approve → assign → upload → review → seminar → sidang → graduate (backend integration test sudah ada, frontend E2E belum) |

---

## 👥 Fase 4 — UAT & Persiapan Go-Live (1–2 minggu)

| # | Item | Detail |
|---|------|--------|
| 1 | **UAT dengan user nyata** | 2–3 mahasiswa, 2–3 dosen, 1 kaprodi, 1 admin. Pakai skenario di `docs/pertanyaan-kaprodi-flow.md` + daftar UAT |
| 2 | **Training Admin & Kaprodi** | Sesi singkat: buat akun, import user, review judul, assign pembimbing, jadwal seminar/sidang |
| 3 | **Smoke Test Production** | Per role: login → action utama → verifikasi email terkirim |
| 4 | **Seed Data Production** | Roles, tahun akademik, admin & kaprodi dengan password acak + `must_change_password` |
| 5 | **Pengumuman Go-Live** | Email ke civitas akademika: URL, cara dapat akun, kontak support, tanggal wajib pakai |

---

## 🚀 Fase 5 — Go-Live (hari H)

Urutan:

```
08.00  Verifikasi checklist go-live (infra, security, data, monitoring)
08.30  Smoke test production final
09.00  Kirim pengumuman resmi go-live
09.15  Admin import data mahasiswa angkatan berjalan
10.00  Sistem resmi OPEN — pantau log + uptime 2 jam pertama
```

---

## 📋 Checkpoint Harian

| Item | Checklist |
|------|-----------|
| Backend | `cd backend && go build ./... && go vet ./... && go test ./... -race` |
| Backend integrasi | `cd backend && make test-integration` |
| Coverage | `cd backend && make coverage-check` (≥80% usecase) |
| Frontend | `cd frontend && npm run lint && npm run type-check && npm run build` |
| CI | Push → GitHub Actions hijau di `develop` lalu `main` |
| Swagger | `cd backend && make docs` — tidak boleh ada diff |

---

## Post-Launch (minggu pertama)

- [ ] Pantau error log & Sentry setiap hari
- [ ] Patch v1.0.1 dalam 2 minggu jika ada bug signifikan
- [ ] Kumpulkan feedback user awal → buat GitHub Issues

---

## Catatan Versi

| Versi | Target | Keterangan |
|-------|--------|------------|
| v1.0.0 | Go-live | Sesuai roadmap ini |
| v1.0.1 | +2 minggu pasca-live | Bug fixes |
| v1.1.0 | TBD | Export PDF/Excel, document preview, dark mode |
