# SIMTAS FILKOM — Test tooling

## API smoke/load tests

Requires [k6](https://k6.io/docs/get-started/installation/).

```bash
k6 run tests/k6/api-smoke.js \
  -e BASE_URL=https://api.simtas.filkom.unida.ac.id \
  -e EMAIL=mahasiswa@filkom.unida.ac.id -e PASSWORD='...' \
  -e ADMIN_EMAIL=admin@filkom.unida.ac.id -e ADMIN_PASSWORD='...'

k6 run tests/k6/load.js \
  -e BASE_URL=https://api.simtas.filkom.unida.ac.id \
  -e EMAIL=admin@filkom.unida.ac.id -e PASSWORD='...' \
  -e THESIS_ID=<id-aktif>
```

## Browser E2E tests

The active frontend is `apps/web` and runs on Bun. Install Playwright once
with Bun, then run the configured browser tests from that workspace when E2E
coverage is added:

```bash
cd apps/web
bunx playwright install
bun test
```

API integration coverage lives in `apps/api/test` and can be run with:

```bash
cd apps/api
bun test test/
```

## Security scan

OWASP ZAP can scan the staging URL:

```bash
zap-full-scan.py -t https://<staging-url> -r report.html
```
