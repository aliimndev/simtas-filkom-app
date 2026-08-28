import { writable } from "svelte/store";
import { api } from "./api";

export const auth = writable<{ accessToken: string | null; user: any | null }>({
  accessToken: null,
  user: null,
});

export async function login(email: string, password: string) {
  const res = await api.api.v1.auth.login.$post({ json: { email, password } });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error((body as any)?.error?.message ?? "Login failed");
  }
  const data = (await res.json()) as any;
  auth.set({ accessToken: data.accessToken, user: data.user });
  return data;
}

export async function logout(refreshToken: string) {
  await api.api.v1.auth.logout.$post({ json: { refreshToken } });
  auth.set({ accessToken: null, user: null });
}
