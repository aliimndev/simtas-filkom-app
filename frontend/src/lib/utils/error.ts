import { AxiosError } from 'axios'

interface ApiErrorBody {
  message?: string
  error?: string
  errors?: Record<string, string>
}

/** Menarik pesan error yang user-friendly dari AxiosError / error biasa. */
export function getErrorMessage(err: unknown, fallback = 'Terjadi kesalahan. Silakan coba lagi.'): string {
  if (err instanceof AxiosError) {
    const body = err.response?.data as ApiErrorBody | undefined
    if (body?.message) return body.message
    if (body?.errors) {
      const first = Object.values(body.errors)[0]
      if (first) return first
    }
    if (body?.error) return body.error
    if (err.code === 'ECONNABORTED') return 'Koneksi timeout. Silakan coba lagi.'
    if (!err.response) return 'Tidak dapat terhubung ke server. Periksa koneksi Anda.'
    if (err.response.status === 401) return 'Sesi Anda telah berakhir. Silakan login kembali.'
    if (err.response.status === 403) return 'Anda tidak memiliki akses untuk melakukan tindakan ini.'
    if (err.response.status === 404) return 'Data tidak ditemukan.'
    if (err.response.status >= 500) return 'Terjadi kesalahan pada server. Silakan coba lagi nanti.'
  }
  if (err instanceof Error && err.message) return err.message
  return fallback
}

/** Memetakan kode error backend (jika tersedia di `message`) ke pesan yang lebih jelas. */
export function mapAuthError(message: string): string {
  const lower = message.toLowerCase()
  if (lower.includes('locked') || lower.includes('terkunci')) {
    return 'Akun Anda terkunci. Silakan hubungi admin fakultas.'
  }
  if (lower.includes('must change') || lower.includes('ubah password')) {
    return 'Anda harus mengganti password sebelum melanjutkan.'
  }
  if (lower.includes('invalid') || lower.includes('salah')) return 'Email atau password salah.'
  return message
}
