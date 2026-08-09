import apiClient from './client'
import type { Thesis, ThesisFilter, CreateThesisRequest, ReviewThesisRequest, AssignSupervisorRequest } from '@/types/thesis'
import type { PaginatedResponse } from '@/types/api'

export const thesisApi = {
  async list(params?: ThesisFilter): Promise<PaginatedResponse<Thesis>> {
    const res = await apiClient.get('/theses', { params })
    return res.data
  },

  async get(id: string): Promise<Thesis> {
    const res = await apiClient.get(`/theses/${id}`)
    return res.data.data
  },

  async create(data: CreateThesisRequest): Promise<Thesis> {
    const form = new FormData()
    form.append('title', data.title)
    form.append('abstract', data.abstract)
    form.append('field_of_study', data.field_of_study)
    form.append('thesis_type', data.thesis_type)
    form.append('file', data.file)
    const res = await apiClient.post('/theses', form, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return res.data.data
  },

  async review(id: string, data: ReviewThesisRequest): Promise<Thesis> {
    const res = await apiClient.put(`/theses/${id}/review`, data)
    return res.data.data
  },

  async assignSupervisor(id: string, data: AssignSupervisorRequest): Promise<Thesis> {
    const res = await apiClient.put(`/theses/${id}/assign-supervisor`, data)
    return res.data.data
  },

  async cancel(id: string, reason?: string): Promise<Thesis> {
    const res = await apiClient.patch(`/theses/${id}/cancel`, { reason })
    return res.data.data
  },
}
