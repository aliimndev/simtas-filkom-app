# ADR-002: TypeScript Schema — Manual Mirror vs `drizzle-kit pull`

**Status:** Accepted  \
**Date:** 2026-08-28  \
**Area:** `packages/db` (TypeScript rewrite, Phase 1)  \
**Related plan:** `docs/superpowers/plans/2026-07-28-ts-rewrite-phase1-walking-skeleton.md` (Task 2)

---

## Context

The Phase 1 walking-skeleton plan (Task 2) specified that `packages/db/src/schema.ts`
be **generated** from the live Postgres database via `drizzle-kit pull`, matching the
snake_case column names verbatim, and that the result must never be hand-edited
("Zero schema drift").

The committed `packages/db/src/schema.ts` is **not** a `drizzle-kit pull` artifact. Its
header explicitly says:

> "MANUAL mirror of the live Postgres schema (camelCase props -> snake_case columns).
> ponytail: we hand-mirror because `drizzle-kit pull` emits snake_case property names
> that break the camelCase query code."

So there is a documented deviation between the plan's stated mechanism and the
implementation reality. This ADR records that deviation so the plan and reality stay
in sync (Ticket 6).

---

## Decision

Keep `packages/db/src/schema.ts` as a **hand-maintained mirror** of the live Postgres
schema, not a regenerated `pull` artifact.

- Property names are **camelCase** (`userId`, `tokenJti`, `expiresAt`) mapping to
  snake_case columns (`user_id`, `token_jti`, `expires_at`) via `drizzle-orm/pg-core`.
- All 22 live Postgres tables are mirrored (roles, users, refresh_token_families,
  token_blacklist, password_reset_tokens, academic_years, theses,
  thesis_supervisors, seminars, seminar_examiners, seminar_scores, thesis_defenses,
  defense_examiners, defense_scores, documents, thesis_archives, notifications,
  email_logs, consultation_logs, audit_logs, title_change_requests).
- The mirror is kept in sync **by code review** against the Go migrations, not by a
  generative CLI. Any new Go migration that adds/renames a column must update this
  mirror in the same PR.

### Why not `drizzle-kit pull`?
- `drizzle-kit@0.28` / `drizzle-orm@0.36` refuse to run `drizzle-kit check`
  ("Please install latest version of drizzle-orm"); the whole Phase 1 toolchain is
  pinned to that pair.
- `pull` emits snake_case property keys, which would force every query to use
  `user_id`, `token_jti` etc. — inconsistent with the camelCase codebase and the
  `{error:{code,message}}`/camelCase wire contract (see ADR on parity
  decision in the cutover governance doc).

### Trade-off
- ❌ Loses the "regen from live DB" guarantee; drift must be caught by review + the
  parity gate, not by `drizzle-kit check`.
- ✅ Keeps one runtime/one language ergonomics and the camelCase query style; avoids
  a toolchain upgrade mid-Phase-1.

---

## Consequences

- `drizzle-kit check` remains **deferred** in the CI gate (already noted in the
  cutover governance doc §4). The schema-accuracy gate is human review of the mirror
  against Go migrations.
- The database is the source of truth for shape; the mirror is a typed convenience.

---

## Links
- Plan Task 2: `docs/superpowers/plans/2026-07-28-ts-rewrite-phase1-walking-skeleton.md`
- Governance/CI gate: `docs/superpowers/plans/2026-07-28-ts-rewrite-cutover-governance.md` §4
