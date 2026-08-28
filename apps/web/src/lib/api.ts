import { hc } from "hono/client";
import type { AppType } from "@sims/api";
import { getAccessToken } from "./session";

const apiFetch = (input: RequestInfo | URL, init?: RequestInit): Promise<Response> => {
  const headers = new Headers(init?.headers);
  const token = getAccessToken();
  if (token && !headers.has("Authorization")) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  return fetch(input, { ...init, headers });
};

// ponytail: typed hc client — cast to any to keep the existing Svelte route calls concise.
export const api: any = hc<AppType>(import.meta.env.VITE_API_ORIGIN ?? "http://localhost:3001", {
  fetch: apiFetch as typeof fetch,
});
