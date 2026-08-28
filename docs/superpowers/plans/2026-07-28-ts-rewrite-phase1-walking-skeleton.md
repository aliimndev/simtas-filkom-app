# TypeScript Rewrite — Phase 1: Walking Skeleton + Auth Slice Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

## Agent Execution Contract (READ FIRST)

An AI agent executing this plan MUST obey the following rules. They are not optional.

1. **One task at a time.** Complete Task N fully (all steps + verification green) before starting Task N+1. Do not batch-edit unrelated files.
2. **Track with checkboxes.** Tick `- [ ]` → `- [x]` as you finish each step. A step is only done when its **Verification** command passes with the stated output.
3. **One conventional commit per task.** Use the exact `git commit -m "..."` message provided at the end of each task. Do not squash tasks. Do not commit `backend/` (Go) — it is out of scope for Phase 1.
4. **Run `bun install` after every change to any `package.json`.** Workspace links (`workspace:*`) are resolved by Bun; missing install = broken imports.
5. **When a step is ambiguous, STOP and ask.** Do not invent secrets, env values, or schema column names. The DB contract is generated, not assumed — always read `packages/db/src/schema.ts` (produced in Task 2) before writing any query.
6. **Package namespace is `@sims/*` everywhere.** `@sims/api`, `@sims/web`, `@sims/db`, `@sims/shared`. (The original draft used a mix of `@baz/*` and `@sims/web`; that was a bug. This version is normalized to `@sims/*`.)
7. **Column naming caveat.** `drizzle-kit pull` preserves the database's snake_case column names as the Drizzle property keys (e.g. `user_id`, `family_id`, `token_jti`, `expires_at`, `revoked`, `password_hash`). The code snippets below use readable names; **always substitute the exact property names from `packages/db/src/schema.ts`** when you implement.
8. **Do not modify `backend/` (Go) in Phase 1.** The Go backend stays deployable on `develop`. You may READ Go source (handlers, migrations, tests) to port behavior 1:1, but never write to it.

### Agent Execution Loop (Structured Flow)

Every task is executed through the same 6-phase loop. Do not skip phases. If a phase fails, loop back — never advance.

```
┌─────────────────────────────────────────────────────────────────┐
│                    AGENT LOOP (per Task)                         │
│                                                                  │
│  ┌──────┐    ┌──────┐    ┌─────────┐    ┌────────┐    ┌──────┐  │
│  │ READ │───>│ PLAN │───>│ EXECUTE │───>│ VERIFY │───>│COMMIT│  │
│  └──┬───┘    └──────┘    └────┬────┘    └───┬────┘    └──┬───┘  │
│     │                  fail   │         fail│            │      │
│     │              ┌─────────┘         ┌────┘            │      │
│     │              ▼                   ▼                 │      │
│     │        ┌──────────┐        ┌─────────┐            │      │
│     └───────>│ REFLECT  │<───────│ REFLECT │            │      │
│              │ (fix)    │        │ (fix)   │            │      │
│              └────┬─────┘        └────┬────┘            │      │
│                   └──────> retry ─────┘                 │      │
│                                                          ▼      │
│                                                   ┌──────────┐  │
│                                                   │ NEXT TASK│  │
│                                                   └──────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

**Phase definitions:**

| Phase | What the agent does | Exit criteria | On failure |
|-------|---------------------|---------------|------------|
| **1. READ** | Read the task's Goal, Produces, Done-when, all steps, and referenced files (`schema.ts`, Go handlers/tests, existing `app.ts`). Never assume schema column names. | Can list every file to create/modify and every verification command for the task. | Re-read; ask human if spec is contradictory. |
| **2. PLAN** | Micro-plan: order the steps, identify dependencies (e.g. Task 5 needs `schema.ts` from Task 2), note `ponytail:` simplifications and their ceilings. Write a 3-5 line plan in the subagent's scratchpad — not in the repo. | Plan fits in one screen, no invented env vars. | Revise plan. |
| **3. EXECUTE** | Create/modify files exactly as spec'd. After any `package.json` change, run `bun install`. Keep diff minimal — shortest working change wins (`ponytail: full`). | Files exist, `bunx tsc --noEmit` has no new errors in touched package. | Fix, stay in EXECUTE until typecheck passes. |
| **4. VERIFY** | Run the task's **Done when** command(s) verbatim. Capture output. Example: `bun test packages/shared`, `cd apps/api && bun test`, `bun run typecheck`. | Output matches the expected result stated in the task (e.g. `2 passing`, `exit 0`). | Go to REFLECT. |
| **5. REFLECT** | On verification failure: read the error, grep the failing function's callers, fix root cause (not symptom) in the shared function where all callers route through. Re-run VERIFY. Max 3 retries then ask human. | Verification green. | Ask human, do not commit red. |
| **6. COMMIT** | `git add <only task files>` + `git commit -m "<exact message>"` + tick checkboxes in this plan file if tracking locally. One commit per task, conventional commits. | `git log --oneline -1` shows the expected message; `git status` clean for task files. | Amend only if commit message was wrong; never force-push. |

**Loop invariants (hold for every task):**
- **Idempotent reruns:** re-running a completed task's VERIFY must still pass without re-creating files.
- **Evidence before claim:** never say "done" without pasting the verification command output.
- **Subagent mapping:** when using `superpowers:subagent-driven-development`, spawn one subagent per task, passing the task's Goal + Steps + Done-when as the prompt. The main agent only runs VERIFY and COMMIT reviews between subagents.
- **Checkpoint:** after COMMIT, the main agent reads the next task's Goal and repeats the loop. No parallel tasks.

**Quick checklist (copy into each subagent prompt):**
```
- [ ] READ task + referenced schema/Go source
- [ ] PLAN micro-steps
- [ ] EXECUTE file changes + bun install if needed
- [ ] VERIFY via task's Done-when command(s) — paste output
- [ ] REFLECT if red — fix root cause, re-verify (max 3x)
- [ ] COMMIT with exact message
```

**Goal:** Stand up the full TypeScript fullstack pipeline (Bun monorepo → Hono API → Drizzle Postgres → SvelteKit SPA) with a working end-to-end **auth + user + RBAC** vertical slice, proving every architectural assumption before porting the remaining 14 domain modules.

**Architecture:** Monorepo split into `apps/api` (Hono + Drizzle), `apps/web` (SvelteKit, SSR disabled / SPA mode), `packages/db` (Drizzle schema + connection), and `packages/shared` (Zod schemas + types, consumed by both apps). The API is the single source of truth for the domain; the web app is a pure presentation client calling the API over HTTP via Hono's `hc` RPC client. JWT access + refresh-token **rotation with family revoke-on-reuse** is ported 1:1 from the existing Go backend.

**Tech Stack:** Bun (runtime + package manager + test runner), Hono, Drizzle ORM + `drizzle-kit`, Zod, jose (JWT), bcryptjs (password hashing, compatible with Go's bcrypt hashes), PostgreSQL (existing, reused schema via `drizzle-kit pull`), SvelteKit (SPA mode), TypeScript 5.

## System Design — Senior Review (Added for Fullstack Migration)

This section fills the gaps the original skeleton left as implicit. Read it before Task 1.

### C4 / Component Overview

```
Level 1 (Context):  Browser (SvelteKit SPA) ──HTTPS──> Hono API (Bun) ──postgres.js──> PostgreSQL (existing)
                                                     │
                                                     └─> reads Go migrations (source of truth, read-only in Phase 1)

Level 2 (Containers):
  apps/web  : SvelteKit adapter-static, ssr:false, fallback: index.html. Served via static hosting (Cloudflare Pages / Nginx). Env: VITE_API_ORIGIN.
  apps/api  : Bun.serve + Hono. Single container, stateless, horizontal scale readiness via ponytail in-memory limiter (see Task 7).
  packages/db: Drizzle schema (generated) + postgres.js pool (max 10). No DDL in Phase 1.
  packages/shared: Zod schemas, role constants, TokenPair. Consumed by both apps — single source of truth.

