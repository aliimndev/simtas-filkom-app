#!/bin/sh
# SIMTAS FILKOM — Production backup script
# Run from host: cd /opt/simtas-filkom && ./deploy/scripts/backup.sh
set -euo pipefail

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/opt/simtas-filkom/backups"
RETENTION_DAYS=30
COMPOSE="docker compose --env-file /opt/simtas-filkom/deploy/.env.production -f /opt/simtas-filkom/deploy/docker-compose.prod.yml"

echo "[$(date)] Starting backup..."

# ── PostgreSQL backup ────────────────────────────────────────────────────
echo "[$(date)] Backing up PostgreSQL..."
mkdir -p "${BACKUP_DIR}/postgres"
$COMPOSE exec -T postgres pg_dump -U ${DB_USER:-postgres} -d ${DB_NAME:-simtas_filkom} \
  -F c -b -v -f "/tmp/simtas_${TIMESTAMP}.dump"
$COMPOSE cp postgres:"/tmp/simtas_${TIMESTAMP}.dump" "${BACKUP_DIR}/postgres/simtas_${TIMESTAMP}.dump"
$COMPOSE exec -T postgres rm "/tmp/simtas_${TIMESTAMP}.dump"
echo "[$(date)] PostgreSQL backup complete."

# ── MinIO backup ─────────────────────────────────────────────────────────
echo "[$(date)] Backing up MinIO..."
mkdir -p "${BACKUP_DIR}/minio"
$COMPOSE exec -T minio mc mirror --overwrite local/simtas-documents "${BACKUP_DIR}/minio/documents_${TIMESTAMP}" || true
$COMPOSE exec -T minio mc mirror --overwrite local/simtas-archives "${BACKUP_DIR}/minio/archives_${TIMESTAMP}" || true
echo "[$(date)] MinIO backup complete."

# ── Retention: delete backups older than RETENTION_DAYS ──────────────────
echo "[$(date)] Cleaning up old backups (older than ${RETENTION_DAYS} days)..."
find "${BACKUP_DIR}/postgres" -type f -mtime +${RETENTION_DAYS} -delete || true
find "${BACKUP_DIR}/minio" -type f -mtime +${RETENTION_DAYS} -delete || true

echo "[$(date)] Backup complete."
