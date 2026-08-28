import { Hono } from "hono";
import { Authenticate } from "../middleware/auth";
import { RequireRole } from "../middleware/rbac";
import { getSummary } from "../services/dashboard";

export const dashboardRoutes = new Hono();

dashboardRoutes.get(
  "/summary",
  Authenticate(),
  RequireRole("ADMIN_FAKULTAS", "KAPRODI"),
  async (c: any) => {
    const summary = await getSummary();
    return c.json(summary, 200);
  },
);
