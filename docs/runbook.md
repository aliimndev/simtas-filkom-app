# Runbook — SIMTAS FILKOM (Job 27)

Prosedur operasional untuk server production (`/opt/simtas-filkom`).
Semua perintah dijalankan sebagai user `deployer` di VPS.

---

## Restart Backend

```bash
cd /opt/simtas-filkom
docker compose -f docker-compose.prod.yml restart backend
```

## Restart Semua Service

```bash
cd /opt/simtas-filkom
docker compose -f docker-compose.prod.yml down
docker compose -f docker-compose.prod.yml up -d
```

## Cek Status Semua Service

```bash
docker compose -f docker-compose.prod.yml ps
docker stats --no-stream
```

## Backup Database (Manual)

```bash
/opt/simtas-filkom/deploy/scripts/backup.sh
```

## Automated Backup (Cron)

```bash
# Sekali saja setelah deploy:
/opt/simtas-filkom/deploy/scripts/install-backup-cron.sh
```

Installs: backup harian 02.00 (PostgreSQL + MinIO), alert error 5xx per jam,
alert disk usage >80% per hari. Lihat `docs/runbook.md#alerting`.

## Restore Database dari Backup

```bash
# Dump format custom (pg_restore)
docker exec -i simtas-postgres pg_restore -U postgres -d simtas_filkom \
  --clean --if-exists /opt/simtas-filkom/backups/postgres/simtas_YYYYMMDD_HHMMSS.dump
```

## SSL/TLS (Let's Encrypt)

```bash
# Sekali saja (root di VPS):
/opt/simtas-filkom/deploy/scripts/ssl-setup.sh

# Verifikasi
curl -v https://api.simtas.filkom.unida.ac.id/api/v1/health
# Auto-renewal: cron 2x/hari + nginx reload (didaftarkan otomatis oleh script)
```

## Monitoring (Prometheus + Grafana)

```bash
docker compose -f /opt/simtas-filkom/deploy/docker-compose.monitoring.yml up -d
# Grafana UI: http://<vps-ip>:3001  (login: GRAFANA_ADMIN_USER / GRAFANA_ADMIN_PASSWORD)
# Dashboard: SIMTAS FILKOM — API Overview (request rate, p95, error rate, up)
# Backend metrics: http://backend:8080/metrics (internal, tidak dipublikasikan)
```

## Error Tracking (Sentry)

- `SENTRY_DSN` di `.env.production` backend → panic/error dikirim otomatis.
- `NEXT_PUBLIC_SENTRY_DSN` di frontend → ErrorBoundary & error ter-capture.
- Tanpa DSN kedua-duanya no-op, aman untuk dev.

## Lihat Log

```bash
# Log backend real-time
docker compose -f docker-compose.prod.yml logs -f backend

# 100 baris terakhir
docker compose -f docker-compose.prod.yml logs --tail=100 backend

# Error terbaru
docker compose -f docker-compose.prod.yml logs --tail=100 backend | grep -i error

# Log PostgreSQL
docker compose -f docker-compose.prod.yml logs -f postgres
```

## Analisis Nginx Log

```bash
# 10 URL paling sering di-request
awk '{print $7}' /var/log/nginx/access.log | sort | uniq -c | sort -rn | head -10

# Request dengan status 5xx
grep ' 5[0-9][0-9] ' /var/log/nginx/access.log | tail -50
```

## Rollback ke Image Sebelumnya

```bash
cd /opt/simtas-filkom
docker images | grep simtas-backend

# Deploy image dengan SHA tertentu
export IMAGE_TAG=sha-abc1234
docker compose -f docker-compose.prod.yml up -d --no-deps backend
```

## Run Migrasi

```bash
docker compose -f docker-compose.prod.yml run --rm backend ./simtas-api --migrate
```

## Run Seed Data

Ada dua jenis seed:

```bash
# 1. Seed SQL dasar (roles, tahun akademik, akun default) — idempotent
docker compose -f docker-compose.prod.yml run --rm backend ./simtas-api --seed

# 2. Seed akun production (Job 29) — buat akun admin/kaprodi dengan password
#    acak kuat + must_change_password=true. Password dicetak sekali ke stdout.
docker compose -f docker-compose.prod.yml run --rm -T backend \
  env EMAIL_ADMIN=admin@filkom.unida.ac.id EMAIL_KAPRODI=kaprodi@filkom.unida.ac.id \
  ./simtas-seed
```

## Health Check

```bash
curl http://localhost:8080/api/v1/health
# {"status":"ok","version":"1.0.0","timestamp":"...","uptime_seconds":...,"database":"ok"}
```

---

## Alerting

- **UptimeRobot** memonitor `https://api.simtas.filkom.unida.ac.id/api/v1/health` dan frontend URL (interval 5 menit).
- **Error 5xx**: `scripts/check-errors.sh` via cron setiap jam → email `admin@filkom.unida.ac.id`.
- **Disk usage** >80%: cron harian jam 8 pagi → email alert.

## Kontak Support

- Developer / IT FILKOM: [isi sesuai tim]
