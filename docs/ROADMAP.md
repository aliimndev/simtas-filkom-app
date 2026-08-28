# ROADMAP

## Phase 1 — Walking Skeleton + Auth Slice (2026-07-28)
- **Status:** Green
- **Plan:** `docs/superpowers/plans/2026-07-28-ts-rewrite-phase1-walking-skeleton.md`
- **Governance:** `docs/superpowers/plans/2026-07-28-ts-rewrite-cutover-governance.md` (parity gate + rollback runbook)
- **Scope:** Bun monorepo → Hono API → Drizzle Postgres → SvelteKit SPA, auth + user + RBAC vertical slice, JWT rotation with family revoke-on-reuse, rate limiting, error envelope.

## Phase N — Remaining 14 Modules
Order: `users → roles → academic_years → theses → title_change_requests → consultation_logs → seminars → thesis_defenses → documents → thesis_archives → notifications → email_logs → audit_logs → dashboard`
Each follows the 6-phase Agent Loop and must add its row to the governance parity checklist before merge.

## Infra
- API: `oven/bun:1` Docker, `Bun.serve` port 3001
- Web: `adapter-static` SPA, `fallback: index.html`
- DB: Postgres 18 (existing, `drizzle-kit pull` mirror)
