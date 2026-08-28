# Cutover Governance — Phase 1 Parity Gate & Rollback Runbook

> Phase 1 is the walking skeleton. This doc is the gate that Phase N must pass before TS replaces Go. It is the single source of truth for "done".

## 1. Parity Checklist

Each row is a scenario ported from the Go suite (`backend/internal/handler/*_test.go` and `backend/internal/usecase/*_test.go`). `Status`: Pending / Ported / Green (CI green on seeded DB).

| # | Module | Scenario (Go source) | Route / Slice | Expected | Status |
|---|--------|----------------------|---------------|----------|--------|
| 1 | auth | login success (valid email + password) | `POST /api/v1/auth/login` | 200 + `{accessToken, refreshToken, user}` with correct role | Green |
| 2 | auth | login wrong password | `POST /api/v1/auth/login` | 401 `UNAUTHORIZED` | Green |
| 3 | auth | login unknown email (timing-equalized) | `POST /api/v1/auth/login` | 401 no enumeration leak | Green |
| 4 | auth | login locked account (locked_until in future) | `POST /api/v1/auth/login` | 423 `LOCKED` | Green |
| 5 | auth | login lockout after 5 failures (increment + locked_until) | `POST /api/v1/auth/login` | 5th failure → 423, 6th still 423 | Green |
| 6 | auth | login inactive user (`is_active=false`) | `POST /api/v1/auth/login` | 403 | Green |
| 7 | auth | `GET /auth/me` with valid Bearer | `GET /api/v1/auth/me` | 200 `{id,email,fullName,role}` | Green |
| 8 | auth | `GET /auth/me` with bad Bearer / missing | `GET /api/v1/auth/me` | 401 `UNAUTHORIZED` | Green |
| 9 | auth | refresh rotation (present current jti) | `POST /api/v1/auth/refresh` | 200 new pair, old jti revoked | Green |
| 10 | auth | refresh reuse (replay stale jti) | `POST /api/v1/auth/refresh` | 401 `TOKEN_REUSE` + family revoked | Green |
| 11 | auth | logout (revoke family) | `POST /api/v1/auth/logout` | 204, subsequent refresh → 401 | Green |
| 12 | auth | health DB up/down | `GET /api/v1/health` | 200 `healthy` / 503 `unreachable` | Green |
| 13 | rbac | RequireRole matrix (route × role) | middleware `RequireRole` | 403 on wrong role, 401 on no token | Green |
| 14 | rate-limit | global 100/min | `*` | 429 after 100 | Green |
| 15 | rate-limit | login 10/min | `POST /api/v1/auth/login` | 429 after 10 | Green |
| 16 | error | envelope shape | all errors | `{error:{code,message}}` + correct status (400/401/403/404/409/423/429) | Green |
| 17 | validation | loginSchema email format, password min 8 | shared | Zod rejects, handler returns 400 | Green |
| 18 | pagination | paginationSchema (Phase N) | shared | `page, limit, sort` reused across modules | Pending |
| 19 | file-upload | S3/local switch (Phase N) | documents | not in Phase 1 | Pending |
| 20 | forgot/reset | password reset flow (Phase N) | auth | not in Phase 1 | Pending |

**Gate:** rows 1-17 must be Green before any Phase N module is merged. Rows 18-20 are deferred but tracked here to prevent 14 divergent implementations.

## 2. Rollback Runbook

### How to fail back to Go
1. Go container is defined in `docker-compose.yml` (`backend` service) and remains deployable on `develop`. It owns the migrations; TS owns zero DDL in Phase 1.
2. Nginx / reverse proxy routes `/api/v1/auth/*` via flag `USE_TS_AUTH` (env: `true` → TS, `false` → Go). The flag is read at Nginx reload, no redeploy needed.
3. To rollback: `USE_TS_AUTH=false && nginx -s reload` or revert the `docker-compose.override.yml` that points the auth upstream to `http://api-ts:3001`. No DB migration to revert because Phase 1 is schema-read-only. Traffic flips in <5s.

### Data safety
- Both backends write the same Postgres (`simtas` / `simtas_filkom`). No `DROP` / `TRUNCATE` in Phase 1; all writes respect FKs.
- Refresh-token families are the only new table touched; both implementations use `family_id + jti` semantics, so a TS-issued family is readable by Go and vice versa (UUIDs are DB-generated, no codegen mismatch).
- A rollback does not strand users: access tokens remain valid until `exp` (15 min), refresh tokens remain valid in their families.

### Verification after rollback
- `curl /api/v1/health` → 200, `db:healthy`.
- `POST /api/v1/auth/login` with seeded user → 200.
- Monitor `audit_logs` for gapless `USER_LOGIN` sequence (no missing log due to backend switch).

## 3. Cookie / Refresh Cross-Origin Decision

**Phase 1 (current):** refresh token via JSON body (`POST /api/v1/auth/refresh {refreshToken}`; `POST /api/v1/auth/logout {refreshToken}`). SPA keeps `refreshToken` in memory + `localStorage` with silent refresh on 401. This is the simplest that proves rotation without cookie plumbing.

