# simtas-filkom-app

Sistem Manajemen Tugas Akhir dan Skripsi Fakultas Ilmu Komputer Universitas Djuanda

**Version:** 1.0
**Status:** Development

## Tech Stack

- **Frontend:** Next.js 16 + React 19 + TypeScript + Tailwind CSS v4
- **Backend:** Go 1.24+ (Gin Framework, GORM, Clean Architecture)
- **Database:** PostgreSQL 16
- **Storage:** Supabase Storage
- **Email:** Resend
- **Auth:** JWT

## Prerequisites

- Go 1.24+
- Node.js 20+
- Docker & Docker Compose
- PostgreSQL 16 (if running locally without Docker)

## Setup Lokal

### 1. Clone Repository

```bash
git clone https://github.com/aliimndev/simtas-filkom-app.git
cd simtas-filkom-app
```

## Struktur Repository

```
simtas-filkom-app/
├── backend/          # Go API (Gin, GORM, Clean Architecture)
├── frontend/         # Next.js 16 + React 19 + TypeScript
├── docs/             # Dokumentasi & timeline (30 job)
├── infrastructure/   # Nginx, monitoring, konfigurasi deploy
├── scripts/          # Script operasional (backup DB, dll)
├── .github/          # CI/CD (GitHub Actions)
├── docker-compose.yml
└── README.md
```

### 2. Backend Setup

```bash
cd backend
cp .env.example .env
# Edit .env sesuai konfigurasi lokal Anda
go mod download
go run ./cmd/server
```

### 3. Frontend Setup

```bash
cd frontend
cp .env.local.example .env.local
npm install
npm run dev
```

### 4. Docker (Opsional)

```bash
# Jalankan semua service (PostgreSQL + Backend)
docker compose up -d
```

Akses aplikasi:
- Frontend: http://localhost:3000
- API: http://localhost:8080/api/v1
- Health Check: http://localhost:8080/api/v1/health
