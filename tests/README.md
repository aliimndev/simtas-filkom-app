# SIMTAS FILKOM — Test Tooling

## k6 (API smoke + load, Fase 3)

Requires [k6](https://k6.io/docs/get-started/installation/).

```bash
# API smoke test collection (auth, RBAC, pagination, error shape)
k6 run tests/k6/api-smoke.js \
  -e BASE_URL=https://api.simtas.filkom.unida.ac.id \
  -e EMAIL=mahasiswa@filkom.unida.ac.id -e PASSWORD='...' \
  -e ADMIN_EMAIL=admin@filkom.unida.ac.id -e ADMIN_PASSWORD='...'

# Load test (50 VU ramp, 10 min, baseline p95)
k6 run tests/k6/load.js \
  -e BASE_URL=https://api.simtas.filkom.unida.ac.id \
  -e EMAIL=admin@filkom.unida.ac.id -e PASSWORD='...' \
  -e THESIS_ID=<id-aktif>   # opsional: aktifkan branch upload dokumen
```

Target baseline (PRODUCTION-READINESS-REVIEW §7E): p95 < 500 ms at 50 VU.

## Playwright (E2E, Fase 3 #4)

Runs against a live stack (frontend + backend + seeded DB).

```bash
cd frontend
npm run e2e:install          # sekali: install Chromium
E2E_EMAIL=mahasiswa@... E2E_PASSWORD='...' npm run e2e
```

Backend full-lifecycle coverage sudah ada di `TestFullThesisLifecycle`
(`backend/internal/handler/lifecycle_test.go`); E2E menutup lapisan UI.

## Security scan (Fase 3 #3)

OWASP ZAP aktif scan pada environment staging:
`zap-full-scan.py -t https://<staging-url> -r report.html`
