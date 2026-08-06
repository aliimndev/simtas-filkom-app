# Infrastructure — Monitoring (Job 27)

Folder ini berisi konfigurasi monitoring & logging untuk deployment production
(akan diisi saat Job 27 — Monitoring & Logging dikerjakan).

Referensi: `docs/ROADMAP.md` (item Monitoring).

Rencana isi folder ini:
- `prometheus.yml` — scrape config untuk backend API
- `grafana/dashboards/` — dashboard metrik & uptime
- `loki/` — aggregasi log backend (opsional)
- Alert rules untuk notifikasi (email/Slack)
