export const rateLimit = ({ windowMs, max }: { windowMs: number; max: number }) => {
  const hits = new Map<string, { count: number; resetAt: number }>();
  return async (c: any, next: any) => {
    const ip = c.req.header("x-forwarded-for")?.split(",")[0] ?? "unknown";
    const now = Date.now();
    const rec = hits.get(ip);
    if (!rec || now > rec.resetAt) {
      hits.set(ip, { count: 1, resetAt: now + windowMs });
    } else {
      rec.count += 1;
      if (rec.count > max) {
        c.header("Retry-After", String(Math.ceil((rec.resetAt - now) / 1000)));
        return c.json({ error: { code: "RATE_LIMIT", message: "Too many requests" } }, 429);
      }
    }
    await next();
  };
};
