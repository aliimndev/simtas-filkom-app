import type { RoleName } from "@sims/shared";

export const RequireRole = (...roles: RoleName[]) => async (c: any, next: any) => {
  const user = c.get("user");
  if (!user || !roles.includes(user.role)) {
    return c.json({ error: { code: "FORBIDDEN", message: "Insufficient role" } }, 403);
  }
  await next();
};
