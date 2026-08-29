import { Hono } from "hono";
import { cors } from "hono/cors";
import { sql } from "drizzle-orm";
import { setRuntimeConfig, type Config } from "./config";
import { getDb } from "./db";
import { authRoutes, meRoutes, passwordRoutes } from "./modules/auth";
import { usersRoutes } from "./modules/users";
import { rolesRoutes } from "./modules/roles";
import { academicYearsRoutes } from "./modules/academic-years";
import { thesesRoutes } from "./modules/theses";
import { archiveCreationRoutes, archivesRoutes } from "./modules/archives";
import { documentsRoutes } from "./modules/documents";
import { titleChangeRequestsRoutes } from "./modules/title-change-requests";
import { consultationLogsRoutes } from "./modules/consultation-logs";
import { dashboardRoutes } from "./modules/dashboard";
import { notificationsRoutes } from "./modules/notifications";
import { auditLogsRoutes } from "./modules/audit-logs";
import { emailLogsRoutes } from "./modules/email-logs";
import { defenseSubmissionRoutes, defensesRoutes, suratTugasRoutes } from "./modules/defenses";
import { seminarSubmissionRoutes, seminarsRoutes } from "./modules/seminars";
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

  app.get("/metrics", (c) =>
    c.text("# HELP simtas_api_up Whether the SIMTAS API is available.\n# TYPE simtas_api_up gauge\nsimtas_api_up 1\n"),
  );

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
  app.route("/api/v1/archives", archivesRoutes);
  app.route("/api/v1/theses", archiveCreationRoutes);
  app.route("/api/v1/title-change-requests", titleChangeRequestsRoutes);
  app.route("/api/v1/consultation-logs", consultationLogsRoutes);
  app.route("/api/v1/dashboard", dashboardRoutes);
  app.route("/api/v1/notifications", notificationsRoutes);
  app.route("/api/v1/audit-logs", auditLogsRoutes);
  app.route("/api/v1/email-logs", emailLogsRoutes);
  app.route("/api/v1/defenses", defensesRoutes);
  app.route("/api/v1/surat-tugas", suratTugasRoutes);
  app.route("/api/v1/theses", defenseSubmissionRoutes);
  app.route("/api/v1/seminars", seminarsRoutes);
  app.route("/api/v1/theses", seminarSubmissionRoutes);

  return app;
}

export type AppType = ReturnType<typeof createApp>;
