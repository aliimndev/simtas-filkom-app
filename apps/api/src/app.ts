import { Hono } from "hono";
import { cors } from "hono/cors";
import { sql } from "drizzle-orm";
import type { Config } from "./config";
import { getDb } from "./db";
import { authRoutes } from "./routes/auth";
import { meRoutes } from "./routes/me";
import { rateLimit } from "./middleware/rateLimit";
import { errorHandler } from "./middleware/error";
import { requestId, logger } from "./middleware/logger";

export function createApp(cfg: Config) {
  const app = new Hono();

  app.onError(errorHandler);
  app.use("*", requestId);
  app.use("*", logger);
  app.use("*", rateLimit({ windowMs: 60_000, max: 100 }));

  // CORS via native middleware (ponytail: native). Allowlist supports comma-separated CORS_ORIGIN in prod.
  // Unknown origin => no ACAO header (never fall back to a allowlisted origin — that leaks cross-origin).
  const allowlist = cfg.corsOrigin.split(",").map((s) => s.trim());
  app.use(
    "*",
    cors({
      origin: (origin) =>
        origin && (allowlist.includes(origin) || allowlist.includes("*")) ? origin : undefined,
      allowHeaders: ["Content-Type", "Authorization"],
      allowMethods: ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"],
      credentials: true,
    }),
  );

  // Health does a live DB probe via SELECT 1 (not hardcoded). 503 when unreachable.
  app.route(
    "/api/v1/health",
    new Hono().get("/", async (c) => {
      try {
        const db = getDb(cfg.databaseUrl);
        await db.execute(sql`SELECT 1`);
        return c.json({ status: "ok", db: "healthy" });
      } catch {
        return c.json({ status: "error", db: "unreachable" }, 503);
      }
    }),
  );

  app.route("/api/v1/auth", authRoutes);
  app.route("/api/v1/auth", meRoutes);

  return app;
}

export type AppType = ReturnType<typeof createApp>;
