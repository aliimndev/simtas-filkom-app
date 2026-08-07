# Release Candidate 1 — SIMTAS FILKOM

- **Status:** 🔍 Release Candidate (menunggu verifikasi produksi VPS)
- **Target branch:** `main` (disatukan dari `develop`)
- **Dokumen:** RC-1.md · Release Manager
- **Tanggal:** 2026-08-07

---

## 1. Executive Summary

RC-1 adalah kandidat rilis pertama SIMTAS FILKOM menuju produksi. Rilis ini
mengunci tiga fix **Critical** yang sudah divalidasi independen:

| ID | Temuan | Status |
|----|--------|--------|
| **C1** | Endpoint autentikasi (login/refresh/forgot/reset) gagal selalu 403 oleh global CSRF middleware | Fixed, teruji |
| **C2** | Migration title-change menabrak realita schema — GORM map `reviewed_by`/`cancelled_by` vs kolom `..._by_id` | Fixed, migration data-safe terbukti |
| **C3** | Server berjalan tanpa env yang jelas (APP_ENV kosong), CORS terbuka di produksi | Fixed, fail-fast |

Selain itu RC-1 membawa pengerasan produksi dari sesi sebelumnya: observability
(Prometheus/Grafana + sentry), notifikasi in-app dengan durable email retry
queue (dead-letter cap), skeleton UX, skrip deployment/monitoring/backup/SSL,
infra test (k6 + Playwright), dan dokumen perencanaan.

**Verdict:** **Ready for Release Candidate** — tidak ada release blocker.
Belum **Ready for Production** karena verifikasi e2e + load test terhadap stack
yang sudah ter-deploy (item H7) dan validasi hidup di VPS belum dijalankan; itu
justru pekerjaan utama fase RC.

---

## 2. Files Changed

### 2.1 Fix Critical (inti rilis)

**C1 — CSRF auth**
- `backend/internal/middleware/csrf.go` — `CSRFMiddleware(exemptPaths ...string)`, pencocokan via `c.FullPath()`
- `backend/internal/handler/router.go` — exempt 4 path auth (`/api/v1/auth/login`, `/refresh`, `/forgot-password`, `/reset-password`)
- `backend/internal/middleware/csrf_test.go` *(baru)* — kasus exempt (200) dan non-exempt (403)

**C2 — Migration title-change**
- `backend/migrations/000018_add_deleted_at_to_title_change_requests.up.sql` *(baru)* — RENAME `requested_by`→`requested_by_id`, `reviewed_by`→`reviewed_by_id`, `cancelled_by`→`cancelled_by_id`; ADD `deleted_at TIMESTAMPTZ`; INDEX `idx_title_change_requests_deleted_at`
- `backend/migrations/000018_add_deleted_at_to_title_change_requests.down.sql` *(baru)* — rollback lengkap
- `backend/internal/repository/title_change_request_repository_impl.go` — map key `reviewed_by_id`/`cancelled_by_id`
- `backend/internal/handler/title_change_request_integration_test.go` *(baru)* — regression; sengaja di package `handler` (bukan `repository`) untuk menghindari race DB paralel di test suite

**C3 — APP_ENV fail-fast**
- `backend/pkg/config/config.go` — `requireEnv("APP_ENV")` di `Load()` + whitelist `development|production` di `Validate()`
- `backend/pkg/config/config_test.go` *(baru)* — 2 kasus panic
- `backend/internal/testutil/setup.go` — set `APP_ENV=development` di setup test DB/router
- `.env.production` + `deploy/.env.production` *(keduanya untracked, bukan bagian commit)* — ditambah `APP_ENV=production`

### 2.2 Observability & error tracking
- `backend/pkg/metrics/` *(baru)* — kolektor Prometheus
- `backend/cmd/server/main.go`, `backend/internal/middleware/error_handler.go` — init sentry + report
- `backend/go.mod`, `backend/go.sum` — `prometheus/client_golang`, `getsentry/sentry-go`
- `infrastructure/monitoring/prometheus.yml`, `infrastructure/monitoring/grafana/` *(baru)*, `deploy/docker-compose.monitoring.yml` *(baru)*
- `frontend/src/lib/utils/report-error.ts` *(baru)*, `frontend/src/components/features/error-boundary.tsx`

### 2.3 Notifikasi & email
- Sudah di-commit pada `develop`: `d4c438f` (notifikasi in-app + durable email retry + race guard), `2694b7e` (bell notifikasi), `6de87e5` (test + coverage gate), `e522e9a` (swagger). Bagian dari scope RC-1.