Level 3 (API internals):  Hono app factory (createApp) → middleware chain (rateLimit → CORS → auth → rbac) → routes (health, auth, me) → services (token, password) → db
```

### Deployment Topology (Phase 1 Decision)

- **API:** one Docker image `FROM oven/bun:1` running `bun run --watch` in dev, `bun src/index.ts` in prod. Port `3001`. Behind Nginx / reverse proxy that terminates TLS and sets `x-forwarded-for` (trusted proxy).
- **Web:** `adapter-static` SPA built via `vite build`, deployed to static host. Deep links work via `fallback: index.html`. No SSR server needed.
- **Env contract:** `DATABASE_URL`, `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`, `CORS_ORIGIN` (comma-separated allowlist in prod — see Security), `ACCESS_TTL_SEC` (default 900), `PORT`. Missing secrets throw at boot (`required()` in config.ts) except `NODE_ENV=test`.
- **Strangler fig:** Go stays on `develop`, TS on `ts/*`. Nginx routes `/api/v1/auth/*` to TS first with flag `USE_TS_AUTH`; rollback = flip flag, zero schema change.

### Data Flow — Auth Slice (the only domain in Phase 1)

```
login:  SPA POST /auth/login {email,password} → validate loginSchema → SELECT users WHERE email → check locked_until → bcrypt compare
        → on fail: increment login_attempt_count, lock after 5 for 15m → 401/423
        → on success: reset counter, set last_login_at, issueTokens() → INSERT refresh_token_families (family_id, jti, revoked:false) → return {accessToken, refreshToken, user}

me:     GET /auth/me Authorization: Bearer <access> → Authenticate() verifyJwt(accessSecret) → c.set(user) → SELECT users WHERE id → return {id,email,fullName,role,mustChangePassword}

refresh: POST /auth/refresh {refreshToken} → verifyJwt(refreshSecret) → SELECT refresh_token_families WHERE family_id AND jti AND revoked=false
        → if miss → revokeRefreshFamily(familyId) (revoke all jtis in family) → throw TOKEN_REUSE → 401 (forces re-login)
        → if hit → UPDATE revoked=true for old jti → INSERT new jti → sign new pair → return {accessToken,refreshToken}

logout: POST /auth/logout {refreshToken} → verifyJwt(refreshToken) → revokeRefreshFamily(familyId) → 204
```

JWT claims parity with Go: `{sub: userId, role, tokenVersion, iat, exp}` for access; refresh adds `{familyId, jti}`. `aud/iss` optional, keep absent for parity unless Go sets them — check `backend/internal/auth`.

### Security Additions (beyond the skeleton)

- **Password:** `bcryptjs` cost 12 for new hashes; existing Go hashes (cost 10-12) verified via `bcrypt.compare` — wire compatible.
- **Refresh storage Phase 1:** via request body (SPA keeps refresh in memory + localStorage with silent refresh on 401). Phase N ADR: move to `HttpOnly; Secure; SameSite=Lax` cookie; then refresh endpoint reads cookie, not body. Record in Task 9.
- **CORS:** use `hono/cors` with allowlist: `origin: (o)=> allowlist.includes(o) ? o : ""`, `credentials:true`, `allowHeaders: ["Content-Type","Authorization"]`. Manual header code in draft replaced by this middleware (ponytail: native).
- **Rate limit key:** `x-forwarded-for` is trusted only behind proxy; add `app.set('trust proxy',1)` equivalent or read `cf-connecting-ip` fallback. In-memory limiter is single-instance only (ponytail note in Task 7) — acceptable for ~100 users, Redis when scaling.
- **Lockout:** `login_attempt_count` + `locked_until` copied 1:1 from Go; 423 not 429 so clients can distinguish lockout vs rate limit.

### Observability & Quality Gates

- **Health:** `GET /api/v1/health` must do `SELECT 1` against Postgres and return `{status:"ok", db:"healthy"}` or `{status:"error", db:"unreachable"}` with 503 — not hardcoded.
- **Logging:** `hono-pino` or `hono/logger` with request-id (`X-Request-Id`) propagated to error envelope. No `console.log` in prod path except boot message.
- **CI gate (Phase 1):** `bunx tsc --noEmit`, `drizzle-kit check`, `svelte-check`, `bun test` (api + shared) must all pass on PR. Integration tests run against ephemeral Postgres via `docker compose -f docker-compose.test.yml up -d`.

### API Contract

- **Error envelope (uniform):** `{error:{code:string, message:string}}` with status map `VALIDATION:400, UNAUTHORIZED:401, FORBIDDEN:403, NOT_FOUND:404, CONFLICT:409, LOCKED:423, RATE_LIMIT:429`. Implemented via `app.onError` + `throwError(c,code,msg,status)` from Task 7, used from Task 4 onward.
- **Pagination (for Phase N):** add `paginationSchema` to `packages/shared` now (`page, limit, sort`) so later modules reuse it — prevents 14 divergent implementations.
- **OpenAPI (optional, 1 line):** `hono/openapi` + Scalar UI at `GET /doc` for manual QA during migration.

### Frontend Architecture

- Auth state: `writable<{accessToken:string|null, user:User|null}>` in memory; refresh token in `localStorage` for silent refresh on 401. `+layout.ts` (`ssr:false`) guards unauthenticated routes via `goto("/login")`.
- `hc<AppType>` typing: `AppType` re-exported from `@sims/api/src/index.ts`; client exposes `api.api.v1.auth.login.$post` — verify with `svelte-check`.

### Migration Order (Post-Phase 1)

Phase 1 proves the slice. Remaining 14 modules port in dependency order: `users → roles → academic_years → theses → title_change_requests → consultation_logs → seminars → thesis_defenses → documents → thesis_archives → notifications → email_logs → audit_logs → dashboard`. Each follows the same 6-phase loop and must add its row to the parity checklist (Task 9).

## Global Constraints

- **Go backend untouched:** the existing Go backend (`backend/`) stays fully deployable and on `develop` until Phase N cutover; this plan introduces the TS workspace alongside it. Do **not** modify `backend/` in Phase 1 (reading is allowed for porting).
- **Zero schema drift:** `packages/db/src/schema.ts` is generated from the live Postgres DB via `drizzle-kit pull`. Never hand-edit introspected schema to "fix" design; design fixes are deferred.
- **Zero data loss:** no `DROP`, `TRUNCATE`, or destructive migration. Respect all existing FKs/constraints.
- **Auth parity non-negotiable:** JWT access + refresh with per-family rotation and reuse detection (revoke family) must behave identically to Go. Password hashing stays **bcrypt** — existing hashes in `users.password_hash` must keep logging in (use `bcryptjs`, which is wire-compatible with Go's `golang.org/x/crypto/bcrypt`).
- **Route prefixes unchanged:** all routes under `/api/v1/*` exactly as the Go router (`POST /api/v1/auth/login`, `/refresh`, `/logout`, `GET /api/v1/auth/me`, etc.).
- **Test-parity gate:** a Hono/Bun integration test that hits the live DB must pass before any further module is ported. TS cutover only allowed when the ported integration suite is 100% green.
- **One runtime, one language:** everything runs on Bun and is TypeScript. No separate Node runtime.
- **Frequent commits:** one conventional-commit message per task.

---

## File Structure (Phase 1)

```
.
├── package.json                  # Bun workspace root (private)
├── bunfig.toml
├── tsconfig.base.json
├── apps/
│   ├── api/
│   │   ├── package.json         # @sims/api
│   │   ├── tsconfig.json
│   │   ├── src/
│   │   │   ├── index.ts         # Bun.serve entry
│   │   │   ├── app.ts           # Hono app factory (importable for tests)
│   │   │   ├── config.ts        # typed env config
│   │   │   ├── db.ts            # drizzle client singleton (wraps @sims/db)
│   │   │   ├── middleware/      # auth.ts, rbac.ts, rateLimit.ts, error.ts
│   │   │   ├── routes/          # health.ts, auth.ts, me.ts
│   │   │   └── services/        # token.ts (JWT+rotation), password.ts
│   │   └── test/                # health, token, auth.integration, helpers
│   └── web/
│       ├── package.json         # @sims/web (SvelteKit)
│       ├── svelte.config.js     # adapter-static, ssr:false
│       ├── vite.config.ts
│       ├── tsconfig.json
│       └── src/
│           ├── lib/             # api.ts (hc client), auth.store.ts
│           ├── routes/          # +layout.svelte, +layout.ts, login/+page.svelte, +page.svelte
│           └── app.html         # SvelteKit HTML shell
├── packages/
│   ├── db/
│   │   ├── package.json         # @sims/db
│   │   ├── drizzle.config.ts
│   │   └── src/
│   │       ├── index.ts         # createDb(), schema re-export, Db type
│   │       └── schema.ts        # GENERATED via pull (do not edit)
│   └── shared/
│       ├── package.json         # @sims/shared
│       └── src/                 # auth.ts, roles.ts, index.ts
└── scripts/
    └── dev.sh                   # run api + web concurrently
```

---

## Task 1: Bootstrap the Bun Workspace Monorepo

**Goal:** Establish root scaffolding every later task depends on: workspace manifests, shared TS config, single lockfile.

**Produces:** Bun resolves workspace packages `@sims/api`, `@sims/web`, `@sims/db`, `@sims/shared` by name.

**Done when:** `bun install` produces a single `bun.lock` at root and `bunx tsc --noEmit -p tsconfig.base.json` exits 0.

- [ ] **Step 1: Create root `package.json`**

```jsonc
{
  "name": "simtas-ts",
  "private": true,
  "type": "module",
  "workspaces": ["apps/*", "packages/*"],
  "scripts": { "dev": "bash scripts/dev.sh" },
  "devDependencies": { "typescript": "^5.5.0", "@types/bun": "latest" }
}
```

- [ ] **Step 2: Create `bunfig.toml`**

```toml
[install]
# workspace tooling defaults; install scoping reserved for later
```

- [ ] **Step 3: Create `tsconfig.base.json`**

```jsonc
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "moduleResolution": "Bundler",
    "lib": ["ES2022"],
    "strict": true,
    "noEmit": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "types": ["bun"],
    "forceConsistentCasingInFileNames": true
  }
}
```

- [ ] **Step 4: Append to `.gitignore`** (keep existing entries)

```bash
# TypeScript rewrite (Phase 1)
node_modules
*.tsbuildinfo
apps/api/.env
apps/web/.env
```

- [ ] **Step 5: Verify & commit**

Run: `bun install` → expect a single `bun.lock` at root; `bunx tsc --noEmit -p tsconfig.base.json` → exit 0.

```bash
git add package.json bunfig.toml tsconfig.base.json .gitignore bun.lock
git commit -m "chore(ts): bootstrap Bun workspace monorepo (Phase 1)"
```

---

## Task 2 — `packages/db`: Introspect Postgres schema

**Goal:** Lock the database contract by introspecting the same dev Postgres that holds the Go migrations, so the generated schema is an exact mirror.

**Produces:** `schema` (all tables/columns), `createDb(url): Db`, `type Db`.

**Done when:** `bunx drizzle-kit check` reports no errors and `bun test packages/db` exits 0 (0 tests ok).

- [ ] **Step 1: Create `packages/db/package.json`**

```jsonc
{
  "name": "@sims/db",
  "version": "0.1.0",
  "type": "module",
  "exports": { ".": "./src/index.ts" },
  "scripts": { "generate": "drizzle-kit pull", "check": "drizzle-kit check" },
  "dependencies": { "drizzle-orm": "^0.36.0", "postgres": "^3.4.4" },
  "devDependencies": { "drizzle-kit": "^0.28.0" }
}
```

- [ ] **Step 2: Create `packages/db/drizzle.config.ts`**

```ts
import { defineConfig } from "drizzle-kit";
export default defineConfig({
  dialect: "postgresql",
  schema: "./src/schema.ts",
  out: "./drizzle",
  dbCredentials: { url: process.env.DATABASE_URL ?? "postgres://simtas:simtas@localhost:5432/simtas" },
});
```

- [ ] **Step 3: Start the dev DB and generate schema (do not hand-edit result)**

Start Postgres first if not running (see root `docker-compose.yml`): `docker compose up -d db`.
Point `DATABASE_URL` at the same Postgres the Go backend uses, then:

```bash
cd packages/db
DATABASE_URL="postgres://simtas:simtas@localhost:5432/simtas" bunx drizzle-kit pull --config drizzle.config.ts
```

This writes `packages/db/src/schema.ts`. **Read it.** Note the exact property names it generated (snake_case). You will use these property names in Tasks 5 and 6. Expected tables include at least: `users`, `roles`, `refresh_token_families`, plus the 13 other domain tables.

- [ ] **Step 4: Create `packages/db/src/index.ts`**

```ts
import { drizzle, type PostgresJsDatabase } from "drizzle-orm/postgres-js";
import postgres from "postgres";
export * as schema from "./schema";

export function createDb(url: string) {
  return drizzle(postgres(url, { max: 10 }), { schema });
}
export type Db = PostgresJsDatabase<typeof schema>;
```

- [ ] **Step 5: Verify & commit**

Run: `bunx drizzle-kit check` (no errors) and `bun test packages/db` (0 tests, exit 0).

```bash
git add packages/db bun.lock
git commit -m "feat(db): introspect existing Postgres schema via drizzle-kit pull (Phase 1)"
```

> **Agent note:** Every later task that queries the DB imports the table objects via `import { schema } from "@sims/db"` (this gives you `schema.users`, `schema.refreshTokenFamilies`, etc.). There is no `@sims/db/schema` subpath export — use the namespace re-export above.

---

## Task 3 — `packages/shared`: typed auth contract

**Goal:** Single source of truth for request/response shapes shared by API and web. Zod is the validator.

**Produces:** `loginSchema`, `LoginInput`, `loginSuccessSchema`, `LoginSuccessPayload`, `tokenPairSchema`, `TokenPair`, role constants `ADMIN_FAKULTAS`, `ROLE_KAPRODI`, `ROLE_DOSEN_PEMBIMBING`, `ROLE_MAHASISWA`, `RoleName`, `isRoleName()`.

**Done when:** `bun test packages/shared` shows 2 passing.

- [ ] **Step 1: Create `packages/shared/package.json`**

```jsonc
{
  "name": "@sims/shared",
  "version": "0.1.0",
  "type": "module",
  "exports": { ".": "./src/index.ts" },
  "scripts": { "test": "bun test" },
  "dependencies": { "zod": "^3.23.8" }
}
```

- [ ] **Step 2: Create `packages/shared/src/roles.ts`**

```ts
export const ADMIN_FAKULTAS = "ADMIN_FAKULTAS";
export const ROLE_KAPRODI = "KAPRODI";
export const ROLE_DOSEN_PEMBIMBING = "DOSEN_PEMBIMBING";
export const ROLE_MAHASISWA = "MAHASISWA";
export type RoleName = typeof ADMIN_FAKULTAS | typeof ROLE_KAPRODI | typeof ROLE_DOSEN_PEMBIMBING | typeof ROLE_MAHASISWA;
export function isRoleName(v: unknown): v is RoleName {
  return [ADMIN_FAKULTAS, ROLE_KAPRODI, ROLE_DOSEN_PEMBIMBING, ROLE_MAHASISWA].includes(v as RoleName);
}
```

- [ ] **Step 3: Create `packages/shared/src/auth.ts`**

```ts
import { z } from "zod";

export const loginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8),
});
export type LoginInput = z.infer<typeof loginSchema>;

export const loginSuccessSchema = z.object({
  accessToken: z.string(),
  refreshToken: z.string(),
  user: z.object({
    id: z.string().uuid(),
    email: z.string().email(),
    fullName: z.string(),
    role: z.string(),
    mustChangePassword: z.boolean(),
  }),
});
export type LoginSuccessPayload = z.infer<typeof loginSuccessSchema>;

// Reused by the /refresh route so the pair shape stays consistent.
export const tokenPairSchema = z.object({
  accessToken: z.string(),
  refreshToken: z.string(),
});
export type TokenPair = z.infer<typeof tokenPairSchema>;
```

- [ ] **Step 4: Create `packages/shared/src/index.ts`**

```ts
export * from "./auth";
export * from "./roles";
```

- [ ] **Step 5: Create test `packages/shared/test/auth.test.ts`**

```ts
import { describe, test, expect } from "bun:test";
import { loginSchema, isRoleName, ROLE_KAPRODI } from "@sims/shared";

describe("auth contract", () => {
  test("accepts valid creds, rejects short password", () => {
    expect(loginSchema.safeParse({ email: "a@b.c", password: "longenough" }).success).toBe(true);
    expect(loginSchema.safeParse({ email: "a@b.c", password: "short" }).success).toBe(false);
  });
  test("role constants are valid", () => {
    expect(isRoleName(ROLE_KAPRODI)).toBe(true);
    expect(isRoleName("VISITOR")).toBe(false);
  });
});
```

- [ ] **Step 6: Run & verify pass** — `bun test packages/shared` → 2 passing.
- [ ] **Step 7: Commit**

```bash
git add packages/shared bun.lock
git commit -m "feat(shared): define typed auth contract & roles (Phase 1)"
```

---

## Task 4 — Hono API app + health + error shape

**Goal:** Stand up the Hono app factory, typed config, DB singleton, CORS, and the health endpoint; the app must be importable in tests.

**Produces:** `createApp(cfg): Hono`, `AppType`, `loadConfig(env?): Config`, `getDb(url): Db`, `ApiErrorBody` type.

**Done when:** `cd apps/api && bun test` → 2 passing (health + 404). Typecheck passes.

- [ ] **Step 1: Create `apps/api/package.json`**

```jsonc
{
  "name": "@sims/api",
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "bun --watch src/index.ts",
    "test": "bun test test/",
    "typecheck": "bunx tsc --noEmit"
  },
  "dependencies": {
    "hono": "^4.5.0",
    "postgres": "^3.4.4",
    "jose": "^5.9.0",
    "bcryptjs": "^2.4.3",
    "@sims/db": "workspace:*",
    "@sims/shared": "workspace:*"
  },
  "devDependencies": { "@types/bun": "latest", "typescript": "^5.5.0", "@types/bcryptjs": "^2.4.6" }
}
```

- [ ] **Step 2: Create `apps/api/tsconfig.json`**

```jsonc
{ "extends": "../../tsconfig.base.json", "compilerOptions": { "types": ["bun"] }, "include": ["src", "test"] }
```

- [ ] **Step 3: Create `apps/api/src/config.ts`**

```ts
export type Config = {
  databaseUrl: string;
  jwtAccessSecret: string;
  jwtRefreshSecret: string;
  accessTtlSec: number;
  refreshTtlSec: number;
  corsOrigin: string;
  port: number;
  isTest: boolean;
};

export function loadConfig(env: Record<string, string | undefined> = process.env): Config {
  return {
    databaseUrl: env.DATABASE_URL ?? "postgres://simtas:simtas@localhost:5432/simtas",
    jwtAccessSecret:
      env.JWT_ACCESS_SECRET ?? (env.NODE_ENV === "test" ? "test-access-secret" : required("JWT_ACCESS_SECRET")),
    jwtRefreshSecret:
      env.JWT_REFRESH_SECRET ?? (env.NODE_ENV === "test" ? "test-refresh-secret" : required("JWT_REFRESH_SECRET")),
    accessTtlSec: Number(env.ACCESS_TTL_SEC ?? 15 * 60),
    refreshTtlSec: 90 * 24 * 60 * 60,
    corsOrigin: env.CORS_ORIGIN ?? "http://localhost:5173",
    port: Number(env.PORT ?? 3001),
    isTest: env.NODE_ENV === "test",
  };
}

function required(k: string): string {
  throw new Error(`Missing env var ${k}`);
}
```

- [ ] **Step 4: Create `apps/api/src/db.ts`** (singleton that reuses `@sims/db`)

```ts
import { createDb, type Db } from "@sims/db";

let shared: Db | undefined;

export function getDb(url: string): Db {
  if (!shared) shared = createDb(url);
  return shared;
}
```

- [ ] **Step 5: Create `apps/api/src/app.ts`**

> Ponytail fix: use `hono/cors` (native middleware) instead of hand-rolled headers — fewer lines, correct preflight handling.

```ts
import { Hono } from "hono";
import { cors } from "hono/cors";
import type { Config } from "./config";

export function createApp(cfg: Config) {
  const app = new Hono();

  // CORS via native middleware (ponytail: native). Allowlist supports comma-separated CORS_ORIGIN in prod.
  const allowlist = cfg.corsOrigin.split(",").map((s) => s.trim());
  app.use(
    "*",
    cors({
      origin: (origin) => (allowlist.includes(origin) || allowlist.includes("*") ? origin : allowlist[0]),
      allowHeaders: ["Content-Type", "Authorization"],
      allowMethods: ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"],
      credentials: true,
    }),
  );

  // Health does a live DB probe via SELECT 1 (see System Design — not hardcoded).
  app.route(
    "/api/v1/health",
    new Hono().get("/", async (c) => {
      try {
        const db = getDb(cfg.databaseUrl);
        await db.execute("SELECT 1" as any);
        return c.json({ status: "ok", db: "healthy" });
      } catch {
        return c.json({ status: "error", db: "unreachable" }, 503);
      }
    }),
  );

  return app;
}

export type AppType = ReturnType<typeof createApp>;
```

- [ ] **Step 6: Create `apps/api/src/index.ts`**

```ts
import { loadConfig } from "./config";
import { createApp } from "./app";

const cfg = loadConfig();
Bun.serve({ port: cfg.port, fetch: createApp(cfg).fetch });
console.log(`api listening on http://localhost:${cfg.port}`);
```

- [ ] **Step 7: Create `apps/api/src/middleware/error.ts`** (type stub; full renderer added in Task 7)

```ts
import type { MiddlewareHandler } from "hono";

export interface ApiErrorBody {
  error: { code: string; message: string };
}

// Full onError renderer + throwError() helper are implemented in Task 7.
export const errorStub: MiddlewareHandler = async (_c, next) => {
  await next();
};
```

- [ ] **Step 8: Create `apps/api/test/health.test.ts`**

```ts
import { describe, expect, it } from "bun:test";
import { createApp } from "../src/app";
import { loadConfig } from "../src/config";

const app = createApp(loadConfig({ NODE_ENV: "test" }));

describe("health", () => {
  it("returns 200 status ok", async () => {
    const res = await app.request("/api/v1/health");
    expect(res.status).toBe(200);
    expect(await res.json()).toEqual({ status: "ok", db: "healthy" });
  });
  it("404 on unknown route", async () => {
    expect((await app.request("/api/v1/nope")).status).toBe(404);
  });
});
```

- [ ] **Step 9: Run & verify pass** — `cd apps/api && bun test` → 2 passing. Also `bun run typecheck` → exit 0.
- [ ] **Step 10: Commit**

```bash
git add apps/api bun.lock
git commit -m "feat(api): Hono app entry, CORS, health endpoint (Phase 1)"
```

---

## Task 5 — Token service: JWT issue + refresh-token family rotation

**Goal:** Implement JWT signing/verification and the refresh-token family rotation with revoke-on-reuse, ported 1:1 from the Go `refresh_token_families` CAS logic.

**Requires:** Task 2 completed — you MUST have read `packages/db/src/schema.ts` to know the exact `refresh_token_families` column property names. Below uses readable names; substitute the real ones (e.g. `user_id`, `family_id`, `token_jti`, `expires_at`, `revoked`).

**Produces:**
- `signAccessToken(userId, role, tokenVersion): Promise<string>`
- `signRefreshToken(familyId, jti, userId, role, tokenVersion): Promise<string>`
- `verifyJwt(token, secret): Promise<JwtClaims | null>`
- `issueTokens({ userId, role, tokenVersion }): Promise<TokenPair>` — creates a new family, inserts the first refresh row, returns signed pair
- `rotateRefresh(refreshToken): Promise<TokenPair>` — throws `{ code: "TOKEN_REUSE" }` on stale-token reuse (revokes the whole family)
- `revokeRefreshFamily(familyId): Promise<void>`

**Done when:** `cd apps/api && bun test test/token.test.ts` → passing (pure sign/verify; DB-backed rotation is covered by the Task 6 integration parity test).

- [ ] **Step 1: Write failing-free unit test `apps/api/test/token.test.ts`** (pure signing only — no DB needed)

```ts
import { describe, expect, it } from "bun:test";
import { signAccessToken, verifyJwt } from "../src/services/token";
import { loadConfig } from "../src/config";
import { ROLE_MAHASISWA } from "@sims/shared";

const cfg = loadConfig({ NODE_ENV: "test" });

describe("token service", () => {
  it("issues access JWT with sub + role + tokenVersion claims", async () => {
    const accessToken = await signAccessToken(
      "00000000-0000-0000-0000-000000000001",
      ROLE_MAHASISWA,
      0,
    );
    const claims = await verifyJwt(accessToken, cfg.jwtAccessSecret);
    expect(claims?.role).toBe(ROLE_MAHASISWA);
    expect(claims?.sub).toBe("00000000-0000-0000-0000-000000000001");
    expect(claims?.tokenVersion).toBe(0);
  });

  it("verifyJwt returns null for a bad token", async () => {
    expect(await verifyJwt("not-a-jwt", cfg.jwtAccessSecret)).toBeNull();
  });
});
```

- [ ] **Step 2: Implement `apps/api/src/services/token.ts`**

> Port the rotation semantics from the Go handler/service that manages `refresh_token_families`. The Go logic: on refresh, the presented token must be the *current* unrevoked jti for its family; if a previously-used (revoked) jti is presented, the entire family is revoked and all refresh tokens in it are invalidated (TOKEN_REUSE → force re-login).

```ts
import { SignJWT, jwtVerify, type JWTPayload } from "jose";
import { eq, and } from "drizzle-orm";
import { getDb } from "../db";
import { loadConfig } from "../config";
import { schema } from "@sims/db";

const enc = (s: string) => new TextEncoder().encode(s);

export interface JwtClaims extends JWTPayload {
  sub: string;
  role: string;
  tokenVersion: number;
  familyId?: string;
  jti?: string;
}

export async function signAccessToken(userId: string, role: string, tokenVersion: number): Promise<string> {
  const cfg = loadConfig();
  return new SignJWT({ role, tokenVersion })
    .setProtectedHeader({ alg: "HS256" })
    .setSubject(userId)
    .setIssuedAt()
    .setExpirationTime(Math.floor(Date.now() / 1000) + cfg.accessTtlSec)
    .sign(enc(cfg.jwtAccessSecret));
}

export async function signRefreshToken(
  familyId: string,
  jti: string,
  userId: string,
  role: string,
  tokenVersion: number,
): Promise<string> {
  const cfg = loadConfig();
  return new SignJWT({ familyId, jti, role, tokenVersion })
    .setProtectedHeader({ alg: "HS256" })
    .setSubject(userId)
    .setIssuedAt()
    .setExpirationTime(Math.floor(Date.now() / 1000) + cfg.refreshTtlSec)
    .sign(enc(cfg.jwtRefreshSecret));
}

export async function verifyJwt(token: string, secret: string): Promise<JwtClaims | null> {
  try {
    const { payload } = await jwtVerify(token, enc(secret));
    return payload as JwtClaims;
  } catch {
    return null;
  }
}

// Issue a brand-new refresh family and return the signed pair.
export async function issueTokens(input: {
  userId: string;
  role: string;
  tokenVersion: number;
}): Promise<{ accessToken: string; refreshToken: string }> {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);
  const familyId = crypto.randomUUID();
  const jti = crypto.randomUUID();

  // INSERT the first refresh row for this family (substitute real column names).
  await db.insert(schema.refreshTokenFamilies).values({
    userId: input.userId,
    familyId,
    tokenJti: jti,
    expiresAt: new Date(Date.now() + cfg.refreshTtlSec * 1000),
    revoked: false,
  });

  const [accessToken, refreshToken] = await Promise.all([
    signAccessToken(input.userId, input.role, input.tokenVersion),
    signRefreshToken(familyId, jti, input.userId, input.role, input.tokenVersion),
  ]);
  return { accessToken, refreshToken };
}

// Rotate: the presented refresh token must be the current unrevoked jti.
// Reuse of an old jti => revoke the whole family (TOKEN_REUSE).
export async function rotateRefresh(refreshToken: string): Promise<{ accessToken: string; refreshToken: string }> {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);

  const claims = await verifyJwt(refreshToken, cfg.jwtRefreshSecret);
  if (!claims?.sub || !claims.familyId || !claims.jti) {
    throw { code: "INVALID_REFRESH" };
  }

  const current = await db
    .select()
    .from(schema.refreshTokenFamilies)
    .where(
      and(
        eq(schema.refreshTokenFamilies.familyId, claims.familyId),
        eq(schema.refreshTokenFamilies.tokenJti, claims.jti),
        eq(schema.refreshTokenFamilies.revoked, false),
      ),
    );

  if (current.length === 0) {
    // The presented jti is not the current valid one => reuse detected.
    await revokeRefreshFamily(claims.familyId);
    throw { code: "TOKEN_REUSE" };
  }

  // Revoke the used jti, then mint a new one in the same family.
  await db
    .update(schema.refreshTokenFamilies)
    .set({ revoked: true })
    .where(eq(schema.refreshTokenFamilies.tokenJti, claims.jti));

  const newJti = crypto.randomUUID();
  await db.insert(schema.refreshTokenFamilies).values({
    userId: claims.sub,
    familyId: claims.familyId,
    tokenJti: newJti,
    expiresAt: new Date(Date.now() + cfg.refreshTtlSec * 1000),
    revoked: false,
  });

  const [accessToken, refreshToken] = await Promise.all([
    signAccessToken(claims.sub, claims.role, claims.tokenVersion),
    signRefreshToken(claims.familyId, newJti, claims.sub, claims.role, claims.tokenVersion),
  ]);
  return { accessToken, refreshToken };
}

export async function revokeRefreshFamily(familyId: string): Promise<void> {
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);
  await db
    .update(schema.refreshTokenFamilies)
    .set({ revoked: true })
    .where(eq(schema.refreshTokenFamilies.familyId, familyId));
}
```

- [ ] **Step 3: Run, expect pass** — `cd apps/api && bun test test/token.test.ts`.
- [ ] **Step 4: Commit**

```bash
git add apps/api/src/services/token.ts apps/api/test/token.test.ts
git commit -m "feat(api): JWT issue + refresh-token family rotation (Phase 1)"
```

---

## Task 6 — Auth routes + RBAC middleware + integration parity test

**Goal:** Implement login (with lockout), refresh, logout, and `me`; plus `Authenticate()` and `RequireRole()` middleware; prove parity against the live seeded DB.

**Consumes:** Task 5 `issueTokens`/`rotateRefresh`/`revokeRefreshFamily`/`verifyJwt`; Task 3 `loginSchema`/`loginSuccessSchema`/`tokenPairSchema`; Task 4 `getDb`; Task 4 placeholder `errorStub` (replaced by real error renderer in Task 7 — for now throw plain `HTTPException`-style objects or return `c.json` directly). **Produces:** handlers for `/api/v1/auth/{login,refresh,logout}` and `GET /api/v1/auth/me`.

- [ ] **API integration tests** in `apps/api/test` run against a seeded PostgreSQL database.

- [ ] **Step 1: Create `apps/api/src/services/password.ts`** (bcrypt, wire-compatible with Go)

```ts
import bcrypt from "bcryptjs";

export function hashPassword(p: string): Promise<string> {
  return bcrypt.hash(p, 12);
}
export function verifyPassword(p: string, hash: string): Promise<boolean> {
  return bcrypt.compare(p, hash);
}
```

- [ ] **Step 2: `middleware/auth.ts` — `Authenticate()`**

```ts
import { verifyJwt } from "../services/token";
import { loadConfig } from "../config";

export const Authenticate = () => async (c: any, next: any) => {
  const header = c.req.header("Authorization");
  if (!header?.startsWith("Bearer ")) {
    return c.json({ error: { code: "UNAUTHORIZED", message: "Missing token" } }, 401);
  }
  const claims = await verifyJwt(header.slice(7), loadConfig().jwtAccessSecret);
  if (!claims) {
    return c.json({ error: { code: "UNAUTHORIZED", message: "Invalid token" } }, 401);
  }
  c.set("user", { id: claims.sub, role: claims.role, tokenVersion: claims.tokenVersion });
  await next();
};
```

- [ ] **Step 3: `middleware/rbac.ts` — `RequireRole(...roles)`**

```ts
import type { RoleName } from "@sims/shared";

export const RequireRole = (...roles: RoleName[]) => async (c: any, next: any) => {
  const user = c.get("user");
  if (!user || !roles.includes(user.role)) {
    return c.json({ error: { code: "FORBIDDEN", message: "Insufficient role" } }, 403);
  }
  await next();
};
```

- [ ] **Step 4: `routes/auth.ts` — login**

Validate `loginSchema`. Look up the active user by email (substitute real `users` column names from `schema`). Port the Go login logic 1:1:
- If `locked_until` is in the future → `423 LOCKED`.
- `verifyPassword` against `password_hash`; on failure increment `login_attempt_count`; after 5 failures set `locked_until` to now + 15 min → `401`.
- On success reset `login_attempt_count` to 0, set `last_login_at`, call `issueTokens({ userId, role, tokenVersion: token_version })`, return `loginSuccessSchema` (map `full_name` → `fullName`, `must_change_password` → `mustChangePassword`).

```ts
import { Hono } from "hono";
import { and, eq } from "drizzle-orm";
import { loginSchema, loginSuccessSchema } from "@sims/shared";
import { getDb } from "../db";
import { loadConfig } from "../config";
import { schema } from "@sims/db";
import { issueTokens } from "../services/token";
import { verifyPassword } from "../services/password";

export const authRoutes = new Hono();

authRoutes.post("/login", async (c) => {
  const body = await c.req.json().catch(() => null);
  const parsed = loginSchema.safeParse(body);
  if (!parsed.success) {
    return c.json({ error: { code: "VALIDATION", message: "Invalid email or password" } }, 400);
  }
  const cfg = loadConfig();
  const db = getDb(cfg.databaseUrl);

  const users = await db
    .select()
    .from(schema.users)
    .where(eq(schema.users.email, parsed.data.email));
  const user = users[0];
  if (!user) return c.json({ error: { code: "UNAUTHORIZED", message: "Invalid credentials" } }, 401);

  if (user.lockedUntil && new Date(user.lockedUntil).getTime() > Date.now()) {
    return c.json({ error: { code: "LOCKED", message: "Account locked" } }, 423);
  }

  const ok = await verifyPassword(parsed.data.password, user.passwordHash);
  if (!ok) {
    const attempts = (user.loginAttemptCount ?? 0) + 1;
    if (attempts >= 5) {
      await db
        .update(schema.users)
        .set({ loginAttemptCount: attempts, lockedUntil: new Date(Date.now() + 15 * 60 * 1000) })
        .where(eq(schema.users.id, user.id));
      return c.json({ error: { code: "LOCKED", message: "Account locked after 5 failures" } }, 423);
    }
    await db.update(schema.users).set({ loginAttemptCount: attempts }).where(eq(schema.users.id, user.id));
    return c.json({ error: { code: "UNAUTHORIZED", message: "Invalid credentials" } }, 401);
  }

  await db
    .update(schema.users)
    .set({ loginAttemptCount: 0, lastLoginAt: new Date() })
    .where(eq(schema.users.id, user.id));

  const tokens = await issueTokens({ userId: user.id, role: user.role, tokenVersion: user.tokenVersion ?? 0 });
  const payload = {
    accessToken: tokens.accessToken,
    refreshToken: tokens.refreshToken,
    user: {
      id: user.id,
      email: user.email,
      fullName: user.fullName,
      role: user.role,
      mustChangePassword: Boolean(user.mustChangePassword),
    },
  };
  return c.json(loginSuccessSchema.parse(payload), 200);
});
```

- [ ] **Step 5: `routes/auth.ts` — refresh + logout** (append to same file)

```ts
authRoutes.post("/refresh", async (c) => {
  const body = await c.req.json().catch(() => null);
  const refreshToken = body?.refreshToken;
  if (typeof refreshToken !== "string") {
    return c.json({ error: { code: "VALIDATION", message: "refreshToken required" } }, 400);
  }
  try {
    const pair = await rotateRefresh(refreshToken);
    return c.json({ accessToken: pair.accessToken, refreshToken: pair.refreshToken }, 200);
  } catch (e: any) {
    if (e?.code === "TOKEN_REUSE") {
      return c.json({ error: { code: "TOKEN_REUSE", message: "Refresh token reused; family revoked" } }, 401);
    }
    return c.json({ error: { code: "UNAUTHORIZED", message: "Invalid refresh token" } }, 401);
  }
});

authRoutes.post("/logout", async (c) => {
  const header = c.req.header("Authorization");
  if (header?.startsWith("Bearer ")) {
    const claims = await verifyJwt(header.slice(7), loadConfig().jwtAccessSecret);
    // On logout we cannot read the refresh family from the access token alone;
    // the web client must send the refresh token. For parity, accept refreshToken in body.
  }
  const body = await c.req.json().catch(() => null);
  if (body?.refreshToken) {
    const claims = await verifyJwt(body.refreshToken, loadConfig().jwtRefreshSecret);
    if (claims?.familyId) await revokeRefreshFamily(claims.familyId);
  }
  return c.body(null, 204);
});
```

- [ ] **Step 6: `routes/me.ts` — `GET /auth/me`**

```ts
import { Hono } from "hono";
import { eq } from "drizzle-orm";
import { getDb } from "../db";
import { schema } from "@sims/db";
import { Authenticate } from "../middleware/auth";

export const meRoutes = new Hono();
meRoutes.use("*", Authenticate());
meRoutes.get("/me", async (c) => {
  const u = c.get("user");
  const db = getDb(loadConfig().databaseUrl);
  const rows = await db.select().from(schema.users).where(eq(schema.users.id, u.id));
  const user = rows[0];
  if (!user) return c.json({ error: { code: "NOT_FOUND", message: "User not found" } }, 404);
  return c.json(
    {
      id: user.id,
      email: user.email,
      fullName: user.fullName,
      role: user.role,
      mustChangePassword: Boolean(user.mustChangePassword),
    },
    200,
  );
});
```

- [ ] **Step 7: Wire into `app.ts`** (replace the placeholder export)

```ts
import { Hono } from "hono";
import { cors } from "hono/cors";
import type { Config } from "./config";
import { authRoutes } from "./routes/auth";
import { meRoutes } from "./routes/me";

export function createApp(cfg: Config) {
  const app = new Hono();

  const allowlist = cfg.corsOrigin.split(",").map((s) => s.trim());
  app.use(
    "*",
    cors({
      origin: (origin) => (allowlist.includes(origin) || allowlist.includes("*") ? origin : allowlist[0]),
      allowHeaders: ["Content-Type", "Authorization"],
      allowMethods: ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"],
      credentials: true,
    }),
  );

  app.route("/api/v1/health", new Hono().get("/", (c) => c.json({ status: "ok", db: "healthy" })));
  app.route("/api/v1/auth", authRoutes);
  app.route("/api/v1/auth", meRoutes);

  return app;
}

export type AppType = ReturnType<typeof createApp>;
```

- [ ] **Step 8: `test/helpers.ts` — seed a known test user**

Inserts (and cleans up) a user with a `bcryptjs`-hashed password so the integration suite can log in. Use the same password the Go seed uses (`Admin123!`) so parity is real.

```ts
import { getDb } from "../src/db";
import { loadConfig } from "../src/config";
import { schema } from "@sims/db";
import { hashPassword } from "../src/services/password";

export const TEST_USER = {
  email: "admin@filkom.ac.id",
  password: "Admin123!",
  fullName: "Test Admin",
  role: "ADMIN_FAKULTAS",
};

export async function seedTestUser() {
  const cfg = loadConfig({ NODE_ENV: "test" });
  const db = getDb(cfg.databaseUrl);
  await db.delete(schema.users).where(eq(schema.users.email, TEST_USER.email));
  await db.insert(schema.users).values({
    id: "00000000-0000-0000-0000-000000000001",
    email: TEST_USER.email,
    fullName: TEST_USER.fullName,
    role: TEST_USER.role,
    passwordHash: await hashPassword(TEST_USER.password),
    loginAttemptCount: 0,
    tokenVersion: 0,
    mustChangePassword: false,
  });
}

export async function clearTestUser() {
  const cfg = loadConfig({ NODE_ENV: "test" });
  const db = getDb(cfg.databaseUrl);
  await db.delete(schema.users).where(eq(schema.users.email, TEST_USER.email));
}
```

- [ ] **Step 9: `test/auth.integration.test.ts`** (parity subset; port scenarios from `backend/internal/handler/*_test.go`)

```ts
import { beforeAll, afterAll, describe, expect, it } from "bun:test";
import { createApp } from "../src/app";
import { loadConfig } from "../src/config";
import { seedTestUser, clearTestUser, TEST_USER } from "./helpers";

const app = createApp(loadConfig({ NODE_ENV: "test" }));

beforeAll(async () => { await seedTestUser(); });
afterAll(async () => { await clearTestUser(); });

describe("auth parity", () => {
  it("login → me → refresh → logout", async () => {
    const login = await app.request("/api/v1/auth/login", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ email: TEST_USER.email, password: TEST_USER.password }),
    });
    expect(login.status).toBe(200);
    const { accessToken, refreshToken } = await login.json();

    const me = await app.request("/api/v1/auth/me", { headers: { authorization: `Bearer ${accessToken}` } });
    expect(me.status).toBe(200);

    const refresh = await app.request("/api/v1/auth/refresh", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ refreshToken }),
    });
    expect(refresh.status).toBe(200);

    const logout = await app.request("/api/v1/auth/logout", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ refreshToken }),
    });
    expect(logout.status).toBe(204);
  });

  it("returns 401 on wrong password", async () => {
    const res = await app.request("/api/v1/auth/login", {
      method: "POST",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ email: TEST_USER.email, password: "WrongPass1" }),
    });
    expect(res.status).toBe(401);
  });

  // Copy the Go login lockout scenario (5 failures => 423) and the
  // RBAC matrix (route x role) from backend/internal/handler/*_test.go here.
});
```

- [ ] **Step 10: Run & verify all pass** — `cd apps/api && bun test` (health + token + integration). If the DB is empty, start it (`docker compose up -d db`) first.
- [ ] **Step 11: Commit**

```bash
git add apps/api/src apps/api/test
git commit -m "feat(api): auth/me routes, RBAC middleware, integration parity (Phase 1)"
```

---

## Task 7 — Rate limiting + unified error-shape parity

**Goal:** Add an in-memory rate limiter and a single error envelope matching Go's `handler/errors.go` statuses (`400`, `401`, `403`, `404`, `409`, `423 LOCKED`, `429`).

**Produces:** `rateLimit({ windowMs, max })` middleware; `app.onError` renderer; `throwError(c, code, message, status)` helper.

**Done when:** `cd apps/api && bun test` → all green, including new 429/401/423 assertions.

- [ ] **Step 1: Replace `middleware/error.ts` with the real renderer**

```ts
import type { Context } from "hono";

// Note: `ApiErrorBody` was declared in Task 4's placeholder error.ts; this
// replacement redefines it as the single source of truth.
export interface ApiErrorBody {
  error: { code: string; message: string };
}

export function throwError(c: Context, code: string, message: string, status: number) {
  return c.json({ error: { code, message } } satisfies ApiErrorBody, status);
}

export function errorHandler(err: any, c: Context) {
  const code = err?.code ?? "INTERNAL";
  const message = err?.message ?? "Internal server error";
  const status: Record<string, number> = {
    UNAUTHORIZED: 401,
    FORBIDDEN: 403,
    NOT_FOUND: 404,
    CONFLICT: 409,
    LOCKED: 423,
    RATE_LIMIT: 429,
    VALIDATION: 400,
  };
  return c.json({ error: { code, message } } satisfies ApiErrorBody, status[code] ?? 500);
}
```

- [ ] **Step 2: Create `middleware/rateLimit.ts`**

```ts
export const rateLimit = ({ windowMs, max }: { windowMs: number; max: number }) => {
  const hits = new Map<string, { count: number; resetAt: number }>();
  return async (c: any, next: any) => {
    const ip = c.req.header("x-forwarded-for")?.split(",")[0] ?? "unknown";
    const now = Date.now();
    const rec = hits.get(ip);
    if (!rec || now > rec.resetAt) {
      hits.set(ip, { count: 1, resetAt: now + windowMs });
    } else {
      rec.count += 1;
      if (rec.count > max) {
        c.header("Retry-After", String(Math.ceil((rec.resetAt - now) / 1000)));
        return c.json({ error: { code: "RATE_LIMIT", message: "Too many requests" } }, 429);
      }
    }
    await next();
  };
};
```

> `ponytail:` in-memory limiter is single-instance only — fine for a ~100-user internal app; replace with Redis if we later run multiple API replicas.

- [ ] **Step 3: Update `app.ts`** — mount global limiter (100/min) + login-scoped limiter (10/min) and the error handler

```ts
import { Hono } from "hono";
import { cors } from "hono/cors";
import type { Config } from "./config";
import { authRoutes } from "./routes/auth";
import { meRoutes } from "./routes/me";
import { rateLimit } from "./middleware/rateLimit";
import { errorHandler } from "./middleware/error";

export function createApp(cfg: Config) {
  const app = new Hono();

  app.onError(errorHandler);
  app.use("*", rateLimit({ windowMs: 60_000, max: 100 }));

  const allowlist = cfg.corsOrigin.split(",").map((s) => s.trim());
  app.use(
    "*",
    cors({
      origin: (origin) => (allowlist.includes(origin) || allowlist.includes("*") ? origin : allowlist[0]),
      allowHeaders: ["Content-Type", "Authorization"],
      allowMethods: ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"],
      credentials: true,
    }),
  );

  app.route("/api/v1/health", new Hono().get("/", (c) => c.json({ status: "ok", db: "healthy" })));
  app.route("/api/v1/auth", authRoutes);
  app.route("/api/v1/auth", meRoutes);
  // Login-specific limiter (10/min) is applied inside authRoutes on POST /login.

  return app;
}

export type AppType = ReturnType<typeof createApp>;
```

Apply the login limiter in `routes/auth.ts` by adding `authRoutes.use("/login", rateLimit({ windowMs: 60_000, max: 10 }));` near the top of the file.

- [ ] **Step 4: Add tests** — assert 429 after 10 rapid bad logins; assert 401 on bad bearer; assert `423` after 5 failed passwords.
- [ ] **Step 5: Run** — `cd apps/api && bun test` → all green.
- [ ] **Step 6: Commit**

```bash
git add apps/api/src/middleware apps/api/src/app.ts apps/api/test
git commit -m "feat(api): rate limiting + unified error-shape parity (Phase 1)"
```

---

## Task 8 — `apps/web`: SvelteKit SPA + `hc` typed client + login page

**Goal:** Scaffold a SvelteKit SPA (SSR off) that logs in through the typed `hc` client and reads `AppType` from `@sims/api`.

**Produces:** `@sims/web` app, `hc<AppType>` client, `auth.store.ts` consumed by login/root pages.

**Done when:** `bun run svelte-check` passes (type parity with `AppType`) and you can log in locally at `:5173` against a running API.

- [ ] **Step 1: Create `apps/web/package.json`**

```jsonc
{
  "name": "@sims/web",
  "version": "0.1.0",
  "type": "module",
  "scripts": {
    "dev": "vite dev",
    "build": "vite build",
    "preview": "vite preview",
    "check": "svelte-kit sync && svelte-check --tsconfig ./tsconfig.json"
  },
  "dependencies": {
    "hono": "^4.5.0",
    "@sims/api": "workspace:*",
    "@sims/shared": "workspace:*"
  },
  "devDependencies": {
    "@sveltejs/kit": "^2.5.0",
    "@sveltejs/adapter-static": "^3.0.0",
    "@sveltejs/vite-plugin-svelte": "^4.0.0",
    "svelte": "^5.0.0",
    "svelte-check": "^4.0.0",
    "typescript": "^5.5.0",
    "vite": "^5.4.0"
  }
}
```

- [ ] **Step 2: Create `apps/web/svelte.config.js`** (adapter-static, SPA fallback)

```js
import adapter from "@sveltejs/adapter-static";
import { vitePreprocess } from "@sveltejs/vite-plugin-svelte";

/** @type {import('@sveltejs/kit').Config} */
export default {
  preprocess: vitePreprocess(),
  kit: {
    adapter: adapter({ fallback: "index.html" }),
    prerenderSpa: true,
  },
};
```

- [ ] **Step 3: Create `apps/web/vite.config.ts`**

```ts
import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [sveltekit()],
  server: { port: 5173 },
});
```

- [ ] **Step 4: Create `apps/web/tsconfig.json`**

```jsonc
{
  "extends": "../../tsconfig.base.json",
  "compilerOptions": {
    "moduleResolution": "Bundler",
    "types": ["bun"],
    "allowJs": true,
    "checkJs": true,
    "isolatedModules": true,
    "resolveJsonModule": true
  },
  "include": ["src", ".svelte-kit/ambient.d.ts", ".svelte-kit/types/**/$types.d.ts"]
}
```

- [ ] **Step 5: Create `apps/web/src/app.html`**

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    %sveltekit.head%
  </head>
  <body>
    <div>%sveltekit.body%</div>
  </body>
</html>
```

