FROM oven/bun:1.3.14

WORKDIR /app

COPY package.json bun.lock ./
COPY apps/api/package.json apps/api/package.json
COPY apps/web/package.json apps/web/package.json
COPY packages/db/package.json packages/db/package.json
COPY packages/shared/package.json packages/shared/package.json
RUN bun install --frozen-lockfile

COPY apps/api apps/api
COPY apps/web apps/web
COPY packages packages

ARG VITE_API_ORIGIN=
ENV VITE_API_ORIGIN=${VITE_API_ORIGIN}
RUN bun run --cwd apps/web build

ENV NODE_ENV=production
EXPOSE 4173
CMD ["bun", "run", "--cwd", "apps/web", "start"]
