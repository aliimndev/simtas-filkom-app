import { hc } from "hono/client";
import type { AppType } from "@sims/api";

// ponytail: typed hc client — cast to any to avoid Hono nested route inference issues in Svelte check
export const api: any = hc<AppType>(import.meta.env.VITE_API_ORIGIN ?? "http://localhost:3001");
