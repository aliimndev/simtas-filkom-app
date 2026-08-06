import axios from 'axios'
import { API_BASE } from '@/constants/api'
import { useAuthStore } from '@/lib/stores/auth-store'

// readCookie returns the value of a document cookie by name (browser only).
function readCookie(name: string): string | undefined {
  if (typeof document === 'undefined') return undefined
  const match = document.cookie
    .split('; ')
    .find((row) => row.startsWith(`${name}=`))
  return match ? decodeURIComponent(match.split('=')[1]) : undefined
}

const apiClient = axios.create({
  baseURL: API_BASE,
  headers: { 'Content-Type': 'application/json' },
  // Required so the HttpOnly refresh-token cookie is sent on cross-origin
  // requests to the API (and so the CSRF cookie is available to read).
  withCredentials: true,
})

apiClient.interceptors.request.use((config) => {
  const token = useAuthStore.getState().accessToken
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  // Double-Submit Cookie CSRF: echo the XSRF-TOKEN cookie value as a header.
  const xsrf = readCookie('XSRF-TOKEN')
  if (xsrf) {
    config.headers['X-XSRF-TOKEN'] = xsrf
  }
  return config
})

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().clearAuth()
      if (typeof window !== 'undefined' && !window.location.pathname.startsWith('/login')) {
        window.location.href = '/login'
      }
    } else if (error.response?.status === 403) {
      if (typeof window !== 'undefined') {
        window.dispatchEvent(
          new CustomEvent('app:forbidden', {
            detail: error.response.data?.message ?? 'Akses ditolak',
          }),
        )
      }
    }
    return Promise.reject(error)
  },
)

export default apiClient
