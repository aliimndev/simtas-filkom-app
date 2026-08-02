# Job 24 — Deployment Backend (VPS Ubuntu + Docker + Nginx)

**Phase:** 6 — Deployment
**Referensi PRD:** Section 11 (Deployment), Section 7 (Non-Functional Requirements)
**Prerequisites:** Job 23 (Testing Strategy) ✅
**Estimasi:** 2 hari

---

## Objective

Deploy backend Go ke VPS Ubuntu menggunakan Docker, konfigurasi Nginx sebagai reverse proxy dengan TLS, dan setup PostgreSQL production. Setelah job ini selesai, backend API berjalan di production dan bisa diakses dari internet via HTTPS.

---

## Checklist

### Persiapan VPS

- [ ] Spesifikasi minimum VPS: 2 vCPU, 2 GB RAM, 20 GB SSD, Ubuntu 22.04 LTS
- [ ] Setup awal server:
  ```bash
  # Update system
  apt update && apt upgrade -y

  # Install Docker
  curl -fsSL https://get.docker.com | sh
  usermod -aG docker $USER

  # Install Docker Compose
  apt install docker-compose-plugin -y

  # Install Nginx
  apt install nginx -y

  # Install Certbot (untuk SSL)
  apt install certbot python3-certbot-nginx -y

  # Setup firewall
  ufw allow 22    # SSH
  ufw allow 80    # HTTP
  ufw allow 443   # HTTPS
  ufw enable
  ```
- [ ] Buat user non-root untuk deploy: `useradd -m -s /bin/bash deployer`
- [ ] Setup SSH key untuk user deployer (akan digunakan GitHub Actions)

### Struktur Folder di VPS

```
/opt/simtas-filkom/
├── docker-compose.prod.yml
├── .env.production          # JANGAN commit ke repo
├── nginx/
│   └── simtas.conf
└── backups/                 # backup database
```

### Docker Compose Production

- [x] Buat `docker-compose.prod.yml` di VPS (bukan di repo): template tersedia di deploy/docker-compose.prod.yml (copy + isi .env.production di VPS)
  ```yaml
  version: '3.8'
  services:
    postgres:
      image: postgres:16-alpine
      restart: always
      environment:
        POSTGRES_USER: ${DB_USER}
        POSTGRES_PASSWORD: ${DB_PASSWORD}
        POSTGRES_DB: ${DB_NAME}
      volumes:
        - postgres_data:/var/lib/postgresql/data
      networks:
        - simtas-network
      # Tidak expose port ke host — hanya internal network

    backend:
      image: ghcr.io/${GITHUB_REPO}/simtas-backend:${IMAGE_TAG}
      restart: always
      env_file: .env.production
      ports:
        - "127.0.0.1:8080:8080"  # hanya localhost — Nginx yang expose ke luar
      depends_on:
        - postgres
      networks:
        - simtas-network
      healthcheck:
        test: ["CMD", "wget", "-qO-", "http://localhost:8080/api/v1/health"]
        interval: 30s
        timeout: 10s
        retries: 3

  volumes:
    postgres_data:

  networks:
    simtas-network:
      driver: bridge
  ```

### Konfigurasi Nginx

