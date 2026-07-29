# Job 30 — Go-Live Checklist

**Phase:** 7 — Launch
**Referensi PRD:** Section 23 (Milestones / Roadmap v1.0)
**Prerequisites:** Job 29 (Seed Data Production) ✅
**Estimasi:** 1 hari

---

## Objective

Checklist final sebelum sistem dibuka untuk pengguna nyata. Setiap item harus diverifikasi dan dicentang. Setelah semua item ✅, lakukan pengumuman go-live resmi kepada civitas akademika FILKOM Unida.

---

## Go-Live Checklist

### ✅ 1. Infrastructure

- [ ] Backend production berjalan: `https://api.simtas.filkom.unida.ac.id/api/v1/health` → `{ "status": "ok" }`
- [ ] Frontend production berjalan: `https://simtas.filkom.unida.ac.id` → halaman login tampil
- [ ] SSL certificate valid untuk kedua domain (tidak ada warning browser)
- [ ] Database PostgreSQL berjalan dan sehat di dalam Docker
- [ ] Disk space VPS: tersisa minimal 50% (untuk pertumbuhan data)
- [ ] Memory usage VPS: tidak lebih dari 70% saat idle
- [ ] Docker containers semua `healthy` (bukan `starting` atau `unhealthy`)
- [ ] Auto-restart aktif: matikan container → otomatis restart dalam <30 detik

### ✅ 2. Security

- [ ] Tidak ada secret/password yang ter-commit di repository GitHub
- [ ] `.env.production` di VPS tidak bisa dibaca oleh user selain deployer dan root
- [ ] JWT secret key: panjang minimal 256-bit (32 karakter random)
- [ ] Database password: kuat (minimal 20 karakter random)
- [ ] Swagger UI **tidak** bisa diakses di production (`APP_ENV=production`)
- [ ] CORS hanya mengizinkan origin `https://simtas.filkom.unida.ac.id`
- [ ] Rate limiting aktif pada endpoint login
- [ ] Semua akun default sudah ganti password (`must_change_password = false` untuk akun yang sudah login)
- [ ] Firewall VPS: hanya port 22, 80, 443 yang terbuka

### ✅ 3. Data & Database

- [ ] Migrasi database production sudah berjalan (semua tabel ada)
- [ ] Seed data production sudah ada (roles, tahun akademik, admin, kaprodi, dosen)
- [ ] Tidak ada data test/dummy di database production
- [ ] Backup pertama sudah berjalan: file `.sql.gz` ada di folder backup
- [ ] Crontab backup harian sudah aktif (`crontab -l` → terlihat job backup)
- [ ] Test restore backup: ambil file backup, restore ke DB temporary, verifikasi data

### ✅ 4. Monitoring & Alerting

- [ ] UptimeRobot monitor aktif untuk API endpoint
- [ ] UptimeRobot monitor aktif untuk frontend URL
- [ ] Alert email dikonfigurasi ke alamat Admin Fakultas
- [ ] Test alert: matikan backend sementara → konfirmasi email alert diterima
- [ ] Log rotation aktif (Docker: max-size 50m, max-file 5)
- [ ] Runbook tersedia di `docs/runbook.md` dan dapat diakses tim

### ✅ 5. CI/CD

- [ ] GitHub Actions CI berjalan hijau pada commit terakhir di `main`
- [ ] GitHub Actions CD deploy berhasil (tidak ada failed job)
- [ ] Branch protection aktif di `main`: PR wajib pass CI sebelum merge
- [ ] Test rollback: deploy image SHA sebelumnya berhasil

### ✅ 6. Fungsionalitas (Smoke Test Production)

Jalankan smoke test cepat di production environment:

