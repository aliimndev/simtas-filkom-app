import apiClient from './client'
import type { Consultation, CreateConsultationRequest, UpdateConsultationRequest, ConsultationSummary } from '@/types/consultation'
import type { PaginatedResponse } from '@/types/api'

export const consultationApi = {
  async list(thesisId: string, params?: { page?: number; per_page?: number; status?: string }): Promise<PaginatedResponse<Consultation>> {
    const res = await apiClient.get(`/theses/${thesisId}/consultations`, { params })
    return res.data
  },

  async get(thesisId: string, id: string): Promise<Consultation> {
    const res = await apiClient.get(`/theses/${thesisId}/consultations/${id}`)
    return res.data.data
  },

  async summary(thesisId: string): Promise<ConsultationSummary> {
    const res = await apiClient.get(`/theses/${thesisId}/consultations/summary`)
    return res.data.data
  },

  async create(thesisId: string, data: CreateConsultationRequest): Promise<Consultation> {
    const res = await apiClient.post(`/theses/${thesisId}/consultations`, data)
    return res.data.data
  },

  async update(thesisId: string, id: string, data: UpdateConsultationRequest): Promise<Consultation> {
    const res = await apiClient.put(`/theses/${thesisId}/consultations/${id}`, data)
    return res.data.data
  },

  async approve(thesisId: string, id: string, notes?: string): Promise<Consultation> {
    const res = await apiClient.patch(`/theses/${thesisId}/consultations/${id}/approve`, { notes })
    return res.data.data
  },

  async remove(thesisId: string, id: string): Promise<void> {
    await apiClient.delete(`/theses/${thesisId}/consultations/${id}`)
  },
}
