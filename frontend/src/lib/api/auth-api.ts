import apiClient from './client'
import type { LoginRequest, LoginResponse, User, ForgotPasswordRequest, ResetPasswordRequest } from '@/types/auth'

export const authApi = {
  async login(data: LoginRequest): Promise<LoginResponse> {
    const res = await apiClient.post('/auth/login', data)
    return res.data.data
  },

  async logout(): Promise<void> {
    await apiClient.post('/auth/logout')
  },

  async refresh(refreshToken: string): Promise<{ access_token: string; refresh_token: string; expires_in: number }> {
    const res = await apiClient.post('/auth/refresh', { refresh_token: refreshToken })
    return res.data.data
  },

  async getMe(): Promise<User> {
    const res = await apiClient.get('/auth/me')
    return res.data.data
  },

  async forgotPassword(data: ForgotPasswordRequest): Promise<void> {
    await apiClient.post('/auth/forgot-password', data)
  },

  async resetPassword(data: ResetPasswordRequest): Promise<void> {
    await apiClient.post('/auth/reset-password', data)
  },
}
