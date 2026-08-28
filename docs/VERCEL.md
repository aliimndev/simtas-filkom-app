# Vercel deployment

SIMTAS deploys `apps/web` to Vercel. The API is a separate service; the frontend does not bundle or deploy the Bun API.

## Vercel project settings

The repository includes `vercel.json` with the required monorepo settings:

- Install command: `bun install --frozen-lockfile`
- Build command: `bun run verify:vercel-env && bun run --cwd apps/web build`
- Output directory: `apps/web/build`

Connect the Vercel project to this repository and use the repository root as the project root. Do not set `apps/web` as a second nested root; the workspace dependencies are installed from the repository root.

## Environment variables

Set this variable in Vercel Project Settings → Environment Variables:

| Variable | Preview | Production | Meaning |
|---|---|---|---|
| `VITE_API_ORIGIN` | API preview/staging URL | API production URL | Absolute `http://` or `https://` origin used by the browser |

The build fails with an actionable message when `VITE_API_ORIGIN` is missing or is not an absolute HTTP(S) URL. Never put API secrets, database credentials, JWT secrets, or storage credentials in `VITE_*` variables; Vite exposes them to browser code.

## Branch and deployment flow

```text
Pull request
  ├─ GitHub Actions: typecheck, unit tests, integration tests, build
  └─ Vercel: Preview Deployment

Merge to main
  ├─ GitHub Actions: required checks
  └─ Vercel: Production Deployment
```

Protect `main` in GitHub and require the CI workflow to pass before merging. Keep the Vercel Production branch set to `main`. The workflow in `.github/workflows/deploy.yml` is intentionally manual and is only for the optional VPS/API stack; it is not part of the Vercel frontend deployment.

## Smoke test after a production deploy

Open the Vercel production URL and verify:

1. The landing page loads and the theme toggle works.
2. Login submits to the configured API origin.
3. A protected dashboard route loads after authentication.
4. Browser DevTools shows no requests to `localhost` or an empty API origin.

If the frontend loads but login fails, check `VITE_API_ORIGIN` first, then the API's CORS allowlist for the Vercel production URL.
