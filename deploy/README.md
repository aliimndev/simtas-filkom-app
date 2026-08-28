# SIMTAS FILKOM — Production deployment

The production stack runs the Hono API and Svelte static web server on Bun.
PostgreSQL, PgBouncer, MinIO, Nginx, Prometheus, and Grafana remain separate
infrastructure services.

## Quick deploy

```bash
cp deploy/.env.production.example deploy/.env.production
# Edit every CHANGE_ME value.
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml up -d --build
curl https://your-domain.com/health
```

Services:

| Service | Role | Internal port |
| --- | --- | ---: |
| `api` | Hono REST API on Bun | 3001 |
| `web` | SvelteKit static files served by Bun | 4173 |
| `postgres` | PostgreSQL | 5432 |
| `pgbouncer` | Transaction pooler | 6432 |
| `minio` | S3-compatible storage | 9000 |
| `nginx` | TLS and reverse proxy | 80/443 |

Nginx routes `/api/` and `/health` to `api`, and all other requests to `web`.

## Operations

```bash
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml ps
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml logs -f api
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml logs -f web
docker compose --env-file deploy/.env.production -f deploy/docker-compose.prod.yml restart api
```

Back up PostgreSQL and MinIO with `deploy/scripts/backup.sh`. Never commit
`deploy/.env.production` or TLS private keys.
