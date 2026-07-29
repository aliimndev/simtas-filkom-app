# Job 01 — Project Setup & Infrastructure

## Progress Tracking

### 1. Repository & Struktur Proyek
- [x] Buat struktur monorepo (api/, client/, docker-compose.yml, .gitignore, .env.example)
- [x] Update README.md dengan instruksi setup lokal

### 2. API — Go Setup
- [x] Init Go module & install dependencies
- [x] Buat struktur Clean Architecture
- [x] Implementasi `GET /api/v1/health` endpoint
- [x] Buat `pkg/response/response.go`
- [x] Konfigurasi `.env.example`

### 3. Client — Next.js Setup
- [x] Init Next.js project
- [x] Install dependencies
- [x] Buat struktur folder
- [x] Konfigurasi next.config.ts, Tailwind, TanStack Query
- [x] Buat lib/api/client.ts (Axios instance)
- [x] Buat .env.local.example

### 4. Docker & Local Development
- [x] Buat api/Dockerfile
- [x] Buat docker-compose.yml

### 5. CI/CD Pipeline
- [x] Buat .github/workflows/ci.yml

### 6. Done Criteria Verification
- [ ] Test semua service berjalan (requires docker/postgres running)