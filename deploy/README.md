# SIMTAS FILKOM — Production Deployment Guide

Single-VPS deployment (4 vCPU, 8 GB RAM, 150-200 GB SSD) using Docker Compose.

## Architecture

```
Internet
    |
    v
Nginx Reverse Proxy (port 80/443)
    |
    +----------------+----------------+
    |                |                |
    v                v                v
Backend API      Frontend         MinIO Storage
(:8080)          (:3000)          (:9000)
    |
    v
PgBouncer (:6432)
    |
    v
PostgreSQL (:5432)
```

## Prerequisites

- Ubuntu Server LTS (22.04/24.04)
- Docker Engine 24+ and Docker Compose plugin
- Domain name pointing to VPS IP (for SSL)
- At least 150 GB SSD storage

## Quick Deploy

```bash
# 1. Clone repository
git clone https://github.com/aliimndev/simtas-filkom-app.git /opt/simtas-filkom
cd /opt/simtas-filkom

# 2. Copy environment file
cp deploy/.env.production.example .env.production
nano .env.production  # Edit all CHANGE_ME values

# 3. Build and start all services
docker compose -f deploy/docker-compose.prod.yml up -d --build

# 4. Run database migrations
docker compose -f deploy/docker-compose.prod.yml run --rm backend ./simtas-api --migrate

# 5. Seed initial data (roles, admin user, academic year)
docker compose -f deploy/docker-compose.prod.yml run --rm backend ./simtas-seed

# 6. Verify health
curl https://your-domain.com/health
```

## SSL/TLS Certificates

Place your SSL certificates in `infrastructure/nginx/ssl/`:
- `fullchain.pem`
- `privkey.pem`

Or use Let's Encrypt with certbot:
```bash
sudo certbot certonly --standalone -d simtas.filkom.unida.ac.id
sudo cp /etc/letsencrypt/live/simtas.filkom.unida.ac.id/*.pem infrastructure/nginx/ssl/
```

## Backup Strategy

### Automated Daily Backup

Add to crontab (`crontab -e`):
```bash
0 2 * * * cd /opt/simtas-filkom && docker compose -f deploy/docker-compose.prod.yml run --rm --no-deps backup /backup.sh >> /var/log/simtas-backup.log 2>&1
```

### Manual Backup
```bash
docker compose -f deploy/docker-compose.prod.yml run --rm --no-deps backup /backup.sh
```

### Restore
```bash
# Restore PostgreSQL
gunzip -c backups/postgres/simtas_YYYYMMDD_HHMMSS.dump.gz | \
  docker compose -f deploy/docker-compose.prod.yml exec -T postgres pg_restore -U postgres -d simtas_filkom

# Restore MinIO (requires mc client)
mc mirror backups/minio/documents_YYYYMMDD_HHMMSS/ local/simtas-documents
mc mirror backups/minio/archives_YYYYMMDD_HHMMSS/ local/simtas-archives
```

## Resource Limits

| Service | Memory | CPU |
|---------|--------|-----|
| PostgreSQL | 2 GB | 2 cores |
| PgBouncer | 256 MB | 0.5 cores |
| Backend | 1 GB | 1 core |
| Frontend | 512 MB | 0.5 cores |
| Nginx | 128 MB | 0.25 cores |
| MinIO | 2 GB | 1 core |
| **Total** | **~6 GB** | **~5.25 cores** |

Leaves ~2 GB RAM headroom for OS and burst traffic.

## Monitoring

```bash
# Container status
docker compose -f deploy/docker-compose.prod.yml ps

# Logs
docker compose -f deploy/docker-compose.prod.yml logs -f backend
docker compose -f deploy/docker-compose.prod.yml logs -f postgres

# Resource usage
docker stats

# Database connections
docker compose -f deploy/docker-compose.prod.yml exec postgres psql -U postgres -d simtas_filkom -c "SELECT count(*) FROM pg_stat_activity;"
```

## Maintenance

### Update Application
```bash
cd /opt/simtas-filkom
git pull origin main
docker compose -f deploy/docker-compose.prod.yml up -d --build
```

### Database Migration
```bash
docker compose -f deploy/docker-compose.prod.yml run --rm backend ./simtas-api --migrate
```

### Restart Single Service
```bash
docker compose -f deploy/docker-compose.prod.yml restart backend
```

## Security Notes

1. **Never commit `.env.production`** — contains secrets
2. **Rotate credentials** periodically (JWT_SECRET, DB_PASSWORD, MINIO_ROOT_PASSWORD)
3. **Firewall**: Only expose ports 80/443 to internet; block 5432, 9000, 9001
4. **Updates**: Regularly update base images (`docker compose pull`)
5. **Backups**: Store backups off-VPS (rsync to another server or cloud)

## Troubleshooting

### Backend won't start
```bash
docker compose -f deploy/docker-compose.prod.yml logs backend
# Check DB_HOST is set to pgbouncer (not postgres)
```

### MinIO connection refused
```bash
docker compose -f deploy/docker-compose.prod.yml logs minio
# Verify MINIO_ROOT_USER and MINIO_ROOT_PASSWORD are set
```

### Nginx 502 Bad Gateway
```bash
docker compose -f deploy/docker-compose.prod.yml ps
# Ensure backend healthcheck passes: wget http://localhost:8080/api/v1/health
```
