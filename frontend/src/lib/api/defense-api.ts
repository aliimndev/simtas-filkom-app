import apiClient from './client'
import type { Defense, ScheduleDefenseRequest, SubmitDefenseScoresRequest, GraduationRequest } from '@/types/defense'
import type { PaginatedResponse } from '@/types/api'
import type { UpcomingSchedules } from '@/types/dashboard'

export const defenseApi = {
  async list(params?: { page?: number; per_page?: number; status?: string }): Promise<PaginatedResponse<Defense>> {
    const res = await apiClient.get('/defenses', { params })
    return res.data
  },

  async get(id: string): Promise<Defense> {
    const res = await apiClient.get(`/defenses/${id}`)
    return res.data.data
  },

  async result(id: string): Promise<Defense> {
    const res = await apiClient.get(`/defenses/${id}/result`)
    return res.data.data
  },

  async submit(thesisId: string): Promise<Defense> {
    const res = await apiClient.post(`/theses/${thesisId}/defenses`)
    return res.data.data
  },

  async schedule(id: string, data: ScheduleDefenseRequest): Promise<Defense> {
    const res = await apiClient.put(`/defenses/${id}/schedule`, data)
    return res.data.data
  },

  async setRevisionNotes(id: string, notes: string): Promise<Defense> {
    const res = await apiClient.put(`/defenses/${id}/revision`, { notes })
    return res.data.data
  },

  async submitScores(id: string, data: SubmitDefenseScoresRequest): Promise<Defense> {
    const res = await apiClient.post(`/defenses/${id}/scores`, data)
    return res.data.data
  },

  async graduate(thesisId: string, data: GraduationRequest): Promise<Defense> {
    const res = await apiClient.put(`/theses/${thesisId}/graduation`, data)
    return res.data.data
  },

  /** Jadwal seminar + sidang mendatang (Admin/Kaprodi). Backend mengembalikan
   *  { seminars: [...], defenses: [...] } dari DefenseUseCase.Upcoming. */
  async upcoming(): Promise<UpcomingSchedules> {
    const res = await apiClient.get('/schedules/upcoming')
    return res.data.data
  },
}
