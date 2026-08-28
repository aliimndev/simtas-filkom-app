#!/usr/bin/env bash
set -e
# Run api + web concurrently (Phase 1)
concurrently="bunx concurrently"
if ! command -v concurrently >/dev/null 2>&1; then
  echo "starting api and web..."
  (cd apps/api && bun run dev) &
  (cd apps/web && bun run dev) &
  wait
else
  $concurrently "bun --cwd apps/api run dev" "bun --cwd apps/web run dev"
fi
