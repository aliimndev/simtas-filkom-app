# Job 27 — Monitoring & Logging (Production)

**Phase:** 6 — Deployment
**Referensi PRD:** Section 7 (Non-Functional Requirements — Logging & Error Handling)
**Prerequisites:** Job 26 (CI/CD Pipeline) ✅
**Estimasi:** 1 hari

---

## Objective

Setup monitoring dan logging production: akses log aplikasi via Docker, uptime monitoring eksternal, dan alert sederhana jika sistem down. Setelah job ini selesai, tim bisa mendeteksi masalah production secara proaktif.

---

## Checklist

### Application Logging — Docker

Backend sudah menggunakan `slog` (dari Job 13). Di production, log ditulis ke stdout dalam format JSON dan ditangkap Docker.

- [ ] Verifikasi log format production di `cmd/server/main.go`:
  ```go
  if cfg.Env == "production" {
    logger.Init("production") // JSON format ke stdout
  }
  ```
- [ ] Perintah untuk melihat log real-time di VPS:
  ```bash
  # Log backend real-time
  docker compose -f docker-compose.prod.yml logs -f backend

  # Log 100 baris terakhir
  docker compose -f docker-compose.prod.yml logs --tail=100 backend

  # Log dengan timestamp
  docker compose -f docker-compose.prod.yml logs -t backend

  # Log PostgreSQL
  docker compose -f docker-compose.prod.yml logs -f postgres
  ```
- [ ] Konfigurasi Docker log rotation (agar disk tidak penuh):
  ```yaml
  # Tambah ke docker-compose.prod.yml pada service backend & postgres
  logging:
    driver: "json-file"
    options:
      max-size: "50m"
      max-file: "5"
  ```

### Nginx Access Log

- [ ] Verifikasi Nginx access log aktif di `/var/log/nginx/access.log`
- [ ] Nginx error log di `/var/log/nginx/error.log`
- [ ] Konfigurasi logrotate untuk Nginx (sudah otomatis di Ubuntu, verifikasi saja):
  ```bash
  cat /etc/logrotate.d/nginx
  ```
- [ ] Perintah useful untuk analisis log Nginx:
  ```bash
  # 10 URL yang paling banyak di-request
  awk '{print $7}' /var/log/nginx/access.log | sort | uniq -c | sort -rn | head -10

  # Request dengan status 5xx
  grep ' 5[0-9][0-9] ' /var/log/nginx/access.log | tail -50

  # Request dari IP tertentu
  grep '192.168.1.1' /var/log/nginx/access.log
  ```

### Uptime Monitoring (Eksternal)

- [ ] Daftar ke **UptimeRobot** (gratis untuk 50 monitor, check interval 5 menit):
  - Buka https://uptimerobot.com → daftar akun gratis
  - Add monitor: HTTP(s), URL: `https://api.simtas.filkom.unida.ac.id/api/v1/health`
  - Alert contact: email Admin Fakultas
  - Add monitor kedua: `https://simtas.filkom.unida.ac.id` (frontend)
- [ ] Konfigurasi alert: email notifikasi saat down + saat kembali up
- [ ] Status page UptimeRobot bisa di-share ke stakeholder (opsional)

### Health Check Endpoint Enhancement

- [ ] Update `GET /api/v1/health` agar lebih informatif:
  ```go
  // Response production health check
  {
    "status": "ok",
    "version": "1.0.0",
    "timestamp": "2027-01-15T10:00:00Z",
    "uptime_seconds": 86400,
    "database": "ok"   // ping database, return "error" jika gagal
  }
  ```
- [ ] Jika database ping gagal → status 503 (UptimeRobot akan alert)

### Error Alerting Sederhana

- [ ] Buat script `scripts/check-errors.sh` yang berjalan via cron setiap jam:
  ```bash
  #!/bin/bash
  # Cek apakah ada error 5xx dalam 1 jam terakhir
  ERROR_COUNT=$(grep ' 5[0-9][0-9] ' /var/log/nginx/access.log | \
    awk -v d="$(date -d '1 hour ago' '+%d/%b/%Y:%H')" '$4 > "["d' | wc -l)

  if [ "$ERROR_COUNT" -gt 10 ]; then
    echo "ALERT: $ERROR_COUNT error 5xx dalam 1 jam terakhir di SIMTAS FILKOM" | \
      mail -s "SIMTAS Alert: Server Errors" admin@filkom.unida.ac.id
  fi
  ```
- [ ] Pasang di crontab: `0 * * * * /opt/simtas-filkom/scripts/check-errors.sh`
- [ ] Install `mailutils` jika belum ada: `apt install mailutils -y`

### Disk Space Monitoring

- [ ] Tambah ke crontab — alert jika disk usage > 80%:
  ```bash
  # Cek disk setiap hari jam 8 pagi
  0 8 * * * df -h / | awk 'NR==2 {if(int($5)>80) print "ALERT: Disk usage "$5" on SIMTAS server"}' | \
    mail -s "SIMTAS Disk Alert" admin@filkom.unida.ac.id
  ```

### Runbook — Prosedur Operasional

- [ ] Buat file `docs/runbook.md` dengan prosedur:

  **Restart backend:**
  ```bash
  cd /opt/simtas-filkom
  docker compose -f docker-compose.prod.yml restart backend
  ```

  **Restart semua service:**
  ```bash
  docker compose -f docker-compose.prod.yml down
  docker compose -f docker-compose.prod.yml up -d
  ```

  **Manual backup database:**
  ```bash
  /opt/simtas-filkom/backups/backup.sh
  ```

  **Restore database dari backup:**
  ```bash
  gunzip -c /opt/simtas-filkom/backups/db_20270115_020000.sql.gz | \
    docker exec -i simtas-postgres psql -U simtas_prod simtas_filkom_prod
  ```

  **Cek status semua service:**
  ```bash
  docker compose -f docker-compose.prod.yml ps
  docker stats --no-stream
  ```

  **Lihat log error terbaru:**
  ```bash
  docker compose -f docker-compose.prod.yml logs --tail=100 backend | grep -i error
  ```

---

## Done Criteria

- [ ] `docker compose logs -f backend` → log JSON muncul real-time
- [ ] Log rotation aktif: max 50MB × 5 file per service
- [ ] UptimeRobot monitor aktif untuk API dan frontend
- [ ] Matikan backend sementara → UptimeRobot kirim email alert dalam <10 menit
- [ ] `GET /api/v1/health` → response mencakup status database
- [ ] Matikan database → health check return 503
- [ ] Disk usage alert script berfungsi (test manual dengan threshold rendah)
- [ ] Runbook tersedia di `docs/runbook.md`
- [ ] **MILESTONE Phase 6:** Sistem production berjalan penuh, terpantau, dan terdokumentasi operasionalnya