### 2.4 Frontend & UX
- `frontend/src/components/ui/skeleton.tsx` *(baru)* + `ListSkeleton` di 12 halaman list
- `frontend/src/app/(dashboard)/...` — 12 halaman (admin/academic-years, audit-logs, users, archives, defenses, documents, schedules, seminars, supervision, theses, thesis, title-change-reviews)
- `frontend/package.json`, `frontend/package-lock.json` — bump next 16.3.0 + audit npm

### 2.5 Deployment & operasi
- `deploy/scripts/install-backup-cron.sh`, `deploy/scripts/ssl-setup.sh` *(baru)*
- `deploy/docker-compose.monitoring.yml` *(baru)*
- `.env.example` — dokumentasi var env baru

### 2.6 Infra test & dokumen
- `tests/` *(baru)* — skenario k6, `frontend/e2e/` *(baru)*, `tests/README.md`
- `docs/PRODUCTION-READINESS-REVIEW.md`, `docs/ROADMAP.md`, `docs/runbook.md`, `infrastructure/monitoring/README.md`

---

## 3. Database Migration Plan

Satu migration baru dalam rilis ini:

```
backend/migrations/000018_add_deleted_at_to_title_change_requests.{up,down}.sql
```

- **Ekseskusi:** `./simtas-api --migrate` (golang-migrate, file-source) — berjalan berurutan setelah 000001..000017.
- **Isi up:** rename 3 kolom (`requested_by`→`requested_by_id`, dst.) + `deleted_at` + index.
- **Sifat:** **data-preserving** (ALTER RENAME) — terbukti di DB scratch: baris APPROVED utuh melewati up→down→up.
- **Keamanan waktu henti:** sebelum fix, fitur title-change *sudah rusak* (GORM query `requested_by_id` vs kolom `requested_by`). Tidak ada jendela "fitur pernah berfungsi lalu rusak" — karenanya memilih migration perbaikan atas restore DB.
- **Uji regression:** integration test gagal tanpa migration (`column "requested_by_id" does not exist`), hijau setelah migration dijalankan.

### Rollback migration
`--migrate` + `--migrate-down` (atau jalankan `000018.down.sql`). Kolom kembali ke nama lama; `deleted_at` & index dihapus. Backward compatible dengan binary lama (yang memang tak pernah bisa memakai fitur ini).

---

## 4. Deployment Procedure

Mengikuti `deploy/README.md` (contoh dengan `deploy/docker-compose.prod.yml`, `env_file: .env.production`):

1. **Persiapan** — tarik kode, salin `.env.production` dari `deploy/.env.production.example`, isi seluruh `CHANGE_ME` (termasuk `APP_ENV=production`, `CORS_ALLOWED_ORIGINS=https://<domain>`).
2. **Backup** — jalankan skrip backup manual/harian (cron 02:00 via `install-backup-cron.sh`) sebelum up.
3. **Build & start** — `docker compose -f deploy/docker-compose.prod.yml up -d --build` (postgres → pgbouncer → backend → frontend → nginx → minio).
4. **Migrasi** — `docker compose -f deploy/docker-compose.prod.yml run --rm backend ./simtas-api --migrate` (menjalankan 000018).
5. **Seed** — hanya saat deploy pertama: `... backend ./simtas-seed`.
6. **Verifikasi health** — `curl https://<domain>/health`.
7. **Observability** — `docker compose -f deploy/docker-compose.monitoring.yml up -d` (prometheus/grafana).

---

## 5. Rollback Procedure

1. **Code** — kembalikan image/binary ke build sebelumnya (git revert merge C1–C3).
2. **DB** — jalankan down migration 000018. Tanpa itu pun aman: binary lama tidak membaca kolom baru; schema baru hanya berisi rename+kolom tambahan, tidak merusak path lain.
3. **Konfigurasi** — binary lama tidak membaca `APP_ENV`; tidak ada langkah tambahan.
4. **Verifikasi** — ulangi checklist §6.

---

## 6. Post-Deployment Verification Checklist

| Area | Cara verifikasi | Kriteria lolos |
|------|-----------------|----------------|
| Auth | Login/logout/refresh via browser + curl | Tidak ada 403 CSRF; refresh memperpanjang sesi |
| Authorization | Login sebagai admin, kaprodi, dosen, mahasiswa | RBAC menolak/tidak menolak sesuai peran |
| CSRF | Browser fresh (tanpa cookie) submit POST login | 200; POST non-auth tanpa token → 403 |
| Title Change | Submit → list → approve/reject/cancel | Alur lengkap, `requested_by_id` terisi |
| Dashboard | Buka halaman dashboard | Metrik/grafik tampil, tanpa error |
| File Upload | Upload dokumen (jpeg/png/pdf) | Tersimpan; tipe lain ditolak |
| Notifications | Trigger notifikasi, buka bell | Badge muncul, mark-as-read berfungsi |
| Email | Event memicu email (signup/reset) | Email terkirim; `email_logs` tidak menumpuk `failed` |
| Database | `--migrate` log + `\dt` / schema | Version = 18; kolom baru ada |
| Monitoring | Grafana target up, sentry dashboard | Target hijau; error terekam |
| Logs | `docker compose logs backend` | JSON terstruktur, tanpa stacktrace berlebihan |