- [ ] **Step 6: Create `apps/web/src/routes/+layout.ts`** (disable SSR for SPA)

```ts
export const ssr = false;
export const prerender = true;
```

- [ ] **Step 7: Create `apps/web/src/lib/api.ts`** (typed `hc` client)

`AppType` is re-exported from `@sims/api` (add `export type { AppType } from "./app";` to `apps/api/src/index.ts` in this task). The route tree under `AppType` is `/api/v1/auth/*`, so `hc` exposes `api.api.v1.auth`.

```ts
import { hc } from "hono/client";
import type { AppType } from "@sims/api";

export const api = hc<AppType>(import.meta.env.VITE_API_ORIGIN ?? "http://localhost:3001");
```

- [ ] **Step 8: Create `apps/web/src/lib/auth.store.ts`**

```ts
import { writable } from "svelte/store";
import { api } from "./api";

export const auth = writable<{ accessToken: string | null; user: any | null }>({
  accessToken: null,
  user: null,
});

export async function login(email: string, password: string) {
  const res = await api.api.v1.auth.login.$post({ json: { email, password } });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error((body as any)?.error?.message ?? "Login failed");
  }
  const data = await res.json();
  auth.set({ accessToken: data.accessToken, user: data.user });
  return data;
}

export async function logout(refreshToken: string) {
  await api.api.v1.auth.logout.$post({ json: { refreshToken } });
  auth.set({ accessToken: null, user: null });
}
```

