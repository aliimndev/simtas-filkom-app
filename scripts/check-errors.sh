#!/bin/bash
# SIMTAS FILKOM — Error alerting (Job 27)
# Deploy ke VPS: /opt/simtas-filkom/scripts/check-errors.sh
# Crontab: 0 * * * * /opt/simtas-filkom/scripts/check-errors.sh
#
# Kirim email alert jika ada >10 error 5xx dalam 1 jam terakhir di Nginx log.

set -uo pipefail

ALERT_EMAIL="${ALERT_EMAIL:-admin@filkom.unida.ac.id}"
ACCESS_LOG="${ACCESS_LOG:-/var/log/nginx/access.log}"
THRESHOLD="${THRESHOLD:-10}"

if [ ! -f "$ACCESS_LOG" ]; then
  echo "Access log tidak ditemukan: $ACCESS_LOG"
  exit 1
fi

# Hitung 5xx dalam 1 jam terakhir
ERROR_COUNT=$(grep ' 5[0-9][0-9] ' "$ACCESS_LOG" | \
  awk -v d="$(date -d '1 hour ago' '+%d/%b/%Y:%H')" '$4 > "["d' | wc -l)

if [ "$ERROR_COUNT" -gt "$THRESHOLD" ]; then
  echo "ALERT: $ERROR_COUNT error 5xx dalam 1 jam terakhir di SIMTAS FILKOM" | \
    mail -s "SIMTAS Alert: Server Errors" "$ALERT_EMAIL"
  echo "Alert terkirim: $ERROR_COUNT error"
else
  echo "OK: $ERROR_COUNT error 5xx (threshold $THRESHOLD)"
fi
