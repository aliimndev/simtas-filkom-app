import apiClient from './client'
import type {
  TitleChangeRequest,
  CreateTitleChangeRequest,
  ReviewTitleChangeRequest,
} from '@/types/title-change'
import type { ApiResponse } from '@/types/api'

export const titleChangeApi = {
  /** POST /theses/:thesis_id/title-change-requests — Mahasiswa pemilik. */
  async create(thesisId: string, data: CreateTitleChangeRequest): Promise<TitleChangeRequest> {
    const res = await apiClient.post<ApiResponse<TitleChangeRequest>>(
      `/theses/${thesisId}/title-change-requests`,
      data,
    )
    return res.data.data
  },

  /** GET /theses/:thesis_id/title-change-requests — riwayat sebuah thesis. */
  async list(thesisId: string): Promise<TitleChangeRequest[]> {
    const res = await apiClient.get<ApiResponse<TitleChangeRequest[]>>(
      `/theses/${thesisId}/title-change-requests`,
    )
    return res.data.data
  },

  /** POST /title-change-requests/:id/cancel — Mahasiswa pemilik, PENDING only. */
  async cancel(id: string): Promise<TitleChangeRequest> {
    const res = await apiClient.post<ApiResponse<TitleChangeRequest>>(`/title-change-requests/${id}/cancel`)
    return res.data.data
  },

  /** GET /title-change-requests — antrian PENDING untuk dosen pembimbing. */
  async listPending(): Promise<TitleChangeRequest[]> {
    const res = await apiClient.get<ApiResponse<TitleChangeRequest[]>>('/title-change-requests')
    return res.data.data
  },

  /** POST /title-change-requests/:id/approve — Dosen pembimbing assigned. */
  async approve(id: string, data?: ReviewTitleChangeRequest): Promise<TitleChangeRequest> {
    const res = await apiClient.post<ApiResponse<TitleChangeRequest>>(`/title-change-requests/${id}/approve`, data ?? {})
    return res.data.data
  },

  /** POST /title-change-requests/:id/reject — catatan penolakan wajib. */
  async reject(id: string, data: ReviewTitleChangeRequest): Promise<TitleChangeRequest> {
    const res = await apiClient.post<ApiResponse<TitleChangeRequest>>(`/title-change-requests/${id}/reject`, data)
    return res.data.data
  },
}