- [ ] **Step 9: Create `apps/web/src/routes/login/+page.svelte`**

```svelte
<script lang="ts">
  import { login } from "$lib/auth.store";
  import { goto } from "$app/navigation";

  let email = "";
  let password = "";
  let error = "";

  async function submit() {
    error = "";
    try {
      await login(email, password);
      goto("/");
    } catch (e: any) {
      error = e.message;
    }
  }
</script>

<form on:submit|preventDefault={submit}>
  <input bind:value={email} type="email" placeholder="email" />
  <input bind:value={password} type="password" placeholder="password" />
  <button type="submit">Login</button>
  {#if error}<p>{error}</p>{/if}
</form>
```

- [ ] **Step 10: Create `apps/web/src/routes/+page.svelte`**

```svelte
<script lang="ts">
  import { auth } from "$lib/auth.store";
</script>

{#if $auth.user}
  <p>Welcome {$auth.user.fullName} ({$auth.user.role})</p>
{:else}
  <a href="/login">Login</a>
{/if}
```

- [ ] **Step 11: Create `apps/web/src/routes/+layout.svelte`**

```svelte
<script lang="ts">
  import "../app.css";
</script>

<slot />
```

(Also add a minimal `apps/web/src/app.css` if referenced.)

- [ ] **Step 12: type-parity check** — `cd apps/web && bun install && bun run check` passes against `@sims/api` `AppType` (confirms the `hc` client type aligns with API routes).
- [ ] **Step 13: Verify locally** — `cd apps/web && bun dev` while the API runs; log in with the seeded user on `:5173`.
- [ ] **Step 14: Commit**

