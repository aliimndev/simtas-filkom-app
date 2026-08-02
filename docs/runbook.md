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
/opt/simtas-filkom/backups/backup.sh
```

## Restore Database dari Backup

```bash
gunzip -c /opt/simtas-filkom/backups/db_20270115_020000.sql.gz | \
  docker exec -i simtas-postgres psql -U simtas_prod simtas_filkom_prod
```

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
