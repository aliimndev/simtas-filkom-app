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

  // The refresh token is sent automatically via the HttpOnly cookie set at
  // login, so no token needs to be passed in the request body.
  async refresh(): Promise<{ access_token: string; expires_in: number }> {
    const res = await apiClient.post('/auth/refresh')
    return res.data.data
  },

  async getMe(): Promise<User> {
    const res = await apiClient.get('/auth/me')
    return res.data.data
  },

  // Returns the reset link in development (reset_url) so the flow can be
  // tested without a working email provider. Production omits it.
  async forgotPassword(data: ForgotPasswordRequest): Promise<{ reset_url?: string } | undefined> {
    const res = await apiClient.post('/auth/forgot-password', data)
    return res.data.data
  },

  async resetPassword(data: ResetPasswordRequest): Promise<void> {
    await apiClient.post('/auth/reset-password', data)
  },
}
