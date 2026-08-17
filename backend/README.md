# SIMTAS FILKOM — Backend API

RESTful API for the **Sistem Manajemen Tugas Akhir Skripsi** (Thesis & Final Project Management System) of the Faculty of Computer Science, Universitas Djuanda.

Built with **Go**, **Gin**, **GORM**, and **PostgreSQL**. It powers authentication, thesis/consultation/document/seminar/defense workflows, archive management, dashboards, notifications, and audit logging consumed by the frontend.

---

## Tech Stack

| Layer      | Technology                                            |
| ---------- | ------------------------------------------------------ |
| Language   | Go 1.26                                                |
| HTTP       | Gin-gonic                                              |
| ORM        | GORM (PostgreSQL driver)                               |
| Database   | PostgreSQL 17 / 18                                     |
| Migrations | golang-migrate (SQL files under `migrations/`)         |
| Auth       | JWT (access + refresh tokens), bcrypt                  |
| Storage    | Local disk / Supabase / S3-compatible (MinIO, R2, B2) |
| Email      | Resend                                                 |
| Docs       | Swagger / OpenAPI (generated)                          |

---

## Prerequisites

- **Go 1.26+** — check with `go version`
- **PostgreSQL 17 or 18** — either a local install or Docker
- **Docker** (optional, only if you run the DB via `docker compose`)
- **`swag` CLI** (optional) — only needed to regenerate Swagger docs

---

## Project Layout

```
backend/
├── cmd/
│   ├── server/main.go    # Entry point: API server + CLI flags (--migrate, --seed)
│   └── seed/main.go      # Production admin/kaprodi account seed
├── internal/
│   ├── domain/           # Entities, repository & service interfaces
│   ├── repository/       # GORM data access
│   ├── usecase/          # Business logic
│   ├── handler/          # HTTP handlers + router + DTOs
│   └── middleware/       # Auth, role, rate-limit middlewares
├── migrations/           # SQL migrations + seed data
├── pkg/                  # Shared infra (config, database, jwt, storage, email, ...)
├── docs/                 # Generated Swagger docs
├── Makefile              # Convenience dev/CI targets
└── .env.example          # Template for configuration
```

---

## 1. Configuration

Copy the example file and adjust values:

```bash
cp .env.example .env
```

The backend loads `.env` automatically (via godotenv) and also falls back to system environment variables.

Key variables (see `.env.example` for the full list):

| Variable                         | Default                          | Description                                    |
| -------------------------------- | -------------------------------- | ---------------------------------------------- |
| `APP_PORT`                       | `8080`                           | HTTP listen port                               |
| `APP_ENV`                        | `development`                    | `development` or `production` (must be explicit) |
| `DB_HOST`, `DB_PORT`, `DB_USER`  | `localhost`, `5432`, `postgres`  | PostgreSQL connection                          |
| `DB_PASSWORD`, `DB_NAME`         | `postgres`, `simtas_filkom`      | PostgreSQL credentials / database name         |
| `JWT_SECRET`, `JWT_EXPIRY`       | — / `24h`                        | JWT signing secret + access-token lifetime     |
| `STORAGE_PROVIDER`               | `local`                          | `local` (dev) / `supabase` / `s3`              |
| `CORS_ALLOWED_ORIGINS`           | `http://localhost:3000`          | Frontend origin allowed for CORS               |
| `RESEND_API_KEY`, `EMAIL_FROM`   | —                                | Email sending (Resend) — dev mode logs instead |

> **Development** auto-runs migrations on startup and exposes Swagger.
> **Production** panics on insecure defaults (weak `JWT_SECRET`, default DB password, localhost CORS) — always set real values.

---

## 2. Start the Database

**Option A — Docker (recommended):**

```bash
docker compose up -d postgres
```

From the repo root this starts a PostgreSQL container and creates the `simtas_filkom` database with `postgres` / `postgres` credentials on port `5432` (see `docker-compose.yml` at the repo root).

**Option B — Local PostgreSQL:**

Make sure a server is running on `localhost:5432` and create the database:

```bash
createdb -U postgres simtas_filkom
# or via psql:  CREATE DATABASE simtas_filkom;
```

---

## 3. Run Migrations

With the database up, apply the SQL migrations:

```bash
go run ./cmd/server -migrate
```

Expected output: `database migrations OK — version 18 (dirty=false)`.

> In `development` mode, migrations also run automatically whenever the server starts.

---

## 4. (Optional) Seed the Database

To seed roles, the initial academic year, and default admin/kaprodi accounts:

```bash
go run ./cmd/server -seed
```

This creates the following default dev accounts (`must_change_password=true`):

| Role    | Email                        | Password        |
| ------- | ---------------------------- | --------------- |
| Admin   | `admin@filkom.unida.ac.id`   | `Admin@2027!`   |
| Kaprodi | `kaprodi@filkom.unida.ac.id` | `Kaprodi@2027!` |

These accounts must change their password on first login.

> The separate `cmd/seed` command (`go run ./cmd/seed`) targets **production**: it generates a strong random password per account and is idempotent. Use `go run ./cmd/seed --dev` for an easier, rememberable dev password.

---

## 5. Start the Server

```bash
go run ./cmd/server
# or, with a production build:
go build -o simtas-api ./cmd/server && ./simtas-api
```

The server listens on port **`8080`** (configurable via `APP_PORT`). A convenience Make target is also available:

```bash
make run
```

---

## 6. Verify It Is Running

```bash
# Health check (public)
curl http://localhost:8080/api/v1/health
# → {"success":true,"data":{"database":"ok","status":"ok",...}}

# OpenAPI / Swagger UI (development only)
open http://localhost:8080/swagger/index.html

# Metrics
curl http://localhost:8080/metrics
```

Example login:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@filkom.unida.ac.id","password":"Admin@2027!"}'
```

---

## Makefile Targets

| Target                     | Description                                        |
| -------------------------- | -------------------------------------------------- |
| `make build`               | Compile all packages (`go build ./...`)                                                                          |
| `make run`                 | Run the server (`go run ./cmd/server`)                                                                           |
| `make vet`                 | Run `go vet ./...`                                                                                                |
| `make test`                | Run all unit tests                                                                                                |
| `make test-unit`           | Unit tests for `pkg/...` and `internal/usecase/...` (verbose)                                                     |
| `make test-integration`    | Integration tests (build tag `integration`)                                                                       |
| `make test-coverage`       | Generate a coverage report (`coverage.html`)                                                                      |
| `make coverage-check`      | Enforce ≥80% coverage on the usecase layer                                                                        |
| `make docs` / `make swagger` | Regenerate Swagger docs from annotations                                                                         |

---

## Running Tests

```bash
make test                 # unit tests
make test-integration     # integration tests (require a real DB)
```

---

## Production Notes

- Always set a strong, unique `JWT_SECRET` (≥32 chars) and change `DB_PASSWORD`.
- Set `APP_ENV=production` with proper `CORS_ALLOWED_ORIGINS`.
- Use `STORAGE_PROVIDER=supabase` or `s3` with real credentials; local disk is for development only.
- Configure `TRUSTED_PROXIES` when running behind a reverse proxy so per-IP rate limits see real client IPs.
- The default dev accounts must change their passwords on first login.
