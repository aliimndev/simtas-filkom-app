#!/bin/bash
# SIMTAS FILKOM — One-time Let's Encrypt SSL setup (Job 30).
# Run as root on the VPS:
#   ./deploy/scripts/ssl-setup.sh
#
# 1. Stops the nginx container (frees port 80 for certbot standalone).
# 2. Issues certs for api + www domains via certbot --standalone.
# 3. Copies fullchain.pem/privkey.pem into infrastructure/nginx/ssl/ (the
#    volume mounted by the nginx container).
# 4. Restarts nginx and registers a renewal cron job.
set -euo pipefail

BASE_DIR="${BASE_DIR:-/opt/simtas-filkom}"
SSL_DIR="$BASE_DIR/infrastructure/nginx/ssl"
DOMAINS=("api.simtas.filkom.unida.ac.id" "simtas.filkom.unida.ac.id")
EMAIL="${EMAIL:-admin@filkom.unida.ac.id}"
COMPOSE="docker compose --env-file $BASE_DIR/deploy/.env.production -f $BASE_DIR/deploy/docker-compose.prod.yml"

if ! command -v certbot >/dev/null 2>&1; then
  apt-get update && apt-get install -y certbot
fi

mkdir -p "$SSL_DIR"

echo "[1/4] Stopping nginx to free port 80 for issuance..."
$COMPOSE stop nginx

CERT_ARGS="certonly --standalone --agree-tos -m $EMAIL --non-interactive --cert-name simtas"
for d in "${DOMAINS[@]}"; do
  CERT_ARGS="$CERT_ARGS -d $d"
done

echo "[2/4] Issuing certificate for ${DOMAINS[*]}..."
certbot $CERT_ARGS

echo "[3/4] Copying certificates into the nginx ssl volume..."
cp /etc/letsencrypt/live/simtas/fullchain.pem "$SSL_DIR/fullchain.pem"
cp /etc/letsencrypt/live/simtas/privkey.pem "$SSL_DIR/privkey.pem"

echo "[4/4] Restarting nginx + registering auto-renewal cron..."
$COMPOSE start nginx

# Renew twice daily (certbot recommends 2x) + reload nginx on renewal.
(crontab -l 2>/dev/null | grep -v -F "simtas-filkom-certbot" || true; \
 echo "17 3,15 * * * certbot renew --quiet --deploy-hook '$COMPOSE exec -T nginx nginx -s reload' # simtas-filkom-certbot") | crontab -

echo "SSL done. Verify: curl -v https://${DOMAINS[0]}/api/v1/health"