**Phase N (required before production cutover):** move to `HttpOnly` cookie.

```
Set-Cookie: simtas_refresh_token=<refreshJwt>; Path=/; HttpOnly; Secure; SameSite=None; Max-Age=7776000
```

- API sets the cookie on `login` and `refresh`, clears on `logout`. Web no longer sends `refreshToken` in body; `fetch` uses `credentials: "include"`.
- For cross-origin (web on `https://sims.filkom.ac.id`, api on `https://api.filkom.ac.id`) the cookie must be `SameSite=None; Secure` — requires HTTPS. In dev (HTTP) use `SameSite=Lax; Secure=false`.
- The `hc` client must be constructed with `fetch: (url, init) => fetch(url, { ...init, credentials: "include" })`.

This decision is recorded here and must not be re-debated per feature; implement the cookie switch as a single ADR + one PR that touches `auth.ts` (cookie read/write) and `auth.store.ts` (remove localStorage).

## 4. CI Gate (enforced on PR)

```
bunx tsc --noEmit -p tsconfig.base.json
bunx tsc --noEmit -p apps/api/tsconfig.json
./apps/web/node_modules/.bin/svelte-check
DATABASE_URL=... bun test apps/api packages/shared
bunx drizzle-kit check --config packages/db/drizzle.config.ts
```

All must be green before merge to `develop`.

**Status (Phase 1):** tsc (base + api) ✅, svelte-check ✅, `bun test` (17 passing) ✅.
`drizzle-kit check` is **deferred**: the installed `drizzle-kit@0.28`/`drizzle-orm@0.36` pair refuses to run `check`
("Please install latest version of drizzle-orm"), and the schema is a **hand-maintained mirror** of all live
Postgres tables (see ADR-002) rather than a `pull` artifact. The mirror is kept in sync by code review against
Go migrations; full `drizzle-kit` introspection in Phase N remains optional. Until then this gate item is a
known, documented exception.

## 5. Wire-Contract Parity Decision (Ticket 3)

**Decision: the TypeScript API keeps the new TS wire contract as its target contract.** We do **not** port the
Go wire contract verbatim (snake_case JSON keys, `{success,...}` envelope, lowercase role names, HttpOnly-cookie
refresh transport). Rationale: the TS rewrite intentionally normalized payload fields to camelCase, uses a
uniform `{error:{code,message}}` envelope, and Phase 1 proves the slice before tackling cookie plumbing. The
parity rule therefore means **behavior parity (status codes, lockout, rotation/reuse, RBAC) on top of the TS
contract**, not byte-for-byte JSON parity.

### Deliberate deviations from Go (recorded, not bugs)

| Concern | Go behavior | TS behavior | Status |
|---|---|---|---|
| Success envelope | `{success, data, meta}` | flat JSON (e.g. `{accessToken,user}`) / `{data}` for lists | intentional |
| Error envelope | `{success:false, message}` | `{error:{code,message}}` | intentional |
| Field casing | snake_case (`full_name`, `access_token`) | camelCase (`fullName`, `accessToken`) | intentional |
| Role names in claims/responses | lowercase (`admin_fakultas`) | uppercase (`ADMIN_FAKULTAS`) | intentional |
| Lockout HTTP status | `403` for `ErrAccountLocked` | `423 LOCKED` | intentional (TS uses 423 so clients can tell lockout from RBAC 403) |
| Lock duration | 30 minutes | 15 minutes | **divergence — tracked** (plan default; align with Go if product wants 30) |
| Refresh transport | HttpOnly cookie | JSON body (Phase 1) → cookie in Phase N | intentional, Phase-N item |
| Access JWT | `sub`, role, `email`, `jti`, `token_version` | `sub`, role, `tokenVersion`, `jti` | ✔ jti added (Ticket 4) for logout blacklist |
| Access-token revocation on logout | blacklist jti | blacklist access jti in `token_blacklist` (Ticket 4) | ✔ parity |
| Session version check | reject on token_version mismatch | reject on mismatch when user row exists (Ticket 4) | ✔ parity (synthetic-token tests fall through) |

### Authentication security improvements shipped (Ticket 4)
- Access tokens now carry a `jti` so they can be revoked individually.
- `Authenticate()` rejects a blacklisted access token on **every** protected route (not just logout).
- `Authenticate()` enforces `token_version` (session bump) and active-account check when the user row exists.
- `POST /auth/logout` blacklists the presented access token's jti **and** revokes the refresh family.

These match the Go middleware's intent (blacklist + session-version) on the TS contract. The one non-parity
fall-through (synthetic non-UUID user ids in tests) is documented inline in `middleware/auth.ts` and guarded so
it cannot 500.

## 6. Links

- Walking skeleton plan: `docs/superpowers/plans/2026-07-28-ts-rewrite-phase1-walking-skeleton.md`
- ADR-002 (schema manual mirror): `docs/adr/ADR-002-ts-schema-manual-mirror.md`
- ROADMAP: `docs/ROADMAP.md` → add row: `Phase 1 — Walking Skeleton + Auth Slice (Tasks 1-9) — status: Green (see this doc)`