- [ ] Buat `/etc/nginx/sites-available/simtas.conf`:
  ```nginx
  # Redirect HTTP ke HTTPS
  server {
      listen 80;
      server_name api.simtas.filkom.unida.ac.id;
      return 301 https://$host$request_uri;
  }

  server {
      listen 443 ssl http2;
      server_name api.simtas.filkom.unida.ac.id;

      ssl_certificate     /etc/letsencrypt/live/api.simtas.filkom.unida.ac.id/fullchain.pem;
      ssl_certificate_key /etc/letsencrypt/live/api.simtas.filkom.unida.ac.id/privkey.pem;
      ssl_protocols       TLSv1.2 TLSv1.3;
      ssl_ciphers         HIGH:!aNULL:!MD5;

      # Upload file size limit
      client_max_body_size 25M;

      # Security headers
      add_header X-Frame-Options "DENY" always;
      add_header X-Content-Type-Options "nosniff" always;
      add_header X-XSS-Protection "1; mode=block" always;
      add_header Strict-Transport-Security "max-age=31536000" always;

      # Rate limiting
      limit_req_zone $binary_remote_addr zone=api:10m rate=30r/s;
      limit_req zone=api burst=50 nodelay;

      location / {
          proxy_pass         http://127.0.0.1:8080;
          proxy_http_version 1.1;
          proxy_set_header   Host $host;
          proxy_set_header   X-Real-IP $remote_addr;
          proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
          proxy_set_header   X-Forwarded-Proto $scheme;
          proxy_read_timeout 120s;
          proxy_connect_timeout 10s;
      }
  }
  ```
- [ ] Enable config: `ln -s /etc/nginx/sites-available/simtas.conf /etc/nginx/sites-enabled/`
- [ ] Test config: `nginx -t`
- [ ] Setup SSL: `certbot --nginx -d api.simtas.filkom.unida.ac.id`
- [ ] Auto-renew certbot: `crontab -e` → `0 12 * * * certbot renew --quiet`

### Database Production

- [ ] PostgreSQL berjalan di dalam Docker (tidak expose port ke luar)
- [ ] Buat `.env.production` di VPS dengan value production:
  ```env
  APP_ENV=production
  APP_PORT=8080
  DB_HOST=postgres
  DB_PORT=5432
  DB_USER=simtas_prod
  DB_PASSWORD=[strong-random-password]
  DB_NAME=simtas_filkom_prod
  JWT_SECRET=[strong-random-secret-256-bit]
  SUPABASE_URL=https://xxx.supabase.co
  SUPABASE_KEY=[production-key]
  RESEND_API_KEY=[production-key]
  EMAIL_FROM=noreply@filkom.unida.ac.id
  CORS_ALLOWED_ORIGINS=https://simtas.filkom.unida.ac.id
  FRONTEND_URL=https://simtas.filkom.unida.ac.id
  EMAIL_DEV_MODE=false
  STORAGE_PROVIDER=supabase
  # Wajib untuk docker-compose.prod.yml (nama image GHCR)
  GITHUB_REPO=aliimndev/simtas-filkom-app
  IMAGE_TAG=latest
  ```
- [ ] Jalankan migrasi saat pertama deploy:
  ```bash
  docker compose -f docker-compose.prod.yml run --rm backend ./simtas-backend --migrate
  ```

### Database Backup

- [x] Script backup harian `backups/backup.sh` (ada di scripts/backup.sh):
  ```bash
  #!/bin/bash
  TIMESTAMP=$(date +%Y%m%d_%H%M%S)
  docker exec simtas-postgres pg_dump -U $DB_USER $DB_NAME \
    | gzip > /opt/simtas-filkom/backups/db_$TIMESTAMP.sql.gz
  # Hapus backup lebih dari 30 hari
  find /opt/simtas-filkom/backups/ -name "*.sql.gz" -mtime +30 -delete
  ```
- [ ] Setup crontab: `0 2 * * * /opt/simtas-filkom/backups/backup.sh`

---

## Done Criteria

- [ ] `https://api.simtas.filkom.unida.ac.id/api/v1/health` → `{ "status": "ok" }`
- [ ] HTTP → HTTPS redirect berfungsi
- [ ] SSL certificate valid (A rating di ssllabs.com)
- [ ] Upload file 20MB berhasil (client_max_body_size 25M)
- [ ] `docker compose -f docker-compose.prod.yml ps` → semua service status "healthy"
- [ ] Database tidak bisa diakses dari luar server (port tidak terbuka)
- [ ] Backup database berjalan: file `.sql.gz` terbuat di folder backups
- [ ] Restart server → backend otomatis restart (`restart: always`)
- [ ] Security headers muncul di response production
