import { Hono } from "hono";
import { cors } from "hono/cors";
import { sql } from "drizzle-orm";
import { setRuntimeConfig, type Config } from "./config";
import { getDb } from "./db";
import { authRoutes } from "./routes/auth";
import { meRoutes } from "./routes/me";
import { passwordRoutes } from "./routes/password";
import { usersRoutes } from "./routes/users";
import { rolesRoutes } from "./routes/roles";
import { academicYearsRoutes } from "./routes/academicYears";
import { thesesRoutes } from "./routes/theses";
import { documentsRoutes } from "./routes/documents";
import { titleChangeRequestsRoutes } from "./routes/titleChangeRequests";
import { consultationLogsRoutes } from "./routes/consultationLogs";
import { dashboardRoutes } from "./routes/dashboard";
import { notificationsRoutes } from "./routes/notifications";
import { auditLogsRoutes } from "./routes/auditLogs";
import { emailLogsRoutes } from "./routes/emailLogs";
import { rateLimit } from "./middleware/rateLimit";
import { errorHandler } from "./middleware/error";
import { requestId, logger } from "./middleware/logger";

export function createApp(cfg: Config) {
  setRuntimeConfig(cfg);
  const app = new Hono();

  // Keep the legacy loadConfig() callers aligned with this app instance even when
  // multiple app instances are created in the same process (for example in tests).
  app.use("*", async (_c, next) => {
    setRuntimeConfig(cfg);
    await next();
  });

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
  app.route("/api/v1/auth/password", passwordRoutes);
  app.route("/api/v1/users", usersRoutes);
  app.route("/api/v1/roles", rolesRoutes);
  app.route("/api/v1/academic-years", academicYearsRoutes);
  app.route("/api/v1/theses", thesesRoutes);
  app.route("/api/v1/documents", documentsRoutes);
  app.route("/api/v1/title-change-requests", titleChangeRequestsRoutes);
  app.route("/api/v1/consultation-logs", consultationLogsRoutes);
  app.route("/api/v1/dashboard", dashboardRoutes);
  app.route("/api/v1/notifications", notificationsRoutes);
  app.route("/api/v1/audit-logs", auditLogsRoutes);
  app.route("/api/v1/email-logs", emailLogsRoutes);

  return app;
}

export type AppType = ReturnType<typeof createApp>;
