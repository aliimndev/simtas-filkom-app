#!/bin/bash
# SIMTAS FILKOM — Install backup + alerting cron jobs (idempotent).
# Run once on the VPS as the deployer user:
#   ./deploy/scripts/install-backup-cron.sh
#
# Installs:
#   - Daily PostgreSQL+MinIO backup at 02:00        -> deploy/scripts/backup.sh
#   - Hourly 5xx error alert                        -> scripts/check-errors.sh
#   - Daily disk-usage alert at 08:00 (>80% used)
set -euo pipefail

BASE_DIR="${BASE_DIR:-/opt/simtas-filkom}"
CRON_MARKER="# simtas-filkom"

# Remove existing SIMTAS entries so re-running never duplicates.
crontab -l 2>/dev/null | grep -v -F "$CRON_MARKER" | crontab - || true

{
  cat <<EOF
# ── SIMTAS FILKOM automated jobs ($CRON_MARKER)
0 2 * * * $BASE_DIR/deploy/scripts/backup.sh >> $BASE_DIR/logs/backup.log 2>&1 $CRON_MARKER
0 * * * * $BASE_DIR/scripts/check-errors.sh $CRON_MARKER
0 8 * * * df -h / | awk 'NR==2 && \$5+0 >= 80 {print "ALERT: disk usage " \$5} ' | mail -s "SIMTAS Alert: Disk Usage" "${ALERT_EMAIL:-admin@filkom.unida.ac.id}" $CRON_MARKER
EOF
  crontab -l 2>/dev/null
} | crontab -

echo "Crontab installed:"
crontab -l | grep -F "$CRON_MARKER"
echo ""
echo "Manual test backup: $BASE_DIR/deploy/scripts/backup.sh"
echo "Manual test restore: see docs/runbook.md (# Restore Database dari Backup)"
