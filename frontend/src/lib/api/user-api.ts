import apiClient from './client'
import type { User } from '@/types/auth'
import type { PaginatedResponse } from '@/types/api'

export interface CreateUserRequest {
  email: string
  full_name: string
  role: string
  nim_nidn?: string
  study_program?: string
  initial_password?: string
}

export interface UpdateUserRequest {
  full_name?: string
  role?: string
  nim_nidn?: string
  study_program?: string
  place_of_birth?: string
  address?: string
  phone?: string
  birth_date?: string
  faculty?: string
  semester?: number
}

export interface AcademicYear {
  id: string
  name: string
  start_date: string
  end_date: string
  is_active: boolean
}

export const userApi = {
  async list(params?: { q?: string; role?: string; page?: number; per_page?: number }): Promise<PaginatedResponse<User>> {
    const res = await apiClient.get('/admin/users', { params })
    return res.data
  },

  /** Daftar dosen (endpoint /lecturers, tersedia untuk semua user terautentikasi).
   *  Backend mengembalikan array langsung (bukan paginated). */
  async lecturers(): Promise<User[]> {
    const res = await apiClient.get('/lecturers')
    return res.data.data
  },

  async get(id: string): Promise<User> {
    const res = await apiClient.get(`/admin/users/${id}`)
    return res.data.data
  },

  async create(data: CreateUserRequest): Promise<User> {
    const res = await apiClient.post('/admin/users', data)
    return res.data.data
  },

  async update(id: string, data: UpdateUserRequest): Promise<User> {
    const res = await apiClient.put(`/admin/users/${id}`, data)
    return res.data.data
  },

  /** Update profil sendiri (PATCH /users/me). */
  async updateMe(data: UpdateUserRequest): Promise<User> {
    const res = await apiClient.patch('/users/me', data)
    return res.data.data
  },

  async deactivate(id: string): Promise<User> {
    const res = await apiClient.patch(`/admin/users/${id}/deactivate`)
    return res.data.data
  },

  async activate(id: string): Promise<User> {
    const res = await apiClient.patch(`/admin/users/${id}/activate`)
    return res.data.data
  },

  async resetPassword(id: string): Promise<void> {
    await apiClient.post(`/admin/users/${id}/reset-password`)
  },

  async changeMyPassword(currentPassword: string, newPassword: string): Promise<void> {
    await apiClient.put('/users/me/password', {
      current_password: currentPassword,
      new_password: newPassword,
    })
  },
}

export const academicYearApi = {
  async list(): Promise<AcademicYear[]> {
    const res = await apiClient.get('/academic-years')
    return res.data.data
  },

  async create(data: Omit<AcademicYear, 'id' | 'is_active'>): Promise<AcademicYear> {
    const res = await apiClient.post('/admin/academic-years', data)
    return res.data.data
  },

  async activate(id: string): Promise<AcademicYear> {
    const res = await apiClient.patch(`/admin/academic-years/${id}/activate`)
    return res.data.data
  },
}
