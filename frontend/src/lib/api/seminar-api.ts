import apiClient from './client'
import type { Seminar, SubmitSeminarRequest, ScheduleSeminarRequest, SubmitScoresRequest } from '@/types/seminar'
import type { PaginatedResponse } from '@/types/api'

export const seminarApi = {
  async list(params?: { page?: number; per_page?: number; status?: string; stage?: string }): Promise<PaginatedResponse<Seminar>> {
    const res = await apiClient.get('/seminars', { params })
    return res.data
  },

  async get(id: string): Promise<Seminar> {
    const res = await apiClient.get(`/seminars/${id}`)
    return res.data.data
  },

  async result(id: string): Promise<Seminar> {
    const res = await apiClient.get(`/seminars/${id}/result`)
    return res.data.data
  },

  async submit(thesisId: string, data: SubmitSeminarRequest): Promise<Seminar> {
    const res = await apiClient.post(`/theses/${thesisId}/seminars`, data)
    return res.data.data
  },

  async schedule(id: string, data: ScheduleSeminarRequest): Promise<Seminar> {
    const res = await apiClient.put(`/seminars/${id}/schedule`, data)
    return res.data.data
  },

  async setRevisionNotes(id: string, notes: string): Promise<Seminar> {
    const res = await apiClient.put(`/seminars/${id}/revision`, { notes })
    return res.data.data
  },

  async submitScores(id: string, data: SubmitScoresRequest): Promise<Seminar> {
    const res = await apiClient.post(`/seminars/${id}/scores`, data)
    return res.data.data
  },
}
