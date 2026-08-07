export type UserRole =
  | 'admin_fakultas'
  | 'kaprodi'
  | 'mahasiswa'
  | 'dosen_pembimbing'
  | 'dosen_penguji'

export interface User {
  id: string
  email: string
  full_name: string
  role: string
  nim_nidn?: string
  study_program?: string
  is_active?: boolean
  must_change_password?: boolean
  created_at?: string
  last_login_at?: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface LoginResponse {
  access_token: string
  expires_in: number
  user: User
}

export interface ForgotPasswordRequest {
  email: string
}

export interface ResetPasswordRequest {
  token: string
  new_password: string
  confirm_password: string
}

export interface AuthState {
  user: User | null
  accessToken: string | null
  isAuthenticated: boolean
  isLoading: boolean
}
