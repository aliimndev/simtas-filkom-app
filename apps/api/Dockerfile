FROM oven/bun:1.3.14

WORKDIR /app

COPY package.json bun.lock ./
COPY apps/api/package.json apps/api/package.json
COPY apps/web/package.json apps/web/package.json
COPY packages/db/package.json packages/db/package.json
COPY packages/shared/package.json packages/shared/package.json
RUN bun install --frozen-lockfile

COPY apps/api apps/api
COPY packages packages

ENV NODE_ENV=production
EXPOSE 3001
CMD ["bun", "run", "--cwd", "apps/api", "start"]
