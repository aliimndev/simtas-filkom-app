# Infrastructure — Monitoring

Stack Prometheus + Grafana untuk deployment production.

- `prometheus.yml` — scrape config untuk Bun API (`api:3001/metrics`)
- `grafana/provisioning/` — datasource & dashboard auto-provisioning
- `grafana/dashboards/simtas-api-overview.json` — metrik & uptime

Deploy (setelah stack utama up):

```bash
docker compose -f deploy/docker-compose.monitoring.yml up -d
```

Bun API `/metrics` (Prometheus text format) di-serve langsung oleh Hono,
internal di docker network (tidak dipublikasikan via nginx).