---

## 7. Remaining Risks (H1–H7)

Status setelah audit (penilaian konsolidasi). **Tidak ada yang menjadi release blocker.**

| ID | Item | Dampak | Mitigasi yang ada |
|----|------|--------|-------------------|
| **H3** | Upload file: nama file asli dipakai langsung tanpa re-encode; MIME hanya dicek via ekstensi | Menengah — file berbahaya bisa di-upload tersamar | Validasi tipe via ekstensi; tindakan di v1.0.1 |
| **H2** | Backup MinIO via `mc mirror`; jika storage placeholder (bukan S3) dokumen tak masuk backup | Menengah — restore dokumen bisa inkonsisten | Backup postgres solid; verifikasi konfigurasi S3 saat deploy |
| **H1** | Logout tidak invalidasi JWT server-side (stateless) | Menengah — token berlaku sampai expiry | JWT expiry pendek (24h) + `token_version` dibump saat ganti password |
| **H7** | k6 & Playwright belum terinstall di VPS; e2e/load test belum pernah jalan | Menengah — gap verifikasi release | Infra test tersedia; jadi tugas utama fase RC |
| **H5** | Hash password memakai `bcrypt(cost 12)`; belum argon2id | Rendah — modernisasi, bukan vuln aktif | bcrypt cost 12 memadai; migrasi hash di v1.0.1 |
| **H6** | Alur seminar/defense: ada `ErrSeminarActiveExists`, tetapi belum ada guard jadwal duplikat (slot sama) | Rendah | Validasi manual; guard di v1.0.1 |
| **H4** | ~~Retry email tanpa batas~~ | **RESOLVED** — tidak lagi menjadi risiko | Dead-letter cap `emailMaxSchedulerRuns=5` + `emailMaxRetries=3` di rilis ini |

---

## 8. Deferred Work (v1.0.1)

- **H1–H3, H5, H6** — lihat §7 (prioritas di §9).
- **Cleanup over-engineering** (hasil audit): hapus dead code (`pagination.go`, `dto/request.go`, `RollbackAll`, var `apperror` tak terpakai, `GetFileSizeInMB`, `grading` tak terpakai, statemachine mati, helper CSRF mati, `newIPRateLimiter`, `backup.sh` duplikat terorak, `assets/style.css`, SVG frontend); shrink (`gin.Default()`→`gin.New()`, `date.ts`→`Intl`, `next.config` remotePatterns, `client.ts` `readCookie`→axios `xsrfConfig`); hati-hati dengan `seeds/003_admin_user.sql`.
- **Verifikasi** — k6/Playwright di VPS (§10).

---

## 9. Recommended Priority (H1–H7)

Urutan berdampak × kemungkinan terjadi (tinggi → rendah):

1. **H3** — keamanan upload (vektor langsung, user-facing) → tertinggi.
2. **H2** — kelengkapan backup/restore dokumen.
3. **H1** — invalidasi sesi saat logout.
4. **H7** — tutup gap verifikasi (e2e + load test).
5. **H5** — modernisasi hash password (argon2id).
6. **H6** — guard jadwal seminar/defense duplikat.
7. **H4** — ✅ sudah ter-resolve di rilis ini; tidak ada aksi.

---

## 10. Final Release Assessment

**Status: ✅ READY FOR RELEASE CANDIDATE**

**Lulus menuju RC:**
- Ketiga fix Critical C1–C3 selesai, diverifikasi independen, semua test (unit + integration) hijau, code style bersih (gofmt/build/vet), review disetujui.
- Migration 000018 terbukti data-safe; tidak ada jendela rusak.
- Tidak ada release blocker tersisa; sisa risiko semuanya Medium/Low dengan mitigasi berjalan.

**Syarat menjadi Ready for Production (kerjaan fase RC):**
1. Deploy penuh ke VPS mengikuti §4, jalankan checklist §6.
2. Install & jalankan k6 + Playwright (H7) terhadap stack ter-deploy.
3. Verifikasi nyata backup (restore drill) dan konfigurasi S3 (H2).

Setelah ketiga syarat di atas hijau, lakukan **RC-2 → Promotion Review** untuk keputusan produksi.
