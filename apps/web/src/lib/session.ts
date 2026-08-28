let accessToken: string | null = null;
let refreshToken: string | null = null;

export function getAccessToken(): string | null {
  return accessToken;
}

export function getRefreshToken(): string | null {
  return refreshToken;
}

export function setSession(tokens: { accessToken: string | null; refreshToken?: string | null }): void {
  accessToken = tokens.accessToken;
  if (tokens.refreshToken !== undefined) refreshToken = tokens.refreshToken;
}

export function clearSession(): void {
  accessToken = null;
  refreshToken = null;
}
