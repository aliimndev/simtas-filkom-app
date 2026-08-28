import { verifyJwt } from "../services/token";
import { loadConfig } from "../config";

export const Authenticate = () => async (c: any, next: any) => {
  const header = c.req.header("Authorization");
  if (!header?.startsWith("Bearer ")) {
    return c.json({ error: { code: "UNAUTHORIZED", message: "Missing token" } }, 401);
  }
  const claims = await verifyJwt(header.slice(7), loadConfig().jwtAccessSecret);
  if (!claims) {
    return c.json({ error: { code: "UNAUTHORIZED", message: "Invalid token" } }, 401);
  }
  c.set("user", { id: claims.sub, role: claims.role, tokenVersion: claims.tokenVersion });
  await next();
};
