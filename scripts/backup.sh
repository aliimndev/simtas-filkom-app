#!/bin/bash
# SIMTAS FILKOM — Database backup harian (Job 24)
# Deploy ke VPS: /opt/simtas-filkom/backups/backup.sh
# Crontab: 0 2 * * * /opt/simtas-filkom/backups/backup.sh

set -euo pipefail

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/opt/simtas-filkom/backups"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-simtas_filkom}"
# Nama container PostgreSQL — set container_name di docker-compose prod agar konsisten.
PG_CONTAINER="${PG_CONTAINER:-simtas-postgres}"

mkdir -p "$BACKUP_DIR"

docker exec "$PG_CONTAINER" pg_dump -U "$DB_USER" "$DB_NAME" \
    | gzip > "$BACKUP_DIR/db_${TIMESTAMP}.sql.gz"

# Hapus backup lebih dari 30 hari
find "$BACKUP_DIR" -name "*.sql.gz" -mtime +30 -delete

echo "Backup selesai: $BACKUP_DIR/db_${TIMESTAMP}.sql.gz"
