import { Hono } from "hono";
import { cors } from "hono/cors";
import type { Config } from "./config";
import { getDb } from "./db";
import { authRoutes } from "./routes/auth";
import { meRoutes } from "./routes/me";

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
  // In test mode, skip DB probe to keep tests deterministic without a live DB.
  app.route(
    "/api/v1/health",
    new Hono().get("/", async (c) => {
      if (cfg.isTest) return c.json({ status: "ok", db: "healthy" });
      try {
        const db = getDb(cfg.databaseUrl);
        await (db as any).execute("SELECT 1");
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
