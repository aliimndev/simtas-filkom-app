# Infrastructure — Monitoring

Stack Prometheus + Grafana untuk deployment production.

- `prometheus.yml` — scrape config untuk backend API (`backend:8080/metrics`)
- `grafana/provisioning/` — datasource & dashboard auto-provisioning
- `grafana/dashboards/simtas-api-overview.json` — metrik & uptime

Deploy (setelah stack utama up):

```bash
docker compose -f deploy/docker-compose.monitoring.yml up -d
```

Backend `/metrics` (Prometheus text format) di-serve langsung oleh Gin,
internal di docker network (tidak dipublikasikan via nginx).
