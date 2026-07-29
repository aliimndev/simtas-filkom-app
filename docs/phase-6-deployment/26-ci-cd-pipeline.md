# Job 26 — CI/CD Pipeline (GitHub Actions)

**Phase:** 6 — Deployment
**Referensi PRD:** Section 11 (Deployment — GitHub Actions CI/CD)
**Prerequisites:** Job 25 (Deployment Frontend) ✅
**Estimasi:** 1 hari

---

## Objective

Implementasi pipeline CI/CD penuh menggunakan GitHub Actions: CI (lint, build, test) pada setiap PR, dan CD (build Docker image, push ke registry, deploy ke VPS) otomatis saat merge ke `main`. Setelah job ini selesai, setiap push ke `main` otomatis ter-deploy ke production tanpa intervensi manual.

---

## Checklist

### GitHub Secrets Setup

- [ ] Di GitHub repo → Settings → Secrets and Variables → Actions, tambahkan:
  ```
  VPS_HOST          = IP atau domain VPS
  VPS_USER          = deployer
  VPS_SSH_KEY       = private key SSH untuk user deployer
  VPS_DEPLOY_PATH   = /opt/simtas-filkom

  GHCR_TOKEN        = GitHub Personal Access Token (packages:write)

  VERCEL_TOKEN      = Vercel deployment token
  VERCEL_ORG_ID     = dari vercel.json atau dashboard
  VERCEL_PROJECT_ID = dari vercel.json atau dashboard
  ```

### Workflow 1 — CI (Pull Request)

**File:** `.github/workflows/ci.yml`

```yaml
name: CI

on:
  pull_request:
    branches: [main, develop]
  push:
    branches: [develop]

jobs:
  backend-ci:
    name: Backend — Lint, Build, Test
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: simtas_test
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'
          cache: true
          cache-dependency-path: backend/go.sum

      - name: Install dependencies
        run: go mod download
        working-directory: backend

      - name: Lint
        uses: golangci/golangci-lint-action@v4
        with:
          working-directory: backend
          version: latest

      - name: Build
        run: go build ./...
        working-directory: backend

      - name: Run unit tests
        run: go test ./pkg/... ./internal/usecase/... -v -count=1
        working-directory: backend

      - name: Run integration tests
        run: go test ./internal/handler/... -v -count=1 -tags integration
        working-directory: backend
        env:
          DB_URL: postgres://postgres:postgres@localhost:5432/simtas_test
          JWT_SECRET: test-secret-key-for-ci-only
          APP_ENV: test

      - name: Check Swagger docs
        run: |
          go install github.com/swaggo/swag/cmd/swag@latest
          swag init -g cmd/server/main.go -o /tmp/swagger-check
          diff docs/swagger/swagger.json /tmp/swagger-check/swagger.json || \
            (echo "❌ Swagger docs tidak up-to-date. Jalankan 'make docs' dan commit hasilnya." && exit 1)
        working-directory: backend

  frontend-ci:
    name: Frontend — Lint, Type Check, Build, Test
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Install dependencies
        run: npm ci
        working-directory: frontend

      - name: Type check
        run: npm run type-check
        working-directory: frontend

      - name: Lint
        run: npm run lint
        working-directory: frontend

      - name: Run tests
        run: npm test -- --watchAll=false --coverage
        working-directory: frontend
        env:
          NEXT_PUBLIC_API_URL: http://localhost:8080/api/v1

      - name: Build
        run: npm run build
        working-directory: frontend
        env:
          NEXT_PUBLIC_API_URL: http://localhost:8080/api/v1
          NEXT_PUBLIC_APP_NAME: SIMTAS FILKOM
```

- [ ] Tambah ke `frontend/package.json` script:
  ```json
  "type-check": "tsc --noEmit"
  ```

### Workflow 2 — CD (Deploy ke Production)

**File:** `.github/workflows/cd.yml`

```yaml
name: CD — Deploy to Production

on:
  push:
    branches: [main]

jobs:
  build-and-push-backend:
    name: Build & Push Docker Image
    runs-on: ubuntu-latest
    outputs:
      image-tag: ${{ steps.meta.outputs.version }}

    steps:
      - uses: actions/checkout@v4

      - name: Docker meta
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ghcr.io/${{ github.repository }}/simtas-backend
          tags: |
            type=sha,prefix=sha-
            type=raw,value=latest

      - name: Login ke GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GHCR_TOKEN }}

      - name: Build and Push Docker image
        uses: docker/build-push-action@v5
        with:
          context: ./backend
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  deploy-backend:
    name: Deploy Backend ke VPS
    runs-on: ubuntu-latest
    needs: build-and-push-backend

    steps:
      - name: Deploy via SSH
        uses: appleboy/ssh-action@v1
        with:
          host: ${{ secrets.VPS_HOST }}
          username: ${{ secrets.VPS_USER }}
          key: ${{ secrets.VPS_SSH_KEY }}
          script: |
            cd ${{ secrets.VPS_DEPLOY_PATH }}

            # Pull image terbaru
            echo ${{ secrets.GHCR_TOKEN }} | \
              docker login ghcr.io -u ${{ github.actor }} --password-stdin
            docker pull ghcr.io/${{ github.repository }}/simtas-backend:latest

            # Update IMAGE_TAG di environment
            export IMAGE_TAG=latest

            # Rolling restart (zero-downtime)
            docker compose -f docker-compose.prod.yml up -d --no-deps backend

            # Jalankan migrasi jika ada
            docker compose -f docker-compose.prod.yml run --rm \
              backend ./simtas-backend --migrate

            # Health check
            sleep 5
            curl -f http://localhost:8080/api/v1/health || \
              (echo "❌ Health check gagal!" && exit 1)

            # Cleanup image lama
            docker image prune -f

  deploy-frontend:
    name: Deploy Frontend ke Vercel
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Deploy ke Vercel Production
        uses: amondnet/vercel-action@v25
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
          vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
          working-directory: frontend
          vercel-args: '--prod'
```

### Branch Protection Rules

- [ ] Di GitHub repo → Settings → Branches → Add rule untuk `main`:
  - ✅ Require a pull request before merging
  - ✅ Require status checks to pass: `backend-ci`, `frontend-ci`
  - ✅ Require branches to be up to date before merging
  - ✅ Do not allow bypassing the above settings

### Rollback Strategy

- [ ] Dokumentasikan cara rollback manual jika deploy bermasalah:
  ```bash
  # Di VPS — rollback ke image sebelumnya
  cd /opt/simtas-filkom

  # Lihat image yang tersedia
  docker images | grep simtas-backend

  # Deploy image dengan SHA tertentu
  export IMAGE_TAG=sha-abc1234
  docker compose -f docker-compose.prod.yml up -d --no-deps backend
  ```
- [ ] GitHub Actions menyimpan image dengan tag `sha-{commit}` — bisa digunakan untuk rollback

---

## Done Criteria

- [ ] Buat PR → CI workflow berjalan otomatis (backend + frontend)
- [ ] CI gagal jika ada test yang merah → PR tidak bisa di-merge
- [ ] CI gagal jika Swagger docs tidak up-to-date
- [ ] Merge ke `main` → CD workflow berjalan otomatis
- [ ] CD berhasil → Docker image terbaru di-push ke GHCR
- [ ] CD berhasil → backend ter-deploy ke VPS, health check pass
- [ ] CD berhasil → frontend ter-deploy ke Vercel
- [ ] Rollback: deploy image SHA sebelumnya berfungsi
- [ ] Branch protection aktif: tidak bisa push langsung ke `main` tanpa PR