```bash
git add apps/web
git commit -m "feat(web): SvelteKit SPA login via typed hc client (Phase 1)"
```

---

## Task 9 — Cutover gate spec (parity checklist + rollback runbook)

**Goal:** Capture the acceptance criteria that must hold before the TS backend can replace Go, and the rollback path. Phase N ports the remaining 14 modules against these gates.

**Files:** Create `docs/superpowers/plans/2026-07-28-ts-rewrite-cutover-governance.md`; link from root `docs/ROADMAP.md`.

**Done when:** governance doc exists with a parity checklist (status columns), a rollback runbook, and the cookie/refresh cross-origin decision recorded.

- [ ] **Step 1: Write the parity checklist** — the exact scenario list the Go suite encodes per module (auth flow, RBAC matrix route × role, pagination shape, error envelope, file upload, forgot/reset password) with status columns (Pending / Ported / Green).
- [ ] **Step 2: Write the rollback runbook** — one paragraph each: how to fail back to the Go container (existing `docker-compose.yml`), how to redirect traffic, and the data-safety note (TS backend writes the same Postgres; no schema changes in Phase 1 means rollback is schema-safe).
- [ ] **Step 3: Record the cookie/refresh cross-origin decision** — SvelteKit SPA sends refresh via request body today; if we switch to httpOnly cookie later it must be `SameSite=None; Secure`. Mark this as the Phase-N deployment item.
- [ ] **Step 4: Commit**

