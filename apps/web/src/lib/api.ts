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

const apiOrigin = import.meta.env.VITE_API_ORIGIN || (import.meta.env.DEV ? "http://localhost:3001" : "");

// ponytail: typed hc client — cast to any to keep the existing Svelte route calls concise.
export const api: any = hc<AppType>(apiOrigin, {
  fetch: apiFetch as typeof fetch,
});
