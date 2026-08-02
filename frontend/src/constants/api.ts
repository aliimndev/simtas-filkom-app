/** Base URL API backend. Endpoint paths didefinisikan di tiap modul API. */
export const API_BASE: string =
  process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080/api/v1'