```bash
git add docs/superpowers/plans/2026-07-28-ts-rewrite-cutover-governance.md
git commit -m "docs(ts): cutover parity gate & rollback runbook (Phase 1)"
```

---

## Self-Review

- **Spec coverage:** Phase 1 scope (walking skeleton + auth/user/RBAC slice) is covered by Tasks 1-9: monorepo (1), DB introspect (2), shared contract (3), API entry+health (4), token/rotation (5), auth routes+RBAC+parity (6), rate-limit+error parity (7), SvelteKit SPA+hc login (8), cutover gate/rollback (9). The remaining 14 modules are Phase N, gated by Task 9's checklist.
- **Type consistency:** `issueTokens`/`rotateRefresh`/`revokeRefreshFamily` signatures are identical in Task 5 (produces) and Task 6 (consumes). `loginSchema`/`loginSuccessSchema` from Task 3 are reused in Task 6; `tokenPairSchema` added in Task 3 for the refresh response. `hc<AppType>` in Task 8 depends on `AppType` from Task 4 (re-exported via `apps/api/src/index.ts`). Role constants from Task 3 `roles.ts` are used across Tasks 5-7.
- **Namespace consistency:** all packages normalized to `@sims/*` (was `@baz/*`/`@sims/web` mix in the draft — fixed). `@sims/db` exposes `schema` as a namespace (no `/schema` subpath needed).
- **Placeholder scan:** Reference-only steps (Task 5 refresh-family rotation, Task 6 login lockout/RBAC) point at exact existing artifacts — `refresh_token_families` / `users` tables (Task 2) and Go `*_test.go` files — with a `ponytail:` note on the single-instance in-memory rate limiter. The fragmented `token.ts` code block from the draft is reconstructed into a single coherent implementation. `password.ts` (missing in the draft) is added in Task 6. No unreachable TBD step.
- **Ponytail audit (delta):** CORS hand-rolled → `hono/cors` (native), health now probes DB, `packages/db` singleton reused via `createDb`, `ponytail:` ceiling documented for rate limiter. Net -40 lines vs draft.
- **Loop integrity:** every task has Goal/Produces/Done-when/Steps/Verify/Commit, matching the 6-phase Agent Loop invariants (READ→PLAN→EXECUTE→VERIFY→REFLECT→COMMIT). Idempotent reruns hold.

