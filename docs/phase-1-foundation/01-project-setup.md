# Job 01 — Project Setup & Infrastructure

**Phase:** 1 — Foundation
**Referensi PRD:** Section 3 (Scope), Section 9 (System Architecture), Section 10 (Technology Stack)
**Prerequisites:** Tidak ada (job pertama)
**Estimasi:** 3–4 hari

---

## Objective

Menyiapkan seluruh fondasi proyek: struktur repository, konfigurasi environment development, Docker untuk local development, dan CI pipeline dasar. Setelah job ini selesai, seluruh service dapat berjalan secara lokal dengan satu perintah.

---

## Checklist

### Repository & Struktur Proyek
- [ ] Buat repository GitHub `simtas-filkom-app`
- [ ] Setup struktur monorepo:
  ```
  simtas-filkom-app/
  ├── api/          # Go application
  ├── client/         # Next.js application
  ├── docs/             # Dokumentasi
  ├── docker-compose.yml
  ├── .gitignore
  ├── .env.example
  └── README.md
  ```
- [ ] Setup `.gitignore` untuk Go, Node.js, dan environment files
- [ ] Buat `README.md` dengan instruksi setup lokal

### Backend — Go Setup
- [ ] Inisiasi Go module:
  ```bash
  go mod init github.com/aliimndev/simtas-filkom-app/backend
  ```
- [ ] Install dependencies awal:
  ```bash
  go get github.com/gin-gonic/gin
  go get gorm.io/gorm
  go get gorm.io/driver/postgres
  go get github.com/golang-jwt/jwt/v5
  go get github.com/joho/godotenv
  go get github.com/go-playground/validator/v10
  go get github.com/google/uuid
  go get golang.org/x/crypto
  ```
- [ ] Buat struktur Clean Architecture:
  ```
  backend/
  ├── cmd/
  │   └── server/
  │       └── main.go
  ├── internal/
  │   ├── domain/
  │   │   ├── entity/       # Struct entitas database
  │   │   └── repository/   # Interface repository
  │   ├── usecase/          # Business logic
  │   ├── repository/       # Implementasi repository (GORM)
  │   ├── handler/          # HTTP handlers (Gin)
  │   │   └── dto/          # Request/Response structs
  │   └── middleware/       # Auth, RBAC, rate limit, logger
  ├── pkg/
  │   ├── config/           # App configuration loader
  │   ├── database/         # DB connection & migration runner
  │   ├── storage/          # Supabase storage client
  │   ├── email/            # Resend email client
  │   └── response/         # Standard API response helpers
  ├── migrations/           # SQL migration files (golang-migrate)
  ├── Dockerfile
  ├── .env.example
  └── go.mod
  ```
- [ ] Implementasi `GET /api/v1/health` endpoint — return `{ "status": "ok", "version": "1.0.0", "timestamp": "..." }`
- [ ] Buat `pkg/response/response.go` — standard response format:
  ```json
  {
    "success": true,
    "message": "...",
    "data": {},
    "meta": { "page": 1, "per_page": 20, "total": 100 }
  }
  ```
- [ ] Konfigurasi environment variables di `.env.example`:
  ```env
  # App
  APP_PORT=8080
  APP_ENV=development

  # Database
  DB_HOST=localhost
  DB_PORT=5432
  DB_USER=postgres
  DB_PASSWORD=postgres
  DB_NAME=simtas_filkom

  # JWT
  JWT_SECRET=your-super-secret-key
  JWT_EXPIRY=24h
  JWT_REFRESH_EXPIRY=168h

  # Storage (Supabase)
  SUPABASE_URL=https://xxx.supabase.co
  SUPABASE_KEY=your-supabase-anon-key
  SUPABASE_BUCKET=simtas-documents

  # Email (Resend)
  RESEND_API_KEY=re_xxxxxxxxxxxx
  EMAIL_FROM=noreply@filkom.unida.ac.id
  EMAIL_FROM_NAME=SIMTAS FILKOM

  # CORS
  CORS_ALLOWED_ORIGINS=http://localhost:3000
  ```

### Frontend — Next.js Setup
- [ ] Inisiasi project:
  ```bash
  npx create-next-app@latest frontend \
    --typescript \
    --tailwind \
    --eslint \
    --app \
    --src-dir \
    --import-alias "@/*"
  ```
- [ ] Install dependencies:
  ```bash
  npm install @tanstack/react-query @tanstack/react-query-devtools
  npm install react-hook-form @hookform/resolvers zod
  npm install axios
  npm install clsx tailwind-merge
  npm install lucide-react
  npm install next-themes
  ```