- [ ] Login sebagai Admin → berhasil
- [ ] Login sebagai Kaprodi → berhasil
- [ ] Buat 1 user mahasiswa test dari Admin panel → welcome email terkirim ke inbox
- [ ] Login sebagai mahasiswa test → berhasil, dashboard tampil
- [ ] Submit pengajuan judul test → Kaprodi menerima email notifikasi
- [ ] Kaprodi approve judul + tetapkan pembimbing → mahasiswa menerima email
- [ ] Upload dokumen PDF kecil (<1MB) → berhasil tersimpan, bisa didownload
- [ ] Cek arsip search berfungsi (mungkin kosong dulu, tapi tidak error)
- [ ] **Hapus data smoke test setelah selesai**

### ✅ 7. Documentation

- [ ] `docs/00-overview.md` sudah diupdate dengan tanggal go-live aktual
- [ ] `docs/runbook.md` tersedia dan akurat
- [ ] Swagger API docs dapat diakses di staging/development
- [ ] README.md di root repo berisi instruksi setup untuk developer baru
- [ ] Semua 30 job docs sudah dibuat dan ada di `docs/`

### ✅ 8. Team Readiness

- [ ] Admin Fakultas sudah training singkat: cara buat akun, import user, jadwalkan seminar/sidang
- [ ] Kaprodi sudah training: cara review pengajuan, tetapkan pembimbing, yudisium
- [ ] Panduan singkat user tersedia (bisa berupa dokumen Word/PDF 1–2 halaman per role)
- [ ] Nomor kontak support (developer atau IT FILKOM) sudah dibagikan ke pengguna
- [ ] Prosedur laporan bug sudah dikomunikasikan ke pengguna

### ✅ 9. Komunikasi Go-Live

- [ ] Draft pengumuman go-live sudah siap (email ke civitas akademika atau grup WhatsApp)
- [ ] Isi pengumuman:
  - URL sistem: `https://simtas.filkom.unida.ac.id`
  - Cara mendapatkan akun: hubungi Admin Fakultas
  - Kontak support
  - Tanggal mulai berlaku wajib menggunakan sistem
- [ ] Pengumuman disetujui oleh Kaprodi atau Dekan sebelum disebarkan

---

## Prosedur Go-Live Hari H

Urutan tindakan pada hari go-live:

```
08.00  Verifikasi semua item checklist di atas ✅
08.30  Jalankan smoke test production
09.00  Kirim pengumuman resmi go-live ke civitas akademika
09.15  Admin Fakultas mulai import data mahasiswa angkatan berjalan
10.00  Sistem resmi OPEN untuk pengguna
       Pantau log dan UptimeRobot selama 2 jam pertama
```

---

## Post-Launch — Minggu Pertama

- [ ] Pantau error log setiap hari: `docker compose logs --tail=200 backend | grep -i error`
- [ ] Pantau UptimeRobot dashboard setiap hari
- [ ] Kumpulkan feedback dari pengguna awal (Admin, Kaprodi, Mahasiswa pertama)
- [ ] Buat GitHub Issues untuk setiap bug yang ditemukan
- [ ] Target: patch v1.0.1 dalam 2 minggu setelah go-live jika ada bug signifikan

---

## Done Criteria — SIMTAS FILKOM v1.0 LIVE ✅

- [ ] **Semua 30 item checklist dicentang**
- [ ] Pengumuman go-live sudah terkirim
- [ ] Sistem dapat diakses di `https://simtas.filkom.unida.ac.id`
- [ ] Tidak ada incident dalam 2 jam pertama setelah go-live
- [ ] **🎉 SIMTAS FILKOM v1.0 resmi diluncurkan!**

---

## Catatan Versi

| Versi | Tanggal | Keterangan |
|-------|---------|------------|
| v1.0.0 | 2027-XX-XX | Initial release — go-live |
| v1.0.1 | TBD | Bug fixes pasca go-live |
| v1.1.0 | TBD | Export PDF/Excel, Dashboard analitik lanjutan |
| v2.0.0 | TBD | Integrasi SIAKAD, WhatsApp notif, E-signature |