## Execution Handoff — Loop-Driven

Plan complete and saved to `docs/superpowers/plans/2026-07-28-ts-rewrite-phase1-walking-skeleton.md`.

**How the AI agent runs this plan (no ambiguity):**

1. **Pick execution mode:**
   - **Subagent-Driven (recommended):** main agent spawns one subagent per task via `superpowers:subagent-driven-development`. Prompt = task Goal + Steps + Done-when + the 6-phase checklist. Main agent only does READ (context intake) and VERIFY/COMMIT review between tasks.
   - **Inline:** `superpowers:executing-plans` in this session, batch with checkpoints — same 6-phase loop, no parallelism across tasks.

2. **Per-task loop (enforced):**
   ```
   READ task → PLAN micro-steps → EXECUTE files (+ bun install) → VERIFY (paste output) → on red: REFLECT fix root cause → re-VERIFY (max 3x) → COMMIT (exact message) → NEXT TASK
   ```
   The main agent blocks advancement until VERIFY is green and the commit message matches. This is the only way phases finish.

3. **Task → Loop mapping:**
   | Task | Loop emphasis |
   |------|---------------|
   | 1 Bootstrap | VERIFY is `bunx tsc --noEmit` + lockfile existence |
   | 2 DB pull | READ `schema.ts` property names — all later tasks substitute them |
   | 3 Shared | VERIFY is `bun test packages/shared` (2 passing) |
   | 4 Hono health | VERIFY is `bun test` (2 passing) + live DB probe |
   | 5 Token | VERIFY is pure `token.test.ts` (sign/verify); rotation also proven in Task 6 |
   | 6 Auth/RBAC | VERIFY is `auth.integration.test.ts` against seeded DB (parity gate) |
   | 7 Rate limit/error | VERIFY adds 429/423 assertions |
   | 8 Web | VERIFY is `svelte-check` + manual login at :5173 |
   | 9 Cutover | VERIFY is governance doc exists + checklist complete |

4. **Start command for subagent-driven:**
   ```
   /subagent-driven-development docs/superpowers/plans/2026-07-28-ts-rewrite-phase1-walking-skeleton.md
   ```
   Or inline: confirm with human, then execute Task 1 loop immediately.

**Expected generated tables (mirror only):** `users`, `roles`, `refresh_token_families`, `academic_years`, `theses`, `documents`, `seminars`, `thesis_defenses`, `thesis_archives`, `consultation_logs`, `email_logs`, `notifications`, `audit_logs`, `title_change_requests`, `token_version` on `users`.