- [ ] Setup struktur folder:
  ```
  frontend/src/
  ├── app/
  │   ├── (auth)/
  │   │   └── login/
  │   │       └── page.tsx
  │   ├── (dashboard)/
  │   │   ├── layout.tsx       # Layout dengan sidebar
  │   │   └── dashboard/
  │   │       └── page.tsx
  │   ├── layout.tsx           # Root layout
  │   └── page.tsx             # Redirect ke /dashboard atau /login
  ├── components/
  │   ├── ui/                  # Reusable: Button, Input, Table, Modal, Badge, etc.
  │   └── features/            # Feature-specific components
  ├── lib/
  │   ├── api/
  │   │   ├── client.ts        # Axios instance dengan interceptors
  │   │   └── endpoints/       # API endpoint functions per module
  │   ├── hooks/               # Custom React hooks
  │   ├── stores/              # Zustand stores (jika diperlukan)
  │   └── utils/               # Helper functions, formatters
  ├── types/                   # TypeScript global types & interfaces
  └── constants/               # App constants (routes, labels, etc.)
  ```
- [ ] Konfigurasi `next.config.ts`:
  ```ts
  const nextConfig = {
    env: { NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL },
    images: { domains: ['xxx.supabase.co'] }
  }
  ```
- [ ] Setup Tailwind CSS custom theme di `tailwind.config.ts` — warna brand FILKOM Unida
- [ ] Setup TanStack Query provider di `app/layout.tsx`
- [ ] Buat `lib/api/client.ts` — Axios instance dengan:
  - `baseURL` dari env
  - Request interceptor: attach JWT token dari localStorage/cookie
  - Response interceptor: handle 401 (redirect ke login), 403 (toast error)
- [ ] Tambah `.env.local.example`:
  ```env
  NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
  ```

### Docker & Local Development
- [ ] Buat `backend/Dockerfile` (multi-stage build):
  ```dockerfile
  # Stage 1: Build
  FROM golang:1.24-alpine AS builder
  WORKDIR /app
  COPY go.mod go.sum ./
  RUN go mod download
  COPY . .
  RUN go build -o simtas-backend ./cmd/server

  # Stage 2: Run
  FROM alpine:latest
  WORKDIR /app
  COPY --from=builder /app/simtas-backend .
  EXPOSE 8080
  CMD ["./simtas-backend"]
  ```
- [ ] Buat `docker-compose.yml` di root:
  ```yaml
  services:
    postgres:
      image: postgres:16-alpine
      environment:
        POSTGRES_USER: postgres
        POSTGRES_PASSWORD: postgres
        POSTGRES_DB: simtas_filkom
      ports:
        - "5432:5432"
      volumes:
        - postgres_data:/var/lib/postgresql/data

    backend:
      build: ./backend
      ports:
        - "8080:8080"
      env_file: ./backend/.env
      depends_on:
        - postgres

  volumes:
    postgres_data:
  ```
- [ ] Test: `docker compose up` → semua service running tanpa error
- [ ] Pastikan backend dapat connect ke PostgreSQL di dalam Docker network

### CI/CD Pipeline Awal
- [ ] Buat `.github/workflows/ci.yml`:
  ```yaml
  name: CI
  on:
    push:
      branches: ['**']
    pull_request:
      branches: [main]
  jobs:
    backend:
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@v4
        - uses: actions/setup-go@v5
          with: { go-version: '1.24' }
        - run: go build ./...
          working-directory: backend
        - run: go vet ./...
          working-directory: backend
    frontend:
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@v4
        - uses: actions/setup-node@v4
          with: { node-version: '20' }
        - run: npm ci
          working-directory: frontend
        - run: npm run build
          working-directory: frontend
  ```
- [ ] Setup branch protection rule di GitHub: PR ke `main` wajib pass CI

---

## Done Criteria

- [ ] `docker compose up` berhasil → PostgreSQL dan backend berjalan tanpa error
- [ ] `GET http://localhost:8080/api/v1/health` → `{ "status": "ok" }`
- [ ] `cd frontend && npm run dev` → Next.js berjalan di `http://localhost:3000`
- [ ] `cd frontend && npm run build` → build sukses tanpa error TypeScript
- [ ] GitHub Actions CI berjalan hijau pada push pertama
- [ ] Tidak ada file `.env` yang ter-commit (hanya `.env.example`)
- [ ] Struktur folder backend dan frontend sesuai diagram di atas
