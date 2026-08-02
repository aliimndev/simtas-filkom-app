import axios from 'axios'
import { API_BASE } from '@/constants/api'
import { useAuthStore } from '@/lib/stores/auth-store'

const apiClient = axios.create({
  baseURL: API_BASE,
  headers: { 'Content-Type': 'application/json' },
})

apiClient.interceptors.request.use((config) => {
  const token = useAuthStore.getState().accessToken
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
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
