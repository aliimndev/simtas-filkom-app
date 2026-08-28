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

The frontend is deployed through Vercel's Git integration:

- Pull requests create Preview Deployments.
- Production deploys come from the protected `main` branch.
- GitHub Actions runs the quality gates in `.github/workflows/ci.yml`.
- Configure `VITE_API_ORIGIN` in Vercel for both Preview and Production environments. Local development can use `VITE_API_ORIGIN=http://localhost:3001` in `.env.local`.

The Docker Compose production stack remains available for the API/VPS deployment.
See [`deploy/README.md`](deploy/README.md) and [`docs/VERCEL.md`](docs/VERCEL.md).
