# SIMTAS FILKOM — Bun workspace

## Local development

Requirements: Bun 1.3+ and Docker (for PostgreSQL).

```bash
bun install
bun run dev
```

The development process starts:

- Hono API on `http://localhost:3001`
- SvelteKit web app on `http://localhost:5173`

Start PostgreSQL locally with `docker compose up postgres`, or use the
connection settings in `.env.example`.

## Workspace layout

```text
apps/api       Hono REST API running on Bun
apps/web       SvelteKit frontend and Bun static production server
packages/db    Drizzle schema and SQL migrations
packages/shared Shared TypeScript contracts
```

## Validation

```bash
bun run check
bun run typecheck
bun run test
bun run build
```

## Production

The Docker Compose production stack uses Bun images for both application
services. See [`deploy/README.md`](deploy/README.md).
