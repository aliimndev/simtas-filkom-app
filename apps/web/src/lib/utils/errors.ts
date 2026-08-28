export type ApiError = {
  error?: { message?: string };
  message?: string;
};

export function apiErrorMessage(err: unknown, fallback: string): string {
  const e = err as ApiError | undefined;
  return e?.error?.message ?? e?.message ?? fallback;
}

export async function readError(res: Response, fallback: string): Promise<string> {
  try {
    const body = (await res.json()) as ApiError;
    return body?.error?.message ?? body?.message ?? fallback;
  } catch {
    return fallback;
  }
}
